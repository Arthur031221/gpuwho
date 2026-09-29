// Package tui is the interactive bubbletea view: a sortable process
// table refreshed on a timer, with a header showing system GPU and ANE
// state.
package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Arthur031221/gpuwho/internal/model"
	"github.com/Arthur031221/gpuwho/internal/render"
	"github.com/Arthur031221/gpuwho/internal/snapshot"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type tickMsg time.Time

type snapMsg struct {
	snap model.Snapshot
}

// Model is the bubbletea application state.
type Model struct {
	ctx      context.Context
	cfg      snapshot.Config
	interval time.Duration
	sortCol  string

	snap  model.Snapshot
	tbl   table.Model
	width int
	err   string
}

// New builds the initial TUI model.
func New(ctx context.Context, cfg snapshot.Config, interval time.Duration, sortCol string) Model {
	columns := []table.Column{
		{Title: "PID", Width: 7},
		{Title: "KIND", Width: 12},
		{Title: "NAME", Width: 24},
		{Title: "CPU%", Width: 6},
		{Title: "RSS", Width: 9},
		{Title: "MODEL", Width: 16},
		{Title: "TOK/S", Width: 6},
		{Title: "NOTE", Width: 30},
	}
	t := table.New(table.WithColumns(columns), table.WithFocused(true), table.WithHeight(20))
	return Model{ctx: ctx, cfg: cfg, interval: interval, sortCol: sortCol, tbl: t}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchCmd(m.ctx, m.cfg), tickCmd(m.interval))
}

func fetchCmd(ctx context.Context, cfg snapshot.Config) tea.Cmd {
	return func() tea.Msg {
		return snapMsg{snap: snapshot.Build(ctx, cfg)}
	}
}

func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "c":
			m.sortCol = "cpu"
			snapshot.SortBy(m.snap.Processes, m.sortCol)
			m.applyRows()
		case "r":
			m.sortCol = "rss"
			snapshot.SortBy(m.snap.Processes, m.sortCol)
			m.applyRows()
		case "p":
			m.sortCol = "pid"
			snapshot.SortBy(m.snap.Processes, m.sortCol)
			m.applyRows()
		case "n":
			m.sortCol = "name"
			snapshot.SortBy(m.snap.Processes, m.sortCol)
			m.applyRows()
		case "k":
			m.sortCol = "kind"
			snapshot.SortBy(m.snap.Processes, m.sortCol)
			m.applyRows()
		default:
			var cmd tea.Cmd
			m.tbl, cmd = m.tbl.Update(msg)
			return m, cmd
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.tbl.SetHeight(msg.Height - 8)
		return m, nil
	case tickMsg:
		return m, tea.Batch(fetchCmd(m.ctx, m.cfg), tickCmd(m.interval))
	case snapMsg:
		m.snap = msg.snap
		snapshot.SortBy(m.snap.Processes, m.sortCol)
		m.applyRows()
		return m, nil
	}
	return m, nil
}

func (m *Model) applyRows() {
	rows := make([]table.Row, 0, len(m.snap.Processes))
	for _, p := range m.snap.Processes {
		kind := string(p.Kind)
		name := p.Name
		if kind == "" {
			kind = "-"
		} else {
			name = "* " + name
		}
		modelName := p.Model
		if modelName == "" {
			modelName = "-"
		}
		toks := "-"
		if p.TokensPerSec != nil {
			toks = fmt.Sprintf("%.1f", *p.TokensPerSec)
		}
		note := p.Note
		if note == "" {
			note = "-"
		}
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.PID),
			kind,
			name,
			fmt.Sprintf("%.1f", p.CPUPercent),
			humanBytes(p.RSSBytes),
			modelName,
			toks,
			note,
		})
	}
	m.tbl.SetRows(rows)
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func (m Model) View() string {
	header := headerStyle.Render("gpuwho") + "  " + dimStyle.Render("live per-process GPU/ANE attribution for Apple Silicon")
	body := render.Header(m.snap)
	footer := footerStyle.Render("sort: c=cpu r=rss p=pid n=name k=kind   q=quit   sorted by " + m.sortCol)
	return header + "\n" + body + "\n" + m.tbl.View() + "\n" + footer + "\n"
}

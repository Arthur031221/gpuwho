// Package render formats a model.Snapshot as plain text, shared by
// --once output and the TUI's static header/footer.
package render

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/Arthur031221/gpuwho/internal/model"
)

// Header renders the GPU and power summary lines.
func Header(snap model.Snapshot) string {
	var b strings.Builder
	if snap.GPU.Available {
		chip := snap.GPU.ChipModel
		if chip == "" {
			chip = "unknown chip"
		}
		fmt.Fprintf(&b, "GPU  %3d%%  (%s, %d cores, tiler %d%%, renderer %d%%)  [ioreg, system-wide]\n",
			snap.GPU.UtilPercent, chip, snap.GPU.CoreCount, snap.GPU.TilerPercent, snap.GPU.RendererPercent)
	} else {
		fmt.Fprintf(&b, "GPU  unavailable (%s)\n", orDash(snap.GPU.Error))
	}
	if snap.Power.Available {
		fmt.Fprintf(&b, "ANE  %.0f mW   GPU power %.0f mW   [powermetrics, system-wide, root]\n", snap.Power.ANEMilliW, snap.Power.GPUMilliW)
	} else {
		fmt.Fprintf(&b, "ANE  unavailable (%s)\n", orDash(snap.Power.Unavailable))
	}
	return b.String()
}

func orDash(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}

// Table renders the process table with a fixed column layout.
func Table(w io.Writer, rows []model.ProcessRow) {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "PID\tKIND\tNAME\tCPU%\tRSS\tMODEL\tTOK/S\tNOTE")
	for _, r := range rows {
		kind := string(r.Kind)
		if kind == "" {
			kind = "-"
		}
		modelName := r.Model
		if modelName == "" {
			modelName = "-"
		}
		toks := "-"
		if r.TokensPerSec != nil {
			toks = fmt.Sprintf("%.1f", *r.TokensPerSec)
		}
		note := r.Note
		if note == "" {
			note = "-"
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%.1f\t%s\t%s\t%s\t%s\n",
			r.PID, kind, truncate(r.Name, 28), r.CPUPercent, humanBytes(r.RSSBytes), modelName, toks, truncate(note, 40))
	}
	tw.Flush()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
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

// Package snapshot combines the individual samplers into one
// model.Snapshot for the TUI and --json/--once output to share.
package snapshot

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/Arthur031221/gpuwho/internal/gpuinfo"
	"github.com/Arthur031221/gpuwho/internal/llamametrics"
	"github.com/Arthur031221/gpuwho/internal/model"
	"github.com/Arthur031221/gpuwho/internal/ollamaps"
	"github.com/Arthur031221/gpuwho/internal/power"
	"github.com/Arthur031221/gpuwho/internal/procs"
)

// Config controls which optional samplers Build reaches for.
type Config struct {
	OllamaURL         string
	LlamaServerURL    string
	EnablePower       bool
	PowerSampleMillis int
}

// Build takes one full sample. Every sampler fails soft: if ps, ioreg,
// Ollama or llama-server are unavailable, the corresponding field is
// left empty or marked Available=false rather than returning an error,
// so a partial environment still renders something useful.
func Build(ctx context.Context, cfg Config) model.Snapshot {
	snap := model.Snapshot{Time: time.Now()}

	// Sample fills Available=false and Error itself on failure, so the
	// error return here is informational only.
	snap.GPU, _ = gpuinfo.Sample(ctx)

	if cfg.EnablePower {
		p, _ := power.Sample(ctx, cfg.PowerSampleMillis)
		snap.Power = p
	} else {
		snap.Power = model.PowerStat{Available: false, Unavailable: "not requested: pass --powermetrics to try (needs root)"}
	}

	rows, _ := procs.List(ctx)

	loaded, _ := ollamaps.Fetch(ctx, cfg.OllamaURL)
	attachOllamaModel(rows, loaded)

	if cfg.LlamaServerURL != "" {
		if metrics, err := llamametrics.Fetch(ctx, cfg.LlamaServerURL); err == nil && metrics.HasPredicted {
			attachTokensPerSec(rows, metrics.PredictedTokensPerSec)
		}
	}

	sortRelevance(rows)
	snap.Processes = rows
	return snap
}

// attachOllamaModel labels llama-server rows with the model name when
// exactly one model is loaded in Ollama. Ollama's /api/ps has no PID
// field, so with more than one model loaded gpuwho cannot tell which
// runner is which and leaves Model blank rather than guessing.
func attachOllamaModel(rows []model.ProcessRow, loaded []ollamaps.Model) {
	if len(loaded) != 1 {
		return
	}
	for i := range rows {
		if rows[i].Kind == model.KindLlamaServer {
			rows[i].Model = loaded[0].Name
		}
	}
}

// attachTokensPerSec puts the llama-server /metrics reading on the first
// standalone llama-server row (one that is not Ollama's vendored
// runner). Ollama does not start its runner with --metrics, so its rows
// are skipped.
func attachTokensPerSec(rows []model.ProcessRow, tps float64) {
	for i := range rows {
		if rows[i].Kind == model.KindLlamaServer && !strings.Contains(rows[i].Note, "ollama") {
			v := tps
			rows[i].TokensPerSec = &v
			return
		}
	}
}

func sortRelevance(rows []model.ProcessRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		ki, kj := rows[i].Kind != "", rows[j].Kind != ""
		if ki != kj {
			return ki
		}
		return rows[i].CPUPercent > rows[j].CPUPercent
	})
}

// SortBy reorders an existing snapshot's process rows by a named column,
// used by both --json (via --sort) and the TUI's sort keys.
func SortBy(rows []model.ProcessRow, column string) {
	switch column {
	case "cpu":
		sort.SliceStable(rows, func(i, j int) bool { return rows[i].CPUPercent > rows[j].CPUPercent })
	case "rss":
		sort.SliceStable(rows, func(i, j int) bool { return rows[i].RSSBytes > rows[j].RSSBytes })
	case "pid":
		sort.SliceStable(rows, func(i, j int) bool { return rows[i].PID < rows[j].PID })
	case "name":
		sort.SliceStable(rows, func(i, j int) bool { return strings.ToLower(rows[i].Name) < strings.ToLower(rows[j].Name) })
	default:
		sortRelevance(rows)
	}
}

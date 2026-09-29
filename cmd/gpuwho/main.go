// Command gpuwho shows live per-process GPU and Neural Engine
// attribution on Apple Silicon, highlighting ollama, llama-server and
// MLX processes.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Arthur031221/gpuwho/internal/render"
	"github.com/Arthur031221/gpuwho/internal/snapshot"
	"github.com/Arthur031221/gpuwho/internal/tui"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gpuwho", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprint(stderr, usageText)
	}

	jsonOut := fs.Bool("json", false, "print one snapshot as JSON and exit")
	once := fs.Bool("once", false, "print one snapshot as a table and exit (no TUI)")
	interval := fs.Duration("interval", time.Second, "TUI refresh interval")
	sortCol := fs.String("sort", "relevance", "sort column: relevance, cpu, rss, pid, name")
	ollamaURL := fs.String("ollama-url", "http://localhost:11434", "Ollama base URL for model attribution")
	llamaServerURL := fs.String("llama-server-url", "", "llama-server base URL to read tokens/sec from /metrics (started with --metrics)")
	powermetricsFlag := fs.Bool("powermetrics", false, "also sample ANE and GPU power with powermetrics (needs root, see README)")
	powerSampleMillis := fs.Int("powermetrics-sample-ms", 1000, "powermetrics sample duration in milliseconds")
	showVersion := fs.Bool("version", false, "print version and exit")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	if *showVersion {
		fmt.Fprintf(stdout, "gpuwho %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return 0
	}

	if runtime.GOOS != "darwin" {
		fmt.Fprintln(stderr, "gpuwho only supports macOS (Apple Silicon). It reads ioreg and proc data that do not exist on other platforms.")
		return 1
	}

	cfg := snapshot.Config{
		OllamaURL:         *ollamaURL,
		LlamaServerURL:    *llamaServerURL,
		EnablePower:       *powermetricsFlag,
		PowerSampleMillis: *powerSampleMillis,
	}

	ctx := context.Background()

	if *jsonOut {
		snap := snapshot.Build(ctx, cfg)
		snapshot.SortBy(snap.Processes, normalizeSort(*sortCol))
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(snap); err != nil {
			fmt.Fprintln(stderr, "gpuwho: encode json:", err)
			return 1
		}
		return 0
	}

	if *once {
		snap := snapshot.Build(ctx, cfg)
		snapshot.SortBy(snap.Processes, normalizeSort(*sortCol))
		fmt.Fprint(stdout, render.Header(snap))
		fmt.Fprintln(stdout)
		render.Table(stdout, snap.Processes)
		return 0
	}

	m := tui.New(ctx, cfg, *interval, normalizeSort(*sortCol))
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(stderr, "gpuwho:", err)
		return 1
	}
	return 0
}

func normalizeSort(s string) string {
	switch s {
	case "cpu", "rss", "pid", "name", "kind":
		return s
	default:
		return "relevance"
	}
}

const usageText = `gpuwho: Activity Monitor's GPU column for local LLM people.

Live per-process GPU and Neural Engine attribution on Apple Silicon,
highlighting ollama, llama-server and MLX processes.

Usage:
  gpuwho [flags]

Flags:
  --json                       print one snapshot as JSON and exit
  --once                       print one snapshot as a table and exit (no TUI)
  --interval duration          TUI refresh interval (default 1s)
  --sort string                relevance, cpu, rss, pid, name (default relevance)
  --ollama-url string          Ollama base URL (default http://localhost:11434)
  --llama-server-url string    llama-server base URL, reads /metrics for tokens/sec
  --powermetrics                also sample ANE and GPU power (needs root)
  --powermetrics-sample-ms int  powermetrics sample duration in ms (default 1000)
  --version                    print version and exit
  --help                       show this message

In the TUI: c/r/p/n/k sort by CPU, RSS, PID, name or kind. q quits.

macOS has no public per-process GPU time counter. gpuwho reads
system-wide GPU utilization from ioreg (no root needed) and highlights
which processes are known local inference runtimes. See README.md "How
it works" for what was measured on this machine.
`

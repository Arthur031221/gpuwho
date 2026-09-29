// Package procs lists local processes and flags the ones known to run
// local LLM inference: ollama, its vendored llama-server runner, and
// python processes using MLX.
package procs

import (
	"context"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Arthur031221/gpuwho/internal/model"
)

var psLine = regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+([\d.]+)\s+(\d+)\s+(.*)$`)

// List shells out to ps and returns every process on the system.
func List(ctx context.Context) ([]model.ProcessRow, error) {
	cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(cctx, "ps", "-axo", "pid=,ppid=,pcpu=,rss=,command=").Output()
	if err != nil {
		return nil, err
	}
	return ParsePS(string(out)), nil
}

// ParsePS parses `ps -axo pid=,ppid=,pcpu=,rss=,command=` output. It is a
// pure function so tests do not depend on the local process table.
func ParsePS(out string) []model.ProcessRow {
	var rows []model.ProcessRow
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		m := psLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		pid, _ := strconv.Atoi(m[1])
		ppid, _ := strconv.Atoi(m[2])
		cpu, _ := strconv.ParseFloat(m[3], 64)
		rssKB, _ := strconv.ParseInt(m[4], 10, 64)
		command := strings.TrimSpace(m[5])
		if command == "" {
			continue
		}

		row := model.ProcessRow{
			PID:        pid,
			PPID:       ppid,
			Name:       processName(command),
			Command:    command,
			CPUPercent: cpu,
			RSSBytes:   rssKB * 1024,
		}
		row.Kind, row.Note = classify(row.Name, command)
		rows = append(rows, row)
	}
	return rows
}

// processName takes the display name from the first token of the
// command line (the executable path), matching what Activity Monitor
// shows.
func processName(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return command
	}
	exe := fields[0]
	// Some kernel/system processes are wrapped in brackets, e.g. "(mdworker)".
	exe = strings.TrimPrefix(exe, "(")
	exe = strings.TrimSuffix(exe, ")")
	return filepath.Base(exe)
}

var mlxHint = regexp.MustCompile(`(?i)mlx[_.-](lm|vlm|whisper)|mlx_lm|mlx_vlm|\bmlx\b`)

// classify tags a process by name and command line. It returns an empty
// Kind for anything gpuwho does not recognize.
func classify(name, command string) (model.ProcKind, string) {
	lname := strings.ToLower(name)

	switch {
	case lname == "ollama":
		return model.KindOllama, ""
	case lname == "llama-server":
		note := ""
		if strings.Contains(command, "/ollama/") || strings.Contains(command, "lib/ollama") {
			note = "ollama's vendored llama.cpp runner, not a standalone llama-server"
		}
		return model.KindLlamaServer, note
	case strings.Contains(lname, "python") && mlxHint.MatchString(command):
		return model.KindMLX, ""
	default:
		return "", ""
	}
}

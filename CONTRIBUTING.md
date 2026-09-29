# Contributing

## Setup

```sh
git clone https://github.com/Arthur031221/gpuwho
cd gpuwho
go build ./...
go test ./...
```

Go 1.27 or newer. macOS to actually run it (it shells out to `ioreg` and `ps`), any OS to build and run the unit tests, which use fixtures instead of the live system.

## Before sending a pull request

- `gofmt -l .` must print nothing.
- `go vet ./...` must be clean.
- `go test ./...` must pass.
- New parsing logic (ioreg, powermetrics, ps, Ollama, llama-server) needs a fixture-based test next to the existing ones. Do not add a test that depends on a live model being loaded or on root.

## Project layout

- `internal/gpuinfo`: system GPU utilization from ioreg, no root needed.
- `internal/power`: optional ANE and GPU power from powermetrics, needs root.
- `internal/procs`: process listing and classification (ollama, llama-server, MLX).
- `internal/ollamaps`, `internal/llamametrics`: HTTP clients for Ollama and llama-server.
- `internal/snapshot`: combines the samplers into one snapshot.
- `internal/render`, `internal/tui`: the plain-text and bubbletea views.
- `probe/`: the throwaway Swift probe used to check the proc_pid_rusage technical bet. Kept for reference, not built by the main tool.

## What gpuwho will not do

Per-process GPU time attribution through a public macOS API does not exist as of macOS 26.6 (see README, "How it works"). Do not add a feature that fakes a per-process GPU percentage. System-wide numbers plus honest process highlighting is the ceiling until Apple ships a real counter.

## Reporting a security issue

Open a regular issue. There is no separate disclosure process for a project this size.

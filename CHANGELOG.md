# Changelog

## 0.1.0

Initial release.

- Live TUI (bubbletea) showing system GPU utilization (ioreg, no root) and a sortable process table.
- `--json` and `--once` output for scripting and CI.
- Process highlighting for `ollama`, `llama-server` and python MLX processes, with a note when a `llama-server` row is Ollama's vendored runner rather than a standalone one.
- Model attribution from Ollama's `/api/ps` when exactly one model is loaded.
- Tokens per second from a llama-server instance's `/metrics` endpoint (`--llama-server-url`), when it was started with `--metrics`.
- Optional ANE and GPU power draw from `powermetrics` (`--powermetrics`, needs root).
- goreleaser config, Homebrew formula template, CI, demo tape.
- Documented technical finding: macOS 26.6 has no per-process GPU time field in `proc_pid_rusage` (checked `rusage_info_v4` through `v6`, public and private headers). See README, "How it works".

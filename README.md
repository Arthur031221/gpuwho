# gpuwho

Activity Monitor's GPU column for local LLM people. A live terminal view of which processes on your Mac are using the GPU, with ollama, llama-server and MLX processes named and highlighted.

On this machine (Apple M5, 10-core GPU, macOS 26.6), system GPU utilization measured through gpuwho went from 7% idle to 96% within about two seconds of sending one prompt to a locally running `qwen3:1.7b` model in Ollama, no root privileges required.[^1]

[![CI](https://github.com/Arthur031221/gpuwho/actions/workflows/ci.yml/badge.svg)](https://github.com/Arthur031221/gpuwho/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/tag/Arthur031221/gpuwho)](https://github.com/Arthur031221/gpuwho/releases)

![gpuwho demo](demo/demo.gif)

## Why

Apple never shipped a GPU column to `top` or a per-process GPU meter to Activity Monitor. If you run local models you end up alt-tabbing to Activity Monitor's separate GPU History window, which shows one system-wide graph and no process names. `asitop` used to fill this gap and stopped working after Sonoma broke its `powermetrics` parsing (its last commit was April 2024, 66 open issues). `mactop` is alive but tells you the GPU is busy, not which of your five terminal tabs running `ollama`, `llama-server`, or an MLX script is the one doing it. Nobody has fixed either of these for a Mac full of local LLM tools.

## Install

```sh
brew install --formula Formula/gpuwho.rb   # from a local clone until the tap is published
```

Or build from source:

```sh
git clone https://github.com/Arthur031221/gpuwho
cd gpuwho
go build -o gpuwho ./cmd/gpuwho
```

Go 1.27 or newer, macOS on Apple Silicon.

## Quick start

```sh
./gpuwho
```

Opens a live table: system GPU percent at the top, then every process on the machine, with `ollama`, `llama-server` and MLX rows sorted to the top and marked with `*`. Press `c`, `r`, `p`, `n` or `k` to sort by CPU, RSS, PID, name or kind. Press `q` to quit.

```sh
./gpuwho --once      # one table, no TUI, for scripts and screenshots
./gpuwho --json       # one JSON snapshot, for piping into jq
```

## How it works

The technical bet behind gpuwho was `proc_pid_rusage`, which on Linux-adjacent Darwin folklore is rumored to expose a per-process GPU time field. It does not, at least not on macOS 26.6.

I checked this directly before writing any product code. `rusage_info_v4`, `v5` and `v6` (`RUSAGE_INFO_CURRENT` on this SDK) are fully declared in `/usr/include/sys/resource.h` on the MacOSX26.5 and MacOSX27.0 SDKs, plus the private `Kernel.framework` header that ships with the Command Line Tools. None of the three versions has a `ri_gpu_time_ns` field or anything like it. `v6` adds `ri_neural_footprint`, `ri_lifetime_max_neural_footprint` and `ri_interval_max_neural_footprint`, which are Apple Neural Engine *memory* counters in bytes, not time. A small Swift probe (`probe/rusage_probe.swift`, built with `swiftc`) confirms this live: it calls `proc_pid_rusage(pid, RUSAGE_INFO_V6, ...)` against Ollama's running `llama-server` process while it generates tokens, and every CPU, memory and energy field moves as expected, but there is no GPU field to read because the struct does not have one.[^2]

So gpuwho does not use cgo or a compiled helper for the shipped binary. It falls back to two sources, both checked live on this machine:

- **System GPU utilization**, from `ioreg -r -d 1 -c IOAccelerator`. The AGXAccelerator entry publishes a `PerformanceStatistics` dictionary with `Device Utilization %`, `Tiler Utilization %` and `Renderer Utilization %`, refreshed on every read. This needs no root and no `powermetrics`. It is the number Activity Monitor's GPU History graph itself is built from.
- **ANE and GPU power draw**, optionally, from `powermetrics --samplers gpu_power,ane_power`. This does need root. gpuwho tries `sudo -n` (which fails instantly instead of hanging on a password prompt) and shows a clear reason when it cannot get a sample, instead of pretending to have one. Pass `--powermetrics` to try, or run gpuwho itself under `sudo`.

Process identification comes from `ps -axo pid=,ppid=,pcpu=,rss=,command=`, matched by executable name and, for MLX, by command-line pattern (`python` plus `mlx_lm` or `mlx_vlm`). Ollama's runner is a real find worth calling out: it is not a separate binary, it is Ollama's own vendored copy of `llama.cpp`'s server, installed at `lib/ollama/llama-server` and literally named `llama-server` in the process table. gpuwho tags it as `llama-server` and adds a note that it is Ollama's runner, not a standalone one, since a standalone `llama-server` you started yourself shows up identically otherwise.

**The honest limit**: macOS has no public per-process GPU time or GPU percent counter, on this OS version or any other one I could find documentation for. gpuwho's GPU number is system-wide. What it adds over Activity Monitor and `mactop` is naming which of your processes are the local-inference ones currently alive, so when the system number jumps you know it is `llama-server` and not Safari's WebKit GPU process. It does not, and currently cannot, tell you that PID 22934 personally accounts for 61 of that 88 percent.

### Measured: one process the system can name, one it cannot split

While building this, another build on the same shared machine was running a genuine `qwen3:4b` generation in Ollama (PID 22934). gpuwho sampled every 3 to 4 seconds for 15 seconds while it ran:[^3]

| time | GPU % (ioreg) | known process | its CPU% |
|---|---|---|---|
| 02:09:04 | 89 | llama-server (qwen3:4b) | 102.3 |
| 02:09:08 | 90 | llama-server (qwen3:4b) | 120.0 |
| 02:09:11 | 88 | llama-server (qwen3:4b) | 150.4 |
| 02:09:15 | 86 | llama-server (qwen3:4b) | 136.0 |

With one known GPU-using process alive, it is reasonable to guess the 86 to 90 percent belongs mostly to it. gpuwho does not print that guess as a number, because with two known processes running at once there is no macOS API that tells you how the system total splits between them, and a tool that fabricates a split is worse than one that says so. This is the fallback the brief asked for: system totals plus honest process highlighting, stated plainly, not a fake per-process percent.

### Tokens per second

Ollama's `/api/ps` (checked against Ollama 0.34.4 on this machine) reports the loaded model's name, size and VRAM use, but not a live tokens/second figure, so gpuwho uses it only for model name attribution, shown in the `MODEL` column when exactly one model is loaded.

`llama-server` started with `--metrics` exposes a Prometheus endpoint at `/metrics` with `llamacpp:predicted_tokens_seconds`. Pass `--llama-server-url http://localhost:8080` (or wherever yours listens) and gpuwho reads it into the `TOK/S` column. This was implemented against llama.cpp's documented metric names and covered by a fixture test. It was not exercised against a live standalone `llama-server` on this machine, since only Ollama's internal one (which Ollama does not start with `--metrics`) was running during the build. Ollama rows never get a `TOK/S` value for this reason.

## Comparison

| Tool | Per-process GPU | ANE | Sudo needed | Last push | Notes |
|---|---|---|---|---|---|
| **gpuwho** | Highlights known LLM processes by name, system % only | Optional, via powermetrics | No (yes only for ANE) | this release | Apple Silicon only |
| [tlkh/asitop](https://github.com/tlkh/asitop) | No, system-wide only | Yes | Yes | 2024-04-18 | 4,645 stars, 66 open issues, stopped working on recent macOS |
| [context-labs/mactop](https://github.com/context-labs/mactop) | No, system-wide only | Yes | Yes | 2025-12-21 | 2,340 stars, actively maintained fork |
| [metaspartan/mactop](https://github.com/metaspartan/mactop) | No, system-wide only | Yes | Yes | active | 1,653 stars, the original mactop, also system-wide |
| [Syllo/nvtop](https://github.com/Syllo/nvtop) | Yes, real per-process VRAM and compute | No (not applicable) | No | 2026-09-27 | 11,034 stars, NVIDIA/AMD/Intel Linux only, no Apple Silicon support. What per-process GPU monitoring looks like when the vendor exposes it, which Apple does not |

None of the three Apple Silicon tools break GPU usage out by process. gpuwho does not either, for the reason explained above, but it does tell you which of your processes are the ones worth suspecting.

## Command reference

```
gpuwho [flags]

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
```

In the TUI: `c` `r` `p` `n` `k` sort by CPU, RSS, PID, name or kind. `q` quits.

## Limits and FAQ

- **No per-process GPU percent.** Explained above. If Apple ships `ri_gpu_time_ns` or an equivalent, this is the first thing gpuwho will add.
- **Apple Silicon only.** Intel Macs do not have an AGXAccelerator entry in ioreg, so the GPU reading will fail with a clear "unavailable" message. gpuwho refuses to run on non-macOS platforms outright.
- **ANE needs root.** Without it, the ANE line shows "unavailable" and says why, rather than a fake zero.
- **Tokens/sec is opt-in and narrow.** Ollama does not expose it at all through its API today. llama-server does, only with `--metrics`, and only if you tell gpuwho the URL.
- **This is a monitor, not a profiler.** It samples on an interval. It will not catch a GPU spike shorter than your `--interval`.

## Related projects

- [gpuwait](https://github.com/Arthur031221/gpuwait): gpuwho shows which process is on the GPU right now. gpuwait shows how much of a serving window the GPU sits idle.
- [llm-doctor](https://github.com/Arthur031221/llm-doctor): Diagnoses the local model store that the processes gpuwho lists are usually reading from.
- [mlxtrace](https://github.com/Arthur031221/mlxtrace): Also samples GPU power on Apple Silicon, but per training step inside one MLX run instead of system wide.

## Contributing and license

See [CONTRIBUTING.md](CONTRIBUTING.md). MIT licensed, see [LICENSE](LICENSE).

[^1]: Measured with `gpuwho --json`, GPU reading from `ioreg -r -d 1 -c IOAccelerator`, `Device Utilization %` field. Idle sample and post-prompt sample taken about 2 seconds apart, `qwen3:1.7b` served by Ollama 0.34.4, Apple M5, macOS 26.6 (build 25G72), 2026-09-30, n=1.
[^2]: Checked against `/Library/Developer/CommandLineTools/SDKs/MacOSX26.5.sdk/usr/include/sys/resource.h`, the `MacOSX27.0.sdk` equivalent, and `Kernel.framework/Versions/A/Headers/sys/resource_private.h` on the same machine, 2026-09-30. Probe source in `probe/rusage_probe.swift`.
[^3]: Samples taken with `gpuwho --once --json` every 3 to 4 seconds, 2026-09-30, Apple M5, macOS 26.6, while another local build's Ollama session ran `qwen3:4b`. Raw PIDs and CPU percent from the same snapshots.

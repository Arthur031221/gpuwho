// Package model holds the shared data types gpuwho passes between its
// samplers, its renderers (TUI and JSON) and its tests.
package model

import "time"

// ProcKind marks a process as belonging to a known local inference
// runtime. An empty ProcKind means gpuwho does not recognize the process.
type ProcKind string

const (
	KindOllama      ProcKind = "ollama"
	KindLlamaServer ProcKind = "llama-server"
	KindMLX         ProcKind = "mlx"
)

// ProcessRow is one row of the process table.
type ProcessRow struct {
	PID          int      `json:"pid"`
	PPID         int      `json:"ppid"`
	Name         string   `json:"name"`
	Command      string   `json:"command"`
	CPUPercent   float64  `json:"cpu_percent"`
	RSSBytes     int64    `json:"rss_bytes"`
	Kind         ProcKind `json:"kind,omitempty"`
	Note         string   `json:"note,omitempty"`
	Model        string   `json:"model,omitempty"`
	TokensPerSec *float64 `json:"tokens_per_sec,omitempty"`
}

// GPUStat is a system-wide GPU utilization reading.
//
// macOS has no public per-process GPU time API (see README, "How it
// works"), so this number is system-wide. gpuwho pairs it with process
// highlighting instead of pretending to split it per process.
type GPUStat struct {
	Available       bool   `json:"available"`
	UtilPercent     int    `json:"util_percent"`
	TilerPercent    int    `json:"tiler_percent"`
	RendererPercent int    `json:"renderer_percent"`
	CoreCount       int    `json:"core_count"`
	ChipModel       string `json:"chip_model"`
	Source          string `json:"source"`
	Error           string `json:"error,omitempty"`
}

// PowerStat is an optional, root-only reading from powermetrics. It
// carries Apple Neural Engine and GPU power draw, which ioreg does not
// expose.
type PowerStat struct {
	Available   bool    `json:"available"`
	ANEMilliW   float64 `json:"ane_milliwatts,omitempty"`
	GPUMilliW   float64 `json:"gpu_milliwatts,omitempty"`
	Source      string  `json:"source,omitempty"`
	Unavailable string  `json:"unavailable_reason,omitempty"`
}

// Snapshot is one full sample: system GPU state plus the process table.
type Snapshot struct {
	Time      time.Time    `json:"time"`
	GPU       GPUStat      `json:"gpu"`
	Power     PowerStat    `json:"power"`
	Processes []ProcessRow `json:"processes"`
}

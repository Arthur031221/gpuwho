// Package gpuinfo samples system-wide GPU utilization on Apple Silicon
// through ioreg, which needs no root privileges.
//
// gpuwho's original technical bet was per-process GPU time through
// proc_pid_rusage (ri_gpu_time_ns). That field does not exist on macOS
// 26.6 (see probe/rusage_probe.swift and README.md "How it works" for
// the measurement). ioreg's AGXAccelerator PerformanceStatistics block
// is the fallback: it gives a live, system-wide "Device Utilization %"
// without sudo, refreshed on every read.
package gpuinfo

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/Arthur031221/gpuwho/internal/model"
)

var (
	reDeviceUtil   = regexp.MustCompile(`"Device Utilization %"\s*=\s*(\d+)`)
	reTilerUtil    = regexp.MustCompile(`"Tiler Utilization %"\s*=\s*(\d+)`)
	reRendererUtil = regexp.MustCompile(`"Renderer Utilization %"\s*=\s*(\d+)`)
	reCoreCount    = regexp.MustCompile(`"gpu-core-count"\s*=\s*(\d+)`)
	reChipModel    = regexp.MustCompile(`"model"\s*=\s*"([^"]+)"`)
)

// Sample shells out to ioreg and parses the result. It never requires
// root.
func Sample(ctx context.Context) (model.GPUStat, error) {
	cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(cctx, "ioreg", "-r", "-d", "1", "-c", "IOAccelerator").Output()
	if err != nil {
		return model.GPUStat{Available: false, Source: "ioreg", Error: err.Error()}, err
	}
	return ParseIoreg(string(out))
}

// ParseIoreg extracts the fields gpuwho needs from `ioreg -r -d 1 -c
// IOAccelerator` output. It is a pure function so it can be tested
// against a captured fixture without shelling out.
func ParseIoreg(output string) (model.GPUStat, error) {
	stat := model.GPUStat{Source: "ioreg"}

	m := reDeviceUtil.FindStringSubmatch(output)
	if m == nil {
		return stat, fmt.Errorf("gpuinfo: no \"Device Utilization %%\" field in ioreg output (no AGXAccelerator on this Mac, or Apple changed the key)")
	}
	stat.UtilPercent, _ = strconv.Atoi(m[1])
	stat.Available = true

	if m := reTilerUtil.FindStringSubmatch(output); m != nil {
		stat.TilerPercent, _ = strconv.Atoi(m[1])
	}
	if m := reRendererUtil.FindStringSubmatch(output); m != nil {
		stat.RendererPercent, _ = strconv.Atoi(m[1])
	}
	if m := reCoreCount.FindStringSubmatch(output); m != nil {
		stat.CoreCount, _ = strconv.Atoi(m[1])
	}
	if m := reChipModel.FindStringSubmatch(output); m != nil {
		stat.ChipModel = m[1]
	}
	return stat, nil
}

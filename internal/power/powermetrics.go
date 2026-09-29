// Package power reads Apple Neural Engine and GPU power draw from
// powermetrics. Unlike gpuinfo's ioreg sampler, powermetrics needs root,
// so this sampler is optional and fails soft.
package power

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Arthur031221/gpuwho/internal/model"
)

var (
	reANEPower = regexp.MustCompile(`ANE Power:\s*([\d.]+)\s*mW`)
	reGPUPower = regexp.MustCompile(`GPU Power:\s*([\d.]+)\s*mW`)
)

// Sample runs a single powermetrics sample. If the process is not root
// and cannot sudo without a password prompt, it returns immediately with
// Available=false and a reason instead of blocking on a prompt.
func Sample(ctx context.Context, sampleMillis int) (model.PowerStat, error) {
	stat := model.PowerStat{Source: "powermetrics"}

	args := []string{"powermetrics", "-n", "1", "-i", strconv.Itoa(sampleMillis), "--samplers", "gpu_power,ane_power"}
	var name string
	var cmdArgs []string
	if os.Geteuid() == 0 {
		name = args[0]
		cmdArgs = args[1:]
	} else {
		name = "sudo"
		cmdArgs = append([]string{"-n"}, args...)
	}

	cctx, cancel := context.WithTimeout(ctx, time.Duration(sampleMillis+2000)*time.Millisecond)
	defer cancel()
	out, err := exec.CommandContext(cctx, name, cmdArgs...).CombinedOutput()
	if err != nil {
		reason := "requires root: run gpuwho with sudo, or add a NOPASSWD sudoers rule for powermetrics (see README)"
		if strings.Contains(string(out), "password") || strings.Contains(string(out), "a password is required") {
			reason = "sudo needs a password in this shell: run gpuwho with sudo, or add a NOPASSWD sudoers rule for powermetrics (see README)"
		}
		stat.Unavailable = reason
		return stat, err
	}
	return ParsePowermetrics(string(out))
}

// ParsePowermetrics is a pure parser over powermetrics text output so it
// can be tested with a fixture instead of a live root session.
func ParsePowermetrics(out string) (model.PowerStat, error) {
	stat := model.PowerStat{Source: "powermetrics"}
	found := false
	if m := reANEPower.FindStringSubmatch(out); m != nil {
		stat.ANEMilliW, _ = strconv.ParseFloat(m[1], 64)
		found = true
	}
	if m := reGPUPower.FindStringSubmatch(out); m != nil {
		stat.GPUMilliW, _ = strconv.ParseFloat(m[1], 64)
		found = true
	}
	stat.Available = found
	if !found {
		stat.Unavailable = "powermetrics ran but did not report ANE or GPU power lines"
	}
	return stat, nil
}

package power

import "testing"

const fixture = `Machine model: Mac15,1
OS version: 26.6
Boot arguments:
Boot time: Mon Sep 29 00:00:00 2026

*** Sampled system activity (Tue Sep 29 01:56:00 2026 -0700) (1000.02ms elapsed) ***

**** GPU usage ****

GPU HW active residency:  93.40% (600MHz: .5% 828MHz: 1% ...)
GPU Power: 4211 mW

**** ANE usage ****

ANE Power: 128 mW
`

func TestParsePowermetrics(t *testing.T) {
	stat, err := ParsePowermetrics(fixture)
	if err != nil {
		t.Fatalf("ParsePowermetrics: %v", err)
	}
	if !stat.Available {
		t.Fatal("expected Available=true")
	}
	if stat.GPUMilliW != 4211 {
		t.Errorf("GPUMilliW = %v, want 4211", stat.GPUMilliW)
	}
	if stat.ANEMilliW != 128 {
		t.Errorf("ANEMilliW = %v, want 128", stat.ANEMilliW)
	}
}

func TestParsePowermetricsMissing(t *testing.T) {
	stat, err := ParsePowermetrics("nothing useful here")
	if err != nil {
		t.Fatalf("ParsePowermetrics: %v", err)
	}
	if stat.Available {
		t.Fatal("expected Available=false when no power lines are present")
	}
	if stat.Unavailable == "" {
		t.Error("expected a reason when unavailable")
	}
}

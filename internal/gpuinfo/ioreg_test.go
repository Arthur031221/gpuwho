package gpuinfo

import (
	"os"
	"testing"
)

func TestParseIoregFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/ioreg_sample.txt")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	stat, err := ParseIoreg(string(data))
	if err != nil {
		t.Fatalf("ParseIoreg: %v", err)
	}
	if !stat.Available {
		t.Fatal("expected Available=true")
	}
	if stat.ChipModel != "Apple M5" {
		t.Errorf("ChipModel = %q, want Apple M5", stat.ChipModel)
	}
	if stat.CoreCount != 10 {
		t.Errorf("CoreCount = %d, want 10", stat.CoreCount)
	}
	if stat.UtilPercent != 94 {
		t.Errorf("UtilPercent = %d, want 94", stat.UtilPercent)
	}
}

func TestParseIoregEmpty(t *testing.T) {
	_, err := ParseIoreg("no accelerator info here")
	if err == nil {
		t.Fatal("expected an error for output with no Device Utilization field")
	}
}

func TestParseIoregSpacedForm(t *testing.T) {
	// Some ioreg builds print key = value with spaces around '='.
	stat, err := ParseIoreg(`"Device Utilization %" = 42, "Tiler Utilization %" = 3, "Renderer Utilization %" = 5, "gpu-core-count" = 10, "model" = "Apple M5"`)
	if err != nil {
		t.Fatalf("ParseIoreg: %v", err)
	}
	if stat.UtilPercent != 42 {
		t.Errorf("UtilPercent = %d, want 42", stat.UtilPercent)
	}
}

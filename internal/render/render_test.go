package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Arthur031221/gpuwho/internal/model"
)

func TestHeaderAvailable(t *testing.T) {
	snap := model.Snapshot{
		GPU: model.GPUStat{Available: true, UtilPercent: 88, ChipModel: "Apple M5", CoreCount: 10},
	}
	out := Header(snap)
	if !strings.Contains(out, "88%") || !strings.Contains(out, "Apple M5") {
		t.Errorf("header missing expected content: %q", out)
	}
}

func TestHeaderUnavailable(t *testing.T) {
	snap := model.Snapshot{GPU: model.GPUStat{Available: false, Error: "boom"}}
	out := Header(snap)
	if !strings.Contains(out, "unavailable") || !strings.Contains(out, "boom") {
		t.Errorf("header missing unavailable reason: %q", out)
	}
}

func TestTable(t *testing.T) {
	toks := 33.4
	rows := []model.ProcessRow{
		{PID: 1, Name: "llama-server", Kind: model.KindLlamaServer, CPUPercent: 12.3, RSSBytes: 2 * 1024 * 1024 * 1024, Model: "qwen3:4b", TokensPerSec: &toks},
		{PID: 2, Name: "Finder", CPUPercent: 0.1, RSSBytes: 1024},
	}
	var buf bytes.Buffer
	Table(&buf, rows)
	out := buf.String()
	if !strings.Contains(out, "llama-server") || !strings.Contains(out, "qwen3:4b") || !strings.Contains(out, "33.4") {
		t.Errorf("table missing expected content: %q", out)
	}
	if !strings.Contains(out, "Finder") {
		t.Errorf("table missing unknown-kind row: %q", out)
	}
}

func TestHumanBytes(t *testing.T) {
	cases := map[int64]string{
		500:                    "500B",
		2048:                   "2.0KiB",
		3 * 1024 * 1024:        "3.0MiB",
		5 * 1024 * 1024 * 1024: "5.0GiB",
	}
	for in, want := range cases {
		if got := humanBytes(in); got != want {
			t.Errorf("humanBytes(%d) = %q, want %q", in, got, want)
		}
	}
}

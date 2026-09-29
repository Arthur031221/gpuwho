package snapshot

import (
	"testing"

	"github.com/Arthur031221/gpuwho/internal/model"
	"github.com/Arthur031221/gpuwho/internal/ollamaps"
)

func TestAttachOllamaModelSingleModel(t *testing.T) {
	rows := []model.ProcessRow{
		{PID: 1, Kind: model.KindLlamaServer},
		{PID: 2, Kind: model.KindOllama},
		{PID: 3},
	}
	attachOllamaModel(rows, []ollamaps.Model{{Name: "qwen3:4b"}})
	if rows[0].Model != "qwen3:4b" {
		t.Errorf("llama-server row Model = %q, want qwen3:4b", rows[0].Model)
	}
	if rows[1].Model != "" {
		t.Errorf("ollama daemon row Model = %q, want empty", rows[1].Model)
	}
}

func TestAttachOllamaModelAmbiguous(t *testing.T) {
	rows := []model.ProcessRow{{PID: 1, Kind: model.KindLlamaServer}}
	attachOllamaModel(rows, []ollamaps.Model{{Name: "a"}, {Name: "b"}})
	if rows[0].Model != "" {
		t.Errorf("Model = %q, want empty when more than one model is loaded", rows[0].Model)
	}
}

func TestAttachTokensPerSecSkipsOllamaRunner(t *testing.T) {
	rows := []model.ProcessRow{
		{PID: 1, Kind: model.KindLlamaServer, Note: "ollama's vendored llama.cpp runner"},
		{PID: 2, Kind: model.KindLlamaServer},
	}
	attachTokensPerSec(rows, 42.0)
	if rows[0].TokensPerSec != nil {
		t.Error("ollama runner row should not get a standalone llama-server tokens/sec reading")
	}
	if rows[1].TokensPerSec == nil || *rows[1].TokensPerSec != 42.0 {
		t.Errorf("standalone llama-server row TokensPerSec = %v, want 42.0", rows[1].TokensPerSec)
	}
}

func TestSortRelevancePutsKnownKindsFirst(t *testing.T) {
	rows := []model.ProcessRow{
		{PID: 1, CPUPercent: 99},
		{PID: 2, Kind: model.KindOllama, CPUPercent: 1},
	}
	sortRelevance(rows)
	if rows[0].PID != 2 {
		t.Errorf("expected known-kind row first regardless of CPU, got PID %d first", rows[0].PID)
	}
}

func TestSortByCPU(t *testing.T) {
	rows := []model.ProcessRow{{PID: 1, CPUPercent: 1}, {PID: 2, CPUPercent: 50}}
	SortBy(rows, "cpu")
	if rows[0].PID != 2 {
		t.Errorf("expected PID 2 first (higher CPU), got %d", rows[0].PID)
	}
}

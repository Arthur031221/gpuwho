package procs

import (
	"testing"

	"github.com/Arthur031221/gpuwho/internal/model"
)

const fixture = `  18897     1  12.5  1812104 /opt/homebrew/Cellar/ollama/0.34.4/libexec/lib/ollama/llama-server --model /Users/x/.ollama/models/blobs/sha256-abc --port 52433
  62319     1   0.1    21024 ollama serve
   4210 18897   0.0     3200 /usr/bin/llama-server --port 8080 --metrics
   9001     1  88.0   512000 /opt/homebrew/opt/python@3.12/bin/python3.12 -m mlx_lm.server --model mlx-community/Qwen3-4B
   1200     1   0.2    12000 /System/Library/CoreServices/Finder.app/Contents/MacOS/Finder
`

func TestParsePS(t *testing.T) {
	rows := ParsePS(fixture)
	if len(rows) != 5 {
		t.Fatalf("got %d rows, want 5", len(rows))
	}

	byPID := map[int]model.ProcessRow{}
	for _, r := range rows {
		byPID[r.PID] = r
	}

	if got := byPID[18897]; got.Kind != model.KindLlamaServer || got.Note == "" {
		t.Errorf("pid 18897 = %+v, want KindLlamaServer with an ollama note", got)
	}
	if got := byPID[62319]; got.Kind != model.KindOllama {
		t.Errorf("pid 62319 kind = %q, want ollama", got.Kind)
	}
	if got := byPID[4210]; got.Kind != model.KindLlamaServer || got.Note != "" {
		t.Errorf("pid 4210 = %+v, want KindLlamaServer with no ollama note", got)
	}
	if got := byPID[9001]; got.Kind != model.KindMLX {
		t.Errorf("pid 9001 kind = %q, want mlx", got.Kind)
	}
	if got := byPID[1200]; got.Kind != "" {
		t.Errorf("pid 1200 kind = %q, want empty (Finder is not a known runtime)", got.Kind)
	}
	if byPID[18897].RSSBytes != 1812104*1024 {
		t.Errorf("RSSBytes = %d, want %d", byPID[18897].RSSBytes, 1812104*1024)
	}
}

func TestParsePSIgnoresBlankLines(t *testing.T) {
	rows := ParsePS("\n\n  1 1 0.0 100 /bin/launchd\n\n")
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
}

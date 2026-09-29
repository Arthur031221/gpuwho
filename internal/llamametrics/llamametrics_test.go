package llamametrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const fixture = `# HELP llamacpp:prompt_tokens_total Number of prompt tokens processed.
# TYPE llamacpp:prompt_tokens_total counter
llamacpp:prompt_tokens_total 512
# HELP llamacpp:prompt_tokens_seconds Average prompt throughput in tokens/s.
# TYPE llamacpp:prompt_tokens_seconds gauge
llamacpp:prompt_tokens_seconds 812.500000
# HELP llamacpp:predicted_tokens_total Number of generated tokens.
# TYPE llamacpp:predicted_tokens_total counter
llamacpp:predicted_tokens_total 128
# HELP llamacpp:predicted_tokens_seconds Average generation throughput in tokens/s.
# TYPE llamacpp:predicted_tokens_seconds gauge
llamacpp:predicted_tokens_seconds 33.450000
`

func TestParse(t *testing.T) {
	m, err := Parse(strings.NewReader(fixture))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !m.HasPredicted || m.PredictedTokensPerSec != 33.45 {
		t.Errorf("PredictedTokensPerSec = %v (has=%v), want 33.45", m.PredictedTokensPerSec, m.HasPredicted)
	}
	if !m.HasPrompt || m.PromptTokensPerSec != 812.5 {
		t.Errorf("PromptTokensPerSec = %v (has=%v), want 812.5", m.PromptTokensPerSec, m.HasPrompt)
	}
}

func TestParseNoMatchingGauges(t *testing.T) {
	m, err := Parse(strings.NewReader("# just comments\nsome_other_metric 1\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if m.HasPrompt || m.HasPredicted {
		t.Errorf("expected no gauges found, got %+v", m)
	}
}

func TestFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metrics" {
			t.Errorf("path = %q, want /metrics", r.URL.Path)
		}
		w.Write([]byte(fixture))
	}))
	defer srv.Close()

	m, err := Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if m.PredictedTokensPerSec != 33.45 {
		t.Errorf("PredictedTokensPerSec = %v, want 33.45", m.PredictedTokensPerSec)
	}
}

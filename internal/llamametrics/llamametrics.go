// Package llamametrics reads tokens-per-second from a llama.cpp
// llama-server instance started with --metrics, which exposes a
// Prometheus text endpoint at /metrics with gauges named
// llamacpp:prompt_tokens_seconds and llamacpp:predicted_tokens_seconds.
//
// This was not exercised against a live standalone llama-server while
// building gpuwho on this machine (only Ollama's vendored llama-server
// was running, and Ollama does not pass --metrics to it), so treat this
// as implemented against llama.cpp's documented metric names and covered
// by a fixture test, not a live measurement. It fails soft: if the
// endpoint is unreachable or the gauges are missing, gpuwho just shows
// no tokens/sec for that row.
package llamametrics

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Metrics struct {
	PromptTokensPerSec    float64
	PredictedTokensPerSec float64
	HasPrompt             bool
	HasPredicted          bool
}

// Fetch calls GET <baseURL>/metrics and parses the Prometheus text
// format. baseURL has no default: the caller must opt in with
// --llama-server-url since there is no fixed convention for the port.
func Fetch(ctx context.Context, baseURL string) (Metrics, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/metrics", nil)
	if err != nil {
		return Metrics{}, err
	}
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return Metrics{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Metrics{}, &statusError{resp.StatusCode}
	}
	return Parse(resp.Body)
}

// Parse reads Prometheus text-format exposition and pulls out the two
// gauges gpuwho cares about. It is a pure function over an io.Reader so
// tests can feed it a fixture file.
func Parse(r io.Reader) (Metrics, error) {
	var m Metrics
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		name, valStr := fields[0], fields[1]
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}
		switch name {
		case "llamacpp:prompt_tokens_seconds":
			m.PromptTokensPerSec = val
			m.HasPrompt = true
		case "llamacpp:predicted_tokens_seconds":
			m.PredictedTokensPerSec = val
			m.HasPredicted = true
		}
	}
	if err := scanner.Err(); err != nil {
		return m, err
	}
	return m, nil
}

type statusError struct{ code int }

func (e *statusError) Error() string { return http.StatusText(e.code) }

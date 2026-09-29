// Package ollamaps queries Ollama's /api/ps endpoint for the models it
// currently has loaded in memory. As of Ollama 0.34.4 this endpoint does
// not report tokens per second (only name, size, VRAM use and expiry),
// so gpuwho uses it for model attribution, not for throughput. See
// README.md "How it works" for what was checked.
package ollamaps

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Model struct {
	Name          string `json:"name"`
	Model         string `json:"model"`
	Size          int64  `json:"size"`
	SizeVRAM      int64  `json:"size_vram"`
	ExpiresAt     string `json:"expires_at"`
	ContextLength int    `json:"context_length"`
}

type psResponse struct {
	Models []Model `json:"models"`
}

// Fetch calls GET <baseURL>/api/ps and returns the loaded models. baseURL
// defaults to http://localhost:11434 when empty.
func Fetch(ctx context.Context, baseURL string) ([]Model, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/ps", nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &statusError{resp.StatusCode}
	}
	var parsed psResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed.Models, nil
}

type statusError struct{ code int }

func (e *statusError) Error() string { return http.StatusText(e.code) }

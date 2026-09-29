package ollamaps

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ps" {
			t.Errorf("path = %q, want /api/ps", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"models":[{"name":"qwen3:1.7b","model":"qwen3:1.7b","size":1873427823,"size_vram":1873427823,"expires_at":"2026-09-30T02:00:00Z","context_length":4096}]}`))
	}))
	defer srv.Close()

	models, err := Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("got %d models, want 1", len(models))
	}
	if models[0].Name != "qwen3:1.7b" {
		t.Errorf("Name = %q, want qwen3:1.7b", models[0].Name)
	}
	if models[0].SizeVRAM != 1873427823 {
		t.Errorf("SizeVRAM = %d", models[0].SizeVRAM)
	}
}

func TestFetchUnreachable(t *testing.T) {
	_, err := Fetch(context.Background(), "http://127.0.0.1:1")
	if err == nil {
		t.Fatal("expected an error for an unreachable server")
	}
}

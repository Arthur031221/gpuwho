package main

import (
	"bytes"
	"encoding/json"
	"runtime"
	"testing"
)

// captureRun runs the CLI with args and returns stdout, stderr and the
// exit code, using in-memory buffers so there is no pipe buffer to fill
// or drain.
func captureRun(t *testing.T, args []string) (stdout, stderr string, code int) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	code = run(args, &outBuf, &errBuf)
	return outBuf.String(), errBuf.String(), code
}

func TestRunOnce(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("gpuwho only supports macOS")
	}
	stdout, _, code := captureRun(t, []string{"--once", "--ollama-url", "http://127.0.0.1:1"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !bytes.Contains([]byte(stdout), []byte("GPU")) {
		t.Errorf("stdout missing GPU header: %q", stdout)
	}
}

func TestRunJSON(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("gpuwho only supports macOS")
	}
	stdout, _, code := captureRun(t, []string{"--json", "--ollama-url", "http://127.0.0.1:1"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout)
	}
	if _, ok := parsed["processes"]; !ok {
		t.Error("json output missing processes field")
	}
	if _, ok := parsed["gpu"]; !ok {
		t.Error("json output missing gpu field")
	}
}

func TestRunVersion(t *testing.T) {
	stdout, _, code := captureRun(t, []string{"--version"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !bytes.Contains([]byte(stdout), []byte("gpuwho")) {
		t.Errorf("stdout = %q, want it to mention gpuwho", stdout)
	}
}

func TestRunHelp(t *testing.T) {
	_, stderr, code := captureRun(t, []string{"--help"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !bytes.Contains([]byte(stderr), []byte("Usage")) {
		t.Errorf("stderr = %q, want usage text", stderr)
	}
}

func TestNormalizeSort(t *testing.T) {
	cases := map[string]string{
		"cpu":       "cpu",
		"rss":       "rss",
		"bogus":     "relevance",
		"":          "relevance",
		"relevance": "relevance",
	}
	for in, want := range cases {
		if got := normalizeSort(in); got != want {
			t.Errorf("normalizeSort(%q) = %q, want %q", in, got, want)
		}
	}
}

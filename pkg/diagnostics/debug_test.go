package diagnostics

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer

	fn()

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestSetVerboseAndInfo(t *testing.T) {
	SetVerbose(false)
	if IsVerbose() {
		t.Fatal("verbose should be disabled")
	}
	quiet := captureStdout(t, func() {
		Info("hidden %s", "message")
	})
	if quiet != "" {
		t.Fatalf("expected no output when verbose=false, got %q", quiet)
	}

	SetVerbose(true)
	defer SetVerbose(false)
	if !IsVerbose() {
		t.Fatal("verbose should be enabled")
	}
	output := captureStdout(t, func() {
		Info("visible %s", "message")
	})
	if !strings.Contains(output, "INFO -   visible message") {
		t.Fatalf("expected verbose info output, got %q", output)
	}
}

func TestDebugRunsOnlyWhenEnvIsSet(t *testing.T) {
	t.Setenv("DEBUG", "")
	called := false
	Debug(func() {
		called = true
	})
	if called {
		t.Fatal("debug callback should not run without DEBUG")
	}

	t.Setenv("DEBUG", "1")
	Debug(func() {
		called = true
	})
	if !called {
		t.Fatal("debug callback should run with DEBUG")
	}
}

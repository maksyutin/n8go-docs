package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureCommandStdout(t *testing.T, fn func()) string {
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

func writeCommandTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestVersionCommandRunPlainAndJSON(t *testing.T) {
	plain := captureCommandStdout(t, func() {
		if err := (&VersionCommand{}).Run(&CLI{}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(plain, "n8go-docs "+appVersion) {
		t.Fatalf("unexpected plain version output: %q", plain)
	}

	jsonOutput := captureCommandStdout(t, func() {
		if err := (&VersionCommand{}).Run(&CLI{JSON: true}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(jsonOutput, `"version": "`+appVersion+`"`) {
		t.Fatalf("unexpected JSON version output: %q", jsonOutput)
	}
}

func TestBuildCommandRunReturnsJSONError(t *testing.T) {
	missingConfig := filepath.Join(t.TempDir(), "missing.yaml")

	output := captureCommandStdout(t, func() {
		err := (&BuildCommand{}).Run(&CLI{Config: missingConfig, JSON: true})
		if err == nil {
			t.Fatal("expected missing config error")
		}
	})
	if !strings.Contains(output, `"status": "error"`) {
		t.Fatalf("expected JSON error output, got %q", output)
	}
}

func TestRunGeneratorUsesThemeDirAndBuildsSite(t *testing.T) {
	root := t.TempDir()
	docsDir := filepath.Join(root, "docs")
	siteDir := filepath.Join(root, "site")
	themeBaseDir := filepath.Join(root, "themes")
	themeDir := filepath.Join(themeBaseDir, "default")
	configPath := filepath.Join(root, "n8go-docs.yaml")

	writeCommandTestFile(t, filepath.Join(docsDir, "index.md"), "# Home")
	writeCommandTestFile(t, filepath.Join(themeDir, "theme.yaml"), `theme:
  name: Default
  version: 1.0.0
highlighting:
  style: github
`)
	writeCommandTestFile(t, filepath.Join(themeDir, "main.html"), `<!doctype html><main class="main-content">{{.Page.Body}}</main>`)
	writeCommandTestFile(t, configPath, "name: Docs\ndocs_dir: docs\nsite_dir: site\ntheme: default\n")

	t.Setenv("THEME_DIR", themeBaseDir)
	if err := runGenerator(configPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(siteDir, "index.html")); err != nil {
		t.Fatalf("generated index.html not found: %v", err)
	}
}

func TestFileServerWithCustom404(t *testing.T) {
	root := t.TempDir()
	writeCommandTestFile(t, filepath.Join(root, "file.txt"), "ok")

	handler := FileServerWithCustom404(http.Dir(root), 9080)

	okRequest := httptest.NewRequest(http.MethodGet, "/file.txt", nil)
	okResponse := httptest.NewRecorder()
	handler.ServeHTTP(okResponse, okRequest)
	if okResponse.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", okResponse.Code)
	}

	missingRequest := httptest.NewRequest(http.MethodGet, "/missing.html", nil)
	missingResponse := httptest.NewRecorder()
	handler.ServeHTTP(missingResponse, missingRequest)
	if missingResponse.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", missingResponse.Code)
	}
	if !strings.Contains(missingResponse.Body.String(), "404 - Not Found") {
		t.Fatalf("unexpected 404 body: %q", missingResponse.Body.String())
	}
}

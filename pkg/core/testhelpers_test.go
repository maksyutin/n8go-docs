package core

import (
	"os"
	"path/filepath"
	"testing"

	"n8go-docs/manifest"
)

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func testThemeManifest() manifest.ThemeManifest {
	return manifest.ThemeManifest{
		Name:    "Test theme",
		Version: "1.0.0",
		Highlighting: manifest.HighlightingConfig{
			Style: "github",
		},
	}
}

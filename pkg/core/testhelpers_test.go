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

func makeTestTheme(t *testing.T, root string) string {
	t.Helper()
	themeDir := filepath.Join(root, "theme")
	writeTestFile(t, filepath.Join(themeDir, "main.html"), `<!doctype html>
<html>
<body>
<nav>{% for nav_item in nav %}<a href="{{ nav_item.url|url }}" data-active="{{ nav_item.active }}">{{ nav_item.title }}</a>{% endfor %}</nav>
<main class="main-content">{{ page.content }}</main>
</body>
</html>`)
	writeTestFile(t, filepath.Join(themeDir, "css", "theme.css"), "body { color: #111; }\n")
	return themeDir
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

package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
<nav>{{range .Nav}}<a href="{{.Url}}" data-active="{{.Active}}">{{.Name}}</a>{{end}}</nav>
<main class="main-content">{{.Page.Body}}</main>
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

func TestGenerateDocumentationBuildsPagesSearchIndexAndStaticFiles(t *testing.T) {
	root := t.TempDir()
	inputDir := filepath.Join(root, "docs")
	outputDir := filepath.Join(root, "site")
	themeDir := makeTestTheme(t, root)

	writeTestFile(t, filepath.Join(inputDir, "index.md"), "# Home\n\n[Guide](guide/README.md)\n\n<script>alert(1)</script>")
	writeTestFile(t, filepath.Join(inputDir, "guide", "README.md"), "# Guide\n\nContent")
	writeTestFile(t, filepath.Join(inputDir, "draft.md"), "# Draft\n\nHidden")
	writeTestFile(t, filepath.Join(inputDir, "asset.txt"), "copied from docs")

	siteManifest := manifest.SiteManifest{
		Name:          "Docs",
		ThemeId:       "test",
		InputPath:     inputDir,
		OutputPath:    outputDir,
		DefaultSearch: true,
		ExcludeDocs:   []string{"draft.md"},
	}

	if err := GenerateDocumentation(siteManifest, testThemeManifest(), themeDir); err != nil {
		t.Fatal(err)
	}

	indexHTML, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(indexHTML), "<script>alert(1)</script>") {
		t.Fatalf("unsafe markdown HTML was rendered:\n%s", string(indexHTML))
	}
	if !strings.Contains(string(indexHTML), `href="guide/"`) {
		t.Fatalf("markdown link was not rewritten:\n%s", string(indexHTML))
	}

	if _, err := os.Stat(filepath.Join(outputDir, "guide", "index.html")); err != nil {
		t.Fatalf("README page was not generated as directory index: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "draft", "index.html")); !os.IsNotExist(err) {
		t.Fatalf("excluded draft page should not be generated, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "asset.txt")); err != nil {
		t.Fatalf("docs static file was not copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "css", "theme.css")); err != nil {
		t.Fatalf("theme static file was not copied: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "search", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var entries []SearchIndexEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 indexed pages, got %d: %#v", len(entries), entries)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Title, "Draft") || strings.Contains(entry.Content, "Hidden") {
			t.Fatalf("excluded page was indexed: %#v", entries)
		}
	}
}

func TestGenerateDocumentationReturnsErrorWhenNoPagesExist(t *testing.T) {
	root := t.TempDir()
	inputDir := filepath.Join(root, "docs")
	outputDir := filepath.Join(root, "site")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatal(err)
	}

	err := GenerateDocumentation(manifest.SiteManifest{
		Name:       "Docs",
		ThemeId:    "test",
		InputPath:  inputDir,
		OutputPath: outputDir,
	}, testThemeManifest(), makeTestTheme(t, root))
	if err == nil || !strings.Contains(err.Error(), "no pages found") {
		t.Fatalf("expected no pages error, got %v", err)
	}
}

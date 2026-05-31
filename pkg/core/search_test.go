package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"n8go-docs/manifest"
)

func TestSearchIndexSanitizesEntriesAndWritesJSONOnce(t *testing.T) {
	outputDir := t.TempDir()
	index := NewSearchIndex()

	index.Add(SearchIndexEntry{
		Title:   `<img src=x onerror=alert(1)>Unsafe title`,
		Url:     "javascript:alert(1)",
		Content: "<p>Hello</p><script>alert(2)</script>\nworld",
	})

	if err := index.Write(manifest.SiteManifest{OutputPath: outputDir}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "search", "index.json"))
	if err != nil {
		t.Fatal(err)
	}

	var entries []SearchIndexEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("search index is not a JSON array: %v\n%s", err, string(data))
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Title != "Unsafe title" {
		t.Fatalf("title was not sanitized: %q", entry.Title)
	}
	if entry.Url != "./javascript:alert(1)" {
		t.Fatalf("dangerous URL was not made relative: %q", entry.Url)
	}
	if entry.Content != "Hello world" {
		t.Fatalf("content was not normalized and sanitized: %q", entry.Content)
	}
}

func TestIndexPageContentReturnsSanitizedLimitedContent(t *testing.T) {
	ctx := &pageContext{
		Page: pageInfo{
			FileName: "index.md",
			Title:    "Home",
		},
		Url: ".",
	}

	entry, ok := indexPageContent(ctx, manifest.SiteManifest{SearchContentLimit: 5}, `
<html><body>
	<div class="main-content">
		<p>Hello <strong>world</strong></p>
		<script>alert(1)</script>
	</div>
</body></html>`)
	if !ok {
		t.Fatal("expected index entry")
	}
	if entry.Content != "Hello" {
		t.Fatalf("expected limited sanitized content, got %q", entry.Content)
	}
}

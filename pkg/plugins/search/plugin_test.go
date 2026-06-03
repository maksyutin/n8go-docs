package search_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"n8go-docs/core"
	"n8go-docs/plugins/search"
)

func TestSearchPluginSanitizesContentAndWritesJSON(t *testing.T) {
	outputDir := t.TempDir()
	p := search.New(0, "", outputDir)

	// Simulate build lifecycle: OnInit → OnPageRendered (unsafe HTML) → OnBuildComplete
	if err := p.OnInit(&core.BuildContext{OutputDir: outputDir}); err != nil {
		t.Fatal(err)
	}

	html := `<html><body><div class="main-content"><p>Hello <strong>world</strong></p><script>alert(1)</script></div></body></html>`
	if _, err := p.OnPageRendered(&core.PageResult{Title: "My Page", URL: "."}, html); err != nil {
		t.Fatal(err)
	}

	if err := p.OnBuildComplete(&core.BuildContext{OutputDir: outputDir}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "search", "index.json"))
	if err != nil {
		t.Fatal(err)
	}

	var entries []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("search index is not valid JSON: %v\n%s", err, string(data))
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Title != "My Page" {
		t.Fatalf("unexpected title: %q", entry.Title)
	}
	// script tag stripped, text nodes joined
	if entry.Content != "Hello world" {
		t.Fatalf("content was not normalized and sanitized: %q", entry.Content)
	}
}

func TestSearchPluginRespectsContentLimit(t *testing.T) {
	outputDir := t.TempDir()
	p := search.New(5, "", outputDir)

	_ = p.OnInit(&core.BuildContext{OutputDir: outputDir})

	html := `<html><body><div class="main-content"><p>Hello world</p></div></body></html>`
	if _, err := p.OnPageRendered(&core.PageResult{Title: "Home", URL: "."}, html); err != nil {
		t.Fatal(err)
	}

	_ = p.OnBuildComplete(&core.BuildContext{OutputDir: outputDir})

	data, _ := os.ReadFile(filepath.Join(outputDir, "search", "index.json"))
	var entries []struct {
		Content string `json:"content"`
	}
	_ = json.Unmarshal(data, &entries)

	if len([]rune(entries[0].Content)) > 5 {
		t.Fatalf("content limit not respected: %q", entries[0].Content)
	}
}

func TestSearchPluginPrefixesSiteURL(t *testing.T) {
	outputDir := t.TempDir()
	p := search.New(0, "https://example.com/docs/", outputDir)

	_ = p.OnInit(&core.BuildContext{OutputDir: outputDir})

	html := `<html><body><div class="main-content">text</div></body></html>`
	if _, err := p.OnPageRendered(&core.PageResult{Title: "Page", URL: "guide/"}, html); err != nil {
		t.Fatal(err)
	}

	_ = p.OnBuildComplete(&core.BuildContext{OutputDir: outputDir})

	data, _ := os.ReadFile(filepath.Join(outputDir, "search", "index.json"))
	var entries []struct {
		URL string `json:"url"`
	}
	_ = json.Unmarshal(data, &entries)

	want := "https://example.com/docs/guide/"
	if entries[0].URL != want {
		t.Fatalf("URL = %q, want %q", entries[0].URL, want)
	}
}

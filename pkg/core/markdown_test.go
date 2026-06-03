package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"n8go-docs/manifest"
)

func TestRenderMarkdownPageSanitizesUnsafeHTML(t *testing.T) {
	dir := t.TempDir()
	mdFile := filepath.Join(dir, "index.md")
	source := `# Safe title

<strong>allowed</strong>
<script>alert("xss")</script>
<img src="javascript:alert(1)" onerror="alert(2)">

` + "```go\nfmt.Println(\"ok\")\n```"

	if err := os.WriteFile(mdFile, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	page, err := renderMarkdownPage(mdFile, manifest.ThemeManifest{
		Highlighting: manifest.HighlightingConfig{Style: "github"},
	}, manifest.SiteManifest{})
	if err != nil {
		t.Fatal(err)
	}

	for _, blocked := range []string{"<script", "alert(\"xss\")", "javascript:alert", "onerror="} {
		if strings.Contains(page.Body, blocked) {
			t.Fatalf("rendered body contains unsafe fragment %q:\n%s", blocked, page.Body)
		}
	}
	if !strings.Contains(page.Body, "<strong>allowed</strong>") {
		t.Fatalf("expected safe inline HTML to be preserved:\n%s", page.Body)
	}
	if !strings.Contains(page.Body, "class=") {
		t.Fatalf("expected sanitizer to preserve generated highlighting classes:\n%s", page.Body)
	}
	if !strings.Contains(page.Body, `<div class="highlight"><pre`) {
		t.Fatalf("expected code blocks to use Bootstrap-style highlight wrapper:\n%s", page.Body)
	}
	if !strings.Contains(page.Body, `<code class="language-go" data-lang="go">`) {
		t.Fatalf("expected highlighted code block to preserve language attributes:\n%s", page.Body)
	}
}

func TestRenderMarkdownPagePreservesThemeCustomElements(t *testing.T) {
	dir := t.TempDir()
	mdFile := filepath.Join(dir, "components.md")
	source := `# Components

<n8go-alert type="info" message="Heads up"></n8go-alert>
`

	if err := os.WriteFile(mdFile, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	page, err := renderMarkdownPage(mdFile, manifest.ThemeManifest{
		Highlighting: manifest.HighlightingConfig{Style: "github"},
	}, manifest.SiteManifest{})
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"<n8go-alert", `type="info"`, `message="Heads up"`} {
		if !strings.Contains(page.Body, want) {
			t.Fatalf("expected rendered body to preserve theme custom element fragment %q:\n%s", want, page.Body)
		}
	}
}

func TestRenderMarkdownPageWrapsUnhighlightedCodeBlocks(t *testing.T) {
	dir := t.TempDir()
	mdFile := filepath.Join(dir, "themeing.md")
	source := "# Themeing\n\n```jinja2\n<!DOCTYPE html>\n<title>{{ page_title }}</title>\n```\n"

	if err := os.WriteFile(mdFile, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	page, err := renderMarkdownPage(mdFile, manifest.ThemeManifest{
		Highlighting: manifest.HighlightingConfig{Style: "github"},
	}, manifest.SiteManifest{})
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{`<div class="highlight"><pre><code class="language-jinja2" data-lang="jinja2">`, "&lt;!DOCTYPE html&gt;", "</code></pre></div>"} {
		if !strings.Contains(page.Body, want) {
			t.Fatalf("expected unhighlighted code block to preserve Bootstrap-style code markup %q:\n%s", want, page.Body)
		}
	}
}

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
}

// Package search provides a Pipeline plugin that builds a client-side search
// index (JSON) from rendered page HTML. Supports flexsearch and fuse.js.
package search

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"n8go-docs/core"
	"n8go-docs/diagnostics"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
)

type entry struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

// Plugin collects per-page search entries and writes search/index.json on build
// completion. Wire it up via pipeline.Register(search.New(site)).
type Plugin struct {
	contentLimit int
	siteURL      string
	outputDir    string
	entries      []entry
}

// New creates a search plugin.
//
//	contentLimit – max rune count per page (0 = unlimited).
//	siteURL      – absolute site URL for entry URLs; empty means relative.
//	outputDir    – site output directory.
func New(contentLimit int, siteURL, outputDir string) *Plugin {
	return &Plugin{
		contentLimit: contentLimit,
		siteURL:      siteURL,
		outputDir:    outputDir,
	}
}

func (p *Plugin) Name() string { return "search" }

func (p *Plugin) OnInit(_ *core.BuildContext) error {
	p.entries = p.entries[:0]
	return nil
}

func (p *Plugin) OnPageRendered(page *core.PageResult, html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		diagnostics.PrintError(err, "search: parse HTML for "+page.FilePath)
		return html, nil
	}

	content, _ := doc.Find(".main-content").Html()
	content = bluemonday.StrictPolicy().Sanitize(content)
	content = strings.Join(strings.Fields(content), " ")

	if p.contentLimit > 0 && len([]rune(content)) > p.contentLimit {
		content = string([]rune(content)[:p.contentLimit])
	}

	url := page.URL
	if p.siteURL != "" {
		url = strings.TrimRight(p.siteURL, "/") + "/" + strings.TrimLeft(page.URL, "/")
	}

	p.entries = append(p.entries, entry{
		Title:   page.Title,
		URL:     url,
		Content: content,
	})
	return html, nil
}

func (p *Plugin) OnBuildComplete(_ *core.BuildContext) error {
	indexPath := filepath.Join(p.outputDir, "search", "index.json")
	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(p.entries)
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath, data, 0o644)
}

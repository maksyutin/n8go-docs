// Package unsaved provides an EditorPlugin that injects an "unsaved changes"
// banner into preview HTML so the user knows they are viewing a draft.
package unsaved

import (
	"n8go-docs/core"
	"n8go-docs/editor"
	"net/http"
	"strings"
)

const bannerHTML = `<div style="position:fixed;top:0;left:0;right:0;z-index:9999;background:#f5a623;color:#fff;text-align:center;padding:6px 12px;font-family:sans-serif;font-size:13px;">
  ⚠ Preview — unsaved changes
</div>`

// Plugin injects a warning banner into preview renders and exposes a status
// route that returns whether there are pending edits.
type Plugin struct {
	dirty map[string]bool
}

// New creates the unsaved-changes plugin.
func New() *Plugin {
	return &Plugin{dirty: make(map[string]bool)}
}

// --- core.Plugin hooks -------------------------------------------------------

func (p *Plugin) Name() string { return "unsaved" }

// OnPreview injects the banner and records the file as dirty.
func (p *Plugin) OnPreview(filePath string, html string) (string, error) {
	p.dirty[filePath] = true
	// Inject banner after <body>
	return strings.Replace(html, "<body>", "<body>"+bannerHTML, 1), nil
}

// OnFileWrite clears the dirty flag when a file is saved.
func (p *Plugin) OnFileWrite(filePath string, content []byte) ([]byte, error) {
	delete(p.dirty, filePath)
	return content, nil
}

// --- editor.EditorPlugin routes ----------------------------------------------

func (p *Plugin) Routes() []editor.Route {
	return []editor.Route{
		{Pattern: "/api/unsaved", Handler: p.handleUnsaved},
	}
}

// handleUnsaved returns a JSON list of files with unsaved previews.
// GET /_editor/api/unsaved
func (p *Plugin) handleUnsaved(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	files := make([]string, 0, len(p.dirty))
	for f := range p.dirty {
		files = append(files, f)
	}
	w.Header().Set("Content-Type", "application/json")
	// Simple JSON array without import overhead
	_, _ = w.Write([]byte(`["` + strings.Join(files, `","`) + `"]`))
}

// Ensure Plugin satisfies all interfaces at compile time.
var _ core.Plugin = (*Plugin)(nil)
var _ core.EditorPreviewHook = (*Plugin)(nil)
var _ core.EditorFileWriteHook = (*Plugin)(nil)
var _ editor.EditorPlugin = (*Plugin)(nil)

// Package editor provides a pluggable HTTP API for editing documentation
// source files in a browser. It delegates all generation logic to a
// core.Pipeline so that the same rendering pipeline is used for both the
// static build and the live preview.
package editor

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"n8go-docs/core"
	"n8go-docs/diagnostics"
)

// Editor mounts the editor API under a given HTTP mux prefix.
// All heavy lifting is delegated to the Pipeline; the editor only handles
// HTTP plumbing and calls the registered plugin hooks.
type Editor struct {
	pipeline *core.Pipeline
	plugins  []EditorPlugin
}

// EditorPlugin extends the editor's HTTP behaviour. A plugin may handle
// custom API routes, inject toolbar buttons, or add WebSocket channels.
// All methods are optional — implement only the ones you need.
type EditorPlugin interface {
	// Name uniquely identifies the plugin.
	Name() string

	// Routes returns additional HTTP handlers to register.
	// Each entry is (pattern, handler) relative to the editor mount path.
	// Example: "/api/comments" → your handler.
	Routes() []Route
}

// Route pairs a URL pattern with its handler.
type Route struct {
	Pattern string
	Handler http.HandlerFunc
}

// New creates an Editor backed by the given Pipeline.
func New(pipeline *core.Pipeline) *Editor {
	return &Editor{pipeline: pipeline}
}

// Register adds an editor-level plugin (e.g. a comments side-panel).
func (e *Editor) Register(plugin EditorPlugin) {
	e.plugins = append(e.plugins, plugin)
}

// Mount registers all editor routes on mux under the given prefix.
// prefix should start with "/" and not end with "/", e.g. "/_editor".
func (e *Editor) Mount(mux *http.ServeMux, prefix string) {
	// Built-in routes
	mux.HandleFunc(prefix+"/api/files", e.handleFiles)
	mux.HandleFunc(prefix+"/api/file", e.handleFile)
	mux.HandleFunc(prefix+"/api/preview", e.handlePreview)

	// Plugin routes
	for _, pl := range e.plugins {
		for _, r := range pl.Routes() {
			pattern := prefix + r.Pattern
			mux.HandleFunc(pattern, r.Handler)
		}
	}
}

// ---- handlers ---------------------------------------------------------------

// handleFiles returns a JSON array of all .md files relative to docs_dir.
// GET /_editor/api/files
func (e *Editor) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	docsDir := e.pipeline.Site().InputPath
	var files []string

	err := filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.ToLower(filepath.Ext(path)) == ".md" {
			rel, _ := filepath.Rel(docsDir, path)
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		diagnostics.PrintError(err, "editor: list files")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, files)
}

// handleFile reads or writes a single .md file.
//
//	GET  /_editor/api/file?path=guide/index.md  → returns raw markdown
//	POST /_editor/api/file?path=guide/index.md  → body is the new markdown content
func (e *Editor) handleFile(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Query().Get("path")
	if relPath == "" {
		http.Error(w, "missing path parameter", http.StatusBadRequest)
		return
	}

	absPath := filepath.Join(e.pipeline.Site().InputPath, filepath.FromSlash(relPath))
	if !isInsideDocsDir(absPath, e.pipeline.Site().InputPath) {
		http.Error(w, "path escapes docs_dir", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		content, err := e.pipeline.ReadFile(absPath)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		_, _ = w.Write(content)

	case http.MethodPost:
		var body struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if err := e.pipeline.WriteFile(absPath, []byte(body.Content)); err != nil {
			diagnostics.PrintError(err, "editor: write file")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{"status": "ok"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePreview renders a single page and returns the HTML without writing to disk.
//
//	POST /_editor/api/preview?path=guide/index.md
//	body: { "content": "# My title\n..." }   (optional — if omitted, reads from disk)
func (e *Editor) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	relPath := r.URL.Query().Get("path")
	if relPath == "" {
		http.Error(w, "missing path parameter", http.StatusBadRequest)
		return
	}

	absPath := filepath.Join(e.pipeline.Site().InputPath, filepath.FromSlash(relPath))
	if !isInsideDocsDir(absPath, e.pipeline.Site().InputPath) {
		http.Error(w, "path escapes docs_dir", http.StatusBadRequest)
		return
	}

	// If the client sent unsaved content, write it to a temp file and render that.
	var body struct {
		Content *string `json:"content"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	renderPath := absPath
	if body.Content != nil {
		tmp, err := writeTempMarkdown(*body.Content)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer os.Remove(tmp)
		renderPath = tmp
	}

	html, err := e.pipeline.RenderPageToHTML(renderPath)
	if err != nil {
		diagnostics.PrintError(err, "editor: preview")
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

// ---- helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// isInsideDocsDir prevents path-traversal attacks.
func isInsideDocsDir(absPath, docsDir string) bool {
	rel, err := filepath.Rel(docsDir, absPath)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

// writeTempMarkdown writes content to a temporary .md file and returns its path.
func writeTempMarkdown(content string) (string, error) {
	f, err := os.CreateTemp("", "n8go-preview-*.md")
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return f.Name(), err
}

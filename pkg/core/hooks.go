package core

import "github.com/yuin/goldmark"

// Plugin is the base interface every plugin must implement.
type Plugin interface {
	Name() string
}

// BuildContext carries site-wide state available to plugins during a build.
type BuildContext struct {
	Site      interface{ GetInputPath() string }
	NavTree   []*navNode
	PageIndex PageIndex
	OutputDir string
}

// InitHook is called once before any page is processed.
// Use it to initialise per-build state (open files, reset counters, etc.).
type InitHook interface {
	OnInit(ctx *BuildContext) error
}

// MarkdownExtensionHook lets a plugin inject goldmark extensions into the
// Markdown parser. Called once when the parser is constructed.
type MarkdownExtensionHook interface {
	GoldmarkExtensions() []goldmark.Extender
}

// PageRenderedHook is called after every page has been rendered to HTML but
// before it is written to disk. The returned string replaces the HTML.
// Plugins can transform markup or harvest data (e.g. search indexing).
type PageRenderedHook interface {
	OnPageRendered(page *PageResult, html string) (string, error)
}

// BuildCompleteHook is called once after all pages have been written and
// static files have been copied. Use it to write aggregate artefacts
// (search index, sitemap, etc.).
type BuildCompleteHook interface {
	OnBuildComplete(ctx *BuildContext) error
}

// EditorFileReadHook is called when the editor reads a source file before
// returning its content to the client. Plugins may decrypt, preprocess, or
// enrich the raw markdown.
type EditorFileReadHook interface {
	OnFileRead(filePath string, content []byte) ([]byte, error)
}

// EditorFileWriteHook is called when the editor is about to persist changes
// made by the client. Plugins may validate, lint, or post-process the content.
type EditorFileWriteHook interface {
	OnFileWrite(filePath string, content []byte) ([]byte, error)
}

// EditorPreviewHook is called after a preview render. Plugins may inject
// editor-specific UI (e.g. unsaved-changes banner, collaboration cursors).
type EditorPreviewHook interface {
	OnPreview(filePath string, html string) (string, error)
}

// PageResult is the public view of a rendered page exposed to plugins.
type PageResult struct {
	FilePath string
	Title    string
	URL      string
	RootPath string
}

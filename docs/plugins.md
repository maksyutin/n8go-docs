# Plugins

n8go-docs has a pluggable pipeline architecture. Plugins extend the build process and the editor API without modifying core code.

## Concepts

The build is orchestrated by a `core.Pipeline`. Plugins are registered on the pipeline before `Build()` is called. Each plugin implements one or more hook interfaces; the pipeline calls the relevant hooks at defined points in the pipeline lifecycle.

```text
Pipeline.Register(plugin) → Pipeline.Build()
                                │
                    ┌───────────▼────────────┐
                    │    InitHook            │  once, before any page
                    │    ┌──────────────┐    │
                    │    │ per page:    │    │
                    │    │ PageRendered │    │  after MD→HTML, before write
                    │    └──────────────┘    │
                    │    BuildComplete       │  once, after all pages written
                    └────────────────────────┘
```

## Pipeline Hooks (`core` package)

### `InitHook`

Called once before any page is processed. Use it to initialise per-build state.

```go
type InitHook interface {
    OnInit(ctx *BuildContext) error
}
```

`BuildContext` provides:

| Field | Type | Description |
|-------|------|-------------|
| `NavTree` | `[]*navNode` | Full navigation tree |
| `PageIndex` | `PageIndex` | Map of input file path → output URL for all pages |
| `OutputDir` | `string` | Absolute path to the output directory |

### `MarkdownExtensionHook`

Called once when the Goldmark Markdown parser is constructed. Return additional [Goldmark extensions](https://github.com/yuin/goldmark#extensions) to enable.

```go
type MarkdownExtensionHook interface {
    GoldmarkExtensions() []goldmark.Extender
}
```

### `PageRenderedHook`

Called for every page after it has been rendered to HTML but before it is written to disk. The returned string replaces the HTML passed in. Use it to transform markup or harvest data (e.g. build a search index).

```go
type PageRenderedHook interface {
    OnPageRendered(page *PageResult, html string) (string, error)
}
```

`PageResult` provides:

| Field | Description |
|-------|-------------|
| `FilePath` | Absolute path to the source `.md` file |
| `Title` | Page title (from first `# H1`) |
| `URL` | Output directory path relative to the site root |
| `RootPath` | Relative prefix to the site root (e.g. `../../`) |

### `BuildCompleteHook`

Called once after all pages have been written and static files have been copied. Use it to write aggregate artefacts (search index, sitemap, etc.).

```go
type BuildCompleteHook interface {
    OnBuildComplete(ctx *BuildContext) error
}
```

## Editor Hooks (`core` package)

These hooks are called by `core.Pipeline` when the editor API reads, writes, or previews a file.

### `EditorFileReadHook`

```go
type EditorFileReadHook interface {
    OnFileRead(filePath string, content []byte) ([]byte, error)
}
```

Called before returning file content to the client. Plugins may decrypt, preprocess, or enrich the raw Markdown.

### `EditorFileWriteHook`

```go
type EditorFileWriteHook interface {
    OnFileWrite(filePath string, content []byte) ([]byte, error)
}
```

Called before persisting client changes. Plugins may validate, lint, or post-process the content.

### `EditorPreviewHook`

```go
type EditorPreviewHook interface {
    OnPreview(filePath string, html string) (string, error)
}
```

Called after a preview render. Plugins may inject editor-specific UI into the HTML (e.g. unsaved-changes banners, collaboration cursors).

## Editor API Plugins (`editor` package)

The editor server (`editor.Editor`) supports its own plugin interface for adding HTTP routes:

```go
type EditorPlugin interface {
    Name() string
    Routes() []Route
}

type Route struct {
    Pattern string
    Handler http.HandlerFunc
}
```

Register an `EditorPlugin` with `editor.Editor.Register(plugin)` to mount additional HTTP handlers under the editor prefix (`/_editor` by default).

## Bundled Plugins

### `plugins/search`

Collects page content during `OnPageRendered` and writes `search/index.json` on `OnBuildComplete`. Registered automatically when `default_search: true`.

```go
import "n8go-docs/plugins/search"

p := core.NewPipeline(site, theme, themeDir)
p.Register(search.New(site.SearchContentLimit, site.SiteURL, site.OutputPath))
```

Constructor parameters:

| Parameter | Type | Description |
|-----------|------|-------------|
| `contentLimit` | `int` | Max rune count per page (`0` = unlimited) |
| `siteURL` | `string` | Absolute site URL for index entry URLs; empty = relative |
| `outputDir` | `string` | Absolute path to the output directory |

The plugin extracts text from the `.main-content` element of each rendered page, strips HTML, and normalises whitespace before indexing.

### `plugins/social`

Generates Open Graph / Twitter Card social preview images (1200×630 PNG) for every documentation page and injects the corresponding `<meta>` tags into `<head>`. Configuration mirrors the [MkDocs-Material social plugin](https://squidfunk.github.io/mkdocs-material/plugins/social/).

```go
import "n8go-docs/plugins/social"

p.Register(social.NewWithDefaults(social.Config{
    CardsLayoutOptions: social.LayoutOptions{
        BackgroundColor: "#1976d2",
        Color:           "#ffffff",
    },
}))
```

`NewWithDefaults` fills every unset field with the documented defaults. Use `New` when you need explicit control over boolean fields.

#### Config options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Enabled` | bool | `true` | Disable the plugin entirely when `false` |
| `Concurrency` | int | `NumCPU-1` | Parallel card generation workers |
| `Cache` | bool | `true` | Reuse unchanged cards; disable to force regeneration |
| `CacheDir` | string | `.cache/plugin/social` | Directory for cache key files |
| `LogLevel` | string | `warn` | Error reporting: `warn`, `info`, `ignore` |
| `Cards` | bool | `true` | Enable/disable card generation without disabling meta-tag injection |
| `CardsDir` | string | `assets/images/social` | Output path inside `site_dir` |
| `CardsLayoutDir` | string | `layouts` | Directory for custom layout definitions |
| `CardsLayout` | string | `default` | Active layout name |
| `CardsLayoutOptions` | `LayoutOptions` | — | Visual overrides applied to every card |
| `CardsInclude` | `[]string` | — | Glob patterns: only matching pages get a card |
| `CardsExclude` | `[]string` | — | Glob patterns: matching pages are skipped |
| `Debug` | bool | `false` | Overlay dot grid and layer outlines |
| `DebugOnBuild` | bool | `false` | Keep debug overlays during `build` (not just `serve`) |
| `DebugGrid` | bool | `true` | Draw a dot grid when `Debug` is enabled |
| `DebugGridStep` | int | `32` | Dot grid step in pixels |
| `DebugColor` | string | `grey` | Color of debug outlines and grid |

#### LayoutOptions

| Field | Default | Description |
|-------|---------|-------------|
| `BackgroundColor` | `#176bfb` | Card background. Accepts `#rgb`, `#rrggbb`, named colors, `transparent` |
| `BackgroundImage` | — | Path to a background image (tinted by `BackgroundColor`) |
| `Color` | `#ffffff` | Foreground/text color |
| `FontFamily` | `Roboto` | Typeface (Google Fonts name) |
| `FontVariant` | — | Style modifier, e.g. `Condensed` |
| `Logo` | — | Path to logo image (top-left corner) |
| `Title` | page title | Overrides the page title shown on the card |
| `Description` | — | Overrides the description line on the card |

#### Generated meta tags

For each page the plugin injects into `<head>`:

```html
<meta property="og:type" content="website">
<meta property="og:title" content="Page Title">
<meta property="og:image" content="https://example.com/assets/images/social/guide-abc12345.png">
<meta property="og:image:width" content="1200">
<meta property="og:image:height" content="630">
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="Page Title">
<meta name="twitter:image" content="https://example.com/assets/images/social/guide-abc12345.png">
```

---

### `plugins/unsaved`

An example `EditorPlugin` that injects an "unsaved changes" warning banner into preview HTML and exposes a route that reports which files have been previewed but not saved.

```go
import "n8go-docs/plugins/unsaved"

p := core.NewPipeline(site, theme, themeDir)
u := unsaved.New()
p.Register(u)    // registers OnPreview + OnFileWrite hooks

ed := editor.New(p)
ed.Register(u)   // registers GET /_editor/api/unsaved route
```

Editor routes added by this plugin:

| Route | Method | Description |
|-------|--------|-------------|
| `/_editor/api/unsaved` | `GET` | Returns a JSON array of files with unsaved previews |

## Writing a Plugin

A plugin only needs to implement the hooks it cares about. Implement `core.Plugin` (`Name() string`) plus any combination of the hook interfaces.

```go
package myplugin

import "n8go-docs/core"

type Plugin struct{}

func (p *Plugin) Name() string { return "my-plugin" }

// OnInit is optional — only implement hooks you need.
func (p *Plugin) OnInit(_ *core.BuildContext) error {
    // initialise per-build state
    return nil
}

func (p *Plugin) OnPageRendered(page *core.PageResult, html string) (string, error) {
    // transform html or harvest data
    return html, nil
}

func (p *Plugin) OnBuildComplete(_ *core.BuildContext) error {
    // write aggregate artefacts
    return nil
}

// Compile-time interface checks
var _ core.Plugin           = (*Plugin)(nil)
var _ core.InitHook         = (*Plugin)(nil)
var _ core.PageRenderedHook = (*Plugin)(nil)
var _ core.BuildCompleteHook = (*Plugin)(nil)
```

Register the plugin before calling `Build`:

```go
p := core.NewPipeline(site, theme, themeDir)
p.Register(myplugin.New())
p.Build()
```

## Editor API

When `n8go-docs serve` runs, the editor API is mounted at `/_editor`.

### Built-in endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/_editor/api/files` | JSON array of all `.md` files relative to `docs_dir` |
| `GET` | `/_editor/api/file?path=<rel>` | Raw Markdown content of the given file |
| `POST` | `/_editor/api/file?path=<rel>` | Save file. Body: `{"content": "..."}` |
| `POST` | `/_editor/api/preview?path=<rel>` | Render and return HTML. Body: `{"content": "..."}` (optional; omit to render from disk) |

All paths are relative to `docs_dir`. Requests that escape `docs_dir` via path traversal are rejected with `400`.

### Preview without saving

The `content` field in the preview request body is optional. When present, n8go-docs renders the supplied Markdown directly without writing to disk, using the existing nav tree and page index for link resolution.

```bash
curl -X POST http://localhost:9080/_editor/api/preview?path=guide/setup.md \
  -H 'Content-Type: application/json' \
  -d '{"content": "# Draft heading\n\nSome **draft** content."}'
```

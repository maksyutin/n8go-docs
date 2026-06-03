# Changelog

## Unreleased

### Added
- Pluggable pipeline architecture (`core.Pipeline`, `core.Plugin` and hook interfaces).
- Seven hook interfaces: `InitHook`, `MarkdownExtensionHook`, `PageRenderedHook`, `BuildCompleteHook`, `EditorFileReadHook`, `EditorFileWriteHook`, `EditorPreviewHook`.
- Bundled `plugins/search` — extracts `.main-content` text and writes `search/index.json`; replaces the previously inlined search indexing in the generator.
- Bundled `plugins/unsaved` — example editor plugin that injects unsaved-changes banner and exposes `GET /_editor/api/unsaved`.
- Web editor API mounted at `/_editor` by `n8go-docs serve`:
  - `GET /_editor/api/files` — list all `.md` files in `docs_dir`.
  - `GET/POST /_editor/api/file?path=<rel>` — read and write source files.
  - `POST /_editor/api/preview?path=<rel>` — render Markdown to HTML without writing to disk; accepts unsaved content in the request body.
- Path-traversal protection on all editor API endpoints.
- `editor.EditorPlugin` interface for registering additional HTTP routes under `/_editor`.
- `THEME_DIR` environment variable to override the themes root directory.
- `n8go-docs build` command (alias for `generate`).
- `--json` flag on `build` and `version` commands for machine-readable output.
- CI/CD pipeline configuration for GitHub Actions, GitLab CI, GitVerse, and GitFlic.

### Changed
- `GenerateDocumentation` now accepts variadic `...core.Plugin` — pass plugins explicitly instead of relying on internal wiring.
- `generateThemedHtmlForPage` replaced by `Pipeline.renderContextToHTML` + `writePageHTML`; rendering and I/O are now decoupled.
- `serve.go` uses a shared `http.ServeMux` combining the static file server and the editor API router.

## Earlier Changes

- YAML configuration (`n8go-docs.yaml`), FlexSearch/Fuse engine selection, `search_content_limit`, `strip_md_extension`.
- `StripMdExtension` option to strip `.md` from internal links in generated HTML.
- Full-text search index generation (`search/index.json`).
- Jinja2-compatible theme engine (`gonja/v2`); Go `text/template` syntax deprecated and rejected at build time.
- Stricter theme validation: detects legacy Go template syntax, missing `page.content`, and hard-coded local asset paths.
- Bundled themes: `default` and `material`.
- Live-reload development server (`n8go-docs serve`).
- Automatic and explicit (`nav:`) navigation tree building.
- Syntax highlighting via Chroma; emoji support via Twemoji.
- `exclude_docs` glob patterns (list and multiline-string forms).
- `extra_css` and `extra_javascript` support.
- `custom_font`, `head_tags`, `logo` configuration options.

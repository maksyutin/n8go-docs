# n8go-docs

A static documentation generator written in Go, inspired by MkDocs.

## Features

- Write documentation in Markdown (GitHub Flavored)
- Outputs static HTML — host anywhere (GitHub Pages, S3, nginx, …)
- Syntax highlighting via [Chroma](https://github.com/alecthomas/chroma), emoji support (Twemoji)
- Extensible with custom Jinja2-compatible themes (engine: [`gonja/v2`](https://github.com/nikolalohinski/gonja))
- Full-text search powered by [FlexSearch](https://github.com/nextapps-de/flexsearch) or [Fuse.js](https://fusejs.io/)
- YAML configuration (`n8go-docs.yaml`)
- `.md` extension stripping for clean URLs
- Pluggable pipeline — extend build and editor behaviour via hooks without forking
- Built-in web editor API (`/_editor`) for live Markdown editing and preview

## Installation

Download the latest binary from the [releases](https://github.com/maksyutin/n8go-docs/releases) page.

The `themes/` directory must be located next to the binary. Override with `THEME_DIR`:

```bash
THEME_DIR=/path/to/themes n8go-docs build
```

## Quick Start

**1. Create `n8go-docs.yaml` in your project root:**

```yaml
site_name: My Docs
site_description: Project documentation
site_url: https://example.com/docs/
docs_dir: docs
site_dir: site
theme: default
search_engine: flexsearch
search_content_limit: 500
strip_md_extension: true
```

**2. Build the site:**

```bash
n8go-docs build
```

**3. Start the development server with live reload:**

```bash
n8go-docs serve
# or on a custom port:
n8go-docs serve --port 3000
```

Default port: `9080`. The editor API is available at `http://localhost:9080/_editor`.

## CLI Commands

| Command | Description |
|---------|-------------|
| `n8go-docs build` | Build the static site once and exit. Alias: `generate`. |
| `n8go-docs serve` | Start the dev server with live reload and editor API. |
| `n8go-docs version` | Print the version and exit. |

Common flags: `--config <path>` (default: `n8go-docs.yaml`), `--verbose` / `-v`, `--json`.

## Theme Notes

- Templates use Go-native Jinja2-compatible syntax (`gonja/v2`). Go `text/template` syntax is **not** supported.
- Theme static assets are emitted to `<site_dir>/assets/css/`, `/assets/js/`, `/assets/img/`.
- All asset and page links in templates **must** use the `|url` filter to produce correct paths on nested pages.
- Both bundled themes (`default`, `material`) follow this convention.

## Documentation

- [Configuration](docs/config.md) — all options with defaults
- [Themeing](docs/themeing.md) — creating and customising themes, full template variable reference
- [Components](docs/components.md) — built-in UI components (default theme)
- [Deploy](DEPLOY.md) — CI/CD pipeline, deploy providers, secrets reference

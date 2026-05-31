# n8go-docs

A clean static documentation generator written in Go.

## Features

- Write documentation in Markdown (GitHub Flavored)
- Outputs static HTML — host anywhere (GitHub Pages, S3, nginx, …)
- Syntax highlighting, emoji support
- Extensible with custom Jinja2-compatible themes
- Full-text search powered by [FlexSearch](https://github.com/nextapps-de/flexsearch) or [Fuse.js](https://fusejs.io/)
- YAML configuration (`n8go-docs.yaml`)
- `.md` extension stripping for clean URLs

## Theme runtime notes

- Theme templates are rendered by a Go-native Jinja2-compatible engine (Jinja2go via `gonja/v2`).
- Theme static files are emitted under `/assets/css`, `/assets/js`, and `/assets/img`.
- In the bundled `material` theme, internal navigation updates page content and sidebars without reloading header/footer.

## Installation

Download the latest binary from the [releases](https://github.com/maksyutin/n8go-docs/releases) page.

## Quick start

**1. Create `n8go-docs.yaml` in your project root:**

```yaml
site_name: My Docs
site_description: Project documentation for My Docs
site_url: https://example.com/docs/
docs_dir: docs
site_dir: docs_gen
use_directory_urls: true
search_engine: flexsearch   # flexsearch | fuse
search_content_limit: 500
strip_md_extension: true
```

**2. Generate the site:**

```bash
n8go-docs generate
```

**3. Start the dev server (with live reload):**

```bash
n8go-docs serve
```

The server starts on port `9080` by default. Pass a custom port as an argument:

```bash
n8go-docs serve --port 3000
```

## Configuration

See [Configuration](docs/config.md) for all available options.

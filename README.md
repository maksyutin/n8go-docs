# n8go-docs

A clean static documentation generator written in Go.

## Features

- Write documentation in Markdown (GitHub Flavored)
- Outputs static HTML — host anywhere (GitHub Pages, S3, nginx, …)
- Syntax highlighting, emoji support
- Extensible with custom themes (Go templates)
- Full-text search powered by [FlexSearch](https://github.com/nextapps-de/flexsearch) or [Fuse.js](https://fusejs.io/)
- YAML configuration (`n8go-docs.yaml`)
- `.md` extension stripping for clean URLs

## Installation

Download the latest binary from the [releases](https://github.com/maksyutin/n8go-docs/releases) page.

## Quick start

**1. Create `n8go-docs.yaml` in your project root:**

```yaml
name: My Docs
input: docs
output: docs_gen
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
n8go-docs serve 3000
```

## Configuration

See [Configuration](docs_src/config.md) for all available options.

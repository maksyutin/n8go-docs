# Introduction

n8go-docs is a static documentation generator written in Go, inspired by MkDocs.

## Features

- Write documentation in Markdown (GitHub Flavored)
- Outputs static HTML — host anywhere (GitHub Pages, S3, nginx, …)
- Syntax highlighting via [Chroma](https://github.com/alecthomas/chroma), emoji support (Twemoji)
- Extensible with custom Jinja2-compatible themes (engine: `gonja/v2`)
- Full-text search powered by [FlexSearch](https://github.com/nextapps-de/flexsearch) or [Fuse.js](https://fusejs.io/)
- YAML configuration (`n8go-docs.yaml`)
- `.md` extension stripping for clean URLs
- Pluggable pipeline architecture — extend build and editor behaviour without forking
- Built-in web editor API (`/_editor`) for live Markdown editing and preview

## Installation

Download the latest binary for your platform from the [releases](https://github.com/maksyutin/n8go-docs/releases) page.

The themes directory is expected to be located next to the binary at `./themes/`. Override it with the `THEME_DIR` environment variable.

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

**2. Generate the site:**

```bash
n8go-docs build
```

**3. Start the development server with live reload:**

```bash
n8go-docs serve
```

The server starts on port `9080` by default. Pass `--port` to override:

```bash
n8go-docs serve --port 3000
```

The editor API is available at `http://localhost:9080/_editor` when the server is running.

## CLI Commands

| Command | Description |
|---------|-------------|
| `n8go-docs build` | Build the static site once and exit. Alias: `generate`. |
| `n8go-docs serve` | Start the development server with live reload and editor API. |
| `n8go-docs version` | Print the version and exit. |

Common flags available on all commands:

| Flag | Default | Description |
|------|---------|-------------|
| `--config <path>` | `n8go-docs.yaml` | Path to the config file. |
| `--verbose` / `-v` | `false` | Print detailed build progress to stderr. |
| `--json` | `false` | Output machine-readable JSON (supported on `build` and `version`). |

## Navigation

Navigation is built in two ways:

- **Automatic** (default): n8go-docs scans `docs_dir` recursively. Directories become sections; `index.md` or `README.md` (when no `index.md` is present) become section indexes.
- **Explicit**: define the full tree in `n8go-docs.yaml` under the `nav` key. See [Configuration](config.md) for the syntax.

## Further Reading

- [Configuration](config.md) — all configuration options with defaults
- [Themeing](themeing.md) — how to create and customize themes
- [Components](components.md) — built-in UI components available in the default theme

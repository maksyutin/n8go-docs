# Configuration

n8go-docs is configured via a `n8go-docs.yaml` file. By default the file is read from the current working directory; pass `--config <path>` to use a different location.

## Full Example

```yaml
site_name: My Docs
site_description: Project documentation for My Docs
site_url: https://example.com/docs/

# Optional: override the local development address used by `n8go-docs serve`.
# dev_addr: 127.0.0.1:8000

use_directory_urls: true

docs_dir: docs
site_dir: site
theme: default

default_search: true
search_engine: flexsearch
search_content_limit: 500
strip_md_extension: true

custom_font: Roboto
head_tags:
  - '<link rel="preconnect" href="https://fonts.googleapis.com">'
  - '<link href="https://fonts.googleapis.com/css2?family=Roboto&display=swap" rel="stylesheet">'

logo: assets/img/logo.svg

extra_css:
  - css/extra.css

extra_javascript:
  - js/extra.js

exclude_docs:
  - drafts/
  - "**/secret-*.md"

nav:
  - Home: index.md
  - User Guide:
      - Installation: guide/installation.md
      - Configuration: guide/configuration.md
```

## Options

### site_name *(required)*

The site title. Used in HTML `<title>` tags as `Page — My Docs`.

```yaml
site_name: My Docs
```

---

### site_description

Short site description for themes, SEO metadata, and auxiliary files.

```yaml
site_description: Project documentation for My Docs
```

---

### site_url

Public base URL of the generated site. Recommended for any site published to a fixed address.

```yaml
site_url: https://example.com/docs
```

With this value a page at `guide/setup/` is linked as `https://example.com/docs/guide/setup/`.

`site_url` affects: absolute links in navigation, `base_url` template variable, static asset URLs, and search index entry URLs. The trailing slash is optional — n8go-docs normalizes it.

When omitted:
- all links are relative to the current page;
- `base_url` equals the page's `root_path` value;
- search index entries use relative URLs.

---

### dev_addr

Local address used by `n8go-docs serve` when `--port` is not passed on the command line.

```yaml
dev_addr: 127.0.0.1:8000
```

Default: not set (server binds to `:9080`).

---

### use_directory_urls

When `true`, every page is generated as `page/index.html` and linked as `/page/`. When `false`, pages are generated as `page.html`. Default: `true`.

```yaml
use_directory_urls: true
```

---

### docs_dir

Directory containing the Markdown source files, resolved relative to the config file. Default: `docs`.

```yaml
docs_dir: docs
```

---

### site_dir

Directory where the generated HTML is written, resolved relative to the config file. Default: `site`.

```yaml
site_dir: site
```

---

### theme

Theme ID. n8go-docs looks for the theme in the `themes/` directory located next to the binary. Default: `default`.

```yaml
theme: default
```

Override the themes root with the `THEME_DIR` environment variable:

```bash
THEME_DIR=/path/to/themes n8go-docs build
```

Bundled themes: `default`, `material`.

---

### default_search

Set to `false` to disable the built-in search entirely. Default: `true`.

```yaml
default_search: false
```

When `true`, n8go-docs registers the `search` plugin which extracts text from `.main-content` and writes `search/index.json` on build completion.

---

### search_engine

Client-side search engine to use. Default: `flexsearch`.

| Value | Library |
|-------|---------|
| `flexsearch` | [FlexSearch](https://github.com/nextapps-de/flexsearch) — fast tokenized search |
| `fuse` | [Fuse.js](https://fusejs.io/) — fuzzy search |

```yaml
search_engine: flexsearch
```

---

### search_content_limit

Maximum number of characters (Unicode code points) stored per page in the search index. Reduces index size without affecting quality for typical queries. Set to `0` for no limit. Default: `500`.

```yaml
search_content_limit: 500
```

---

### strip_md_extension

When `true`, `.md` extensions are stripped from all internal links in the generated HTML. Useful when source files reference each other with `.md` links for editor navigation, while the published site uses clean URLs. Default: `false`.

```yaml
strip_md_extension: true
```

---

### custom_font

CSS font-family name applied site-wide. Load the font itself via `head_tags`.

```yaml
custom_font: Roboto
```

---

### head_tags

Raw HTML strings injected into `<head>` on every page. Use for fonts, analytics, meta tags, or any other head-level markup.

```yaml
head_tags:
  - '<link rel="stylesheet" href="https://example.com/style.css">'
  - '<meta name="author" content="Your Name">'
```

---

### logo

Path to the logo image, relative to the output directory. Default: `assets/img/logo.svg`.

```yaml
logo: assets/img/my-logo.svg
```

---

### extra_css

CSS files relative to `docs_dir`. Each file is copied to the output directory and linked on every page.

```yaml
extra_css:
  - css/extra.css
```

---

### extra_javascript

JavaScript files relative to `docs_dir`. Each file is copied to the output directory and injected as `<script defer>` on every page.

```yaml
extra_javascript:
  - js/extra.js
```

---

### nav

Explicit navigation tree. Each entry is either a leaf (`Title: file.md`) or a section with nested children. When omitted, n8go-docs builds the menu automatically from the filesystem — directories become sections, `index.md` or `README.md` becomes the section index.

```yaml
nav:
  - Home: index.md
  - User Guide:
      - Installation: guide/installation.md
      - Configuration: guide/configuration.md
  - Reference: reference.md
```

Paths are relative to `docs_dir`.

---

### exclude_docs

Glob patterns of paths to exclude entirely — matching files are neither rendered nor copied as static assets. Patterns are matched relative to `docs_dir`.

Accepted as a YAML list:

```yaml
exclude_docs:
  - drafts/
  - "**/secret-*.md"
  - wip.md
```

Or as a multiline block scalar:

```yaml
exclude_docs: |
  drafts/
  **/secret-*.md
  wip.md
```

Both forms support inline `# comments` and ignore blank lines.

Pattern rules:
- A pattern without `/` matches the **basename** at any depth.
- `**` matches any number of path segments.
- A trailing `/` matches a directory and everything inside it.
- In the list form, patterns starting with `*` or `**` must be quoted to avoid YAML parsing them as aliases.

## Default Values

| Option | Default |
|--------|---------|
| `docs_dir` | `docs` |
| `site_dir` | `site` |
| `theme` | `default` |
| `use_directory_urls` | `true` |
| `default_search` | `true` |
| `search_engine` | `flexsearch` |
| `search_content_limit` | `500` |
| `strip_md_extension` | `false` |
| `logo` | `assets/img/logo.svg` |

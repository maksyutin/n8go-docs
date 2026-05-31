# Configuration

n8go-docs is configured via a `n8go-docs.yaml` file in the working directory from which `n8go-docs` is run.

## Full example

```yaml
site_name: My Docs
site_description: Project documentation for My Docs
site_url: https://example.com/docs/

# Optional local development address.
#dev_addr: 127.0.0.1:8000

# Use directory URLs such as /guide/ instead of /guide.html.
use_directory_urls: true

docs_dir: docs
site_dir: site
theme: default

search_engine: flexsearch
search_content_limit: 500
strip_md_extension: true

custom_font: Roboto
head_tags:
  - '<link rel="preconnect" href="https://fonts.googleapis.com">'
  - '<link href="https://fonts.googleapis.com/css2?family=Roboto&display=swap" rel="stylesheet">'

exclude_docs:
  - drafts/
  - "**/secret-*.md"
```

## Options

### site_name *(required)*

The site title. Used in HTML `<title>` tags: `Page — My Docs`.

```yaml
site_name: My Docs
```

---

### site_description

Short site description for themes, SEO metadata, and generated auxiliary files.

```yaml
site_description: Project documentation for My Docs
```

---

### site_url

Public base URL of the generated site. This is the standard parameter for production builds and is recommended whenever the site is published.

```yaml
site_url: https://108n.online/xboiler
```

With this value, a page at `adr/0016-versioning-and-compatibility/` is linked as:

```text
https://108n.online/xboiler/adr/0016-versioning-and-compatibility/
```

`site_url` is used for absolute links in SEO, sitemap generation, metadata, `base_url`, static asset URLs, navigation URLs, and search index URLs. Do not confuse it with the template variable `base_url` or the local development option `dev_addr`.

If `site_url` is omitted:

- canonical links are not generated;
- `sitemap.xml` uses relative paths;
- search engines may see duplicate content when the same site is served from multiple hosts or paths.

The trailing slash is optional; n8go-docs normalizes it.

---

### dev_addr

Local development address used by `n8go-docs serve` when `--port` is not passed.

```yaml
dev_addr: 127.0.0.1:8000
```

---

### use_directory_urls

When `true`, pages are generated as directory indexes such as `guide/index.html` and linked as `/guide/`. Default: `true`.

```yaml
use_directory_urls: true
```

---

### docs_dir

Directory containing the Markdown source files. Default: `docs`.

```yaml
docs_dir: docs
```

---

### site_dir

Directory where static HTML is written. Default: `site`.

```yaml
site_dir: site
```

---

### theme

Theme ID. n8go-docs looks for the theme in the `themes/` directory next to the binary. Default: `default`.

```yaml
theme: default
```

The theme directory can be overridden with the `THEME_DIR` environment variable.

---

### search_engine

Full-text search engine to use. Default: `flexsearch`.

| Value | Library |
|-------|---------|
| `flexsearch` | [FlexSearch](https://github.com/nextapps-de/flexsearch) — fast tokenized search |
| `fuse` | [Fuse.js](https://fusejs.io/) — fuzzy search |

```yaml
search_engine: flexsearch
```

---

### search_content_limit

Maximum number of characters stored per page in the search index. Reduces index size without affecting search quality for typical queries. Set to `0` for no limit. Default: `500`.

```yaml
search_content_limit: 500
```

---

### strip_md_extension

When `true`, strips `.md` extensions from all internal links in generated HTML. Useful when source files reference each other with `.md` links (e.g. for editor navigation) while the site serves clean URLs. Default: `false`.

```yaml
strip_md_extension: true
```

---

### default_search

Set to `false` to disable the built-in search entirely. Default: `true`.

```yaml
default_search: false
```

---

### custom_font

Name of a custom font to apply site-wide. Make sure to load the font via `head_tags`.

```yaml
custom_font: Roboto
```

---

### head_tags

List of raw HTML tags injected into `<head>` on every page. Use for fonts, analytics, meta tags, etc.

```yaml
head_tags:
  - '<link rel="stylesheet" href="https://example.com/style.css">'
  - '<meta name="author" content="Your Name">'
```

---

### logo

Path to the logo image (relative to the output directory). Default: `assets/img/logo.svg`.

```yaml
logo: assets/img/my-logo.svg
```

---

### extra_css

List of CSS files (paths relative to `docs_dir`) copied to the output and linked on every page.

```yaml
extra_css:
  - css/extra.css
```

---

### extra_javascript

List of JavaScript files (paths relative to `docs_dir`) copied to the output and added as `<script defer>` on every page.

```yaml
extra_javascript:
  - js/extra.js
```

---

### nav

Explicit navigation tree. Each entry is either a single page (`Title: file.md`) or a section with nested children. When omitted, the menu is built automatically from the filesystem.

```yaml
nav:
  - Home: index.md
  - User Guide:
      - Installation: guide/installation.md
      - Configuration: guide/configuration.md
```

---

### exclude_docs

Glob patterns of files and directories to exclude from the build entirely — they are neither rendered nor copied as static assets. Patterns are matched relative to `docs_dir`.

Can be given as a multiline string:

```yaml
exclude_docs: |
  drafts/
  **/secret-*.md
  wip.md
```

…or as a YAML list:

```yaml
exclude_docs:
  - drafts/
  - "**/secret-*.md"
  - wip.md
```

Both forms support inline comments (`# …`) and ignore blank lines. Supported wildcards: `*`, `**`, `?`. A pattern without `/` matches the basename at any depth; `**` matches any number of path segments; a trailing `/` matches a directory and everything inside it.

> In the list form, a pattern starting with `*` or `**` must be quoted (`- "**/secret-*.md"`) — otherwise YAML treats it as an alias. In the multiline-string form quoting is not needed.

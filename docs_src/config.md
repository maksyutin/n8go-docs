# Configuration

n8go-docs is configured via a `n8go-docs.yaml` file in the working directory from which `n8go-docs` is run.

## Full example

```yaml
name: My Docs
input: docs
output: docs_gen
theme: default

search_engine: flexsearch
search_content_limit: 500
strip_md_extension: true

custom_font: Roboto
head_tags:
  - '<link rel="preconnect" href="https://fonts.googleapis.com">'
  - '<link href="https://fonts.googleapis.com/css2?family=Roboto&display=swap" rel="stylesheet">'
```

## Options

### name *(required)*

The site title. Used in HTML `<title>` tags: `Page — My Docs`.

```yaml
name: My Docs
```

---

### input

Directory containing the Markdown source files. Default: `docs`.

```yaml
input: docs
```

---

### output

Directory where static HTML is written. Default: `docs_gen`.

```yaml
output: docs_gen
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

Path to the logo image (relative to the output directory). Default: `img/book.svg`.

```yaml
logo: img/my-logo.svg
```

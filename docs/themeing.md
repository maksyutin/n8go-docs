# Themeing

Themes define the HTML structure, static assets, navigation rendering, and page chrome for an n8go-docs site.

> **Engine:** n8go-docs renders themes with the Go-native Jinja2-compatible engine backed by [`gonja/v2`](https://github.com/nikolalohinski/gonja). Legacy Go `text/template` syntax is **not** supported and will cause a build error.

## Theme Directory Layout

```text
my_theme/
├── theme.yaml       # Required: theme metadata and highlighting config
├── main.html        # Required: root Jinja2 template
├── body.html        # Optional: included template for page body
├── nav.html         # Optional: included template for navigation
├── css/
│   └── style.css    # Copied to output/assets/css/
├── js/
│   └── script.js    # Copied to output/assets/js/
└── img/
    └── logo.svg     # Copied to output/assets/img/
```

`main.html` and `theme.yaml` are the only required files. During build, theme subdirectories are emitted to:

| Source | Output |
|--------|--------|
| `css/*` | `<site_dir>/assets/css/*` |
| `js/*` | `<site_dir>/assets/js/*` |
| `img/*` | `<site_dir>/assets/img/*` |

The theme directory is resolved from `themes/<theme_id>/` next to the binary. Override the root with `THEME_DIR`:

```bash
THEME_DIR=/path/to/themes n8go-docs build
```

## theme.yaml

`theme.yaml` is required. It provides metadata and controls syntax highlighting.

```yaml
theme:
  name: My Theme
  version: 1.0.0
  description: A clean documentation theme
  author: Your Name
  repository: https://github.com/you/my-theme
  license: MIT

highlighting:
  style: github        # Any Chroma style name
  line_numbers: false
```

| Field | Required | Description |
|-------|:--------:|-------------|
| `theme.name` | ✓ | Theme display name |
| `theme.version` | ✓ | Semantic version |
| `theme.description` | — | Short description |
| `theme.author` | — | Author name |
| `theme.repository` | — | Repository URL |
| `theme.license` | — | SPDX license identifier |
| `highlighting.style` | ✓ | [Chroma style](https://xyproto.github.io/splash/docs/) (e.g. `github`, `monokai`, `xcode-dark`) |
| `highlighting.line_numbers` | — | Show line numbers in code blocks. Default: `false` |

## main.html

`main.html` must contain the complete HTML document and render the current page content via `{{ page.content }}`. The build fails with a clear error if the file is missing or does not render `page.content`.

Minimal working example:

```jinja2
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{% if page_title %}{{ page_title }} — {% endif %}{{ config.site_name }}</title>
  {% for tag in config.head_tags %}{{ tag }}{% endfor %}
  <link href="{{ 'css/style.css'|url }}" rel="stylesheet">
  {% for path in config.extra_css %}
    <link href="{{ path|url }}" rel="stylesheet">
  {% endfor %}
</head>
<body>
  <nav>
    {% for item in nav %}
      <a href="{{ item.url|url }}" class="{% if item.active %}active{% endif %}">
        {{ item.title }}
      </a>
    {% endfor %}
  </nav>

  <main class="main-content">
    {{ page.content }}
  </main>

  <script src="{{ 'js/script.js'|url }}"></script>
  {% for path in config.extra_javascript %}
    <script src="{{ path|url }}"></script>
  {% endfor %}
</body>
</html>
```

> **Important:** the `main-content` CSS class on the `<main>` element is required for the search plugin to extract page text correctly.

## Template Variables

All variables are available inside every template file included from `main.html`.

### Top-level

| Variable | Type | Description |
|----------|------|-------------|
| `config` | object | All fields from `n8go-docs.yaml` (see below) |
| `page` | object | Current page data (see below) |
| `nav` | list | Navigation tree (see below) |
| `base_url` | string | `site_url` when set; otherwise the page's `root_path` |
| `root_path` | string | Relative path prefix to the site root (e.g. `../../`) |
| `site_url` | string | Value of `site_url` from config, or empty string |
| `homepage_url` | string | Absolute or relative URL of the homepage |
| `page_title` | string | Current page title (same as `page.title`) |
| `url` | string | Output directory path of the current page |
| `input_dir` | string | Absolute path to the directory of the source `.md` file |
| `generator.name` | string | `n8go-docs` |
| `generator.version` | string | Current application version |
| `now` | string | Build timestamp in `YYYY-MM-DD HH:MM:SS.mmm` format |
| `site` | object | Alias for `config` |

### config object

Mirrors `n8go-docs.yaml` directly:

| Key | Description |
|-----|-------------|
| `config.site_name` | Site title |
| `config.site_url` | Public base URL |
| `config.site_description` | Site description |
| `config.dev_addr` | Development server address |
| `config.use_directory_urls` | Boolean |
| `config.theme` | Theme ID |
| `config.docs_dir` | Resolved absolute path to docs directory |
| `config.site_dir` | Resolved absolute path to output directory |
| `config.default_search` | Boolean |
| `config.search_engine` | `flexsearch` or `fuse` |
| `config.search_content_limit` | Integer |
| `config.head_tags` | List of raw HTML strings |
| `config.custom_font` | Font family name |
| `config.logo` | Logo path |
| `config.strip_md_extension` | Boolean |
| `config.extra_css` | List of CSS file paths |
| `config.extra_javascript` | List of JS file paths |
| `config.exclude_docs` | List of glob patterns |

### page object

| Key | Description |
|-----|-------------|
| `page.title` | Page title extracted from the first `# H1` heading |
| `page.content` | Rendered page body as sanitized HTML (safe to output unescaped) |
| `page.body` | Alias for `page.content` |
| `page.toc` | Table of contents — list of `{id, name, title, level}` for H1–H3 headings |
| `page.file_path` | Absolute path to the source `.md` file |
| `page.file_name` | Base name of the source file without extension |
| `page.url` | Output directory path relative to the site root |

### nav items

Each item in `nav` (and recursively in `item.children`) has:

| Key | Description |
|-----|-------------|
| `item.title` | Display name |
| `item.name` | Alias for `title` |
| `item.url` | Output URL of the page |
| `item.active` | `true` when this item (or a descendant) is the current page |
| `item.children` | List of child nav items (empty list for leaf pages) |

## Filters

| Filter | Description |
|--------|-------------|
| `\|url` | Resolves a path relative to the site root, respecting `root_path` and `site_url`. **Must** be used for all asset and page URLs. |
| `\|urlquery` | URL-encodes a string (equivalent to `url.QueryEscape`). |

### Navigation with active state

```jinja2
<ul>
  {% for item in nav %}
    <li class="{% if item.active %}active{% endif %}">
      <a href="{{ item.url|url }}">{{ item.title }}</a>
      {% if item.children %}
        <ul>
          {% for child in item.children %}
            <li class="{% if child.active %}active{% endif %}">
              <a href="{{ child.url|url }}">{{ child.title }}</a>
            </li>
          {% endfor %}
        </ul>
      {% endif %}
    </li>
  {% endfor %}
</ul>
```

### Table of contents

```jinja2
{% if page.toc %}
<nav class="toc">
  {% for entry in page.toc %}
    <a href="#{{ entry.id }}" data-level="{{ entry.level }}">{{ entry.name }}</a>
  {% endfor %}
</nav>
{% endif %}
```

## Control Structures

In addition to standard Jinja2, the engine supports:

| Syntax | Description |
|--------|-------------|
| `{% do expression %}` | Evaluate an expression without producing output (e.g. `{% do list.append(item) %}`). |
| `{% break %}` | Break out of a `{% for %}` loop. |
| `{% continue %}` | Skip to the next iteration of a `{% for %}` loop. |
| `{% for k, v in dict %}` | Iterate over key-value pairs. |

## Build-time Validation

n8go-docs validates every theme before rendering. The build fails if:

- `main.html` is missing.
- Any `.html` file uses Go `text/template` syntax (`{{.`, `{{ range`, `{{ if`).
- Any `.html` file hard-codes local asset paths (e.g. `href="css/style.css"`) instead of using the `url` filter.
- No template file renders `page.content`.

## Bundled Themes

| Theme ID | Description |
|----------|-------------|
| `default` | Clean, accessible theme with light/dark toggle and FlexSearch integration. |
| `material` | Material Design-inspired theme with a collapsible sidebar. |

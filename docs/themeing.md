# Themeing

Themes define the HTML structure, static assets, navigation rendering, and page chrome for an n8go-docs site.

> Note: n8go-docs renders themes with the Go-native Jinja2-compatible engine Jinja2go, currently backed by `github.com/nikolalohinski/gonja/v2`. Legacy Go `text/template` syntax is not supported in theme files.

## Minimal Theme Directory

A theme directory must follow this minimum layout:

```text
my_theme/
├── main.html        # Required: main template
├── base.html        # Optional: base template
├── css/
│   └── style.css
├── js/
│   └── script.js
├── img/
│   └── logo.png
└── theme.yaml       # Optional: theme config
```

`main.html` is the only required template file. Static files must live inside theme subdirectories such as `css/`, `js/`, and `img/`.

During build, these theme directories are emitted to:

- `css/*` -> `/assets/css/*`
- `js/*` -> `/assets/js/*`
- `img/*` -> `/assets/img/*`

## Required Template: main.html

`main.html` must contain the base HTML document structure and render the current page content.

Target Jinja2-compatible example:

```jinja2
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{% if page_title %}{{ page_title }} - {% endif %}{{ config.site_name }}</title>
    {% for path in config.extra_css %}
      <link href="{{ path|url }}" rel="stylesheet">
    {% endfor %}
  </head>
  <body>
    <nav>
      <a href="{{ nav.homepage.url|url }}">{{ config.site_name }}</a>
    </nav>

    <main>
      {{ page.content }}
    </main>

    {% for path in config.extra_javascript %}
      <script src="{{ path|url }}"></script>
    {% endfor %}
  </body>
</html>
```

The builder must fail with a clear error if `main.html` is missing.

## Template Variables

Templates must have access to these required variables:

| Variable | Type | Description |
| --- | --- | --- |
| `config` | `n8go-docs.config.Config` | Full configuration from `n8go-docs.yml` or `n8go-docs.yaml`. |
| `nav` | `n8go-docs.structure.nav.Navigation` | Navigation object for the site. |
| `page` | `n8go-docs.structure.pages.Page` | Current page object. |
| `base_url` | `str` | Base URL of the generated site. |
| `homepage_url` | `str` | URL of the homepage. |

The renderer maps the internal Go page context to the Jinja2-style names above.

## Filters And Functions

The theme engine must support these filters:

```jinja2
<!-- url: normalizes paths for nested pages -->
<a href="{{ 'page.md'|url }}">Link</a>

<!-- tojson: serializes values as JSON -->
<script>
  var config = {{ config|tojson }};
</script>
```

The `url` filter is required for correct links in subdirectories. Themes must not hard-code root-relative or page-relative paths when the `url` filter can be used.

## Optional Theme Config

`theme.yaml` is optional. When present, it may define theme metadata and rendering defaults such as:

```yaml
theme:
  name: My Theme
  version: 1.0.0
  author: You
  license: MIT

highlighting:
  style: bw
  line_numbers: false
```

## Default Jinja2 Extensions

The Jinja2-compatible engine must enable these extensions by default:

| Extension | Purpose |
| --- | --- |
| `jinja2.ext.do` | Enables `{% do ... %}`. |
| `jinja2.ext.loopcontrols` | Enables `break` and `continue` in loops. |

If native Jinja2 is unavailable, n8go-docs must provide an internal Go implementation or compatibility layer that supports the required variables, filters, functions, and extensions.

## Static Files

All CSS, JavaScript, and image files must be stored under theme subdirectories and referenced through the `url` filter:

```jinja2
<link href="{{ 'assets/css/style.css'|url }}" rel="stylesheet">
<script src="{{ 'assets/js/script.js'|url }}"></script>
<img src="{{ 'assets/img/logo.png'|url }}" alt="Logo">
```

This prevents broken assets on nested pages and avoids hard-coded paths.

## Navigation

Themes must handle nested navigation correctly and preserve active state:

```jinja2
<ul class="nav">
  {% for nav_item in nav %}
    <li class="{% if nav_item.active %}active{% endif %}">
      <a href="{{ nav_item.url|url }}">{{ nav_item.title }}</a>
      {% if nav_item.children %}
        <ul class="subnav">
          {% for child in nav_item.children %}
            <li>
              <a href="{{ child.url|url }}">{{ child.title }}</a>
            </li>
          {% endfor %}
        </ul>
      {% endif %}
    </li>
  {% endfor %}
</ul>
```

## MkDocs Compatibility

n8go-docs themes should be compatible with the MkDocs/Jinja2 theme model where practical. The key compatibility requirements are:

- `main.html` is the minimum required file.
- `page.content` must render the current page body as sanitized HTML.
- The `url` filter must normalize links and static asset paths for nested pages.
- `config` and `nav` must be available for site settings and menus.
- Static assets must be linked with `url`.
- Themes must use `base_url` instead of hard-coded root paths.

If a theme does not satisfy these requirements, the build must fail with a clear, actionable error message.

## Implementation Plan

1. Done: documented and froze the target theme contract.
2. Done: added Jinja2go rendering with a clear build error when `main.html` is missing or invalid.
3. Done: added a compatibility context that maps internal Go data to `config`, `nav`, `page`, `base_url`, and `homepage_url`.
4. Done: implemented `url` support and use the engine-provided `tojson` support.
5. Done: migrated bundled themes to Jinja2-compatible syntax.
6. Done: added tests for missing `main.html`, bundled theme rendering, nested static URLs, `page.content`, nav nesting, and `tojson`.
7. Done: added stricter validation for legacy Go-template syntax, missing `page.content`, and hard-coded local static asset paths.
8. Next: expand validation for more MkDocs-compatible constructs as themes start using them.

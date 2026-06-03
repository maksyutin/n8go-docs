package core

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"n8go-docs/manifest"
)

func TestGenerateTemplateRequiresMainHTML(t *testing.T) {
	_, err := generateTemplate(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "theme must provide required main.html") {
		t.Fatalf("expected required main.html error, got %v", err)
	}
}

func TestGenerateTemplateRejectsInvalidThemeContract(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "legacy Go syntax",
			content: `<main>{{ .Page.Body }}</main>`,
			want:    "legacy Go template syntax",
		},
		{
			name:    "hard-coded static asset",
			content: `<link rel="shortcut icon" href="assets/img/logo.svg"><main>{{ page.content }}</main>`,
			want:    "hard-codes local static asset paths",
		},
		{
			name:    "missing page content",
			content: `<main>{{ page.title }}</main>`,
			want:    "must render page.content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			themeDir := t.TempDir()
			writeTestFile(t, themeDir+"/"+RootTemplateName, tt.content)

			_, err := generateTemplate(themeDir)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestBundledThemesRenderWithJinja2go(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve test file path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))

	for _, themeName := range []string{"default", "material"} {
		t.Run(themeName, func(t *testing.T) {
			tmpl, err := generateTemplate(filepath.Join(repoRoot, "themes", themeName))
			if err != nil {
				t.Fatal(err)
			}

			html, err := renderTemplateHTML(tmpl, &pageContext{
				Page: pageInfo{
					FilePath: "docs/index.md",
					FileName: "index.md",
					Title:    "Home",
					Body:     "<h1>Home</h1><p>Content</p>",
					Toc: []tocEntry{
						{Id: "home", Name: "Home", Level: 1},
					},
				},
				Generator: generatorInfo{Name: "n8go-docs", Version: "test"},
				Site: manifest.SiteManifest{
					SiteName:      "Docs",
					DefaultSearch: true,
					Logo:          "assets/img/logo.svg",
				},
				Nav: []*navNode{
					{
						Name:   "Section",
						Active: true,
						Children: []*navNode{
							{Name: "Home", Url: ".", Active: true},
						},
					},
				},
				RootPath: "../",
				Now:      "now",
			})
			if err != nil {
				t.Fatal(err)
			}

			for _, want := range []string{"Docs", "<h1>Home</h1>", "../assets/img/logo.svg"} {
				if !strings.Contains(html, want) {
					t.Fatalf("rendered bundled theme does not contain %q:\n%s", want, html)
				}
			}
		})
	}
}

func TestDefaultThemeIncludesChromaStyles(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve test file path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))

	data, err := os.ReadFile(filepath.Join(repoRoot, "themes", "default", "css", "theme.css"))
	if err != nil {
		t.Fatal(err)
	}

	css := string(data)
	for _, want := range []string{".main-content .chroma", ".main-content .chroma .k", ".main-content .chroma .s"} {
		if !strings.Contains(css, want) {
			t.Fatalf("default theme is missing bundled Chroma styles %q", want)
		}
	}
}

func TestJinja2goTemplateRendersThemeContextAndFilters(t *testing.T) {
	themeDir := t.TempDir()
	writeTestFile(t, themeDir+"/"+RootTemplateName, `<!doctype html>
<html>
<body>
<a href="{{ 'assets/css/style.css'|url }}">CSS</a>
<script>var config = {{ config|tojson }};</script>
<nav>
{% for nav_item in nav %}
  <a href="{{ nav_item.url|url }}" data-active="{{ nav_item.active }}">{{ nav_item.title }}</a>
  {% for child in nav_item.children %}
    <a href="{{ child.url|url }}">{{ child.title }}</a>
  {% endfor %}
{% endfor %}
</nav>
<main>{{ page.content }}</main>
</body>
</html>`)

	tmpl, err := generateTemplate(themeDir)
	if err != nil {
		t.Fatal(err)
	}

	html, err := renderTemplateHTML(tmpl, &pageContext{
		Page: pageInfo{
			Title: "Home",
			Body:  "<h1>Home</h1>",
		},
		Site: manifest.SiteManifest{
			SiteName: "Docs",
		},
		Nav: []*navNode{
			{
				Name:   "Section",
				Active: true,
				Children: []*navNode{
					{Name: "Child", Url: "child/"},
				},
			},
		},
		RootPath: "../",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		`href="../assets/css/style.css"`,
		`"site_name":"Docs"`,
		`href="../child/" data-active="True">Section</a>`,
		`href="../child/">Child</a>`,
		`<main><h1>Home</h1></main>`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML does not contain %q:\n%s", want, html)
		}
	}
}

func TestJinja2goURLFilterUsesConfiguredSiteURL(t *testing.T) {
	themeDir := t.TempDir()
	writeTestFile(t, themeDir+"/"+RootTemplateName, `<!doctype html>
<a href="{{ 'adr/0016-versioning-and-compatibility'|url }}">ADR</a>
<img src="{{ 'assets/img/logo.svg'|url }}">
<main>{{ page.content }}</main>`)

	tmpl, err := generateTemplate(themeDir)
	if err != nil {
		t.Fatal(err)
	}

	html, err := renderTemplateHTML(tmpl, &pageContext{
		Page: pageInfo{Body: "<p>Body</p>"},
		Site: manifest.SiteManifest{
			SiteName: "Docs",
			SiteURL:  "https://108n.online/xboiler",
		},
		RootPath: "../../",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		`href="https://108n.online/xboiler/adr/0016-versioning-and-compatibility/"`,
		`src="https://108n.online/xboiler/assets/img/logo.svg"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML does not contain %q:\n%s", want, html)
		}
	}
}

func TestJinja2goTemplateSupportsDoBreakAndContinue(t *testing.T) {
	themeDir := t.TempDir()
	writeTestFile(t, themeDir+"/"+RootTemplateName, `{{ page.content }}
{% do config|tojson %}
{% for nav_item in nav %}
{% if nav_item.title == "Skip" %}{% continue %}{% endif %}
{{ nav_item.title }}
{% if nav_item.title == "Stop" %}{% break %}{% endif %}
{% endfor %}`)

	tmpl, err := generateTemplate(themeDir)
	if err != nil {
		t.Fatal(err)
	}

	html, err := renderTemplateHTML(tmpl, &pageContext{
		Site: manifest.SiteManifest{SiteName: "Docs"},
		Nav: []*navNode{
			{Name: "One", Url: "one/"},
			{Name: "Skip", Url: "skip/"},
			{Name: "Stop", Url: "stop/"},
			{Name: "After", Url: "after/"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"One", "Stop"} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML does not contain %q:\n%s", want, html)
		}
	}
	for _, unwanted := range []string{"Skip", "After"} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("rendered HTML unexpectedly contains %q:\n%s", unwanted, html)
		}
	}
}

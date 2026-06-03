package core

import (
	"strings"
	"testing"

	"n8go-docs/manifest"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func makeCtx(inputDir, pageUrl, rootPath string, index PageIndex) *pageContext {
	return &pageContext{
		InputDir: inputDir,
		Url:      pageUrl,
		RootPath: rootPath,
		Index:    index,
		Site: manifest.SiteManifest{
			InputPath: "docs",
		},
	}
}

// ── relativeUrl ──────────────────────────────────────────────────────────────

func TestRelativeUrl(t *testing.T) {
	cases := []struct {
		cur, tgt, want string
	}{
		// same page
		{".", ".", "./"},
		{"guide/page", "guide/page", "./"},
		// root → child
		{".", "components", "components/"},
		{".", "guide/page", "guide/page/"},
		// child → root
		{"guide/page", ".", "../../"},
		{"guide/sub/deep", ".", "../../../"},
		// siblings at root level
		{"components", "config", "../config/"},
		// cross-directory
		{"guide/page", "components", "../../components/"},
		{"guide/page", "guide/other", "../other/"},
		{"guide/sub/deep", "guide/page", "../../page/"},
	}

	for _, c := range cases {
		got := relativeUrl(c.cur, c.tgt)
		if got != c.want {
			t.Errorf("relativeUrl(%q, %q) = %q, want %q", c.cur, c.tgt, got, c.want)
		}
	}
}

// ── resolveHref ──────────────────────────────────────────────────────────────

func TestResolveHref_AnchorOnly(t *testing.T) {
	ctx := makeCtx("docs/guide", "guide/page", "../../", nil)
	if got := resolveHref("#section", ctx); got != "#section" {
		t.Errorf("got %q, want %q", got, "#section")
	}
}

func TestResolveHref_External(t *testing.T) {
	ctx := makeCtx("docs", ".", "", nil)
	for _, link := range []string{"https://example.com", "http://foo.bar", "mailto:user@example.com"} {
		if got := resolveHref(link, ctx); got != link {
			t.Errorf("resolveHref(%q): got %q, want unchanged", link, got)
		}
	}
}

func TestResolveHref_MdLink_KnownPage(t *testing.T) {
	index := PageIndex{
		"docs/index.md":      ".",
		"docs/components.md": "components",
		"docs/config.md":     "config",
	}

	// from root index page: links to siblings
	ctx := makeCtx("docs", ".", "", index)
	if got := resolveHref("components.md", ctx); got != "components/" {
		t.Errorf("got %q, want %q", got, "components/")
	}

	// from nested page: ../index.md resolves to root
	ctx2 := makeCtx("docs/guide", "guide/page", "../../", index)
	if got := resolveHref("../index.md", ctx2); got != "../../" {
		t.Errorf("got %q, want %q", got, "../../")
	}

	// link with anchor
	if got := resolveHref("../config.md#settings", ctx2); got != "../../config/#settings" {
		t.Errorf("got %q, want %q", got, "../../config/#settings")
	}
}

func TestResolveHref_UsesConfiguredSiteURL(t *testing.T) {
	index := PageIndex{
		"docs/index.md":      ".",
		"docs/adr/page.md":   "adr/page",
		"docs/components.md": "components",
	}
	ctx := makeCtx("docs/adr", "adr/current", "../../", index)
	ctx.Site.SiteURL = "https://108n.online/xboiler"

	if got := resolveHref("../components.md", ctx); got != "https://108n.online/xboiler/components/" {
		t.Errorf("got %q, want configured public URL", got)
	}
	if got := resolveHref("img/photo.png", ctx); got != "https://108n.online/xboiler/img/photo.png" {
		t.Errorf("got %q, want configured public static URL", got)
	}
	if got := resolveHref(".", ctx); got != "https://108n.online/xboiler/" {
		t.Errorf("got %q, want configured public homepage URL", got)
	}
}

func TestResolveHref_AbsoluteMdLink(t *testing.T) {
	index := PageIndex{"docs/guide/page.md": "guide/page"}
	ctx := makeCtx("docs/other", "other", "../", index)
	ctx.Site.InputPath = "docs"

	if got := resolveHref("/guide/page.md", ctx); got != "../guide/page/" {
		t.Errorf("got %q, want %q", got, "../guide/page/")
	}
}

func TestResolveHref_StaticResource(t *testing.T) {
	// .png is not a md link → fallback with RootPath
	ctx := makeCtx("docs/guide", "guide/page", "../../", PageIndex{})
	if got := resolveHref("img/photo.png", ctx); got != "../../img/photo.png" {
		t.Errorf("got %q, want %q", got, "../../img/photo.png")
	}
	// strip leading ../
	if got := resolveHref("../img/photo.png", ctx); got != "../../img/photo.png" {
		t.Errorf("got %q, want %q", got, "../../img/photo.png")
	}
}

func TestResolveHref_DotLink_Root(t *testing.T) {
	// "." from root page (RootPath="") should return "./"
	ctx := makeCtx("docs", ".", "", PageIndex{})
	if got := resolveHref(".", ctx); got != "./" {
		t.Errorf("got %q, want %q", got, "./")
	}
	// "." from nested page returns RootPath
	ctx2 := makeCtx("docs/guide", "guide/page", "../../", PageIndex{})
	if got := resolveHref(".", ctx2); got != "../../" {
		t.Errorf("got %q, want %q", got, "../../")
	}
}

func TestResolveHref_DirectoryLink_Fallback(t *testing.T) {
	// link without extension not in index → directory fallback
	ctx := makeCtx("docs", ".", "", PageIndex{})
	if got := resolveHref("some/dir", ctx); got != "some/dir/" {
		t.Errorf("got %q, want %q", got, "some/dir/")
	}
}

func TestResolveHref_EmptyAfterAnchorSplit(t *testing.T) {
	// href="#only-anchor" → anchor only, already handled above
	// href is just "#" with empty path
	ctx := makeCtx("docs", ".", "", PageIndex{})
	if got := resolveHref("#", ctx); got != "#" {
		t.Errorf("got %q, want %q", got, "#")
	}
}

// ── processHtml ──────────────────────────────────────────────────────────────

func TestProcessHtml_RewritesLinks(t *testing.T) {
	index := PageIndex{
		"docs/index.md":      ".",
		"docs/components.md": "components",
	}
	ctx := makeCtx("docs", ".", "", index)

	input := `<html><body>
<a href="components.md">Components</a>
<a href="https://example.com">External</a>
<a href="#anchor">Anchor</a>
<img src="img/logo.png"/>
</body></html>`

	var out strings.Builder
	err := processHtml(strings.NewReader(input), &out, ctx)
	if err != nil {
		t.Fatal(err)
	}

	result := out.String()

	if !strings.Contains(result, `href="components/"`) {
		t.Errorf("expected href=components/ in:\n%s", result)
	}
	if !strings.Contains(result, `href="https://example.com"`) {
		t.Errorf("external link should be unchanged in:\n%s", result)
	}
	if !strings.Contains(result, `href="#anchor"`) {
		t.Errorf("anchor link should be unchanged in:\n%s", result)
	}
	if !strings.Contains(result, `src="img/logo.png"`) {
		t.Errorf("img src should be rewritten with RootPath (empty here) in:\n%s", result)
	}
}

func TestProcessHtml_SrcRewrittenWithRootPath(t *testing.T) {
	ctx := makeCtx("docs/guide", "guide/page", "../../", PageIndex{})

	input := `<html><body><img src="img/logo.png"/><script src="../js/app.js"></script></body></html>`
	var out strings.Builder
	if err := processHtml(strings.NewReader(input), &out, ctx); err != nil {
		t.Fatal(err)
	}
	result := out.String()

	if !strings.Contains(result, `src="../../img/logo.png"`) {
		t.Errorf("expected ../../img/logo.png in:\n%s", result)
	}
	if !strings.Contains(result, `src="../../js/app.js"`) {
		t.Errorf("expected ../../js/app.js (stripped ../) in:\n%s", result)
	}
}

func TestProcessHtml_RewritesLocalLinksWithConfiguredSiteURL(t *testing.T) {
	ctx := makeCtx("docs/adr", "adr/current", "../../", PageIndex{
		"docs/adr/target.md": "adr/target",
	})
	ctx.Site.SiteURL = "https://108n.online/xboiler"
	input := `<html><body><a href="target.md">Target</a><img src="img/logo.png"/></body></html>`
	var out strings.Builder

	if err := processHtml(strings.NewReader(input), &out, ctx); err != nil {
		t.Fatal(err)
	}
	result := out.String()
	if !strings.Contains(result, `href="https://108n.online/xboiler/adr/target/"`) {
		t.Errorf("expected public href in:\n%s", result)
	}
	if !strings.Contains(result, `src="https://108n.online/xboiler/img/logo.png"`) {
		t.Errorf("expected public src in:\n%s", result)
	}
}

func TestProcessHtml_ExtraCssJs(t *testing.T) {
	// Simulate what the template emits for extra_css / extra_javascript.
	// The paths are relative to the site root; postproc adds RootPath via fallback.
	ctx := makeCtx("docs/user-guide", "user-guide/installation", "../../", PageIndex{})

	input := `<html><head>
<link rel="stylesheet" href="css/custom.css">
<script src="js/analytics.js" defer=""></script>
</head><body></body></html>`

	var out strings.Builder
	if err := processHtml(strings.NewReader(input), &out, ctx); err != nil {
		t.Fatal(err)
	}
	result := out.String()

	if !strings.Contains(result, `href="../../css/custom.css"`) {
		t.Errorf("expected ../../css/custom.css in:\n%s", result)
	}
	if !strings.Contains(result, `src="../../js/analytics.js"`) {
		t.Errorf("expected ../../js/analytics.js in:\n%s", result)
	}
}

func TestProcessHtml_NestedPageLinks(t *testing.T) {
	index := PageIndex{
		"docs/index.md":  ".",
		"docs/config.md": "config",
	}
	ctx := makeCtx("docs/guide", "guide/page", "../../", index)

	input := `<html><body>
<a href="../index.md">Home</a>
<a href="../config.md#section">Config</a>
</body></html>`

	var out strings.Builder
	if err := processHtml(strings.NewReader(input), &out, ctx); err != nil {
		t.Fatal(err)
	}
	result := out.String()

	if !strings.Contains(result, `href="../../"`) {
		t.Errorf("expected href=../../ for index in:\n%s", result)
	}
	if !strings.Contains(result, `href="../../config/#section"`) {
		t.Errorf("expected href=../../config/#section in:\n%s", result)
	}
}

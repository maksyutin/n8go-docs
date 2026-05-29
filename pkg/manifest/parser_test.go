package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

// ── ParseSiteManifest: YAML keys ─────────────────────────────────────────────

// Regression: generator used to silently read "output" key instead of "site_dir",
// falling back to the default "site/" while deploy scripts expected "sites/".
func TestParseSiteManifest_SiteDirKey(t *testing.T) {
	path := writeTempConfig(t, "name: s\nsite_dir: sites\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean(filepath.Join(filepath.Dir(path), "sites"))
	if m.OutputPath != want {
		t.Errorf("OutputPath: got %q, want %q — check that parser reads site_dir, not output", m.OutputPath, want)
	}
}

// Regression: old "output" key must NOT be silently accepted.
func TestParseSiteManifest_OutputKeyIgnored(t *testing.T) {
	path := writeTempConfig(t, "name: s\noutput: custom_out\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	// "output" is unknown — parser should fall back to default "site", not "custom_out"
	defaultOut := filepath.Clean(filepath.Join(filepath.Dir(path), "site"))
	if m.OutputPath == filepath.Clean(filepath.Join(filepath.Dir(path), "custom_out")) {
		t.Errorf("OutputPath: parser still reads deprecated key \"output\"; use \"site_dir\" instead")
	}
	if m.OutputPath != defaultOut {
		t.Errorf("OutputPath: got %q, want default %q", m.OutputPath, defaultOut)
	}
}

func TestParseSiteManifest_DocsDirKey(t *testing.T) {
	path := writeTempConfig(t, "name: s\ndocs_dir: content\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean(filepath.Join(filepath.Dir(path), "content"))
	if m.InputPath != want {
		t.Errorf("InputPath: got %q, want %q — check that parser reads docs_dir, not input", m.InputPath, want)
	}
}

// Regression: old "input" key must NOT be silently accepted.
func TestParseSiteManifest_InputKeyIgnored(t *testing.T) {
	path := writeTempConfig(t, "name: s\ninput: custom_in\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	defaultIn := filepath.Clean(filepath.Join(filepath.Dir(path), "docs"))
	if m.InputPath == filepath.Clean(filepath.Join(filepath.Dir(path), "custom_in")) {
		t.Errorf("InputPath: parser still reads deprecated key \"input\"; use \"docs_dir\" instead")
	}
	if m.InputPath != defaultIn {
		t.Errorf("InputPath: got %q, want default %q", m.InputPath, defaultIn)
	}
}

// ── ParseSiteManifest: defaults ──────────────────────────────────────────────

func TestParseSiteManifest_Defaults(t *testing.T) {
	path := writeTempConfig(t, "name: site\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)

	cases := []struct {
		field string
		got   string
		want  string
	}{
		{"ThemeId", m.ThemeId, "default"},
		{"InputPath", m.InputPath, filepath.Clean(filepath.Join(dir, "docs"))},
		{"OutputPath", m.OutputPath, filepath.Clean(filepath.Join(dir, "site"))},
		{"SearchEngine", m.SearchEngine, "flexsearch"},
		{"Logo", m.Logo, "img/book.svg"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s default: got %q, want %q", c.field, c.got, c.want)
		}
	}

	if !m.DefaultSearch {
		t.Error("DefaultSearch default: want true")
	}
	if m.SearchContentLimit != 500 {
		t.Errorf("SearchContentLimit default: got %d, want 500", m.SearchContentLimit)
	}
	if m.StripMdExtension {
		t.Error("StripMdExtension default: want false")
	}
}

// ── ParseSiteManifest: path resolution ──────────────────────────────────────

// Paths must be resolved relative to the config file, not os.Getwd().
func TestParseSiteManifest_PathsRelativeToConfig(t *testing.T) {
	path := writeTempConfig(t, "name: s\ndocs_dir: ./src\nsite_dir: ../build\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)
	wantIn := filepath.Clean(filepath.Join(dir, "src"))
	wantOut := filepath.Clean(filepath.Join(dir, "../build"))
	if m.InputPath != wantIn {
		t.Errorf("InputPath: got %q, want %q", m.InputPath, wantIn)
	}
	if m.OutputPath != wantOut {
		t.Errorf("OutputPath: got %q, want %q", m.OutputPath, wantOut)
	}
}

// ── ParseSiteManifest: required fields ──────────────────────────────────────

func TestParseSiteManifest_MissingName(t *testing.T) {
	path := writeTempConfig(t, "theme: default\n")
	_, err := ParseSiteManifest(path)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

func TestParseSiteManifest_FileNotFound(t *testing.T) {
	_, err := ParseSiteManifest("/nonexistent/n8go-docs.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestParseSiteManifest_InvalidYaml(t *testing.T) {
	path := writeTempConfig(t, "name: [invalid\n")
	_, err := ParseSiteManifest(path)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

// ── ParseSiteManifest: all fields ────────────────────────────────────────────

func TestParseSiteManifest_AllFields(t *testing.T) {
	yaml := `
name: my-site
theme: custom
docs_dir: src
site_dir: out
default_search: false
search_engine: fuse
search_content_limit: 200
custom_font: Roboto
logo: img/logo.svg
strip_md_extension: true
head_tags:
  - '<meta name="foo" content="bar">'
extra_css:
  - css/custom.css
  - assets/style.css
extra_javascript:
  - js/analytics.js
`
	path := writeTempConfig(t, yaml)
	dir := filepath.Dir(path)
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}

	if m.Name != "my-site" {
		t.Errorf("Name: got %q", m.Name)
	}
	if m.ThemeId != "custom" {
		t.Errorf("ThemeId: got %q", m.ThemeId)
	}
	if m.InputPath != filepath.Clean(filepath.Join(dir, "src")) {
		t.Errorf("InputPath: got %q", m.InputPath)
	}
	if m.OutputPath != filepath.Clean(filepath.Join(dir, "out")) {
		t.Errorf("OutputPath: got %q", m.OutputPath)
	}
	if m.DefaultSearch {
		t.Error("DefaultSearch: want false")
	}
	if m.SearchEngine != "fuse" {
		t.Errorf("SearchEngine: got %q", m.SearchEngine)
	}
	if m.SearchContentLimit != 200 {
		t.Errorf("SearchContentLimit: got %d", m.SearchContentLimit)
	}
	if m.CustomFont != "Roboto" {
		t.Errorf("CustomFont: got %q", m.CustomFont)
	}
	if m.Logo != "img/logo.svg" {
		t.Errorf("Logo: got %q", m.Logo)
	}
	if !m.StripMdExtension {
		t.Error("StripMdExtension: want true")
	}
	if len(m.HeadTags) != 1 || m.HeadTags[0] != `<meta name="foo" content="bar">` {
		t.Errorf("HeadTags: got %v", m.HeadTags)
	}
	if len(m.ExtraCss) != 2 || m.ExtraCss[0] != "css/custom.css" {
		t.Errorf("ExtraCss: got %v", m.ExtraCss)
	}
	if len(m.ExtraJavascript) != 1 || m.ExtraJavascript[0] != "js/analytics.js" {
		t.Errorf("ExtraJavascript: got %v", m.ExtraJavascript)
	}
}

// ── ParseSiteManifest: nav ───────────────────────────────────────────────────

func TestParseSiteManifest_NavFlat(t *testing.T) {
	path := writeTempConfig(t, "name: s\nnav:\n  - Home: index.md\n  - About: about.md\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Nav) != 2 {
		t.Fatalf("Nav len: got %d, want 2", len(m.Nav))
	}
	if m.Nav[0].Title != "Home" || m.Nav[0].File != "index.md" {
		t.Errorf("Nav[0]: got %+v", m.Nav[0])
	}
	if m.Nav[1].Title != "About" || m.Nav[1].File != "about.md" {
		t.Errorf("Nav[1]: got %+v", m.Nav[1])
	}
}

func TestParseSiteManifest_NavNested(t *testing.T) {
	content := `
name: s
nav:
  - Home: index.md
  - Guide:
      - Start: guide/start.md
      - Advanced:
          - Plugins: guide/advanced/plugins.md
`
	path := writeTempConfig(t, content)
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Nav) != 2 {
		t.Fatalf("Nav len: got %d, want 2", len(m.Nav))
	}
	guide := m.Nav[1]
	if guide.Title != "Guide" || guide.File != "" || len(guide.Children) != 2 {
		t.Errorf("Guide: got %+v", guide)
	}
	advanced := guide.Children[1]
	if len(advanced.Children) != 1 || advanced.Children[0].File != "guide/advanced/plugins.md" {
		t.Errorf("Advanced.Children[0]: got %+v", advanced.Children)
	}
}

func TestParseSiteManifest_NavEmpty(t *testing.T) {
	path := writeTempConfig(t, "name: s\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Nav) != 0 {
		t.Errorf("Nav: expected empty, got %v", m.Nav)
	}
}

// ── ParseSiteManifest: exclude_docs ─────────────────────────────────────────

func TestParseSiteManifest_ExcludeDocs(t *testing.T) {
	content := "name: s\nexclude_docs: |\n  draft.md\n  private/*\n  **/secret-*.md\n"
	path := writeTempConfig(t, content)
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"draft.md", "private/*", "**/secret-*.md"}
	if len(m.ExcludeDocs) != len(want) {
		t.Fatalf("ExcludeDocs: got %v, want %v", m.ExcludeDocs, want)
	}
	for i, w := range want {
		if m.ExcludeDocs[i] != w {
			t.Errorf("ExcludeDocs[%d]: got %q, want %q", i, m.ExcludeDocs[i], w)
		}
	}
}

func TestParseSiteManifest_ExcludeDocsStripsComments(t *testing.T) {
	content := "name: s\nexclude_docs: |\n  draft.md   # черновик\n  private/*\n"
	path := writeTempConfig(t, content)
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.ExcludeDocs) != 2 || m.ExcludeDocs[0] != "draft.md" {
		t.Errorf("ExcludeDocs: got %v", m.ExcludeDocs)
	}
}

func TestParseSiteManifest_ExcludeDocsList(t *testing.T) {
	content := "name: s\nexclude_docs:\n  - draft.md\n  - private/*\n  - \"**/secret-*.md\"\n"
	path := writeTempConfig(t, content)
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"draft.md", "private/*", "**/secret-*.md"}
	if len(m.ExcludeDocs) != len(want) {
		t.Fatalf("ExcludeDocs: got %v, want %v", m.ExcludeDocs, want)
	}
	for i, w := range want {
		if m.ExcludeDocs[i] != w {
			t.Errorf("ExcludeDocs[%d]: got %q, want %q", i, m.ExcludeDocs[i], w)
		}
	}
}

func TestParseSiteManifest_ExcludeDocsListStripsComments(t *testing.T) {
	content := "name: s\nexclude_docs:\n  - draft.md   # черновик\n  - private/*\n"
	path := writeTempConfig(t, content)
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"draft.md", "private/*"}
	if len(m.ExcludeDocs) != len(want) {
		t.Fatalf("ExcludeDocs: got %v, want %v", m.ExcludeDocs, want)
	}
	for i, w := range want {
		if m.ExcludeDocs[i] != w {
			t.Errorf("ExcludeDocs[%d]: got %q, want %q", i, m.ExcludeDocs[i], w)
		}
	}
}

// ── ParseThemeManifest ───────────────────────────────────────────────────────

func TestParseThemeManifest_Full(t *testing.T) {
	yaml := `
theme:
  name: MyTheme
  description: A test theme
  repository: https://github.com/test/theme
  version: 2.0.0
  author: Tester
  license: Apache-2.0
highlighting:
  style: monokai
  line_numbers: true
`
	path := writeTempConfig(t, yaml)
	m, err := ParseThemeManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "MyTheme" {
		t.Errorf("Name: got %q", m.Name)
	}
	if m.Version != "2.0.0" {
		t.Errorf("Version: got %q", m.Version)
	}
	if m.Highlighting.Style != "monokai" {
		t.Errorf("Style: got %q", m.Highlighting.Style)
	}
	if !m.Highlighting.LineNumbers {
		t.Error("LineNumbers: want true")
	}
}

func TestParseThemeManifest_DefaultHighlightingStyle(t *testing.T) {
	path := writeTempConfig(t, "theme:\n  name: T\n  version: 1.0.0\n")
	m, err := ParseThemeManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Highlighting.Style != "bw" {
		t.Errorf("Style default: got %q, want %q", m.Highlighting.Style, "bw")
	}
}

func TestParseThemeManifest_MissingName(t *testing.T) {
	path := writeTempConfig(t, "theme:\n  version: 1.0.0\nhighlighting:\n  style: bw\n")
	_, err := ParseThemeManifest(path)
	if err == nil {
		t.Error("expected error for missing theme name, got nil")
	}
}

func TestParseThemeManifest_MissingVersion(t *testing.T) {
	path := writeTempConfig(t, "theme:\n  name: T\nhighlighting:\n  style: bw\n")
	_, err := ParseThemeManifest(path)
	if err == nil {
		t.Error("expected error for missing theme version, got nil")
	}
}

func TestParseThemeManifest_FileNotFound(t *testing.T) {
	_, err := ParseThemeManifest("/nonexistent/theme.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func TestBoolOr(t *testing.T) {
	tr, fa := true, false
	if !boolOr(&tr, false) {
		t.Error("boolOr(&true, false): want true")
	}
	if boolOr(&fa, true) {
		t.Error("boolOr(&false, true): want false")
	}
	if !boolOr(nil, true) {
		t.Error("boolOr(nil, true): want true")
	}
	if boolOr(nil, false) {
		t.Error("boolOr(nil, false): want false")
	}
}

func TestIntOr(t *testing.T) {
	n := 42
	if intOr(&n, 0) != 42 {
		t.Error("intOr(&42, 0): want 42")
	}
	if intOr(nil, 99) != 99 {
		t.Error("intOr(nil, 99): want 99")
	}
}

func TestStrOr(t *testing.T) {
	if strOr("hello", "default") != "hello" {
		t.Error("strOr(non-empty): want original")
	}
	if strOr("", "default") != "default" {
		t.Error("strOr(empty): want default")
	}
}

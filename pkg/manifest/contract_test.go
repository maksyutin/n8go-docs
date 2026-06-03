// contract_test.go verifies the contract between the YAML config schema and
// the parsed SiteManifest / ThemeManifest structs. Each test documents the
// regression it prevents; the grouping mirrors the prompt spec.
package manifest

import (
	"path/filepath"
	"testing"
)

// ── YAML keys ────────────────────────────────────────────────────────────────

// Prevents: renaming the YAML key "site_description" without updating the parser.
func TestContract_SiteDescriptionKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nsite_description: My docs\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.SiteDescription != "My docs" {
		t.Errorf("SiteDescription: got %q, want %q — parser must read site_description", m.SiteDescription, "My docs")
	}
}

// Prevents: renaming the YAML key "dev_addr" without updating the parser.
func TestContract_DevAddrKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\ndev_addr: 127.0.0.1:3000\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.DevAddr != "127.0.0.1:3000" {
		t.Errorf("DevAddr: got %q, want %q — parser must read dev_addr", m.DevAddr, "127.0.0.1:3000")
	}
}

// Prevents: renaming the YAML key "use_directory_urls" without updating the parser.
func TestContract_UseDirectoryURLsKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nuse_directory_urls: false\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.UseDirectoryURLs {
		t.Error("UseDirectoryURLs: got true, want false — parser must read use_directory_urls")
	}
}

// Prevents: renaming the YAML key "theme" without updating the parser.
func TestContract_ThemeKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\ntheme: material\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.ThemeId != "material" {
		t.Errorf("ThemeId: got %q, want %q — parser must read theme", m.ThemeId, "material")
	}
}

// Prevents: renaming the YAML key "default_search" without updating the parser.
func TestContract_DefaultSearchKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\ndefault_search: false\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.DefaultSearch {
		t.Error("DefaultSearch: got true, want false — parser must read default_search")
	}
}

// Prevents: renaming the YAML key "search_engine" without updating the parser.
func TestContract_SearchEngineKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nsearch_engine: fuse\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.SearchEngine != "fuse" {
		t.Errorf("SearchEngine: got %q, want %q — parser must read search_engine", m.SearchEngine, "fuse")
	}
}

// Prevents: renaming the YAML key "search_content_limit" without updating the parser.
func TestContract_SearchContentLimitKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nsearch_content_limit: 250\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.SearchContentLimit != 250 {
		t.Errorf("SearchContentLimit: got %d, want 250 — parser must read search_content_limit", m.SearchContentLimit)
	}
}

// Prevents: renaming the YAML key "strip_md_extension" without updating the parser.
func TestContract_StripMdExtensionKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nstrip_md_extension: true\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if !m.StripMdExtension {
		t.Error("StripMdExtension: got false, want true — parser must read strip_md_extension")
	}
}

// Prevents: renaming the YAML key "custom_font" without updating the parser.
func TestContract_CustomFontKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\ncustom_font: Inter\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.CustomFont != "Inter" {
		t.Errorf("CustomFont: got %q, want %q — parser must read custom_font", m.CustomFont, "Inter")
	}
}

// Prevents: renaming the YAML key "logo" without updating the parser.
func TestContract_LogoKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nlogo: img/brand.svg\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Logo != "img/brand.svg" {
		t.Errorf("Logo: got %q, want %q — parser must read logo", m.Logo, "img/brand.svg")
	}
}

// Prevents: renaming the YAML key "head_tags" without updating the parser.
func TestContract_HeadTagsKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nhead_tags:\n  - '<meta name=\"og:type\" content=\"website\">'\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.HeadTags) != 1 {
		t.Fatalf("HeadTags: got %d entries, want 1 — parser must read head_tags", len(m.HeadTags))
	}
}

// Prevents: renaming the YAML key "extra_css" without updating the parser.
func TestContract_ExtraCssKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nextra_css:\n  - css/custom.css\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.ExtraCss) != 1 || m.ExtraCss[0] != "css/custom.css" {
		t.Errorf("ExtraCss: got %v, want [css/custom.css] — parser must read extra_css", m.ExtraCss)
	}
}

// Prevents: renaming the YAML key "extra_javascript" without updating the parser.
func TestContract_ExtraJavascriptKey(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nextra_javascript:\n  - js/analytics.js\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.ExtraJavascript) != 1 || m.ExtraJavascript[0] != "js/analytics.js" {
		t.Errorf("ExtraJavascript: got %v, want [js/analytics.js] — parser must read extra_javascript", m.ExtraJavascript)
	}
}

// Prevents: renaming highlighting.style in the theme YAML schema.
func TestContract_ThemeHighlightingStyleKey(t *testing.T) {
	path := writeTempConfig(t, "theme:\n  name: T\n  version: 1.0.0\nhighlighting:\n  style: dracula\n")
	m, err := ParseThemeManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Highlighting.Style != "dracula" {
		t.Errorf("Highlighting.Style: got %q, want %q — parser must read highlighting.style", m.Highlighting.Style, "dracula")
	}
}

// Prevents: renaming highlighting.line_numbers in the theme YAML schema.
func TestContract_ThemeHighlightingLineNumbersKey(t *testing.T) {
	path := writeTempConfig(t, "theme:\n  name: T\n  version: 1.0.0\nhighlighting:\n  style: bw\n  line_numbers: true\n")
	m, err := ParseThemeManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if !m.Highlighting.LineNumbers {
		t.Error("Highlighting.LineNumbers: got false, want true — parser must read highlighting.line_numbers")
	}
}

// ── defaults ─────────────────────────────────────────────────────────────────

// Prevents: changing the default output directory from "site" to anything else
// (the exact regression described in the bug report).
func TestContract_DefaultOutputDirIsSite(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean(filepath.Join(filepath.Dir(path), "site"))
	if m.OutputPath != want {
		t.Errorf("OutputPath default: got %q, want %q — changing the default output dir is a breaking change; update docs and deploy scripts first", m.OutputPath, want)
	}
}

// Prevents: changing the default input directory from "docs".
func TestContract_DefaultInputDirIsDocs(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean(filepath.Join(filepath.Dir(path), "docs"))
	if m.InputPath != want {
		t.Errorf("InputPath default: got %q, want %q — changing default docs_dir breaks projects that omit it", m.InputPath, want)
	}
}

// Prevents: changing the default theme from "default".
func TestContract_DefaultThemeIsDefault(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.ThemeId != "default" {
		t.Errorf("ThemeId default: got %q, want %q", m.ThemeId, "default")
	}
}

// Prevents: changing the default search engine from "flexsearch" to something else.
func TestContract_DefaultSearchEngineIsFlexsearch(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.SearchEngine != "flexsearch" {
		t.Errorf("SearchEngine default: got %q, want %q — client-side JS expects flexsearch unless explicitly configured", m.SearchEngine, "flexsearch")
	}
}

// Prevents: changing the default search content limit from 500.
func TestContract_DefaultSearchContentLimitIs500(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.SearchContentLimit != 500 {
		t.Errorf("SearchContentLimit default: got %d, want 500 — raising this increases index size; lowering drops content silently", m.SearchContentLimit)
	}
}

// Prevents: flipping the default for default_search to false (search index stops generating).
func TestContract_DefaultSearchIsEnabled(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if !m.DefaultSearch {
		t.Error("DefaultSearch default: got false, want true — disabling search by default would break existing sites silently")
	}
}

// Prevents: enabling strip_md_extension by default (changes all output URLs).
func TestContract_DefaultStripMdExtensionIsFalse(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.StripMdExtension {
		t.Error("StripMdExtension default: got true, want false — enabling it by default would rewrite all links in existing sites")
	}
}

// Prevents: disabling use_directory_urls by default (all page paths change).
func TestContract_DefaultUseDirectoryURLsIsTrue(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if !m.UseDirectoryURLs {
		t.Error("UseDirectoryURLs default: got false, want true — disabling it changes every page URL in the generated site")
	}
}

// Prevents: changing the default logo path (breaks themes that rely on the default).
func TestContract_DefaultLogoPath(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	const want = "assets/img/logo.svg"
	if m.Logo != want {
		t.Errorf("Logo default: got %q, want %q — themes reference this path; change it in the theme templates first", m.Logo, want)
	}
}

// Prevents: changing the default highlighting style from "bw".
func TestContract_DefaultThemeHighlightingStyleIsBw(t *testing.T) {
	path := writeTempConfig(t, "theme:\n  name: T\n  version: 1.0.0\n")
	m, err := ParseThemeManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Highlighting.Style != "bw" {
		t.Errorf("Highlighting.Style default: got %q, want %q — changing default style affects all themes that omit highlighting config", m.Highlighting.Style, "bw")
	}
}

// ── path resolution ──────────────────────────────────────────────────────────

// Prevents: resolving docs_dir relative to os.Getwd() instead of the config file.
// Symptom: works when run from project root, fails from any other directory.
func TestContract_DocsDirRelativeToConfigFile(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\ndocs_dir: content\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean(filepath.Join(filepath.Dir(path), "content"))
	if m.InputPath != want {
		t.Errorf("InputPath: got %q, want %q — path must be resolved relative to the config file, not cwd", m.InputPath, want)
	}
}

// Prevents: resolving site_dir relative to os.Getwd() instead of the config file.
// Symptom: deploy picks up the wrong output directory when CI runs from a subdirectory.
func TestContract_SiteDirRelativeToConfigFile(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nsite_dir: public\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean(filepath.Join(filepath.Dir(path), "public"))
	if m.OutputPath != want {
		t.Errorf("OutputPath: got %q, want %q — path must be resolved relative to the config file, not cwd", m.OutputPath, want)
	}
}

// Prevents: losing the ".." traversal when resolving paths outside the config directory.
func TestContract_SiteDirParentTraversal(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nsite_dir: ../dist\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean(filepath.Join(filepath.Dir(path), "../dist"))
	if m.OutputPath != want {
		t.Errorf("OutputPath: got %q, want %q — parent-relative paths must be preserved", m.OutputPath, want)
	}
}

// ── deprecated / renamed keys ────────────────────────────────────────────────

// Prevents: silently accepting the old "output" key for the output directory.
// Root cause of the production bug: config used "output: sites" but parser only
// reads "site_dir", so it fell back to default "site" with no warning.
func TestContract_DeprecatedOutputKeyIsIgnored(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\noutput: custom_out\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	defaultOut := filepath.Clean(filepath.Join(filepath.Dir(path), "site"))
	customOut := filepath.Clean(filepath.Join(filepath.Dir(path), "custom_out"))
	if m.OutputPath == customOut {
		t.Errorf("OutputPath: deprecated key \"output\" is being read; parser must only accept \"site_dir\"")
	}
	if m.OutputPath != defaultOut {
		t.Errorf("OutputPath: got %q, want default %q", m.OutputPath, defaultOut)
	}
}

// Prevents: silently accepting the old "input" key for the docs directory.
func TestContract_DeprecatedInputKeyIsIgnored(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\ninput: custom_in\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	defaultIn := filepath.Clean(filepath.Join(filepath.Dir(path), "docs"))
	customIn := filepath.Clean(filepath.Join(filepath.Dir(path), "custom_in"))
	if m.InputPath == customIn {
		t.Errorf("InputPath: deprecated key \"input\" is being read; parser must only accept \"docs_dir\"")
	}
	if m.InputPath != defaultIn {
		t.Errorf("InputPath: got %q, want default %q", m.InputPath, defaultIn)
	}
}

// Prevents: silently accepting the old "name" key instead of "site_name".
// If this were accepted the site_name requirement becomes bypassable via a legacy key.
func TestContract_DeprecatedNameKeyIsIgnored(t *testing.T) {
	// "name: s" alone must not satisfy the required site_name check
	_, err := ParseSiteManifest(writeTempConfig(t, "name: s\n"))
	if err == nil {
		t.Error("deprecated key \"name\" must not satisfy site_name requirement; parser must only accept \"site_name\"")
	}
}

// Prevents: silently accepting the old "url" key instead of "site_url".
func TestContract_DeprecatedUrlKeyIsIgnored(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nurl: https://legacy.example/docs\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.SiteURL != "" {
		t.Errorf("SiteURL: deprecated key \"url\" was accepted, got %q; parser must only accept \"site_url\"", m.SiteURL)
	}
}

// Prevents: silently accepting "docs" as a YAML key when "docs_dir" is expected.
func TestContract_DeprecatedDocsKeyIsIgnored(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\ndocs: custom_docs\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	defaultIn := filepath.Clean(filepath.Join(filepath.Dir(path), "docs"))
	customIn := filepath.Clean(filepath.Join(filepath.Dir(path), "custom_docs"))
	if m.InputPath == customIn {
		t.Errorf("InputPath: deprecated key \"docs\" is being read; parser must only accept \"docs_dir\"")
	}
	if m.InputPath != defaultIn {
		t.Errorf("InputPath: got %q, want default %q", m.InputPath, defaultIn)
	}
}

// Prevents: silently accepting "site" as a YAML key when "site_dir" is expected.
func TestContract_DeprecatedSiteKeyIsIgnored(t *testing.T) {
	path := writeTempConfig(t, "site_name: s\nsite: custom_site\n")
	m, err := ParseSiteManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	defaultOut := filepath.Clean(filepath.Join(filepath.Dir(path), "site"))
	customOut := filepath.Clean(filepath.Join(filepath.Dir(path), "custom_site"))
	if m.OutputPath == customOut {
		t.Errorf("OutputPath: YAML key \"site\" is being read; parser must only accept \"site_dir\"")
	}
	if m.OutputPath != defaultOut {
		t.Errorf("OutputPath: got %q, want default %q", m.OutputPath, defaultOut)
	}
}

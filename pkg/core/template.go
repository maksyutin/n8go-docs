package core

import (
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/nikolalohinski/gonja/v2/loaders"
)

var registerJinja2goFiltersOnce sync.Once

var hardCodedThemeAssetPattern = regexp.MustCompile(`(?i)\b(?:href|src)\s*=\s*["'](?:\.\./|\.\/|/)*(?:(?:assets/)?(?:css|js|img))/`)

type Jinja2goTemplate struct {
	template *exec.Template
}

func generateTemplate(themeDir string) (*Jinja2goTemplate, error) {
	mainTemplate := filepath.Join(themeDir, RootTemplateName)
	if stat, err := os.Stat(mainTemplate); err != nil {
		return nil, fmt.Errorf("theme must provide required %s: %w", RootTemplateName, err)
	} else if stat.IsDir() {
		return nil, fmt.Errorf("theme must provide required %s as a file", RootTemplateName)
	}
	if err := validateJinja2goTheme(themeDir); err != nil {
		return nil, err
	}
	registerJinja2goFilters()

	loader, err := loaders.NewFileSystemLoader(themeDir)
	if err != nil {
		return nil, err
	}

	tmpl, err := exec.NewTemplate(RootTemplateName, gonja.DefaultConfig, loader, gonja.DefaultEnvironment)
	if err != nil {
		return nil, fmt.Errorf("theme must provide a valid %s Jinja2 template: %w", RootTemplateName, err)
	}

	return &Jinja2goTemplate{template: tmpl}, nil
}

func validateJinja2goTheme(themeDir string) error {
	hasPageContent := false
	err := filepath.WalkDir(themeDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(contentBytes)
		relPath, err := filepath.Rel(themeDir, path)
		if err != nil {
			relPath = path
		}

		if strings.Contains(content, "{{.") || strings.Contains(content, "{{ .") || strings.Contains(content, "{{ range") || strings.Contains(content, "{{ if") {
			return fmt.Errorf("theme template %s uses legacy Go template syntax; use Jinja2-compatible syntax", relPath)
		}
		if hardCodedThemeAssetPattern.MatchString(content) {
			return fmt.Errorf("theme template %s hard-codes local static asset paths; use the url filter, for example {{ 'img/logo.png'|url }}", relPath)
		}
		if strings.Contains(content, "page.content") {
			hasPageContent = true
		}

		return nil
	})
	if err != nil {
		return err
	}
	if !hasPageContent {
		return fmt.Errorf("theme must render page.content in %s or an included template", RootTemplateName)
	}
	return nil
}

func (tmpl *Jinja2goTemplate) Execute(w io.Writer, ctx *pageContext) error {
	return tmpl.template.Execute(w, exec.NewContext(jinja2goContext(ctx)))
}

func registerJinja2goFilters() {
	registerJinja2goFiltersOnce.Do(func() {
		_ = gonja.DefaultEnvironment.Filters.Register("url", jinja2goURLFilter)
		_ = gonja.DefaultEnvironment.Filters.Register("urlquery", jinja2goURLQueryFilter)
		_ = gonja.DefaultEnvironment.ControlStructures.Register("do", jinja2goDoParser)
		_ = gonja.DefaultEnvironment.ControlStructures.Register("break", jinja2goBreakParser)
		_ = gonja.DefaultEnvironment.ControlStructures.Register("continue", jinja2goContinueParser)
		_ = gonja.DefaultEnvironment.ControlStructures.Replace("for", jinja2goForParser)
	})
}

func jinja2goURLFilter(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if in.IsError() {
		return in
	}
	if err := params.Take(); err != nil {
		return exec.AsValue(exec.ErrInvalidCall(err))
	}

	rootPath := ""
	if rawRootPath, ok := e.Environment.Context.Get("root_path"); ok {
		if root, ok := rawRootPath.(string); ok {
			rootPath = root
		}
	}
	if rawSiteURL, ok := e.Environment.Context.Get("site_url"); ok {
		if siteURL, ok := rawSiteURL.(string); ok && siteURL != "" {
			return exec.AsSafeValue(resolvePublicURL(siteURL, in.String()))
		}
	}

	return exec.AsSafeValue(resolveTemplateURL(in.String(), rootPath))
}

func jinja2goURLQueryFilter(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if in.IsError() {
		return in
	}
	if err := params.Take(); err != nil {
		return exec.AsValue(exec.ErrInvalidCall(err))
	}

	return exec.AsSafeValue(url.QueryEscape(in.String()))
}

func resolveTemplateURL(rawURL string, rootPath string) string {
	if isAnchorOnly(rawURL) || isExternalLink(rawURL) {
		return rawURL
	}

	hasTrailingSlash := strings.HasSuffix(rawURL, "/") && rawURL != "/"
	url := filepath.ToSlash(filepath.Clean(rawURL))
	if url == "." {
		if rootPath == "" {
			return "./"
		}
		return rootPath
	}
	for strings.HasPrefix(url, "../") {
		url = strings.TrimPrefix(url, "../")
	}
	url = strings.TrimPrefix(url, "/")
	if hasTrailingSlash && !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return rootPath + url
}

func jinja2goContext(ctx *pageContext) map[string]any {
	baseURL := ctx.RootPath
	homepageURL := ctx.RootPath
	if ctx.Site.SiteURL != "" {
		baseURL = ctx.Site.SiteURL
		homepageURL = resolvePublicURL(ctx.Site.SiteURL, ".")
	}

	config := map[string]any{
		"site_name":            ctx.Site.SiteName,
		"site_url":             ctx.Site.SiteURL,
		"site_description":     ctx.Site.SiteDescription,
		"dev_addr":             ctx.Site.DevAddr,
		"use_directory_urls":   ctx.Site.UseDirectoryURLs,
		"theme":                ctx.Site.ThemeId,
		"docs_dir":             ctx.Site.InputPath,
		"site_dir":             ctx.Site.OutputPath,
		"default_search":       ctx.Site.DefaultSearch,
		"search_engine":        ctx.Site.SearchEngine,
		"search_content_limit": ctx.Site.SearchContentLimit,
		"head_tags":            ctx.Site.HeadTags,
		"custom_font":          ctx.Site.CustomFont,
		"logo":                 ctx.Site.Logo,
		"strip_md_extension":   ctx.Site.StripMdExtension,
		"extra_css":            ctx.Site.ExtraCss,
		"extra_javascript":     ctx.Site.ExtraJavascript,
		"exclude_docs":         ctx.Site.ExcludeDocs,
	}

	page := map[string]any{
		"title":     ctx.Page.Title,
		"content":   exec.AsSafeValue(ctx.Page.Body),
		"body":      exec.AsSafeValue(ctx.Page.Body),
		"file_path": ctx.Page.FilePath,
		"file_name": ctx.Page.FileName,
		"toc":       jinja2goToc(ctx.Page.Toc),
		"url":       ctx.Url,
	}

	return map[string]any{
		"config":       config,
		"nav":          jinja2goNav(ctx.Nav),
		"page":         page,
		"page_title":   ctx.Page.Title,
		"base_url":     baseURL,
		"root_path":    ctx.RootPath,
		"site_url":     ctx.Site.SiteURL,
		"homepage_url": homepageURL,
		"generator": map[string]any{
			"name":    ctx.Generator.Name,
			"version": ctx.Generator.Version,
		},
		"now":       ctx.Now,
		"url":       ctx.Url,
		"input_dir": ctx.InputDir,

		// Compatibility aliases for existing project vocabulary.
		"site": config,
	}
}

func jinja2goNav(nodes []*navNode) []map[string]any {
	result := make([]map[string]any, 0, len(nodes))
	for _, node := range nodes {
		children := jinja2goNav(node.Children)
		nodeURL := node.Url
		if nodeURL == "" && len(children) > 0 {
			if firstChildURL, ok := children[0]["url"].(string); ok {
				nodeURL = firstChildURL
			}
		}
		result = append(result, map[string]any{
			"title":    node.Name,
			"name":     node.Name,
			"url":      nodeURL,
			"active":   node.Active,
			"children": children,
		})
	}
	return result
}

func jinja2goToc(entries []tocEntry) []map[string]any {
	result := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		result = append(result, map[string]any{
			"id":    entry.Id,
			"name":  entry.Name,
			"title": entry.Name,
			"level": entry.Level,
		})
	}
	return result
}

func renderTemplateHTML(themeTemplate *Jinja2goTemplate, ctx *pageContext) (string, error) {
	var htmlBuf strings.Builder
	if err := themeTemplate.Execute(&htmlBuf, ctx); err != nil {
		return "", err
	}
	return htmlBuf.String(), nil
}

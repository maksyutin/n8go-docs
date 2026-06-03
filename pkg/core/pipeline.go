package core

import (
	"errors"
	"strings"

	"n8go-docs/diagnostics"
	"n8go-docs/manifest"

	"github.com/fatih/color"
)

// Pipeline holds the site-wide state produced by a full build and exposes
// incremental rendering for the editor. Register plugins before calling Build.
type Pipeline struct {
	site     manifest.SiteManifest
	theme    manifest.ThemeManifest
	themeDir string
	plugins  []Plugin

	// state populated by Build; reused by RenderPageToHTML
	navChildren   []*navNode
	pageIndex     PageIndex
	themeTemplate *Jinja2goTemplate
}

// NewPipeline creates a pipeline for the given site/theme configuration.
func NewPipeline(site manifest.SiteManifest, theme manifest.ThemeManifest, themeDir string) *Pipeline {
	return &Pipeline{
		site:     site,
		theme:    theme,
		themeDir: themeDir,
	}
}

// Register appends a plugin. Call before Build.
func (p *Pipeline) Register(plugin Plugin) {
	p.plugins = append(p.plugins, plugin)
}

// Build performs a full site generation: builds the nav tree, renders all
// pages through registered plugins, writes output files, copies static assets.
func (p *Pipeline) Build() error {
	var sw diagnostics.Stopwatch
	sw.Reset()

	diagnostics.Info("Building documentation...")

	var contexts []pageContext

	if len(p.site.Nav) > 0 {
		children, err := buildNavFromConfig(p.site.Nav, p.site, p.theme, &contexts)
		if err != nil {
			return err
		}
		p.navChildren = children
	} else {
		var root navNode
		if err := prepareDocumentationTree(p.site.InputPath, "", &root, p.site, p.theme, &contexts); err != nil {
			return err
		}
		p.navChildren = root.Children
	}

	if len(contexts) == 0 {
		return errors.New("no pages found in docs_dir: add at least index.md or README.md")
	}

	tmpl, err := generateTemplate(p.themeDir)
	if err != nil {
		return err
	}
	p.themeTemplate = tmpl
	p.pageIndex = buildPageIndex(contexts)

	bctx := p.buildContext()

	for _, pl := range p.plugins {
		if h, ok := pl.(InitHook); ok {
			if err := h.OnInit(bctx); err != nil {
				return err
			}
		}
	}

	for i := range contexts {
		nav := cloneNavTree(p.navChildren)
		markActiveNodes(nav, contexts[i].Url)
		contexts[i].Nav = nav
		contexts[i].Index = p.pageIndex

		html, err := p.renderContextToHTML(&contexts[i])
		if err != nil {
			diagnostics.PrintError(err, "failed to render "+contexts[i].Page.FileName)
			continue
		}

		result := pageResultFromContext(&contexts[i])

		for _, pl := range p.plugins {
			if h, ok := pl.(PageRenderedHook); ok {
				html, err = h.OnPageRendered(result, html)
				if err != nil {
					diagnostics.PrintError(err, "plugin "+pl.Name()+" OnPageRendered failed for "+contexts[i].Page.FileName)
				}
			}
		}

		if err := writePageHTML(&contexts[i], p.site, html); err != nil {
			diagnostics.PrintError(err, "failed to write "+contexts[i].Page.FileName)
		}
	}

	for _, pl := range p.plugins {
		if h, ok := pl.(BuildCompleteHook); ok {
			if err := h.OnBuildComplete(bctx); err != nil {
				diagnostics.PrintError(err, "plugin "+pl.Name()+" OnBuildComplete failed")
			}
		}
	}

	if err := copyStaticFiles(p.site, p.themeDir); err != nil {
		return err
	}

	color.Green("Documentation generated in %dms", sw.Milliseconds())
	return nil
}

// RenderPageToHTML re-renders a single .md file using the current nav tree and
// page index without writing anything to disk. Returns ready-to-serve HTML.
// Build must have been called at least once before using this method.
func (p *Pipeline) RenderPageToHTML(mdFilePath string) (string, error) {
	if p.themeTemplate == nil {
		return "", errors.New("pipeline: Build must be called before RenderPageToHTML")
	}

	depth := 0
	if url, ok := p.pageIndex[mdFilePath]; ok && url != "" && url != "." {
		depth = len(strings.Split(url, "/"))
	}
	rootPath := strings.Repeat("../", depth)

	ctx, err := createPageContext(mdFilePath, rootPath, p.site, p.theme)
	if err != nil {
		return "", err
	}

	nav := cloneNavTree(p.navChildren)
	markActiveNodes(nav, ctx.Url)
	ctx.Nav = nav
	ctx.Index = p.pageIndex

	html, err := p.renderContextToHTML(&ctx)
	if err != nil {
		return "", err
	}

	result := pageResultFromContext(&ctx)

	for _, pl := range p.plugins {
		if h, ok := pl.(EditorPreviewHook); ok {
			html, err = h.OnPreview(mdFilePath, html)
			if err != nil {
				diagnostics.PrintError(err, "plugin "+pl.Name()+" OnPreview failed")
			}
		}
	}

	_ = result
	return html, nil
}

// ReadFile reads the raw markdown for the given file, applying EditorFileReadHooks.
func (p *Pipeline) ReadFile(filePath string) ([]byte, error) {
	content, err := readFileBytes(filePath)
	if err != nil {
		return nil, err
	}
	for _, pl := range p.plugins {
		if h, ok := pl.(EditorFileReadHook); ok {
			content, err = h.OnFileRead(filePath, content)
			if err != nil {
				return nil, err
			}
		}
	}
	return content, nil
}

// WriteFile persists content to filePath, applying EditorFileWriteHooks first.
func (p *Pipeline) WriteFile(filePath string, content []byte) error {
	var err error
	for _, pl := range p.plugins {
		if h, ok := pl.(EditorFileWriteHook); ok {
			content, err = h.OnFileWrite(filePath, content)
			if err != nil {
				return err
			}
		}
	}
	return writeFileBytes(filePath, content)
}

// Site returns the site manifest (read-only access for the editor layer).
func (p *Pipeline) Site() manifest.SiteManifest { return p.site }

// ---- internal helpers -------------------------------------------------------

func (p *Pipeline) buildContext() *BuildContext {
	return &BuildContext{
		NavTree:   p.navChildren,
		PageIndex: p.pageIndex,
		OutputDir: p.site.OutputPath,
	}
}

// renderContextToHTML renders a page context to an HTML string (no disk I/O).
func (p *Pipeline) renderContextToHTML(ctx *pageContext) (string, error) {
	raw, err := renderTemplateHTML(p.themeTemplate, ctx)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := processHtml(strings.NewReader(raw), &sb, ctx); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func pageResultFromContext(ctx *pageContext) *PageResult {
	return &PageResult{
		FilePath: ctx.Page.FilePath,
		Title:    ctx.Page.Title,
		URL:      ctx.Url,
		RootPath: ctx.RootPath,
	}
}

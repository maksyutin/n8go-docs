package core

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"n8go-docs/diagnostics"
	"n8go-docs/manifest"
	"n8go-docs/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
)

func findDirForPage(page pageInfo, siteManifest manifest.SiteManifest) string {
	relativePath := page.FilePath[len(siteManifest.InputPath)+1:]
	outputDirPath := filepath.Dir(relativePath)

	if !isIndexFile(page.FilePath) && !isReadmeIndexForFile(page.FilePath) {
		outputDirPath = filepath.Join(outputDirPath, page.FileName)
	}

	return outputDirPath
}

// isReadmeIndexForFile reports whether filePath is README.md acting as index
// by checking the filesystem for absence of index.md in the same directory.
func isReadmeIndexForFile(filePath string) bool {
	if utils.GetFileName(filePath) != ReadmeFileName {
		return false
	}
	dir := filepath.Dir(filePath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return !hasIndexFile(entries)
}

func createPageContext(mdFile string, rootPath string, siteManifest manifest.SiteManifest, themeManifest manifest.ThemeManifest) (pageContext, error) {
	page, err := renderMarkdownPage(mdFile, themeManifest, siteManifest)
	if err != nil {
		return pageContext{}, err
	}
	return pageContext{
		Page: page,
		Site: siteManifest,
		Generator: generatorInfo{
			Name:    ProgramName,
			Version: ProgramVersion,
		},
		Now:      time.Now().Format("2006-01-02 15:04:05.000"),
		RootPath: rootPath,
		Url:      filepath.ToSlash(findDirForPage(page, siteManifest)),
		InputDir: filepath.ToSlash(filepath.Dir(filepath.Clean(mdFile))),
	}, nil
}

func openOutputFileForPage(ctx *pageContext, siteManifest manifest.SiteManifest) (*os.File, error) {
	outPath := filepath.Join(siteManifest.OutputPath, ctx.Url)
	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(filepath.Join(outPath, "index.html"))
}

func generateThemedHtmlForPage(ctx *pageContext, siteManifest manifest.SiteManifest, themeTemplate *template.Template) {
	mdFile := ctx.Page.FileName

	writer, err := openOutputFileForPage(ctx, siteManifest)
	if err != nil {
		diagnostics.PrintError(err, "failed to open output file for "+mdFile)
		return
	}

	var htmlBuf strings.Builder
	if err = themeTemplate.ExecuteTemplate(&htmlBuf, RootTemplateName, ctx); err != nil {
		diagnostics.PrintError(err, "failed to execute template for "+mdFile)
		return
	}

	if err = processHtml(strings.NewReader(htmlBuf.String()), writer, ctx); err != nil {
		diagnostics.PrintError(err, "failed to run HTML postproc for "+mdFile)
		return
	}

	if siteManifest.DefaultSearch {
		indexPageContent(ctx, siteManifest, htmlBuf.String())
	}

	if err = writer.Close(); err != nil {
		diagnostics.PrintError(err, "failed to close output file for "+mdFile)
	}
}

func indexPageContent(ctx *pageContext, siteManifest manifest.SiteManifest, rawHtml string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHtml))
	if err != nil {
		diagnostics.PrintError(err, "failed to parse HTML for search indexing: "+ctx.Page.FileName)
		return
	}

	content, err := doc.Find(".main-content").Html()
	if err != nil {
		diagnostics.PrintError(err, "failed to find main content for "+ctx.Page.FileName)
		return
	}

	content = bluemonday.StrictPolicy().Sanitize(content)
	content = strings.ReplaceAll(content, "\n", " ")

	if lim := siteManifest.SearchContentLimit; lim > 0 && len([]rune(content)) > lim {
		content = string([]rune(content)[:lim])
	}

	AddToSearchIndex(siteManifest, SearchIndexEntry{
		Title:   ctx.Page.Title,
		Url:     ctx.Url,
		Content: content,
	})
}

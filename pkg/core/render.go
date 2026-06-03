package core

import (
	"os"
	"path/filepath"
	"time"

	"n8go-docs/manifest"
	"n8go-docs/utils"
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

// writePageHTML writes the provided HTML string to the output directory for ctx.
func writePageHTML(ctx *pageContext, siteManifest manifest.SiteManifest, html string) error {
	outPath := filepath.Join(siteManifest.OutputPath, ctx.Url)
	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outPath, "index.html"), []byte(html), 0o644)
}

// readFileBytes reads a file from disk.
func readFileBytes(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// writeFileBytes writes content to a file, creating directories as needed.
func writeFileBytes(filePath string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filePath, content, 0o644)
}


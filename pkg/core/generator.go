package core

import (
	"errors"
	"n8go-docs/diagnostics"
	"n8go-docs/manifest"
	"n8go-docs/utils"

	"github.com/fatih/color"
)

func GenerateDocumentation(siteManifest manifest.SiteManifest, themeManifest manifest.ThemeManifest, themeDir string) error {
	var stopwatch diagnostics.Stopwatch
	stopwatch.Reset()

	diagnostics.Info("Building documentation...")

	var contexts []pageContext
	var navChildren []*navNode

	if len(siteManifest.Nav) > 0 {
		children, err := buildNavFromConfig(siteManifest.Nav, siteManifest, themeManifest, &contexts)
		if err != nil {
			return err
		}
		navChildren = children
	} else {
		var navTreeRoot navNode
		if err := prepareDocumentationTree(siteManifest.InputPath, "", &navTreeRoot, siteManifest, themeManifest, &contexts); err != nil {
			return err
		}
		navChildren = navTreeRoot.Children
	}

	if len(contexts) == 0 {
		return errors.New("no pages found in docs_dir: add at least index.md or README.md")
	}

	themeTemplate, err := generateTemplate(themeDir)
	if err != nil {
		return err
	}

	pageIndex := buildPageIndex(contexts)
	var searchIndex *SearchIndex
	if siteManifest.DefaultSearch {
		searchIndex = NewSearchIndex()
	}

	for i := range contexts {
		nav := cloneNavTree(navChildren)
		markActiveNodes(nav, contexts[i].Url)
		contexts[i].Nav = nav
		contexts[i].Index = pageIndex
		generateThemedHtmlForPage(&contexts[i], siteManifest, themeTemplate, searchIndex)
	}

	if searchIndex != nil {
		if err := searchIndex.Write(siteManifest); err != nil {
			return err
		}
	}

	if err := copyStaticFiles(siteManifest, themeDir); err != nil {
		return err
	}

	color.Green("Documentation generated in %dms", stopwatch.Milliseconds())
	return nil
}

func copyStaticFiles(siteManifest manifest.SiteManifest, themeDir string) error {
	isStaticFile := func(ext string) bool {
		return ext != ".md" && ext != ".html" && ext != ".yaml"
	}

	err := utils.CopyDirContentsWithHook(
		siteManifest.InputPath,
		siteManifest.OutputPath,
		isStaticFile,
		func(relPath string) bool {
			if utils.MatchesExclude(relPath, siteManifest.ExcludeDocs) {
				return false
			}
			diagnostics.Info("Copying '%s' (static)", relPath)
			return true
		},
	)
	if err != nil {
		return err
	}

	return utils.CopyDirContents(themeDir, siteManifest.OutputPath, isStaticFile)
}

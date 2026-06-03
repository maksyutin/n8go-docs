package core

import (
	"io/fs"
	"n8go-docs/manifest"
	"n8go-docs/utils"
	"os"
	"path/filepath"
)

// GenerateDocumentation builds the full documentation site.
// Plugins are registered on the pipeline before Build is called.
// This function is the entry point used by the CLI build command.
func GenerateDocumentation(siteManifest manifest.SiteManifest, themeManifest manifest.ThemeManifest, themeDir string, plugins ...Plugin) error {
	p := NewPipeline(siteManifest, themeManifest, themeDir)
	for _, pl := range plugins {
		p.Register(pl)
	}
	return p.Build()
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
			return true
		},
	)
	if err != nil {
		return err
	}

	return copyThemeAssets(themeDir, siteManifest.OutputPath, isStaticFile)
}

func copyThemeAssets(themeDir string, outputDir string, isStaticFile func(ext string) bool) error {
	return filepath.WalkDir(themeDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !isStaticFile(filepath.Ext(path)) {
			return nil
		}

		relPath, err := filepath.Rel(themeDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		var assetPath string
		switch {
		case relPath == "css" || relPath == "js" || relPath == "img":
			return nil
		case len(relPath) > 4 && relPath[:4] == "css/":
			assetPath = filepath.Join(outputDir, "assets", relPath)
		case len(relPath) > 3 && relPath[:3] == "js/":
			assetPath = filepath.Join(outputDir, "assets", relPath)
		case len(relPath) > 4 && relPath[:4] == "img/":
			assetPath = filepath.Join(outputDir, "assets", relPath)
		default:
			return nil
		}

		return copyThemeAssetFile(path, assetPath)
	})
}

func copyThemeAssetFile(src string, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	inFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = outFile.ReadFrom(inFile)
	return err
}

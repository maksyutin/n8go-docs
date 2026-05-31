package main

import (
	"n8go-docs/core"
	"n8go-docs/diagnostics"
	"n8go-docs/manifest"
	"os"

	"github.com/fatih/color"
)

func runGenerator(configPath string) error {
	siteManifest, err := manifest.ParseSiteManifest(configPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(siteManifest.OutputPath, os.ModePerm)
	if err != nil {
		return err
	}

	themeBaseDir, err := findThemesBaseDir()
	if err != nil {
		return err
	}

	themeManifest, themeDir, err := findThemeConfig(themeBaseDir, siteManifest.ThemeId)
	if err != nil {
		return err
	}

	diagnostics.Debug(func() {
		color.Yellow("Using theme: %s by %s", themeManifest.Name, themeManifest.Author)
	})
	return core.GenerateDocumentation(siteManifest, themeManifest, themeDir)
}

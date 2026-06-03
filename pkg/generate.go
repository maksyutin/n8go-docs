package main

import (
	"n8go-docs/core"
	"n8go-docs/diagnostics"
	"n8go-docs/manifest"
	"n8go-docs/plugins/search"
	"os"

	"github.com/fatih/color"
)

func runGenerator(configPath string) error {
	siteManifest, themeManifest, themeDir, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	pipeline := buildPipeline(siteManifest, themeManifest, themeDir)
	return pipeline.Build()
}

// buildPipeline constructs a Pipeline with the standard set of plugins.
func buildPipeline(site manifest.SiteManifest, theme manifest.ThemeManifest, themeDir string) *core.Pipeline {
	p := core.NewPipeline(site, theme, themeDir)

	if site.DefaultSearch {
		p.Register(search.New(site.SearchContentLimit, site.SiteURL, site.OutputPath))
	}

	return p
}

// loadConfig parses the site manifest and resolves the theme directory.
func loadConfig(configPath string) (manifest.SiteManifest, manifest.ThemeManifest, string, error) {
	siteManifest, err := manifest.ParseSiteManifest(configPath)
	if err != nil {
		return manifest.SiteManifest{}, manifest.ThemeManifest{}, "", err
	}

	if err := os.MkdirAll(siteManifest.OutputPath, os.ModePerm); err != nil {
		return manifest.SiteManifest{}, manifest.ThemeManifest{}, "", err
	}

	themeBaseDir, err := findThemesBaseDir()
	if err != nil {
		return manifest.SiteManifest{}, manifest.ThemeManifest{}, "", err
	}

	themeManifest, themeDir, err := findThemeConfig(themeBaseDir, siteManifest.ThemeId)
	if err != nil {
		return manifest.SiteManifest{}, manifest.ThemeManifest{}, "", err
	}

	diagnostics.Debug(func() {
		color.Yellow("Using theme: %s by %s", themeManifest.Name, themeManifest.Author)
	})

	return siteManifest, themeManifest, themeDir, nil
}

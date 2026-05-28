package manifest

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

type siteYaml struct {
	Name               string   `yaml:"name"`
	Theme              string   `yaml:"theme"`
	Input              string   `yaml:"input"`
	Output             string   `yaml:"output"`
	DefaultSearch      *bool    `yaml:"default_search"`
	SearchEngine       string   `yaml:"search_engine"`
	SearchContentLimit *int     `yaml:"search_content_limit"`
	HeadTags           []string `yaml:"head_tags"`
	CustomFont         string   `yaml:"custom_font"`
	Logo               string   `yaml:"logo"`
	StripMdExtension   *bool    `yaml:"strip_md_extension"`
}

func boolOr(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

func intOr(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}

func strOr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// ParseSiteManifest loads site config from a YAML file.
// Falls back to legacy INI format (utdocs.ini) if the file has a .ini extension.
func ParseSiteManifest(path string) (SiteManifest, error) {
	ext := filepath.Ext(path)
	if ext == ".ini" {
		return parseSiteIni(path)
	}
	return parseSiteYaml(path)
}

func parseSiteYaml(path string) (SiteManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SiteManifest{}, err
	}

	var cfg siteYaml
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return SiteManifest{}, err
	}

	result := SiteManifest{
		Name:               cfg.Name,
		ThemeId:            strOr(cfg.Theme, "default"),
		InputPath:          strOr(cfg.Input, "docs"),
		OutputPath:         strOr(cfg.Output, "docs_gen"),
		DefaultSearch:      boolOr(cfg.DefaultSearch, true),
		SearchEngine:       strOr(cfg.SearchEngine, "flexsearch"),
		SearchContentLimit: intOr(cfg.SearchContentLimit, 500),
		HeadTags:           cfg.HeadTags,
		CustomFont:         cfg.CustomFont,
		Logo:               strOr(cfg.Logo, "img/book.svg"),
		StripMdExtension:   boolOr(cfg.StripMdExtension, false),
	}

	if !result.IsValid() {
		return result, errors.New("missing required field: name")
	}

	result.InputPath = filepath.Clean(result.InputPath)
	result.OutputPath = filepath.Clean(result.OutputPath)

	return result, nil
}

// parseSiteIni keeps backward compatibility with the old INI format.
func parseSiteIni(path string) (SiteManifest, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return SiteManifest{}, err
	}

	s := cfg.Section("Site")
	result := SiteManifest{
		Name:               s.Key("Name").String(),
		ThemeId:            s.Key("Theme").MustString("default"),
		DefaultSearch:      s.Key("DefaultSearch").MustBool(true),
		SearchEngine:       s.Key("SearchEngine").MustString("flexsearch"),
		SearchContentLimit: s.Key("SearchContentLimit").MustInt(500),
		CustomFont:         s.Key("CustomFont").String(),
		InputPath:          s.Key("Input").MustString("docs"),
		OutputPath:         s.Key("Output").MustString("docs_gen"),
		Logo:               s.Key("Logo").MustString("img/book.svg"),
		StripMdExtension:   s.Key("StripMdExtension").MustBool(false),
	}
	for _, t := range s.Key("HeadTags").Strings(",") {
		if t != "" {
			result.HeadTags = append(result.HeadTags, t)
		}
	}

	if !result.IsValid() {
		return result, errors.New("missing required parameters")
	}

	result.InputPath = filepath.Clean(result.InputPath)
	result.OutputPath = filepath.Clean(result.OutputPath)

	return result, nil
}

func ParseThemeManifest(path string) (ThemeManifest, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return ThemeManifest{}, err
	}

	result := ThemeManifest{}

	if s := cfg.Section("Theme"); s != nil {
		result.Name = s.Key("Name").String()
		result.Description = s.Key("Description").String()
		result.Repository = s.Key("Repository").String()
		result.Version = s.Key("Version").String()
		result.Author = s.Key("Author").String()
		result.License = s.Key("License").String()
	}

	if s := cfg.Section("Highlighting"); s != nil {
		result.Highlighting.Style = s.Key("Style").MustString("bw")
		result.Highlighting.LineNumbers = s.Key("LineNumbers").MustBool(false)
	}

	if !result.IsValid() {
		return result, errors.New("missing required parameters")
	}

	return result, nil
}

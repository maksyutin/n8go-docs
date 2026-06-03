package manifest

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// navItemYaml is the raw YAML representation of a nav entry.
// Each entry is a single-key map: { "Title": "file.md" } or { "Title": [child, ...] }.
type navItemYaml struct {
	Title    string
	File     string
	Children []navItemYaml
}

func (n *navItemYaml) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode || len(value.Content) < 2 {
		return errors.New("nav entry must be a mapping with one key")
	}
	n.Title = value.Content[0].Value
	val := value.Content[1]
	switch val.Kind {
	case yaml.ScalarNode:
		n.File = val.Value
	case yaml.SequenceNode:
		for _, child := range val.Content {
			var item navItemYaml
			if err := child.Decode(&item); err != nil {
				return err
			}
			n.Children = append(n.Children, item)
		}
	default:
		return errors.New("nav entry value must be a string or a list")
	}
	return nil
}

func convertNavItems(raw []navItemYaml) []NavItem {
	result := make([]NavItem, 0, len(raw))
	for _, r := range raw {
		item := NavItem{Title: r.Title, File: r.File}
		if len(r.Children) > 0 {
			item.Children = convertNavItems(r.Children)
		}
		result = append(result, item)
	}
	return result
}

type siteYaml struct {
	SiteName           string          `yaml:"site_name"`
	SiteURL            string          `yaml:"site_url"`
	SiteDescription    string          `yaml:"site_description"`
	DevAddr            string          `yaml:"dev_addr"`
	UseDirectoryURLs   *bool           `yaml:"use_directory_urls"`
	Theme              string          `yaml:"theme"`
	DocsDir            string          `yaml:"docs_dir"`
	SiteDir            string          `yaml:"site_dir"`
	DefaultSearch      *bool           `yaml:"default_search"`
	SearchEngine       string          `yaml:"search_engine"`
	SearchContentLimit *int            `yaml:"search_content_limit"`
	HeadTags           []string        `yaml:"head_tags"`
	CustomFont         string          `yaml:"custom_font"`
	Logo               string          `yaml:"logo"`
	StripMdExtension   *bool           `yaml:"strip_md_extension"`
	ExtraCss           []string        `yaml:"extra_css"`
	ExtraJavascript    []string        `yaml:"extra_javascript"`
	Nav                []navItemYaml   `yaml:"nav"`
	ExcludeDocs        excludeDocsYaml `yaml:"exclude_docs"`
}

// excludeDocsYaml accepts exclude_docs as either a multiline string (block
// scalar) or a YAML list of patterns. In both cases it normalizes the value
// into a slice of patterns.
type excludeDocsYaml []string

func (e *excludeDocsYaml) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*e = parseExcludePatterns(value.Value)
		return nil
	case yaml.SequenceNode:
		var items []string
		if err := value.Decode(&items); err != nil {
			return err
		}
		var patterns []string
		for _, item := range items {
			for _, p := range parseExcludePatterns(item) {
				patterns = append(patterns, p)
			}
		}
		*e = patterns
		return nil
	default:
		return errors.New("exclude_docs must be a string or a list")
	}
}

// parseExcludePatterns splits a multiline exclude_docs string into individual patterns,
// stripping blank lines and inline comments.
func parseExcludePatterns(raw string) []string {
	var patterns []string
	for _, line := range splitLines(raw) {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if line != "" {
			patterns = append(patterns, line)
		}
	}
	return patterns
}

func splitLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

type themeYaml struct {
	Theme struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Repository  string `yaml:"repository"`
		Version     string `yaml:"version"`
		Author      string `yaml:"author"`
		License     string `yaml:"license"`
	} `yaml:"theme"`
	Highlighting struct {
		Style       string `yaml:"style"`
		LineNumbers bool   `yaml:"line_numbers"`
	} `yaml:"highlighting"`
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

func normalizeURL(rawURL string) string {
	normalized := strings.TrimSpace(rawURL)
	for strings.HasSuffix(normalized, "/") && !strings.HasSuffix(normalized, "://") {
		normalized = strings.TrimSuffix(normalized, "/")
	}
	return normalized
}

func ParseSiteManifest(path string) (SiteManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SiteManifest{}, err
	}

	var cfg siteYaml
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return SiteManifest{}, err
	}

	result := SiteManifest{
		SiteName:           cfg.SiteName,
		SiteURL:            normalizeURL(cfg.SiteURL),
		SiteDescription:    cfg.SiteDescription,
		DevAddr:            strings.TrimSpace(cfg.DevAddr),
		UseDirectoryURLs:   boolOr(cfg.UseDirectoryURLs, true),
		ThemeId:            strOr(cfg.Theme, "default"),
		InputPath:          strOr(cfg.DocsDir, "docs"),
		OutputPath:         strOr(cfg.SiteDir, "site"),
		DefaultSearch:      boolOr(cfg.DefaultSearch, true),
		SearchEngine:       strOr(cfg.SearchEngine, "flexsearch"),
		SearchContentLimit: intOr(cfg.SearchContentLimit, 500),
		HeadTags:           cfg.HeadTags,
		CustomFont:         cfg.CustomFont,
		Logo:               strOr(cfg.Logo, "assets/img/logo.svg"),
		StripMdExtension:   boolOr(cfg.StripMdExtension, false),
		ExtraCss:           cfg.ExtraCss,
		ExtraJavascript:    cfg.ExtraJavascript,
		Nav:                convertNavItems(cfg.Nav),
		ExcludeDocs:        []string(cfg.ExcludeDocs),
	}

	if result.SiteName == "" {
		return result, errors.New("missing required field: site_name")
	}
	if result.ThemeId == "" {
		return result, errors.New("missing required field: theme")
	}
	if result.InputPath == "" {
		return result, errors.New("missing required field: input")
	}
	if result.OutputPath == "" {
		return result, errors.New("missing required field: output")
	}

	configDir := filepath.Dir(path)
	result.InputPath = filepath.Clean(filepath.Join(configDir, result.InputPath))
	result.OutputPath = filepath.Clean(filepath.Join(configDir, result.OutputPath))

	return result, nil
}

func ParseThemeManifest(path string) (ThemeManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ThemeManifest{}, err
	}

	var cfg themeYaml
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ThemeManifest{}, err
	}

	result := ThemeManifest{
		Name:        cfg.Theme.Name,
		Description: cfg.Theme.Description,
		Repository:  cfg.Theme.Repository,
		Version:     cfg.Theme.Version,
		Author:      cfg.Theme.Author,
		License:     cfg.Theme.License,
		Highlighting: HighlightingConfig{
			Style:       strOr(cfg.Highlighting.Style, "bw"),
			LineNumbers: cfg.Highlighting.LineNumbers,
		},
	}

	if result.Name == "" {
		return result, errors.New("theme: missing required field: name")
	}
	if result.Version == "" {
		return result, errors.New("theme: missing required field: version")
	}
	if result.Highlighting.Style == "" {
		return result, errors.New("theme: missing required field: highlighting.style")
	}

	return result, nil
}

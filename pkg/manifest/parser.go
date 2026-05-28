package manifest

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

func getEnvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getEnvStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// ParseSiteManifest loads site config from environment variables.
// Falls back to the .env file (key=value format) if present.
func ParseSiteManifest(envFile string) (SiteManifest, error) {
	loadEnvFile(envFile)

	result := SiteManifest{
		Name:               getEnvStr("UTDOCS_NAME", ""),
		ThemeId:            getEnvStr("UTDOCS_THEME", "default"),
		InputPath:          getEnvStr("UTDOCS_INPUT", "docs"),
		OutputPath:         getEnvStr("UTDOCS_OUTPUT", "docs_gen"),
		DefaultSearch:      getEnvBool("UTDOCS_DEFAULT_SEARCH", true),
		SearchEngine:       getEnvStr("UTDOCS_SEARCH_ENGINE", "fuse"),
		SearchContentLimit: getEnvInt("UTDOCS_SEARCH_CONTENT_LIMIT", 500),
		CustomFont:         getEnvStr("UTDOCS_CUSTOM_FONT", ""),
		Logo:               getEnvStr("UTDOCS_LOGO", "img/book.svg"),
		StripMdExtension:   getEnvBool("UTDOCS_STRIP_MD_EXTENSION", false),
	}

	if tags := os.Getenv("UTDOCS_HEAD_TAGS"); tags != "" {
		for _, t := range strings.Split(tags, ",") {
			if t = strings.TrimSpace(t); t != "" {
				result.HeadTags = append(result.HeadTags, t)
			}
		}
	}

	if !result.IsValid() {
		return result, errors.New("missing required parameters: UTDOCS_NAME must be set")
	}

	result.InputPath = filepath.Clean(result.InputPath)
	result.OutputPath = filepath.Clean(result.OutputPath)

	return result, nil
}

// loadEnvFile reads a key=value file and sets env vars that are not already set.
func loadEnvFile(path string) {
	cfg, err := ini.Load(path)
	if err != nil {
		return
	}
	section := cfg.Section("Site")
	mapping := map[string]string{
		"Name":               "UTDOCS_NAME",
		"Theme":              "UTDOCS_THEME",
		"Input":              "UTDOCS_INPUT",
		"Output":             "UTDOCS_OUTPUT",
		"DefaultSearch":      "UTDOCS_DEFAULT_SEARCH",
		"SearchEngine":       "UTDOCS_SEARCH_ENGINE",
		"SearchContentLimit": "UTDOCS_SEARCH_CONTENT_LIMIT",
		"HeadTags":           "UTDOCS_HEAD_TAGS",
		"CustomFont":         "UTDOCS_CUSTOM_FONT",
		"Logo":               "UTDOCS_LOGO",
		"StripMdExtension":   "UTDOCS_STRIP_MD_EXTENSION",
	}
	for iniKey, envKey := range mapping {
		if os.Getenv(envKey) == "" {
			if v := section.Key(iniKey).String(); v != "" {
				os.Setenv(envKey, v)
			}
		}
	}
}

func ParseThemeManifest(path string) (ThemeManifest, error) {
	manifest, err := ini.Load(path)
	if err != nil {
		return ThemeManifest{}, err
	}

	result := ThemeManifest{}

	rootSection := manifest.Section("Theme")
	if rootSection != nil {
		result.Name = rootSection.Key("Name").String()
		result.Description = rootSection.Key("Description").String()
		result.Repository = rootSection.Key("Repository").String()
		result.Version = rootSection.Key("Version").String()
		result.Author = rootSection.Key("Author").String()
		result.License = rootSection.Key("License").String()
	}

	highlightingSection := manifest.Section("Highlighting")
	if highlightingSection != nil {
		result.Highlighting.Style = highlightingSection.Key("Style").MustString("bw")
		result.Highlighting.LineNumbers = highlightingSection.Key("LineNumbers").MustBool(false)
	}

	if !result.IsValid() {
		return result, errors.New("missing required parameters")
	}

	return result, nil
}

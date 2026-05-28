package manifest

type SiteManifest struct {
	Name               string
	ThemeId            string
	InputPath          string
	OutputPath         string
	DefaultSearch      bool
	SearchEngine       string // "fuse" | "flexsearch"
	SearchContentLimit int    // max chars stored per page in search index (0 = unlimited)
	HeadTags           []string
	CustomFont         string
	Logo               string
	StripMdExtension   bool
}

type HighlightingConfig struct {
	Style       string
	LineNumbers bool
}

type ThemeManifest struct {
	Name         string
	Description  string
	Repository   string
	Version      string
	Author       string
	License      string
	Highlighting HighlightingConfig
}

func (manifest *SiteManifest) IsValid() bool {
	return manifest.Name != "" && manifest.ThemeId != "" && manifest.InputPath != "" && manifest.OutputPath != ""
}

func (manifest *ThemeManifest) IsValid() bool {
	return manifest.Name != "" && manifest.Version != "" && manifest.Highlighting.Style != ""
}

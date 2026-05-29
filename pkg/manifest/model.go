package manifest

// NavItem represents one entry in the explicit nav config.
// Either File is set (leaf page) or Children is set (section).
type NavItem struct {
	Title    string
	File     string    // path relative to InputPath, empty for sections
	Children []NavItem // non-nil only for sections
}

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
	ExtraCss           []string  // paths relative to InputPath, copied to output and linked on every page
	ExtraJavascript    []string  // paths relative to InputPath, copied to output and added as <script> on every page
	Nav                []NavItem // explicit navigation; empty = auto-build from filesystem
	ExcludeDocs        []string  // glob patterns (relative to InputPath) of files/dirs to exclude entirely
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

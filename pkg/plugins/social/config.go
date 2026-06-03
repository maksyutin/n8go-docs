package social

// Config mirrors the configuration schema of the MkDocs-Material social plugin.
// All fields are optional; see defaults in New().
type Config struct {
	// Enabled disables the plugin entirely when false.
	// Default: true
	Enabled bool

	// Concurrency is the number of cards generated in parallel.
	// Default: runtime.NumCPU() - 1, minimum 1
	Concurrency int

	// Cache controls whether previously generated cards are reused.
	// Default: true
	Cache bool

	// CacheDir is the directory used to store cached card images.
	// Default: ".cache/plugin/social"
	CacheDir string

	// LogLevel controls how generation errors are reported.
	// "warn" (default) – logged as warnings, build continues.
	// "info"           – logged at info level.
	// "ignore"         – silently dropped.
	LogLevel string

	// Cards enables or disables social card generation.
	// Default: true
	Cards bool

	// CardsDir is the path inside the output directory where generated cards
	// are written.
	// Default: "assets/images/social"
	CardsDir string

	// CardsLayoutDir is the directory where custom layout definitions live.
	// Default: "layouts"
	CardsLayoutDir string

	// CardsLayout selects the card layout.
	// Supported: "default" (only built-in layout at the moment).
	// Default: "default"
	CardsLayout string

	// CardsLayoutOptions contains layout-level overrides applied to every card
	// unless overridden per page.
	CardsLayoutOptions LayoutOptions

	// CardsInclude is a list of glob patterns (relative to docs_dir).
	// When non-empty only matching pages get a card.
	CardsInclude []string

	// CardsExclude is a list of glob patterns (relative to docs_dir).
	// Matching pages are skipped.
	CardsExclude []string

	// Debug renders a dot-grid and layer outlines on top of cards.
	// Default: false
	Debug bool

	// DebugOnBuild keeps debug overlays during a normal `build` run.
	// Default: false
	DebugOnBuild bool

	// DebugGrid draws a dot grid when Debug is true.
	// Default: true
	DebugGrid bool

	// DebugGridStep is the pixel step between grid dots.
	// Default: 32
	DebugGridStep int

	// DebugColor is the color of layer outlines and the grid.
	// Default: "grey"
	DebugColor string
}

// LayoutOptions are the per-card visual overrides that can be set in
// CardsLayoutOptions or overridden per page via front-matter.
type LayoutOptions struct {
	// BackgroundColor overrides the card background.
	// Accepts CSS color syntax: #rgb, #rrggbb, named colors.
	BackgroundColor string

	// BackgroundImage is a path (relative to project root) to a background
	// image. It is tinted by BackgroundColor when both are set.
	BackgroundImage string

	// Color is the foreground / text color.
	Color string

	// FontFamily selects the typeface. Must be a valid Google Fonts name.
	// The font is downloaded on first use and cached.
	// Default: "Roboto"
	FontFamily string

	// FontVariant is an optional style modifier, e.g. "Condensed".
	FontVariant string

	// Logo is the path (relative to project root) to the logo image.
	Logo string

	// Title overrides the page title shown on the card.
	Title string

	// Description overrides the page description shown on the card.
	Description string
}

func defaultConfig() Config {
	return Config{
		Enabled:        true,
		Concurrency:    0, // resolved to max(1, NumCPU-1) at runtime
		Cache:          true,
		CacheDir:       ".cache/plugin/social",
		LogLevel:       "warn",
		Cards:          true,
		CardsDir:       "assets/images/social",
		CardsLayoutDir: "layouts",
		CardsLayout:    "default",
		Debug:          false,
		DebugOnBuild:   false,
		DebugGrid:      true,
		DebugGridStep:  32,
		DebugColor:     "grey",
		CardsLayoutOptions: LayoutOptions{
			FontFamily: "Roboto",
		},
	}
}

// Package social generates Open Graph / Twitter Card social preview images
// for every documentation page and injects the corresponding <meta> tags.
//
// It mirrors the configuration schema of the MkDocs-Material social plugin:
// https://squidfunk.github.io/mkdocs-material/plugins/social/
//
// Wiring example:
//
//	p := core.NewPipeline(site, theme, themeDir)
//	p.Register(social.New(social.Config{
//	    Cards: true,
//	    CardsLayoutOptions: social.LayoutOptions{
//	        BackgroundColor: "#1976d2",
//	        Color:           "#ffffff",
//	    },
//	}))
package social

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"n8go-docs/core"
	"n8go-docs/diagnostics"
	"n8go-docs/utils"
)

// Plugin generates social preview cards and injects Open Graph meta tags.
type Plugin struct {
	cfg     Config
	siteURL string
	outDir  string

	mu    sync.Mutex
	pages []pageJob // populated by OnPageRendered, consumed by OnBuildComplete
}

type pageJob struct {
	result  *core.PageResult
	opts    LayoutOptions // effective options (global merged with page-level)
	cardRel string        // relative path inside CardsDir, e.g. "guide/setup.png"
}

// New creates the social plugin.  All Config fields are optional.
// Bool fields default to true (Enabled, Cache, Cards, DebugGrid) unless
// the caller passes a Config with explicit zero values via WithDisabled /
// WithCacheDisabled helpers, so we use a sentinel-based approach: callers
// should use NewWithDefaults() when they want all defaults applied.
func New(cfg Config) *Plugin {
	return newWithDefaults(cfg, false)
}

// NewWithDefaults creates the plugin and applies the default value for every
// unset field, including boolean fields whose zero value (false) would
// otherwise be ambiguous.
func NewWithDefaults(overrides Config) *Plugin {
	return newWithDefaults(overrides, true)
}

func newWithDefaults(cfg Config, applyBoolDefaults bool) *Plugin {
	d := defaultConfig()

	if cfg.Concurrency == 0 {
		cfg.Concurrency = d.Concurrency
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = d.CacheDir
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = d.LogLevel
	}
	if cfg.CardsDir == "" {
		cfg.CardsDir = d.CardsDir
	}
	if cfg.CardsLayoutDir == "" {
		cfg.CardsLayoutDir = d.CardsLayoutDir
	}
	if cfg.CardsLayout == "" {
		cfg.CardsLayout = d.CardsLayout
	}
	if cfg.DebugGridStep == 0 {
		cfg.DebugGridStep = d.DebugGridStep
	}
	if cfg.DebugColor == "" {
		cfg.DebugColor = d.DebugColor
	}
	if cfg.CardsLayoutOptions.FontFamily == "" {
		cfg.CardsLayoutOptions.FontFamily = d.CardsLayoutOptions.FontFamily
	}
	if applyBoolDefaults {
		if !cfg.Enabled {
			cfg.Enabled = d.Enabled
		}
		if !cfg.Cache {
			cfg.Cache = d.Cache
		}
		if !cfg.Cards {
			cfg.Cards = d.Cards
		}
		if !cfg.DebugGrid {
			cfg.DebugGrid = d.DebugGrid
		}
	}

	return &Plugin{cfg: cfg}
}

// --- core.Plugin -------------------------------------------------------------

func (p *Plugin) Name() string { return "social" }

// OnInit captures site-level context needed for URL construction.
func (p *Plugin) OnInit(ctx *core.BuildContext) error {
	p.outDir = ctx.OutputDir
	p.mu.Lock()
	p.pages = p.pages[:0]
	p.mu.Unlock()
	return nil
}

// OnPageRendered records the page for card generation and injects <meta> tags.
func (p *Plugin) OnPageRendered(page *core.PageResult, html string) (string, error) {
	if !p.cfg.Enabled || !p.cfg.Cards {
		return html, nil
	}
	if p.excluded(page.FilePath) {
		return html, nil
	}
	if !p.included(page.FilePath) {
		return html, nil
	}

	opts := p.cfg.CardsLayoutOptions // page-level overrides not yet implemented

	// Derive card relative path from the page URL.
	cardRel := cardRelPath(page.URL)

	p.mu.Lock()
	p.pages = append(p.pages, pageJob{result: page, opts: opts, cardRel: cardRel})
	p.mu.Unlock()

	// Inject Open Graph / Twitter Card meta tags into <head>.
	cardURL := p.cardPublicURL(cardRel)
	meta := p.metaTags(page, cardURL)
	html = strings.Replace(html, "</head>", meta+"</head>", 1)
	return html, nil
}

// OnBuildComplete generates all queued cards in parallel.
func (p *Plugin) OnBuildComplete(ctx *core.BuildContext) error {
	if !p.cfg.Enabled || !p.cfg.Cards {
		return nil
	}

	p.mu.Lock()
	jobs := make([]pageJob, len(p.pages))
	copy(jobs, p.pages)
	p.mu.Unlock()

	if len(jobs) == 0 {
		return nil
	}

	workers := p.workers()
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var errs []string

	for _, job := range jobs {
		job := job
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if err := p.generateCard(job); err != nil {
				p.logError(err, job.result.FilePath, &errMu, &errs)
			}
		}()
	}
	wg.Wait()

	if len(errs) > 0 && p.cfg.LogLevel != "ignore" {
		diagnostics.PrintError(fmt.Errorf("%s", strings.Join(errs, "; ")), "social: card generation errors")
	}
	return nil
}

// --- card generation ---------------------------------------------------------

func (p *Plugin) generateCard(job pageJob) error {
	outPath := filepath.Join(p.outDir, filepath.FromSlash(p.cfg.CardsDir), filepath.FromSlash(job.cardRel))

	params := p.resolveParams(job)

	// Cache check: skip re-generation if the key file matches.
	if p.cfg.Cache {
		key := cacheKey(params)
		cacheFile := filepath.Join(p.cfg.CacheDir, job.cardRel+".key")
		if existingKey, err := os.ReadFile(cacheFile); err == nil && string(existingKey) == key {
			// Key matches — check whether the output file still exists.
			if _, err := os.Stat(outPath); err == nil {
				return nil // cache hit
			}
		}
		// Store key after successful generation (see below).
		defer func() {
			_ = os.MkdirAll(filepath.Dir(cacheFile), 0o755)
			_ = os.WriteFile(cacheFile, []byte(key), 0o644)
		}()
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("create card dir: %w", err)
	}

	var buf bytes.Buffer
	if err := renderCard(&buf, params); err != nil {
		return fmt.Errorf("render %s: %w", job.cardRel, err)
	}

	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

func (p *Plugin) resolveParams(job pageJob) cardParams {
	opts := job.opts
	page := job.result

	bgColor := mustParseColor(opts.BackgroundColor, color.RGBA{R: 0x17, G: 0x6b, B: 0xfb, A: 0xff})
	fgColor := mustParseColor(opts.Color, color.White)

	title := page.Title
	if opts.Title != "" {
		title = opts.Title
	}
	description := opts.Description

	debugColor := mustParseColor(p.cfg.DebugColor, color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff})
	debug := p.cfg.Debug && p.cfg.DebugOnBuild

	return cardParams{
		Title:       title,
		Description: description,
		BgColor:     bgColor,
		FgColor:     fgColor,
		LogoPath:    opts.Logo,
		BgImagePath: opts.BackgroundImage,
		Debug:       debug,
		DebugGrid:   p.cfg.DebugGrid,
		DebugStep:   p.cfg.DebugGridStep,
		DebugColor:  debugColor,
	}
}

// --- meta tag injection ------------------------------------------------------

func (p *Plugin) metaTags(page *core.PageResult, cardURL string) string {
	if cardURL == "" {
		return ""
	}
	title := htmlEsc(page.Title)
	var b strings.Builder
	// Open Graph
	fmt.Fprintf(&b, "\n  <meta property=\"og:type\" content=\"website\">")
	fmt.Fprintf(&b, "\n  <meta property=\"og:title\" content=\"%s\">", title)
	fmt.Fprintf(&b, "\n  <meta property=\"og:image\" content=\"%s\">", htmlEsc(cardURL))
	fmt.Fprintf(&b, "\n  <meta property=\"og:image:width\" content=\"%d\">", cardWidth)
	fmt.Fprintf(&b, "\n  <meta property=\"og:image:height\" content=\"%d\">", cardHeight)
	// Twitter Card
	fmt.Fprintf(&b, "\n  <meta name=\"twitter:card\" content=\"summary_large_image\">")
	fmt.Fprintf(&b, "\n  <meta name=\"twitter:title\" content=\"%s\">", title)
	fmt.Fprintf(&b, "\n  <meta name=\"twitter:image\" content=\"%s\">", htmlEsc(cardURL))
	return b.String()
}

func (p *Plugin) cardPublicURL(cardRel string) string {
	if p.siteURL != "" {
		return strings.TrimRight(p.siteURL, "/") + "/" +
			strings.TrimLeft(filepath.ToSlash(p.cfg.CardsDir)+"/"+cardRel, "/")
	}
	// Relative URL — useful when site_url is not set.
	return "/" + filepath.ToSlash(p.cfg.CardsDir) + "/" + cardRel
}

// --- filtering ---------------------------------------------------------------

func (p *Plugin) included(filePath string) bool {
	if len(p.cfg.CardsInclude) == 0 {
		return true
	}
	for _, pattern := range p.cfg.CardsInclude {
		if utils.MatchesExclude(relFilePath(filePath), []string{pattern}) {
			return true
		}
	}
	return false
}

func (p *Plugin) excluded(filePath string) bool {
	if len(p.cfg.CardsExclude) == 0 {
		return false
	}
	return utils.MatchesExclude(relFilePath(filePath), p.cfg.CardsExclude)
}

func relFilePath(absPath string) string {
	// Strip leading path separators for glob matching.
	return filepath.ToSlash(strings.TrimLeft(absPath, string(os.PathSeparator)+"/"))
}

// --- helpers -----------------------------------------------------------------

func (p *Plugin) workers() int {
	n := p.cfg.Concurrency
	if n <= 0 {
		n = runtime.NumCPU() - 1
		if n < 1 {
			n = 1
		}
	}
	return n
}

func (p *Plugin) logError(err error, context string, mu *sync.Mutex, errs *[]string) {
	msg := fmt.Sprintf("%s: %v", context, err)
	mu.Lock()
	*errs = append(*errs, msg)
	mu.Unlock()
	if p.cfg.LogLevel == "warn" || p.cfg.LogLevel == "info" {
		diagnostics.PrintError(err, "social: "+context)
	}
}

// cardRelPath converts a page URL (output directory path) to the card filename.
// "guide/setup" → "guide/setup.png"
// "."           → "index.png"
func cardRelPath(pageURL string) string {
	u := strings.Trim(filepath.ToSlash(pageURL), "/")
	if u == "" || u == "." {
		u = "index"
	}
	h := sha256.Sum256([]byte(u))
	// Use URL path as directory structure + short hash suffix to avoid collisions.
	return u + "-" + hex.EncodeToString(h[:4]) + ".png"
}

func htmlEsc(s string) string {
	r := strings.NewReplacer(
		`&`, "&amp;",
		`"`, "&quot;",
		`<`, "&lt;",
		`>`, "&gt;",
	)
	return r.Replace(s)
}

// Compile-time interface checks.
var _ core.Plugin = (*Plugin)(nil)
var _ core.InitHook = (*Plugin)(nil)
var _ core.PageRenderedHook = (*Plugin)(nil)
var _ core.BuildCompleteHook = (*Plugin)(nil)

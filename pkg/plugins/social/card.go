package social

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

const (
	cardWidth  = 1200
	cardHeight = 630
)

// cardParams holds all resolved values needed to render one card.
type cardParams struct {
	Title       string
	Description string
	BgColor     color.Color
	FgColor     color.Color
	LogoPath    string // empty = no logo
	BgImagePath string // empty = solid background
	Debug       bool
	DebugGrid   bool
	DebugStep   int
	DebugColor  color.Color
}

// cacheKey returns a hex digest that uniquely identifies the rendered output.
// If the key is unchanged from a previous run the card can be reused.
func cacheKey(p cardParams) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%v|%v|%s|%s",
		p.Title, p.Description, p.BgColor, p.FgColor, p.LogoPath, p.BgImagePath)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// renderCard writes a 1200×630 PNG social card to w.
func renderCard(w io.Writer, p cardParams) error {
	img := image.NewRGBA(image.Rect(0, 0, cardWidth, cardHeight))

	// --- Background ---------------------------------------------------
	if p.BgImagePath != "" {
		if err := drawBackgroundImage(img, p.BgImagePath, p.BgColor); err != nil {
			// Fallback to solid color on image load failure.
			fillRect(img, img.Bounds(), p.BgColor)
		}
	} else {
		fillRect(img, img.Bounds(), p.BgColor)
	}

	// --- Logo (top-left) ---------------------------------------------
	logoSize := 48
	logoPad := 60
	if p.LogoPath != "" {
		_ = drawLogo(img, p.LogoPath, logoPad, logoPad, logoSize)
	}

	// --- Typography --------------------------------------------------
	ft, err := loadFontFace()
	if err != nil {
		return fmt.Errorf("social: load font: %w", err)
	}

	titleY := 340
	descY := 420

	// Title — bold, large
	drawText(img, ft, p.Title, p.FgColor, logoPad, titleY, 64, cardWidth-2*logoPad)

	// Description — regular, smaller
	if p.Description != "" {
		drawText(img, ft, p.Description, withAlpha(p.FgColor, 0xbb), logoPad, descY, 32, cardWidth-2*logoPad)
	}

	// --- Debug overlays ----------------------------------------------
	if p.Debug {
		if p.DebugGrid {
			drawDotGrid(img, p.DebugStep, p.DebugColor)
		}
		// Outline the title text area
		drawRect(img, image.Rect(logoPad, titleY-70, cardWidth-logoPad, titleY+10), p.DebugColor)
		// Outline the description text area
		drawRect(img, image.Rect(logoPad, descY-40, cardWidth-logoPad, descY+10), p.DebugColor)
	}

	return png.Encode(w, img)
}

// ---- helpers ----------------------------------------------------------------

func fillRect(img draw.Image, r image.Rectangle, c color.Color) {
	draw.Draw(img, r, image.NewUniform(c), image.Point{}, draw.Src)
}

func drawRect(img *image.RGBA, r image.Rectangle, c color.Color) {
	r8, g8, b8, a8 := toRGBA8(c)
	col := color.RGBA{R: r8, G: g8, B: b8, A: a8}
	for x := r.Min.X; x <= r.Max.X; x++ {
		img.SetRGBA(x, r.Min.Y, col)
		img.SetRGBA(x, r.Max.Y, col)
	}
	for y := r.Min.Y; y <= r.Max.Y; y++ {
		img.SetRGBA(r.Min.X, y, col)
		img.SetRGBA(r.Max.X, y, col)
	}
}

func drawDotGrid(img *image.RGBA, step int, c color.Color) {
	if step <= 0 {
		step = 32
	}
	r8, g8, b8, a8 := toRGBA8(c)
	col := color.RGBA{R: r8, G: g8, B: b8, A: a8}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			img.SetRGBA(x, y, col)
		}
	}
}

func drawBackgroundImage(dst *image.RGBA, path string, tint color.Color) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return err
	}
	// Scale to fill the card
	scaled := image.NewRGBA(image.Rect(0, 0, cardWidth, cardHeight))
	xdraw.BiLinear.Scale(scaled, scaled.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	draw.Draw(dst, dst.Bounds(), scaled, image.Point{}, draw.Src)

	// Apply tint with 50% opacity if provided
	if tint != nil {
		r, g, b, _ := tint.RGBA()
		tintColor := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 0x80}
		draw.Draw(dst, dst.Bounds(), image.NewUniform(tintColor), image.Point{}, draw.Over)
	}
	return nil
}

func drawLogo(dst *image.RGBA, path string, x, y, size int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return err
	}
	dstRect := image.Rect(x, y, x+size, y+size)
	scaled := image.NewRGBA(dstRect)
	xdraw.BiLinear.Scale(scaled, dstRect, src, src.Bounds(), xdraw.Over, nil)
	draw.Draw(dst, dstRect, scaled, dstRect.Min, draw.Over)
	return nil
}

// loadFontFace returns the embedded Go Regular font.
// When a real font download mechanism is added it can override this.
func loadFontFace() (*truetype.Font, error) {
	return freetype.ParseFont(goregular.TTF)
}

// drawText renders text with word-wrap inside maxWidth at (x, y).
func drawText(dst *image.RGBA, ft *truetype.Font, text string, c color.Color, x, y, size, maxWidth int) {
	ctx := freetype.NewContext()
	ctx.SetDPI(96)
	ctx.SetFont(ft)
	ctx.SetFontSize(float64(size))
	ctx.SetClip(dst.Bounds())
	ctx.SetDst(dst)
	ctx.SetSrc(image.NewUniform(c))
	ctx.SetHinting(font.HintingFull)

	face := truetype.NewFace(ft, &truetype.Options{
		Size: float64(size),
		DPI:  96,
	})
	defer face.Close()

	lineHeight := int(math.Ceil(float64(size) * 1.3))
	lines := wrapText(face, text, maxWidth)
	for i, line := range lines {
		pt := freetype.Pt(x, y+i*lineHeight)
		_, _ = ctx.DrawString(line, pt)
	}
}

// wrapText splits text into lines that fit within maxWidth pixels.
func wrapText(face font.Face, text string, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		candidate := current + " " + w
		if measureText(face, candidate) <= maxWidth {
			current = candidate
		} else {
			lines = append(lines, current)
			current = w
		}
	}
	return append(lines, current)
}

func measureText(face font.Face, s string) int {
	var advance fixed.Int26_6
	for _, r := range s {
		a, ok := face.GlyphAdvance(r)
		if ok {
			advance += a
		}
	}
	return int(advance >> 6)
}

// withAlpha returns c with its alpha replaced by a (0–255).
func withAlpha(c color.Color, a uint8) color.Color {
	r, g, b, _ := toRGBA8(c)
	return color.RGBA{R: r, G: g, B: b, A: a}
}

func toRGBA8(c color.Color) (r, g, b, a uint8) {
	r32, g32, b32, a32 := c.RGBA()
	return uint8(r32 >> 8), uint8(g32 >> 8), uint8(b32 >> 8), uint8(a32 >> 8)
}

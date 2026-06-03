package social

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// parseColor parses CSS-subset color strings into color.Color.
// Supported: #rgb, #rgba, #rrggbb, #rrggbbaa, named colors (subset).
// Returns color.Black on parse failure.
func parseColor(s string) (color.Color, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return color.Black, fmt.Errorf("empty color string")
	}
	if s == "transparent" {
		return color.RGBA{}, nil
	}

	// Named colors (small subset covering common use cases)
	if c, ok := namedColors[strings.ToLower(s)]; ok {
		return c, nil
	}

	// Hex notation
	if strings.HasPrefix(s, "#") {
		hex := s[1:]
		switch len(hex) {
		case 3: // #rgb → #rrggbb
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
			fallthrough
		case 6: // #rrggbb
			hex += "ff"
			fallthrough
		case 8: // #rrggbbaa
			v, err := strconv.ParseUint(hex, 16, 32)
			if err != nil {
				return color.Black, fmt.Errorf("invalid hex color %q: %w", s, err)
			}
			return color.RGBA{
				R: uint8(v >> 24),
				G: uint8(v >> 16),
				B: uint8(v >> 8),
				A: uint8(v),
			}, nil
		case 4: // #rgba
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2], hex[3], hex[3]})
			v, err := strconv.ParseUint(hex, 16, 32)
			if err != nil {
				return color.Black, fmt.Errorf("invalid hex color %q: %w", s, err)
			}
			return color.RGBA{
				R: uint8(v >> 24),
				G: uint8(v >> 16),
				B: uint8(v >> 8),
				A: uint8(v),
			}, nil
		}
	}

	return color.Black, fmt.Errorf("unsupported color format %q", s)
}

// mustParseColor returns the parsed color or a fallback.
func mustParseColor(s string, fallback color.Color) color.Color {
	c, err := parseColor(s)
	if err != nil {
		return fallback
	}
	return c
}

var namedColors = map[string]color.Color{
	"black":   color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff},
	"white":   color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
	"red":     color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff},
	"green":   color.RGBA{R: 0x00, G: 0x80, B: 0x00, A: 0xff},
	"blue":    color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff},
	"grey":    color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff},
	"gray":    color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff},
	"navy":    color.RGBA{R: 0x00, G: 0x00, B: 0x80, A: 0xff},
	"teal":    color.RGBA{R: 0x00, G: 0x80, B: 0x80, A: 0xff},
	"purple":  color.RGBA{R: 0x80, G: 0x00, B: 0x80, A: 0xff},
	"orange":  color.RGBA{R: 0xff, G: 0xa5, B: 0x00, A: 0xff},
	"yellow":  color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff},
	"indigo":  color.RGBA{R: 0x4b, G: 0x00, B: 0x82, A: 0xff},
	"default": color.RGBA{R: 0x17, G: 0x6b, B: 0xfb, A: 0xff},
}

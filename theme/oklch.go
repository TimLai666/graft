package theme

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gogpu/ui/widget"
)

// OKLCH converts an OKLCH color to a widget.Color (sRGB, gamma-encoded).
//
// Arguments mirror the CSS oklch() function: l is perceptual lightness in
// [0, 1], c is chroma (0 = gray, ~0.37 = max sRGB chroma), h is hue in
// degrees. Colors outside the sRGB gamut are clipped per linear channel —
// the same fallback browsers and Tailwind v4 use on sRGB displays — so
// out-of-gamut tokens like shadcn's destructive red render to the exact
// same pixels as in Chrome (oklch(0.577 0.245 27.325) → #e7000b).
//
// This allows shadcn/ui token values to be copied verbatim:
//
//	primary := theme.OKLCH(0.205, 0, 0) // --primary: oklch(0.205 0 0)
func OKLCH(l, c, h float64) widget.Color {
	return OKLCHA(l, c, h, 1)
}

// OKLCHA is OKLCH with an explicit alpha in [0, 1].
func OKLCHA(l, c, h, alpha float64) widget.Color {
	r, g, b := oklchToSRGB(l, c, h)
	return widget.Color{
		R: float32(clamp01(r)),
		G: float32(clamp01(g)),
		B: float32(clamp01(b)),
		A: float32(clamp01(alpha)),
	}
}

// ParseOKLCH parses a CSS oklch() string such as
//
//	"oklch(0.205 0 0)"
//	"oklch(0.577 0.245 27.325)"
//	"oklch(0.708 0 0 / 50%)"
//	"oklch(1 0 0 / 10%)"
//
// so values can be pasted directly from shadcn/ui theme CSS.
func ParseOKLCH(s string) (widget.Color, error) {
	body, ok := strings.CutPrefix(strings.TrimSpace(s), "oklch(")
	if !ok {
		return widget.Color{}, fmt.Errorf("theme: not an oklch() value: %q", s)
	}
	body, ok = strings.CutSuffix(body, ")")
	if !ok {
		return widget.Color{}, fmt.Errorf("theme: unterminated oklch() value: %q", s)
	}

	main, alphaPart, hasAlpha := strings.Cut(body, "/")
	fields := strings.Fields(main)
	if len(fields) != 3 {
		return widget.Color{}, fmt.Errorf("theme: oklch() needs 3 components, got %d: %q", len(fields), s)
	}

	l, err := parseComponent(fields[0], true)
	if err != nil {
		return widget.Color{}, fmt.Errorf("theme: bad oklch lightness in %q: %w", s, err)
	}
	c, err := parseComponent(fields[1], false)
	if err != nil {
		return widget.Color{}, fmt.Errorf("theme: bad oklch chroma in %q: %w", s, err)
	}
	h, err := parseHue(fields[2])
	if err != nil {
		return widget.Color{}, fmt.Errorf("theme: bad oklch hue in %q: %w", s, err)
	}

	alpha := 1.0
	if hasAlpha {
		alpha, err = parseComponent(strings.TrimSpace(alphaPart), true)
		if err != nil {
			return widget.Color{}, fmt.Errorf("theme: bad oklch alpha in %q: %w", s, err)
		}
	}
	return OKLCHA(l, c, h, alpha), nil
}

// MustOKLCH is ParseOKLCH that panics on malformed input. Intended for
// package-level token tables where the input is a compile-time constant.
func MustOKLCH(s string) widget.Color {
	c, err := ParseOKLCH(s)
	if err != nil {
		panic(err)
	}
	return c
}

// parseComponent parses a number or percentage. When percentRelative is
// true, "40%" means 0.4; otherwise percentages scale the reference chroma
// 0.4 (per CSS Color 4 for oklch chroma).
func parseComponent(s string, percentRelative bool) (float64, error) {
	if p, ok := strings.CutSuffix(s, "%"); ok {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return 0, err
		}
		if percentRelative {
			return v / 100, nil
		}
		return v / 100 * 0.4, nil
	}
	return strconv.ParseFloat(s, 64)
}

func parseHue(s string) (float64, error) {
	s = strings.TrimSuffix(s, "deg")
	if s == "none" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

// oklchToSRGB converts OKLCH to gamma-encoded sRGB components (unclamped),
// using Björn Ottosson's reference OKLab matrices.
func oklchToSRGB(l, c, h float64) (r, g, b float64) {
	hRad := h * math.Pi / 180
	a := c * math.Cos(hRad)
	bb := c * math.Sin(hRad)
	return oklabToSRGB(l, a, bb)
}

func oklabToSRGB(l, a, b float64) (float64, float64, float64) {
	l_ := l + 0.3963377774*a + 0.2158037573*b
	m_ := l - 0.1055613458*a - 0.0638541728*b
	s_ := l - 0.0894841775*a - 1.2914855480*b

	ll := l_ * l_ * l_
	mm := m_ * m_ * m_
	ss := s_ * s_ * s_

	rLin := +4.0767416621*ll - 3.3077115913*mm + 0.2309699292*ss
	gLin := -1.2684380046*ll + 2.6097574011*mm - 0.3413193965*ss
	bLin := -0.0041960863*ll - 0.7034186147*mm + 1.7076147010*ss

	// Clip in linear space: matches browser/Tailwind sRGB fallback exactly.
	return linearToGamma(clamp01(rLin)), linearToGamma(clamp01(gLin)), linearToGamma(clamp01(bLin))
}

func linearToGamma(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1/2.4) - 0.055
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

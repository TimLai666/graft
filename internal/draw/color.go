package draw

import (
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
)

// Alpha returns c with its alpha channel replaced by a.
//
// Use it for opaque tokens that shadcn draws at a fixed opacity, e.g.
// bg-primary/90 is Alpha(t.Primary, 0.9). For tokens that already carry
// alpha (dark-mode border/input) use [MulAlpha] instead.
func Alpha(c widget.Color, a float32) widget.Color {
	c.A = a
	return c
}

// MulAlpha returns c with its existing alpha multiplied by f.
//
// Use it for tokens that already carry alpha, such as the dark-mode
// border token oklch(1 0 0 / 10%): dark outline hover bg-input/50 is
// MulAlpha(t.Input, 0.5), not Alpha(t.Input, 0.5).
func MulAlpha(c widget.Color, f float32) widget.Color {
	c.A *= f
	return c
}

// Fade dims c for the disabled state (disabled:opacity-50).
//
// When disabled is false c is returned unchanged; when true the alpha is
// multiplied by metrics.DisabledOpacity. Painters apply Fade to every color
// they draw (fills, borders, text, icons) because the canvas has no
// whole-subtree opacity layer (DESIGN.md section 5.3).
func Fade(c widget.Color, disabled bool) widget.Color {
	if disabled {
		return MulAlpha(c, metrics.DisabledOpacity)
	}
	return c
}

// Over composites fg over bg using source-over blending of straight
// (non-premultiplied) alpha colors and returns the straight-alpha result.
//
// This mirrors what the canvas does when fg is drawn on top of an already
// painted bg, so spec tests can compute expected pixels, and callers can
// precomposite where a single flat color is needed.
func Over(bg, fg widget.Color) widget.Color {
	fa := clamp01(fg.A)
	ba := clamp01(bg.A)
	outA := fa + ba*(1-fa)
	if outA <= 0 {
		return widget.Color{}
	}
	// Weight each straight-alpha channel by its coverage, then un-premultiply.
	bw := ba * (1 - fa)
	return widget.Color{
		R: (fg.R*fa + bg.R*bw) / outA,
		G: (fg.G*fa + bg.G*bw) / outA,
		B: (fg.B*fa + bg.B*bw) / outA,
		A: outA,
	}
}

// clamp01 clamps v to the range [0, 1].
func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

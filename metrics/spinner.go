package metrics

import (
	"math"
	"time"
)

// Spinner holds the constants for the shadcn Spinner
// (docs/research/03-shadcn-pixel-spec.md §5 "Spinner").
//
// Source (new-york-v4):
//
//	<Loader2Icon className="size-4 animate-spin" />
//	animation: spin 1s linear infinite   // @keyframes spin { to { rotate: 360deg } }
//
// shadcn renders a spinning lucide Loader2 (a circle with a ~90° gap). The
// gogpu/ui canvas has no rotation transform (Report 1 §7), so graft draws the
// arc directly with StrokeArc and advances the start angle over time instead
// of rotating an icon.
//
// The color (currentColor → Foreground by default) routes through the theme;
// only the geometry and timing are literals here.
var Spinner = struct {
	// DefaultSize is the spinner side length in px (size-4 = 16px).
	DefaultSize float32

	// SpinPeriod is the full rotation period (animate-spin = 1s linear).
	SpinPeriod time.Duration

	// SweepAngle is the visible arc sweep in radians. Lucide's loader-circle
	// leaves roughly a quarter-circle gap, so the arc spans ~270°.
	SweepAngle float64

	// StartAngle is the base arc start angle in radians at rotation phase 0,
	// measured clockwise from the +x axis. -90° puts the leading cap at the
	// top of the circle.
	StartAngle float64

	// StrokeRatio is the arc stroke width as a fraction of the spinner size.
	// Lucide draws its 24-unit icons at stroke-width 2, i.e. 1/12 of the box.
	StrokeRatio float32
}{
	DefaultSize: 16,
	SpinPeriod:  time.Second,
	SweepAngle:  math.Pi * 1.5, // 270°
	StartAngle:  -math.Pi / 2,  // top of the circle
	StrokeRatio: 2.0 / 24.0,
}

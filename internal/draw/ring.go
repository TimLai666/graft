package draw

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
)

// FocusRing draws the standard shadcn focus ring: a 3px band hugging the
// outside of bounds (focus-visible:ring-[3px], DESIGN.md section 5.2).
//
// The ring color is passed in already resolved (typically
// Alpha(t.Ring, metrics.RingAlpha), or the destructive variant for the
// invalid state). Canvas strokes are center-drawn, so the band occupying
// [bounds, bounds+3] is a width-3 stroke at Expand(1.5) with radius+1.5.
func FocusRing(c widget.Canvas, bounds geometry.Rect, radius float32, ring widget.Color) {
	FocusRingWidth(c, bounds, radius, ring, metrics.RingWidth)
}

// FocusRingWidth draws a focus ring band of arbitrary width outside bounds.
//
// It generalizes [FocusRing] for the components that deviate from the 3px
// standard, e.g. the slider thumb uses metrics.SliderRingWidth.
func FocusRingWidth(c widget.Canvas, bounds geometry.Rect, radius float32, ring widget.Color, width float32) {
	if width <= 0 {
		return
	}
	c.StrokeRoundRect(bounds.Expand(width/2), ring, radius+width/2, width)
}

// OffsetRing draws the legacy "ring-N ring-offset-M" recipe used by the
// Dialog/Sheet close button (DESIGN.md section 5.2): a ring band of the
// given width separated from bounds by an offset gap, with the gap filled
// in gapFill (the theme Background) to reproduce ring-offset compositing.
//
// Band layout outside bounds: [bounds, bounds+offset] is gapFill,
// [bounds+offset, bounds+offset+width] is ring.
func OffsetRing(c widget.Canvas, bounds geometry.Rect, radius float32, ring widget.Color, width, offset float32, gapFill widget.Color) {
	if offset > 0 && gapFill.A > 0 {
		c.StrokeRoundRect(bounds.Expand(offset/2), gapFill, radius+offset/2, offset)
	}
	if width > 0 {
		c.StrokeRoundRect(bounds.Expand(offset+width/2), ring, radius+offset+width/2, width)
	}
}

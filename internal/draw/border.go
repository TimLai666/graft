package draw

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// InsideBorder strokes a border that sits entirely inside bounds, matching
// CSS box-sizing: border-box (DESIGN.md section 5.4).
//
// Canvas strokes are center-drawn, so the stroke is inset by w/2 and the
// radius reduced by w/2 (clamped at zero). All 1px component borders in
// graft go through this helper.
//
// NOTE: the GPU scene renderer currently fills thin round-rect strokes as solid
// boxes (gg bug, reported upstream). Prefer [BorderFill] for any element that
// has a fill behind the border — it composes the border from convex fills,
// which render correctly. InsideBorder remains for border-only cases (no fill
// to redraw, e.g. focus rings and separators) where the artifact is absent or
// acceptable.
func InsideBorder(c widget.Canvas, bounds geometry.Rect, radius float32, col widget.Color, w float32) {
	if w <= 0 {
		return
	}
	r := radius - w/2
	if r < 0 {
		r = 0
	}
	c.StrokeRoundRect(bounds.Expand(-w/2), col, r, w)
}

// BorderFill paints a filled rounded rect with a w-px inner (border-box) border,
// composed from two convex fills: an outer round-rect in the border color, then
// an inset round-rect in the fill color. This avoids the canvas stroke path,
// which the GPU scene renderer mis-renders for thin strokes (fills the whole
// shape solid). Call it where the element fill would be drawn; paint content on
// top afterward. With w<=0 it degrades to a plain filled round-rect.
//
// The fill must be opaque for the border to read as a thin ring (a translucent
// fill lets the underlying border color show through the interior).
func BorderFill(c widget.Canvas, bounds geometry.Rect, fill, border widget.Color, radius, w float32) {
	if w <= 0 {
		c.DrawRoundRect(bounds, fill, radius)
		return
	}
	c.DrawRoundRect(bounds, border, radius)
	ir := radius - w
	if ir < 0 {
		ir = 0
	}
	c.DrawRoundRect(bounds.Expand(-w), fill, ir)
}

// Corners is a bitmask selecting rectangle corners for [SquareCorners].
type Corners uint8

// Corner flags. Combine with bitwise OR, e.g. TopLeft|BottomLeft selects
// the left edge corners.
const (
	// TopLeft selects the top-left corner.
	TopLeft Corners = 1 << iota

	// TopRight selects the top-right corner.
	TopRight

	// BottomLeft selects the bottom-left corner.
	BottomLeft

	// BottomRight selects the bottom-right corner.
	BottomRight

	// AllCorners selects every corner.
	AllCorners = TopLeft | TopRight | BottomLeft | BottomRight
)

// Has reports whether the mask includes all corners in c.
func (cs Corners) Has(c Corners) bool {
	return cs&c == c
}

// SquareCorners fills a rounded rectangle, then overpaints the selected
// corner quadrants with plain rectangles of the same fill so those corners
// render square (DESIGN.md section 5.4).
//
// Per-corner radii do not exist on the canvas; this compositing trick gives
// ToggleGroup/ButtonGroup their fused segments (only the outer corners of
// the first/last item stay rounded). The fill must be opaque for the
// overpaint to be invisible — group fills are opaque token colors.
func SquareCorners(c widget.Canvas, bounds geometry.Rect, radius float32, fill widget.Color, corners Corners) {
	c.DrawRoundRect(bounds, fill, radius)
	if radius <= 0 || corners == 0 {
		return
	}
	// The renderer clamps the round-rect radius to half the smaller
	// dimension; clamp the quadrant size identically so the overpaint
	// covers exactly the rounded region.
	q := radius
	if half := bounds.Width() / 2; q > half {
		q = half
	}
	if half := bounds.Height() / 2; q > half {
		q = half
	}
	if q <= 0 {
		return
	}
	if corners.Has(TopLeft) {
		c.DrawRect(geometry.NewRect(bounds.Min.X, bounds.Min.Y, q, q), fill)
	}
	if corners.Has(TopRight) {
		c.DrawRect(geometry.NewRect(bounds.Max.X-q, bounds.Min.Y, q, q), fill)
	}
	if corners.Has(BottomLeft) {
		c.DrawRect(geometry.NewRect(bounds.Min.X, bounds.Max.Y-q, q, q), fill)
	}
	if corners.Has(BottomRight) {
		c.DrawRect(geometry.NewRect(bounds.Max.X-q, bounds.Max.Y-q, q, q), fill)
	}
}

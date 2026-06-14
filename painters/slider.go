package painters

import (
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// sliderFallbackTheme backs nil-theme painters (painters.New(nil)) with the
// stock neutral palette, per the Painters bundle contract.
var sliderFallbackTheme = theme.New()

// PaintSlider renders a core/slider in shadcn style
// (metrics.Slider; docs/research/03-shadcn-pixel-spec.md section 5 "Slider"):
//
//   - track: 4px pill in --muted spanning the full main axis,
//   - range: --primary fill from the start to the thumb, square leading
//     edge via the track clip (overflow-hidden),
//   - thumb: 12px circle, always-white fill in both modes (bg-white),
//     1px --primary border, shadow-sm,
//   - hover/drag/focus: 4px ring in --ring at 50% alpha around the thumb
//     (hover:ring-4 focus-visible:ring-4 ring-ring/50),
//   - disabled: every color at 50% opacity, no ring (disabled:opacity-50).
//
// Colors resolve from the theme at paint time; PaintState.ColorScheme is
// ignored (graft resolves tokens itself, DESIGN.md section 8 item 13).
func (p Slider) PaintSlider(canvas widget.Canvas, ps slider.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}
	th := p.Theme
	if th == nil {
		th = sliderFallbackTheme
	}
	tok := th.Active()

	muted := draw.Fade(tok.Muted, ps.Disabled)
	primary := draw.Fade(tok.Primary, ps.Disabled)
	thumbFill := draw.Fade(widget.ColorWhite, ps.Disabled) // bg-white, both modes

	if ps.Orientation == slider.Vertical {
		p.paintVertical(canvas, ps, tok, muted, primary, thumbFill)
		return
	}

	bounds := ps.Bounds
	cy := bounds.Min.Y + bounds.Height()/2
	half := metrics.Slider.TrackThickness / 2
	trackRect := geometry.NewRect(bounds.Min.X, cy-half, bounds.Width(), metrics.Slider.TrackThickness)

	// Track (bg-muted, rounded-full).
	canvas.DrawRoundRect(trackRect, muted, metrics.Slider.TrackRadius)

	// Thumb center travels inset by half the thumb so the thumb stays
	// inside the track at the extremes, like Radix.
	thumbR := metrics.Slider.ThumbSize / 2
	travel := bounds.Width() - metrics.Slider.ThumbSize
	thumbX := bounds.Min.X + thumbR + ps.Progress*travel
	center := geometry.Pt(thumbX, cy)

	// Range (bg-primary) clipped to the track pill: square edge at the
	// thumb, rounded start (overflow-hidden on the track).
	if thumbX > bounds.Min.X {
		canvas.PushClipRoundRect(trackRect, metrics.Slider.TrackRadius)
		canvas.DrawRect(geometry.NewRect(bounds.Min.X, cy-half, thumbX-bounds.Min.X, metrics.Slider.TrackThickness), primary)
		canvas.PopClip()
	}

	p.paintThumb(canvas, ps, tok, primary, thumbFill, center)
}

// paintVertical mirrors the horizontal painting along the Y axis
// (progress 0 at the bottom, like core/slider's hit testing).
func (p Slider) paintVertical(canvas widget.Canvas, ps slider.PaintState, tok *theme.Tokens, muted, primary, thumbFill widget.Color) {
	bounds := ps.Bounds
	cx := bounds.Min.X + bounds.Width()/2
	half := metrics.Slider.TrackThickness / 2
	trackRect := geometry.NewRect(cx-half, bounds.Min.Y, metrics.Slider.TrackThickness, bounds.Height())

	canvas.DrawRoundRect(trackRect, muted, metrics.Slider.TrackRadius)

	thumbR := metrics.Slider.ThumbSize / 2
	travel := bounds.Height() - metrics.Slider.ThumbSize
	thumbY := bounds.Max.Y - thumbR - ps.Progress*travel
	center := geometry.Pt(cx, thumbY)

	if thumbY < bounds.Max.Y {
		canvas.PushClipRoundRect(trackRect, metrics.Slider.TrackRadius)
		canvas.DrawRect(geometry.NewRect(cx-half, thumbY, metrics.Slider.TrackThickness, bounds.Max.Y-thumbY), primary)
		canvas.PopClip()
	}

	p.paintThumb(canvas, ps, tok, primary, thumbFill, center)
}

// paintThumb draws the thumb shadow, ring, fill, and border.
func (p Slider) paintThumb(canvas widget.Canvas, ps slider.PaintState, tok *theme.Tokens, primary, thumbFill widget.Color, center geometry.Point) {
	thumbR := metrics.Slider.ThumbSize / 2
	thumbRect := geometry.NewRect(center.X-thumbR, center.Y-thumbR, metrics.Slider.ThumbSize, metrics.Slider.ThumbSize)

	// shadow-sm under the thumb (faded with the rest when disabled).
	shadow := metrics.ShadowSM
	if ps.Disabled {
		faded := make([]metrics.ShadowLayer, len(shadow))
		for i, l := range shadow {
			l.Alpha *= metrics.DisabledOpacity
			faded[i] = l
		}
		shadow = faded
	}
	draw.Shadow(canvas, thumbRect, thumbR, shadow)

	// hover:ring-4 focus-visible:ring-4 ring-ring/50 — band hugging the
	// thumb outside edge; never on a disabled slider. Dragging keeps the
	// ring (the thumb holds focus while dragged in the browser too).
	if !ps.Disabled && (ps.Hovered || ps.Dragging || ps.Focused) {
		ring := draw.Alpha(tok.Ring, metrics.RingAlpha)
		canvas.StrokeCircle(center, thumbR+metrics.SliderRingWidth/2, ring, metrics.SliderRingWidth)
	}

	// Thumb fill (bg-white) + inside 1px --primary border.
	canvas.DrawCircle(center, thumbR, thumbFill)
	canvas.StrokeCircle(center, thumbR-metrics.Slider.ThumbBorderWidth/2, primary, metrics.Slider.ThumbBorderWidth)
}

// Compile-time interface check.
var _ slider.Painter = Slider{}

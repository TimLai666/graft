package graft

import (
	"math"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// RangeSliderWidget is the shadcn Slider in its two-thumb RANGE form
// (value={[lo, hi]}): a track with a Primary fill BETWEEN two circular
// thumbs (docs/research/03-shadcn-pixel-spec.md §5 "Slider").
//
// OWNED widget (DESIGN.md §3.2): unlike the single-value SliderWidget, which
// wraps gogpu/ui core/slider, this control draws and drags itself. The core
// slider is single-value only (core/slider/config.go has one `value
// float32`, no array), so a range needs a graft-owned reimplementation. It
// reuses metrics.Slider so the two thumbs are pixel-identical to the single
// slider's thumb, and the same TrackThickness/TrackRadius so a row of single
// and range sliders reads as one family.
//
// SCREEN-SPACE DRAG (the bug class that bit the single slider): on press the
// widget captures the pointer (widget.PointerCapturer), after which gogpu's
// Window delivers MouseMove/Release DIRECTLY to this widget with RAW SCREEN
// coordinates, bypassing parent translation (app/window.go). To make the
// captured path and the normal (parent-local) path agree, every mouse
// position is converted to SCREEN space before mapping it to a value, and the
// value mapping runs against the track in SCREEN space (origin =
// ScreenOrigin(), which Draw stamps each frame). Painting, by contrast, is
// done in this widget's own local space via PushTransform(Bounds().Min), so
// the rendered pixels are independent of the screen offset.
type RangeSliderWidget struct {
	widget.WidgetBase

	th *theme.Theme

	min, max, step float32
	lo, hi         float32

	loSig, hiSig state.Signal[float32]
	onChange     func(lo, hi float32)

	disabled bool
	width    float32

	// active is the thumb currently being dragged / keyboard-targeted:
	// thumbNone when idle, thumbLo or thumbHi otherwise.
	active   thumbID
	dragging bool
	// hovered is the thumb under the cursor for the hover ring (thumbNone
	// when the cursor is off both thumbs).
	hovered thumbID
}

// thumbID identifies which of the two thumbs an interaction targets.
type thumbID uint8

const (
	thumbNone thumbID = iota
	thumbLo
	thumbHi
)

// RangeSlider creates a two-thumb range slider over [0, 100] with the thumbs
// at 0 and 100, matching shadcn's defaults.
func RangeSlider() *RangeSliderWidget {
	s := &RangeSliderWidget{th: CurrentTheme(), max: 100, hi: 100}
	s.SetVisible(true)
	s.SetEnabled(true)
	return s
}

// Min sets the minimum value (default 0).
func (s *RangeSliderWidget) Min(v float32) *RangeSliderWidget { s.min = v; return s }

// Max sets the maximum value (default 100).
func (s *RangeSliderWidget) Max(v float32) *RangeSliderWidget { s.max = v; return s }

// Step sets the snap increment; 0 (default) is continuous.
func (s *RangeSliderWidget) Step(v float32) *RangeSliderWidget { s.step = v; return s }

// Values sets the initial (uncontrolled) low and high thumb values. They are
// normalized so lo<=hi and both fall within [min, max].
func (s *RangeSliderWidget) Values(lo, hi float32) *RangeSliderWidget {
	s.lo, s.hi = lo, hi
	s.normalize()
	return s
}

// BindRange makes the slider controlled by a pair of signals: the thumbs
// reflect loSig/hiSig and dragging writes back to them. The bindings are
// registered in Mount.
func (s *RangeSliderWidget) BindRange(loSig, hiSig state.Signal[float32]) *RangeSliderWidget {
	s.loSig, s.hiSig = loSig, hiSig
	if loSig != nil {
		s.lo = loSig.Get()
	}
	if hiSig != nil {
		s.hi = hiSig.Get()
	}
	s.normalize()
	return s
}

// OnChange registers an observer fired with (lo, hi) on every change.
func (s *RangeSliderWidget) OnChange(fn func(lo, hi float32)) *RangeSliderWidget {
	s.onChange = fn
	return s
}

// Disabled sets the disabled state (50% opacity, no interaction).
func (s *RangeSliderWidget) Disabled(v bool) *RangeSliderWidget { s.disabled = v; return s }

// W pins an explicit width in px instead of stretching to the container.
func (s *RangeSliderWidget) W(px float32) *RangeSliderWidget { s.width = px; return s }

// Theme pins a specific theme instead of the snapshotted current theme.
func (s *RangeSliderWidget) Theme(th *theme.Theme) *RangeSliderWidget { s.th = th; return s }

// Lo reports the current low value (signal wins when bound).
func (s *RangeSliderWidget) Lo() float32 {
	if s.loSig != nil {
		return s.loSig.Get()
	}
	return s.lo
}

// Hi reports the current high value (signal wins when bound).
func (s *RangeSliderWidget) Hi() float32 {
	if s.hiSig != nil {
		return s.hiSig.Get()
	}
	return s.hi
}

func (s *RangeSliderWidget) resolvedTheme() *theme.Theme {
	if s.th != nil {
		return s.th
	}
	return CurrentTheme()
}

// IsFocusable reports whether the slider can take keyboard focus.
func (s *RangeSliderWidget) IsFocusable() bool {
	return s.IsVisible() && s.IsEnabled() && !s.disabled
}

// Layout sizes the slider: explicit .W, else the bounded container width
// (w-full), else metrics.Slider.DefaultWidth; height is the 12px thumb.
//
// Layout seeds the screen origin from the last-known position so the value
// mapping is sane before the first Draw; Draw re-stamps it to the current
// frame's screen origin.
func (s *RangeSliderWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	w := s.width
	if w <= 0 {
		if c.MaxWidth > 0 && c.MaxWidth < geometry.Infinity {
			w = c.MaxWidth
		} else {
			w = metrics.Slider.DefaultWidth
		}
	}
	size := c.Constrain(geometry.Sz(w, metrics.Slider.Height))
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

// Draw paints the track, the Primary fill between the two thumbs, and the two
// thumbs (each with shadow, optional ring, white fill, Primary border).
//
// Painting happens in this widget's local space: PushTransform(Bounds().Min)
// then shapes are computed against a (0,0)-origin track, so the rendered
// pixels never depend on the screen offset. Colors resolve from the active
// token set here so a mode switch repaints without rebuilding.
func (s *RangeSliderWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !s.IsVisible() {
		return
	}
	th := s.resolvedTheme()
	tok := th.Active()
	size := s.Bounds().Size()

	canvas.PushTransform(s.Bounds().Min)
	defer canvas.PopTransform()

	cy := size.Height / 2
	half := metrics.Slider.TrackThickness / 2
	trackRect := geometry.NewRect(0, cy-half, size.Width, metrics.Slider.TrackThickness)

	muted := draw.Fade(tok.Muted, s.disabled)
	primary := draw.Fade(tok.Primary, s.disabled)
	thumbFill := draw.Fade(widget.ColorWhite, s.disabled) // bg-white, both modes

	// Track (bg-muted, rounded-full).
	canvas.DrawRoundRect(trackRect, muted, metrics.Slider.TrackRadius)

	loX := s.thumbX(size.Width, s.Lo())
	hiX := s.thumbX(size.Width, s.Hi())

	// Range fill (bg-primary) BETWEEN the two thumbs, clipped to the track
	// pill (overflow-hidden) so its ends follow the pill's rounding.
	if hiX > loX {
		canvas.PushClipRoundRect(trackRect, metrics.Slider.TrackRadius)
		canvas.DrawRect(geometry.NewRect(loX, cy-half, hiX-loX, metrics.Slider.TrackThickness), primary)
		canvas.PopClip()
	}

	// Draw the inactive thumb first so the active one (with its ring) wins on
	// overlap at the extremes.
	if s.active == thumbHi {
		s.paintThumb(canvas, tok, primary, thumbFill, geometry.Pt(loX, cy), thumbLo)
		s.paintThumb(canvas, tok, primary, thumbFill, geometry.Pt(hiX, cy), thumbHi)
	} else {
		s.paintThumb(canvas, tok, primary, thumbFill, geometry.Pt(hiX, cy), thumbHi)
		s.paintThumb(canvas, tok, primary, thumbFill, geometry.Pt(loX, cy), thumbLo)
	}
	_ = ctx
}

// paintThumb draws one thumb: shadow, optional hover/focus/drag ring (only on
// the active or hovered thumb), white fill, 1px Primary border.
func (s *RangeSliderWidget) paintThumb(canvas widget.Canvas, tok *theme.Tokens, primary, thumbFill widget.Color, center geometry.Point, id thumbID) {
	thumbR := metrics.Slider.ThumbSize / 2
	thumbRect := geometry.NewRect(center.X-thumbR, center.Y-thumbR, metrics.Slider.ThumbSize, metrics.Slider.ThumbSize)

	// shadow-sm under the thumb (faded with the rest when disabled).
	shadow := metrics.ShadowSM
	if s.disabled {
		faded := make([]metrics.ShadowLayer, len(shadow))
		for i, l := range shadow {
			l.Alpha *= metrics.DisabledOpacity
			faded[i] = l
		}
		shadow = faded
	}
	draw.Shadow(canvas, thumbRect, thumbR, shadow)

	// hover:ring-4 focus-visible:ring-4 ring-ring/50 — band hugging the active
	// or hovered thumb; never on a disabled slider. A dragged thumb keeps the
	// ring (focus stays on it while dragging, like the browser).
	ringed := s.hovered == id || ((s.dragging || s.IsFocused()) && s.active == id)
	if !s.disabled && ringed {
		ring := draw.Alpha(tok.Ring, metrics.RingAlpha)
		canvas.StrokeCircle(center, thumbR+metrics.SliderRingWidth/2, ring, metrics.SliderRingWidth)
	}

	canvas.DrawCircle(center, thumbR, thumbFill)
	canvas.StrokeCircle(center, thumbR-metrics.Slider.ThumbBorderWidth/2, primary, metrics.Slider.ThumbBorderWidth)
}

// thumbX returns the LOCAL x of a thumb center for value v, inset by the thumb
// radius on each side so the thumb stays inside the track at the extremes
// (Radix). The same inset is used by valueFromX, so the visual thumb position
// and the value its X maps to are exactly consistent.
func (s *RangeSliderWidget) thumbX(width, v float32) float32 {
	thumbR := metrics.Slider.ThumbSize / 2
	travel := width - metrics.Slider.ThumbSize
	if travel <= 0 {
		return thumbR
	}
	rng := s.max - s.min
	if rng <= 0 {
		return thumbR
	}
	progress := (v - s.min) / rng
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	return thumbR + progress*travel
}

// Event handles hover, press (thumb selection + capture), captured drag, and
// keyboard adjustment.
//
// Coordinate handling: positions arriving on the non-captured path are in this
// widget's PARENT space (the widget sits at Bounds().Min). The captured path
// (post-press drag/release) bypasses the parent and delivers RAW SCREEN
// coords. screenPos converts the parent-local case to screen space; the
// captured case is already screen space (its position equals its global
// position) so screenPos must NOT be applied there. valueFromX then maps
// against the SCREEN-space track (origin = ScreenOrigin()) so both paths agree.
func (s *RangeSliderWidget) Event(ctx widget.Context, e event.Event) bool {
	if !s.IsVisible() || !s.IsEnabled() || s.disabled {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return s.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		return s.keyEvent(ctx, ev)
	}
	return false
}

// screenPos converts a parent-local event position to screen space, shifting
// it by (ScreenOrigin - Bounds.Min). It is applied ONLY to the non-captured
// path; captured drag/release positions are already raw screen coords.
func (s *RangeSliderWidget) screenPos(p geometry.Point) geometry.Point {
	return p.Sub(s.Bounds().Min).Add(s.ScreenOrigin())
}

func (s *RangeSliderWidget) mouseEvent(ctx widget.Context, ev *event.MouseEvent) bool {
	switch ev.MouseType {
	case event.MouseEnter, event.MouseMove:
		if s.dragging {
			// Captured drag: position is raw screen coords (no screenPos). Move
			// the active thumb to track the cursor.
			s.dragTo(ctx, ev.Position)
			return true
		}
		// Hover: highlight the thumb under the cursor (parent-local coords).
		s.updateHover(ctx, s.screenPos(ev.Position))
		return ev.MouseType == event.MouseEnter
	case event.MouseLeave:
		if !s.dragging && s.hovered != thumbNone {
			s.hovered = thumbNone
			ctx.SetCursor(widget.CursorDefault)
			s.SetNeedsRedraw(true)
			ctx.InvalidateRect(s.Bounds())
		}
		return true
	case event.MousePress:
		if ev.Button != event.ButtonLeft {
			return false
		}
		return s.press(ctx, s.screenPos(ev.Position))
	case event.MouseRelease:
		if ev.Button != event.ButtonLeft {
			return false
		}
		return s.release(ctx, ev.Position)
	}
	return false
}

// press selects the thumb to drag (the nearer one to the click in SCREEN
// space), captures the pointer, takes focus, and snaps that thumb to the
// click position so a click on the track jumps the nearer thumb.
func (s *RangeSliderWidget) press(ctx widget.Context, screen geometry.Point) bool {
	ctx.RequestFocus(s)
	s.active = s.nearestThumb(screen.X)
	s.dragging = true
	s.hovered = s.active
	ctx.SetCursor(widget.CursorPointer)
	if pc, ok := ctx.(widget.PointerCapturer); ok {
		pc.CapturePointer(s)
	}
	s.setActiveValue(ctx, s.valueFromX(screen.X))
	s.SetNeedsRedraw(true)
	ctx.InvalidateRect(s.Bounds())
	return true
}

// dragTo moves the active thumb to the screen X of a captured drag move.
func (s *RangeSliderWidget) dragTo(ctx widget.Context, screen geometry.Point) {
	s.setActiveValue(ctx, s.valueFromX(screen.X))
	ctx.SetCursor(widget.CursorPointer)
}

// release ends the drag and releases the pointer capture. pos is whatever the
// captured release delivered (screen coords); we use it only to decide the
// resting hover state.
func (s *RangeSliderWidget) release(ctx widget.Context, pos geometry.Point) bool {
	wasDragging := s.dragging
	s.dragging = false
	if pc, ok := ctx.(widget.PointerCapturer); ok {
		pc.ReleasePointer(s)
	}
	if s.ScreenBounds().Contains(pos) || s.Bounds().Contains(pos) {
		s.hovered = s.active
	} else {
		s.hovered = thumbNone
		ctx.SetCursor(widget.CursorDefault)
	}
	s.SetNeedsRedraw(true)
	ctx.InvalidateRect(s.Bounds())
	return wasDragging
}

// updateHover sets the hovered thumb from a screen-space cursor position.
func (s *RangeSliderWidget) updateHover(ctx widget.Context, screen geometry.Point) {
	var want thumbID
	if s.thumbContains(screen, thumbLo) || s.thumbContains(screen, thumbHi) {
		want = s.nearestThumb(screen.X)
	}
	if want != s.hovered {
		s.hovered = want
		if want == thumbNone {
			ctx.SetCursor(widget.CursorDefault)
		} else {
			ctx.SetCursor(widget.CursorPointer)
		}
		s.SetNeedsRedraw(true)
		ctx.InvalidateRect(s.Bounds())
	}
}

// thumbContains reports whether a screen point is within a thumb's circle
// (plus the ring band, so the hover target matches the visible affordance).
func (s *RangeSliderWidget) thumbContains(screen geometry.Point, id thumbID) bool {
	v := s.Lo()
	if id == thumbHi {
		v = s.Hi()
	}
	cx := s.ScreenOrigin().X + s.thumbX(s.Bounds().Width(), v)
	cy := s.ScreenOrigin().Y + s.Bounds().Height()/2
	dx := screen.X - cx
	dy := screen.Y - cy
	r := metrics.Slider.ThumbSize/2 + metrics.SliderRingWidth
	return dx*dx+dy*dy <= r*r
}

// nearestThumb returns the thumb whose center is closer to screen X. On a tie
// it favors the low thumb for clicks left of center and the high thumb for
// clicks to the right, so clicking the empty track moves the side you clicked.
func (s *RangeSliderWidget) nearestThumb(screenX float32) thumbID {
	loCx := s.ScreenOrigin().X + s.thumbX(s.Bounds().Width(), s.Lo())
	hiCx := s.ScreenOrigin().X + s.thumbX(s.Bounds().Width(), s.Hi())
	dLo := absf(screenX - loCx)
	dHi := absf(screenX - hiCx)
	if dLo < dHi {
		return thumbLo
	}
	if dHi < dLo {
		return thumbHi
	}
	// Equidistant (thumbs coincident, e.g. lo==hi): pick by side of click.
	if screenX < loCx {
		return thumbLo
	}
	return thumbHi
}

// valueFromX maps a SCREEN-space X to a value, against the track in screen
// space (origin = ScreenOrigin()) inset by the thumb radius on each side. This
// is the inverse of thumbX, so dragging keeps the thumb under the cursor.
func (s *RangeSliderWidget) valueFromX(screenX float32) float32 {
	thumbR := metrics.Slider.ThumbSize / 2
	width := s.Bounds().Width()
	travel := width - metrics.Slider.ThumbSize
	if travel <= 0 {
		return s.min
	}
	trackLeft := s.ScreenOrigin().X + thumbR
	progress := (screenX - trackLeft) / travel
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	return s.min + progress*(s.max-s.min)
}

// setActiveValue writes raw into the active thumb, snapping to step, clamping
// to bounds and to the other thumb (lo<=hi), then publishes via signals and
// OnChange and requests a repaint. No-ops when the value is unchanged.
func (s *RangeSliderWidget) setActiveValue(ctx widget.Context, raw float32) {
	v := clampSnap(raw, s.min, s.max, s.step)
	changed := false
	switch s.active {
	case thumbLo:
		if v > s.Hi() {
			v = s.Hi()
		}
		if v != s.Lo() {
			s.lo = v
			if s.loSig != nil {
				s.loSig.Set(v)
			}
			changed = true
		}
	case thumbHi:
		if v < s.Lo() {
			v = s.Lo()
		}
		if v != s.Hi() {
			s.hi = v
			if s.hiSig != nil {
				s.hiSig.Set(v)
			}
			changed = true
		}
	}
	if !changed {
		return
	}
	if s.onChange != nil {
		s.onChange(s.Lo(), s.Hi())
	}
	s.SetNeedsRedraw(true)
	ctx.InvalidateRect(s.Bounds())
}

// keyEvent adjusts the active thumb with the arrow keys / Home / End while the
// slider is focused. Left/Down decrement, Right/Up increment by the step (or a
// 1% default), Home/End jump to the bound (clamped to the other thumb).
func (s *RangeSliderWidget) keyEvent(ctx widget.Context, ev *event.KeyEvent) bool {
	if !s.IsFocused() {
		return false
	}
	if ev.KeyType != event.KeyPress && ev.KeyType != event.KeyRepeat {
		return false
	}
	if s.active == thumbNone {
		s.active = thumbLo
	}
	rng := s.max - s.min
	if rng <= 0 {
		return false
	}
	step := s.step
	if step <= 0 {
		step = rng * 0.01
		if step > 1 {
			step = 1
		}
	}
	current := s.Lo()
	if s.active == thumbHi {
		current = s.Hi()
	}
	switch ev.Key {
	case event.KeyRight, event.KeyUp:
		s.setActiveValue(ctx, current+step)
	case event.KeyLeft, event.KeyDown:
		s.setActiveValue(ctx, current-step)
	case event.KeyHome:
		s.setActiveValue(ctx, s.min)
	case event.KeyEnd:
		s.setActiveValue(ctx, s.max)
	default:
		return false
	}
	return true
}

// normalize enforces min<=lo<=hi<=max on the stored values, snapping to step.
func (s *RangeSliderWidget) normalize() {
	s.lo = clampSnap(s.lo, s.min, s.max, s.step)
	s.hi = clampSnap(s.hi, s.min, s.max, s.step)
	if s.lo > s.hi {
		s.lo, s.hi = s.hi, s.lo
	}
}

// Children returns nil; the range slider is a leaf.
func (s *RangeSliderWidget) Children() []widget.Widget { return nil }

// Mount registers the controlled-signal bindings so external writes to loSig /
// hiSig repaint the thumbs.
func (s *RangeSliderWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if s.loSig != nil {
		s.AddBinding(state.BindToScheduler(s.loSig, s, sched))
	}
	if s.hiSig != nil {
		s.AddBinding(state.BindToScheduler(s.hiSig, s, sched))
	}
}

// Unmount has no resources to free beyond the bindings WidgetBase cleans up
// automatically; it exists to satisfy widget.Lifecycle alongside Mount.
func (s *RangeSliderWidget) Unmount() {}

// SetFocused tracks focus with MarkRedrawLocal (NOT SetNeedsRedraw): SetFocused
// runs inside RequestFocus, which already holds the context lock, so calling
// SetNeedsRedraw there would re-enter that lock and deadlock (the button.go
// pattern). MarkRedrawLocal only flips the local dirty flag.
func (s *RangeSliderWidget) SetFocused(focused bool) {
	s.WidgetBase.SetFocused(focused)
	s.MarkRedrawLocal()
}

// clampSnap clamps v to [min, max] and snaps to the nearest step (relative to
// min) when step > 0, mirroring core/slider's clampAndSnap.
func clampSnap(v, min, max, step float32) float32 {
	if v < min {
		v = min
	}
	if v > max {
		v = max
	}
	if step > 0 {
		steps := float32(math.Round(float64((v - min) / step)))
		v = min + steps*step
		if v < min {
			v = min
		}
		if v > max {
			v = max
		}
	}
	return v
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*RangeSliderWidget)(nil)
	_ widget.Focusable = (*RangeSliderWidget)(nil)
	_ widget.Lifecycle = (*RangeSliderWidget)(nil)
)

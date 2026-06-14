package graft_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// layoutRangeAt lays out a range slider of the given width and places it at the
// supplied screen offset, stamping the screen origin the way a parent's Draw
// pass would and running Draw so the widget pins its screen-space track. The
// returned values match the geometry the captured drag path computes against.
func layoutRangeAt(s *graft.RangeSliderWidget, width, offsetX float32) {
	uitest.LayoutWidget(s, 400, metrics.Slider.Height)
	s.SetBounds(geometry.FromPointSize(geometry.Pt(offsetX, 0),
		geometry.Sz(width, metrics.Slider.Height)))
	s.SetScreenOrigin(geometry.Pt(offsetX, 0))
	uitest.DrawWidget(s)
}

// rangeTrack returns the inset track geometry in screen space for a width/offset
// (inset by the thumb radius on each side, matching valueFromX / thumbX).
func rangeTrack(width, offsetX float32) (left, w float32) {
	r := metrics.Slider.ThumbSize / 2
	left = offsetX + r
	w = width - metrics.Slider.ThumbSize
	return left, w
}

// dragMove builds a captured-drag MouseMove with the left button held and raw
// screen coordinates, exactly as gogpu's Window delivers them post-capture.
func dragMove(x, y float32) *event.MouseEvent {
	return event.NewMouseEvent(event.MouseMove, event.ButtonNone, event.ButtonStateLeft,
		geometry.Pt(x, y), geometry.Pt(x, y), event.ModNone)
}

// releaseAt builds a left-button MouseRelease at raw screen coordinates.
func releaseAt(x, y float32) *event.MouseEvent {
	return event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, event.ButtonState(0),
		geometry.Pt(x, y), geometry.Pt(x, y), event.ModNone)
}

func TestRangeSliderInitialClamp(t *testing.T) {
	lightTokens(t)

	// Out-of-order Values are swapped so lo<=hi.
	s := graft.RangeSlider().Values(70, 20)
	if s.Lo() != 20 || s.Hi() != 70 {
		t.Fatalf("Values(70,20) normalized to (%v,%v), want (20,70)", s.Lo(), s.Hi())
	}

	// Both ends are clamped into [min,max].
	s2 := graft.RangeSlider().Min(0).Max(100).Values(-30, 250)
	if s2.Lo() != 0 || s2.Hi() != 100 {
		t.Fatalf("Values(-30,250) clamped to (%v,%v), want (0,100)", s2.Lo(), s2.Hi())
	}

	// Default thumbs sit at the bounds.
	d := graft.RangeSlider()
	if d.Lo() != 0 || d.Hi() != 100 {
		t.Fatalf("default range = (%v,%v), want (0,100)", d.Lo(), d.Hi())
	}
}

// TestRangeSliderDragLoThumbAtNonZeroOffset drives the LOW thumb via the
// captured-drag path at a non-zero screen offset. This is the regression class
// that bit the single slider: with the bounds/track kept in screen space and
// valueFromX mapping against ScreenOrigin, a raw screen X computes the right
// value instead of saturating.
func TestRangeSliderDragLoThumbAtNonZeroOffset(t *testing.T) {
	lightTokens(t)

	const offsetX float32 = 200
	const width float32 = 200

	loSig := state.NewSignal(float32(10))
	hiSig := state.NewSignal(float32(90))
	var obsLo, obsHi float32
	s := graft.RangeSlider().BindRange(loSig, hiSig).
		OnChange(func(lo, hi float32) { obsLo, obsHi = lo, hi }).W(width)
	layoutRangeAt(s, width, offsetX)

	trackLeft, trackW := rangeTrack(width, offsetX)

	// Press near the LOW thumb (10% across) so the low thumb becomes active.
	pressX := trackLeft + 0.10*trackW
	ctx := uitest.NewMockContext()
	if !s.Event(ctx, uitest.Click(pressX, 6)) {
		t.Fatal("press not consumed")
	}
	if got := loSig.Get(); !approx(got, 10) {
		t.Fatalf("press at 10%% set lo = %v, want ~10", got)
	}

	// Captured drag: the Window delivers MouseMove with RAW SCREEN coords and
	// the left button held. Drive the widget's Event directly (dragging is now
	// true, so it treats the position as screen space).
	const wantProgress float32 = 0.40
	dragX := trackLeft + wantProgress*trackW
	if !s.Event(ctx, dragMove(dragX, 6)) {
		t.Fatal("captured drag move not consumed")
	}

	wantLo := wantProgress * 100 // 40
	if got := loSig.Get(); !approx(got, wantLo) {
		t.Fatalf("drag lo at non-zero offset: lo = %v, want ~%v (pre-fix this "+
			"saturates because raw screen X is compared against a (0,0) track)", got, wantLo)
	}
	if got := hiSig.Get(); !approx(got, 90) {
		t.Fatalf("hi thumb moved while dragging lo: hi = %v, want 90", got)
	}
	if !approx(obsLo, wantLo) || !approx(obsHi, 90) {
		t.Fatalf("OnChange observed (%v,%v), want (~%v,90)", obsLo, obsHi, wantLo)
	}

	// Release ends the drag.
	s.Event(ctx, releaseAt(dragX, 6))
}

// TestRangeSliderDragHiThumbAtNonZeroOffset drives the HIGH thumb via the
// captured path at a non-zero offset, mirroring the low-thumb test.
func TestRangeSliderDragHiThumbAtNonZeroOffset(t *testing.T) {
	lightTokens(t)

	const offsetX float32 = 320
	const width float32 = 240

	loSig := state.NewSignal(float32(10))
	hiSig := state.NewSignal(float32(90))
	s := graft.RangeSlider().BindRange(loSig, hiSig).W(width)
	layoutRangeAt(s, width, offsetX)

	trackLeft, trackW := rangeTrack(width, offsetX)

	// Press near the HIGH thumb (90% across) so the high thumb becomes active.
	pressX := trackLeft + 0.90*trackW
	ctx := uitest.NewMockContext()
	if !s.Event(ctx, uitest.Click(pressX, 6)) {
		t.Fatal("press not consumed")
	}
	if got := hiSig.Get(); !approx(got, 90) {
		t.Fatalf("press at 90%% set hi = %v, want ~90", got)
	}

	// Captured drag toward the middle.
	const wantProgress float32 = 0.60
	dragX := trackLeft + wantProgress*trackW
	if !s.Event(ctx, dragMove(dragX, 6)) {
		t.Fatal("captured drag move not consumed")
	}

	wantHi := wantProgress * 100 // 60
	if got := hiSig.Get(); !approx(got, wantHi) {
		t.Fatalf("drag hi at non-zero offset: hi = %v, want ~%v", got, wantHi)
	}
	if got := loSig.Get(); !approx(got, 10) {
		t.Fatalf("lo thumb moved while dragging hi: lo = %v, want 10", got)
	}
}

// TestRangeSliderThumbCrossClamp verifies a thumb cannot cross past the other:
// dragging the low thumb beyond the high thumb pins it at the high value.
func TestRangeSliderThumbCrossClamp(t *testing.T) {
	lightTokens(t)

	const offsetX float32 = 50
	const width float32 = 200

	loSig := state.NewSignal(float32(20))
	hiSig := state.NewSignal(float32(60))
	s := graft.RangeSlider().BindRange(loSig, hiSig).W(width)
	layoutRangeAt(s, width, offsetX)

	trackLeft, trackW := rangeTrack(width, offsetX)

	// Press the low thumb (20%).
	ctx := uitest.NewMockContext()
	if !s.Event(ctx, uitest.Click(trackLeft+0.20*trackW, 6)) {
		t.Fatal("press not consumed")
	}

	// Drag the low thumb to 95%, well past the high thumb at 60.
	dragX := trackLeft + 0.95*trackW
	s.Event(ctx, dragMove(dragX, 6))

	if got := loSig.Get(); !approx(got, 60) {
		t.Fatalf("lo dragged past hi: lo = %v, want clamped to 60", got)
	}
	if got := hiSig.Get(); !approx(got, 60) {
		t.Fatalf("hi must be unchanged at 60, got %v", got)
	}
	if s.Lo() > s.Hi() {
		t.Fatalf("invariant lo<=hi violated: (%v,%v)", s.Lo(), s.Hi())
	}
}

func TestRangeSliderStepSnap(t *testing.T) {
	lightTokens(t)

	const offsetX float32 = 0
	const width float32 = 200

	loSig := state.NewSignal(float32(0))
	hiSig := state.NewSignal(float32(100))
	s := graft.RangeSlider().Min(0).Max(100).Step(10).
		BindRange(loSig, hiSig).W(width)
	layoutRangeAt(s, width, offsetX)

	trackLeft, trackW := rangeTrack(width, offsetX)

	// Press + drag the low thumb to 23%, which must snap to the nearest 10 (20).
	ctx := uitest.NewMockContext()
	s.Event(ctx, uitest.Click(trackLeft, 6))
	dragX := trackLeft + 0.23*trackW
	s.Event(ctx, dragMove(dragX, 6))

	if got := loSig.Get(); !approx(got, 20) {
		t.Fatalf("step snap: lo = %v, want 20", got)
	}
}

func TestRangeSliderKeyboard(t *testing.T) {
	lightTokens(t)

	loSig := state.NewSignal(float32(20))
	hiSig := state.NewSignal(float32(80))
	s := graft.RangeSlider().Min(0).Max(100).Step(5).BindRange(loSig, hiSig).W(200)
	layoutRangeAt(s, 200, 0)

	ctx := uitest.NewMockContext()
	// Press the low thumb so it is the active thumb, then release the drag.
	trackLeft, trackW := rangeTrack(200, 0)
	s.Event(ctx, uitest.Click(trackLeft+0.20*trackW, 6))
	s.Event(ctx, releaseAt(trackLeft, 6))
	ctx.RequestFocus(s)

	// Right arrow increments the active (low) thumb by the step.
	if !s.Event(ctx, uitest.KeyPress(event.KeyRight, event.ModNone)) {
		t.Fatal("KeyRight not consumed")
	}
	if got := loSig.Get(); !approx(got, 25) {
		t.Fatalf("KeyRight: lo = %v, want 25", got)
	}
	// Left arrow decrements.
	s.Event(ctx, uitest.KeyPress(event.KeyLeft, event.ModNone))
	if got := loSig.Get(); !approx(got, 20) {
		t.Fatalf("KeyLeft: lo = %v, want 20", got)
	}
}

func TestGoldenRangeSlider(t *testing.T) {
	gtest.GoldenLightDark(t, "range-slider", func() widget.Widget {
		return primitives.Box(graft.RangeSlider().Values(25, 75).W(300)).Padding(16)
	})

	gtest.GoldenLightDark(t, "range-slider-values", func() widget.Widget {
		return primitives.VBox(
			graft.RangeSlider().Values(0, 100).W(300),
			graft.RangeSlider().Values(20, 60).W(300),
			graft.RangeSlider().Values(40, 50).W(300),
		).Gap(24).Padding(16)
	})

	gtest.GoldenLightDark(t, "range-slider-disabled", func() widget.Widget {
		return primitives.Box(graft.RangeSlider().Values(25, 75).W(300).Disabled(true)).Padding(16)
	})
}

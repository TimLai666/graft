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

func approx(a, b float32) bool {
	d := a - b
	return d < 0.01 && d > -0.01
}

func TestSliderSpecDefault(t *testing.T) {
	tok := lightTokens(t)

	s := graft.Slider().Value(50).W(200)
	size := uitest.LayoutWidget(s, 400, 100)
	if size.Width != 200 || size.Height != metrics.Slider.Height {
		t.Fatalf("layout: got %vx%v, want 200x%v", size.Width, size.Height, metrics.Slider.Height)
	}

	mc := uitest.DrawWidget(s)

	// Track: full-width 4px pill in muted, vertically centered (y=4..8).
	trackTop := metrics.Slider.Height/2 - metrics.Slider.TrackThickness/2
	foundTrack := false
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Muted &&
			approx(rr.Bounds.Min.X, 0) && approx(rr.Bounds.Width(), 200) &&
			approx(rr.Bounds.Height(), metrics.Slider.TrackThickness) &&
			approx(rr.Bounds.Min.Y, trackTop) && approx(rr.Radius, metrics.Slider.TrackRadius) {
			foundTrack = true
		}
	}
	if !foundTrack {
		t.Fatalf("track pill not drawn; roundrects: %+v", mc.RoundRects)
	}

	// Range: primary rect from 0 to the thumb center (6 + 0.5*188 = 100).
	foundRange := false
	for _, r := range mc.Rects {
		if r.Color == tok.Primary && approx(r.Bounds.Min.X, 0) && approx(r.Bounds.Width(), 100) &&
			approx(r.Bounds.Height(), metrics.Slider.TrackThickness) {
			foundRange = true
		}
	}
	if !foundRange {
		t.Fatalf("primary range not drawn; rects: %+v", mc.Rects)
	}
	if len(mc.ClipRoundRects) == 0 {
		t.Fatal("range must be clipped to the track pill (overflow-hidden)")
	}

	// Thumb: 6px-radius white circle at (100, 6) with a 1px primary border.
	thumbR := metrics.Slider.ThumbSize / 2
	cy := metrics.Slider.Height / 2
	foundThumb := false
	for _, c := range mc.Circles {
		if c.Color == widget.ColorWhite && approx(c.Radius, thumbR) &&
			approx(c.Center.X, 100) && approx(c.Center.Y, cy) {
			foundThumb = true
		}
	}
	if !foundThumb {
		t.Fatalf("white thumb not drawn; circles: %+v", mc.Circles)
	}
	foundBorder := false
	for _, c := range mc.StrokeCircles {
		if c.Color == tok.Primary && approx(c.Radius, thumbR-metrics.Slider.ThumbBorderWidth/2) && approx(c.StrokeWidth, 1) {
			foundBorder = true
		}
		if approx(c.StrokeWidth, metrics.SliderRingWidth) {
			t.Fatal("no hover/focus ring expected in the default state")
		}
	}
	if !foundBorder {
		t.Fatalf("1px primary thumb border not drawn; strokes: %+v", mc.StrokeCircles)
	}
}

func TestSliderSpecHoverRing(t *testing.T) {
	tok := lightTokens(t)

	s := graft.Slider().Value(50).W(200)
	uitest.LayoutWidget(s, 400, 100)
	ctx := uitest.NewMockContext()
	if !s.Event(ctx, uitest.MouseEnter(100, 8)) {
		t.Fatal("MouseEnter not consumed")
	}

	mc := uitest.DrawWidget(s)
	want := tok.Ring
	want.A = metrics.RingAlpha
	found := false
	for _, c := range mc.StrokeCircles {
		if c.Color == want && approx(c.StrokeWidth, metrics.SliderRingWidth) &&
			approx(c.Radius, metrics.Slider.ThumbSize/2+metrics.SliderRingWidth/2) {
			found = true
		}
	}
	if !found {
		t.Fatalf("hover ring (4px ring/50) not drawn; strokes: %+v", mc.StrokeCircles)
	}
}

func TestSliderSpecDisabled(t *testing.T) {
	tok := lightTokens(t)

	s := graft.Slider().Value(50).W(200).Disabled(true)
	uitest.LayoutWidget(s, 400, 100)
	// Hover must NOT produce a ring when disabled (and is not consumed).
	s.Event(uitest.NewMockContext(), uitest.MouseEnter(100, 8))

	mc := uitest.DrawWidget(s)
	fadedMuted := tok.Muted
	fadedMuted.A *= metrics.DisabledOpacity
	foundTrack := false
	for _, rr := range mc.RoundRects {
		if rr.Color == fadedMuted && approx(rr.Bounds.Height(), metrics.Slider.TrackThickness) {
			foundTrack = true
		}
	}
	if !foundTrack {
		t.Fatalf("disabled track must be muted at 50%%; roundrects: %+v", mc.RoundRects)
	}
	for _, c := range mc.StrokeCircles {
		if approx(c.StrokeWidth, metrics.SliderRingWidth) {
			t.Fatal("disabled slider must not draw a ring")
		}
	}
	fadedWhite := widget.ColorWhite
	fadedWhite.A *= metrics.DisabledOpacity
	foundThumb := false
	for _, c := range mc.Circles {
		if c.Color == fadedWhite && approx(c.Radius, metrics.Slider.ThumbSize/2) {
			foundThumb = true
		}
	}
	if !foundThumb {
		t.Fatalf("disabled thumb must be white at 50%%; circles: %+v", mc.Circles)
	}
}

func TestSliderInteraction(t *testing.T) {
	lightTokens(t)

	sig := state.NewSignal(float32(20))
	var observed float32
	s := graft.Slider().Bind(sig).OnChange(func(v float32) { observed = v })
	uitest.LayoutWidget(s, 200, 16)

	// Click at the exact center of the 200px slider: the core maps the
	// inner track (inset by its 10px layout constant) to the range, so
	// x=100 is exactly 50%.
	uitest.SimulateClick(s, 100, 8)
	if got := sig.Get(); got != 50 {
		t.Fatalf("click at center: signal = %v, want 50", got)
	}
	if observed != 50 {
		t.Fatalf("OnChange observed %v, want 50", observed)
	}
}

// TestSliderDragAtNonZeroOffset is a regression test for the drag bug where a
// slider placed anywhere other than the window origin would jump to the end on
// drag instead of tracking the cursor.
//
// Root cause (coordinate terms): on MousePress the wrapped core/slider captures
// the pointer, after which gogpu's Window delivers MouseMove/Release DIRECTLY to
// the core with RAW SCREEN coordinates, bypassing the wrapper's translation
// (app/window.go capturedWidget.Event). The core derives the value from that
// position against core.Bounds() (core/slider valueFromPosition). The wrapper
// used to pin those bounds at (0,0), so a raw screen X (e.g. 345 for a slider at
// x=200) was compared against a (0,0)-origin track and saturated to 100%.
//
// The fix keeps the core's bounds in SCREEN space between frames so the captured
// path computes correctly, while Draw temporarily swaps to (0,0)-origin bounds so
// the render is byte-identical. This test simulates the capture path by driving
// the core's Event directly with raw screen coords, exactly as the Window does.
func TestSliderDragAtNonZeroOffset(t *testing.T) {
	lightTokens(t)

	const offsetX float32 = 200
	const width float32 = 200

	sig := state.NewSignal(float32(0))
	s := graft.Slider().Bind(sig).W(width) // range [0, 100], fixed 200px width

	// Lay out the slider (sizes the core to width) then place it at a non-zero
	// screen position, as a real layout inside a positioned parent would.
	uitest.LayoutWidget(s, 400, metrics.Slider.Height)
	s.SetBounds(geometry.FromPointSize(geometry.Pt(offsetX, 0),
		geometry.Sz(width, metrics.Slider.Height)))
	// Stamp the screen origin the way a parent's Draw pass does, then run Draw
	// so the wrapper pins the core's bounds to that screen origin.
	s.SetScreenOrigin(geometry.Pt(offsetX, 0))
	uitest.DrawWidget(s)

	// Reach the wrapped core/slider — the widget the Window captures on press.
	core := s.Children()[0]

	// Track geometry in SCREEN space: the core insets the track by the 10px
	// thumb radius on each side (core/slider valueFromPosition / painter).
	const thumbRadius float32 = 10
	trackLeft := offsetX + thumbRadius          // 210
	trackRight := offsetX + width - thumbRadius // 390
	trackW := trackRight - trackLeft            // 180

	// Press near the LEFT of the track via the wrapper (this is the real entry
	// point for a press — it gets translated into the core's space). The wrapper
	// receives PARENT-LOCAL coords; we model a parent at the window origin, so
	// parent-local == screen coords here, and the slider sits at offsetX inside
	// that parent (s.Bounds().Min == s.ScreenOrigin() == offsetX).
	pressScreenX := trackLeft + 0.05*trackW // 5% across
	ctx := uitest.NewMockContext()
	if !s.Event(ctx, uitest.Click(pressScreenX, 8)) {
		t.Fatal("press not consumed by slider")
	}
	if got := sig.Get(); got > 20 {
		t.Fatalf("press near track-left set value too high: got %v, want <=20", got)
	}

	// Captured drag: the Window now delivers MouseMove straight to the CORE with
	// RAW SCREEN coordinates. Drive that path directly.
	const wantProgress float32 = 0.75
	dragScreenX := trackLeft + wantProgress*trackW // 210 + 135 = 345
	// The Window dispatches captured drags as MouseMove (app/window.go), so the
	// raw event must carry the left button held for the core to treat it as a
	// drag. uitest.MouseMove has no buttons, so build the event explicitly.
	drag := event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, event.ButtonStateLeft,
		geometry.Pt(dragScreenX, 8), geometry.Pt(dragScreenX, 8), event.ModNone,
	)
	if !core.Event(ctx, drag) {
		t.Fatal("captured drag move not consumed by core")
	}

	want := wantProgress * 100 // 75
	if got := sig.Get(); !approx(got, want) {
		t.Fatalf("drag at non-zero offset: value = %v, want ~%v "+
			"(pre-fix this saturates to 100 because raw screen X is compared "+
			"against a (0,0)-origin track)", got, want)
	}
}

func TestGoldenSlider(t *testing.T) {
	gtest.GoldenLightDark(t, "slider-value-50", func() widget.Widget {
		return primitives.Box(graft.Slider().Value(50).W(300)).Padding(16)
	})

	gtest.GoldenLightDark(t, "slider-hover-ring", func() widget.Widget {
		s := graft.Slider().Value(50).W(300)
		uitest.LayoutWidget(s, 300, metrics.Slider.Height)
		s.Event(uitest.NewMockContext(), uitest.MouseEnter(150, 8))
		return primitives.Box(s).Padding(16)
	})

	gtest.GoldenLightDark(t, "slider-disabled", func() widget.Widget {
		return primitives.Box(graft.Slider().Value(50).W(300).Disabled(true)).Padding(16)
	})

	gtest.GoldenLightDark(t, "slider-values", func() widget.Widget {
		return primitives.VBox(
			graft.Slider().Value(0).W(300),
			graft.Slider().Value(25).W(300),
			graft.Slider().Value(100).W(300),
		).Gap(24).Padding(16)
	})
}

// silence unused import when geometry assertions change during edits.
var _ = geometry.Pt

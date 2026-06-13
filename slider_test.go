package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// lightTokens forces light mode and returns the active token set plus a
// restore function.
func lightTokens(t *testing.T) *theme.Tokens {
	t.Helper()
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	t.Cleanup(func() { th.SetMode(prev) })
	return th.Active()
}

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

	// Track: full-width 6px pill in muted, vertically centered (y=5..11).
	foundTrack := false
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Muted &&
			approx(rr.Bounds.Min.X, 0) && approx(rr.Bounds.Width(), 200) &&
			approx(rr.Bounds.Height(), metrics.Slider.TrackThickness) &&
			approx(rr.Bounds.Min.Y, 5) && approx(rr.Radius, metrics.Slider.TrackRadius) {
			foundTrack = true
		}
	}
	if !foundTrack {
		t.Fatalf("track pill not drawn; roundrects: %+v", mc.RoundRects)
	}

	// Range: primary rect from 0 to the thumb center (8 + 0.5*184 = 100).
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

	// Thumb: 8px-radius white circle at (100, 8) with a 1px primary border.
	foundThumb := false
	for _, c := range mc.Circles {
		if c.Color == widget.ColorWhite && approx(c.Radius, 8) &&
			approx(c.Center.X, 100) && approx(c.Center.Y, 8) {
			foundThumb = true
		}
	}
	if !foundThumb {
		t.Fatalf("white thumb not drawn; circles: %+v", mc.Circles)
	}
	foundBorder := false
	for _, c := range mc.StrokeCircles {
		if c.Color == tok.Primary && approx(c.Radius, 7.5) && approx(c.StrokeWidth, 1) {
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
			approx(c.Radius, 8+metrics.SliderRingWidth/2) {
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
		if c.Color == fadedWhite && approx(c.Radius, 8) {
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

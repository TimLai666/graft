package graft_test

import (
	"math"
	"testing"
	"time"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/theme"
)

// TestSpinnerSpec verifies the default spinner draws a 270° arc in the
// Foreground color at rotation phase 0 (leading cap at the top), with the
// lucide 1/12 stroke ratio and the arc inset by half the stroke.
func TestSpinnerSpec(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()

	s := graft.Spinner() // 16px default
	size := s.Layout(nil, fixedWidthLoose(16))
	if size.Width != 16 || size.Height != 16 {
		t.Errorf("spinner size = %vx%v, want 16x16", size.Width, size.Height)
	}

	// Draw with a zero-delta frame (the deterministic golden path) so the spin
	// stays at rotation phase 0 (leading cap at the top).
	canvas := uitest.DrawWidgetWithContext(s, zeroDeltaContext())
	// The mock canvas implements ArcStroker, so the round-capped arc lands in
	// StrokeArcStyleds.
	if len(canvas.StrokeArcStyleds) != 1 {
		t.Fatalf("spinner drew %d styled arcs, want 1", len(canvas.StrokeArcStyleds))
	}
	arc := canvas.StrokeArcStyleds[0]
	if arc.Cap != widget.LineCapRound {
		t.Errorf("spinner cap = %v, want LineCapRound", arc.Cap)
	}
	if arc.Color != tok.Foreground {
		t.Errorf("spinner color = %v, want Foreground %v", arc.Color, tok.Foreground)
	}
	if math.Abs(arc.StartAngle-(-math.Pi/2)) > 1e-6 {
		t.Errorf("spinner start angle = %v, want -pi/2 (top)", arc.StartAngle)
	}
	if math.Abs(arc.SweepAngle-math.Pi*1.5) > 1e-6 {
		t.Errorf("spinner sweep = %v, want 270deg (3pi/2)", arc.SweepAngle)
	}
	const wantStroke = 16.0 * 2.0 / 24.0 // 1.333..
	if math.Abs(float64(arc.StrokeWidth-wantStroke)) > 1e-4 {
		t.Errorf("spinner stroke = %v, want %v", arc.StrokeWidth, wantStroke)
	}
	wantRadius := float32(16)/2 - arc.StrokeWidth/2
	if math.Abs(float64(arc.Radius-wantRadius)) > 1e-4 {
		t.Errorf("spinner radius = %v, want %v (inset by half stroke)", arc.Radius, wantRadius)
	}
	if arc.Center.X != 8 || arc.Center.Y != 8 {
		t.Errorf("spinner center = %v, want (8,8)", arc.Center)
	}
}

// TestSpinnerColorToken verifies ColorToken overrides the arc color.
func TestSpinnerColorToken(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()

	s := graft.Spinner().ColorToken(func(t *theme.Tokens) widget.Color { return t.MutedForeground })
	s.Layout(nil, fixedWidthLoose(16))
	canvas := uitest.DrawWidget(s)
	if len(canvas.StrokeArcStyleds) != 1 {
		t.Fatalf("spinner drew %d styled arcs, want 1", len(canvas.StrokeArcStyleds))
	}
	if canvas.StrokeArcStyleds[0].Color != tok.MutedForeground {
		t.Errorf("spinner color = %v, want MutedForeground %v", canvas.StrokeArcStyleds[0].Color, tok.MutedForeground)
	}
}

// TestSpinnerSize verifies Size scales the spinner square.
func TestSpinnerSize(t *testing.T) {
	alertForceLight(t)
	s := graft.Spinner().Size(32)
	size := s.Layout(nil, fixedWidthLoose(32))
	if size.Width != 32 || size.Height != 32 {
		t.Errorf("spinner size = %vx%v, want 32x32", size.Width, size.Height)
	}
}

// TestSpinnerRotationAdvancesInDraw verifies the spin is driven by Draw, not
// Layout. Layout is NOT re-run every frame in a continuous-render app (only
// Draw is), so a rotation ticked only in Layout would freeze at phase 0. We
// Layout ONCE, then draw two successive frames with a non-zero frame delta
// WITHOUT re-laying-out, and assert the arc start angle changed. With the
// pre-fix code (tick in Layout) both frames render at -pi/2 and this fails.
func TestSpinnerRotationAdvancesInDraw(t *testing.T) {
	alertForceLight(t)

	s := graft.Spinner()
	s.Layout(nil, fixedWidthLoose(16)) // single layout pass, as in a real app

	ctx := uitest.NewMockContext()
	ctx.DeltaVal = 250 * time.Millisecond // quarter turn per frame

	first := uitest.DrawWidgetWithContext(s, ctx)
	second := uitest.DrawWidgetWithContext(s, ctx)
	if len(first.StrokeArcStyleds) != 1 || len(second.StrokeArcStyleds) != 1 {
		t.Fatalf("expected 1 styled arc per draw, got %d then %d",
			len(first.StrokeArcStyleds), len(second.StrokeArcStyleds))
	}
	a0 := first.StrokeArcStyleds[0].StartAngle
	a1 := second.StrokeArcStyleds[0].StartAngle
	if math.Abs(a0-a1) < 1e-6 {
		t.Errorf("spin start angle did not advance across draw frames (both %.6f); "+
			"the rotation must be ticked in Draw, not Layout", a0)
	}
}

// TestGoldenSpinner renders spinners at a few sizes at rotation phase 0 in
// light and dark modes (deterministic: a single offscreen frame has dt=0).
func TestGoldenSpinner(t *testing.T) {
	gtest.GoldenLightDark(t, "spinner-sizes", func() widget.Widget {
		return primitives.HBox(
			graft.Spinner().Size(16),
			graft.Spinner().Size(24),
			graft.Spinner().Size(32),
			graft.Spinner().Size(48),
		).Gap(20).Padding(24).CrossAlign(primitives.CrossAxisCenter)
	})
}

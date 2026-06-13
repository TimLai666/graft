package graft_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
)

// TestSkeletonSpec verifies the skeleton fills with the Accent token at
// rounded-md, sized to its explicit W/H, at full opacity (pulse phase 0).
func TestSkeletonSpec(t *testing.T) {
	th := alertForceLight(t) // also loads assets + pins light mode
	tok := th.Active()

	s := graft.Skeleton().Size(200, 16)
	size := s.Layout(nil, fixedWidthLoose(200))
	if size.Width != 200 || size.Height != 16 {
		t.Errorf("skeleton size = %vx%v, want 200x16", size.Width, size.Height)
	}

	// Draw with a zero-delta frame (the deterministic golden path) so the pulse
	// stays at phase 0 (full opacity), matching what the offscreen renderer sees.
	canvas := uitest.DrawWidgetWithContext(s, zeroDeltaContext())
	if len(canvas.RoundRects) != 1 {
		t.Fatalf("skeleton drew %d round-rects, want 1", len(canvas.RoundRects))
	}
	rr := canvas.RoundRects[0]
	if rr.Color != tok.Accent {
		t.Errorf("skeleton fill = %v, want Accent %v (full opacity at phase 0)", rr.Color, tok.Accent)
	}
	if rr.Radius != th.RadiusMD() {
		t.Errorf("skeleton radius = %v, want RadiusMD %v", rr.Radius, th.RadiusMD())
	}
}

// TestSkeletonCircleRadius verifies Circle uses the full pill radius.
func TestSkeletonCircleRadius(t *testing.T) {
	th := alertForceLight(t)
	s := graft.Skeleton().Size(40, 40).Circle()
	s.Layout(nil, fixedWidthLoose(40))
	canvas := uitest.DrawWidget(s)
	if len(canvas.RoundRects) != 1 {
		t.Fatalf("skeleton drew %d round-rects, want 1", len(canvas.RoundRects))
	}
	if got := canvas.RoundRects[0].Radius; got != th.RadiusFull() {
		t.Errorf("circle skeleton radius = %v, want RadiusFull %v", got, th.RadiusFull())
	}
}

// TestSkeletonPulseAdvancesInDraw verifies the pulse animation is driven by
// Draw, not Layout. Layout is NOT re-run every frame in a continuous-render
// app (only Draw is), so a pulse ticked only in Layout would freeze at full
// opacity. We Layout ONCE, then draw two successive frames with a non-zero
// frame delta WITHOUT re-laying-out, and assert the fill opacity changed
// between frames. With the pre-fix code (tick in Layout) both frames render at
// phase 0 and this fails.
func TestSkeletonPulseAdvancesInDraw(t *testing.T) {
	alertForceLight(t)

	s := graft.Skeleton().Size(200, 16)
	s.Layout(nil, fixedWidthLoose(200)) // single layout pass, as in a real app

	// Advance ~quarter of the pulse period per draw frame so the opacity
	// visibly moves toward the trough.
	ctx := uitest.NewMockContext()
	ctx.DeltaVal = 500 * time.Millisecond

	first := uitest.DrawWidgetWithContext(s, ctx)
	second := uitest.DrawWidgetWithContext(s, ctx)
	if len(first.RoundRects) != 1 || len(second.RoundRects) != 1 {
		t.Fatalf("expected 1 round-rect per draw, got %d then %d", len(first.RoundRects), len(second.RoundRects))
	}
	a0 := first.RoundRects[0].Color.A
	a1 := second.RoundRects[0].Color.A
	if a0 == a1 {
		t.Errorf("pulse opacity did not advance across draw frames (both %.4f); "+
			"the pulse must be ticked in Draw, not Layout", a0)
	}
}

// TestGoldenSkeleton renders a small composition of skeletons (avatar circle
// plus two text lines, the canonical shadcn card-loading layout) at pulse
// phase 0 in light and dark modes.
func TestGoldenSkeleton(t *testing.T) {
	gtest.GoldenLightDark(t, "skeleton-card", func() widget.Widget {
		return primitives.HBox(
			graft.Skeleton().Size(48, 48).Circle(),
			primitives.VBox(
				graft.Skeleton().Size(220, 16),
				graft.Skeleton().Size(180, 16),
			).Gap(8),
		).Gap(16).Padding(24)
	})
}

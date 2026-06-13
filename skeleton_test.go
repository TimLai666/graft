package graft_test

import (
	"testing"

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

	canvas := uitest.DrawWidget(s)
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

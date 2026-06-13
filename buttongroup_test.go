package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// buttonGroupForceLight pins the theme to light and returns active tokens plus
// a restore function.
func buttonGroupForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

// TestButtonGroupLayoutOverlap verifies adjacent buttons overlap by 1px so the
// total width is the sum minus the overlaps, and the height is uniform.
func TestButtonGroupLayoutOverlap(t *testing.T) {
	_, restore := buttonGroupForceLight(t)
	defer restore()

	b1 := graft.Button("One").Outline()
	b2 := graft.Button("Two").Outline()
	b3 := graft.Button("Three").Outline()
	g := graft.ButtonGroup(b1, b2, b3)
	size := g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(600, 100)))

	w1 := b1.Bounds().Width()
	w2 := b2.Bounds().Width()
	w3 := b3.Bounds().Width()
	overlap := metrics.ButtonGroup.Overlap
	wantW := w1 + w2 + w3 - 2*overlap
	if size.Width != wantW {
		t.Fatalf("group width: got %v want %v", size.Width, wantW)
	}
	if size.Height != metrics.Button.Default.Height {
		t.Fatalf("group height: got %v want %v", size.Height, metrics.Button.Default.Height)
	}

	// Buttons advance left-to-right with the overlap applied.
	if b2.Bounds().Min.X != w1-overlap {
		t.Fatalf("second button x: got %v want %v", b2.Bounds().Min.X, w1-overlap)
	}
}

// TestButtonGroupOuterBorder asserts the group draws a unified 1px outer border
// in the Border token at the group radius and clips to a rounded rect.
func TestButtonGroupOuterBorder(t *testing.T) {
	tok, restore := buttonGroupForceLight(t)
	defer restore()

	g := graft.ButtonGroup(
		graft.Button("One").Outline(),
		graft.Button("Two").Outline(),
		graft.Button("Three").Outline(),
	)
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(600, 100)))
	canvas := uitest.DrawWidget(g)

	radius := graft.CurrentTheme().RadiusMD()

	// The group's rounded clip is present.
	if len(canvas.ClipRoundRects) == 0 {
		t.Fatal("expected a rounded clip for the group")
	}
	if canvas.ClipRoundRects[0].Radius != radius {
		t.Fatalf("clip radius: got %v want %v", canvas.ClipRoundRects[0].Radius, radius)
	}

	// A unified outer border stroke in the Border token at radius-0.5.
	var outer *uitest.StrokeRoundRectCall
	for idx := range canvas.StrokeRoundRects {
		s := &canvas.StrokeRoundRects[idx]
		if s.StrokeWidth == metrics.ButtonGroup.OuterBorderWidth && s.Color == tok.Border {
			outer = s
		}
	}
	if outer == nil {
		t.Fatal("no unified outer Border stroke found")
	}
	if want := radius - metrics.ButtonGroup.OuterBorderWidth/2; outer.Radius != want {
		t.Fatalf("outer border radius: got %v want %v", outer.Radius, want)
	}
}

// TestButtonGroupClickFiresOnlyHitSegment verifies a click in one segment
// fires only that segment's OnClick and focuses only that segment. Before the
// hit-test fix, the group broadcast every mouse event to all segments, so a
// MousePress pressed/focused every button (focus landing on the last one).
func TestButtonGroupClickFiresOnlyHitSegment(t *testing.T) {
	_, restore := buttonGroupForceLight(t)
	defer restore()

	var hits [3]int
	b1 := graft.Button("One").OnClick(func() { hits[0]++ })
	b2 := graft.Button("Two").OnClick(func() { hits[1]++ })
	b3 := graft.Button("Three").OnClick(func() { hits[2]++ })
	g := graft.ButtonGroup(b1, b2, b3)
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(600, 100)))

	// Press inside the middle segment must focus only b2.
	mid := b2.Bounds()
	ctx := uitest.NewMockContext()
	g.Event(ctx, uitest.Click(mid.Min.X+mid.Width()/2, mid.Min.Y+mid.Height()/2))
	if ctx.FocusedWidget() != widget.Widget(b2) {
		t.Fatalf("press focused wrong segment: got %v want b2", ctx.FocusedWidget())
	}

	// A full click on the middle segment fires only b2's OnClick.
	uitest.SimulateClick(g, mid.Min.X+mid.Width()/2, mid.Min.Y+mid.Height()/2)
	if hits[0] != 0 || hits[2] != 0 {
		t.Fatalf("neighbor segments fired: hits=%v", hits)
	}
	if hits[1] != 1 {
		t.Fatalf("middle segment OnClick count: got %d want 1", hits[1])
	}
}

// TestGoldenButtonGroup renders a 3-button outline group light and dark.
func TestGoldenButtonGroup(t *testing.T) {
	build := func() widget.Widget {
		return primitives.VBox(graft.ButtonGroup(
			graft.Button("Years").Outline(),
			graft.Button("Months").Outline(),
			graft.Button("Days").Outline(),
		)).Padding(12).CrossAlign(primitives.CrossAxisStart)
	}
	gtest.GoldenLightDark(t, "buttongroup-outline", build)

	gtest.GoldenLightDark(t, "buttongroup-default", func() widget.Widget {
		return primitives.VBox(graft.ButtonGroup(
			graft.Button("Left"),
			graft.Button("Center"),
			graft.Button("Right"),
		)).Padding(12).CrossAlign(primitives.CrossAxisStart)
	})
}

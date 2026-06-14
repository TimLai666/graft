package graft_test

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// buildToggleGroup returns a standard three-segment group with the first
// selected.
func buildToggleGroup(outline bool) *graft.ToggleGroupWidget {
	g := graft.ToggleGroup(
		graft.ToggleGroupItem("bold", "B"),
		graft.ToggleGroupItem("italic", "I"),
		graft.ToggleGroupItem("underline", "U"),
	).Type("single").Value("bold")
	if outline {
		g.Outline()
	}
	return g
}

func TestToggleGroupSpecFusedLayout(t *testing.T) {
	lightTokens(t)
	g := buildToggleGroup(false)
	uitest.LayoutWidget(g, 400, 200)

	items := g.Children()
	i0 := items[0].(bounded)
	i1 := items[1].(bounded)
	i2 := items[2].(bounded)

	// Group height = toggle-group item height (h-8).
	if !approx(g.Bounds().Height(), metrics.ToggleGroup.ItemHeight) {
		t.Fatalf("group height = %v, want %v", g.Bounds().Height(), metrics.ToggleGroup.ItemHeight)
	}
	// Segments overlap by 1px (fused borders).
	if !approx(i1.Bounds().Min.X, i0.Bounds().Max.X-metrics.ToggleGroup.Overlap) {
		t.Fatalf("segment 2 x = %v, want %v (overlap)", i1.Bounds().Min.X, i0.Bounds().Max.X-metrics.ToggleGroup.Overlap)
	}
	if !approx(i2.Bounds().Min.X, i1.Bounds().Max.X-metrics.ToggleGroup.Overlap) {
		t.Fatalf("segment 3 x = %v, want overlap", i2.Bounds().Min.X)
	}
}

func TestToggleGroupSpecSelectedFill(t *testing.T) {
	tok := lightTokens(t)
	g := buildToggleGroup(false)
	uitest.LayoutWidget(g, 400, 200)
	mc := uitest.DrawWidget(g)

	// Selected segment ("bold") fills with --accent.
	found := false
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Accent {
			found = true
		}
	}
	for _, r := range mc.Rects {
		if r.Color == tok.Accent {
			found = true
		}
	}
	if !found {
		t.Fatalf("selected segment must fill accent; roundrects %+v rects %+v", mc.RoundRects, mc.Rects)
	}
}

func TestToggleGroupSpecOutlineBorders(t *testing.T) {
	tok := lightTokens(t)
	g := buildToggleGroup(true)
	uitest.LayoutWidget(g, 400, 200)
	mc := uitest.DrawWidget(g)

	// Outline group: 1px input borders + a group shadow-xs.
	borders := 0
	for _, sr := range mc.StrokeRoundRects {
		if sr.Color == tok.Input && approx(sr.StrokeWidth, metrics.ToggleGroup.BorderWidth) {
			borders++
		}
	}
	if borders < 3 {
		t.Fatalf("outline group should stroke each of 3 segments; got %d; strokes %+v", borders, mc.StrokeRoundRects)
	}
}

func TestToggleGroupSingleExclusive(t *testing.T) {
	lightTokens(t)
	sig := state.NewSignal("bold")
	g := graft.ToggleGroup(
		graft.ToggleGroupItem("bold", "B"),
		graft.ToggleGroupItem("italic", "I"),
	).Type("single").Bind(sig)
	uitest.LayoutWidget(g, 400, 200)

	ctx := uitest.NewMockContext()
	i1 := g.Children()[1].(bounded)
	pt := i1.Bounds().Center()
	g.Event(ctx, uitest.Click(pt.X, pt.Y))
	g.Event(ctx, uitest.Release(pt.X, pt.Y))
	if sig.Get() != "italic" {
		t.Fatalf("single mode: clicking italic should select it; signal=%q", sig.Get())
	}
}

func TestGoldenToggleGroup(t *testing.T) {
	gtest.GoldenLightDark(t, "togglegroup-default", func() widget.Widget {
		return primitives.Box(buildToggleGroup(false)).Padding(16)
	})
	gtest.GoldenLightDark(t, "togglegroup-outline", func() widget.Widget {
		return primitives.Box(buildToggleGroup(true)).Padding(16)
	})
}

package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// inputGroupForceLight pins the theme to light and returns active tokens plus a
// restore function.
func inputGroupForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

// inputGroupRow wraps a group in padding so the focus ring is not clipped.
func inputGroupRow(w widget.Widget) widget.Widget {
	return primitives.VBox(w).Padding(12)
}

// TestInputGroupLayoutHeight pins the container to h-9 (36px).
func TestInputGroupLayoutHeight(t *testing.T) {
	_, restore := inputGroupForceLight(t)
	defer restore()

	g := graft.InputGroup().Leading(graft.InputGroupAddon(icons.Search)).Placeholder("Search").W(260)
	size := g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(300, 100)))
	if size.Height != metrics.InputGroup.Height {
		t.Fatalf("height: got %v want %v", size.Height, metrics.InputGroup.Height)
	}
	if size.Width != 260 {
		t.Fatalf("width: got %v want 260", size.Width)
	}
}

// TestInputGroupSpecBorder asserts a single 1px Input border at the group
// radius and no opaque fill in light mode.
func TestInputGroupSpecBorder(t *testing.T) {
	tok, restore := inputGroupForceLight(t)
	defer restore()

	g := graft.InputGroup().Leading(graft.InputGroupAddon(icons.Search)).Placeholder("Search").W(260)
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(300, 100)))
	canvas := uitest.DrawWidget(g)

	radius := graft.CurrentTheme().RadiusMD()

	var borders []uitest.StrokeRoundRectCall
	for _, s := range canvas.StrokeRoundRects {
		if s.StrokeWidth == metrics.InputGroup.BorderWidth {
			borders = append(borders, s)
		}
	}
	if len(borders) != 1 {
		t.Fatalf("want 1 border stroke, got %d", len(borders))
	}
	if borders[0].Color != tok.Input {
		t.Fatalf("border color: got %+v want Input", borders[0].Color)
	}
	if want := radius - metrics.InputGroup.BorderWidth/2; borders[0].Radius != want {
		t.Fatalf("border radius: got %v want %v", borders[0].Radius, want)
	}

	for _, rr := range canvas.RoundRects {
		if rr.Color.A == 1 {
			t.Fatalf("unexpected opaque fill in light mode: %+v", rr)
		}
	}
}

// TestInputGroupSpecFocusedRing asserts the focus state lights the ring and
// solid Ring border.
func TestInputGroupSpecFocusedRing(t *testing.T) {
	tok, restore := inputGroupForceLight(t)
	defer restore()

	g := graft.InputGroup().Leading(graft.InputGroupAddon(icons.Search)).Value("query").W(260)
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(300, 100)))
	g.SetFocused(true)
	canvas := uitest.DrawWidget(g)

	var ring *uitest.StrokeRoundRectCall
	for idx := range canvas.StrokeRoundRects {
		if canvas.StrokeRoundRects[idx].StrokeWidth == metrics.RingWidth {
			ring = &canvas.StrokeRoundRects[idx]
		}
	}
	if ring == nil {
		t.Fatal("no focus ring stroke (width 3) found")
	}
	wantRing := tok.Ring
	wantRing.A = metrics.RingAlpha
	if ring.Color != wantRing {
		t.Fatalf("ring color: got %+v want Ring@0.5", ring.Color)
	}

	for _, s := range canvas.StrokeRoundRects {
		if s.StrokeWidth == metrics.InputGroup.BorderWidth && s.Color != tok.Ring {
			t.Fatalf("focused border should be solid Ring, got %+v", s.Color)
		}
	}
}

// TestInputGroupValue verifies the inner field carries the value.
func TestInputGroupValue(t *testing.T) {
	_, restore := inputGroupForceLight(t)
	defer restore()

	g := graft.InputGroup().Value("hello")
	if g.Text() != "hello" {
		t.Fatalf("text: got %q want hello", g.Text())
	}
}

// TestGoldenInputGroup renders a search input and a trailing-button input.
func TestGoldenInputGroup(t *testing.T) {
	gtest.GoldenLightDark(t, "inputgroup-search", func() widget.Widget {
		return inputGroupRow(
			graft.InputGroup().
				Leading(graft.InputGroupAddon(icons.Search)).
				Placeholder("Search...").
				W(260),
		)
	})
	gtest.GoldenLightDark(t, "inputgroup-trailing-button", func() widget.Widget {
		return inputGroupRow(
			graft.InputGroup().
				Value("https://example.com").
				Trailing(graft.InputGroupButton(graft.Button("Copy").Outline())).
				W(300),
		)
	})
	gtest.GoldenLightDark(t, "inputgroup-text-suffix", func() widget.Widget {
		return inputGroupRow(
			graft.InputGroup().
				Value("75").
				Trailing(graft.InputGroupText("kg")).
				W(180),
		)
	})
}

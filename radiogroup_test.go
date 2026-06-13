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
	"github.com/TimLai666/graft/theme"
)

// radioForceLight pins the current theme to light for a spec test.
func radioForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

// TestRadioGroupSpecItemUnselected asserts the circle border (1px Input) and
// no dot when unselected.
func TestRadioGroupSpecItemUnselected(t *testing.T) {
	tok, restore := radioForceLight(t)
	defer restore()

	g := graft.RadioGroup(
		graft.RadioGroupItem("a", "Option A"),
		graft.RadioGroupItem("b", "Option B"),
	).Value("a")
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))

	// Draw item "b" directly (unselected).
	item := graft.RadioGroupItem("b", "Option B")
	g2 := graft.RadioGroup(item).Value("a")
	g2.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	canvas := uitest.DrawWidget(item)

	var border *uitest.StrokeCircleCall
	for idx := range canvas.StrokeCircles {
		if canvas.StrokeCircles[idx].StrokeWidth == metrics.RadioGroup.BorderWidth {
			border = &canvas.StrokeCircles[idx]
		}
	}
	if border == nil {
		t.Fatal("no circle border stroke found")
	}
	if border.Color != tok.Input {
		t.Fatalf("border color: got %+v want Input %+v", border.Color, tok.Input)
	}
	// No Primary dot when unselected.
	for _, c := range canvas.Circles {
		if c.Color == tok.Primary {
			t.Fatalf("unselected item should not draw a Primary dot: %+v", c)
		}
	}
}

// TestRadioGroupSpecItemSelected asserts the 8px Primary dot when selected.
func TestRadioGroupSpecItemSelected(t *testing.T) {
	tok, restore := radioForceLight(t)
	defer restore()

	item := graft.RadioGroupItem("a", "Option A")
	g := graft.RadioGroup(item).Value("a")
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	canvas := uitest.DrawWidget(item)

	var dot *uitest.DrawCircleCall
	for idx := range canvas.Circles {
		if canvas.Circles[idx].Color == tok.Primary {
			dot = &canvas.Circles[idx]
		}
	}
	if dot == nil {
		t.Fatal("selected item should draw a Primary dot")
	}
	if dot.Radius != metrics.RadioGroup.DotSize/2 {
		t.Fatalf("dot radius: got %v want %v", dot.Radius, metrics.RadioGroup.DotSize/2)
	}
}

// TestRadioGroupSpecFocusRing asserts a circular 3px Ring/50 ring when an item
// is focused.
func TestRadioGroupSpecFocusRing(t *testing.T) {
	tok, restore := radioForceLight(t)
	defer restore()

	item := graft.RadioGroupItem("a", "Option A")
	g := graft.RadioGroup(item).Value("a")
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	item.SetFocused(true)
	canvas := uitest.DrawWidget(item)

	var ring *uitest.StrokeCircleCall
	for idx := range canvas.StrokeCircles {
		if canvas.StrokeCircles[idx].StrokeWidth == metrics.RingWidth {
			ring = &canvas.StrokeCircles[idx]
		}
	}
	if ring == nil {
		t.Fatal("no circular focus ring (width 3)")
	}
	wantRing := tok.Ring
	wantRing.A = metrics.RingAlpha
	if ring.Color != wantRing {
		t.Fatalf("ring color: got %+v want Ring@0.5 %+v", ring.Color, wantRing)
	}
}

// TestRadioGroupSingleSelection verifies clicking an item selects only it.
func TestRadioGroupSingleSelection(t *testing.T) {
	_, restore := radioForceLight(t)
	defer restore()

	var got []string
	a := graft.RadioGroupItem("a", "A")
	b := graft.RadioGroupItem("b", "B")
	g := graft.RadioGroup(a, b).Value("a").OnChange(func(v string) { got = append(got, v) })
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))

	// Click item b (inside its bounds).
	ctx := uitest.NewMockContext()
	bp := b.Bounds()
	center := bp.Center()
	b.Event(ctx, uitest.Click(center.X, center.Y))
	b.Event(ctx, uitest.Release(center.X, center.Y))

	if g.Selected() != "b" {
		t.Fatalf("selected: got %q want b", g.Selected())
	}
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("OnChange: got %v want [b]", got)
	}
}

// TestRadioGroupArrowKeys verifies arrow keys move the selection.
func TestRadioGroupArrowKeys(t *testing.T) {
	_, restore := radioForceLight(t)
	defer restore()

	a := graft.RadioGroupItem("a", "A")
	b := graft.RadioGroupItem("b", "B")
	c := graft.RadioGroupItem("c", "C")
	g := graft.RadioGroup(a, b, c).Value("a")
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	a.SetFocused(true)

	ctx := uitest.NewMockContext()
	if !g.Event(ctx, uitest.KeyPress(event.KeyDown, event.ModNone)) {
		t.Fatal("Down arrow should be consumed")
	}
	if g.Selected() != "b" {
		t.Fatalf("after Down: got %q want b", g.Selected())
	}
	g.Event(ctx, uitest.KeyPress(event.KeyUp, event.ModNone))
	if g.Selected() != "a" {
		t.Fatalf("after Up: got %q want a", g.Selected())
	}
	// Wrap-around: Up from first → last.
	g.Event(ctx, uitest.KeyPress(event.KeyUp, event.ModNone))
	if g.Selected() != "c" {
		t.Fatalf("after Up wrap: got %q want c", g.Selected())
	}
}

// TestRadioGroupBind verifies controlled selection.
func TestRadioGroupBind(t *testing.T) {
	_, restore := radioForceLight(t)
	defer restore()

	sig := state.NewSignal("a")
	a := graft.RadioGroupItem("a", "A")
	b := graft.RadioGroupItem("b", "B")
	g := graft.RadioGroup(a, b).Bind(sig)
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))

	ctx := uitest.NewMockContext()
	bc := b.Bounds().Center()
	b.Event(ctx, uitest.Click(bc.X, bc.Y))
	b.Event(ctx, uitest.Release(bc.X, bc.Y))
	if sig.Get() != "b" {
		t.Fatalf("bound signal: got %q want b", sig.Get())
	}
}

// TestRadioGroupDisabledItem verifies a disabled item ignores clicks.
func TestRadioGroupDisabledItem(t *testing.T) {
	_, restore := radioForceLight(t)
	defer restore()

	a := graft.RadioGroupItem("a", "A")
	b := graft.RadioGroupItem("b", "B").Disabled(true)
	g := graft.RadioGroup(a, b).Value("a")
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	if b.IsFocusable() {
		t.Fatal("disabled item should not be focusable")
	}
	ctx := uitest.NewMockContext()
	bc := b.Bounds().Center()
	b.Event(ctx, uitest.Click(bc.X, bc.Y))
	b.Event(ctx, uitest.Release(bc.X, bc.Y))
	if g.Selected() != "a" {
		t.Fatalf("disabled item should not change selection: got %q", g.Selected())
	}
}

// TestGoldenRadioGroup renders a 3-item group (one selected), a disabled item,
// and a focused item in light and dark.
func TestGoldenRadioGroup(t *testing.T) {
	gtest.GoldenLightDark(t, "radiogroup-default", func() widget.Widget {
		return radioWrap(graft.RadioGroup(
			graft.RadioGroupItem("comfortable", "Comfortable"),
			graft.RadioGroupItem("compact", "Compact"),
			graft.RadioGroupItem("spacious", "Spacious"),
		).Value("comfortable"))
	})
	gtest.GoldenLightDark(t, "radiogroup-disabled", func() widget.Widget {
		return radioWrap(graft.RadioGroup(
			graft.RadioGroupItem("a", "Enabled"),
			graft.RadioGroupItem("b", "Disabled").Disabled(true),
		).Value("a"))
	})
	gtest.GoldenLightDark(t, "radiogroup-focused", func() widget.Widget {
		first := graft.RadioGroupItem("a", "Focused option")
		second := graft.RadioGroupItem("b", "Other option")
		g := graft.RadioGroup(first, second).Value("a")
		first.SetFocused(true)
		return radioWrap(g)
	})
	gtest.GoldenLightDark(t, "radiogroup-horizontal", func() widget.Widget {
		return radioWrap(graft.RadioGroup(
			graft.RadioGroupItem("yes", "Yes"),
			graft.RadioGroupItem("no", "No"),
		).Value("yes").Horizontal())
	})
}

func radioWrap(w widget.Widget) widget.Widget {
	return primitives.VBox(w).Padding(12)
}

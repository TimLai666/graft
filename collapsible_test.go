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
)

// bounded is the bounds accessor exposed by every graft widget (via
// WidgetBase) but not part of the widget.Widget interface.
type bounded interface{ Bounds() geometry.Rect }

// buildCollapsible returns a standard collapsible (ghost-button trigger,
// card content) for tests.
func buildCollapsible(open bool) *graft.CollapsibleWidget {
	return graft.Collapsible(
		graft.Button("Can I use this in my project?").Ghost(),
		graft.Card(graft.CardContent(graft.Text("Yes. Free to use for personal and commercial."))),
	).Open(open)
}

func TestCollapsibleOpenShowsContent(t *testing.T) {
	lightTokens(t)
	c := buildCollapsible(true)
	uitest.LayoutWidget(c, 600, 400)
	mc := uitest.DrawWidget(c)

	found := false
	for _, st := range mc.StyledTexts {
		if st.Text == "Yes. Free to use for personal and commercial." {
			found = true
		}
	}
	if !found {
		t.Fatal("open collapsible must draw its content text")
	}
}

func TestCollapsibleClosedHidesContent(t *testing.T) {
	lightTokens(t)
	c := buildCollapsible(false)
	uitest.LayoutWidget(c, 600, 400)
	mc := uitest.DrawWidget(c)

	for _, st := range mc.StyledTexts {
		if st.Text == "Yes. Free to use for personal and commercial." {
			t.Fatal("closed collapsible must not draw its content")
		}
	}

	// Closed root height equals the trigger height only.
	trig := c.Children()[0].(bounded)
	if !approx(c.Bounds().Height(), trig.Bounds().Height()) {
		t.Fatalf("closed root height = %v, want trigger height %v",
			c.Bounds().Height(), trig.Bounds().Height())
	}
}

func TestCollapsibleClickToggles(t *testing.T) {
	lightTokens(t)
	sig := state.NewSignal(false)
	var observed []bool
	c := graft.Collapsible(
		graft.Button("Toggle").Ghost(),
		graft.Text("Body"),
	).Bind(sig).OnOpenChange(func(o bool) { observed = append(observed, o) })
	uitest.LayoutWidget(c, 600, 400)

	ctx := uitest.NewMockContext()
	trig := c.Children()[0].(bounded)
	pt := trig.Bounds().Center()

	c.Event(ctx, uitest.Click(pt.X, pt.Y))
	c.Event(ctx, uitest.Release(pt.X, pt.Y))
	if !sig.Get() {
		t.Fatalf("first click should open; signal=%v", sig.Get())
	}

	uitest.LayoutWidget(c, 600, 400)
	c.Event(ctx, uitest.Click(pt.X, pt.Y))
	c.Event(ctx, uitest.Release(pt.X, pt.Y))
	if sig.Get() {
		t.Fatalf("second click should close; signal=%v", sig.Get())
	}

	if len(observed) != 2 || observed[0] != true || observed[1] != false {
		t.Fatalf("OnOpenChange sequence = %v, want [true false]", observed)
	}
}

func TestGoldenCollapsible(t *testing.T) {
	gtest.GoldenLightDark(t, "collapsible-open", func() widget.Widget {
		return primitives.Box(buildCollapsible(true)).Padding(16)
	})
	gtest.GoldenLightDark(t, "collapsible-closed", func() widget.Widget {
		return primitives.Box(buildCollapsible(false)).Padding(16)
	})
}

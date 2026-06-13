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

func TestToggleSpecSizes(t *testing.T) {
	lightTokens(t)
	cases := []struct {
		name  string
		build func() *graft.ToggleWidget
		want  metrics.ToggleSize
	}{
		{"default", func() *graft.ToggleWidget { return graft.Toggle("B") }, metrics.Toggle.Default},
		{"sm", func() *graft.ToggleWidget { return graft.Toggle("B").Sm() }, metrics.Toggle.SM},
		{"lg", func() *graft.ToggleWidget { return graft.Toggle("B").Lg() }, metrics.Toggle.LG},
	}
	for _, c := range cases {
		w := c.build()
		uitest.LayoutWidget(w, 400, 200)
		if !approx(w.Bounds().Height(), c.want.Height) {
			t.Fatalf("%s height = %v, want %v", c.name, w.Bounds().Height(), c.want.Height)
		}
		// min-width enforced even for a 1-char label.
		if w.Bounds().Width() < c.want.MinWidth {
			t.Fatalf("%s width = %v, want >= min-w %v", c.name, w.Bounds().Width(), c.want.MinWidth)
		}
	}
}

func TestToggleSpecOnState(t *testing.T) {
	tok := lightTokens(t)
	on := graft.Toggle("Bold").On(true)
	uitest.LayoutWidget(on, 400, 200)
	mc := uitest.DrawWidget(on)

	// On fill = --accent at radius-md.
	found := false
	th := graft.CurrentTheme()
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Accent && approx(rr.Radius, th.RadiusMD()) {
			found = true
		}
	}
	if !found {
		t.Fatalf("on toggle must fill with accent; roundrects: %+v", mc.RoundRects)
	}
	// Label color = accent-foreground.
	for _, st := range mc.StyledTexts {
		if st.Text == "Bold" && st.Style.Color != tok.AccentForeground {
			t.Fatalf("on label color = %+v, want accent-foreground", st.Style.Color)
		}
	}
}

func TestToggleSpecOffDefaultNoFill(t *testing.T) {
	tok := lightTokens(t)
	off := graft.Toggle("Bold")
	uitest.LayoutWidget(off, 400, 200)
	mc := uitest.DrawWidget(off)

	// Default variant off: transparent (no accent/muted fill, no border).
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Accent || rr.Color == tok.Muted {
			t.Fatalf("off default toggle must be transparent; got fill %+v", rr)
		}
	}
	if len(mc.StrokeRoundRects) != 0 {
		t.Fatalf("off default toggle must have no border; strokes: %+v", mc.StrokeRoundRects)
	}
	// Label = foreground.
	for _, st := range mc.StyledTexts {
		if st.Text == "Bold" && st.Style.Color != tok.Foreground {
			t.Fatalf("off label color = %+v, want foreground", st.Style.Color)
		}
	}
}

func TestToggleSpecOutlineBorder(t *testing.T) {
	tok := lightTokens(t)
	o := graft.Toggle("Bold").Outline()
	uitest.LayoutWidget(o, 400, 200)
	mc := uitest.DrawWidget(o)

	found := false
	for _, sr := range mc.StrokeRoundRects {
		if sr.Color == tok.Input && approx(sr.StrokeWidth, metrics.Toggle.BorderWidth) {
			found = true
		}
	}
	if !found {
		t.Fatalf("outline toggle must draw a 1px input border; strokes: %+v", mc.StrokeRoundRects)
	}
}

func TestToggleInteraction(t *testing.T) {
	lightTokens(t)
	sig := state.NewSignal(false)
	var got []bool
	tg := graft.Toggle("Bold").Bind(sig).OnChange(func(v bool) { got = append(got, v) })
	uitest.LayoutWidget(tg, 400, 200)

	ctx := uitest.NewMockContext()
	pt := tg.Bounds().Center()
	tg.Event(ctx, uitest.Click(pt.X, pt.Y))
	tg.Event(ctx, uitest.Release(pt.X, pt.Y))
	if !sig.Get() {
		t.Fatal("click should turn the toggle on")
	}
	tg.Event(ctx, uitest.Click(pt.X, pt.Y))
	tg.Event(ctx, uitest.Release(pt.X, pt.Y))
	if sig.Get() {
		t.Fatal("second click should turn the toggle off")
	}
	if len(got) != 2 || got[0] != true || got[1] != false {
		t.Fatalf("OnChange sequence = %v, want [true false]", got)
	}
}

func TestGoldenToggle(t *testing.T) {
	row := func(variant func(*graft.ToggleWidget) *graft.ToggleWidget) widget.Widget {
		mk := func(on bool) *graft.ToggleWidget {
			tg := variant(graft.Toggle("Bold"))
			return tg.On(on)
		}
		return primitives.HBox(mk(false), mk(true)).Gap(12).CrossAlign(primitives.CrossAxisCenter)
	}
	gtest.GoldenLightDark(t, "toggle-default", func() widget.Widget {
		return primitives.Box(row(func(tg *graft.ToggleWidget) *graft.ToggleWidget { return tg })).Padding(16)
	})
	gtest.GoldenLightDark(t, "toggle-outline", func() widget.Widget {
		return primitives.Box(row(func(tg *graft.ToggleWidget) *graft.ToggleWidget { return tg.Outline() })).Padding(16)
	})
}

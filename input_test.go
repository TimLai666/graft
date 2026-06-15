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

// inputRow wraps a control in padding so the focus ring (which paints outside
// bounds) is not clipped at the golden image edge.
func inputRow(w widget.Widget) widget.Widget {
	return primitives.VBox(w).Padding(12)
}

// inputForceLight pins the current theme to light for a spec test and returns
// the active token set plus a restore function.
func inputForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

// laidOut runs Layout with a loose 400-wide box and returns the widget.
func laidOutInput(w widget.Widget) widget.Widget {
	w.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	return w
}

// TestInputLayoutHeight pins the h-9 (36px) override over the inner 48px.
func TestInputLayoutHeight(t *testing.T) {
	_, restore := inputForceLight(t)
	defer restore()

	in := graft.Input().Placeholder("Email")
	size := in.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(300, 400)))
	if size.Height != metrics.Input.Height {
		t.Fatalf("height: got %v want %v", size.Height, metrics.Input.Height)
	}
	if size.Width != 300 {
		t.Fatalf("width: got %v want 300 (fill)", size.Width)
	}

	withW := graft.Input().W(180)
	size = withW.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(300, 400)))
	if size.Width != 180 {
		t.Fatalf("explicit width: got %v want 180", size.Width)
	}
}

// TestInputSpecEmptyLight asserts the default empty input draws a 1px Input
// border via InsideBorder (no fill in light mode) and a muted placeholder.
func TestInputSpecEmptyLight(t *testing.T) {
	tok, restore := inputForceLight(t)
	defer restore()

	in := graft.Input().Placeholder("Email")
	laidOutInput(in)
	canvas := uitest.DrawWidget(in)

	radius := graft.CurrentTheme().RadiusLG()

	// BorderFill in light mode: outer round-rect = Input border at full
	// bounds/radius, inner round-rect = Background fill inset by 1px.
	var border, fill *uitest.DrawRoundRectCall
	for idx := range canvas.RoundRects {
		switch canvas.RoundRects[idx].Color {
		case tok.Input:
			border = &canvas.RoundRects[idx]
		case tok.Background:
			fill = &canvas.RoundRects[idx]
		}
	}
	if border == nil {
		t.Fatal("no Input-colored border round-rect found")
	}
	if border.Radius != radius {
		t.Fatalf("border radius: got %v want %v", border.Radius, radius)
	}
	if fill == nil {
		t.Fatal("no Background-colored inner fill round-rect found")
	}
	if fill.Radius != radius-metrics.Input.BorderWidth {
		t.Fatalf("inner fill radius: got %v want %v", fill.Radius, radius-metrics.Input.BorderWidth)
	}

	// No 1px border stroke any more (border is now a fill).
	for _, s := range canvas.StrokeRoundRects {
		if s.StrokeWidth == metrics.Input.BorderWidth {
			t.Fatalf("unexpected 1px border stroke: %+v", s)
		}
	}

	// Placeholder text in muted-foreground.
	if len(canvas.StyledTexts) != 1 {
		t.Fatalf("want 1 styled text (placeholder), got %d", len(canvas.StyledTexts))
	}
	ph := canvas.StyledTexts[0]
	if ph.Text != "Email" {
		t.Fatalf("placeholder text: got %q", ph.Text)
	}
	if ph.Style.Color != tok.MutedForeground {
		t.Fatalf("placeholder color: got %+v want MutedForeground %+v", ph.Style.Color, tok.MutedForeground)
	}
	if ph.Style.FontSize != metrics.Input.FontSize {
		t.Fatalf("placeholder font size: got %v want %v", ph.Style.FontSize, metrics.Input.FontSize)
	}
}

// TestInputSpecFocusedRing asserts the focus state draws the 3px Ring/50 ring
// and a solid Ring border.
func TestInputSpecFocusedRing(t *testing.T) {
	tok, restore := inputForceLight(t)
	defer restore()

	in := graft.Input().Placeholder("Email")
	laidOutInput(in)
	in.SetFocused(true)
	canvas := uitest.DrawWidget(in)

	radius := graft.CurrentTheme().RadiusLG()

	// 3px ring stroke in Ring @50%.
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
		t.Fatalf("ring color: got %+v want Ring@0.5 %+v", ring.Color, wantRing)
	}
	if ring.Radius != radius+metrics.RingWidth/2 {
		t.Fatalf("ring radius: got %v want %v", ring.Radius, radius+metrics.RingWidth/2)
	}

	// Border now solid Ring.
	for _, s := range canvas.StrokeRoundRects {
		if s.StrokeWidth == metrics.Input.BorderWidth && s.Color != tok.Ring {
			t.Fatalf("focused border should be solid Ring, got %+v", s.Color)
		}
	}
}

// TestInputSpecInvalidRing asserts the invalid state draws a destructive ring
// and border.
func TestInputSpecInvalidRing(t *testing.T) {
	tok, restore := inputForceLight(t)
	defer restore()

	in := graft.Input().Placeholder("Email").Invalid(true)
	laidOutInput(in)
	canvas := uitest.DrawWidget(in)

	var ring *uitest.StrokeRoundRectCall
	for idx := range canvas.StrokeRoundRects {
		if canvas.StrokeRoundRects[idx].StrokeWidth == metrics.RingWidth {
			ring = &canvas.StrokeRoundRects[idx]
		}
	}
	if ring == nil {
		t.Fatal("no invalid ring stroke (width 3) found")
	}
	wantRing := tok.Destructive
	wantRing.A = metrics.InvalidRingAlphaLight
	if ring.Color != wantRing {
		t.Fatalf("ring color: got %+v want Destructive@0.2 %+v", ring.Color, wantRing)
	}

	for _, s := range canvas.StrokeRoundRects {
		if s.StrokeWidth == metrics.Input.BorderWidth && s.Color != tok.Destructive {
			t.Fatalf("invalid border should be Destructive, got %+v", s.Color)
		}
	}
}

// TestInputSpecFilled asserts a filled input draws its text in Foreground.
func TestInputSpecFilled(t *testing.T) {
	tok, restore := inputForceLight(t)
	defer restore()

	in := graft.Input().Value("hello")
	laidOutInput(in)
	canvas := uitest.DrawWidget(in)

	if len(canvas.StyledTexts) != 1 {
		t.Fatalf("want 1 styled text, got %d", len(canvas.StyledTexts))
	}
	txt := canvas.StyledTexts[0]
	if txt.Text != "hello" {
		t.Fatalf("text: got %q want hello", txt.Text)
	}
	if txt.Style.Color != tok.Foreground {
		t.Fatalf("text color: got %+v want Foreground %+v", txt.Style.Color, tok.Foreground)
	}
}

// TestInputDisabledNotFocusable pins disabled inputs out of focus traversal.
func TestInputDisabledNotFocusable(t *testing.T) {
	_, restore := inputForceLight(t)
	defer restore()

	in := graft.Input().Disabled(true)
	if in.IsFocusable() {
		t.Fatal("disabled input should not be focusable")
	}
	enabled := graft.Input()
	if !enabled.IsFocusable() {
		t.Fatal("enabled input should be focusable")
	}
}

// TestInputBindControlled verifies a bound signal drives the rendered text.
func TestInputBindControlled(t *testing.T) {
	_, restore := inputForceLight(t)
	defer restore()

	sig := state.NewSignal("one")
	in := graft.Input().Bind(sig)
	if in.Text() != "one" {
		t.Fatalf("initial bound text: got %q want one", in.Text())
	}
	sig.Set("two")
	if in.Text() != "two" {
		t.Fatalf("after sig.Set: got %q want two", in.Text())
	}
}

// TestGoldenInput renders the five Input states in light and dark.
func TestGoldenInput(t *testing.T) {
	gtest.GoldenLightDark(t, "input-empty", func() widget.Widget {
		return inputRow(graft.Input().Placeholder("Email").W(220))
	})
	gtest.GoldenLightDark(t, "input-filled", func() widget.Widget {
		return inputRow(graft.Input().Value("m@example.com").W(220))
	})
	gtest.GoldenLightDark(t, "input-focused", func() widget.Widget {
		in := graft.Input().Value("m@example.com").W(220)
		in.SetFocused(true)
		return inputRow(in)
	})
	gtest.GoldenLightDark(t, "input-invalid", func() widget.Widget {
		return inputRow(graft.Input().Value("nope").Invalid(true).W(220))
	})
	gtest.GoldenLightDark(t, "input-disabled", func() widget.Widget {
		return inputRow(graft.Input().Value("disabled").Disabled(true).W(220))
	})
}

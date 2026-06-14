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

// selectGoldenWrap pads a widget so its shadow and focus ring are not
// clipped at the canvas edge in golden renders.
func selectGoldenWrap(w widget.Widget) widget.Widget {
	return primitives.VBox(w).Padding(24)
}

// selectLightTheme forces the current theme to light mode and returns its
// active tokens. It restores the previous mode via t.Cleanup. (Component-
// prefixed to avoid colliding with batch-1's shared test helpers.)
func selectLightTheme(t *testing.T) *theme.Tokens {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatalf("LoadAssets: %v", err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	t.Cleanup(func() { th.SetMode(prev) })
	return th.Active()
}

// selectLayout lays out a widget loosely and returns its size.
func selectLayout(t *testing.T, w widget.Widget) geometry.Size {
	t.Helper()
	return w.Layout(nil, geometry.Loose(geometry.Sz(10000, 10000)))
}

func TestSelectTriggerHeight(t *testing.T) {
	selectLightTheme(t)

	def := graft.Select(graft.SelectItem("a", "Apple")).Placeholder("Pick")
	if got := selectLayout(t, def).Height; got != metrics.Select.TriggerHeight {
		t.Errorf("default trigger height = %v, want %v", got, metrics.Select.TriggerHeight)
	}

	sm := graft.Select(graft.SelectItem("a", "Apple")).Placeholder("Pick").Sm()
	if got := selectLayout(t, sm).Height; got != metrics.Select.TriggerHeightSm {
		t.Errorf("sm trigger height = %v, want %v", got, metrics.Select.TriggerHeightSm)
	}
}

func TestSelectTriggerWidthPin(t *testing.T) {
	selectLightTheme(t)
	s := graft.Select(graft.SelectItem("a", "Apple")).Placeholder("Pick").W(240)
	if got := selectLayout(t, s).Width; got != 240 {
		t.Errorf("pinned width = %v, want 240", got)
	}
}

// drawSelectTrigger lays out and draws a trigger, returning the mock canvas.
func drawSelectTrigger(t *testing.T, s *graft.SelectWidget) *uitest.MockCanvas {
	t.Helper()
	selectLayout(t, s)
	return uitest.DrawWidget(s)
}

func TestSelectTriggerDefault(t *testing.T) {
	tok := selectLightTheme(t)
	s := graft.Select(graft.SelectItem("a", "Apple")).Placeholder("Pick a fruit")
	c := drawSelectTrigger(t, s)

	// Inside border in --input (1px), no fill in light mode.
	if len(c.StrokeRoundRects) != 1 {
		t.Fatalf("expected exactly 1 border stroke, got %d", len(c.StrokeRoundRects))
	}
	border := c.StrokeRoundRects[0]
	if border.StrokeWidth != metrics.Select.BorderWidth {
		t.Errorf("border width = %v, want %v", border.StrokeWidth, metrics.Select.BorderWidth)
	}
	if border.Color != tok.Input {
		t.Errorf("border color = %v, want input %v", border.Color, tok.Input)
	}
	// Placeholder text is muted-foreground.
	if len(c.StyledTexts) != 1 {
		t.Fatalf("expected 1 styled text, got %d", len(c.StyledTexts))
	}
	if c.StyledTexts[0].Style.Color != tok.MutedForeground {
		t.Errorf("placeholder color = %v, want muted-foreground %v", c.StyledTexts[0].Style.Color, tok.MutedForeground)
	}
	if c.StyledTexts[0].Text != "Pick a fruit" {
		t.Errorf("placeholder text = %q", c.StyledTexts[0].Text)
	}
	if c.StyledTexts[0].Style.FontSize != metrics.Select.FontSize {
		t.Errorf("font size = %v, want %v", c.StyledTexts[0].Style.FontSize, metrics.Select.FontSize)
	}
}

func TestSelectTriggerSelected(t *testing.T) {
	tok := selectLightTheme(t)
	s := graft.Select(graft.SelectItem("a", "Apple"), graft.SelectItem("b", "Banana")).
		Placeholder("Pick").Value("b")
	c := drawSelectTrigger(t, s)
	if len(c.StyledTexts) != 1 {
		t.Fatalf("expected 1 styled text, got %d", len(c.StyledTexts))
	}
	if c.StyledTexts[0].Text != "Banana" {
		t.Errorf("selected text = %q, want Banana", c.StyledTexts[0].Text)
	}
	if c.StyledTexts[0].Style.Color != tok.Foreground {
		t.Errorf("selected color = %v, want foreground %v", c.StyledTexts[0].Style.Color, tok.Foreground)
	}
}

func TestSelectTriggerFocused(t *testing.T) {
	tok := selectLightTheme(t)
	s := graft.Select(graft.SelectItem("a", "Apple")).Placeholder("Pick")
	selectLayout(t, s)
	// Drive keyboard focus so focus-visible is set.
	s.SetFocused(true)
	ctx := uitest.NewMockContext()
	ctx.FocusedVal = s
	s.Event(ctx, uitest.FocusGained())
	c := uitest.DrawWidgetWithContext(s, ctx)

	bounds := s.Bounds()
	radius := graft.CurrentTheme().RadiusLG()

	// Expect a focus ring stroke: width 3 at Expand(1.5), radius+1.5,
	// ring@0.5.
	wantRing := widget.Color{R: tok.Ring.R, G: tok.Ring.G, B: tok.Ring.B, A: metrics.RingAlpha}
	var foundRing, foundSolidBorder bool
	for _, sr := range c.StrokeRoundRects {
		if sr.StrokeWidth == metrics.RingWidth && sr.Color == wantRing &&
			sr.Bounds == bounds.Expand(metrics.RingWidth/2) && sr.Radius == radius+metrics.RingWidth/2 {
			foundRing = true
		}
		if sr.StrokeWidth == metrics.Select.BorderWidth && sr.Color == tok.Ring {
			foundSolidBorder = true
		}
	}
	if !foundRing {
		t.Errorf("focus ring not drawn (ring=%v); strokes=%+v", wantRing, c.StrokeRoundRects)
	}
	if !foundSolidBorder {
		t.Errorf("solid ring border not drawn; strokes=%+v", c.StrokeRoundRects)
	}
}

func TestSelectTriggerDisabled(t *testing.T) {
	tok := selectLightTheme(t)
	s := graft.Select(graft.SelectItem("a", "Apple")).Placeholder("Pick").Disabled(true)
	c := drawSelectTrigger(t, s)

	// Border is faded input (alpha halved).
	if len(c.StrokeRoundRects) != 1 {
		t.Fatalf("expected 1 border stroke, got %d", len(c.StrokeRoundRects))
	}
	wantBorder := tok.Input
	wantBorder.A *= metrics.DisabledOpacity
	if c.StrokeRoundRects[0].Color != wantBorder {
		t.Errorf("disabled border = %v, want faded input %v", c.StrokeRoundRects[0].Color, wantBorder)
	}
	// No shadow drawn when disabled.
	for _, rr := range c.RoundRects {
		if rr.Color.A > 0 && rr.Color.R == 0 && rr.Color.G == 0 && rr.Color.B == 0 {
			t.Errorf("disabled trigger should not draw a shadow layer: %+v", rr)
		}
	}
}

func TestSelectTriggerInvalid(t *testing.T) {
	tok := selectLightTheme(t)
	s := graft.Select(graft.SelectItem("a", "Apple")).Placeholder("Pick").Invalid(true)
	c := drawSelectTrigger(t, s)
	if len(c.StrokeRoundRects) != 1 {
		t.Fatalf("expected 1 border stroke, got %d", len(c.StrokeRoundRects))
	}
	if c.StrokeRoundRects[0].Color != tok.Destructive {
		t.Errorf("invalid border = %v, want destructive %v", c.StrokeRoundRects[0].Color, tok.Destructive)
	}
}

// TestSelectGoldens renders the trigger states and an open menu, light+dark.
func TestSelectGoldens(t *testing.T) {
	gtest.GoldenLightDark(t, "select-trigger-placeholder", func() widget.Widget {
		return selectGoldenWrap(graft.Select(
			graft.SelectItem("apple", "Apple"),
			graft.SelectItem("banana", "Banana"),
		).Placeholder("Select a fruit").W(220))
	})

	gtest.GoldenLightDark(t, "select-trigger-selected", func() widget.Widget {
		return selectGoldenWrap(graft.Select(
			graft.SelectItem("apple", "Apple"),
			graft.SelectItem("banana", "Banana"),
		).Placeholder("Select a fruit").Value("banana").W(220))
	})

	gtest.GoldenLightDark(t, "select-trigger-disabled", func() widget.Widget {
		return selectGoldenWrap(graft.Select(
			graft.SelectItem("apple", "Apple"),
		).Placeholder("Select a fruit").Disabled(true).W(220))
	})

	gtest.GoldenLightDark(t, "select-trigger-invalid", func() widget.Widget {
		return selectGoldenWrap(graft.Select(
			graft.SelectItem("apple", "Apple"),
		).Placeholder("Select a fruit").Invalid(true).W(220))
	})

	gtest.GoldenLightDark(t, "select-menu", func() widget.Widget {
		return graft.SelectMenuPreview(
			graft.Select(
				graft.SelectGroup("Fruits",
					graft.SelectItem("apple", "Apple"),
					graft.SelectItem("banana", "Banana"),
					graft.SelectItem("blueberry", "Blueberry").Disabled(true),
				),
				graft.SelectSeparator(),
				graft.SelectItem("carrot", "Carrot"),
			).Value("banana").W(220),
		)
	})
}

package graft_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// lightTokens forces light mode and returns the active tokens plus a
// restore function.
func lightTokens(t *testing.T) *theme.Tokens {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatalf("loading assets: %v", err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	t.Cleanup(func() { th.SetMode(prev) })
	return th.Active()
}

// darkTokens forces dark mode and returns the active tokens.
func darkTokens(t *testing.T) *theme.Tokens {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatalf("loading assets: %v", err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeDark)
	t.Cleanup(func() { th.SetMode(prev) })
	return th.Active()
}

func alpha(c widget.Color, a float32) widget.Color    { c.A = a; return c }
func mulAlpha(c widget.Color, f float32) widget.Color { c.A *= f; return c }

func TestButtonLayoutExactWidth(t *testing.T) {
	lightTokens(t)
	b := graft.Button("Button")
	size := uitest.LayoutWidget(b, 800, 600)

	if size.Height != metrics.Button.Default.Height {
		t.Errorf("default height = %v, want %v (h-8)", size.Height, metrics.Button.Default.Height)
	}
	wantW := textmetrics.Width("Geist Medium", 14, "Button") + 2*metrics.Button.Default.PadX
	if size.Width != wantW {
		t.Errorf("width = %v, want %v (text + 2*px-2.5)", size.Width, wantW)
	}
}

func TestButtonSizes(t *testing.T) {
	lightTokens(t)
	cases := []struct {
		name string
		b    *graft.ButtonWidget
		h    float32
	}{
		{"default", graft.Button("B"), 32},
		{"xs", graft.Button("B").XS(), 24},
		{"sm", graft.Button("B").Sm(), 32},
		{"lg", graft.Button("B").Lg(), 40},
	}
	for _, tc := range cases {
		size := uitest.LayoutWidget(tc.b, 800, 600)
		if size.Height != tc.h {
			t.Errorf("%s height = %v, want %v", tc.name, size.Height, tc.h)
		}
	}

	// Square icon sizes: 36/24/32/40.
	squares := []struct {
		name string
		b    *graft.ButtonWidget
		s    float32
	}{
		{"icon", graft.Button("").IconOnly(icons.Search), 36},
		{"icon-xs", graft.Button("").IconOnly(icons.Search).XS(), 24},
		{"icon-sm", graft.Button("").IconOnly(icons.Search).Sm(), 32},
		{"icon-lg", graft.Button("").IconOnly(icons.Search).Lg(), 40},
	}
	for _, tc := range squares {
		size := uitest.LayoutWidget(tc.b, 800, 600)
		if size.Width != tc.s || size.Height != tc.s {
			t.Errorf("%s = %vx%v, want %vx%v", tc.name, size.Width, size.Height, tc.s, tc.s)
		}
	}
}

func TestButtonDefaultDraw(t *testing.T) {
	tok := lightTokens(t)
	b := graft.Button("Button")
	uitest.LayoutWidget(b, 800, 600)
	c := uitest.DrawWidget(b)

	if len(c.RoundRects) != 1 {
		t.Fatalf("round rects = %d, want 1 (primary fill)", len(c.RoundRects))
	}
	fill := c.RoundRects[0]
	if fill.Color != tok.Primary {
		t.Errorf("fill color = %v, want primary %v", fill.Color, tok.Primary)
	}
	if fill.Radius != graft.CurrentTheme().RadiusLG() {
		t.Errorf("fill radius = %v, want %v (rounded-lg)", fill.Radius, graft.CurrentTheme().RadiusLG())
	}
	if fill.Bounds != b.Bounds() {
		t.Errorf("fill bounds = %v, want %v", fill.Bounds, b.Bounds())
	}

	if len(c.StyledTexts) != 1 {
		t.Fatalf("styled texts = %d, want 1", len(c.StyledTexts))
	}
	txt := c.StyledTexts[0]
	if txt.Style.FontFamily != "Geist Medium" {
		t.Errorf("font family = %q, want Geist Medium (font-medium 500)", txt.Style.FontFamily)
	}
	if txt.Style.FontSize != 14 {
		t.Errorf("font size = %v, want 14 (text-sm)", txt.Style.FontSize)
	}
	if txt.Style.Color != tok.PrimaryForeground {
		t.Errorf("text color = %v, want primary-foreground", txt.Style.Color)
	}
}

func TestButtonHoverFills(t *testing.T) {
	tok := lightTokens(t)
	cases := []struct {
		name string
		b    *graft.ButtonWidget
		want widget.Color
	}{
		{"default primary/90", graft.Button("B"), alpha(tok.Primary, 0.9)},
		{"secondary/80", graft.Button("B").Secondary(), alpha(tok.Secondary, 0.8)},
		{"destructive/90", graft.Button("B").Destructive(), alpha(tok.Destructive, 0.9)},
		{"outline accent", graft.Button("B").Outline(), tok.Accent},
		{"ghost accent", graft.Button("B").Ghost(), tok.Accent},
	}
	for _, tc := range cases {
		uitest.LayoutWidget(tc.b, 800, 600)
		ctx := uitest.NewMockContext()
		tc.b.Event(ctx, uitest.MouseEnter(2, 2))
		if got := ctx.Cursor(); got != widget.CursorPointer {
			t.Errorf("%s: hover cursor = %v, want pointer", tc.name, got)
		}
		c := uitest.DrawWidget(tc.b)
		fill := c.RoundRects[len(c.RoundRects)-1]
		if fill.Color != tc.want {
			t.Errorf("%s: hover fill = %v, want %v", tc.name, fill.Color, tc.want)
		}
	}
}

func TestButtonOutlineDraw(t *testing.T) {
	tok := lightTokens(t)
	b := graft.Button("Cancel").Outline()
	uitest.LayoutWidget(b, 800, 600)
	c := uitest.DrawWidget(b)

	// shadow-xs layers, then the border (BorderFill): an outer round-rect in
	// --border at full bounds, then an inner fill round-rect inset by 1px.
	if len(c.RoundRects) != len(metrics.ShadowXS)+2 {
		t.Fatalf("round rects = %d, want %d (shadow layers + border + fill)", len(c.RoundRects), len(metrics.ShadowXS)+2)
	}
	for i, l := range metrics.ShadowXS {
		got := c.RoundRects[i]
		want := widget.RGBA(0, 0, 0, l.Alpha)
		if got.Color != want {
			t.Errorf("shadow layer %d color = %v, want %v", i, got.Color, want)
		}
	}

	// No StrokeRoundRect border anymore — the border is two convex fills.
	if len(c.StrokeRoundRects) != 0 {
		t.Fatalf("strokes = %d, want 0 (border is now BorderFill)", len(c.StrokeRoundRects))
	}

	radius := graft.CurrentTheme().RadiusLG()

	// Outer round-rect = border color (--border) at full bounds, radius RadiusLG.
	border := c.RoundRects[len(metrics.ShadowXS)]
	if border.Color != tok.Border {
		t.Errorf("border color = %v, want border token", border.Color)
	}
	if border.Bounds != b.Bounds() {
		t.Errorf("border bounds = %v, want full bounds", border.Bounds)
	}
	if border.Radius != radius {
		t.Errorf("border radius = %v, want %v", border.Radius, radius)
	}

	// Inner round-rect = fill (--background) inset by the 1px border, radius-1.
	fill := c.RoundRects[len(metrics.ShadowXS)+1]
	if fill.Color != tok.Background {
		t.Errorf("outline fill = %v, want background", fill.Color)
	}
	if fill.Bounds != b.Bounds().Expand(-1) {
		t.Errorf("fill bounds = %v, want inset 1", fill.Bounds)
	}
	if want := radius - 1; fill.Radius != want {
		t.Errorf("fill radius = %v, want %v", fill.Radius, want)
	}
}

func TestButtonOutlineDark(t *testing.T) {
	tok := darkTokens(t)
	b := graft.Button("Cancel").Outline()
	uitest.LayoutWidget(b, 800, 600)
	c := uitest.DrawWidget(b)

	// BorderFill: outer round-rect = border (--input), inner round-rect = fill.
	border := c.RoundRects[len(metrics.ShadowXS)]
	if border.Color != tok.Input {
		t.Errorf("dark outline border = %v, want input token (alpha kept)", border.Color)
	}
	fill := c.RoundRects[len(metrics.ShadowXS)+1]
	if want := mulAlpha(tok.Input, 0.3); fill.Color != want {
		t.Errorf("dark outline fill = %v, want input/30 %v", fill.Color, want)
	}
	if len(c.StrokeRoundRects) != 0 {
		t.Errorf("strokes = %d, want 0 (border is now BorderFill)", len(c.StrokeRoundRects))
	}
}

func TestButtonFocusRing(t *testing.T) {
	tok := lightTokens(t)
	b := graft.Button("Save")
	uitest.LayoutWidget(b, 800, 600)
	b.SetFocused(true) // keyboard focus -> focus-visible
	c := uitest.DrawWidget(b)

	if len(c.StrokeRoundRects) != 1 {
		t.Fatalf("strokes = %d, want 1 (focus ring)", len(c.StrokeRoundRects))
	}
	ring := c.StrokeRoundRects[0]
	if ring.StrokeWidth != 3 {
		t.Errorf("ring width = %v, want 3", ring.StrokeWidth)
	}
	if want := graft.CurrentTheme().RadiusLG() + 1.5; ring.Radius != want {
		t.Errorf("ring radius = %v, want %v", ring.Radius, want)
	}
	if ring.Bounds != b.Bounds().Expand(1.5) {
		t.Errorf("ring bounds = %v, want Expand(1.5)", ring.Bounds)
	}
	if want := alpha(tok.Ring, 0.5); ring.Color != want {
		t.Errorf("ring color = %v, want ring/50 %v", ring.Color, want)
	}
}

func TestButtonDestructiveFocusRing(t *testing.T) {
	tok := lightTokens(t)
	b := graft.Button("Delete").Destructive()
	uitest.LayoutWidget(b, 800, 600)
	b.SetFocused(true)
	c := uitest.DrawWidget(b)

	ring := c.StrokeRoundRects[0]
	if want := alpha(tok.Destructive, 0.2); ring.Color != want {
		t.Errorf("destructive ring = %v, want destructive/20 %v", ring.Color, want)
	}
}

func TestButtonPointerFocusShowsNoRing(t *testing.T) {
	lightTokens(t)
	b := graft.Button("Save")
	uitest.LayoutWidget(b, 800, 600)
	uitest.SimulateClick(b, 2, 2) // pointer focus, not keyboard
	if !b.IsFocused() {
		t.Fatal("click should focus the button")
	}
	c := uitest.DrawWidget(b)
	if len(c.StrokeRoundRects) != 0 {
		t.Errorf("pointer focus drew %d strokes, want 0 (no focus-visible ring)", len(c.StrokeRoundRects))
	}
}

func TestButtonLinkUnderlineOnHover(t *testing.T) {
	tok := lightTokens(t)
	b := graft.Button("Learn more").Link()
	uitest.LayoutWidget(b, 800, 600)

	c := uitest.DrawWidget(b)
	if len(c.Lines) != 0 {
		t.Errorf("link at rest drew %d lines, want 0", len(c.Lines))
	}
	if c.StyledTexts[0].Style.Color != tok.Primary {
		t.Errorf("link text color = %v, want primary", c.StyledTexts[0].Style.Color)
	}

	b.Event(uitest.NewMockContext(), uitest.MouseEnter(2, 2))
	c = uitest.DrawWidget(b)
	if len(c.Lines) != 1 {
		t.Fatalf("hovered link drew %d lines, want 1 (underline)", len(c.Lines))
	}
	ul := c.Lines[0]
	if ul.StrokeWidth != 1 {
		t.Errorf("underline width = %v, want 1", ul.StrokeWidth)
	}
	if ul.To.X-ul.From.X != textmetrics.Width("Geist Medium", 14, "Learn more") {
		t.Errorf("underline span = %v, want text width", ul.To.X-ul.From.X)
	}
}

func TestButtonDisabled(t *testing.T) {
	tok := lightTokens(t)
	fired := false
	b := graft.Button("Save").Disabled(true).OnClick(func() { fired = true })
	uitest.LayoutWidget(b, 800, 600)

	if b.IsFocusable() {
		t.Error("disabled button must not be focusable")
	}
	uitest.SimulateClick(b, 2, 2)
	if fired {
		t.Error("disabled button fired OnClick")
	}
	c := uitest.DrawWidget(b)
	if want := mulAlpha(tok.Primary, 0.5); c.RoundRects[0].Color != want {
		t.Errorf("disabled fill = %v, want primary at 50%% %v", c.RoundRects[0].Color, want)
	}
}

func TestButtonActivation(t *testing.T) {
	lightTokens(t)
	clicks := 0
	b := graft.Button("Go").OnClick(func() { clicks++ })
	uitest.LayoutWidget(b, 800, 600)

	uitest.SimulateClick(b, 2, 2)
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}

	b.SetFocused(true)
	uitest.SimulateKeyPress(b, event.KeyEnter)
	uitest.SimulateKeyPress(b, event.KeySpace)
	if clicks != 3 {
		t.Errorf("clicks after Enter+Space = %d, want 3", clicks)
	}
}

// --- goldens ---

func TestGoldenButtonVariants(t *testing.T) {
	gtest.GoldenLightDark(t, "button-variants", func() widget.Widget {
		return primitives.HBox(
			graft.Button("Default"),
			graft.Button("Secondary").Secondary(),
			graft.Button("Destructive").Destructive(),
			graft.Button("Outline").Outline(),
			graft.Button("Ghost").Ghost(),
			graft.Button("Link").Link(),
		).Gap(8).Padding(24).CrossAlign(primitives.CrossAxisCenter)
	})
}

func TestGoldenButtonSizes(t *testing.T) {
	gtest.GoldenLightDark(t, "button-sizes", func() widget.Widget {
		return primitives.HBox(
			graft.Button("Extra Small").XS(),
			graft.Button("Small").Sm(),
			graft.Button("Default"),
			graft.Button("Large").Lg(),
			graft.Button("").IconOnly(icons.Search).XS(),
			graft.Button("").IconOnly(icons.Search).Sm(),
			graft.Button("").IconOnly(icons.Search),
			graft.Button("").IconOnly(icons.Search).Lg(),
		).Gap(8).Padding(24).CrossAlign(primitives.CrossAxisCenter)
	})
}

func TestGoldenButtonIcon(t *testing.T) {
	gtest.GoldenLightDark(t, "button-icon", func() widget.Widget {
		return primitives.HBox(
			graft.Button("Login with Email").Icon(icons.Check),
			graft.Button("Search").Outline().Icon(icons.Search),
			graft.Button("Tiny").XS().Icon(icons.X),
		).Gap(8).Padding(24).CrossAlign(primitives.CrossAxisCenter)
	})
}

func TestGoldenButtonDisabled(t *testing.T) {
	gtest.GoldenLightDark(t, "button-disabled", func() widget.Widget {
		return primitives.HBox(
			graft.Button("Default").Disabled(true),
			graft.Button("Destructive").Destructive().Disabled(true),
			graft.Button("Outline").Outline().Disabled(true),
			graft.Button("Secondary").Secondary().Disabled(true),
			graft.Button("Ghost").Ghost().Disabled(true),
		).Gap(8).Padding(24).CrossAlign(primitives.CrossAxisCenter)
	})
}

func TestGoldenButtonFocused(t *testing.T) {
	gtest.GoldenLightDark(t, "button-focused", func() widget.Widget {
		b := graft.Button("Focused")
		b.SetFocused(true)
		return primitives.HBox(b).Padding(24)
	})
}

func TestGoldenButtonHovered(t *testing.T) {
	gtest.GoldenLightDark(t, "button-hovered", func() widget.Widget {
		hovered := func(b *graft.ButtonWidget) *graft.ButtonWidget {
			b.Event(uitest.NewMockContext(), uitest.MouseEnter(2, 2))
			return b
		}
		return primitives.HBox(
			hovered(graft.Button("Default")),
			hovered(graft.Button("Secondary").Secondary()),
			hovered(graft.Button("Destructive").Destructive()),
			hovered(graft.Button("Outline").Outline()),
			hovered(graft.Button("Ghost").Ghost()),
			hovered(graft.Button("Link").Link()),
		).Gap(8).Padding(24).CrossAlign(primitives.CrossAxisCenter)
	})
}

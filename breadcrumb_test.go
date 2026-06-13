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

// breadcrumbForceLight pins the theme to light and returns active tokens plus a
// restore function.
func breadcrumbForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

// TestBreadcrumbSpecColors asserts links are MutedForeground, the page is
// Foreground, and both render at 14px.
func TestBreadcrumbSpecColors(t *testing.T) {
	tok, restore := breadcrumbForceLight(t)
	defer restore()

	bc := graft.Breadcrumb(
		graft.BreadcrumbLink("Home"),
		graft.BreadcrumbLink("Docs"),
		graft.BreadcrumbPage("Components"),
	)
	bc.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 100)))
	canvas := uitest.DrawWidget(bc)

	byText := map[string]*uitest.DrawStyledTextCall{}
	for idx := range canvas.StyledTexts {
		byText[canvas.StyledTexts[idx].Text] = &canvas.StyledTexts[idx]
	}
	for _, name := range []string{"Home", "Docs", "Components"} {
		if byText[name] == nil {
			t.Fatalf("missing breadcrumb text %q", name)
		}
		if byText[name].Style.FontSize != metrics.Breadcrumb.FontSize {
			t.Fatalf("%q font size: got %v want %v", name, byText[name].Style.FontSize, metrics.Breadcrumb.FontSize)
		}
	}
	if byText["Home"].Style.Color != tok.MutedForeground {
		t.Fatalf("link color: got %+v want MutedForeground", byText["Home"].Style.Color)
	}
	if byText["Components"].Style.Color != tok.Foreground {
		t.Fatalf("page color: got %+v want Foreground", byText["Components"].Style.Color)
	}
}

// TestBreadcrumbHoverHighlight asserts a hovered link switches to Foreground.
func TestBreadcrumbHoverHighlight(t *testing.T) {
	tok, restore := breadcrumbForceLight(t)
	defer restore()

	bc := graft.Breadcrumb(
		graft.BreadcrumbLink("Home"),
		graft.BreadcrumbPage("Now"),
	)
	bc.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 100)))

	// Hover over the first link.
	ctx := uitest.NewMockContext()
	bounds := bc.Bounds()
	bc.Event(ctx, uitest.MouseMove(bounds.Min.X+2, bounds.Min.Y+2))

	canvas := uitest.DrawWidget(bc)
	for idx := range canvas.StyledTexts {
		if canvas.StyledTexts[idx].Text == "Home" {
			if canvas.StyledTexts[idx].Style.Color != tok.Foreground {
				t.Fatalf("hovered link color: got %+v want Foreground", canvas.StyledTexts[idx].Style.Color)
			}
		}
	}
}

// TestBreadcrumbLinkClick fires the link's OnClick.
func TestBreadcrumbLinkClick(t *testing.T) {
	_, restore := breadcrumbForceLight(t)
	defer restore()

	clicked := false
	bc := graft.Breadcrumb(
		graft.BreadcrumbLink("Home").OnClick(func() { clicked = true }),
		graft.BreadcrumbPage("Now"),
	)
	bc.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 100)))
	bounds := bc.Bounds()
	uitest.SimulateClick(bc, bounds.Min.X+2, bounds.Min.Y+2)
	if !clicked {
		t.Fatal("breadcrumb link OnClick did not fire")
	}
}

// TestGoldenBreadcrumb renders a 3-level trail and one with an ellipsis.
func TestGoldenBreadcrumb(t *testing.T) {
	gtest.GoldenLightDark(t, "breadcrumb-basic", func() widget.Widget {
		return primitives.VBox(graft.Breadcrumb(
			graft.BreadcrumbLink("Home"),
			graft.BreadcrumbLink("Components"),
			graft.BreadcrumbPage("Breadcrumb"),
		)).Padding(12).CrossAlign(primitives.CrossAxisStart)
	})
	gtest.GoldenLightDark(t, "breadcrumb-ellipsis", func() widget.Widget {
		return primitives.VBox(graft.Breadcrumb(
			graft.BreadcrumbLink("Home"),
			graft.BreadcrumbEllipsis(),
			graft.BreadcrumbLink("Components"),
			graft.BreadcrumbPage("Breadcrumb"),
		)).Padding(12).CrossAlign(primitives.CrossAxisStart)
	})
}

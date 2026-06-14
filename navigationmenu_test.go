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
)

// navMenuDemo builds the canonical navigation bar: a "Getting started" trigger,
// a "Components" trigger, and a plain "Docs" link.
func navMenuDemo() *graft.NavigationMenuWidget {
	return graft.NavigationMenu(
		graft.NavigationMenuItem(
			graft.NavigationMenuTrigger("Getting started"),
			graft.NavigationMenuContent(
				graft.NavigationMenuLink("Introduction").Description("Re-usable components built using the gogpu/ui toolkit."),
				graft.NavigationMenuLink("Installation").Description("How to install dependencies and structure your app."),
				graft.NavigationMenuLink("Typography").Description("Styles for headings, paragraphs, lists, and more."),
			),
		),
		graft.NavigationMenuItem(
			graft.NavigationMenuTrigger("Components"),
			graft.NavigationMenuContent(
				graft.NavigationMenuLink("Alert Dialog").Description("A modal dialog that interrupts the user."),
				graft.NavigationMenuLink("Hover Card").Description("Preview content behind a link."),
				graft.NavigationMenuLink("Progress").Description("Displays an indicator for task completion."),
			),
		),
		graft.NavigationMenuLinkItem("Docs"),
	)
}

// navMenuContentDemo is the content panel used for the panel golden.
func navMenuContentDemo() *graft.NavigationMenuContentDef {
	return graft.NavigationMenuContent(
		graft.NavigationMenuLink("Introduction").Description("Re-usable components built using the gogpu/ui toolkit."),
		graft.NavigationMenuLink("Installation").Description("How to install dependencies and structure your app."),
		graft.NavigationMenuLink("Typography").Description("Styles for headings, paragraphs, lists, and more."),
	)
}

// TestNavigationMenuBarGeometry pins the bar height (h-9) and that it has no
// surface chrome (transparent root: no full-bounds background round-rect).
func TestNavigationMenuBarGeometry(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	bar := navMenuDemo()
	size := bar.Layout(uitest.NewMockContext(), looseConstraints())
	if size.Height != metrics.NavigationMenuHeight {
		t.Errorf("bar height = %v, want %v (h-9)", size.Height, metrics.NavigationMenuHeight)
	}

	bar.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	canvas := uitest.DrawWidget(bar)

	// The root list has no surface: there must be no round-rect covering the
	// full bar bounds (only per-item hover/open fills, which are absent here).
	for i := range canvas.RoundRects {
		if canvas.RoundRects[i].Bounds.Size() == size {
			t.Fatalf("navigation bar drew a full-bounds surface; root should be transparent")
		}
	}
}

// TestNavigationMenuTriggerLabels pins that every top-level label renders at the
// trigger font size.
func TestNavigationMenuTriggerLabels(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	bar := navMenuDemo()
	size := bar.Layout(uitest.NewMockContext(), looseConstraints())
	bar.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	canvas := uitest.DrawWidget(bar)

	want := map[string]bool{"Getting started": false, "Components": false, "Docs": false}
	for _, st := range canvas.StyledTexts {
		if _, ok := want[st.Text]; ok && st.Style.FontSize == metrics.NavigationMenuTriggerFontSize {
			want[st.Text] = true
		}
	}
	for label, seen := range want {
		if !seen {
			t.Errorf("top-level item %q not drawn at %vpx", label, metrics.NavigationMenuTriggerFontSize)
		}
	}
}

// TestNavigationMenuClickSwitching drives the open/close/switch state machine
// against a fake overlay manager.
func TestNavigationMenuClickSwitching(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	bar := navMenuDemo()
	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om
	size := bar.Layout(ctx, looseConstraints())
	bar.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))

	start := bar.Children()[0].(interface{ Bounds() geometry.Rect }).Bounds()
	comp := bar.Children()[1].(interface{ Bounds() geometry.Rect }).Bounds()

	// Click first trigger: one panel open.
	bar.Event(ctx, uitest.Click(start.Center().X, start.Center().Y))
	if om.liveCount() != 1 {
		t.Fatalf("first trigger click did not open a panel: live=%d", om.liveCount())
	}

	// Click second trigger: still one panel (switched, not stacked).
	bar.Event(ctx, uitest.Click(comp.Center().X, comp.Center().Y))
	if om.liveCount() != 1 {
		t.Fatalf("switching changed live count: live=%d", om.liveCount())
	}
	if len(om.pushed) != 2 {
		t.Fatalf("switch did not push a second panel: pushed=%d", len(om.pushed))
	}

	// Click second trigger again: closes.
	bar.Event(ctx, uitest.Click(comp.Center().X, comp.Center().Y))
	if om.liveCount() != 0 {
		t.Fatalf("re-click did not close the panel: live=%d", om.liveCount())
	}
}

// TestNavigationMenuHoverSwitch verifies that hovering an adjacent trigger while
// a panel is open switches to it (mirrors menubar hover-switch).
func TestNavigationMenuHoverSwitch(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	bar := navMenuDemo()
	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om
	size := bar.Layout(ctx, looseConstraints())
	bar.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))

	start := bar.Children()[0].(interface{ Bounds() geometry.Rect }).Bounds()
	comp := bar.Children()[1].(interface{ Bounds() geometry.Rect }).Bounds()

	// Open the first panel by click.
	bar.Event(ctx, uitest.Click(start.Center().X, start.Center().Y))
	if om.liveCount() != 1 {
		t.Fatalf("first trigger click did not open a panel: live=%d", om.liveCount())
	}

	// Hover the second trigger: should switch (still one live panel, two pushed).
	bar.Event(ctx, uitest.MouseMove(comp.Center().X, comp.Center().Y))
	if om.liveCount() != 1 {
		t.Fatalf("hover-switch changed live count: live=%d", om.liveCount())
	}
	if len(om.pushed) != 2 {
		t.Fatalf("hover over adjacent trigger did not switch panels: pushed=%d", len(om.pushed))
	}
}

// TestNavigationMenuLinkItemNoChevron verifies that a plain link item does not
// open an overlay (no content) when clicked, and a trigger does.
func TestNavigationMenuLinkItemNoChevron(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	clicked := false
	bar := graft.NavigationMenu(
		graft.NavigationMenuLinkItem("Docs", func() { clicked = true }),
	)
	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om
	size := bar.Layout(ctx, looseConstraints())
	bar.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))

	docs := bar.Children()[0].(interface{ Bounds() geometry.Rect }).Bounds()
	bar.Event(ctx, uitest.Click(docs.Center().X, docs.Center().Y))

	if !clicked {
		t.Errorf("link item onPress did not fire")
	}
	if om.liveCount() != 0 {
		t.Errorf("link item opened an overlay: live=%d", om.liveCount())
	}
}

// TestGoldenNavigationMenu renders the bar and one open content panel directly,
// light + dark.
func TestGoldenNavigationMenu(t *testing.T) {
	gtest.GoldenLightDark(t, "navigationmenu-bar", func() widget.Widget {
		// Pad the frame so any item hover/open fills and the bar are not clipped.
		return primitives.VBox(navMenuDemo()).Padding(16)
	})

	gtest.GoldenLightDark(t, "navigationmenu-content", func() widget.Widget {
		return graft.NavigationMenuMenuPreview(navMenuContentDemo())
	})
}

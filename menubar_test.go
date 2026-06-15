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

// menubarDemo builds the canonical 3-menu demo bar (File / Edit / View).
func menubarDemo() *graft.MenubarWidget {
	return graft.Menubar(
		graft.MenubarMenu("File", graft.MenubarContent(
			graft.MenubarItem("New Tab").Shortcut("T"),
			graft.MenubarItem("New Window").Shortcut("N"),
			graft.MenubarSeparator(),
			graft.MenubarItem("Print").Shortcut("P"),
		)),
		graft.MenubarMenu("Edit", graft.MenubarContent(
			graft.MenubarItem("Undo").Shortcut("Z"),
			graft.MenubarItem("Redo").Shortcut("Y"),
		)),
		graft.MenubarMenu("View", graft.MenubarContent(
			graft.MenubarCheckboxItem("Always Show Bookmarks Bar").Checked(true),
			graft.MenubarCheckboxItem("Always Show Full URLs"),
		)),
	)
}

// TestMenubarBarGeometry pins the bar surface: h-9 (36), rounded-md, bg
// Background, 1px Border.
func TestMenubarBarGeometry(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	tok := graft.CurrentTheme().Active()
	th := graft.CurrentTheme()

	bar := menubarDemo()
	size := bar.Layout(uitest.NewMockContext(), looseConstraints())
	if size.Height != metrics.MenubarHeight {
		t.Errorf("bar height = %v, want %v (h-9)", size.Height, metrics.MenubarHeight)
	}

	bar.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	canvas := uitest.DrawWidget(bar)

	bw := metrics.MenubarBorderWidth

	// BorderFill: outer Border round-rect at full bounds, inner Background
	// surface inset by the border width.
	var border *uitest.DrawRoundRectCall
	for i := range canvas.RoundRects {
		if canvas.RoundRects[i].Bounds.Size() == size && canvas.RoundRects[i].Color == tok.Border {
			border = &canvas.RoundRects[i]
			break
		}
	}
	if border == nil {
		t.Fatalf("no full-bounds Border bar round-rect at size %v", size)
	}
	if border.Radius != th.RadiusMD() {
		t.Errorf("bar border radius = %v, want MD %v", border.Radius, th.RadiusMD())
	}

	var surface *uitest.DrawRoundRectCall
	innerSize := geometry.Sz(size.Width-2*bw, size.Height-2*bw)
	for i := range canvas.RoundRects {
		if canvas.RoundRects[i].Bounds.Size() == innerSize && canvas.RoundRects[i].Color == tok.Background {
			surface = &canvas.RoundRects[i]
			break
		}
	}
	if surface == nil {
		t.Fatalf("no inset Background bar surface at size %v", innerSize)
	}
	if surface.Radius != th.RadiusMD()-bw {
		t.Errorf("bar surface radius = %v, want %v", surface.Radius, th.RadiusMD()-bw)
	}

	// No 1px border stroke any more (border is now a fill).
	for _, s := range canvas.StrokeRoundRects {
		if s.StrokeWidth == metrics.MenubarBorderWidth {
			t.Fatalf("unexpected 1px border stroke: %+v", s)
		}
	}
}

// TestMenubarTriggerLabels pins the three trigger labels at the trigger font
// size.
func TestMenubarTriggerLabels(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	bar := menubarDemo()
	size := bar.Layout(uitest.NewMockContext(), looseConstraints())
	bar.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	canvas := uitest.DrawWidget(bar)

	want := map[string]bool{"File": false, "Edit": false, "View": false}
	for _, st := range canvas.StyledTexts {
		if _, ok := want[st.Text]; ok && st.Style.FontSize == metrics.MenubarTriggerFontSize {
			want[st.Text] = true
		}
	}
	for label, seen := range want {
		if !seen {
			t.Errorf("trigger %q not drawn at %vpx", label, metrics.MenubarTriggerFontSize)
		}
	}
}

// TestMenubarClickSwitching drives the open/close/switch state machine against
// a fake overlay manager.
func TestMenubarClickSwitching(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	bar := menubarDemo()
	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om
	size := bar.Layout(ctx, looseConstraints())
	bar.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))

	// Find the File and Edit trigger centers from their bounds.
	file := bar.Children()[0].(interface{ Bounds() geometry.Rect }).Bounds()
	edit := bar.Children()[1].(interface{ Bounds() geometry.Rect }).Bounds()

	// Click File: one panel open.
	bar.Event(ctx, uitest.Click(file.Center().X, file.Center().Y))
	if om.liveCount() != 1 {
		t.Fatalf("File click did not open a menu: live=%d", om.liveCount())
	}

	// Click Edit while File open: still one panel (switched, not stacked).
	bar.Event(ctx, uitest.Click(edit.Center().X, edit.Center().Y))
	if om.liveCount() != 1 {
		t.Fatalf("switching menus changed live count: live=%d pushed=%d removed=%d",
			om.liveCount(), len(om.pushed), len(om.removed))
	}
	if len(om.pushed) != 2 {
		t.Fatalf("switch did not push a second panel: pushed=%d", len(om.pushed))
	}

	// Click Edit again: closes (no panel open).
	bar.Event(ctx, uitest.Click(edit.Center().X, edit.Center().Y))
	if om.liveCount() != 0 {
		t.Fatalf("re-click did not close the menu: live=%d", om.liveCount())
	}
}

// TestGoldenMenubar renders the bar (3 menus) and one open menu panel directly,
// light + dark.
func TestGoldenMenubar(t *testing.T) {
	gtest.GoldenLightDark(t, "menubar-bar", func() widget.Widget {
		// Pad the frame so the bar's shadow-xs is captured, not clipped.
		return primitives.VBox(menubarDemo()).Padding(16)
	})

	gtest.GoldenLightDark(t, "menubar-menu", func() widget.Widget {
		return graft.MenubarMenuPreview(graft.MenubarContent(
			graft.MenubarItem("New Tab").Shortcut("T"),
			graft.MenubarItem("New Window").Shortcut("N"),
			graft.MenubarSeparator(),
			graft.MenubarItem("Share"),
			graft.MenubarSeparator(),
			graft.MenubarItem("Print").Shortcut("P"),
		))
	})
}

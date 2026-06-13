package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
)

// contextMenuContent builds the canonical demo content used across tests:
// items + separator + checkbox + a destructive item.
func contextMenuContent() *graft.ContextMenuContentWidget {
	return graft.ContextMenuContent(
		graft.ContextMenuLabel("Actions"),
		graft.ContextMenuSeparator(),
		graft.ContextMenuItem("Back").Shortcut("⌘["),
		graft.ContextMenuItem("Forward").Shortcut("⌘]"),
		graft.ContextMenuCheckboxItem("Show Bookmarks").Checked(true),
		graft.ContextMenuSeparator(),
		graft.ContextMenuItem("Delete").Destructive().Shortcut("⌫"),
	)
}

// TestContextMenuRightClickOpens verifies a right-click inside the target
// pushes the menu overlay at the cursor; a left-click does not.
func TestContextMenuRightClickOpens(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	cm := graft.ContextMenu(primitives.Box().Width(200).Height(120), contextMenuContent())
	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om
	cm.Layout(ctx, looseConstraints())

	// A left click does not open the menu.
	cm.Event(ctx, uitest.Click(40, 40))
	if om.liveCount() != 0 {
		t.Fatalf("left click opened the context menu: live=%d", om.liveCount())
	}

	// A right click opens it.
	cm.Event(ctx, uitest.RightClick(40, 40))
	if om.liveCount() != 1 {
		t.Fatalf("right click did not open the context menu: live=%d", om.liveCount())
	}
}

// TestContextMenuPlacesAtCursor pins the panel anchoring: the top-left sits at
// the cursor plus the small gap when there is room.
func TestContextMenuPlacesAtCursor(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	cm := graft.ContextMenu(primitives.Box().Width(400).Height(300), contextMenuContent())
	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om
	ctx.WindowSizeVal = geometry.Sz(1000, 800)
	cm.Layout(ctx, looseConstraints())

	cm.Event(ctx, uitest.RightClick(120, 90))
	if len(om.pushed) != 1 {
		t.Fatalf("panel not pushed")
	}
	// The overlay wrapper adopts the pre-positioned panel bounds during Layout
	// (as the real overlay system does); its top-left should be near the cursor.
	ov := om.pushed[0]
	ov.Layout(ctx, geometry.Loose(ctx.WindowSizeVal))
	got := ov.(interface{ Bounds() geometry.Rect }).Bounds().Min
	if got.X < 120 || got.Y < 90 {
		t.Errorf("panel origin %v should be at/after cursor (120,90)", got)
	}
}

// TestGoldenContextMenu renders the demo panel directly, light + dark.
func TestGoldenContextMenu(t *testing.T) {
	gtest.GoldenLightDark(t, "context-menu", func() widget.Widget {
		return graft.ContextMenuPreview(contextMenuContent())
	})
}

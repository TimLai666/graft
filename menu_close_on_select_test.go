package graft_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"

	"github.com/TimLai666/graft"
)

// TestDropdownMenuItemSelectCloses verifies that activating a DropdownMenuItem
// fires its OnSelect AND closes the menu (shadcn/Radix dismiss-on-select).
func TestDropdownMenuItemSelectCloses(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	selected := ""
	open := state.NewSignal(false)
	d := graft.DropdownMenu(
		graft.DropdownMenuTrigger(graft.Text("Open")),
		graft.DropdownMenuContent(
			graft.DropdownMenuItem("Profile").OnSelect(func() { selected = "Profile" }),
			graft.DropdownMenuItem("Settings").OnSelect(func() { selected = "Settings" }),
		),
	).Bind(open)

	om := &recordingOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om

	size := d.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))
	d.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))

	// Open the menu.
	d.Event(ctx, uitest.Click(size.Width/2, size.Height/2))
	if om.top == nil || !open.Get() {
		t.Fatal("trigger click did not open the menu")
	}

	// Navigate to the second item and press Enter via the pushed overlay.
	om.top.Event(ctx, uitest.KeyPress(event.KeyDown, event.ModNone))
	om.top.Event(ctx, uitest.KeyPress(event.KeyEnter, event.ModNone))

	if selected != "Settings" {
		t.Errorf("OnSelect fired for %q, want Settings", selected)
	}
	if open.Get() {
		t.Error("menu did not close after selecting an item")
	}
	if om.top != nil {
		t.Error("overlay not removed after selecting an item")
	}
}

// TestContextMenuItemSelectCloses verifies a ContextMenu closes on item select.
func TestContextMenuItemSelectCloses(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	picked := ""
	cm := graft.ContextMenu(
		graft.Text("Right-click me"),
		graft.ContextMenuContent(
			graft.ContextMenuItem("Back").OnSelect(func() { picked = "Back" }),
			graft.ContextMenuItem("Reload").OnSelect(func() { picked = "Reload" }),
		),
	)

	om := &recordingOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om

	size := cm.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))
	cm.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))

	// Right-click opens the menu at the cursor.
	cm.Event(ctx, uitest.RightClick(size.Width/2, size.Height/2))
	if om.top == nil {
		t.Fatal("right-click did not open the context menu")
	}

	// Activate the first item via Enter.
	om.top.Event(ctx, uitest.KeyPress(event.KeyEnter, event.ModNone))
	if picked != "Back" {
		t.Errorf("OnSelect fired for %q, want Back", picked)
	}
	if om.top != nil {
		t.Error("context menu did not close after selecting an item")
	}
}

// TestMenubarItemSelectCloses verifies a Menubar menu closes on item select.
func TestMenubarItemSelectCloses(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	fired := ""
	mb := graft.Menubar(
		graft.MenubarMenu("File", graft.MenubarContent(
			graft.MenubarItem("New").OnSelect(func() { fired = "New" }),
			graft.MenubarItem("Open").OnSelect(func() { fired = "Open" }),
		)),
	)

	om := &recordingOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om

	size := mb.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))
	mb.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	// Re-layout so the trigger bounds are stamped under the bar.
	mb.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))

	// Click the "File" trigger (top-left within the bar padding).
	mb.Event(ctx, uitest.Click(20, size.Height/2))
	if om.top == nil {
		t.Fatal("menubar trigger click did not open the menu")
	}

	om.top.Event(ctx, uitest.KeyPress(event.KeyEnter, event.ModNone))
	if fired != "New" {
		t.Errorf("OnSelect fired for %q, want New", fired)
	}
	if om.top != nil {
		t.Error("menubar menu did not close after selecting an item")
	}
}

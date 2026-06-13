package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// dropdownMenuContent builds the canonical demo content used across tests.
func dropdownMenuContent() *graft.DropdownMenuContentWidget {
	return graft.DropdownMenuContent(
		graft.DropdownMenuLabel("My Account"),
		graft.DropdownMenuSeparator(),
		graft.DropdownMenuItem("Profile").Icon(icons.Info).Shortcut("⇧⌘P"),
		graft.DropdownMenuItem("Settings").Icon(icons.Search).Shortcut("⌘,"),
		graft.DropdownMenuCheckboxItem("Show toolbar").Checked(true),
		graft.DropdownMenuSeparator(),
		graft.DropdownMenuItem("Log out").Destructive().Shortcut("⇧⌘Q"),
	)
}

// TestDropdownMenuOpenSignal verifies the open signal toggles when the
// trigger is clicked. The MockContext OverlayManager is nil, so the overlay
// is not pushed (noted gap); the open state still flips, which is what the
// signal-bound consumer observes.
func TestDropdownMenuOpenSignal(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	open := state.NewSignal(false)
	d := graft.DropdownMenu(
		graft.DropdownMenuTrigger(graft.Text("Open")),
		dropdownMenuContent(),
	).Bind(open)

	ctx := uitest.NewMockContext()
	size := d.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))
	d.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))

	// Press inside the trigger toggles open to true.
	d.Event(ctx, uitest.Click(size.Width/2, size.Height/2))
	if !open.Get() {
		t.Fatal("trigger click did not open the menu")
	}
	// Press again toggles closed.
	d.Event(ctx, uitest.Click(size.Width/2, size.Height/2))
	if open.Get() {
		t.Fatal("second trigger click did not close the menu")
	}
}

// TestDropdownMenuOnOpenChange checks the observer fires.
func TestDropdownMenuOnOpenChange(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	var changes []bool
	d := graft.DropdownMenu(
		graft.DropdownMenuTrigger(graft.Text("Open")),
		dropdownMenuContent(),
	).OnOpenChange(func(v bool) { changes = append(changes, v) })

	ctx := uitest.NewMockContext()
	size := d.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))
	d.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	d.Event(ctx, uitest.Click(size.Width/2, size.Height/2))

	if len(changes) != 1 || changes[0] != true {
		t.Fatalf("OnOpenChange = %v, want [true]", changes)
	}
}

// TestDropdownMenuPreviewGeometry pins the panel geometry rendered as direct
// content (the path goldens use, and the path that works without an
// OverlayManager).
func TestDropdownMenuPreviewGeometry(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	tok := th.Active()
	preview := graft.DropdownMenuPreview(dropdownMenuContent())

	preview.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(800, 800)))
	c := uitest.DrawWidget(preview)

	// The popover surface is drawn.
	foundSurface := false
	for _, rr := range c.RoundRects {
		if rr.Color == tok.Popover {
			foundSurface = true
		}
	}
	if !foundSurface {
		t.Error("dropdown panel popover surface not drawn")
	}

	// Two separators → two border-colored 1px rects spanning the panel.
	sepCount := 0
	for _, r := range c.Rects {
		if r.Color == tok.Border && r.Bounds.Height() == 1 {
			sepCount++
		}
	}
	if sepCount != 2 {
		t.Errorf("separator count = %d, want 2", sepCount)
	}

	// Text rows: label + 3 items + checkbox + destructive item = 6 labels,
	// plus shortcuts for 3 items = 9 styled texts total.
	if len(c.StyledTexts) < 6 {
		t.Errorf("styled text count = %d, want >= 6 (label+items)", len(c.StyledTexts))
	}
}

// TestDropdownMenuFontSizes pins item vs shortcut font sizes.
func TestDropdownMenuFontSizes(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	preview := graft.DropdownMenuPreview(graft.DropdownMenuContent(
		graft.DropdownMenuItem("Profile").Shortcut("⌘P"),
	))
	preview.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	c := uitest.DrawWidget(preview)

	var sawItem, sawShortcut bool
	for _, st := range c.StyledTexts {
		if st.Text == "Profile" && st.Style.FontSize == metrics.Menu.FontSize {
			sawItem = true
		}
		if st.Text == "⌘P" && st.Style.FontSize == metrics.Menu.ShortcutFontSize {
			sawShortcut = true
		}
	}
	if !sawItem {
		t.Errorf("item label not drawn at %vpx", metrics.Menu.FontSize)
	}
	if !sawShortcut {
		t.Errorf("shortcut not drawn at %vpx", metrics.Menu.ShortcutFontSize)
	}
}

// TestDropdownMenuGoldens renders the demo panel directly, light + dark.
func TestDropdownMenuGoldens(t *testing.T) {
	gtest.GoldenLightDark(t, "dropdown-menu", func() widget.Widget {
		return graft.DropdownMenuPreview(dropdownMenuContent())
	})

	gtest.GoldenLightDark(t, "dropdown-menu-radio", func() widget.Widget {
		sig := state.NewSignal("comfortable")
		return graft.DropdownMenuPreview(graft.DropdownMenuContent(
			graft.DropdownMenuLabel("Density"),
			graft.DropdownMenuSeparator(),
			graft.DropdownMenuRadioGroup(sig,
				graft.DropdownMenuRadioItem("compact", "Compact"),
				graft.DropdownMenuRadioItem("comfortable", "Comfortable"),
				graft.DropdownMenuRadioItem("spacious", "Spacious"),
			),
		))
	})
}

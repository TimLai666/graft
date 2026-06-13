package graft_test

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// demoCommandItems returns the canonical demo groups used in tests.
func demoCommandItems() (*graft.CommandInputDef, *graft.CommandListDef) {
	input := graft.CommandInput().Placeholder("Type a command or search...")
	list := graft.CommandList(
		graft.CommandGroup("Suggestions",
			graft.CommandItem("Calendar").Icon(icons.Calendar),
			graft.CommandItem("Search").Icon(icons.Search).Shortcut("Ctrl K"),
			graft.CommandItem("Settings").Icon(icons.GripVertical),
		),
		graft.CommandSeparator(),
		graft.CommandGroup("Settings",
			graft.CommandItem("Profile").Icon(icons.Circle),
			graft.CommandItem("Billing").Icon(icons.ChevronsUpDown),
			graft.CommandItem("Notifications"),
		),
	)
	return input, list
}

// TestCommandLayout verifies the command surface sizes to the dialog width.
func TestCommandLayout(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	input, list := demoCommandItems()
	cmd := graft.Command(input, list)
	size := uitest.LayoutWidget(cmd, 900, 900)
	if size.Width != metrics.Command.DialogWidth {
		t.Errorf("command width = %v, want %v", size.Width, metrics.Command.DialogWidth)
	}
	if size.Height <= metrics.Command.InputHeight {
		t.Errorf("command height = %v, want > input height %v", size.Height, metrics.Command.InputHeight)
	}
}

// TestCommandFilter verifies case-insensitive filtering.
func TestCommandFilter(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	input, list := demoCommandItems()
	cmd := graft.Command(input, list)

	// Filter "cal" -> Calendar visible, Billing hidden.
	preview := graft.CommandPreview(cmd, "cal", 0, graft.CurrentTheme())
	uitest.LayoutWidget(preview, 900, 900)
	mc := uitest.DrawWidget(preview)
	if !drewAnyText(mc, "Calendar") {
		t.Errorf("filter 'cal' did not show Calendar")
	}
	if drewAnyText(mc, "Billing") {
		t.Errorf("filter 'cal' should hide Billing")
	}
}

// TestCommandFilterEmpty verifies the empty state when no items match.
func TestCommandFilterEmpty(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	input, list := demoCommandItems()
	cmd := graft.Command(input, list)
	preview := graft.CommandPreview(cmd, "zzz", 0, graft.CurrentTheme())
	uitest.LayoutWidget(preview, 900, 900)
	mc := uitest.DrawWidget(preview)
	if !drewAnyText(mc, metrics.Command.EmptyText) {
		t.Errorf("no-match did not render empty text %q", metrics.Command.EmptyText)
	}
}

// TestCommandGroupHeadings verifies group headings appear.
func TestCommandGroupHeadings(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	input, list := demoCommandItems()
	cmd := graft.Command(input, list)
	preview := graft.CommandPreview(cmd, "", 0, graft.CurrentTheme())
	uitest.LayoutWidget(preview, 900, 900)
	mc := uitest.DrawWidget(preview)
	if !drewAnyText(mc, "Suggestions") {
		t.Error("group heading 'Suggestions' not drawn")
	}
	if !drewAnyText(mc, "Settings") {
		t.Error("group heading 'Settings' not drawn")
	}
}

// TestCommandShortcut verifies shortcut text renders.
func TestCommandShortcut(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	input, list := demoCommandItems()
	cmd := graft.Command(input, list)
	preview := graft.CommandPreview(cmd, "", 0, graft.CurrentTheme())
	uitest.LayoutWidget(preview, 900, 900)
	mc := uitest.DrawWidget(preview)
	if !drewAnyText(mc, "Ctrl K") {
		t.Error("shortcut 'Ctrl K' not drawn")
	}
}

// TestCommandPlaceholder verifies the placeholder text renders when query is empty.
func TestCommandPlaceholder(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	input, list := demoCommandItems()
	cmd := graft.Command(input, list)
	preview := graft.CommandPreview(cmd, "", 0, graft.CurrentTheme())
	uitest.LayoutWidget(preview, 900, 900)
	mc := uitest.DrawWidget(preview)
	if !drewAnyText(mc, "Type a command or search...") {
		t.Error("placeholder text not drawn")
	}
}

// TestCommandDialogOverlay verifies CommandDialog pushes a modal overlay
// when the bound signal opens, following the dialog host pattern.
func TestCommandDialogOverlay(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	open := state.NewSignal(false)
	input, list := demoCommandItems()
	dlg := graft.CommandDialog(input, list).Bind(open)

	ctx := uitest.NewMockContext()
	om := newFakeOverlayManager()
	ctx.OverlayVal = om

	// Layout with closed signal -- no overlay.
	dlg.Layout(ctx, looseConstraints())
	if len(om.pushed) != 0 {
		t.Fatal("command dialog pushed overlay while closed")
	}

	// Open the signal, re-layout (next frame) -- overlay pushed.
	open.Set(true)
	dlg.Layout(ctx, looseConstraints())
	if len(om.pushed) != 1 {
		t.Fatalf("open pushed %d overlays, want 1", len(om.pushed))
	}

	// Close the signal, re-layout -- overlay removed.
	open.Set(false)
	dlg.Layout(ctx, looseConstraints())
	if len(om.pushed) != 0 {
		t.Fatalf("close left %d overlays, want 0", len(om.pushed))
	}
}

// TestGoldenCommand runs golden image tests for the command palette.
func TestGoldenCommand(t *testing.T) {
	// Full command palette (no filter, no hover).
	gtest.GoldenLightDark(t, "command-open", func() widget.Widget {
		input, list := demoCommandItems()
		cmd := graft.Command(input, list)
		preview := graft.CommandPreview(cmd, "", 0, graft.CurrentTheme())
		return primitives.Box(preview).Padding(24)
	})

	// Command palette with first item hovered.
	gtest.GoldenLightDark(t, "command-hovered", func() widget.Widget {
		input, list := demoCommandItems()
		cmd := graft.Command(input, list)
		preview := graft.CommandPreview(cmd, "", 1, graft.CurrentTheme())
		return primitives.Box(preview).Padding(24)
	})

	// Command palette with filter applied.
	gtest.GoldenLightDark(t, "command-filtered", func() widget.Widget {
		input, list := demoCommandItems()
		cmd := graft.Command(input, list)
		preview := graft.CommandPreview(cmd, "cal", 0, graft.CurrentTheme())
		return primitives.Box(preview).Padding(24)
	})
}

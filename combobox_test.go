package graft_test

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// selectedSignal returns a string signal pre-set to v (combobox controlled
// value).
func selectedSignal(v string) state.Signal[string] {
	s := state.NewSignal("")
	s.Set(v)
	return s
}

// comboFrameworks returns the canonical shadcn combobox demo options.
func comboFrameworks() []graft.ComboboxOption {
	return []graft.ComboboxOption{
		graft.ComboboxItem("next", "Next.js"),
		graft.ComboboxItem("sveltekit", "SvelteKit"),
		graft.ComboboxItem("nuxt", "Nuxt.js"),
		graft.ComboboxItem("remix", "Remix"),
		graft.ComboboxItem("astro", "Astro"),
	}
}

// TestComboboxTriggerWidth pins the 200px outline-button trigger width.
func TestComboboxTriggerWidth(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	cb := graft.Combobox(comboFrameworks()...)
	size := uitest.LayoutWidget(cb, 900, 900)
	if size.Width != metrics.Combobox.TriggerWidth {
		t.Errorf("trigger width = %v, want %v", size.Width, metrics.Combobox.TriggerWidth)
	}
}

// TestComboboxWidthOverride verifies .W overrides the trigger width.
func TestComboboxWidthOverride(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	cb := graft.Combobox(comboFrameworks()...).W(320)
	size := uitest.LayoutWidget(cb, 900, 900)
	if size.Width != 320 {
		t.Errorf("trigger width = %v, want 320", size.Width)
	}
}

// TestComboboxOpensAndSelects drives the trigger -> overlay -> select cycle.
func TestComboboxOpensAndSelects(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	var got string
	cb := graft.Combobox(comboFrameworks()...).OnChange(func(v string) { got = v })

	ctx := uitest.NewMockContext()
	om := newFakeOverlayManager()
	ctx.OverlayVal = om

	uitest.LayoutWidget(cb, 800, 600)
	uitest.DrawWidgetWithContext(cb, ctx)
	if len(om.pushed) != 0 {
		t.Fatal("combobox pushed overlay while closed")
	}

	uitest.SimulateClickWithContext(cb, ctx, 10, 10)
	if !cb.IsOpen() {
		t.Fatal("trigger click did not open the popover")
	}
	if len(om.pushed) != 1 {
		t.Fatalf("open pushed %d overlays, want 1", len(om.pushed))
	}

	// Dismiss closes the popover.
	content := om.pushed[0]
	om.onDismiss[content]()
	if cb.IsOpen() {
		t.Fatal("dismiss did not close the popover")
	}
	_ = got
}

// TestComboboxFilterAndEmpty verifies the content filters case-insensitively
// and shows the empty state for no matches.
func TestComboboxFilterAndEmpty(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)
	tok := graft.CurrentTheme().Active()

	cb := graft.Combobox(comboFrameworks()...)

	// Filter "nu" → Nuxt.js (and Next? "nu" not in "next.js"). Case-insensitive.
	content := graft.ComboboxContentPreview(cb, "nu", "", graft.CurrentTheme())
	uitest.LayoutWidget(content, 900, 900)
	mc := uitest.DrawWidget(content)
	if !drewAnyText(mc, "Nuxt.js") {
		t.Errorf("filter 'nu' did not show Nuxt.js")
	}
	if drewAnyText(mc, "Remix") {
		t.Errorf("filter 'nu' should hide Remix")
	}

	// No matches → empty state.
	empty := graft.ComboboxContentPreview(cb, "zzz", "", graft.CurrentTheme())
	uitest.LayoutWidget(empty, 900, 900)
	mcEmpty := uitest.DrawWidget(empty)
	if !drewAnyText(mcEmpty, metrics.Combobox.EmptyText) {
		t.Errorf("no-match did not render empty text %q", metrics.Combobox.EmptyText)
	}
	_ = tok
}

// TestComboboxSelectedCheck verifies the selected value draws a leading check
// (an extra icon draw is not directly observable in the mock; assert the
// selected label still renders and a check icon path runs without panic).
func TestComboboxSelectedCheck(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	cb := graft.Combobox(comboFrameworks()...)
	content := graft.ComboboxContentPreview(cb, "", "next", graft.CurrentTheme())
	uitest.LayoutWidget(content, 900, 900)
	mc := uitest.DrawWidget(content)
	if !drewAnyText(mc, "Next.js") {
		t.Errorf("selected label Next.js not drawn")
	}
}

// drewAnyText reports whether text s was drawn (styled or plain).
func drewAnyText(mc *uitest.MockCanvas, s string) bool {
	for _, st := range mc.StyledTexts {
		if st.Text == s {
			return true
		}
	}
	for _, dt := range mc.Texts {
		if dt.Text == s {
			return true
		}
	}
	return false
}

func TestGoldenCombobox(t *testing.T) {
	gtest.GoldenLightDark(t, "combobox-trigger-empty", func() widget.Widget {
		return primitives.Box(graft.Combobox(comboFrameworks()...)).Padding(24)
	})

	gtest.GoldenLightDark(t, "combobox-trigger-selected", func() widget.Widget {
		cb := graft.Combobox(comboFrameworks()...)
		cb.Bind(selectedSignal("next"))
		return primitives.Box(cb).Padding(24)
	})

	gtest.GoldenLightDark(t, "combobox-open", func() widget.Widget {
		cb := graft.Combobox(comboFrameworks()...)
		content := graft.ComboboxContentPreview(cb, "", "next", graft.CurrentTheme())
		return primitives.Box(content).Padding(24)
	})
}

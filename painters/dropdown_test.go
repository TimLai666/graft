package painters

import (
	"testing"

	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"

	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

func dropdownLightTheme() *theme.Theme {
	th := theme.New()
	th.SetMode(theme.ModeLight)
	return th
}

// TestDropdownPaintTriggerDefault pins the trigger border (1px --input) and
// placeholder color for the raw core/dropdown painter path.
func TestDropdownPaintTriggerDefault(t *testing.T) {
	th := dropdownLightTheme()
	tok := th.Active()
	p := Dropdown{Theme: th}
	canvas := &uitest.MockCanvas{}

	p.PaintTrigger(canvas, &dropdown.TriggerPaintState{
		Bounds:        geometry.NewRect(0, 0, 200, metrics.Select.TriggerHeight),
		SelectedText:  "Pick",
		IsPlaceholder: true,
	})

	if len(canvas.StrokeRoundRects) != 1 {
		t.Fatalf("expected 1 border stroke, got %d", len(canvas.StrokeRoundRects))
	}
	b := canvas.StrokeRoundRects[0]
	if b.StrokeWidth != metrics.Select.BorderWidth {
		t.Errorf("border width = %v, want %v", b.StrokeWidth, metrics.Select.BorderWidth)
	}
	if b.Color != tok.Input {
		t.Errorf("border color = %v, want input %v", b.Color, tok.Input)
	}
	if len(canvas.StyledTexts) != 1 || canvas.StyledTexts[0].Style.Color != tok.MutedForeground {
		t.Errorf("placeholder text not drawn in muted-foreground; texts=%+v", canvas.StyledTexts)
	}
}

// TestDropdownPaintMenu pins the menu surface (popover bg + border) and item
// highlight/selection drawing for the raw painter path.
func TestDropdownPaintMenu(t *testing.T) {
	th := dropdownLightTheme()
	tok := th.Active()
	p := Dropdown{Theme: th}
	canvas := &uitest.MockCanvas{}

	items := []dropdown.ItemDef{
		{Value: "a", Label: "Apple"},
		{Value: "b", Label: "Banana"},
	}
	p.PaintMenu(canvas, &dropdown.MenuPaintState{
		Bounds:           geometry.NewRect(0, 0, 200, 2*metrics.Select.ItemHeight+2*metrics.Select.ContentPad),
		Items:            items,
		HighlightedIndex: 0,
		SelectedIndex:    1,
		VisibleCount:     2,
		ItemHeight:       metrics.Select.ItemHeight,
	})

	// Popover surface fill present.
	foundPopover := false
	foundAccent := false
	for _, rr := range canvas.RoundRects {
		if rr.Color == tok.Popover {
			foundPopover = true
		}
		if rr.Color == tok.Accent {
			foundAccent = true
		}
	}
	if !foundPopover {
		t.Error("menu popover surface not drawn")
	}
	if !foundAccent {
		t.Error("highlighted item accent fill not drawn")
	}
	// Menu border stroke in --border.
	foundBorder := false
	for _, sr := range canvas.StrokeRoundRects {
		if sr.Color == tok.Border && sr.StrokeWidth == metrics.Select.BorderWidth {
			foundBorder = true
		}
	}
	if !foundBorder {
		t.Error("menu border not drawn in --border")
	}
	// Two item labels.
	if len(canvas.StyledTexts) != 2 {
		t.Errorf("item label count = %d, want 2", len(canvas.StyledTexts))
	}
}

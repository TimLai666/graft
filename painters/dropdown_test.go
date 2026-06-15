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

	// BorderFill: outer round-rect = --input border, no stroke.
	if len(canvas.StrokeRoundRects) != 0 {
		t.Fatalf("expected 0 border strokes (border now a fill), got %d", len(canvas.StrokeRoundRects))
	}
	var foundBorder bool
	for _, rr := range canvas.RoundRects {
		if rr.Color == tok.Input {
			foundBorder = true
		}
	}
	if !foundBorder {
		t.Errorf("--input border round-rect not found; round-rects=%+v", canvas.RoundRects)
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
	// Menu border is now the outer BorderFill round-rect in --border.
	foundBorder := false
	for _, rr := range canvas.RoundRects {
		if rr.Color == tok.Border {
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

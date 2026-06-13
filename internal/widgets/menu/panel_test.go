package menu

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

func testTheme() *theme.Theme {
	th := theme.New()
	th.SetMode(theme.ModeLight)
	return th
}

func TestPanelContentSizeMinWidth(t *testing.T) {
	p := NewPanel(testTheme(), NewItem("Cut"), NewItem("Copy"))
	size := p.ContentSize()
	if size.Width < metrics.Menu.MinWidth {
		t.Errorf("panel width %v below min-width %v", size.Width, metrics.Menu.MinWidth)
	}
	// Height = 2*pad + 2 item rows.
	wantH := 2*metrics.Menu.Pad + 2*metrics.Menu.ItemHeight
	if size.Height != wantH {
		t.Errorf("panel height = %v, want %v", size.Height, wantH)
	}
}

func TestPanelSeparatorHeight(t *testing.T) {
	p := NewPanel(testTheme(), NewItem("A"), NewSeparator(), NewItem("B"))
	wantH := 2*metrics.Menu.Pad + 2*metrics.Menu.ItemHeight + metrics.Menu.SeparatorHeight
	if got := p.ContentSize().Height; got != wantH {
		t.Errorf("height with separator = %v, want %v", got, wantH)
	}
}

func TestPanelInitialHighlightSkipsLabelAndSeparator(t *testing.T) {
	p := NewPanel(testTheme(), NewLabel("Section"), NewSeparator(), NewItem("First"))
	if p.Highlighted() != 2 {
		t.Errorf("initial highlight = %d, want 2 (first selectable item)", p.Highlighted())
	}
}

func TestPanelKeyboardNavSkipsDisabled(t *testing.T) {
	p := NewPanel(testTheme(),
		NewItem("A"),
		NewItem("B").SetDisabled(true),
		NewItem("C"),
	)
	if p.Highlighted() != 0 {
		t.Fatalf("initial highlight = %d, want 0", p.Highlighted())
	}
	// Down should skip the disabled B and land on C (index 2).
	p.Event(uitest.NewMockContext(), uitest.KeyPress(event.KeyDown, event.ModNone))
	if p.Highlighted() != 2 {
		t.Errorf("after Down: highlight = %d, want 2 (skip disabled)", p.Highlighted())
	}
	// Up returns to A.
	p.Event(uitest.NewMockContext(), uitest.KeyPress(event.KeyUp, event.ModNone))
	if p.Highlighted() != 0 {
		t.Errorf("after Up: highlight = %d, want 0", p.Highlighted())
	}
}

func TestPanelHomeEnd(t *testing.T) {
	p := NewPanel(testTheme(), NewItem("A"), NewItem("B"), NewItem("C"))
	p.Event(uitest.NewMockContext(), uitest.KeyPress(event.KeyEnd, event.ModNone))
	if p.Highlighted() != 2 {
		t.Errorf("End: highlight = %d, want 2", p.Highlighted())
	}
	p.Event(uitest.NewMockContext(), uitest.KeyPress(event.KeyHome, event.ModNone))
	if p.Highlighted() != 0 {
		t.Errorf("Home: highlight = %d, want 0", p.Highlighted())
	}
}

func TestPanelEnterSelects(t *testing.T) {
	var selected string
	p := NewPanel(testTheme(),
		NewItem("A").OnSelect(func() { selected = "A" }),
		NewItem("B").OnSelect(func() { selected = "B" }),
	)
	p.Event(uitest.NewMockContext(), uitest.KeyPress(event.KeyDown, event.ModNone)) // → B
	p.Event(uitest.NewMockContext(), uitest.KeyPress(event.KeyEnter, event.ModNone))
	if selected != "B" {
		t.Errorf("Enter selected %q, want B", selected)
	}
}

func TestPanelEscapeCloses(t *testing.T) {
	closed := false
	p := NewPanel(testTheme(), NewItem("A")).OnClose(func() { closed = true })
	p.Event(uitest.NewMockContext(), uitest.KeyPress(event.KeyEscape, event.ModNone))
	if !closed {
		t.Error("Escape did not invoke OnClose")
	}
}

func TestPanelCheckboxToggle(t *testing.T) {
	var got bool
	called := false
	cb := NewCheckboxItem("Show", icon.IconData{}).OnChange(func(v bool) { got = v; called = true })
	p := NewPanel(testTheme(), cb)
	p.Event(uitest.NewMockContext(), uitest.KeyPress(event.KeyEnter, event.ModNone))
	if !called {
		t.Fatal("checkbox OnChange not called")
	}
	if !got {
		t.Errorf("checkbox toggled to %v, want true", got)
	}
	if !cb.Checked {
		t.Error("checkbox Checked not updated")
	}
}

func TestPanelRadioSelect(t *testing.T) {
	var picked string
	r1 := NewRadioItem("one", "One", icon.IconData{}).OnSelect(func(v string) { picked = v })
	r2 := NewRadioItem("two", "Two", icon.IconData{}).OnSelect(func(v string) { picked = v })
	p := NewPanel(testTheme(), r1, r2)
	p.Event(uitest.NewMockContext(), uitest.KeyPress(event.KeyDown, event.ModNone)) // → r2
	p.Event(uitest.NewMockContext(), uitest.KeyPress(event.KeyEnter, event.ModNone))
	if picked != "two" {
		t.Errorf("radio select = %q, want two", picked)
	}
}

func TestPanelMouseClickSelects(t *testing.T) {
	var selected string
	p := NewPanel(testTheme(),
		NewItem("A").OnSelect(func() { selected = "A" }),
		NewItem("B").OnSelect(func() { selected = "B" }),
	)
	// Layout at origin so rows have real bounds.
	size := p.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	p.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))

	// Click in the second row (B): y within [pad+ItemHeight, pad+2*ItemHeight).
	clickY := metrics.Menu.Pad + metrics.Menu.ItemHeight + metrics.Menu.ItemHeight/2
	p.Event(uitest.NewMockContext(), uitest.Click(20, clickY))
	if selected != "B" {
		t.Errorf("mouse click selected %q, want B", selected)
	}
}

// TestPanelDrawsRows pins the row geometry via the mock canvas: the panel
// surface and an accent highlight on the first item are drawn.
func TestPanelDrawsRows(t *testing.T) {
	th := testTheme()
	tok := th.Active()
	p := NewPanel(th, NewItem("A"), NewItem("B"))
	size := p.ContentSize()
	p.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))

	c := uitest.DrawWidget(p)

	// Panel surface: a popover-colored round rect.
	foundSurface := false
	for _, rr := range c.RoundRects {
		if rr.Color == tok.Popover {
			foundSurface = true
		}
	}
	if !foundSurface {
		t.Error("panel popover surface not drawn")
	}
	// First item is highlighted by default → accent round rect.
	foundAccent := false
	for _, rr := range c.RoundRects {
		if rr.Color == tok.Accent {
			foundAccent = true
		}
	}
	if !foundAccent {
		t.Error("highlighted item accent fill not drawn")
	}
	// Two item labels rendered.
	if len(c.StyledTexts) != 2 {
		t.Errorf("expected 2 item labels, got %d", len(c.StyledTexts))
	}
}

var _ widget.Widget = (*Panel)(nil)

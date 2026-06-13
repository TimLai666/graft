package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"

	"github.com/TimLai666/graft"
)

// TestPaginationPressFocusesOnlyHitButton verifies that pressing one page
// button focuses ONLY that button, not every button in the row.
//
// Before the hit-test fix, PaginationWidget.Event forwarded the same mouse
// event to every child button, so MousePress called ctx.RequestFocus on all of
// them and focus landed on the LAST button (Next), not the page clicked.
func TestPaginationPressFocusesOnlyHitButton(t *testing.T) {
	_, restore := paginationForceLight(t)
	defer restore()

	p := graft.Pagination().Pages(5, 1)
	p.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(600, 100)))

	// children order: Prev, page1, page2, page3, page4, page5, Next.
	children := p.Children()
	if len(children) != 7 {
		t.Fatalf("unexpected child count %d", len(children))
	}
	page3 := children[3].(*graft.ButtonWidget)
	b := page3.Bounds()

	ctx := uitest.NewMockContext()
	press := uitest.Click(b.Min.X+b.Width()/2, b.Min.Y+b.Height()/2)
	p.Event(ctx, press)

	if ctx.FocusedWidget() != page3 {
		// Identify which button (if any) got focus.
		idx := -1
		for i, c := range children {
			if c == ctx.FocusedWidget() {
				idx = i
				break
			}
		}
		t.Fatalf("press focused wrong widget: got child index %d, want page3 (index 3)", idx)
	}
}

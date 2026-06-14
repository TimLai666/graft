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
	"github.com/TimLai666/graft/theme"
)

// paginationForceLight pins the theme to light and returns active tokens plus a
// restore function.
func paginationForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

// TestPaginationLayoutHeight pins the row to the 32px default button height
// (page/prev/next links are default-size buttons; h-8).
func TestPaginationLayoutHeight(t *testing.T) {
	_, restore := paginationForceLight(t)
	defer restore()

	p := graft.Pagination().Pages(5, 2)
	size := p.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(600, 100)))
	if size.Height != metrics.Button.Default.Height {
		t.Fatalf("height: got %v want %v", size.Height, metrics.Button.Default.Height)
	}
	// 5 pages (≤7 → all shown) + Previous + Next = 7 buttons.
	if got := len(p.Children()); got != 7 {
		t.Fatalf("button count: got %d want 7", got)
	}
}

// TestPaginationActivePageOutline asserts the active page button draws an
// outline (border + shadow) while others are ghost (no border).
func TestPaginationActivePageOutline(t *testing.T) {
	tok, restore := paginationForceLight(t)
	defer restore()

	p := graft.Pagination().Pages(5, 2)
	p.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(600, 100)))
	canvas := uitest.DrawWidget(p)

	// Exactly one 1px border in light mode comes from the active outline page
	// button (ghost pages and ghost prev/next have no border).
	borderCount := 0
	for _, s := range canvas.StrokeRoundRects {
		if s.StrokeWidth == metrics.Button.BorderWidth && (s.Color == tok.Border) {
			borderCount++
		}
	}
	if borderCount != 1 {
		t.Fatalf("active-page outline border: got %d want 1", borderCount)
	}
}

// TestPaginationOnSelect verifies clicking a page fires OnSelect with that page.
func TestPaginationOnSelect(t *testing.T) {
	_, restore := paginationForceLight(t)
	defer restore()

	var picked int
	p := graft.Pagination().Pages(5, 1).OnSelect(func(n int) { picked = n })
	p.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(600, 100)))

	// The page-3 button is the 4th child (Prev, 1, 2, 3, ...). Click its center.
	children := p.Children()
	if len(children) < 4 {
		t.Fatalf("unexpected child count %d", len(children))
	}
	page3 := children[3].(*graft.ButtonWidget) // Prev=0, page1=1, page2=2, page3=3
	b := page3.Bounds()
	uitest.SimulateClick(p, b.Min.X+b.Width()/2, b.Min.Y+b.Height()/2)
	if picked != 3 {
		t.Fatalf("OnSelect page: got %d want 3", picked)
	}
}

// TestPaginationEllipsisForLargeTotals asserts elision inserts ellipsis slots.
func TestPaginationEllipsisForLargeTotals(t *testing.T) {
	_, restore := paginationForceLight(t)
	defer restore()

	p := graft.Pagination().Pages(20, 10)
	p.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(800, 100)))
	canvas := uitest.DrawWidget(p)
	// With 20 pages around current=10, the row shows 1 … 9 10 11 … 20, i.e.
	// fewer than 20 page buttons plus 2 ellipses. Confirm the button count is
	// well under 20+2.
	if got := len(p.Children()); got >= 20 {
		t.Fatalf("expected elided button count, got %d", got)
	}
	_ = canvas
}

// TestGoldenPagination renders pages 1..5 with page 2 active, plus prev/next.
func TestGoldenPagination(t *testing.T) {
	gtest.GoldenLightDark(t, "pagination-basic", func() widget.Widget {
		return primitives.VBox(graft.Pagination().Pages(5, 2)).
			Padding(12).CrossAlign(primitives.CrossAxisStart)
	})
	gtest.GoldenLightDark(t, "pagination-ellipsis", func() widget.Widget {
		return primitives.VBox(graft.Pagination().Pages(20, 10)).
			Padding(12).CrossAlign(primitives.CrossAxisStart)
	})
}

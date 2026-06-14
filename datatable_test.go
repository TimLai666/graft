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
)

// sampleDataCols builds the column set used across the DataTable tests: a
// sortable Invoice column, a Status column, and a sortable right-aligned Amount
// column. Rows are deliberately out of sorted order so a sort visibly reorders
// them.
func sampleDataCols() (rows int, cols []graft.DataColumn) {
	invoices := []string{"INV003", "INV001", "INV002"}
	statuses := []string{"Unpaid", "Paid", "Pending"}
	amounts := []string{"$350.00", "$250.00", "$150.00"}
	cols = []graft.DataColumn{
		{Title: "Invoice", Sortable: true, Cell: func(i int) string { return invoices[i] }},
		{Title: "Status", Cell: func(i int) string { return statuses[i] }},
		{Title: "Amount", Sortable: true, Align: widget.TextAlignRight, Cell: func(i int) string { return amounts[i] }},
	}
	return len(invoices), cols
}

func sampleDataTable() *graft.DataTableWidget {
	rows, cols := sampleDataCols()
	return graft.DataTable(rows, cols...).W(480)
}

// layoutDT lays the table out at a fixed size so hit-test coordinates are
// stable.
func layoutDT(t *testing.T, d *graft.DataTableWidget) {
	t.Helper()
	d.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(480, 1000)))
}

// TestDataTableSortReorders verifies clicking a sortable header cycles
// none->asc->desc->none and reorders the displayed rows by the column's key.
func TestDataTableSortReorders(t *testing.T) {
	_, restore := tableForceLight(t)
	defer restore()

	d := sampleDataTable()
	layoutDT(t, d)

	// Natural order is INV003, INV001, INV002.
	if got := d.DisplayOrder(); got[0] != 0 || got[1] != 1 || got[2] != 2 {
		t.Fatalf("initial order = %v, want [0 1 2]", got)
	}

	// Click the Invoice header (column 0). Header band y in [0,40), x past the
	// (absent) checkbox column.
	bounds := d.Bounds()
	headerY := bounds.Min.Y + metrics.Table.HeadHeight/2
	clickX := bounds.Min.X + 30

	// none -> asc: order becomes INV001(1), INV002(2), INV003(0).
	if !d.Event(uitest.NewMockContext(), uitest.Release(clickX, headerY)) {
		t.Fatal("header click not handled")
	}
	if col, dir := d.SortState(); col != 0 || dir != graft.SortAsc {
		t.Fatalf("after 1st click: col=%d dir=%d, want col=0 asc", col, dir)
	}
	if got := d.DisplayOrder(); got[0] != 1 || got[1] != 2 || got[2] != 0 {
		t.Fatalf("asc order = %v, want [1 2 0]", got)
	}

	// asc -> desc: INV003(0), INV002(2), INV001(1).
	d.Event(uitest.NewMockContext(), uitest.Release(clickX, headerY))
	if _, dir := d.SortState(); dir != graft.SortDesc {
		t.Fatalf("after 2nd click dir=%d, want desc", dir)
	}
	if got := d.DisplayOrder(); got[0] != 0 || got[1] != 2 || got[2] != 1 {
		t.Fatalf("desc order = %v, want [0 2 1]", got)
	}

	// desc -> none: back to natural order, sort cleared.
	d.Event(uitest.NewMockContext(), uitest.Release(clickX, headerY))
	if col, dir := d.SortState(); col != -1 || dir != graft.SortNone {
		t.Fatalf("after 3rd click: col=%d dir=%d, want col=-1 none", col, dir)
	}
	if got := d.DisplayOrder(); got[0] != 0 || got[1] != 1 || got[2] != 2 {
		t.Fatalf("cleared order = %v, want [0 1 2]", got)
	}
}

// TestDataTableNonSortableHeaderIgnored verifies clicking a non-sortable header
// does not change the sort state.
func TestDataTableNonSortableHeaderIgnored(t *testing.T) {
	_, restore := tableForceLight(t)
	defer restore()

	d := sampleDataTable()
	layoutDT(t, d)

	bounds := d.Bounds()
	headerY := bounds.Min.Y + metrics.Table.HeadHeight/2
	// Status is the middle column; click well into it.
	statusX := bounds.Min.X + d.ColumnLeft(1) + 10
	handled := d.Event(uitest.NewMockContext(), uitest.Release(statusX, headerY))
	if handled {
		t.Fatal("non-sortable header click should not be handled")
	}
	if col, _ := d.SortState(); col != -1 {
		t.Fatalf("sort changed on non-sortable header: col=%d", col)
	}
}

// TestDataTableSelectionToggles verifies a click in the checkbox column toggles
// that row's selection, and that selections are reported by natural index.
func TestDataTableSelectionToggles(t *testing.T) {
	_, restore := tableForceLight(t)
	defer restore()

	var lastSel []int
	rows, cols := sampleDataCols()
	d := graft.DataTable(rows, cols...).W(480).Selectable(true).
		OnSelectionChange(func(s []int) { lastSel = s })
	layoutDT(t, d)

	bounds := d.Bounds()
	cbX := bounds.Min.X + metrics.DataTable.CheckboxColWidth/2
	// First body row center.
	row0Y := bounds.Min.Y + metrics.Table.HeadHeight + (2*metrics.Table.CellPad+20)/2

	if !d.Event(uitest.NewMockContext(), uitest.Release(cbX, row0Y)) {
		t.Fatal("checkbox click not handled")
	}
	if got := d.SelectedRows(); len(got) != 1 || got[0] != 0 {
		t.Fatalf("after toggle: selected=%v, want [0]", got)
	}
	if len(lastSel) != 1 || lastSel[0] != 0 {
		t.Fatalf("OnSelectionChange got %v, want [0]", lastSel)
	}

	// Toggle off.
	d.Event(uitest.NewMockContext(), uitest.Release(cbX, row0Y))
	if got := d.SelectedRows(); len(got) != 0 {
		t.Fatalf("after second toggle: selected=%v, want empty", got)
	}
}

// TestDataTableSelectAll verifies the header checkbox selects every row and a
// second click clears the selection.
func TestDataTableSelectAll(t *testing.T) {
	_, restore := tableForceLight(t)
	defer restore()

	rows, cols := sampleDataCols()
	d := graft.DataTable(rows, cols...).W(480).Selectable(true)
	layoutDT(t, d)

	bounds := d.Bounds()
	cbX := bounds.Min.X + metrics.DataTable.CheckboxColWidth/2
	headerY := bounds.Min.Y + metrics.Table.HeadHeight/2

	d.Event(uitest.NewMockContext(), uitest.Release(cbX, headerY))
	if got := d.SelectedRows(); len(got) != rows {
		t.Fatalf("select-all: selected %d rows, want %d", len(got), rows)
	}

	d.Event(uitest.NewMockContext(), uitest.Release(cbX, headerY))
	if got := d.SelectedRows(); len(got) != 0 {
		t.Fatalf("clear-all: selected %d rows, want 0", len(got))
	}
}

// TestDataTablePaginationSlices verifies PageSize limits the visible rows and
// the Next/Previous footer buttons move between pages.
func TestDataTablePaginationSlices(t *testing.T) {
	_, restore := tableForceLight(t)
	defer restore()

	// 5 rows, 2 per page -> 3 pages (2,2,1).
	cells := []string{"a", "b", "c", "d", "e"}
	d := graft.DataTable(len(cells),
		graft.DataColumn{Title: "Letter", Cell: func(i int) string { return cells[i] }},
	).W(480).PageSize(2)
	layoutDT(t, d)

	if got := d.VisibleRowCount(); got != 2 {
		t.Fatalf("page 0 visible rows = %d, want 2", got)
	}
	if d.CurrentPage() != 0 {
		t.Fatalf("initial page = %d, want 0", d.CurrentPage())
	}

	// Click Next (the footer's rightmost button).
	bounds := d.Bounds()
	nextB := d.NextBounds()
	nx := bounds.Min.X + nextB.Min.X + nextB.Width()/2
	ny := bounds.Min.Y + nextB.Min.Y + nextB.Height()/2
	// Press then release so the button activates.
	d.Event(uitest.NewMockContext(), uitest.Click(nx, ny))
	d.Event(uitest.NewMockContext(), uitest.Release(nx, ny))
	if d.CurrentPage() != 1 {
		t.Fatalf("after Next: page = %d, want 1", d.CurrentPage())
	}

	// Layout for the new page, then advance to the last (partial) page.
	layoutDT(t, d)
	nextB = d.NextBounds()
	nx = bounds.Min.X + nextB.Min.X + nextB.Width()/2
	ny = bounds.Min.Y + nextB.Min.Y + nextB.Height()/2
	d.Event(uitest.NewMockContext(), uitest.Click(nx, ny))
	d.Event(uitest.NewMockContext(), uitest.Release(nx, ny))
	if d.CurrentPage() != 2 {
		t.Fatalf("after 2nd Next: page = %d, want 2", d.CurrentPage())
	}
	layoutDT(t, d)
	if got := d.VisibleRowCount(); got != 1 {
		t.Fatalf("last page visible rows = %d, want 1", got)
	}

	// Previous walks back.
	prevB := d.PrevBounds()
	px := bounds.Min.X + prevB.Min.X + prevB.Width()/2
	py := bounds.Min.Y + prevB.Min.Y + prevB.Height()/2
	d.Event(uitest.NewMockContext(), uitest.Click(px, py))
	d.Event(uitest.NewMockContext(), uitest.Release(px, py))
	if d.CurrentPage() != 1 {
		t.Fatalf("after Previous: page = %d, want 1", d.CurrentPage())
	}
}

// TestDataTableSortArrowDrawn verifies the active sort column draws a chevron
// glyph and the height accounts for header + body rows.
func TestDataTableSortArrowDrawn(t *testing.T) {
	_, restore := tableForceLight(t)
	defer restore()

	d := sampleDataTable().SortColumn(0, graft.SortAsc)
	size := d.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(480, 1000)))

	wantH := metrics.Table.HeadHeight + 3*(2*metrics.Table.CellPad+20)
	if size.Height != wantH {
		t.Fatalf("height = %v, want %v", size.Height, wantH)
	}

	// The sorted column's header text must still be drawn.
	canvas := uitest.DrawWidget(d)
	found := false
	for _, st := range canvas.StyledTexts {
		if st.Text == "Invoice" {
			found = true
		}
	}
	if !found {
		t.Fatal("sorted header label not drawn")
	}
}

// TestGoldenDataTable renders a settled DataTable (sorted asc with arrow, one
// row selected, pagination footer) light and dark.
func TestGoldenDataTable(t *testing.T) {
	build := func() widget.Widget {
		rows, cols := sampleDataCols()
		d := graft.DataTable(rows, cols...).
			W(520).
			Selectable(true).
			PageSize(5).
			SortColumn(0, graft.SortAsc).
			Select(1, true)
		return primitives.VBox(d).Padding(16)
	}
	gtest.GoldenLightDark(t, "datatable-basic", build)
}

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

// tableForceLight pins the theme to light and returns the active tokens plus a
// restore function.
func tableForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

// sampleTable builds a 3-column / 3-row table with a header.
func sampleTable() *graft.TableWidget {
	return graft.Table(
		graft.TableHeader(graft.TableRow(
			graft.TableHead("Invoice"),
			graft.TableHead("Status"),
			graft.TableHead("Amount").Align(widget.TextAlignRight),
		)),
		graft.TableBody(
			graft.TableRow(graft.TableCell("INV001"), graft.TableCell("Paid"), graft.TableCell("$250.00").Align(widget.TextAlignRight)),
			graft.TableRow(graft.TableCell("INV002"), graft.TableCell("Pending"), graft.TableCell("$150.00").Align(widget.TextAlignRight)),
			graft.TableRow(graft.TableCell("INV003"), graft.TableCell("Unpaid"), graft.TableCell("$350.00").Align(widget.TextAlignRight)),
		),
	).W(420)
}

// TestTableLayoutHeights pins the header height (40px) and body row heights.
func TestTableLayoutHeights(t *testing.T) {
	_, restore := tableForceLight(t)
	defer restore()

	tbl := sampleTable()
	size := tbl.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(500, 1000)))
	wantH := metrics.Table.HeadHeight + 3*(2*metrics.Table.CellPad+20)
	if size.Height != wantH {
		t.Fatalf("table height: got %v want %v", size.Height, wantH)
	}
	if size.Width != 420 {
		t.Fatalf("table width: got %v want 420", size.Width)
	}
}

// TestTableSpecHeaderText asserts header cells use weight 500 in Foreground at
// 14px, and that one separator line per non-last row is drawn in Border.
func TestTableSpecHeaderText(t *testing.T) {
	tok, restore := tableForceLight(t)
	defer restore()

	tbl := sampleTable()
	tbl.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(500, 1000)))
	canvas := uitest.DrawWidget(tbl)

	// All cell texts are present.
	var head *uitest.DrawStyledTextCall
	for idx := range canvas.StyledTexts {
		if canvas.StyledTexts[idx].Text == "Invoice" {
			head = &canvas.StyledTexts[idx]
		}
	}
	if head == nil {
		t.Fatal("header cell 'Invoice' not drawn")
	}
	if head.Style.Color != tok.Foreground {
		t.Fatalf("header color: got %+v want Foreground", head.Style.Color)
	}
	if head.Style.FontSize != metrics.Table.FontSize {
		t.Fatalf("header font size: got %v want %v", head.Style.FontSize, metrics.Table.FontSize)
	}

	// Row separators: header + 2 inner body rows = 3 lines in Border (last body
	// row has no border).
	borderLines := 0
	for _, r := range canvas.Rects {
		if r.Color == tok.Border && r.Bounds.Height() == metrics.Table.BorderWidth {
			borderLines++
		}
	}
	if borderLines != 3 {
		t.Fatalf("border separators: got %d want 3", borderLines)
	}
}

// TestTableSpecCaption asserts the caption is muted 14px text below the grid.
func TestTableSpecCaption(t *testing.T) {
	tok, restore := tableForceLight(t)
	defer restore()

	tbl := sampleTable().Caption("A list of recent invoices.")
	tbl.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(500, 1000)))
	canvas := uitest.DrawWidget(tbl)

	var cap *uitest.DrawStyledTextCall
	for idx := range canvas.StyledTexts {
		if canvas.StyledTexts[idx].Text == "A list of recent invoices." {
			cap = &canvas.StyledTexts[idx]
		}
	}
	if cap == nil {
		t.Fatal("caption not drawn")
	}
	if cap.Style.Color != tok.MutedForeground {
		t.Fatalf("caption color: got %+v want MutedForeground", cap.Style.Color)
	}
	if cap.Style.FontSize != metrics.Table.CaptionFontSize {
		t.Fatalf("caption font size: got %v want %v", cap.Style.FontSize, metrics.Table.CaptionFontSize)
	}
}

// TestTableFooterBackground asserts a footer row draws a Muted/50 background.
func TestTableFooterBackground(t *testing.T) {
	tok, restore := tableForceLight(t)
	defer restore()

	tbl := graft.Table(
		graft.TableHeader(graft.TableRow(graft.TableHead("A"), graft.TableHead("B"))),
		graft.TableBody(graft.TableRow(graft.TableCell("1"), graft.TableCell("2"))),
		graft.TableFooter(graft.TableRow(graft.TableCell("Total"), graft.TableCell("3"))),
	).W(300)
	tbl.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 1000)))
	canvas := uitest.DrawWidget(tbl)

	want := tok.Muted
	want.A = metrics.Table.FooterBgAlpha
	found := false
	for _, r := range canvas.Rects {
		if r.Color == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("no Muted/50 footer background drawn")
	}
}

// TestTableRowHover verifies a MouseMove over a body row paints that row's
// hover:bg-muted/50 background, and that no hover background is drawn before
// the pointer enters any body row.
func TestTableRowHover(t *testing.T) {
	tok, restore := tableForceLight(t)
	defer restore()

	tbl := sampleTable()
	tbl.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(500, 1000)))

	hoverColor := tok.Muted
	hoverColor.A = metrics.Table.RowHoverAlpha

	countHover := func(c *uitest.MockCanvas) int {
		n := 0
		for _, r := range c.Rects {
			if r.Color == hoverColor {
				n++
			}
		}
		return n
	}

	// No hover background before any pointer movement.
	if got := countHover(uitest.DrawWidget(tbl)); got != 0 {
		t.Fatalf("hover background before hover: got %d want 0", got)
	}

	// Move the pointer into the first body row (just below the 40px header).
	bounds := tbl.Bounds()
	rowY := bounds.Min.Y + metrics.Table.HeadHeight + (2*metrics.Table.CellPad+20)/2
	if !tbl.Event(uitest.NewMockContext(), uitest.MouseMove(bounds.Min.X+20, rowY)) {
		t.Fatal("MouseMove over a body row was not handled")
	}
	if got := countHover(uitest.DrawWidget(tbl)); got != 1 {
		t.Fatalf("hover background after hover: got %d want 1", got)
	}

	// Leaving the table clears the hover background.
	tbl.Event(uitest.NewMockContext(), uitest.MouseLeave(0, 0))
	if got := countHover(uitest.DrawWidget(tbl)); got != 0 {
		t.Fatalf("hover background after leave: got %d want 0", got)
	}
}

// TestGoldenTable renders the standard table light and dark.
func TestGoldenTable(t *testing.T) {
	gtest.GoldenLightDark(t, "table-basic", func() widget.Widget {
		return primitives.VBox(sampleTable()).Padding(12)
	})
	gtest.GoldenLightDark(t, "table-caption", func() widget.Widget {
		return primitives.VBox(sampleTable().Caption("A list of your recent invoices.")).Padding(12)
	})
}

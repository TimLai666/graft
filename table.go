package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// TableWidget is shadcn's static Table: a styled data grid with a header, body,
// optional footer, and optional caption (docs/research/03-shadcn-pixel-spec.md
// §5 "Table"). This is the presentational Table, not DataTable — there is no
// sorting, filtering, or virtualization.
//
// WRAP DECISION (DESIGN.md §3.2): graft-OWNED. Column alignment requires a
// single global width pass over every cell across all sections (a header cell
// and the body cells below it must share one column width). primitives.Box
// cannot express that cross-row constraint, so the table owns its grid layout
// and draws the section chrome (row border-b separators, footer Muted/50 fill,
// per-row hover) itself, resolving colors from the active token set at draw
// time.
type TableWidget struct {
	widget.WidgetBase

	header  *TableSection
	body    *TableSection
	footer  *TableSection
	caption string

	theme *theme.Theme
	width float32 // explicit outer width (w-full otherwise)

	// hoverRow is the global index (across body rows only) of the hovered row,
	// or -1.
	hoverRow int

	// layout cache, rebuilt each Layout.
	colWidths []float32
	rowLayout []rowBox
}

// rowBox positions one laid-out row in widget-local coordinates.
type rowBox struct {
	y, height float32
	kind      tableRowKind
	bodyIndex int // index among body rows, or -1
	row       *TableRowDef
}

type tableRowKind uint8

const (
	tableRowKindHead tableRowKind = iota
	tableRowKindBody
	tableRowKindFooter
)

// TableSection groups rows for a table region (header, body, footer).
type TableSection struct {
	rows []*TableRowDef
	role sectionRole
}

// TableRowDef is one table row holding its cells.
type TableRowDef struct {
	cells []*TableCellDef
}

// cellKind distinguishes header cells (TableHead) from data cells (TableCell).
type cellKind uint8

const (
	cellKindHead cellKind = iota
	cellKindData
)

// TableCellDef is one table cell: its text plus head/data kind and alignment.
type TableCellDef struct {
	text  string
	kind  cellKind
	align widget.TextAlign
}

// Table creates a table from its sections (TableHeader, TableBody, and an
// optional TableFooter) and an optional TableCaption (added via .Caption).
func Table(sections ...*TableSection) *TableWidget {
	t := &TableWidget{theme: CurrentTheme(), hoverRow: -1}
	t.SetVisible(true)
	t.SetEnabled(true)
	for _, s := range sections {
		if s == nil {
			continue
		}
		switch s.role {
		case sectionFooter:
			t.footer = s
		case sectionHeader:
			t.header = s
		default:
			t.body = s
		}
	}
	return t
}

// sectionRole tags a section so Table can route it without positional rules.
type sectionRole uint8

const (
	sectionBody sectionRole = iota
	sectionHeader
	sectionFooter
)

// TableHeader groups the header rows (typically one TableRow of TableHead).
func TableHeader(rows ...*TableRowDef) *TableSection {
	return &TableSection{rows: rows, role: sectionHeader}
}

// TableBody groups the data rows.
func TableBody(rows ...*TableRowDef) *TableSection {
	return &TableSection{rows: rows, role: sectionBody}
}

// TableFooter groups the footer rows (Muted/50 background, medium weight).
func TableFooter(rows ...*TableRowDef) *TableSection {
	return &TableSection{rows: rows, role: sectionFooter}
}

// TableRow groups cells into one row.
func TableRow(cells ...*TableCellDef) *TableRowDef {
	return &TableRowDef{cells: cells}
}

// TableHead creates a header cell: left-aligned, weight 500, h-10.
func TableHead(text string) *TableCellDef {
	return &TableCellDef{text: text, kind: cellKindHead, align: widget.TextAlignLeft}
}

// TableCell creates a data cell: left-aligned, normal weight, p-2.
func TableCell(text string) *TableCellDef {
	return &TableCellDef{text: text, kind: cellKindData, align: widget.TextAlignLeft}
}

// Align overrides the horizontal text alignment of a cell.
func (c *TableCellDef) Align(a widget.TextAlign) *TableCellDef {
	c.align = a
	return c
}

// Caption sets the table caption (rendered below, muted 14px).
func (t *TableWidget) Caption(s string) *TableWidget {
	t.caption = s
	return t
}

// W pins the outer table width in px (otherwise it fills available width).
func (t *TableWidget) W(px float32) *TableWidget {
	t.width = px
	return t
}

// Theme pins a specific theme instead of the process-wide current theme.
func (t *TableWidget) Theme(th *theme.Theme) *TableWidget {
	if th != nil {
		t.theme = th
	}
	return t
}

// columnCount returns the maximum cell count across all rows.
func (t *TableWidget) columnCount() int {
	n := 0
	for _, s := range t.sections() {
		for _, r := range s.rows {
			if len(r.cells) > n {
				n = len(r.cells)
			}
		}
	}
	return n
}

// sections returns the present sections in render order.
func (t *TableWidget) sections() []*TableSection {
	var out []*TableSection
	if t.header != nil {
		out = append(out, t.header)
	}
	if t.body != nil {
		out = append(out, t.body)
	}
	if t.footer != nil {
		out = append(out, t.footer)
	}
	return out
}

// cellFontFamily resolves the family for a cell's weight.
func (t *TableWidget) cellFontFamily(weight int) string {
	if t.theme.FontSans != theme.DefaultFontSans {
		return t.theme.FontSans
	}
	return fonts.Family(weight)
}

// cellWeight returns the font weight for a cell given its kind and section.
func cellWeight(kind cellKind, rk tableRowKind) int {
	switch {
	case kind == cellKindHead:
		return metrics.Table.HeadFontWeight
	case rk == tableRowKindFooter:
		return metrics.Table.FooterFontWeight
	default:
		return metrics.Table.CellFontWeight
	}
}

// measureColumns computes each column's natural width: max over all cells of
// text width + horizontal padding.
func (t *TableWidget) measureColumns() []float32 {
	cols := t.columnCount()
	widths := make([]float32, cols)
	size := metrics.Table.FontSize
	for _, s := range t.sections() {
		rk := sectionRowKind(s.role)
		for _, r := range s.rows {
			for ci, c := range r.cells {
				w := textmetrics.Width(t.cellFontFamily(cellWeight(c.kind, rk)), size, c.text)
				w += 2 * cellPadX(c.kind)
				if w > widths[ci] {
					widths[ci] = w
				}
			}
		}
	}
	return widths
}

// cellPadX returns the horizontal padding for a cell kind (head px-2, cell p-2;
// both 8px).
func cellPadX(kind cellKind) float32 {
	if kind == cellKindHead {
		return metrics.Table.HeadPadX
	}
	return metrics.Table.CellPad
}

// sectionRowKind maps a section role to its row kind.
func sectionRowKind(role sectionRole) tableRowKind {
	switch role {
	case sectionHeader:
		return tableRowKindHead
	case sectionFooter:
		return tableRowKindFooter
	default:
		return tableRowKindBody
	}
}

// rowHeight returns the height of a row of the given kind.
func rowHeight(rk tableRowKind) float32 {
	if rk == tableRowKindHead {
		return metrics.Table.HeadHeight
	}
	// Body/footer rows are p-2 around a single text-sm line (line box 20px).
	return 2*metrics.Table.CellPad + 20
}

// Layout sizes the table grid: column widths scaled to fill w-full, fixed row
// heights, plus the caption below.
func (t *TableWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	w := t.width
	if w <= 0 {
		w = c.MaxWidth
		if w <= 0 || w >= geometry.Infinity {
			w = 400
		}
	}

	natural := t.measureColumns()
	var sum float32
	for _, cw := range natural {
		sum += cw
	}
	// w-full: distribute any slack evenly so the columns span the full width.
	t.colWidths = make([]float32, len(natural))
	copy(t.colWidths, natural)
	if sum < w && len(natural) > 0 {
		extra := (w - sum) / float32(len(natural))
		for i := range t.colWidths {
			t.colWidths[i] += extra
		}
	} else if sum > w {
		w = sum // do not shrink below natural content width
	}

	// Build row layout (vertical stack of section rows).
	t.rowLayout = t.rowLayout[:0]
	var y float32
	bodyIdx := 0
	for _, s := range t.sections() {
		rk := sectionRowKind(s.role)
		for _, r := range s.rows {
			h := rowHeight(rk)
			rb := rowBox{y: y, height: h, kind: rk, row: r, bodyIndex: -1}
			if rk == tableRowKindBody {
				rb.bodyIndex = bodyIdx
				bodyIdx++
			}
			t.rowLayout = append(t.rowLayout, rb)
			y += h
		}
	}

	totalH := y
	if t.caption != "" {
		totalH += metrics.Table.CaptionMarginTop + 20 // caption line box
	}

	size := c.Constrain(geometry.Sz(w, totalH))
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

// Draw paints rows (header text, body cells with hover, footer Muted/50 fill),
// the 1px row separators, and the caption.
func (t *TableWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !t.IsVisible() {
		return
	}
	th := t.theme
	tok := th.Active()
	bounds := t.Bounds()
	size := metrics.Table.FontSize

	for _, rb := range t.rowLayout {
		rowTop := bounds.Min.Y + rb.y
		rowRect := geometry.NewRect(bounds.Min.X, rowTop, bounds.Width(), rb.height)

		// Backgrounds.
		switch rb.kind {
		case tableRowKindFooter:
			canvas.DrawRect(rowRect, draw.Alpha(tok.Muted, metrics.Table.FooterBgAlpha))
		case tableRowKindBody:
			if rb.bodyIndex == t.hoverRow {
				canvas.DrawRect(rowRect, draw.Alpha(tok.Muted, metrics.Table.RowHoverAlpha))
			}
		}

		// Cells.
		var x float32
		for ci, cell := range rb.row.cells {
			if ci >= len(t.colWidths) {
				break
			}
			colW := t.colWidths[ci]
			padX := cellPadX(cell.kind)
			weight := cellWeight(cell.kind, rb.kind)
			family := t.cellFontFamily(weight)
			col := tok.Foreground
			textRect := geometry.NewRect(
				bounds.Min.X+x+padX,
				rowTop,
				colW-2*padX,
				rb.height,
			)
			drawTableText(canvas, cell.text, textRect, family, size, col, weight, cell.align)
			x += colW
		}
	}

	// Row separators: a border-b under every row except the last body row (the
	// body has [&_tr:last-child]:border-0); the footer has a top border instead.
	t.drawSeparators(canvas, bounds, tok)

	// Caption below the grid.
	if t.caption != "" {
		var gridH float32
		if len(t.rowLayout) > 0 {
			last := t.rowLayout[len(t.rowLayout)-1]
			gridH = last.y + last.height
		}
		capRect := geometry.NewRect(
			bounds.Min.X,
			bounds.Min.Y+gridH+metrics.Table.CaptionMarginTop,
			bounds.Width(),
			20,
		)
		drawTableText(canvas, t.caption, capRect, t.cellFontFamily(400),
			metrics.Table.CaptionFontSize, tok.MutedForeground, 400, widget.TextAlignCenter)
	}
}

// drawSeparators draws the 1px row borders matching shadcn's border-b / border-t
// rules: every header row and every body row except the last gets a bottom
// border; the footer gets a top border.
func (t *TableWidget) drawSeparators(canvas widget.Canvas, bounds geometry.Rect, tok *theme.Tokens) {
	bw := metrics.Table.BorderWidth
	lastBodyIdx := -1
	for i, rb := range t.rowLayout {
		if rb.kind == tableRowKindBody {
			lastBodyIdx = i
		}
	}
	for i, rb := range t.rowLayout {
		switch rb.kind {
		case tableRowKindHead:
			// border-b: occupy the row's last pixel row.
			drawHLine(canvas, bounds, rb.y+rb.height-bw, tok.Border, bw)
		case tableRowKindBody:
			if i != lastBodyIdx {
				drawHLine(canvas, bounds, rb.y+rb.height-bw, tok.Border, bw)
			}
		case tableRowKindFooter:
			// border-t: occupy the footer's first pixel row.
			drawHLine(canvas, bounds, rb.y, tok.Border, bw)
		}
	}
}

// drawHLine draws a full-width 1px separator whose top edge is at local topY.
func drawHLine(canvas widget.Canvas, bounds geometry.Rect, topY float32, col widget.Color, w float32) {
	canvas.DrawRect(geometry.NewRect(bounds.Min.X, bounds.Min.Y+topY, bounds.Width(), w), col)
}

// drawTableText draws cell text via StyledTextDrawer, falling back to DrawText
// (bold at weight >= 600) on mock canvases.
func drawTableText(canvas widget.Canvas, text string, bounds geometry.Rect, family string, size float32, col widget.Color, weight int, align widget.TextAlign) {
	if text == "" {
		return
	}
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(text, bounds, widget.TextStyle{
			FontFamily: family,
			FontSize:   size,
			Color:      col,
			Align:      align,
		})
		return
	}
	canvas.DrawText(text, bounds, size, col, weight >= 600, align)
}

// Event handles row hover highlighting for body rows.
func (t *TableWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	bounds := t.Bounds()
	switch me.MouseType {
	case event.MouseLeave:
		if t.hoverRow != -1 {
			t.hoverRow = -1
			t.invalidate(ctx)
		}
		return true
	case event.MouseMove, event.MouseEnter:
		local := me.Position.Y - bounds.Min.Y
		newHover := -1
		for _, rb := range t.rowLayout {
			if rb.kind == tableRowKindBody && local >= rb.y && local < rb.y+rb.height {
				newHover = rb.bodyIndex
				break
			}
		}
		if newHover != t.hoverRow {
			t.hoverRow = newHover
			t.invalidate(ctx)
		}
		return newHover != -1
	}
	return false
}

func (t *TableWidget) invalidate(ctx widget.Context) {
	t.SetNeedsRedraw(true)
	ctx.InvalidateRect(t.Bounds())
}

// Children returns nil; the table renders its cells directly.
func (t *TableWidget) Children() []widget.Widget { return nil }

// Compile-time interface check.
var _ widget.Widget = (*TableWidget)(nil)

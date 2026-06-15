package graft

import (
	"sort"
	"strconv"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// SortDirection is the active sort order of a DataTable column.
type SortDirection uint8

const (
	// SortNone is the unsorted state (rows render in their natural order).
	SortNone SortDirection = iota
	// SortAsc sorts ascending (a→z) with an up-arrow indicator.
	SortAsc
	// SortDesc sorts descending (z→a) with a down-arrow indicator.
	SortDesc
)

// DataColumn declares one column of a DataTable: its header title, a cell
// accessor that returns the display string for a given row, and optional
// sorting. This is a data-driven description (no per-cell widgets), so a table
// of N rows costs N closures, not N×cols widgets.
type DataColumn struct {
	// Title is the header label.
	Title string

	// Cell returns the text rendered for row i (0-based, in the caller's
	// natural row order — the widget remaps display order to natural order
	// before calling this).
	Cell func(row int) string

	// Sortable enables click-to-sort on this column's header.
	Sortable bool

	// SortKey returns the comparison key for row i. When nil it defaults to
	// Cell, so columns sort by their displayed text.
	SortKey func(row int) string

	// Align sets the horizontal text alignment for both the header and cells
	// (default left).
	Align widget.TextAlign
}

// sortKey returns the effective sort key accessor (SortKey, or Cell as a
// fallback).
func (c DataColumn) sortKey() func(int) string {
	if c.SortKey != nil {
		return c.SortKey
	}
	return c.Cell
}

// DataTableWidget is shadcn's DataTable: the presentational Table styling plus
// the recipe's interactive chrome — sortable headers (click to cycle
// none→asc→desc), an optional leading checkbox selection column with a
// select-all header box, and optional pagination with a Previous/Next footer.
//
// WRAP DECISION (DESIGN.md §3.2): graft-OWNED, building on the same grid layout
// the presentational Table owns (a global column-width pass cannot be expressed
// with Box). The widget draws the header/row/separator chrome itself, reusing
// every metrics.Table constant, and adds the sort arrow, checkbox cells, and a
// footer that embeds two graft.Button children for Previous/Next. All hit
// testing is container-local (header cells, checkbox cells, footer buttons):
// mouse events are routed to the slot under the pointer, never broadcast.
type DataTableWidget struct {
	widget.WidgetBase

	cols     []DataColumn
	rowCount int

	theme *theme.Theme
	width float32 // explicit outer width (w-full otherwise)

	// selectable adds the leading checkbox column + select-all header box.
	selectable bool
	// selected tracks selected NATURAL row indices.
	selected map[int]bool

	// pageSize > 0 enables pagination; 0 shows all rows.
	pageSize int
	page     int // 0-based current page

	// sortCol is the index into cols of the active sort column, or -1.
	sortCol int
	sortDir SortDirection

	// order maps display position -> natural row index, rebuilt on sort.
	order []int

	// callbacks
	onSortChange      func(col int, dir SortDirection)
	onSelectionChange func(selected []int)

	// hoverRow is the display index of the hovered body row, or -1.
	hoverRow int

	// pagination footer buttons (nil unless pageSize > 0).
	prevBtn *ButtonWidget
	nextBtn *ButtonWidget

	// layout cache, rebuilt each Layout.
	colWidths []float32 // data column widths (excludes checkbox column)
	headTop   float32   // local y of the header row
	bodyTop   float32   // local y of the first body row
	rowH      float32   // body row height
	footerTop float32   // local y of the footer band (when paginating)
	checkboxW float32   // leading checkbox column width (0 when not selectable)
}

// DataTable creates a data table that renders rowCount rows, drawing each cell
// via the matching column's Cell accessor.
func DataTable(rowCount int, cols ...DataColumn) *DataTableWidget {
	if rowCount < 0 {
		rowCount = 0
	}
	d := &DataTableWidget{
		cols:     cols,
		rowCount: rowCount,
		theme:    CurrentTheme(),
		selected: make(map[int]bool),
		sortCol:  -1,
		sortDir:  SortNone,
		hoverRow: -1,
	}
	d.SetVisible(true)
	d.SetEnabled(true)
	d.rebuildOrder()
	return d
}

// Selectable adds the leading checkbox column (per-row toggles + a select-all
// header box). Off by default.
func (d *DataTableWidget) Selectable(v bool) *DataTableWidget {
	d.selectable = v
	return d
}

// PageSize enables pagination, showing at most n rows per page with a
// Previous/Next footer. n <= 0 disables pagination (all rows shown).
func (d *DataTableWidget) PageSize(n int) *DataTableWidget {
	if n < 0 {
		n = 0
	}
	d.pageSize = n
	if d.page > d.lastPage() {
		d.page = d.lastPage()
	}
	d.buildFooter()
	return d
}

// OnSortChange registers a callback fired whenever the sort column or direction
// changes, with the column index (-1 when cleared) and the new direction.
func (d *DataTableWidget) OnSortChange(fn func(col int, dir SortDirection)) *DataTableWidget {
	d.onSortChange = fn
	return d
}

// OnSelectionChange registers a callback fired whenever the selection changes,
// receiving the selected natural row indices in ascending order.
func (d *DataTableWidget) OnSelectionChange(fn func(selected []int)) *DataTableWidget {
	d.onSelectionChange = fn
	return d
}

// W pins the outer table width in px (otherwise it fills available width).
func (d *DataTableWidget) W(px float32) *DataTableWidget {
	d.width = px
	return d
}

// Theme pins a specific theme instead of the process-wide current theme.
func (d *DataTableWidget) Theme(th *theme.Theme) *DataTableWidget {
	if th != nil {
		d.theme = th
		if d.prevBtn != nil {
			d.prevBtn.Theme(th)
		}
		if d.nextBtn != nil {
			d.nextBtn.Theme(th)
		}
	}
	return d
}

// SortColumn presets the active sort (col index into cols, or -1) and direction
// without firing OnSortChange. Useful for goldens and controlled state.
func (d *DataTableWidget) SortColumn(col int, dir SortDirection) *DataTableWidget {
	if col < 0 || col >= len(d.cols) || dir == SortNone {
		d.sortCol = -1
		d.sortDir = SortNone
	} else {
		d.sortCol = col
		d.sortDir = dir
	}
	d.rebuildOrder()
	return d
}

// Select presets a row's selected state (by natural index) without firing
// OnSelectionChange. Useful for goldens and controlled state.
func (d *DataTableWidget) Select(row int, v bool) *DataTableWidget {
	if row < 0 || row >= d.rowCount {
		return d
	}
	if v {
		d.selected[row] = true
	} else {
		delete(d.selected, row)
	}
	return d
}

// Page presets the 0-based current page (clamped) without firing callbacks.
func (d *DataTableWidget) Page(p int) *DataTableWidget {
	if p < 0 {
		p = 0
	}
	if p > d.lastPage() {
		p = d.lastPage()
	}
	d.page = p
	d.buildFooter()
	return d
}

// SortState reports the active sort column index (-1 when none) and direction.
func (d *DataTableWidget) SortState() (int, SortDirection) { return d.sortCol, d.sortDir }

// SelectedRows returns the selected natural row indices in ascending order.
func (d *DataTableWidget) SelectedRows() []int {
	out := make([]int, 0, len(d.selected))
	for r := range d.selected {
		out = append(out, r)
	}
	sort.Ints(out)
	return out
}

// CurrentPage returns the 0-based current page.
func (d *DataTableWidget) CurrentPage() int { return d.page }

// VisibleRowCount returns the number of body rows drawn on the current page
// (the full row count when pagination is disabled).
func (d *DataTableWidget) VisibleRowCount() int { return d.visibleRowCount() }

// DisplayOrder returns the current display-position -> natural-row mapping
// (after sorting). The returned slice is a copy.
func (d *DataTableWidget) DisplayOrder() []int {
	out := make([]int, len(d.order))
	copy(out, d.order)
	return out
}

// ColumnLeft returns the widget-local x of data column ci's left edge (past the
// checkbox column when selectable). Valid after Layout. Returns 0 for an
// out-of-range index.
func (d *DataTableWidget) ColumnLeft(ci int) float32 {
	if ci < 0 || ci >= len(d.colWidths) {
		return 0
	}
	x := d.checkboxW
	for i := 0; i < ci; i++ {
		x += d.colWidths[i]
	}
	return x
}

// PrevBounds returns the Previous button's widget-local bounds (zero when not
// paginating). Valid after Layout.
func (d *DataTableWidget) PrevBounds() geometry.Rect {
	if d.prevBtn == nil {
		return geometry.Rect{}
	}
	return d.prevBtn.Bounds()
}

// NextBounds returns the Next button's widget-local bounds (zero when not
// paginating). Valid after Layout.
func (d *DataTableWidget) NextBounds() geometry.Rect {
	if d.nextBtn == nil {
		return geometry.Rect{}
	}
	return d.nextBtn.Bounds()
}

// rebuildOrder recomputes the display->natural index mapping from the active
// sort. With no sort it is the identity; with a sort it is a stable sort of the
// natural indices by the column's sort key.
func (d *DataTableWidget) rebuildOrder() {
	d.order = make([]int, d.rowCount)
	for i := range d.order {
		d.order[i] = i
	}
	if d.sortCol < 0 || d.sortCol >= len(d.cols) || d.sortDir == SortNone {
		return
	}
	key := d.cols[d.sortCol].sortKey()
	if key == nil {
		return
	}
	asc := d.sortDir == SortAsc
	sort.SliceStable(d.order, func(a, b int) bool {
		ka, kb := key(d.order[a]), key(d.order[b])
		if asc {
			return ka < kb
		}
		return ka > kb
	})
}

// cycleSort advances the sort state for column col: none→asc→desc→none.
// Selecting a different column starts it at asc.
func (d *DataTableWidget) cycleSort(col int) {
	if d.sortCol != col {
		d.sortCol = col
		d.sortDir = SortAsc
	} else {
		switch d.sortDir {
		case SortNone:
			d.sortDir = SortAsc
		case SortAsc:
			d.sortDir = SortDesc
		default: // SortDesc -> clear
			d.sortCol = -1
			d.sortDir = SortNone
		}
	}
	d.rebuildOrder()
	if d.onSortChange != nil {
		d.onSortChange(d.sortCol, d.sortDir)
	}
}

// pageRows returns the [start, end) display indices visible on the current
// page. Without pagination it spans all rows.
func (d *DataTableWidget) pageRows() (int, int) {
	if d.pageSize <= 0 {
		return 0, d.rowCount
	}
	start := d.page * d.pageSize
	end := start + d.pageSize
	if start > d.rowCount {
		start = d.rowCount
	}
	if end > d.rowCount {
		end = d.rowCount
	}
	return start, end
}

// visibleRowCount is the number of body rows drawn on the current page.
func (d *DataTableWidget) visibleRowCount() int {
	s, e := d.pageRows()
	return e - s
}

// lastPage returns the 0-based index of the final page.
func (d *DataTableWidget) lastPage() int {
	if d.pageSize <= 0 || d.rowCount == 0 {
		return 0
	}
	return (d.rowCount - 1) / d.pageSize
}

// allSelected reports whether every natural row is selected (drives the
// select-all header checkbox; empty tables are not "all selected").
func (d *DataTableWidget) allSelected() bool {
	return d.rowCount > 0 && len(d.selected) == d.rowCount
}

// someSelected reports whether at least one but not all rows are selected
// (drives the header checkbox's indeterminate/minus indicator).
func (d *DataTableWidget) someSelected() bool {
	return len(d.selected) > 0 && !d.allSelected()
}

// toggleRow flips the selected state of a natural row index and fires the
// selection callback.
func (d *DataTableWidget) toggleRow(row int) {
	if d.selected[row] {
		delete(d.selected, row)
	} else {
		d.selected[row] = true
	}
	d.fireSelection()
}

// toggleAll selects every row when not all are selected, otherwise clears the
// selection (matching shadcn's select-all box).
func (d *DataTableWidget) toggleAll() {
	if d.allSelected() {
		d.selected = make(map[int]bool)
	} else {
		d.selected = make(map[int]bool, d.rowCount)
		for r := 0; r < d.rowCount; r++ {
			d.selected[r] = true
		}
	}
	d.fireSelection()
}

func (d *DataTableWidget) fireSelection() {
	if d.onSelectionChange != nil {
		d.onSelectionChange(d.SelectedRows())
	}
}

// goToPage clamps and sets the current page, rebuilding the footer.
func (d *DataTableWidget) goToPage(p int) {
	if p < 0 || p > d.lastPage() || p == d.page {
		return
	}
	d.page = p
	d.buildFooter()
}

// buildFooter (re)creates the Previous/Next footer buttons for the current
// pagination state. Both are default-size outline buttons (shadcn uses
// variant="outline" size="sm"); Previous is disabled on the first page, Next on
// the last.
func (d *DataTableWidget) buildFooter() {
	if d.pageSize <= 0 {
		d.prevBtn = nil
		d.nextBtn = nil
		return
	}
	prev := Button("Previous").Outline().Theme(d.theme)
	prev.OnClick(func() { d.goToPage(d.page - 1) })
	if d.page <= 0 {
		prev.Disabled(true)
	}
	prev.SetParent(d)

	next := Button("Next").Outline().Theme(d.theme)
	next.OnClick(func() { d.goToPage(d.page + 1) })
	if d.page >= d.lastPage() {
		next.Disabled(true)
	}
	next.SetParent(d)

	d.prevBtn = prev
	d.nextBtn = next
}

// cellFontFamily resolves the family for a given weight, honoring a custom
// FontSans the same way the presentational Table does.
func (d *DataTableWidget) cellFontFamily(weight int) string {
	if d.theme.FontSans != theme.DefaultFontSans {
		return d.theme.FontSans
	}
	return fonts.Family(weight)
}

// measureColumns computes each data column's natural width: max over the header
// label (plus a sort-arrow allowance) and every cell of text width + padding.
func (d *DataTableWidget) measureColumns() []float32 {
	widths := make([]float32, len(d.cols))
	size := metrics.Table.FontSize
	pad := 2 * metrics.Table.CellPad
	for ci, col := range d.cols {
		hw := textmetrics.Width(d.cellFontFamily(metrics.Table.HeadFontWeight), size, col.Title) + pad
		if col.Sortable {
			hw += metrics.DataTable.SortIconGap + metrics.DataTable.SortIconSize
		}
		widths[ci] = hw
		if col.Cell == nil {
			continue
		}
		for r := 0; r < d.rowCount; r++ {
			w := textmetrics.Width(d.cellFontFamily(metrics.Table.CellFontWeight), size, col.Cell(r)) + pad
			if w > widths[ci] {
				widths[ci] = w
			}
		}
	}
	return widths
}

// Layout sizes the grid: a header row (h-10), up to pageSize body rows (p-2
// around a text-sm line), an optional pagination footer, and the leading
// checkbox column when selectable. Data columns are scaled to fill w-full.
func (d *DataTableWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	w := d.width
	if w <= 0 {
		w = c.MaxWidth
		if w <= 0 || w >= geometry.Infinity {
			w = 400
		}
	}

	d.checkboxW = 0
	if d.selectable {
		d.checkboxW = metrics.DataTable.CheckboxColWidth
	}

	natural := d.measureColumns()
	var sum float32
	for _, cw := range natural {
		sum += cw
	}
	dataW := w - d.checkboxW
	d.colWidths = make([]float32, len(natural))
	copy(d.colWidths, natural)
	if sum < dataW && len(natural) > 0 {
		extra := (dataW - sum) / float32(len(natural))
		for i := range d.colWidths {
			d.colWidths[i] += extra
		}
	} else if sum > dataW {
		w = sum + d.checkboxW // do not shrink below natural content width
	}

	d.rowH = 2*metrics.Table.CellPad + 20
	d.headTop = 0
	d.bodyTop = metrics.Table.HeadHeight
	bodyH := float32(d.visibleRowCount()) * d.rowH
	totalH := d.bodyTop + bodyH

	if d.pageSize > 0 {
		d.footerTop = totalH
		totalH += metrics.DataTable.FooterHeight
		d.layoutFooter(ctx, w)
	}

	size := c.Constrain(geometry.Sz(w, totalH))
	d.SetBounds(geometry.FromPointSize(d.Position(), size))
	return size
}

// layoutFooter positions the Previous/Next buttons at the right edge of the
// footer band, vertically centered.
func (d *DataTableWidget) layoutFooter(ctx widget.Context, w float32) {
	if d.prevBtn == nil || d.nextBtn == nil {
		return
	}
	loose := geometry.Loose(geometry.Sz(geometry.Infinity, metrics.DataTable.FooterHeight))
	ps := d.prevBtn.Layout(ctx, loose)
	ns := d.nextBtn.Layout(ctx, loose)
	nextX := w - ns.Width
	prevX := nextX - metrics.DataTable.FooterGap - ps.Width
	d.prevBtn.SetBounds(geometry.NewRect(prevX, d.footerTop+(metrics.DataTable.FooterHeight-ps.Height)/2, ps.Width, ps.Height))
	d.nextBtn.SetBounds(geometry.NewRect(nextX, d.footerTop+(metrics.DataTable.FooterHeight-ns.Height)/2, ns.Width, ns.Height))
}

// Draw paints the header (labels + sort arrows + select-all box), the visible
// body rows (cells, per-row selection box, hover/selected backgrounds), the
// row separators, and the pagination footer.
func (d *DataTableWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !d.IsVisible() {
		return
	}
	th := d.theme
	tok := th.Active()
	bounds := d.Bounds()
	size := metrics.Table.FontSize

	d.drawHeader(canvas, th, tok, bounds, size)
	d.drawBody(canvas, th, tok, bounds, size)
	d.drawSeparators(canvas, bounds, tok)
	if d.pageSize > 0 {
		d.drawFooter(ctx, canvas, tok, bounds)
	}
}

// drawHeader draws the header row: optional select-all checkbox, then each
// column's label and (for sortable columns) a sort-direction arrow.
func (d *DataTableWidget) drawHeader(canvas widget.Canvas, th *theme.Theme, tok *theme.Tokens, bounds geometry.Rect, size float32) {
	rowTop := bounds.Min.Y + d.headTop
	family := d.cellFontFamily(metrics.Table.HeadFontWeight)

	var x float32
	if d.selectable {
		d.drawCheckbox(canvas, th, tok,
			geometry.NewRect(bounds.Min.X, rowTop, d.checkboxW, metrics.Table.HeadHeight),
			d.allSelected(), d.someSelected())
		x += d.checkboxW
	}
	for ci, col := range d.cols {
		if ci >= len(d.colWidths) {
			break
		}
		colW := d.colWidths[ci]
		pad := metrics.Table.HeadPadX
		textRect := geometry.NewRect(bounds.Min.X+x+pad, rowTop, colW-2*pad, metrics.Table.HeadHeight)
		drawTableText(canvas, col.Title, textRect, family, size, tok.Foreground, metrics.Table.HeadFontWeight, col.Align)

		if col.Sortable && ci == d.sortCol && d.sortDir != SortNone {
			d.drawSortArrow(canvas, tok, family, size, bounds.Min.X+x, colW, rowTop, col)
		}
		x += colW
	}
}

// drawSortArrow draws the active sort indicator (up for asc, down for desc).
// For left/center-aligned columns it sits just after the label (shadcn's inline
// sort button reads label + arrow); for right-aligned columns the label hugs
// the cell's right edge, so the arrow precedes it at the same trailing edge.
func (d *DataTableWidget) drawSortArrow(canvas widget.Canvas, tok *theme.Tokens, family string, size, cellX, colW, rowTop float32, col DataColumn) {
	ic := icons.ChevronUp
	if d.sortDir == SortDesc {
		ic = icons.ChevronDown
	}
	sz := metrics.DataTable.SortIconSize
	pad := metrics.Table.HeadPadX
	gap := metrics.DataTable.SortIconGap
	labelW := textmetrics.Width(family, size, col.Title)

	var ax float32
	switch col.Align {
	case widget.TextAlignRight:
		// Label is right-aligned; place the arrow just left of it.
		ax = cellX + colW - pad - labelW - gap - sz
		if ax < cellX+pad {
			ax = cellX + pad
		}
	case widget.TextAlignCenter:
		ax = cellX + (colW+labelW)/2 + gap
	default:
		ax = cellX + pad + labelW + gap
	}
	// Clamp inside the cell's trailing padding.
	if max := cellX + colW - pad - sz; ax > max {
		ax = max
	}
	ay := rowTop + (metrics.Table.HeadHeight-sz)/2
	icon.Draw(canvas, ic, geometry.NewRect(ax, ay, sz, sz), tok.Foreground)
}

// drawBody draws the visible page of body rows: backgrounds (selected then
// hover), the per-row selection checkbox, and the data cells.
func (d *DataTableWidget) drawBody(canvas widget.Canvas, th *theme.Theme, tok *theme.Tokens, bounds geometry.Rect, size float32) {
	start, end := d.pageRows()
	family := d.cellFontFamily(metrics.Table.CellFontWeight)

	for disp := start; disp < end; disp++ {
		natural := d.order[disp]
		localIdx := disp - start
		rowTop := bounds.Min.Y + d.bodyTop + float32(localIdx)*d.rowH
		rowRect := geometry.NewRect(bounds.Min.X, rowTop, bounds.Width(), d.rowH)

		// Selected background wins over hover (data-[state=selected]:bg-muted).
		switch {
		case d.selected[natural]:
			canvas.DrawRect(rowRect, draw.Alpha(tok.Muted, metrics.DataTable.SelectedRowAlpha))
		case localIdx == d.hoverRow:
			canvas.DrawRect(rowRect, draw.Alpha(tok.Muted, metrics.Table.RowHoverAlpha))
		}

		var x float32
		if d.selectable {
			d.drawCheckbox(canvas, th, tok,
				geometry.NewRect(bounds.Min.X, rowTop, d.checkboxW, d.rowH),
				d.selected[natural], false)
			x += d.checkboxW
		}
		for ci, col := range d.cols {
			if ci >= len(d.colWidths) {
				break
			}
			colW := d.colWidths[ci]
			pad := metrics.Table.CellPad
			text := ""
			if col.Cell != nil {
				text = col.Cell(natural)
			}
			textRect := geometry.NewRect(bounds.Min.X+x+pad, rowTop, colW-2*pad, d.rowH)
			drawTableText(canvas, text, textRect, family, size, tok.Foreground, metrics.Table.CellFontWeight, col.Align)
			x += colW
		}
	}
}

// drawCheckbox draws a Checkbox-styled box centered in cellRect, matching the
// shadcn Checkbox visuals (16px box, rounded-[4px], Primary fill when marked,
// Check or Minus indicator). It is drawn directly (no child widget) so the
// table owns its hit testing.
func (d *DataTableWidget) drawCheckbox(canvas widget.Canvas, th *theme.Theme, tok *theme.Tokens, cellRect geometry.Rect, checked, indeterminate bool) {
	boxSize := metrics.Checkbox.Size
	box := geometry.NewRect(
		cellRect.Min.X+(cellRect.Width()-boxSize)/2,
		cellRect.Min.Y+(cellRect.Height()-boxSize)/2,
		boxSize, boxSize,
	)
	radius := metrics.Checkbox.Radius
	marked := checked || indeterminate

	draw.Shadow(canvas, box, radius, metrics.ShadowXS)

	boxBg := tok.Background
	switch {
	case marked:
		boxBg = tok.Primary
	case th.IsDark():
		boxBg = draw.MulAlpha(tok.Input, metrics.Checkbox.DarkFillAlpha)
	}

	borderColor := tok.Input
	if marked {
		borderColor = tok.Primary
	}
	// BorderFill (fill + ring) instead of an inside stroke, which renders as
	// a solid gray box on the GPU.
	draw.BorderFill(canvas, box, boxBg, borderColor, radius, metrics.Checkbox.BorderWidth)

	if marked {
		ic := icons.Check
		if indeterminate {
			ic = icons.Minus
		}
		sz := metrics.Checkbox.IconSize
		center := box.Center()
		icon.Draw(canvas, ic, geometry.NewRect(center.X-sz/2, center.Y-sz/2, sz, sz), tok.PrimaryForeground)
	}
}

// drawSeparators draws the 1px border-b under the header and under every body
// row except the last visible one (matching the Table's [&_tr:last-child]:border-0).
func (d *DataTableWidget) drawSeparators(canvas widget.Canvas, bounds geometry.Rect, tok *theme.Tokens) {
	bw := metrics.Table.BorderWidth
	drawHLine(canvas, bounds, d.headTop+metrics.Table.HeadHeight-bw, tok.Border, bw)
	n := d.visibleRowCount()
	for i := 0; i < n-1; i++ {
		y := d.bodyTop + float32(i+1)*d.rowH - bw
		drawHLine(canvas, bounds, y, tok.Border, bw)
	}
}

// drawFooter draws the "N of M row(s) selected." caption (when selectable) and
// the Previous/Next page buttons.
func (d *DataTableWidget) drawFooter(ctx widget.Context, canvas widget.Canvas, tok *theme.Tokens, bounds geometry.Rect) {
	if d.selectable {
		caption := strconv.Itoa(len(d.selected)) + " of " + strconv.Itoa(d.rowCount) + " row(s) selected."
		capRect := geometry.NewRect(
			bounds.Min.X+metrics.Table.CellPad,
			bounds.Min.Y+d.footerTop,
			bounds.Width(),
			metrics.DataTable.FooterHeight,
		)
		drawTableText(canvas, caption, capRect, d.cellFontFamily(metrics.Table.CellFontWeight),
			metrics.DataTable.FooterFontSize, tok.MutedForeground, metrics.Table.CellFontWeight, widget.TextAlignLeft)
	}

	canvas.PushTransform(bounds.Min)
	for _, b := range []*ButtonWidget{d.prevBtn, d.nextBtn} {
		if b == nil {
			continue
		}
		widget.StampScreenOrigin(b, canvas)
		widget.DrawChild(b, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event routes mouse input to the slot under the pointer: header sort cells,
// the select-all / per-row checkbox cells, and the footer Previous/Next
// buttons. Nothing is broadcast — at most one slot handles each event.
func (d *DataTableWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	bounds := d.Bounds()

	// Footer buttons get group-local coordinates and own hover/cursor.
	if d.pageSize > 0 {
		local := *me
		local.Position = me.Position.Sub(bounds.Min)
		for _, b := range []*ButtonWidget{d.prevBtn, d.nextBtn} {
			if b != nil && b.Bounds().Contains(local.Position) {
				return b.Event(ctx, &local)
			}
		}
	}

	switch me.MouseType {
	case event.MouseLeave:
		if d.hoverRow != -1 {
			d.hoverRow = -1
			d.invalidate(ctx)
		}
		return true
	case event.MouseMove, event.MouseEnter:
		return d.hover(ctx, me, bounds)
	case event.MouseRelease:
		if me.Button == event.ButtonLeft {
			return d.click(ctx, me, bounds)
		}
	}
	return false
}

// hover updates the hovered body-row highlight as the pointer moves over the
// grid body.
func (d *DataTableWidget) hover(ctx widget.Context, me *event.MouseEvent, bounds geometry.Rect) bool {
	localY := me.Position.Y - bounds.Min.Y
	newHover := -1
	n := d.visibleRowCount()
	if localY >= d.bodyTop && localY < d.bodyTop+float32(n)*d.rowH {
		newHover = int((localY - d.bodyTop) / d.rowH)
	}
	if newHover != d.hoverRow {
		d.hoverRow = newHover
		d.invalidate(ctx)
	}
	return newHover != -1
}

// click dispatches a left-release: a header click cycles the sort of a sortable
// column (or toggles select-all over the checkbox column), and a body click on
// the checkbox column toggles that row's selection.
func (d *DataTableWidget) click(ctx widget.Context, me *event.MouseEvent, bounds geometry.Rect) bool {
	localX := me.Position.X - bounds.Min.X
	localY := me.Position.Y - bounds.Min.Y

	// Header row.
	if localY >= d.headTop && localY < d.headTop+metrics.Table.HeadHeight {
		if d.selectable && localX < d.checkboxW {
			d.toggleAll()
			d.invalidate(ctx)
			return true
		}
		if ci := d.columnAt(localX); ci >= 0 && d.cols[ci].Sortable {
			d.cycleSort(ci)
			d.invalidate(ctx)
			return true
		}
		return false
	}

	// Body rows: only the checkbox column is interactive.
	if d.selectable && localX < d.checkboxW {
		start, end := d.pageRows()
		n := end - start
		if localY >= d.bodyTop && localY < d.bodyTop+float32(n)*d.rowH {
			localIdx := int((localY - d.bodyTop) / d.rowH)
			natural := d.order[start+localIdx]
			d.toggleRow(natural)
			d.invalidate(ctx)
			return true
		}
	}
	return false
}

// columnAt maps a local x (already past the checkbox column when selectable) to
// a data column index, or -1 when outside the columns.
func (d *DataTableWidget) columnAt(localX float32) int {
	x := d.checkboxW
	if localX < x {
		return -1
	}
	for ci := range d.colWidths {
		if localX >= x && localX < x+d.colWidths[ci] {
			return ci
		}
		x += d.colWidths[ci]
	}
	return -1
}

func (d *DataTableWidget) invalidate(ctx widget.Context) {
	d.SetNeedsRedraw(true)
	ctx.InvalidateRect(d.Bounds())
}

// Children returns the pagination footer buttons (nil when not paginating).
func (d *DataTableWidget) Children() []widget.Widget {
	if d.prevBtn == nil || d.nextBtn == nil {
		return nil
	}
	return []widget.Widget{d.prevBtn, d.nextBtn}
}

// Compile-time interface check.
var _ widget.Widget = (*DataTableWidget)(nil)

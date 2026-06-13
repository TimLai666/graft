package graft

import (
	"strconv"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// PaginationWidget is shadcn's Pagination: a row of page-number buttons flanked
// by Previous/Next controls, with ellipsis gaps for elided ranges
// (docs/research/03-shadcn-pixel-spec.md §5 "Pagination").
//
// WRAP DECISION (DESIGN.md §3.2): graft-OWNED composite over graft.Button. Page
// links are ghost icon buttons (36²), the active page is an outline button, and
// Previous/Next are default-size ghost buttons carrying a chevron plus a label.
// Ellipsis gaps are non-interactive 36px boxes drawn directly. The widget lays
// the buttons out in a gap-1 row and routes their clicks to OnSelect.
type PaginationWidget struct {
	widget.WidgetBase

	theme    *theme.Theme
	total    int
	current  int
	onSelect func(int)

	// items is the resolved render sequence (prev, pages/ellipses, next),
	// rebuilt on demand.
	items []paginationItem
}

// paginationItemKind selects what a pagination slot renders.
type paginationItemKind uint8

const (
	paginationPrev paginationItemKind = iota
	paginationNext
	paginationPage
	paginationEllipsis
)

// paginationItem is one slot in the pagination row.
type paginationItem struct {
	kind   paginationItemKind
	page   int // 1-based page number for paginationPage
	button *ButtonWidget
	x      float32 // local x of the slot, set in Layout
}

// Pagination creates an empty pagination control. Use Pages to populate it.
func Pagination() *PaginationWidget {
	p := &PaginationWidget{theme: CurrentTheme(), current: 1}
	p.SetVisible(true)
	p.SetEnabled(true)
	return p
}

// Pages configures the control to show pages 1..total with current selected.
func (p *PaginationWidget) Pages(total, current int) *PaginationWidget {
	if total < 1 {
		total = 1
	}
	if current < 1 {
		current = 1
	}
	if current > total {
		current = total
	}
	p.total = total
	p.current = current
	p.build()
	return p
}

// OnSelect registers a callback fired with the 1-based page the user picks
// (including Previous/Next).
func (p *PaginationWidget) OnSelect(fn func(int)) *PaginationWidget {
	p.onSelect = fn
	return p
}

// Theme pins a specific theme.
func (p *PaginationWidget) Theme(th *theme.Theme) *PaginationWidget {
	if th != nil {
		p.theme = th
		p.build()
	}
	return p
}

// build constructs the button row for the current total/current state.
func (p *PaginationWidget) build() {
	p.items = p.items[:0]

	prev := Button("Previous").Ghost().Icon(icons.ChevronLeft).Theme(p.theme)
	prev.Style(func(s *Style) {
		padX := float32(10) // px-2.5
		s.PadX = &padX
	})
	target := p.current - 1
	prev.OnClick(func() { p.selectPage(target) })
	if p.current <= 1 {
		prev.Disabled(true)
	}
	p.items = append(p.items, paginationItem{kind: paginationPrev, button: prev})

	for _, page := range p.pageNumbers() {
		if page == ellipsisMarker {
			p.items = append(p.items, paginationItem{kind: paginationEllipsis})
			continue
		}
		// shadcn page links use size="icon" (36² square), but that size renders
		// icon-only in graft.Button; instead use the default size pinned to a
		// 36px-wide square so the page number text still draws.
		b := Button(strconv.Itoa(page)).Theme(p.theme).W(metrics.Button.Icon.Height)
		if page == p.current {
			b.Outline()
		} else {
			b.Ghost()
		}
		pg := page
		b.OnClick(func() { p.selectPage(pg) })
		p.items = append(p.items, paginationItem{kind: paginationPage, page: page, button: b})
	}

	next := Button("Next").Ghost().Theme(p.theme)
	next.Style(func(s *Style) {
		padX := float32(10)
		s.PadX = &padX
	})
	// DEVIATION: shadcn places the chevron-right after the "Next" label, but
	// graft.Button supports a leading icon only. A leading chevron-right still
	// reads naturally as "next"; a trailing-icon Button is out of scope here.
	next.Icon(icons.ChevronRight)
	nt := p.current + 1
	next.OnClick(func() { p.selectPage(nt) })
	if p.current >= p.total {
		next.Disabled(true)
	}
	p.items = append(p.items, paginationItem{kind: paginationNext, button: next})

	for _, it := range p.items {
		if it.button != nil {
			it.button.SetParent(p)
		}
	}
}

// ellipsisMarker is the sentinel page value standing for an elided range.
const ellipsisMarker = -1

// pageNumbers returns the page sequence with ellipsis markers for elision. For
// small totals every page is shown; for large totals it shows
// 1 … current-1 current current+1 … total.
func (p *PaginationWidget) pageNumbers() []int {
	if p.total <= 7 {
		out := make([]int, p.total)
		for i := range out {
			out[i] = i + 1
		}
		return out
	}
	out := []int{1}
	start := p.current - 1
	end := p.current + 1
	if start < 2 {
		start = 2
	}
	if end > p.total-1 {
		end = p.total - 1
	}
	if start > 2 {
		out = append(out, ellipsisMarker)
	}
	for i := start; i <= end; i++ {
		out = append(out, i)
	}
	if end < p.total-1 {
		out = append(out, ellipsisMarker)
	}
	out = append(out, p.total)
	return out
}

// selectPage clamps and fires OnSelect, then rebuilds for the new current.
func (p *PaginationWidget) selectPage(page int) {
	if page < 1 || page > p.total {
		return
	}
	p.current = page
	if p.onSelect != nil {
		p.onSelect(page)
	}
	p.build()
	p.SetNeedsRedraw(true)
}

// Layout places the items in a gap-1 row, all sharing the 36px height.
func (p *PaginationWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	gap := metrics.Pagination.Gap
	loose := c.Loosen()
	var x, height float32
	for i := range p.items {
		it := &p.items[i]
		if i > 0 {
			x += gap
		}
		it.x = x
		var w float32
		switch it.kind {
		case paginationEllipsis:
			w = metrics.Pagination.EllipsisBox
			if metrics.Pagination.EllipsisBox > height {
				height = metrics.Pagination.EllipsisBox
			}
		default:
			sz := it.button.Layout(ctx, loose)
			w = sz.Width
			if sz.Height > height {
				height = sz.Height
			}
			it.button.SetBounds(geometry.NewRect(x, 0, sz.Width, sz.Height))
		}
		x += w
	}

	size := c.Constrain(geometry.Sz(x, height))
	// Vertically center each button within the row height.
	for _, it := range p.items {
		if it.button == nil {
			continue
		}
		bb := it.button.Bounds()
		it.button.SetBounds(geometry.NewRect(bb.Min.X, (size.Height-bb.Height())/2, bb.Width(), bb.Height()))
	}
	p.SetBounds(geometry.FromPointSize(p.Position(), size))
	return size
}

// Draw paints the buttons and the ellipsis boxes.
func (p *PaginationWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !p.IsVisible() {
		return
	}
	tok := p.theme.Active()
	bounds := p.Bounds()

	canvas.PushTransform(bounds.Min)
	defer canvas.PopTransform()

	for _, it := range p.items {
		switch it.kind {
		case paginationEllipsis:
			box := metrics.Pagination.EllipsisBox
			iconSize := metrics.Pagination.EllipsisIcon
			iconRect := geometry.NewRect(
				it.x+(box-iconSize)/2,
				(bounds.Height()-iconSize)/2,
				iconSize, iconSize)
			icon.Draw(canvas, icons.Ellipsis, iconRect, tok.MutedForeground)
		default:
			widget.StampScreenOrigin(it.button, canvas)
			widget.DrawChild(it.button, ctx, canvas)
		}
	}
}

// Event forwards mouse input to the button under the pointer with
// pagination-local coordinates. Only the hit button receives the event: a
// MousePress must press and focus exactly one button, not the whole row (the
// shared-event broadcast pressed every button and left focus on the last one).
func (p *PaginationWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	local := *me
	local.Position = me.Position.Sub(p.Bounds().Min)
	for i := len(p.items) - 1; i >= 0; i-- {
		it := p.items[i]
		if it.button == nil || !it.button.Bounds().Contains(local.Position) {
			continue
		}
		return it.button.Event(ctx, &local)
	}
	return false
}

// Children returns the page/prev/next buttons.
func (p *PaginationWidget) Children() []widget.Widget {
	out := make([]widget.Widget, 0, len(p.items))
	for _, it := range p.items {
		if it.button != nil {
			out = append(out, it.button)
		}
	}
	return out
}

// Compile-time interface check.
var _ widget.Widget = (*PaginationWidget)(nil)

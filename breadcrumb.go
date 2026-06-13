package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// BreadcrumbWidget is shadcn's Breadcrumb: a horizontal trail of links ending
// in the current page, with chevron-right separators auto-inserted between
// items (docs/research/03-shadcn-pixel-spec.md §5 "Breadcrumb").
//
// WRAP DECISION (DESIGN.md §3.2): graft-OWNED composite. It is a single-line
// row of text leaves plus separator icons drawn directly; links highlight to
// Foreground on hover, the page renders in Foreground, and an ellipsis item
// renders as a 36px box with the ellipsis glyph. All colors resolve from the
// active token set at draw time.
type BreadcrumbWidget struct {
	widget.WidgetBase

	items []*BreadcrumbItemDef
	theme *theme.Theme

	// hoverItem is the index of the hovered link, or -1.
	hoverItem int

	// item layout cache (x offsets and widths), rebuilt each Layout.
	itemRects []geometry.Rect
}

// breadcrumbKind selects how an item renders.
type breadcrumbKind uint8

const (
	breadcrumbLink breadcrumbKind = iota
	breadcrumbPage
	breadcrumbEllipsis
)

// BreadcrumbItemDef is one breadcrumb element.
type BreadcrumbItemDef struct {
	kind    breadcrumbKind
	text    string
	onClick func()
}

// BreadcrumbLink creates a clickable trail link (muted, hover Foreground).
func BreadcrumbLink(text string) *BreadcrumbItemDef {
	return &BreadcrumbItemDef{kind: breadcrumbLink, text: text}
}

// BreadcrumbPage creates the current-page item (non-interactive Foreground).
func BreadcrumbPage(text string) *BreadcrumbItemDef {
	return &BreadcrumbItemDef{kind: breadcrumbPage, text: text}
}

// BreadcrumbEllipsis creates a collapsed-segments ellipsis item.
func BreadcrumbEllipsis() *BreadcrumbItemDef {
	return &BreadcrumbItemDef{kind: breadcrumbEllipsis}
}

// OnClick attaches a click handler to a link item.
func (i *BreadcrumbItemDef) OnClick(fn func()) *BreadcrumbItemDef {
	i.onClick = fn
	return i
}

// Breadcrumb creates a breadcrumb trail from its items, inserting chevron
// separators between them.
func Breadcrumb(items ...*BreadcrumbItemDef) *BreadcrumbWidget {
	b := &BreadcrumbWidget{items: items, theme: CurrentTheme(), hoverItem: -1}
	b.SetVisible(true)
	b.SetEnabled(true)
	return b
}

// Theme pins a specific theme.
func (b *BreadcrumbWidget) Theme(th *theme.Theme) *BreadcrumbWidget {
	if th != nil {
		b.theme = th
	}
	return b
}

func (b *BreadcrumbWidget) fontFamily() string {
	if b.theme.FontSans != theme.DefaultFontSans {
		return b.theme.FontSans
	}
	return fonts.Family(metrics.Breadcrumb.LinkFontWeight)
}

// itemWidth returns the natural width of an item.
func (b *BreadcrumbWidget) itemWidth(it *BreadcrumbItemDef) float32 {
	if it.kind == breadcrumbEllipsis {
		return metrics.Breadcrumb.EllipsisBox
	}
	return textmetrics.Width(b.fontFamily(), metrics.Breadcrumb.FontSize, it.text)
}

// Layout lays the items and separators out in a single row.
func (b *BreadcrumbWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	gap := metrics.Breadcrumb.Gap
	sepW := metrics.Breadcrumb.SeparatorSize
	// The row height is the text-sm line box (20px); the ellipsis box overlaps
	// it and centers its glyph vertically.
	rowH := float32(20)

	b.itemRects = b.itemRects[:0]
	var x float32
	for i, it := range b.items {
		w := b.itemWidth(it)
		b.itemRects = append(b.itemRects, geometry.NewRect(x, 0, w, rowH))
		x += w
		if i < len(b.items)-1 {
			x += gap + sepW + gap
		}
	}

	size := c.Constrain(geometry.Sz(x, rowH))
	b.SetBounds(geometry.FromPointSize(b.Position(), size))
	return size
}

// Draw paints the items (link/page/ellipsis) and the chevron separators.
func (b *BreadcrumbWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !b.IsVisible() {
		return
	}
	tok := b.theme.Active()
	bounds := b.Bounds()
	gap := metrics.Breadcrumb.Gap
	sepW := metrics.Breadcrumb.SeparatorSize
	size := metrics.Breadcrumb.FontSize
	family := b.fontFamily()

	for i, it := range b.items {
		r := b.itemRects[i]
		itemRect := geometry.NewRect(bounds.Min.X+r.Min.X, bounds.Min.Y, r.Width(), r.Height())

		switch it.kind {
		case breadcrumbLink:
			col := tok.MutedForeground
			if i == b.hoverItem {
				col = tok.Foreground
			}
			breadcrumbText(canvas, it.text, itemRect, family, size, col)
		case breadcrumbPage:
			breadcrumbText(canvas, it.text, itemRect, family, size, tok.Foreground)
		case breadcrumbEllipsis:
			iconSize := metrics.Breadcrumb.EllipsisIcon
			iconRect := geometry.NewRect(
				itemRect.Min.X+(itemRect.Width()-iconSize)/2,
				itemRect.Min.Y+(itemRect.Height()-iconSize)/2,
				iconSize, iconSize)
			icon.Draw(canvas, icons.Ellipsis, iconRect, tok.MutedForeground)
		}

		// Separator (chevron-right) after every item except the last.
		if i < len(b.items)-1 {
			sepX := bounds.Min.X + r.Max.X + gap
			sepRect := geometry.NewRect(
				sepX,
				bounds.Min.Y+(r.Height()-sepW)/2,
				sepW, sepW)
			icon.Draw(canvas, icons.ChevronRight, sepRect, tok.MutedForeground)
		}
	}
}

// breadcrumbText draws vertically-centered breadcrumb text.
func breadcrumbText(canvas widget.Canvas, text string, bounds geometry.Rect, family string, size float32, col widget.Color) {
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(text, bounds, widget.TextStyle{FontFamily: family, FontSize: size, Color: col, Align: widget.TextAlignLeft})
		return
	}
	canvas.DrawText(text, bounds, size, col, false, widget.TextAlignLeft)
}

// Event tracks link hover for the Foreground highlight and fires link clicks.
func (b *BreadcrumbWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	bounds := b.Bounds()
	switch me.MouseType {
	case event.MouseLeave:
		if b.hoverItem != -1 {
			b.hoverItem = -1
			b.invalidate(ctx)
		}
		ctx.SetCursor(widget.CursorDefault)
		return true
	case event.MouseMove, event.MouseEnter:
		local := me.Position.Sub(bounds.Min)
		newHover := -1
		for i, it := range b.items {
			if it.kind == breadcrumbLink && b.itemRects[i].Contains(local) {
				newHover = i
				break
			}
		}
		if newHover != b.hoverItem {
			b.hoverItem = newHover
			b.invalidate(ctx)
		}
		if newHover != -1 {
			ctx.SetCursor(widget.CursorPointer)
		} else {
			ctx.SetCursor(widget.CursorDefault)
		}
		return newHover != -1
	case event.MousePress:
		if me.Button != event.ButtonLeft {
			return false
		}
		local := me.Position.Sub(bounds.Min)
		for i, it := range b.items {
			if it.kind == breadcrumbLink && b.itemRects[i].Contains(local) {
				if it.onClick != nil {
					it.onClick()
				}
				return true
			}
		}
	}
	return false
}

func (b *BreadcrumbWidget) invalidate(ctx widget.Context) {
	b.SetNeedsRedraw(true)
	ctx.InvalidateRect(b.Bounds())
}

// Children returns nil; the breadcrumb draws its items directly.
func (b *BreadcrumbWidget) Children() []widget.Widget { return nil }

// Compile-time interface check.
var _ widget.Widget = (*BreadcrumbWidget)(nil)

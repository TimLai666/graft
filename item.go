package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// itemVariant selects the Item background/border chrome.
type itemVariant uint8

const (
	itemDefault itemVariant = iota // transparent border, no fill
	itemOutline                    // border-border
	itemMuted                      // bg-muted/50
)

// ItemWidget is shadcn's Item: a flexible horizontal content row used for
// settings rows, list rows, and link cards. Layout is
// [ItemMedia] [ItemContent] [ItemActions] in a vertically-centered flex
// with gap-4 (gap-2.5 sm), p-4 (py-3 px-4 sm) padding, rounded-md corners,
// text-sm, and a 1px border (transparent by default).
//
// Architecture: graft-owned widget, exactly like CardWidget. Layout composes
// a primitives.HBox for the media/content/actions stack, but the Item chrome
// (background, radius, border) is drawn by ItemWidget itself so every color
// resolves from the active token set at draw time — primitives.Box caches its
// Background color and would go stale on a light/dark mode switch.
//
// The shadcn [a&]:hover:bg-accent/50 hover state (which applies only when the
// Item is rendered as a link) is deferred: v1 renders the static variants
// only, keeping ItemWidget a clean Card-like static surface.
type ItemWidget struct {
	widget.WidgetBase

	box     *primitives.BoxWidget
	variant itemVariant
	sm      bool
	theme   *theme.Theme
}

// Item creates a content row around the given slots (ItemMedia, ItemContent,
// ItemActions, or any widget). Width is natural unless W is set. The default
// variant draws a transparent border with no fill.
//
// An ItemContent child is wrapped in primitives.Expanded so it takes the
// remaining row width (flex-1), pushing any trailing ItemActions to the right
// edge when the Item has a fixed width.
func Item(children ...Widget) *ItemWidget {
	i := &ItemWidget{}
	row := make([]Widget, len(children))
	for n, ch := range children {
		if c, ok := ch.(*ItemContentWidget); ok {
			row[n] = primitives.Expanded(c)
			continue
		}
		row[n] = ch
	}
	i.box = primitives.HBox(row...).
		Padding(metrics.ItemPadDefault).
		Gap(metrics.ItemGapDefault).
		CrossAlign(primitives.CrossAxisCenter)
	i.box.SetParent(i)
	i.SetVisible(true)
	i.SetEnabled(true)
	return i
}

// Outline selects the outline variant: the 1px border uses the Border token.
func (i *ItemWidget) Outline() *ItemWidget {
	i.variant = itemOutline
	return i
}

// Muted selects the muted variant: the background fills with the Muted token
// at 50% (bg-muted/50).
func (i *ItemWidget) Muted() *ItemWidget {
	i.variant = itemMuted
	return i
}

// Sm selects the small size: py-3 px-4 padding and a gap-2.5 row gap.
func (i *ItemWidget) Sm() *ItemWidget {
	i.sm = true
	i.box.PaddingXY(metrics.ItemPadXSm, metrics.ItemPadYSm).Gap(metrics.ItemGapSm)
	return i
}

// W pins the item's outer width in px (border included, like CSS
// border-box).
func (i *ItemWidget) W(px float32) *ItemWidget {
	inner := px - 2*metrics.ItemBorderWidth
	if inner < 0 {
		inner = 0
	}
	i.box.Width(inner)
	return i
}

// Theme pins a specific theme instead of the process-wide current theme.
func (i *ItemWidget) Theme(th *theme.Theme) *ItemWidget {
	i.theme = th
	return i
}

func (i *ItemWidget) resolvedTheme() *theme.Theme {
	if i.theme != nil {
		return i.theme
	}
	return CurrentTheme()
}

// Layout sizes the slot row inside the 1px border (CSS border-box) and adds
// the border back to the outer size, mirroring CardWidget.
func (i *ItemWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	bw := metrics.ItemBorderWidth
	inner := cons.Loosen().Deflate(geometry.UniformInsets(bw))
	size := i.box.Layout(ctx, inner)
	i.box.SetBounds(geometry.FromPointSize(geometry.Pt(bw, bw), size))

	outer := cons.Constrain(geometry.Sz(size.Width+2*bw, size.Height+2*bw))
	i.SetBounds(geometry.FromPointSize(i.Position(), outer))
	return outer
}

// Draw paints the variant fill, the slot row, and the 1px inside border,
// resolving all colors from the active token set so mode switches repaint.
func (i *ItemWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !i.IsVisible() {
		return
	}
	th := i.resolvedTheme()
	tok := th.Active()
	bounds := i.Bounds()
	radius := th.RadiusMD() // rounded-md

	if i.variant == itemMuted {
		canvas.DrawRoundRect(bounds, draw.MulAlpha(tok.Muted, metrics.ItemMutedBgAlpha), radius) // bg-muted/50
	}

	// Outline border drawn UNDER the content: BorderFill (fill = page
	// background) instead of an inside stroke, which renders as a solid gray
	// box on the GPU. Its opaque inset fill would clobber the slot row, so it
	// must precede DrawChild.
	if border := i.borderColor(tok); border.A > 0 {
		draw.BorderFill(canvas, bounds, tok.Background, border, radius, metrics.ItemBorderWidth)
	}

	canvas.PushTransform(bounds.Min)
	widget.StampScreenOrigin(i.box, canvas)
	widget.DrawChild(i.box, ctx, canvas)
	canvas.PopTransform()
}

// borderColor resolves the 1px border color: Border token for the outline
// variant, transparent otherwise (default/muted use border-transparent).
func (i *ItemWidget) borderColor(tok *theme.Tokens) widget.Color {
	if i.variant == itemOutline {
		return tok.Border
	}
	return widget.Color{}
}

// Event forwards input to the slot row with item-local coordinates.
func (i *ItemWidget) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.MouseEvent:
		local := *ev
		local.Position = ev.Position.Sub(i.Bounds().Min)
		if !i.box.Bounds().Contains(local.Position) {
			return false
		}
		return i.box.Event(ctx, &local)
	case *event.WheelEvent:
		local := *ev
		local.Position = ev.Position.Sub(i.Bounds().Min)
		return i.box.Event(ctx, &local)
	default:
		return i.box.Event(ctx, e)
	}
}

// Children returns the internal slot row.
func (i *ItemWidget) Children() []widget.Widget { return []widget.Widget{i.box} }

// ── ItemGroup ──────────────────────────────────────────────────────────────

// ItemGroupWidget is a vertical stack of Items (flex flex-col), optionally
// interleaved with ItemSeparator rules. It is a thin named wrapper over
// primitives.Box; all Box builder methods (Gap, Width, ...) remain available.
type ItemGroupWidget struct {
	*primitives.BoxWidget
}

// ItemGroup stacks Items vertically. Insert ItemSeparator children between
// rows for divided lists; the group stretches its children to a common width.
func ItemGroup(children ...Widget) *ItemGroupWidget {
	return &ItemGroupWidget{
		primitives.VBox(children...).CrossAlign(primitives.CrossAxisStretch),
	}
}

// ItemSeparator returns a horizontal 1px Border-token rule for use between
// Items in an ItemGroup (shadcn's ItemSeparator reuses the Separator look).
func ItemSeparator() *SeparatorWidget {
	return Separator()
}

// ── ItemMedia ──────────────────────────────────────────────────────────────

// itemMediaVariant selects the ItemMedia chrome.
type itemMediaVariant uint8

const (
	itemMediaDefault itemMediaVariant = iota // plain, just holds the icon/child
	itemMediaIcon                            // 32px bordered Muted icon chip
	itemMediaImage                           // 40px rounded clipped image box
)

// ItemMediaWidget is the leading media slot of an Item (shrink-0, centered).
// In the default variant it merely centers its icon/child; the icon variant
// draws a 32px square Muted chip with a 1px border and rounded-sm corners
// behind a 16px icon; the image variant draws a 40px rounded-sm clipped box.
//
// Architecture: graft-owned widget so the chip background/border (icon
// variant) resolve from the active token set at draw time, and so the fixed
// 32/40px square sizes are enforced regardless of the child's intrinsic size.
type ItemMediaWidget struct {
	widget.WidgetBase

	child   Widget
	ic      *icon.IconData
	variant itemMediaVariant
	theme   *theme.Theme

	iconRect geometry.Rect // icon bounds in media-local coords (zero if no icon)
}

// ItemMedia creates the leading media slot from a lucide icon (the common
// case, matching graft.ItemMedia(icons.Circle).Icon()). In the default variant
// the icon renders at 20px in the MutedForeground token; the icon-chip variant
// (.Icon()) draws it at 16px inside a 32px Muted chip. Use ItemMediaChild to
// wrap an arbitrary widget instead.
func ItemMedia(ic icon.IconData) *ItemMediaWidget {
	m := &ItemMediaWidget{ic: &ic}
	m.SetVisible(true)
	m.SetEnabled(true)
	return m
}

// ItemMediaChild wraps an arbitrary widget (an avatar, image widget, custom
// control, ...) as the leading media slot for maximum flexibility. The default
// variant centers the child as-is; call Icon or Image for the chip /
// clipped-image variants.
func ItemMediaChild(child Widget) *ItemMediaWidget {
	m := &ItemMediaWidget{child: child}
	if child != nil {
		if ps, ok := child.(interface{ SetParent(widget.Widget) }); ok {
			ps.SetParent(m)
		}
	}
	m.SetVisible(true)
	m.SetEnabled(true)
	return m
}

// Icon selects the icon-chip variant: a 32px square Muted chip with a 1px
// border and rounded-sm corners drawn behind a 16px icon (or the media child).
func (m *ItemMediaWidget) Icon() *ItemMediaWidget {
	m.variant = itemMediaIcon
	return m
}

// Image selects the image variant: a 40px square rounded-sm clipped box.
func (m *ItemMediaWidget) Image() *ItemMediaWidget {
	m.variant = itemMediaImage
	return m
}

// Theme pins a specific theme instead of the process-wide current theme.
func (m *ItemMediaWidget) Theme(th *theme.Theme) *ItemMediaWidget {
	m.theme = th
	return m
}

func (m *ItemMediaWidget) resolvedTheme() *theme.Theme {
	if m.theme != nil {
		return m.theme
	}
	return CurrentTheme()
}

// fixedSize returns the enforced square side length for the variant, or 0 for
// the default variant (which sizes to its icon/child).
func (m *ItemMediaWidget) fixedSize() float32 {
	switch m.variant {
	case itemMediaIcon:
		return metrics.ItemIconChipSize
	case itemMediaImage:
		return metrics.ItemImageSize
	default:
		return 0
	}
}

// defaultIconSize is the icon side length for the plain default media variant
// (no chip): shadcn renders default ItemMedia svgs at size-5 (20px).
const defaultIconSize float32 = 20

// Layout sizes the media slot: a fixed square for the icon/image variants
// (with the icon/child centered inside), or the icon/child's natural size for
// the default variant.
func (m *ItemMediaWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	loose := geometry.Loose(geometry.Sz(geometry.Infinity, geometry.Infinity))

	var childSize geometry.Size
	if m.child != nil {
		childSize = m.child.Layout(ctx, loose)
	}

	side := m.fixedSize()
	var size geometry.Size
	switch {
	case side > 0:
		size = geometry.Sz(side, side)
	case m.child != nil:
		size = childSize
	case m.ic != nil:
		size = geometry.Sz(defaultIconSize, defaultIconSize)
	}
	size = c.Constrain(size)
	m.SetBounds(geometry.FromPointSize(m.Position(), size))

	// Center the icon and/or child inside the media box.
	m.iconRect = geometry.Rect{}
	if m.ic != nil {
		s := defaultIconSize
		if m.variant == itemMediaIcon {
			s = metrics.ItemMediaIconSize
		}
		m.iconRect = geometry.NewRect((size.Width-s)/2, (size.Height-s)/2, s, s)
	}
	if m.child != nil {
		cx := (size.Width - childSize.Width) / 2
		cy := (size.Height - childSize.Height) / 2
		if sb, ok := m.child.(interface{ SetBounds(geometry.Rect) }); ok {
			sb.SetBounds(geometry.FromPointSize(geometry.Pt(cx, cy), childSize))
		}
	}
	return size
}

// Draw paints the variant chrome (icon chip: Muted fill + 1px Border inside
// border; image: clipped rounded box), then the centered icon and/or child,
// resolving every color from the active token set.
func (m *ItemMediaWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !m.IsVisible() {
		return
	}
	th := m.resolvedTheme()
	tok := th.Active()
	bounds := m.Bounds()

	switch m.variant {
	case itemMediaIcon:
		radius := th.RadiusSM() // rounded-sm
		draw.BorderFill(canvas, bounds, tok.Muted, tok.Border, radius, metrics.ItemBorderWidth)
	case itemMediaImage:
		radius := th.RadiusSM() // rounded-sm
		canvas.DrawRoundRect(bounds, tok.Muted, radius)
	}

	canvas.PushTransform(bounds.Min)
	if m.ic != nil {
		icon.Draw(canvas, *m.ic, m.iconRect, tok.MutedForeground)
	}
	if m.child != nil {
		widget.StampScreenOrigin(m.child, canvas)
		widget.DrawChild(m.child, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event forwards input to the child with media-local coordinates.
func (m *ItemMediaWidget) Event(ctx widget.Context, e event.Event) bool {
	if m.child == nil {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		local := *ev
		local.Position = ev.Position.Sub(m.Bounds().Min)
		return m.child.Event(ctx, &local)
	default:
		return m.child.Event(ctx, e)
	}
}

// Children returns the wrapped media child (nil for bare icon media).
func (m *ItemMediaWidget) Children() []widget.Widget {
	if m.child == nil {
		return nil
	}
	return []widget.Widget{m.child}
}

// ── ItemContent ────────────────────────────────────────────────────────────

// ItemContentWidget is the flex-1 text column of an Item (flex-col, gap-1).
// The Item constructor wraps it in primitives.Expanded so it takes the
// remaining row width. It is a thin named wrapper over primitives.Box.
type ItemContentWidget struct {
	*primitives.BoxWidget
}

// ItemContent creates the text column: a vertical stack (gap-1 between title
// and description) that expands to fill the remaining row width.
func ItemContent(children ...Widget) *ItemContentWidget {
	return &ItemContentWidget{
		primitives.VBox(children...).
			Gap(metrics.ItemContentGap).
			CrossAlign(primitives.CrossAxisStart),
	}
}

// ItemTitle creates the Item title: 14px / weight 500 / leading-snug, in the
// Foreground token (text-sm font-medium).
func ItemTitle(text string) *TypographyWidget {
	return styled(text, metrics.ItemTitleFontSize, metrics.ItemTitleFontWeight, metrics.ItemTitleLineHeight)
}

// ItemDescription creates the Item description: 14px muted-foreground,
// leading-normal (text-sm text-muted-foreground font-normal).
func ItemDescription(text string) *TypographyWidget {
	return styled(text, metrics.ItemDescriptionFontSize, metrics.ItemDescriptionFontWeight, metrics.ItemDescriptionLineHeight).Muted()
}

// ── ItemActions ────────────────────────────────────────────────────────────

// ItemActionsWidget is the trailing slot of an Item (shrink-0, horizontal,
// gap-2), holding buttons or other trailing controls. It is a thin named
// wrapper over primitives.Box.
type ItemActionsWidget struct {
	*primitives.BoxWidget
}

// ItemActions creates the trailing action row: a horizontal stack (gap-2),
// vertically centered, that does not grow (shrink-0).
func ItemActions(children ...Widget) *ItemActionsWidget {
	return &ItemActionsWidget{
		primitives.HBox(children...).
			Gap(metrics.ItemActionsGap).
			CrossAlign(primitives.CrossAxisCenter),
	}
}

var (
	_ widget.Widget = (*ItemWidget)(nil)
	_ widget.Widget = (*ItemGroupWidget)(nil)
	_ widget.Widget = (*ItemMediaWidget)(nil)
	_ widget.Widget = (*ItemContentWidget)(nil)
	_ widget.Widget = (*ItemActionsWidget)(nil)
)

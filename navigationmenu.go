package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// NavigationMenuWidget is the shadcn Navigation Menu: a horizontal row of
// top-level items. Each item is either a plain link or a trigger (label + a
// small chevron) that opens a floating content panel — the "viewport" — below
// the bar. The panel is a rounded-lg popover surface holding grouped links /
// rich content.
//
// Clicking a trigger opens its panel; clicking the open trigger again closes
// it; hovering an adjacent trigger while one is open switches to that panel.
// Outside click or Escape dismisses (non-modal overlay). The bar itself has no
// surface chrome — the root list is a transparent flex row.
type NavigationMenuWidget struct {
	widget.WidgetBase

	items []*navMenuItemWidget
	theme *theme.Theme

	openIndex      int // -1 = none open
	overlayContent *navMenuOverlayContent
}

// NavigationMenu builds a horizontal navigation bar from one or more items
// (NavigationMenuItem for dropdown triggers, NavigationMenuLinkItem for plain
// links).
func NavigationMenu(items ...*navMenuItemWidget) *NavigationMenuWidget {
	n := &NavigationMenuWidget{items: items, openIndex: -1, theme: CurrentTheme()}
	n.SetVisible(true)
	n.SetEnabled(true)
	for i, it := range items {
		it.owner = n
		it.index = i
		n.AddChild(it)
	}
	return n
}

// Theme pins a specific theme instead of the process-wide current theme.
func (n *NavigationMenuWidget) Theme(th *theme.Theme) *NavigationMenuWidget {
	n.theme = th
	return n
}

func (n *NavigationMenuWidget) resolvedTheme() *theme.Theme {
	if n.theme != nil {
		return n.theme
	}
	return CurrentTheme()
}

// Layout arranges the top-level items left-to-right with gap-1 spacing, sizing
// the bar to h-9 and the sum of item widths + gaps.
func (n *NavigationMenuWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	y := n.Position().Y
	rowCons := geometry.Loose(geometry.Sz(10000, metrics.NavigationMenuHeight))

	cursorX := n.Position().X
	for i, it := range n.items {
		if i > 0 {
			cursorX += metrics.NavigationMenuGap
		}
		sz := it.Layout(ctx, rowCons)
		it.SetBounds(geometry.FromPointSize(geometry.Pt(cursorX, y), geometry.Sz(sz.Width, metrics.NavigationMenuHeight)))
		cursorX += sz.Width
	}

	totalW := cursorX - n.Position().X
	size := c.Constrain(geometry.Sz(totalW, metrics.NavigationMenuHeight))
	n.SetBounds(geometry.FromPointSize(n.Position(), size))
	return size
}

// Draw paints the top-level items (the bar has no surface chrome).
func (n *NavigationMenuWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !n.IsVisible() {
		return
	}
	for i, it := range n.items {
		it.open = i == n.openIndex
		it.Draw(ctx, canvas)
	}
}

// Event forwards to the items.
func (n *NavigationMenuWidget) Event(ctx widget.Context, e event.Event) bool {
	for _, it := range n.items {
		if it.Event(ctx, e) {
			return true
		}
	}
	return false
}

// Children returns the top-level items.
func (n *NavigationMenuWidget) Children() []widget.Widget {
	out := make([]widget.Widget, len(n.items))
	for i, it := range n.items {
		out[i] = it
	}
	return out
}

// toggle opens item i, closes it if already open, or switches from another.
func (n *NavigationMenuWidget) toggle(ctx widget.Context, i int) {
	if n.openIndex == i {
		n.closePanel(ctx)
		return
	}
	if n.openIndex >= 0 {
		n.removeOverlay(ctx)
	}
	n.openIndex = i
	n.openPanel(ctx, i)
}

func (n *NavigationMenuWidget) openPanel(ctx widget.Context, i int) {
	it := n.items[i]
	if it.content == nil {
		return
	}
	om := ctx.OverlayManager()
	if om == nil {
		return
	}
	th := n.resolvedTheme()
	panel := newNavMenuPanel(th, it.content)

	anchor := it.ScreenBounds()
	size := panel.ContentSize()
	pos := overlay.Position(overlay.PlacementBelow, anchor, size, ctx.WindowSize(), metrics.NavigationMenuSideOffset)
	panel.SetBounds(geometry.FromPointSize(pos, size))

	oc := &navMenuOverlayContent{panel: panel, onDismiss: func() { n.closePanel(ctx) }}
	n.overlayContent = oc
	om.PushOverlay(oc, func() { n.handleDismiss(ctx) })
	n.SetNeedsRedraw(true)
	ctx.Invalidate()
}

func (n *NavigationMenuWidget) closePanel(ctx widget.Context) {
	n.removeOverlay(ctx)
	n.openIndex = -1
	n.SetNeedsRedraw(true)
	ctx.Invalidate()
}

func (n *NavigationMenuWidget) removeOverlay(ctx widget.Context) {
	if n.overlayContent == nil {
		return
	}
	if om := ctx.OverlayManager(); om != nil {
		om.RemoveOverlay(n.overlayContent)
	}
	n.overlayContent = nil
}

// handleDismiss reacts to an overlay-driven dismissal (outside click/Escape).
func (n *NavigationMenuWidget) handleDismiss(ctx widget.Context) {
	n.overlayContent = nil
	n.openIndex = -1
	n.SetNeedsRedraw(true)
	ctx.Invalidate()
}

// --- top-level item (link or trigger) ---------------------------------------

// navMenuItemWidget is a single top-level item: a plain link or a trigger that
// opens a content panel.
type navMenuItemWidget struct {
	widget.WidgetBase

	label   string
	isLink  bool
	onPress func()

	content *NavigationMenuContentDef
	theme   *theme.Theme

	owner *NavigationMenuWidget
	index int
	open  bool
	hover bool
}

// NavigationMenuTriggerDef is the data for a top-level trigger (label + a
// chevron). Pair it with NavigationMenuContent via NavigationMenuItem.
type NavigationMenuTriggerDef struct {
	label string
}

// NavigationMenuTrigger creates a top-level trigger labeled label. It opens
// the paired NavigationMenuContent panel.
func NavigationMenuTrigger(label string) *NavigationMenuTriggerDef {
	return &NavigationMenuTriggerDef{label: label}
}

// NavigationMenuItem builds a top-level dropdown item: the trigger label plus
// its content panel.
func NavigationMenuItem(trigger *NavigationMenuTriggerDef, content *NavigationMenuContentDef) *navMenuItemWidget {
	label := ""
	if trigger != nil {
		label = trigger.label
	}
	it := &navMenuItemWidget{label: label, content: content}
	it.SetVisible(true)
	it.SetEnabled(true)
	return it
}

// NavigationMenuLinkItem builds a top-level plain link (no content panel, no
// chevron). The optional onPress callback fires on click.
func NavigationMenuLinkItem(label string, onPress ...func()) *navMenuItemWidget {
	it := &navMenuItemWidget{label: label, isLink: true}
	if len(onPress) > 0 {
		it.onPress = onPress[0]
	}
	it.SetVisible(true)
	it.SetEnabled(true)
	return it
}

// Theme pins a specific theme for this item.
func (it *navMenuItemWidget) Theme(th *theme.Theme) *navMenuItemWidget {
	it.theme = th
	return it
}

func (it *navMenuItemWidget) resolvedTheme() *theme.Theme {
	if it.theme != nil {
		return it.theme
	}
	if it.owner != nil {
		return it.owner.resolvedTheme()
	}
	return CurrentTheme()
}

func (it *navMenuItemWidget) triggerFamily(th *theme.Theme) string {
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(metrics.NavigationMenuTriggerFontWeight)
}

// Layout sizes the item to its label plus px-4 padding (and the chevron slot
// for triggers).
func (it *navMenuItemWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	th := it.resolvedTheme()
	family := it.triggerFamily(th)
	w := textmetrics.Width(family, metrics.NavigationMenuTriggerFontSize, it.label) + 2*metrics.NavigationMenuTriggerPadX
	if !it.isLink {
		w += metrics.NavigationMenuChevronGap + metrics.NavigationMenuChevronSize
	}
	size := c.Constrain(geometry.Sz(w, metrics.NavigationMenuHeight))
	it.SetBounds(geometry.FromPointSize(it.Position(), size))
	return size
}

// Draw paints the item: an accent fill + accent-foreground text when open or
// hovered, otherwise transparent with foreground text. Triggers append a
// chevron.
func (it *navMenuItemWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !it.IsVisible() {
		return
	}
	th := it.resolvedTheme()
	tok := th.Active()
	b := it.Bounds()

	textColor := tok.Foreground
	if it.open || it.hover {
		fill := tok.Accent
		if it.open && !it.hover {
			fill = draw.MulAlpha(tok.Accent, 0.5) // data-[state=open]:bg-accent/50
		}
		canvas.DrawRoundRect(b, fill, th.RadiusMD())
		textColor = tok.AccentForeground
	}

	labelRect := geometry.NewRect(
		b.Min.X+metrics.NavigationMenuTriggerPadX,
		b.Min.Y,
		b.Width()-2*metrics.NavigationMenuTriggerPadX,
		b.Height(),
	)
	drawNavMenuText(canvas, metrics.NavigationMenuTriggerFontWeight, metrics.NavigationMenuTriggerFontSize, it.label, labelRect, textColor)

	if !it.isLink {
		family := it.triggerFamily(th)
		labelW := textmetrics.Width(family, metrics.NavigationMenuTriggerFontSize, it.label)
		chevX := b.Min.X + metrics.NavigationMenuTriggerPadX + labelW + metrics.NavigationMenuChevronGap
		chevRect := geometry.NewRect(
			chevX,
			b.Center().Y-metrics.NavigationMenuChevronSize/2,
			metrics.NavigationMenuChevronSize,
			metrics.NavigationMenuChevronSize,
		)
		icon.Draw(canvas, icons.ChevronDown, chevRect, textColor)
	}
}

// Event toggles the panel on a left press inside a trigger (or fires the link
// callback); hover updates the accent fill and drives hover-switch.
func (it *navMenuItemWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	inside := it.Bounds().Contains(me.Position)
	switch me.MouseType {
	case event.MouseMove, event.MouseEnter:
		if inside != it.hover {
			it.hover = inside
			if inside {
				ctx.SetCursor(widget.CursorPointer)
			}
			it.SetNeedsRedraw(true)
			ctx.Invalidate()
		}
		// Hover-switch: when a panel is open and the cursor enters a
		// different trigger, switch to it.
		if inside && !it.isLink && it.owner != nil && it.owner.openIndex >= 0 && it.owner.openIndex != it.index {
			it.owner.toggle(ctx, it.index)
		}
	case event.MouseLeave:
		if it.hover {
			it.hover = false
			ctx.SetCursor(widget.CursorDefault)
			it.SetNeedsRedraw(true)
			ctx.Invalidate()
		}
	case event.MousePress:
		if inside && me.Button == event.ButtonLeft {
			if it.isLink {
				if it.onPress != nil {
					it.onPress()
				}
				return true
			}
			if it.owner != nil {
				it.owner.toggle(ctx, it.index)
				return true
			}
		}
	}
	return false
}

// Children returns nil; the trigger draws its own label and the panel lives in
// the overlay.
func (it *navMenuItemWidget) Children() []widget.Widget { return nil }

// --- content definition (data) ----------------------------------------------

// NavigationMenuContentDef holds the children for one trigger's content panel
// (data only); the live panel is built lazily into the overlay on open.
type NavigationMenuContentDef struct {
	children []widget.Widget
}

// NavigationMenuContent collects the panel's children (link rows via
// NavigationMenuLink, or any other widget).
func NavigationMenuContent(children ...widget.Widget) *NavigationMenuContentDef {
	return &NavigationMenuContentDef{children: children}
}

// NavigationMenuLinkDef is a content link row: a font-medium title with an
// optional muted-foreground description.
type NavigationMenuLinkDef struct {
	widget.WidgetBase

	title   string
	desc    string
	onPress func()
	theme   *theme.Theme

	hover bool
}

// NavigationMenuLink creates a content link row labeled label.
func NavigationMenuLink(label string) *NavigationMenuLinkDef {
	l := &NavigationMenuLinkDef{title: label}
	l.SetVisible(true)
	l.SetEnabled(true)
	return l
}

// Description adds a muted-foreground description line under the title.
func (l *NavigationMenuLinkDef) Description(desc string) *NavigationMenuLinkDef {
	l.desc = desc
	return l
}

// OnPress registers a click callback.
func (l *NavigationMenuLinkDef) OnPress(fn func()) *NavigationMenuLinkDef {
	l.onPress = fn
	return l
}

// Theme pins a specific theme for this link row.
func (l *NavigationMenuLinkDef) Theme(th *theme.Theme) *NavigationMenuLinkDef {
	l.theme = th
	return l
}

func (l *NavigationMenuLinkDef) resolvedTheme() *theme.Theme {
	if l.theme != nil {
		return l.theme
	}
	return CurrentTheme()
}

func (l *NavigationMenuLinkDef) rowHeight() float32 {
	h := 2*metrics.NavigationMenuLinkPad + metrics.NavigationMenuLinkTitleLineHeight
	if l.desc != "" {
		h += metrics.NavigationMenuLinkDescGap + metrics.NavigationMenuLinkDescLineHeight
	}
	return h
}

func (l *NavigationMenuLinkDef) naturalWidth() float32 {
	th := l.resolvedTheme()
	titleFamily := fonts.Family(metrics.NavigationMenuLinkTitleWeight)
	descFamily := fonts.Family(400)
	if th.FontSans != theme.DefaultFontSans {
		titleFamily = th.FontSans
		descFamily = th.FontSans
	}
	w := textmetrics.Width(titleFamily, metrics.NavigationMenuLinkTitleSize, l.title)
	if dw := textmetrics.Width(descFamily, metrics.NavigationMenuLinkDescSize, l.desc); dw > w {
		w = dw
	}
	return w + 2*metrics.NavigationMenuLinkPad
}

// Layout sizes the link row to its natural height (the panel sets the width).
func (l *NavigationMenuLinkDef) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	w := c.MaxWidth
	if w <= 0 {
		w = l.naturalWidth()
	}
	size := c.Constrain(geometry.Sz(w, l.rowHeight()))
	l.SetBounds(geometry.FromPointSize(l.Position(), size))
	return size
}

// Draw paints the link row: an accent fill on hover, the title, and the
// optional muted description.
func (l *NavigationMenuLinkDef) Draw(_ widget.Context, canvas widget.Canvas) {
	if !l.IsVisible() {
		return
	}
	th := l.resolvedTheme()
	tok := th.Active()
	b := l.Bounds()

	titleColor := tok.PopoverForeground
	descColor := tok.MutedForeground
	if l.hover {
		canvas.DrawRoundRect(b, tok.Accent, th.RadiusSM())
		titleColor = tok.AccentForeground
		descColor = tok.AccentForeground
	}

	titleRect := geometry.NewRect(
		b.Min.X+metrics.NavigationMenuLinkPad,
		b.Min.Y+metrics.NavigationMenuLinkPad,
		b.Width()-2*metrics.NavigationMenuLinkPad,
		metrics.NavigationMenuLinkTitleLineHeight,
	)
	drawNavMenuText(canvas, metrics.NavigationMenuLinkTitleWeight, metrics.NavigationMenuLinkTitleSize, l.title, titleRect, titleColor)

	if l.desc != "" {
		descRect := geometry.NewRect(
			b.Min.X+metrics.NavigationMenuLinkPad,
			titleRect.Max.Y+metrics.NavigationMenuLinkDescGap,
			b.Width()-2*metrics.NavigationMenuLinkPad,
			metrics.NavigationMenuLinkDescLineHeight,
		)
		drawNavMenuText(canvas, 400, metrics.NavigationMenuLinkDescSize, l.desc, descRect, descColor)
	}
}

// Event highlights the row on hover and fires the callback on click.
func (l *NavigationMenuLinkDef) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	inside := l.Bounds().Contains(me.Position)
	switch me.MouseType {
	case event.MouseMove, event.MouseEnter:
		if inside != l.hover {
			l.hover = inside
			if inside {
				ctx.SetCursor(widget.CursorPointer)
			}
			l.SetNeedsRedraw(true)
			ctx.InvalidateRect(l.Bounds())
		}
		return inside
	case event.MouseLeave:
		if l.hover {
			l.hover = false
			ctx.SetCursor(widget.CursorDefault)
			l.SetNeedsRedraw(true)
			ctx.InvalidateRect(l.Bounds())
		}
	case event.MousePress:
		if inside && me.Button == event.ButtonLeft {
			if l.onPress != nil {
				l.onPress()
			}
			return true
		}
	}
	return false
}

// Children returns nil; the link draws its own title/description.
func (l *NavigationMenuLinkDef) Children() []widget.Widget { return nil }

// --- content panel (live, in overlay) ---------------------------------------

// navMenuPanel is the floating content panel ("viewport"): a rounded-lg
// popover surface (shadow-md, 1px border, bg-popover) that stacks its children
// vertically with p-2 padding. Children are typically NavigationMenuLink rows
// but may be any widget.
type navMenuPanel struct {
	widget.WidgetBase

	th       *theme.Theme
	children []widget.Widget
}

func newNavMenuPanel(th *theme.Theme, def *NavigationMenuContentDef) *navMenuPanel {
	p := &navMenuPanel{th: th}
	if def != nil {
		p.children = def.children
	}
	p.SetVisible(true)
	p.SetEnabled(true)
	return p
}

func (p *navMenuPanel) resolvedTheme() *theme.Theme {
	if p.th != nil {
		return p.th
	}
	return CurrentTheme()
}

// childNaturalSize returns a child's natural size: link rows report their own,
// others are measured loosely.
func (p *navMenuPanel) childNaturalSize(ctx widget.Context, ch widget.Widget) geometry.Size {
	if l, ok := ch.(*NavigationMenuLinkDef); ok {
		return geometry.Sz(l.naturalWidth(), l.rowHeight())
	}
	return ch.Layout(ctx, geometry.Loose(geometry.Sz(10000, 10000)))
}

// ContentSize returns the panel's natural size for the current children.
func (p *navMenuPanel) ContentSize() geometry.Size {
	pad := metrics.NavigationMenuContentPad
	h := 2 * pad
	maxW := float32(0)
	for _, ch := range p.children {
		sz := p.childNaturalSize(nil, ch)
		h += sz.Height
		if sz.Width > maxW {
			maxW = sz.Width
		}
	}
	w := maxW + 2*pad
	if w < metrics.NavigationMenuContentMinWidth {
		w = metrics.NavigationMenuContentMinWidth
	}
	return geometry.Sz(w, h)
}

// Layout sizes the panel to its content and stacks children top-to-bottom
// inside the p-2 inset, stretching each to the inner width.
func (p *navMenuPanel) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := p.ContentSize()
	if c.MinWidth > size.Width {
		size.Width = c.MinWidth
	}
	size = c.Constrain(size)
	p.SetBounds(geometry.FromPointSize(p.Position(), size))

	pad := metrics.NavigationMenuContentPad
	innerW := size.Width - 2*pad
	y := p.Position().Y + pad
	x := p.Position().X + pad
	for _, ch := range p.children {
		setWidgetBounds(ch, geometry.NewRect(x, y, innerW, 0))
		sz := ch.Layout(ctx, geometry.Tight(geometry.Sz(innerW, p.childNaturalSize(ctx, ch).Height)))
		setWidgetBounds(ch, geometry.NewRect(x, y, innerW, sz.Height))
		y += sz.Height
	}
	return size
}

// Draw renders the panel surface then its children.
func (p *navMenuPanel) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !p.IsVisible() {
		return
	}
	th := p.resolvedTheme()
	tok := th.Active()
	b := p.Bounds()
	radius := th.RadiusLG() // content: rounded-lg

	draw.Shadow(canvas, b, radius, metrics.ShadowMD)
	canvas.DrawRoundRect(b, tok.Popover, radius)
	draw.InsideBorder(canvas, b, radius, tok.Border, metrics.NavigationMenuBorderWidth)

	canvas.PushClip(b)
	defer canvas.PopClip()
	for _, ch := range p.children {
		ch.Draw(ctx, canvas)
	}
}

// Event routes to children; consumes mouse moves inside the panel.
func (p *navMenuPanel) Event(ctx widget.Context, e event.Event) bool {
	for _, ch := range p.children {
		if ch.Event(ctx, e) {
			return true
		}
	}
	if me, ok := e.(*event.MouseEvent); ok && p.Bounds().Contains(me.Position) {
		return true
	}
	return false
}

// Children returns the panel's content widgets.
func (p *navMenuPanel) Children() []widget.Widget { return p.children }

// --- overlay wrapper --------------------------------------------------------

// navMenuOverlayContent hosts a content panel in the overlay. It is non-modal:
// clicks outside the panel and Escape dismiss it. The panel is pre-positioned
// (bounds set before push); this wrapper preserves those bounds and routes
// events to the panel.
type navMenuOverlayContent struct {
	widget.WidgetBase

	panel     *navMenuPanel
	onDismiss func()
}

func (o *navMenuOverlayContent) Layout(_ widget.Context, _ geometry.Constraints) geometry.Size {
	o.SetBounds(o.panel.Bounds())
	return o.panel.Bounds().Size()
}

func (o *navMenuOverlayContent) Draw(ctx widget.Context, canvas widget.Canvas) {
	o.panel.Draw(ctx, canvas)
}

func (o *navMenuOverlayContent) Event(ctx widget.Context, e event.Event) bool {
	if o.panel.Event(ctx, e) {
		return true
	}
	if ke, ok := e.(*event.KeyEvent); ok {
		if ke.KeyType == event.KeyPress && ke.Key == event.KeyEscape {
			o.Dismiss()
			return true
		}
	}
	if me, ok := e.(*event.MouseEvent); ok && me.MouseType == event.MousePress {
		if !o.panel.Bounds().Contains(me.Position) {
			o.Dismiss()
			return true
		}
	}
	return false
}

func (o *navMenuOverlayContent) Dismiss() {
	if o.onDismiss != nil {
		o.onDismiss()
	}
}

func (o *navMenuOverlayContent) Modal() bool { return false }

func (o *navMenuOverlayContent) Children() []widget.Widget {
	return []widget.Widget{o.panel}
}

// --- preview (goldens / docs) -----------------------------------------------

// NavigationMenuMenuPreview returns one trigger's content panel rendered as
// direct content for goldens/docs (the path that works without an
// OverlayManager). It pads the frame so the panel's shadow-md is captured.
func NavigationMenuMenuPreview(content *NavigationMenuContentDef) Widget {
	panel := newNavMenuPanel(CurrentTheme(), content)
	return &navMenuPanelPreview{panel: panel}
}

// navMenuPanelPreview lays out a content panel at its natural size with shadow
// padding so the golden does not clip it.
type navMenuPanelPreview struct {
	widget.WidgetBase
	panel *navMenuPanel
}

func (p *navMenuPanelPreview) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	inner := p.panel.ContentSize()
	pad := float32(16)
	size := c.Constrain(geometry.Sz(inner.Width+2*pad, inner.Height+2*pad))
	p.SetBounds(geometry.FromPointSize(p.Position(), size))
	p.panel.SetBounds(geometry.NewRect(p.Position().X+pad, p.Position().Y+pad, inner.Width, inner.Height))
	p.panel.Layout(ctx, geometry.Tight(inner))
	return size
}

func (p *navMenuPanelPreview) Draw(ctx widget.Context, canvas widget.Canvas) {
	p.panel.Draw(ctx, canvas)
}

func (p *navMenuPanelPreview) Event(widget.Context, event.Event) bool { return false }
func (p *navMenuPanelPreview) Children() []widget.Widget              { return nil }

// drawNavMenuText draws a label, honoring StyledTextDrawer with a mock-canvas
// fallback (bold at weight >= 600).
func drawNavMenuText(canvas widget.Canvas, weight int, size float32, text string, bounds geometry.Rect, color widget.Color) {
	if text == "" {
		return
	}
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(text, bounds, widget.TextStyle{
			FontFamily: fonts.Family(weight),
			FontSize:   size,
			Color:      color,
			Align:      widget.TextAlignLeft,
		})
		return
	}
	canvas.DrawText(text, bounds, size, color, weight >= 600, widget.TextAlignLeft)
}

// Compile-time interface checks.
var (
	_ widget.Widget   = (*NavigationMenuWidget)(nil)
	_ widget.Widget   = (*navMenuItemWidget)(nil)
	_ widget.Widget   = (*NavigationMenuLinkDef)(nil)
	_ widget.Widget   = (*navMenuPanel)(nil)
	_ widget.Widget   = (*navMenuOverlayContent)(nil)
	_ overlay.Overlay = (*navMenuOverlayContent)(nil)
	_ widget.Widget   = (*navMenuPanelPreview)(nil)
)

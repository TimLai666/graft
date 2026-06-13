package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/internal/widgets/menu"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// MenubarWidget is the shadcn Menubar: a horizontal bar of menu triggers,
// each opening a dropdown panel below it. It reuses the shared menu engine
// (internal/widgets/menu) for the panels and the same MenuEntry types as
// DropdownMenu/ContextMenu; the bar chrome and trigger row are graft-owned.
//
// Clicking a trigger opens its panel; clicking the open trigger again closes
// it; clicking a different trigger switches to that menu. Outside click or
// Escape dismisses (non-modal overlay).
type MenubarWidget struct {
	widget.WidgetBase

	menus []*MenubarMenuWidget
	theme *theme.Theme

	openIndex      int // -1 = none open
	overlayContent *menuOverlayContent
}

// Menubar builds a horizontal menu bar from one or more MenubarMenu entries.
func Menubar(menus ...*MenubarMenuWidget) *MenubarWidget {
	m := &MenubarWidget{menus: menus, openIndex: -1, theme: CurrentTheme()}
	m.SetVisible(true)
	m.SetEnabled(true)
	for i, mm := range menus {
		mm.owner = m
		mm.index = i
		m.AddChild(mm)
	}
	return m
}

// Theme pins a specific theme instead of the process-wide current theme.
func (m *MenubarWidget) Theme(th *theme.Theme) *MenubarWidget {
	m.theme = th
	return m
}

func (m *MenubarWidget) resolvedTheme() *theme.Theme {
	if m.theme != nil {
		return m.theme
	}
	return CurrentTheme()
}

// Layout arranges the triggers left-to-right inside the bar's padding, sizing
// the bar to h-9 and the sum of trigger widths + gaps + padding.
func (m *MenubarWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	pad := metrics.MenubarPadding
	x := m.Position().X + pad
	y := m.Position().Y + pad
	innerH := metrics.MenubarHeight - 2*pad
	rowCons := geometry.Loose(geometry.Sz(10000, innerH))

	cursorX := x
	for i, mm := range m.menus {
		if i > 0 {
			cursorX += metrics.MenubarGap
		}
		sz := mm.Layout(ctx, rowCons)
		mm.SetBounds(geometry.FromPointSize(geometry.Pt(cursorX, y), geometry.Sz(sz.Width, innerH)))
		cursorX += sz.Width
	}

	totalW := (cursorX - x) + 2*pad
	size := c.Constrain(geometry.Sz(totalW, metrics.MenubarHeight))
	m.SetBounds(geometry.FromPointSize(m.Position(), size))
	return size
}

// Draw paints the bar surface (bg-background, 1px border, rounded-md,
// shadow-xs) then the triggers.
func (m *MenubarWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !m.IsVisible() {
		return
	}
	th := m.resolvedTheme()
	tok := th.Active()
	b := m.Bounds()
	radius := th.RadiusMD()

	draw.Shadow(canvas, b, radius, metrics.ShadowXS)
	canvas.DrawRoundRect(b, tok.Background, radius)
	draw.InsideBorder(canvas, b, radius, tok.Border, metrics.MenubarBorderWidth)

	for i, mm := range m.menus {
		mm.open = i == m.openIndex
		mm.Draw(ctx, canvas)
	}
}

// Event forwards to the triggers.
func (m *MenubarWidget) Event(ctx widget.Context, e event.Event) bool {
	for _, mm := range m.menus {
		if mm.Event(ctx, e) {
			return true
		}
	}
	return false
}

// Children returns the triggers.
func (m *MenubarWidget) Children() []widget.Widget {
	out := make([]widget.Widget, len(m.menus))
	for i, mm := range m.menus {
		out[i] = mm
	}
	return out
}

// toggle opens menu i, closes it if already open, or switches from another.
func (m *MenubarWidget) toggle(ctx widget.Context, i int) {
	if m.openIndex == i {
		m.closeMenu(ctx)
		return
	}
	if m.openIndex >= 0 {
		m.removeOverlay(ctx)
	}
	m.openIndex = i
	m.openMenu(ctx, i)
}

func (m *MenubarWidget) openMenu(ctx widget.Context, i int) {
	mm := m.menus[i]
	if mm.content == nil {
		return
	}
	om := ctx.OverlayManager()
	if om == nil {
		return
	}
	th := m.resolvedTheme()
	panel := mm.content.buildPanel(th)
	panel.OnClose(func() { m.closeMenu(ctx) })

	anchor := mm.ScreenBounds()
	size := panel.ContentSize()
	if size.Width < metrics.MenubarMenuMinWidth {
		size.Width = metrics.MenubarMenuMinWidth
	}
	pos := overlay.Position(overlay.PlacementBelow, anchor, size, ctx.WindowSize(), metrics.MenubarSideOffset)
	panel.SetBounds(geometry.FromPointSize(pos, size))

	oc := &menuOverlayContent{panel: panel, onDismiss: func() { m.closeMenu(ctx) }}
	m.overlayContent = oc
	om.PushOverlay(oc, func() { m.handleDismiss(ctx) })
	ctx.Invalidate()
}

func (m *MenubarWidget) closeMenu(ctx widget.Context) {
	m.removeOverlay(ctx)
	m.openIndex = -1
	ctx.Invalidate()
}

func (m *MenubarWidget) removeOverlay(ctx widget.Context) {
	if m.overlayContent == nil {
		return
	}
	if om := ctx.OverlayManager(); om != nil {
		om.RemoveOverlay(m.overlayContent)
	}
	m.overlayContent = nil
}

// handleDismiss reacts to an overlay-driven dismissal (outside click/Escape).
func (m *MenubarWidget) handleDismiss(ctx widget.Context) {
	m.overlayContent = nil
	m.openIndex = -1
	ctx.Invalidate()
}

// MenubarMenuWidget is a single bar trigger plus its drop-down content.
type MenubarMenuWidget struct {
	widget.WidgetBase

	label   string
	content *MenubarContentWidget
	theme   *theme.Theme

	owner *MenubarWidget
	index int
	open  bool
	hover bool
}

// MenubarMenu builds one menu: a trigger labeled label that opens content.
func MenubarMenu(label string, content *MenubarContentWidget) *MenubarMenuWidget {
	mm := &MenubarMenuWidget{label: label, content: content}
	mm.SetVisible(true)
	mm.SetEnabled(true)
	return mm
}

// Theme pins a specific theme instead of the process-wide current theme.
func (mm *MenubarMenuWidget) Theme(th *theme.Theme) *MenubarMenuWidget {
	mm.theme = th
	return mm
}

func (mm *MenubarMenuWidget) resolvedTheme() *theme.Theme {
	if mm.theme != nil {
		return mm.theme
	}
	if mm.owner != nil {
		return mm.owner.resolvedTheme()
	}
	return CurrentTheme()
}

// Layout sizes the trigger to its label plus px-2 padding.
func (mm *MenubarMenuWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	th := mm.resolvedTheme()
	family := mm.triggerFamily(th)
	textW := textmetrics.Width(family, metrics.MenubarTriggerFontSize, mm.label)
	w := textW + 2*metrics.MenubarTriggerPadX
	size := c.Constrain(geometry.Sz(w, c.MaxHeight))
	mm.SetBounds(geometry.FromPointSize(mm.Position(), size))
	return size
}

func (mm *MenubarMenuWidget) triggerFamily(th *theme.Theme) string {
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(metrics.MenubarTriggerFontWeight)
}

// Draw paints the trigger: an accent fill + accent-foreground text when open or
// hovered, otherwise transparent with foreground text.
func (mm *MenubarMenuWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !mm.IsVisible() {
		return
	}
	th := mm.resolvedTheme()
	tok := th.Active()
	b := mm.Bounds()

	textColor := tok.Foreground
	if mm.open || mm.hover {
		canvas.DrawRoundRect(b, tok.Accent, th.RadiusSM())
		textColor = tok.AccentForeground
	}

	labelRect := geometry.NewRect(
		b.Min.X+metrics.MenubarTriggerPadX,
		b.Min.Y,
		b.Width()-2*metrics.MenubarTriggerPadX,
		b.Height(),
	)
	drawMenubarText(canvas, metrics.MenubarTriggerFontWeight, metrics.MenubarTriggerFontSize, mm.label, labelRect, textColor)
}

// Event toggles the menu on a left press inside the trigger; hover updates the
// accent fill.
func (mm *MenubarMenuWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	inside := mm.Bounds().Contains(me.Position)
	switch me.MouseType {
	case event.MouseMove, event.MouseEnter:
		if inside != mm.hover {
			mm.hover = inside
			if inside {
				ctx.SetCursor(widget.CursorPointer)
			}
			mm.SetNeedsRedraw(true)
			ctx.Invalidate()
		}
	case event.MouseLeave:
		if mm.hover {
			mm.hover = false
			mm.SetNeedsRedraw(true)
		}
	case event.MousePress:
		if inside && me.Button == event.ButtonLeft && mm.owner != nil {
			mm.owner.toggle(ctx, mm.index)
			return true
		}
	}
	return false
}

// Children returns nil; the trigger draws its own label and the panel lives in
// the overlay.
func (mm *MenubarMenuWidget) Children() []widget.Widget { return nil }

// MenubarContentWidget holds one menu's entries (data only); the panel is built
// lazily into the overlay on open (mirrors DropdownMenuContentWidget).
type MenubarContentWidget struct {
	widget.WidgetBase
	entries []MenuEntry
}

// MenubarContent collects the entries for a MenubarMenu.
func MenubarContent(entries ...MenuEntry) *MenubarContentWidget {
	c := &MenubarContentWidget{entries: entries}
	c.SetVisible(false)
	return c
}

// Layout reports zero size (content renders in the overlay).
func (c *MenubarContentWidget) Layout(_ widget.Context, cs geometry.Constraints) geometry.Size {
	return cs.Constrain(geometry.Sz(0, 0))
}

// Draw is a no-op; the panel renders in the overlay layer.
func (c *MenubarContentWidget) Draw(widget.Context, widget.Canvas) {}

// Event is a no-op.
func (c *MenubarContentWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil.
func (c *MenubarContentWidget) Children() []widget.Widget { return nil }

// buildPanel converts the entries into a live menu engine panel.
func (c *MenubarContentWidget) buildPanel(th *theme.Theme) *menu.Panel {
	engineEntries := make([]menu.Entry, 0, len(c.entries))
	for _, e := range c.entries {
		e.build(&engineEntries)
	}
	return menu.NewPanel(th, engineEntries...)
}

// --- entries (share the DropdownMenu MenuEntry types + menu engine) ---------

// MenubarItem creates an actionable item.
func MenubarItem(label string) *MenuItemEntry { return DropdownMenuItem(label) }

// MenubarCheckboxItem creates a checkbox item.
func MenubarCheckboxItem(label string) *MenuCheckboxEntry { return DropdownMenuCheckboxItem(label) }

// MenubarRadioGroup creates a radio group bound to sig.
func MenubarRadioGroup(sig state.Signal[string], items ...*MenuRadioEntry) MenuEntry {
	return DropdownMenuRadioGroup(sig, items...)
}

// MenubarRadioItem creates a radio item.
func MenubarRadioItem(value, label string) *MenuRadioEntry {
	return DropdownMenuRadioItem(value, label)
}

// MenubarLabel creates a section label row.
func MenubarLabel(text string) MenuEntry { return DropdownMenuLabel(text) }

// MenubarSeparator creates a divider row.
func MenubarSeparator() MenuEntry { return DropdownMenuSeparator() }

// MenubarGroup groups entries together.
func MenubarGroup(entries ...MenuEntry) MenuEntry { return DropdownMenuGroup(entries...) }

// MenubarSub renders an inset placeholder label (sub-menus deferred for v1).
func MenubarSub(label string, entries ...MenuEntry) MenuEntry {
	return DropdownMenuSub(label, entries...)
}

// --- preview (goldens / docs) -----------------------------------------------

// MenubarMenuPreview returns one menu's panel rendered as direct content for
// goldens/docs (the path that works without an OverlayManager).
func MenubarMenuPreview(content *MenubarContentWidget) Widget {
	panel := content.buildPanel(CurrentTheme())
	return &menubarPanelPreview{panel: panel}
}

// menubarPanelPreview lays out a menu panel at the menubar minimum width with
// shadow padding so the golden does not clip it.
type menubarPanelPreview struct {
	widget.WidgetBase
	panel *menu.Panel
}

func (p *menubarPanelPreview) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	inner := p.panel.ContentSize()
	if inner.Width < metrics.MenubarMenuMinWidth {
		inner.Width = metrics.MenubarMenuMinWidth
	}
	pad := float32(16)
	size := c.Constrain(geometry.Sz(inner.Width+2*pad, inner.Height+2*pad))
	p.SetBounds(geometry.FromPointSize(p.Position(), size))
	p.panel.SetBounds(geometry.NewRect(p.Position().X+pad, p.Position().Y+pad, inner.Width, inner.Height))
	return size
}

func (p *menubarPanelPreview) Draw(ctx widget.Context, canvas widget.Canvas) {
	p.panel.Draw(ctx, canvas)
}

func (p *menubarPanelPreview) Event(widget.Context, event.Event) bool { return false }
func (p *menubarPanelPreview) Children() []widget.Widget              { return nil }

// drawMenubarText draws the trigger label, honoring StyledTextDrawer with a
// mock-canvas fallback (bold at weight >= 600).
func drawMenubarText(canvas widget.Canvas, weight int, size float32, text string, bounds geometry.Rect, color widget.Color) {
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
	_ widget.Widget = (*MenubarWidget)(nil)
	_ widget.Widget = (*MenubarMenuWidget)(nil)
	_ widget.Widget = (*MenubarContentWidget)(nil)
	_ widget.Widget = (*menubarPanelPreview)(nil)
)

package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/widgets/menu"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// DropdownMenuWidget is the shadcn DropdownMenu: a trigger that opens a
// floating menu panel built from the shared menu engine
// (internal/widgets/menu). The panel is placed bottom-start with a 4px side
// offset and dismisses on outside click or Escape.
//
// The host watches an open signal (or its own internal state) and pushes /
// removes the overlay accordingly.
type DropdownMenuWidget struct {
	widget.WidgetBase

	trigger *DropdownMenuTriggerWidget
	content *DropdownMenuContentWidget

	open         state.Signal[bool]
	onOpenChange func(bool)

	overlayContent *menuOverlayContent
	theme          *theme.Theme
}

// DropdownMenu builds a DropdownMenu from a trigger and a content block.
// Extra children are ignored; the first trigger and first content win.
func DropdownMenu(children ...Widget) *DropdownMenuWidget {
	d := &DropdownMenuWidget{}
	for _, ch := range children {
		switch c := ch.(type) {
		case *DropdownMenuTriggerWidget:
			if d.trigger == nil {
				d.trigger = c
			}
		case *DropdownMenuContentWidget:
			if d.content == nil {
				d.content = c
			}
		}
	}
	if d.trigger != nil {
		d.trigger.owner = d
	}
	d.SetVisible(true)
	d.SetEnabled(true)
	return d
}

// Bind binds the open state to a controlled bool signal.
func (d *DropdownMenuWidget) Bind(open state.Signal[bool]) *DropdownMenuWidget {
	d.open = open
	return d
}

// OnOpenChange registers an open-state observer.
func (d *DropdownMenuWidget) OnOpenChange(fn func(bool)) *DropdownMenuWidget {
	d.onOpenChange = fn
	return d
}

// Theme pins a specific theme instead of the process-wide current theme.
func (d *DropdownMenuWidget) Theme(th *theme.Theme) *DropdownMenuWidget {
	d.theme = th
	return d
}

func (d *DropdownMenuWidget) resolvedTheme() *theme.Theme {
	if d.theme != nil {
		return d.theme
	}
	return CurrentTheme()
}

func (d *DropdownMenuWidget) isOpen() bool {
	if d.open != nil {
		return d.open.Get()
	}
	return d.overlayContent != nil
}

// setOpen updates open state, fires the observer, and pushes/removes the
// overlay.
func (d *DropdownMenuWidget) setOpen(ctx widget.Context, v bool) {
	if d.open != nil {
		d.open.Set(v)
	}
	if d.onOpenChange != nil {
		d.onOpenChange(v)
	}
	if v {
		d.pushOverlay(ctx)
	} else {
		d.removeOverlay(ctx)
	}
}

func (d *DropdownMenuWidget) toggle(ctx widget.Context) {
	d.setOpen(ctx, !d.isOpen())
}

func (d *DropdownMenuWidget) pushOverlay(ctx widget.Context) {
	if d.content == nil || d.overlayContent != nil {
		return
	}
	om := ctx.OverlayManager()
	if om == nil {
		return
	}
	th := d.resolvedTheme()
	panel := d.content.buildPanel(th)
	panel.OnClose(func() { d.setOpen(ctx, false) })
	panel.OnActivate(func() { d.setOpen(ctx, false) })

	anchor := d.trigger.ScreenBounds()
	size := panel.ContentSize()
	pos := overlay.Position(overlay.PlacementBelow, anchor, size, ctx.WindowSize(), metrics.Menu.SideOffset)
	panel.SetBounds(geometry.FromPointSize(pos, size))

	oc := &menuOverlayContent{panel: panel, onDismiss: func() { d.setOpen(ctx, false) }}
	d.overlayContent = oc
	om.PushOverlay(oc, func() { d.setOpen(ctx, false) })
	d.SetNeedsRedraw(true)
	ctx.Invalidate()
}

func (d *DropdownMenuWidget) removeOverlay(ctx widget.Context) {
	if d.overlayContent == nil {
		return
	}
	if om := ctx.OverlayManager(); om != nil {
		om.RemoveOverlay(d.overlayContent)
	}
	d.overlayContent = nil
	d.SetNeedsRedraw(true)
	ctx.Invalidate()
}

// Layout positions the trigger (the host occupies the trigger's space).
func (d *DropdownMenuWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if d.trigger == nil {
		return c.Constrain(geometry.Sz(0, 0))
	}
	size := d.trigger.Layout(ctx, c)
	d.trigger.SetBounds(geometry.FromPointSize(d.Position(), size))
	d.SetBounds(geometry.FromPointSize(d.Position(), size))
	return size
}

// Draw renders the trigger.
func (d *DropdownMenuWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if d.trigger != nil {
		d.trigger.Draw(ctx, canvas)
	}
}

// Event forwards events to the trigger.
func (d *DropdownMenuWidget) Event(ctx widget.Context, e event.Event) bool {
	if d.trigger != nil {
		return d.trigger.Event(ctx, e)
	}
	return false
}

// Children returns the trigger.
func (d *DropdownMenuWidget) Children() []widget.Widget {
	if d.trigger == nil {
		return nil
	}
	return []widget.Widget{d.trigger}
}

// Mount wires the open signal for push-based invalidation.
func (d *DropdownMenuWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil || d.open == nil {
		return
	}
	b := state.BindToScheduler(d.open, d, sched)
	d.AddBinding(b)
}

// Unmount is a no-op; bindings are cleaned up by WidgetBase.
func (d *DropdownMenuWidget) Unmount() {}

// --- trigger ------------------------------------------------------------

// DropdownMenuTriggerWidget wraps an arbitrary widget that toggles the menu
// when clicked.
type DropdownMenuTriggerWidget struct {
	widget.WidgetBase

	child Widget
	owner *DropdownMenuWidget
}

// DropdownMenuTrigger wraps w as the menu trigger.
func DropdownMenuTrigger(w Widget) *DropdownMenuTriggerWidget {
	t := &DropdownMenuTriggerWidget{child: w}
	t.SetVisible(true)
	t.SetEnabled(true)
	return t
}

// Layout lays out the child and adopts its size.
func (t *DropdownMenuTriggerWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if t.child == nil {
		return c.Constrain(geometry.Sz(0, 0))
	}
	size := t.child.Layout(ctx, c)
	setWidgetBounds(t.child, geometry.FromPointSize(t.Position(), size))
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

// Draw renders the child.
func (t *DropdownMenuTriggerWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if t.child != nil {
		t.child.Draw(ctx, canvas)
	}
}

// Event toggles the menu on a left press inside bounds, after giving the
// child a chance to handle the event (e.g. button hover states).
func (t *DropdownMenuTriggerWidget) Event(ctx widget.Context, e event.Event) bool {
	if t.child != nil {
		t.child.Event(ctx, e)
	}
	if me, ok := e.(*event.MouseEvent); ok {
		if me.MouseType == event.MousePress && me.Button == event.ButtonLeft && t.Bounds().Contains(me.Position) {
			if t.owner != nil {
				t.owner.toggle(ctx)
			}
			return true
		}
	}
	return false
}

// Children returns the wrapped child.
func (t *DropdownMenuTriggerWidget) Children() []widget.Widget {
	if t.child == nil {
		return nil
	}
	return []widget.Widget{t.child}
}

// --- content ------------------------------------------------------------

// DropdownMenuContentWidget holds the menu entries. It is a zero-size,
// inert widget (it carries data only); the real panel is built lazily into
// the overlay when the menu opens.
type DropdownMenuContentWidget struct {
	widget.WidgetBase
	entries []MenuEntry
}

// DropdownMenuContent collects the menu entries.
func DropdownMenuContent(entries ...MenuEntry) *DropdownMenuContentWidget {
	c := &DropdownMenuContentWidget{entries: entries}
	c.SetVisible(false)
	return c
}

// Layout reports zero size (content is rendered in the overlay, not inline).
func (c *DropdownMenuContentWidget) Layout(_ widget.Context, cs geometry.Constraints) geometry.Size {
	return cs.Constrain(geometry.Sz(0, 0))
}

// Draw is a no-op; the panel renders in the overlay layer.
func (c *DropdownMenuContentWidget) Draw(widget.Context, widget.Canvas) {}

// Event is a no-op.
func (c *DropdownMenuContentWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil.
func (c *DropdownMenuContentWidget) Children() []widget.Widget { return nil }

// buildPanel converts the entries into a live menu engine panel.
func (c *DropdownMenuContentWidget) buildPanel(th *theme.Theme) *menu.Panel {
	engineEntries := make([]menu.Entry, 0, len(c.entries))
	for _, e := range c.entries {
		e.build(&engineEntries)
	}
	return menu.NewPanel(th, engineEntries...)
}

// menuOverlayContent is the overlay wrapper for a dropdown menu panel. It is
// non-modal: clicks outside the panel and Escape dismiss it. The panel is
// pre-positioned (bounds set before push); this wrapper preserves those
// bounds (it does not re-lay-out the panel to the window origin) and routes
// events to the panel, dismissing on an outside press.
type menuOverlayContent struct {
	widget.WidgetBase

	panel     *menu.Panel
	onDismiss func()
}

func (o *menuOverlayContent) Layout(_ widget.Context, _ geometry.Constraints) geometry.Size {
	// Adopt the (pre-positioned) panel bounds so the overlay stack and
	// dismiss hit-testing see the panel rect.
	o.SetBounds(o.panel.Bounds())
	return o.panel.Bounds().Size()
}

func (o *menuOverlayContent) Draw(ctx widget.Context, canvas widget.Canvas) {
	o.panel.Draw(ctx, canvas)
}

func (o *menuOverlayContent) Event(ctx widget.Context, e event.Event) bool {
	if o.panel.Event(ctx, e) {
		return true
	}
	// Outside press dismisses.
	if me, ok := e.(*event.MouseEvent); ok && me.MouseType == event.MousePress {
		if !o.panel.Bounds().Contains(me.Position) {
			o.Dismiss()
			return true
		}
	}
	return false
}

func (o *menuOverlayContent) Dismiss() {
	if o.onDismiss != nil {
		o.onDismiss()
	}
}

func (o *menuOverlayContent) Modal() bool { return false }

func (o *menuOverlayContent) Children() []widget.Widget {
	return []widget.Widget{o.panel}
}

// --- entries ------------------------------------------------------------

// MenuEntry is one entry passed to DropdownMenuContent. Concrete entries are
// built with DropdownMenuItem, DropdownMenuCheckboxItem,
// DropdownMenuRadioGroup, DropdownMenuLabel, DropdownMenuSeparator, and
// DropdownMenuGroup.
type MenuEntry interface {
	// build appends the engine entry/entries for this MenuEntry.
	build(out *[]menu.Entry)
}

// MenuItemEntry is an actionable dropdown item.
type MenuItemEntry struct {
	label       string
	icon        icon.IconData
	hasIcon     bool
	shortcut    string
	onSelect    func()
	destructive bool
	disabled    bool
	inset       bool
}

// DropdownMenuItem creates an actionable item.
func DropdownMenuItem(label string) *MenuItemEntry {
	return &MenuItemEntry{label: label}
}

// Icon sets the leading icon (16px muted-foreground).
func (e *MenuItemEntry) Icon(ic icon.IconData) *MenuItemEntry {
	e.icon = ic
	e.hasIcon = true
	return e
}

// Shortcut sets the right-aligned shortcut text.
func (e *MenuItemEntry) Shortcut(s string) *MenuItemEntry { e.shortcut = s; return e }

// OnSelect registers the selection callback.
func (e *MenuItemEntry) OnSelect(fn func()) *MenuItemEntry { e.onSelect = fn; return e }

// Destructive marks the item as destructive.
func (e *MenuItemEntry) Destructive() *MenuItemEntry { e.destructive = true; return e }

// Disabled marks the item disabled.
func (e *MenuItemEntry) Disabled(v bool) *MenuItemEntry { e.disabled = v; return e }

// Inset shifts the label right (pl-8).
func (e *MenuItemEntry) Inset() *MenuItemEntry { e.inset = true; return e }

func (e *MenuItemEntry) build(out *[]menu.Entry) {
	item := menu.NewItem(e.label)
	if e.hasIcon {
		item.Icon(e.icon)
	}
	if e.shortcut != "" {
		item.SetShortcut(e.shortcut)
	}
	if e.destructive {
		item.SetDestructive()
	}
	if e.disabled {
		item.SetDisabled(true)
	}
	if e.inset {
		item.SetInset()
	}
	if e.onSelect != nil {
		item.OnSelect(e.onSelect)
	}
	*out = append(*out, item)
}

// MenuCheckboxEntry is a checkbox dropdown item.
type MenuCheckboxEntry struct {
	label    string
	signal   state.Signal[bool]
	checked  bool
	onChange func(bool)
	disabled bool
}

// DropdownMenuCheckboxItem creates a checkbox item.
func DropdownMenuCheckboxItem(label string) *MenuCheckboxEntry {
	return &MenuCheckboxEntry{label: label}
}

// Bind binds the checked state to a controlled bool signal.
func (e *MenuCheckboxEntry) Bind(sig state.Signal[bool]) *MenuCheckboxEntry {
	e.signal = sig
	return e
}

// Checked sets the initial checked state.
func (e *MenuCheckboxEntry) Checked(v bool) *MenuCheckboxEntry { e.checked = v; return e }

// OnChange registers the toggle observer.
func (e *MenuCheckboxEntry) OnChange(fn func(bool)) *MenuCheckboxEntry { e.onChange = fn; return e }

// Disabled marks the checkbox item disabled.
func (e *MenuCheckboxEntry) Disabled(v bool) *MenuCheckboxEntry { e.disabled = v; return e }

func (e *MenuCheckboxEntry) build(out *[]menu.Entry) {
	checked := e.checked
	if e.signal != nil {
		checked = e.signal.Get()
	}
	cb := menu.NewCheckboxItem(e.label, icons.Check).SetChecked(checked)
	if e.disabled {
		cb.SetDisabled(true)
	}
	cb.OnChange(func(v bool) {
		if e.signal != nil {
			e.signal.Set(v)
		}
		if e.onChange != nil {
			e.onChange(v)
		}
	})
	*out = append(*out, cb)
}

// MenuRadioGroupEntry is a group of radio items bound to one signal.
type MenuRadioGroupEntry struct {
	signal state.Signal[string]
	items  []*MenuRadioEntry
}

// DropdownMenuRadioGroup creates a radio group bound to sig.
func DropdownMenuRadioGroup(sig state.Signal[string], items ...*MenuRadioEntry) MenuEntry {
	return &MenuRadioGroupEntry{signal: sig, items: items}
}

func (g *MenuRadioGroupEntry) build(out *[]menu.Entry) {
	cur := ""
	if g.signal != nil {
		cur = g.signal.Get()
	}
	for _, it := range g.items {
		r := menu.NewRadioItem(it.value, it.label, icons.Circle).SetSelected(it.value == cur)
		if it.disabled {
			r.SetDisabled(true)
		}
		value := it.value
		r.OnSelect(func(v string) {
			if g.signal != nil {
				g.signal.Set(value)
			}
			if it.onSelect != nil {
				it.onSelect()
			}
		})
		*out = append(*out, r)
	}
}

// MenuRadioEntry is a single radio item within a radio group.
type MenuRadioEntry struct {
	value    string
	label    string
	disabled bool
	onSelect func()
}

// DropdownMenuRadioItem creates a radio item.
func DropdownMenuRadioItem(value, label string) *MenuRadioEntry {
	return &MenuRadioEntry{value: value, label: label}
}

// Disabled marks the radio item disabled.
func (e *MenuRadioEntry) Disabled(v bool) *MenuRadioEntry { e.disabled = v; return e }

// OnSelect registers a per-item selection callback.
func (e *MenuRadioEntry) OnSelect(fn func()) *MenuRadioEntry { e.onSelect = fn; return e }

// build is unused directly (radio items build via their group), but
// MenuRadioEntry satisfies MenuEntry so a lone radio item can be placed.
func (e *MenuRadioEntry) build(out *[]menu.Entry) {
	r := menu.NewRadioItem(e.value, e.label, icons.Circle)
	if e.disabled {
		r.SetDisabled(true)
	}
	if e.onSelect != nil {
		r.OnSelect(func(string) { e.onSelect() })
	}
	*out = append(*out, r)
}

// menuLabelEntry is a section label.
type menuLabelEntry struct {
	text  string
	inset bool
}

// DropdownMenuLabel creates a section label row.
func DropdownMenuLabel(text string) MenuEntry { return &menuLabelEntry{text: text} }

func (e *menuLabelEntry) build(out *[]menu.Entry) {
	l := menu.NewLabel(e.text)
	if e.inset {
		l.SetInset()
	}
	*out = append(*out, l)
}

// menuSeparatorEntry is a divider.
type menuSeparatorEntry struct{}

// DropdownMenuSeparator creates a divider row.
func DropdownMenuSeparator() MenuEntry { return menuSeparatorEntry{} }

func (menuSeparatorEntry) build(out *[]menu.Entry) {
	*out = append(*out, menu.NewSeparator())
}

// menuGroupEntry groups entries (no visual chrome; semantic grouping).
type menuGroupEntry struct {
	entries []MenuEntry
}

// DropdownMenuGroup groups entries together.
func DropdownMenuGroup(entries ...MenuEntry) MenuEntry {
	return &menuGroupEntry{entries: entries}
}

func (g *menuGroupEntry) build(out *[]menu.Entry) {
	for _, e := range g.entries {
		e.build(out)
	}
}

// DropdownMenuSub is deferred for v1: sub-menus are not yet implemented. It
// renders as an inset, disabled placeholder label so the API exists without
// a working flyout.
func DropdownMenuSub(label string, entries ...MenuEntry) MenuEntry {
	return &menuLabelEntry{text: label, inset: true}
}

// --- preview (goldens / docs) -------------------------------------------

// DropdownMenuPreview returns the menu panel of d rendered as direct content
// (a self-sizing widget) for documentation and golden tests.
func DropdownMenuPreview(content *DropdownMenuContentWidget) Widget {
	return DropdownMenuContentPreview(content, CurrentTheme())
}

// DropdownMenuContentPreview renders content's panel against a specific theme.
func DropdownMenuContentPreview(content *DropdownMenuContentWidget, th *theme.Theme) Widget {
	panel := content.buildPanel(th)
	return &menuPreviewWidget{panel: panel}
}

// menuPreviewWidget lays out a menu panel at its natural size with shadow
// padding so the golden does not clip it.
type menuPreviewWidget struct {
	widget.WidgetBase
	panel *menu.Panel
}

func (p *menuPreviewWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	inner := p.panel.ContentSize()
	pad := float32(16)
	size := c.Constrain(geometry.Sz(inner.Width+2*pad, inner.Height+2*pad))
	p.SetBounds(geometry.FromPointSize(p.Position(), size))
	p.panel.SetBounds(geometry.NewRect(p.Position().X+pad, p.Position().Y+pad, inner.Width, inner.Height))
	return size
}

func (p *menuPreviewWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	p.panel.Draw(ctx, canvas)
}

func (p *menuPreviewWidget) Event(widget.Context, event.Event) bool { return false }
func (p *menuPreviewWidget) Children() []widget.Widget              { return nil }

// Compile-time interface checks.
var (
	_ widget.Widget    = (*DropdownMenuWidget)(nil)
	_ widget.Lifecycle = (*DropdownMenuWidget)(nil)
	_ widget.Widget    = (*DropdownMenuTriggerWidget)(nil)
	_ widget.Widget    = (*menuOverlayContent)(nil)
	_ overlay.Overlay  = (*menuOverlayContent)(nil)
	_ widget.Widget    = (*menuPreviewWidget)(nil)
)

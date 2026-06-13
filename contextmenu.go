package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/widgets/menu"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// ContextMenuWidget is the shadcn ContextMenu: a target region that opens a
// floating menu panel at the cursor on right-click. It reuses the shared menu
// engine (internal/widgets/menu) and the same MenuEntry types as DropdownMenu;
// only the trigger (secondary-button press) and placement (at the pointer's
// GlobalPosition) differ.
//
// The panel dismisses on outside click or Escape (non-modal overlay, the same
// menuOverlayContent wrapper DropdownMenu uses).
type ContextMenuWidget struct {
	widget.WidgetBase

	target  widget.Widget
	content *ContextMenuContentWidget
	theme   *theme.Theme

	overlayContent *menuOverlayContent
}

// ContextMenu wraps target so a right-click reveals content at the cursor.
func ContextMenu(target widget.Widget, content *ContextMenuContentWidget) *ContextMenuWidget {
	c := &ContextMenuWidget{target: target, content: content}
	c.SetVisible(true)
	c.SetEnabled(true)
	if target != nil {
		c.AddChild(target)
	}
	return c
}

// Theme pins a specific theme instead of the process-wide current theme.
func (c *ContextMenuWidget) Theme(th *theme.Theme) *ContextMenuWidget {
	c.theme = th
	return c
}

func (c *ContextMenuWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// Layout sizes the wrapper to its target.
func (c *ContextMenuWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	var size geometry.Size
	if c.target != nil {
		size = c.target.Layout(ctx, cons)
		setWidgetBounds(c.target, geometry.FromPointSize(c.Position(), size))
	} else {
		size = cons.Constrain(geometry.Sz(0, 0))
	}
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw renders the target. The menu paints in its own overlay.
func (c *ContextMenuWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	if c.target != nil {
		setWidgetBounds(c.target, c.Bounds())
		c.target.Draw(ctx, canvas)
	}
}

// Event opens the menu at the cursor on a right-button press inside the target.
func (c *ContextMenuWidget) Event(ctx widget.Context, e event.Event) bool {
	if c.target != nil {
		c.target.Event(ctx, e)
	}
	if me, ok := e.(*event.MouseEvent); ok {
		if me.MouseType == event.MousePress && me.Button == event.ButtonRight && c.Bounds().Contains(me.Position) {
			c.openAt(ctx, me.GlobalPosition)
			return true
		}
	}
	return false
}

// openAt builds the panel and pushes it anchored at the cursor position.
func (c *ContextMenuWidget) openAt(ctx widget.Context, cursor geometry.Point) {
	if c.content == nil || c.overlayContent != nil {
		return
	}
	om := ctx.OverlayManager()
	if om == nil {
		return
	}
	th := c.resolvedTheme()
	panel := c.content.buildPanel(th)
	panel.OnClose(func() { c.close(ctx) })
	panel.OnActivate(func() { c.close(ctx) })

	size := panel.ContentSize()
	pos := c.placeAtCursor(cursor, size, ctx.WindowSize())
	panel.SetBounds(geometry.FromPointSize(pos, size))

	oc := &menuOverlayContent{panel: panel, onDismiss: func() { c.close(ctx) }}
	c.overlayContent = oc
	om.PushOverlay(oc, func() { c.overlayContent = nil })
	ctx.Invalidate()
}

// placeAtCursor anchors the panel's top-left at the cursor (plus a small gap),
// clamping so the panel stays within the window.
func (c *ContextMenuWidget) placeAtCursor(cursor geometry.Point, size, window geometry.Size) geometry.Point {
	x := cursor.X + metrics.ContextMenuCursorGap
	y := cursor.Y + metrics.ContextMenuCursorGap
	if x+size.Width > window.Width {
		x = cursor.X - size.Width
	}
	if x < 0 {
		x = 0
	}
	if y+size.Height > window.Height {
		y = cursor.Y - size.Height
	}
	if y < 0 {
		y = 0
	}
	return geometry.Pt(x, y)
}

// close removes the overlay if shown.
func (c *ContextMenuWidget) close(ctx widget.Context) {
	if c.overlayContent == nil {
		return
	}
	if om := ctx.OverlayManager(); om != nil {
		om.RemoveOverlay(c.overlayContent)
	}
	c.overlayContent = nil
	ctx.Invalidate()
}

// Children returns the wrapped target.
func (c *ContextMenuWidget) Children() []widget.Widget {
	if c.target == nil {
		return nil
	}
	return []widget.Widget{c.target}
}

// ContextMenuContentWidget holds the menu entries. It is a zero-size, inert
// widget (data only); the real panel is built lazily into the overlay when the
// menu opens (mirrors DropdownMenuContentWidget).
type ContextMenuContentWidget struct {
	widget.WidgetBase
	entries []MenuEntry
}

// ContextMenuContent collects the menu entries.
func ContextMenuContent(entries ...MenuEntry) *ContextMenuContentWidget {
	c := &ContextMenuContentWidget{entries: entries}
	c.SetVisible(false)
	return c
}

// Layout reports zero size (content renders in the overlay, not inline).
func (c *ContextMenuContentWidget) Layout(_ widget.Context, cs geometry.Constraints) geometry.Size {
	return cs.Constrain(geometry.Sz(0, 0))
}

// Draw is a no-op; the panel renders in the overlay layer.
func (c *ContextMenuContentWidget) Draw(widget.Context, widget.Canvas) {}

// Event is a no-op.
func (c *ContextMenuContentWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil.
func (c *ContextMenuContentWidget) Children() []widget.Widget { return nil }

// buildPanel converts the entries into a live menu engine panel.
func (c *ContextMenuContentWidget) buildPanel(th *theme.Theme) *menu.Panel {
	engineEntries := make([]menu.Entry, 0, len(c.entries))
	for _, e := range c.entries {
		e.build(&engineEntries)
	}
	return menu.NewPanel(th, engineEntries...)
}

// --- entries (share the DropdownMenu MenuEntry types + menu engine) ---------

// ContextMenuItem creates an actionable item (shares MenuItemEntry).
func ContextMenuItem(label string) *MenuItemEntry { return DropdownMenuItem(label) }

// ContextMenuCheckboxItem creates a checkbox item (shares MenuCheckboxEntry).
func ContextMenuCheckboxItem(label string) *MenuCheckboxEntry { return DropdownMenuCheckboxItem(label) }

// ContextMenuRadioGroup creates a radio group bound to sig (shares the engine).
func ContextMenuRadioGroup(sig state.Signal[string], items ...*MenuRadioEntry) MenuEntry {
	return DropdownMenuRadioGroup(sig, items...)
}

// ContextMenuRadioItem creates a radio item.
func ContextMenuRadioItem(value, label string) *MenuRadioEntry {
	return DropdownMenuRadioItem(value, label)
}

// ContextMenuLabel creates a section label row.
func ContextMenuLabel(text string) MenuEntry { return DropdownMenuLabel(text) }

// ContextMenuSeparator creates a divider row.
func ContextMenuSeparator() MenuEntry { return DropdownMenuSeparator() }

// ContextMenuGroup groups entries together.
func ContextMenuGroup(entries ...MenuEntry) MenuEntry { return DropdownMenuGroup(entries...) }

// ContextMenuSub renders an inset placeholder label (sub-menus deferred for v1,
// matching DropdownMenuSub).
func ContextMenuSub(label string, entries ...MenuEntry) MenuEntry {
	return DropdownMenuSub(label, entries...)
}

// --- preview (goldens / docs) -----------------------------------------------

// ContextMenuPreview returns the menu panel of content rendered as direct
// content (a self-sizing widget) for documentation and golden tests.
func ContextMenuPreview(content *ContextMenuContentWidget) Widget {
	return ContextMenuContentPreview(content, CurrentTheme())
}

// ContextMenuContentPreview renders content's panel against a specific theme.
func ContextMenuContentPreview(content *ContextMenuContentWidget, th *theme.Theme) Widget {
	panel := content.buildPanel(th)
	return &menuPreviewWidget{panel: panel}
}

// Compile-time interface checks.
var (
	_ widget.Widget = (*ContextMenuWidget)(nil)
	_ widget.Widget = (*ContextMenuContentWidget)(nil)
)

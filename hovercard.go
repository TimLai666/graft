package graft

import (
	"time"

	corepopover "github.com/gogpu/ui/core/popover"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// HoverCardWidget wraps a trigger and reveals floating content on hover — the
// shadcn HoverCard (a non-interactive-dismiss cousin of Popover).
//
// Architecture decision: graft-owned widget, modeled on the shipped Tooltip
// (hover-driven overlay timed via ctx.Now()) but rendering Popover-style
// content (w-64, p-4, rounded-md, 1px border, bg-popover, shadow-md). It is
// NOT a wrap of core/popover: shadcn's surface has exact px metrics and the
// hover open/close-delay machine, neither of which core/popover exposes
// through Layout. The overlay is shown via ctx.OverlayManager() positioned
// with core/popover.CalculatePosition (Report 2 §4.4 recipe), placement
// Bottom center, flipping when cramped.
//
// shadcn ships openDelay 700ms / closeDelay 300ms; both are honored via
// ctx.Now(). The card dismisses when the pointer leaves the trigger (after the
// close delay) — it is non-interactive (no click-outside/Escape machinery, no
// pointer events into the card).
type HoverCardWidget struct {
	widget.WidgetBase

	trigger widget.Widget
	content *HoverCardContentWidget
	theme   *theme.Theme

	hovered    bool
	hoverSince time.Time
	leaveSince time.Time
	overlay    widget.Widget // non-nil while shown
}

// HoverCard wraps trigger so hovering it reveals content in a floating card
// positioned below the trigger (auto-flipped above when cramped).
func HoverCard(trigger widget.Widget, content *HoverCardContentWidget) *HoverCardWidget {
	h := &HoverCardWidget{
		trigger: trigger,
		content: content,
		theme:   CurrentTheme(),
	}
	h.SetVisible(true)
	h.SetEnabled(true)
	if trigger != nil {
		h.AddChild(trigger)
	}
	return h
}

// Theme pins a specific theme instead of the process-wide current theme.
func (h *HoverCardWidget) Theme(th *theme.Theme) *HoverCardWidget {
	h.theme = th
	if h.content != nil {
		h.content.Theme(th)
	}
	return h
}

// Placement overrides the default bottom-center placement.
func (h *HoverCardWidget) Placement(p corepopover.Placement) *HoverCardWidget {
	h.content.placement = p
	return h
}

func (h *HoverCardWidget) resolvedTheme() *theme.Theme {
	if h.theme != nil {
		return h.theme
	}
	return CurrentTheme()
}

// Content returns the floating card widget (rendered directly by goldens).
func (h *HoverCardWidget) Content() *HoverCardContentWidget { return h.content }

// Layout sizes the wrapper to its trigger.
func (h *HoverCardWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	var size geometry.Size
	if h.trigger != nil {
		size = h.trigger.Layout(ctx, c)
		setWidgetBounds(h.trigger, geometry.FromPointSize(h.Position(), size))
	} else {
		size = c.Constrain(geometry.Sz(0, 0))
	}
	h.SetBounds(geometry.FromPointSize(h.Position(), size))
	return size
}

// Draw renders the trigger. The card paints in its own overlay.
func (h *HoverCardWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !h.IsVisible() {
		return
	}
	if h.trigger != nil {
		setWidgetBounds(h.trigger, h.Bounds())
		h.trigger.Draw(ctx, canvas)
	}
}

// Event tracks hover and drives the overlay with open/close delays.
func (h *HoverCardWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}

	switch me.MouseType {
	case event.MouseEnter:
		h.enter(ctx)
	case event.MouseMove:
		if h.Bounds().Contains(me.Position) {
			h.enter(ctx)
		} else {
			h.leave(ctx)
		}
	case event.MouseLeave:
		h.leave(ctx)
	}
	return false
}

// enter marks the trigger hovered and schedules the open after the open delay.
func (h *HoverCardWidget) enter(ctx widget.Context) {
	if !h.hovered {
		h.hovered = true
		h.hoverSince = ctx.Now()
	}
	h.maybeShow(ctx)
}

// leave marks the trigger un-hovered and schedules the close after the close
// delay.
func (h *HoverCardWidget) leave(ctx widget.Context) {
	if h.hovered {
		h.hovered = false
		h.leaveSince = ctx.Now()
	}
	h.maybeHide(ctx)
}

// maybeShow opens the overlay once the open-delay grace period has elapsed.
func (h *HoverCardWidget) maybeShow(ctx widget.Context) {
	if h.overlay != nil || !h.hovered {
		return
	}
	if ctx.Now().Sub(h.hoverSince) < metrics.HoverCardOpenDelayMillis*time.Millisecond {
		ctx.Invalidate() // keep ticking so the delay can elapse
		return
	}
	om := ctx.OverlayManager()
	if om == nil {
		return
	}

	h.content.theme = h.resolvedTheme()
	size := h.content.Layout(ctx, geometry.Loose(ctx.WindowSize()))

	anchor := h.ScreenBounds()
	pos := corepopover.CalculatePosition(h.content.placement, anchor, size, ctx.WindowSize(), metrics.HoverCardSideOffset)
	h.content.SetBounds(geometry.FromPointSize(pos, size))

	om.PushOverlay(h.content, func() { h.overlay = nil })
	h.overlay = h.content
	ctx.Invalidate()
}

// maybeHide removes the overlay once the close-delay grace period has elapsed.
func (h *HoverCardWidget) maybeHide(ctx widget.Context) {
	if h.overlay == nil || h.hovered {
		return
	}
	if ctx.Now().Sub(h.leaveSince) < metrics.HoverCardCloseDelayMillis*time.Millisecond {
		ctx.Invalidate() // keep ticking so the close delay can elapse
		return
	}
	if om := ctx.OverlayManager(); om != nil {
		om.RemoveOverlay(h.overlay)
	}
	h.overlay = nil
	ctx.Invalidate()
}

// Children returns the wrapped trigger.
func (h *HoverCardWidget) Children() []widget.Widget {
	if h.trigger == nil {
		return nil
	}
	return []widget.Widget{h.trigger}
}

// HoverCardContentWidget is the floating card surface: w-64, p-4, rounded-md,
// 1px border, bg-popover, shadow-md. It is the widget rendered directly at
// natural size by the goldens.
type HoverCardContentWidget struct {
	widget.WidgetBase

	children  []widget.Widget
	width     float32 // 0 = metrics.HoverCardWidth
	placement corepopover.Placement
	theme     *theme.Theme
}

// HoverCardContent stacks children vertically inside the shadcn hover-card
// surface.
func HoverCardContent(children ...Widget) *HoverCardContentWidget {
	c := &HoverCardContentWidget{
		children:  children,
		placement: corepopover.Bottom,
	}
	c.SetVisible(true)
	c.SetEnabled(true)
	for _, ch := range children {
		ovlSetParent(ch, c)
	}
	return c
}

// W overrides the content width (shadcn default w-64 = 256px).
func (c *HoverCardContentWidget) W(px float32) *HoverCardContentWidget {
	c.width = px
	return c
}

// Theme pins a specific theme instead of the process-wide current theme.
func (c *HoverCardContentWidget) Theme(th *theme.Theme) *HoverCardContentWidget {
	c.theme = th
	return c
}

func (c *HoverCardContentWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// Layout stacks children top-to-bottom inside the p-4 padding at the fixed
// w-64 width, preserving the pre-positioned overlay origin.
func (c *HoverCardContentWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	w := c.width
	if w <= 0 {
		w = metrics.HoverCardWidth
	}
	pad := metrics.HoverCardPadding
	inner := w - 2*pad

	y := pad
	for _, ch := range c.children {
		sz := ch.Layout(ctx, geometry.Constraints{
			MinWidth: 0, MaxWidth: inner,
			MinHeight: 0, MaxHeight: geometry.Infinity,
		})
		ovlSetBounds(ch, geometry.FromPointSize(geometry.Pt(pad, y), sz))
		y += sz.Height
	}
	size := cons.Constrain(geometry.Sz(w, y+pad))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw paints shadow-md, the bg-popover fill, the 1px border, then the
// children. Colors resolve from the active token set at draw time.
func (c *HoverCardContentWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	tok := c.resolvedTheme().Active()
	r := c.resolvedTheme().RadiusMD()
	bounds := c.Bounds()

	draw.Shadow(canvas, bounds, r, metrics.ShadowMD)
	canvas.DrawRoundRect(bounds, tok.Popover, r)
	draw.InsideBorder(canvas, bounds, r, tok.Border, metrics.HoverCardBorderWidth)

	canvas.PushTransform(bounds.Min)
	for _, ch := range c.children {
		widget.StampScreenOrigin(ch, canvas)
		widget.DrawChild(ch, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event ignores input; the card is non-interactive (hover lives on the
// wrapper, and the HoverCard closes on pointer leave).
func (c *HoverCardContentWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns the stacked content children.
func (c *HoverCardContentWidget) Children() []widget.Widget { return c.children }

// Compile-time interface checks.
var (
	_ widget.Widget = (*HoverCardWidget)(nil)
	_ widget.Widget = (*HoverCardContentWidget)(nil)
)

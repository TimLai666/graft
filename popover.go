package graft

import (
	corepopover "github.com/gogpu/ui/core/popover"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// PopoverWidget is graft's shadcn Popover host.
//
// Architecture decision: graft-owned widget. The host renders its trigger
// inline; on open it measures the content, positions it bottom-center of
// the trigger via core/popover.CalculatePosition (flip + clamp), and
// pushes it on ctx.OverlayManager() (Report 2 section 4.4 recipe). The
// window wraps the pushed content in an overlay.Container, which supplies
// click-outside and Escape dismissal. Only the placement helper is reused
// from core/popover; painting is fully graft-owned.
//
// Open state is a signal: Bind makes it controlled; without Bind an
// internal signal is used and the trigger toggles it.
type PopoverWidget struct {
	widget.WidgetBase

	trigger widget.Widget
	content *PopoverContentWidget

	open         state.Signal[bool]
	onOpenChange func(bool)

	shown bool
	om    widget.OverlayManager

	theme *theme.Theme
}

// Popover builds a popover from a trigger (rendered inline) and floating
// content. Clicking the trigger toggles the popover; clicking outside or
// pressing Escape closes it.
func Popover(trigger Widget, content *PopoverContentWidget) *PopoverWidget {
	p := &PopoverWidget{
		trigger: trigger,
		content: content,
		open:    state.NewSignal(false),
	}
	p.SetVisible(true)
	p.SetEnabled(true)
	ovlSetParent(trigger, p)
	return p
}

// Bind makes open the controlled open-state signal (shadcn open/
// onOpenChange). Writing true/false to the signal opens/closes the
// popover; interactions write back to it.
func (p *PopoverWidget) Bind(open state.Signal[bool]) *PopoverWidget {
	if open != nil {
		p.open = open
	}
	return p
}

// OnOpenChange registers an observer fired when an interaction opens or
// closes the popover.
func (p *PopoverWidget) OnOpenChange(fn func(bool)) *PopoverWidget {
	p.onOpenChange = fn
	return p
}

// Theme pins a specific theme instead of the process-wide current theme.
func (p *PopoverWidget) Theme(th *theme.Theme) *PopoverWidget {
	p.theme = th
	if p.content != nil {
		p.content.Theme(th)
	}
	return p
}

// IsOpen reports whether the popover content is currently requested open.
func (p *PopoverWidget) IsOpen() bool { return p.open.Get() }

// Mount binds the open signal so external writes re-render the host.
func (p *PopoverWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	p.AddBinding(state.BindToScheduler(p.open, p, sched))
}

// Unmount implements widget.Lifecycle; bindings are cleaned automatically.
func (p *PopoverWidget) Unmount() {}

// Layout sizes the host to its trigger.
func (p *PopoverWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if p.trigger == nil {
		size := c.Constrain(geometry.Sz(0, 0))
		p.SetBounds(geometry.FromPointSize(p.Position(), size))
		return size
	}
	size := p.trigger.Layout(ctx, c)
	ovlSetBounds(p.trigger, geometry.NewRect(0, 0, size.Width, size.Height))
	p.SetBounds(geometry.FromPointSize(p.Position(), size))
	return size
}

// Draw renders the trigger and reconciles the overlay with the open
// signal (push on open, remove on close).
func (p *PopoverWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !p.IsVisible() {
		return
	}
	if p.trigger != nil {
		bounds := p.Bounds()
		canvas.PushTransform(bounds.Min)
		widget.StampScreenOrigin(p.trigger, canvas)
		widget.DrawChild(p.trigger, ctx, canvas)
		canvas.PopTransform()
	}
	p.syncOverlay(ctx)
}

// Event forwards input to the trigger; an unconsumed left release inside
// the trigger area toggles the popover (so plain widgets work as triggers
// without PopoverTrigger sugar).
func (p *PopoverWidget) Event(ctx widget.Context, e event.Event) bool {
	before := p.open.Get()
	consumed := false
	offset := p.Bounds().Min
	if p.trigger != nil {
		consumed = p.trigger.Event(ctx, ovlTranslate(e, offset))
	}
	if !consumed {
		if me, ok := e.(*event.MouseEvent); ok &&
			me.MouseType == event.MouseRelease && me.Button == event.ButtonLeft &&
			p.Bounds().Contains(me.Position) {
			p.open.Set(!before)
			consumed = true
		}
	}
	// Any open-state flip caused by this interaction (PopoverTrigger sugar
	// or the fallback toggle above) fires the observer.
	if after := p.open.Get(); after != before && p.onOpenChange != nil {
		p.onOpenChange(after)
	}
	p.syncOverlay(ctx)
	return consumed
}

// Children returns the inline trigger; the content lives in the overlay.
func (p *PopoverWidget) Children() []widget.Widget {
	if p.trigger == nil {
		return nil
	}
	return []widget.Widget{p.trigger}
}

// syncOverlay pushes or removes the content overlay to match the open
// signal (Report 2 section 4.4).
func (p *PopoverWidget) syncOverlay(ctx widget.Context) {
	if ctx == nil || p.content == nil {
		return
	}
	om := ctx.OverlayManager()
	if om == nil {
		return
	}
	want := p.open.Get()
	if want == p.shown {
		return
	}
	if want {
		size := p.content.Layout(ctx, geometry.Loose(ctx.WindowSize()))
		pos := corepopover.CalculatePosition(
			corepopover.Bottom, p.anchorBounds(), size, ctx.WindowSize(),
			metrics.Popover.SideOffset)
		p.content.SetBounds(geometry.FromPointSize(pos, size))
		p.om = om
		om.PushOverlay(p.content, p.handleDismiss)
		p.shown = true
	} else {
		om.RemoveOverlay(p.content)
		p.shown = false
	}
	p.SetNeedsRedraw(true)
}

// handleDismiss reacts to overlay dismissal (click outside or Escape).
func (p *PopoverWidget) handleDismiss() {
	if !p.shown {
		return
	}
	p.shown = false
	if p.om != nil {
		p.om.RemoveOverlay(p.content)
	}
	if p.open.Get() {
		p.open.Set(false)
		if p.onOpenChange != nil {
			p.onOpenChange(false)
		}
	}
	p.SetNeedsRedraw(true)
}

// anchorBounds returns the screen-space rect the content anchors to.
func (p *PopoverWidget) anchorBounds() geometry.Rect {
	if sb, ok := p.trigger.(interface{ ScreenBounds() geometry.Rect }); ok {
		if r := sb.ScreenBounds(); !r.IsEmpty() {
			return r
		}
	}
	if r := p.ScreenBounds(); !r.IsEmpty() {
		return r
	}
	return p.Bounds()
}

// PopoverTriggerWidget wires a wrapped widget's click to an open signal —
// the literal shadcn <PopoverTrigger> shape (DESIGN.md section 4.9).
type PopoverTriggerWidget struct {
	widget.WidgetBase

	child widget.Widget
	open  state.Signal[bool]
}

// PopoverTrigger wraps w so a left click toggles the open signal.
func PopoverTrigger(w Widget, open state.Signal[bool]) *PopoverTriggerWidget {
	t := &PopoverTriggerWidget{child: w, open: open}
	t.SetVisible(true)
	t.SetEnabled(true)
	ovlSetParent(w, t)
	return t
}

// Layout sizes the trigger to its child.
func (t *PopoverTriggerWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if t.child == nil {
		size := c.Constrain(geometry.Sz(0, 0))
		t.SetBounds(geometry.FromPointSize(t.Position(), size))
		return size
	}
	size := t.child.Layout(ctx, c)
	ovlSetBounds(t.child, geometry.NewRect(0, 0, size.Width, size.Height))
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

// Draw renders the wrapped child.
func (t *PopoverTriggerWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !t.IsVisible() || t.child == nil {
		return
	}
	bounds := t.Bounds()
	canvas.PushTransform(bounds.Min)
	widget.StampScreenOrigin(t.child, canvas)
	widget.DrawChild(t.child, ctx, canvas)
	canvas.PopTransform()
}

// Event forwards to the child; an unconsumed left release inside the
// trigger toggles the signal.
func (t *PopoverTriggerWidget) Event(ctx widget.Context, e event.Event) bool {
	if t.child != nil && t.child.Event(ctx, ovlTranslate(e, t.Bounds().Min)) {
		return true
	}
	if me, ok := e.(*event.MouseEvent); ok &&
		me.MouseType == event.MouseRelease && me.Button == event.ButtonLeft &&
		t.Bounds().Contains(me.Position) && t.open != nil {
		t.open.Set(!t.open.Get())
		return true
	}
	return false
}

// Children returns the wrapped child.
func (t *PopoverTriggerWidget) Children() []widget.Widget {
	if t.child == nil {
		return nil
	}
	return []widget.Widget{t.child}
}

// PopoverContentWidget is the floating popover surface: w-72, p-4,
// rounded-lg, 1px border, bg-popover, shadow-md.
type PopoverContentWidget struct {
	widget.WidgetBase

	children []widget.Widget
	width    float32 // 0 = metrics.Popover.Width

	theme *theme.Theme
}

// PopoverContent stacks children vertically inside the shadcn popover
// surface.
func PopoverContent(children ...Widget) *PopoverContentWidget {
	c := &PopoverContentWidget{children: children}
	c.SetVisible(true)
	c.SetEnabled(true)
	for _, ch := range children {
		ovlSetParent(ch, c)
	}
	return c
}

// W overrides the content width (shadcn default w-72 = 288px).
func (c *PopoverContentWidget) W(px float32) *PopoverContentWidget {
	c.width = px
	return c
}

// Theme pins a specific theme instead of the process-wide current theme.
func (c *PopoverContentWidget) Theme(th *theme.Theme) *PopoverContentWidget {
	c.theme = th
	return c
}

func (c *PopoverContentWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// Layout stacks children top-to-bottom inside the p-4 padding at the
// fixed w-72 width, preserving the pre-positioned overlay origin.
func (c *PopoverContentWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	w := c.width
	if w <= 0 {
		w = metrics.Popover.Width
	}
	pad := metrics.Popover.Padding
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
func (c *PopoverContentWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	tok := c.resolvedTheme().Active()
	r := c.resolvedTheme().RadiusLG() // content: rounded-lg
	bounds := c.Bounds()

	draw.Shadow(canvas, bounds, r, metrics.ShadowMD)
	canvas.DrawRoundRect(bounds, tok.Popover, r)
	draw.InsideBorder(canvas, bounds, r, tok.Border, metrics.Popover.BorderWidth)

	canvas.PushTransform(bounds.Min)
	for _, ch := range c.children {
		widget.StampScreenOrigin(ch, canvas)
		widget.DrawChild(ch, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event dispatches to children (top-most first), translating positions
// into content-local space.
func (c *PopoverContentWidget) Event(ctx widget.Context, e event.Event) bool {
	offset := c.Bounds().Min
	for i := len(c.children) - 1; i >= 0; i-- {
		if c.children[i].Event(ctx, ovlTranslate(e, offset)) {
			return true
		}
	}
	return false
}

// Children returns the stacked content children.
func (c *PopoverContentWidget) Children() []widget.Widget { return c.children }

// ovlSetParent links child to parent for dirty propagation (the
// primitives.Box adoptChild pattern).
func ovlSetParent(child widget.Widget, parent widget.Widget) {
	if child == nil {
		return
	}
	if ps, ok := child.(interface{ SetParent(widget.Widget) }); ok {
		ps.SetParent(parent)
	}
}

// ovlSetBounds sets bounds on widgets exposing SetBounds.
func ovlSetBounds(w widget.Widget, r geometry.Rect) {
	if sb, ok := w.(interface{ SetBounds(geometry.Rect) }); ok {
		sb.SetBounds(r)
	}
}

// ovlTranslate shifts positional events into child space.
func ovlTranslate(e event.Event, offset geometry.Point) event.Event {
	switch ev := e.(type) {
	case *event.MouseEvent:
		local := *ev
		local.Position = ev.Position.Sub(offset)
		return &local
	case *event.WheelEvent:
		local := *ev
		local.Position = ev.Position.Sub(offset)
		return &local
	default:
		return e
	}
}

var (
	_ widget.Widget    = (*PopoverWidget)(nil)
	_ widget.Lifecycle = (*PopoverWidget)(nil)
	_ widget.Widget    = (*PopoverTriggerWidget)(nil)
	_ widget.Widget    = (*PopoverContentWidget)(nil)
)

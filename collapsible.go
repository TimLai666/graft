package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// CollapsibleWidget is the shadcn Collapsible: an interactive container that
// shows or hides its content via the trigger.
//
// Architecture decision (DESIGN.md sections 3.1/3.2): Collapsible is
// graft-OWNED, not a wrap of core/collapsible. shadcn's Collapsible is an
// unstyled Radix primitive — it has no chrome, header model, or painter of
// its own; the trigger and content are arbitrary caller-supplied widgets
// (CollapsibleTrigger renders asChild). core/collapsible instead owns a
// styled header widget and paints through painters.Collapsible, an anatomy
// that cannot express shadcn's "any widget as trigger, any widget as content"
// shape. So Collapsible owns only the open/closed show-hide of its content
// and the collapse layout; it draws no chrome itself.
//
// Usage mirrors shadcn:
//
//	graft.Collapsible(
//	    graft.Button("Toggle").Ghost(),
//	    graft.Card(graft.CardContent(graft.Text("Hidden details"))),
//	).Open(true).OnOpenChange(func(o bool) { ... })
//
// The collapse transition is the shadcn collapsible-down/up keyframe (0.2s
// ease-out, metrics.Collapsible.DurationMS); graft renders the settled state.
type CollapsibleWidget struct {
	widget.WidgetBase

	st      *collapsibleState
	trigger Widget
	content *CollapsibleContentWidget
}

// collapsibleState is the open/closed state shared by the root, trigger, and
// content of one Collapsible tree.
type collapsibleState struct {
	open         bool
	sig          state.Signal[bool]
	onOpenChange func(bool)
	disabled     bool
	th           *theme.Theme

	// invalidate is set by the root so triggers can request a relayout when
	// the open state flips (the content height changes).
	invalidate func()
}

// isOpen reports the current open state (bound signal wins).
func (st *collapsibleState) isOpen() bool {
	if st.sig != nil {
		return st.sig.Get()
	}
	return st.open
}

// toggle flips the open state, writing through to the bound signal and firing
// the observer.
func (st *collapsibleState) toggle() {
	st.set(!st.isOpen())
}

// set updates the open state (writing through to the bound signal when
// controlled) and fires OnOpenChange.
func (st *collapsibleState) set(open bool) {
	if st.sig != nil {
		st.sig.Set(open)
	} else {
		st.open = open
	}
	if st.onOpenChange != nil {
		st.onOpenChange(open)
	}
	if st.invalidate != nil {
		st.invalidate()
	}
}

// Collapsible assembles a collapsible from a trigger widget and a content
// widget. The content is wrapped in a CollapsibleContent region (pass an
// already-built CollapsibleContent to keep an explicit handle).
func Collapsible(trigger, content Widget) *CollapsibleWidget {
	st := &collapsibleState{th: CurrentTheme()}

	cw, ok := content.(*CollapsibleContentWidget)
	if !ok {
		cw = CollapsibleContent(content)
	}
	cw.st = st

	c := &CollapsibleWidget{
		st:      st,
		trigger: trigger,
		content: cw,
	}
	c.SetVisible(true)
	c.SetEnabled(true)

	// Wire the trigger: an explicit CollapsibleTrigger shares state; any
	// other widget is made into one transparently so its clicks toggle.
	switch t := trigger.(type) {
	case *CollapsibleTriggerWidget:
		t.st = st
	default:
		wrap := CollapsibleTrigger(trigger)
		wrap.st = st
		c.trigger = wrap
	}

	for _, child := range []Widget{c.trigger, c.content} {
		if ps, ok := child.(interface{ SetParent(widget.Widget) }); ok {
			ps.SetParent(c)
		}
	}
	return c
}

// Open sets the initial open state (uncontrolled).
func (c *CollapsibleWidget) Open(v bool) *CollapsibleWidget {
	c.st.open = v
	return c
}

// Bind makes the open state controlled by a boolean signal.
func (c *CollapsibleWidget) Bind(sig state.Signal[bool]) *CollapsibleWidget {
	c.st.sig = sig
	return c
}

// OnOpenChange registers an observer fired whenever the open state changes.
func (c *CollapsibleWidget) OnOpenChange(fn func(bool)) *CollapsibleWidget {
	c.st.onOpenChange = fn
	return c
}

// Disabled disables the trigger (no toggling, not focusable).
func (c *CollapsibleWidget) Disabled(v bool) *CollapsibleWidget {
	c.st.disabled = v
	return c
}

// Theme pins a specific theme instead of the snapshotted current theme.
func (c *CollapsibleWidget) Theme(th *theme.Theme) *CollapsibleWidget {
	if th != nil {
		c.st.th = th
	}
	return c
}

// Layout stacks the trigger and (when open) the content vertically.
func (c *CollapsibleWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	c.st.invalidate = func() {
		c.SetNeedsRedraw(true)
		if ctx != nil {
			ctx.Invalidate()
		}
	}

	loose := cons.Loosen()
	trigSz := c.trigger.Layout(ctx, loose)
	setChildBounds(c.trigger, geometry.FromPointSize(geometry.Pt(0, 0), trigSz))

	y := trigSz.Height
	maxW := trigSz.Width

	contentSz := c.content.Layout(ctx, loose)
	if contentSz.Height > 0 {
		y += metrics.Collapsible.ContentGap
		setChildBounds(c.content, geometry.FromPointSize(geometry.Pt(0, y), contentSz))
		y += contentSz.Height
		if contentSz.Width > maxW {
			maxW = contentSz.Width
		}
	} else {
		setChildBounds(c.content, geometry.FromPointSize(geometry.Pt(0, y), geometry.Sz(0, 0)))
	}

	size := cons.Constrain(geometry.Sz(maxW, y))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw paints the trigger and the (open) content.
func (c *CollapsibleWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	canvas.PushTransform(c.Bounds().Min)
	widget.StampScreenOrigin(c.trigger, canvas)
	widget.DrawChild(c.trigger, ctx, canvas)
	if c.st.isOpen() {
		widget.StampScreenOrigin(c.content, canvas)
		widget.DrawChild(c.content, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event dispatches input to the trigger and the open content.
func (c *CollapsibleWidget) Event(ctx widget.Context, e event.Event) bool {
	if !c.IsVisible() || !c.IsEnabled() {
		return false
	}
	children := []widget.Widget{c.trigger}
	if c.st.isOpen() {
		children = append(children, c.content)
	}
	return dispatchToChildren(ctx, e, c.Bounds(), children)
}

// Children returns the trigger and content.
func (c *CollapsibleWidget) Children() []widget.Widget {
	return []widget.Widget{c.trigger, c.content}
}

// Mount binds the controlled-open signal for push invalidation.
func (c *CollapsibleWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil || c.st.sig == nil {
		return
	}
	c.AddBinding(state.BindToScheduler(c.st.sig, c, sched))
}

// Unmount implements widget.Lifecycle; bindings clean up automatically.
func (c *CollapsibleWidget) Unmount() {}

// ── CollapsibleTrigger ───────────────────────────────────────────────────

// CollapsibleTriggerWidget wraps an arbitrary widget so that clicking it
// (or pressing Enter/Space when focused) toggles the collapsible. It draws
// no chrome of its own — shadcn's CollapsibleTrigger renders asChild.
type CollapsibleTriggerWidget struct {
	widget.WidgetBase

	st    *collapsibleState
	child Widget
}

// CollapsibleTrigger wraps a widget as the toggle for a Collapsible.
func CollapsibleTrigger(child Widget) *CollapsibleTriggerWidget {
	t := &CollapsibleTriggerWidget{child: child}
	t.SetVisible(true)
	t.SetEnabled(true)
	if ps, ok := child.(interface{ SetParent(widget.Widget) }); ok {
		ps.SetParent(t)
	}
	return t
}

// Layout sizes to the wrapped child.
func (t *CollapsibleTriggerWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	sz := t.child.Layout(ctx, cons)
	setChildBounds(t.child, geometry.FromPointSize(geometry.Pt(0, 0), sz))
	t.SetBounds(geometry.FromPointSize(t.Position(), sz))
	return sz
}

// Draw paints the wrapped child.
func (t *CollapsibleTriggerWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !t.IsVisible() {
		return
	}
	canvas.PushTransform(t.Bounds().Min)
	widget.StampScreenOrigin(t.child, canvas)
	widget.DrawChild(t.child, ctx, canvas)
	canvas.PopTransform()
}

// Event toggles the collapsible on a left click inside bounds or on
// Enter/Space when focused, after offering the event to the wrapped child
// (so an inner Button's own hover/focus visuals still work).
func (t *CollapsibleTriggerWidget) Event(ctx widget.Context, e event.Event) bool {
	if t.st != nil && t.st.disabled {
		return false
	}
	// Offer to the child first (translated into trigger-local space).
	childHandled := dispatchToChildren(ctx, e, t.Bounds(), []widget.Widget{t.child})

	switch ev := e.(type) {
	case *event.MouseEvent:
		if ev.MouseType == event.MouseRelease && ev.Button == event.ButtonLeft &&
			t.Bounds().Contains(ev.Position) {
			if t.st != nil {
				t.st.toggle()
			}
			return true
		}
	case *event.KeyEvent:
		if t.IsFocused() && ev.KeyType == event.KeyPress &&
			(ev.Key == event.KeyEnter || ev.Key == event.KeySpace) {
			if t.st != nil {
				t.st.toggle()
			}
			return true
		}
	}
	return childHandled
}

// Children returns the wrapped child.
func (t *CollapsibleTriggerWidget) Children() []widget.Widget {
	return []widget.Widget{t.child}
}

// ── CollapsibleContent ───────────────────────────────────────────────────

// CollapsibleContentWidget is the region of a Collapsible that expands and
// collapses. When closed it lays out to zero height (overflow-hidden), so it
// is skipped by drawing and focus traversal.
type CollapsibleContentWidget struct {
	widget.WidgetBase

	st    *collapsibleState
	child Widget
}

// CollapsibleContent wraps a widget as the expandable region of a Collapsible.
func CollapsibleContent(child Widget) *CollapsibleContentWidget {
	c := &CollapsibleContentWidget{child: child}
	c.SetVisible(true)
	c.SetEnabled(true)
	if ps, ok := child.(interface{ SetParent(widget.Widget) }); ok {
		ps.SetParent(c)
	}
	return c
}

// Layout sizes to the child when open, zero when closed; the hidden child is
// marked invisible so focus traversal skips it.
func (c *CollapsibleContentWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	open := c.st != nil && c.st.isOpen()
	if v, ok := c.child.(interface{ SetVisible(bool) }); ok {
		v.SetVisible(open)
	}
	if !open {
		c.SetBounds(geometry.FromPointSize(c.Position(), geometry.Sz(0, 0)))
		return geometry.Sz(0, 0)
	}
	sz := c.child.Layout(ctx, cons)
	setChildBounds(c.child, geometry.FromPointSize(geometry.Pt(0, 0), sz))
	c.SetBounds(geometry.FromPointSize(c.Position(), sz))
	return sz
}

// Draw renders the child when open.
func (c *CollapsibleContentWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() || c.st == nil || !c.st.isOpen() {
		return
	}
	canvas.PushTransform(c.Bounds().Min)
	widget.StampScreenOrigin(c.child, canvas)
	widget.DrawChild(c.child, ctx, canvas)
	canvas.PopTransform()
}

// Event forwards to the child when open.
func (c *CollapsibleContentWidget) Event(ctx widget.Context, e event.Event) bool {
	if !c.IsVisible() || c.st == nil || !c.st.isOpen() {
		return false
	}
	return dispatchToChildren(ctx, e, c.Bounds(), []widget.Widget{c.child})
}

// Children returns the wrapped child.
func (c *CollapsibleContentWidget) Children() []widget.Widget {
	return []widget.Widget{c.child}
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*CollapsibleWidget)(nil)
	_ widget.Lifecycle = (*CollapsibleWidget)(nil)
	_ widget.Widget    = (*CollapsibleTriggerWidget)(nil)
	_ widget.Widget    = (*CollapsibleContentWidget)(nil)
)

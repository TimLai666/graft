package graft

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// CarouselOrientation selects the slide direction.
type CarouselOrientation uint8

const (
	// CarouselHorizontal slides left/right (default).
	CarouselHorizontal CarouselOrientation = iota

	// CarouselVertical slides up/down.
	CarouselVertical
)

// CarouselWidget is the shadcn Carousel: a swipeable content carousel with
// Previous/Next navigation buttons, keyboard control, optional auto-play,
// and smooth slide-transition animation.
//
// OWNED widget (DESIGN.md section 3.1): the carousel is implemented
// directly on internal/draw + metrics + theme tokens. It manages N child
// items, clips to the viewport, and animates a slide offset driven by
// animation.Controller ticked in Draw (not Layout, since Layout is not
// called every frame).
//
// Usage:
//
//	graft.Carousel(
//	    graft.CarouselItem(widget1),
//	    graft.CarouselItem(widget2),
//	    graft.CarouselItem(widget3),
//	).Orientation(CarouselHorizontal)
type CarouselWidget struct {
	widget.WidgetBase

	items       []*CarouselItemWidget
	orientation CarouselOrientation
	th          *theme.Theme

	// current is the 0-based index of the active slide.
	current int
	sig     state.Signal[int]

	// autoPlay configuration.
	autoPlay     bool
	autoInterval time.Duration
	autoElapsed  time.Duration

	// loop wraps prev/next around the ends (shadcn opts.loop).
	loop bool

	// Animation state: offset is the normalized slide offset where 0.0
	// means fully showing the current slide, -1.0 is one slide left,
	// +1.0 is one slide right. During animation it tweens between the
	// old and new positions.
	offset   float32
	animCtrl *animation.Controller
	animAdpt carouselProgress

	// Nav button interaction state.
	prevHovered bool
	nextHovered bool
	prevPressed bool
	nextPressed bool

	// Focus state.
	focused      bool
	focusVisible bool
	pointerFocus bool
}

// Carousel creates a carousel with the given items.
func Carousel(items ...*CarouselItemWidget) *CarouselWidget {
	c := &CarouselWidget{
		th:       CurrentTheme(),
		items:    items,
		animCtrl: animation.NewController(),
	}
	c.animAdpt.w = c
	c.SetVisible(true)
	c.SetEnabled(true)
	for _, it := range items {
		it.carousel = c
		it.SetParent(c)
	}
	return c
}

// Orientation sets the slide direction. Default is CarouselHorizontal.
func (c *CarouselWidget) Orientation(o CarouselOrientation) *CarouselWidget {
	c.orientation = o
	return c
}

// Vertical is a convenience for Orientation(CarouselVertical).
func (c *CarouselWidget) Vertical() *CarouselWidget {
	return c.Orientation(CarouselVertical)
}

// Bind makes the current slide index controlled by a signal.
func (c *CarouselWidget) Bind(sig state.Signal[int]) *CarouselWidget {
	c.sig = sig
	c.current = sig.Get()
	return c
}

// Index sets the initially selected slide index (0-based, uncontrolled).
func (c *CarouselWidget) Index(idx int) *CarouselWidget {
	if idx >= 0 && idx < len(c.items) {
		c.current = idx
	}
	return c
}

// AutoPlay enables auto-advancing slides every interval. Pass 0 for the
// default interval (3 seconds).
func (c *CarouselWidget) AutoPlay(interval time.Duration) *CarouselWidget {
	c.autoPlay = true
	if interval <= 0 {
		interval = 3 * time.Second
	}
	c.autoInterval = interval
	return c
}

// Loop wraps the carousel so Previous on the first slide goes to the last and
// Next on the last goes to the first (shadcn opts={{ loop: true }}). The nav
// buttons stay enabled at the ends.
func (c *CarouselWidget) Loop(v bool) *CarouselWidget {
	c.loop = v
	return c
}

// Theme pins a specific theme instead of the snapshotted current theme.
func (c *CarouselWidget) Theme(th *theme.Theme) *CarouselWidget {
	c.th = th
	return c
}

// currentIndex returns the active slide index, reading from the bound
// signal when controlled.
func (c *CarouselWidget) currentIndex() int {
	if c.sig != nil {
		return c.sig.Get()
	}
	return c.current
}

// setIndex updates the current slide index, writing through to the bound
// signal when controlled.
func (c *CarouselWidget) setIndex(idx int) {
	if c.sig != nil {
		c.sig.Set(idx)
		return
	}
	c.current = idx
}

// slideCount returns the number of slides.
func (c *CarouselWidget) slideCount() int {
	return len(c.items)
}

// canPrev reports whether the Previous button is enabled. With loop, it stays
// enabled as long as there is more than one slide.
func (c *CarouselWidget) canPrev() bool {
	if c.loop {
		return c.slideCount() > 1
	}
	return c.slideCount() > 0 && c.currentIndex() > 0
}

// canNext reports whether the Next button is enabled. With loop, it stays
// enabled as long as there is more than one slide.
func (c *CarouselWidget) canNext() bool {
	if c.loop {
		return c.slideCount() > 1
	}
	return c.slideCount() > 0 && c.currentIndex() < c.slideCount()-1
}

// goTo navigates to the given slide index with animation.
func (c *CarouselWidget) goTo(ctx widget.Context, idx int) {
	n := c.slideCount()
	if n == 0 || idx < 0 || idx >= n {
		return
	}
	cur := c.currentIndex()
	if idx == cur {
		return
	}
	// Determine animation direction: offset starts from the visual
	// displacement of the old slide relative to the new one.
	from := float32(cur - idx)
	c.setIndex(idx)
	c.startAnimation(ctx, from)
	c.autoElapsed = 0
	c.SetNeedsRedraw(true)
	if ctx != nil {
		ctx.Invalidate()
	}
}

// prev navigates to the previous slide, wrapping to the last when loop is on.
func (c *CarouselWidget) prev(ctx widget.Context) {
	if !c.canPrev() {
		return
	}
	idx := c.currentIndex() - 1
	if idx < 0 && c.loop {
		idx = c.slideCount() - 1
	}
	c.goTo(ctx, idx)
}

// next navigates to the next slide, wrapping to the first when loop is on.
func (c *CarouselWidget) next(ctx widget.Context) {
	if !c.canNext() {
		return
	}
	idx := c.currentIndex() + 1
	if idx >= c.slideCount() && c.loop {
		idx = 0
	}
	c.goTo(ctx, idx)
}

// viewportRect returns the content area excluding nav button overflow.
func (c *CarouselWidget) viewportRect() geometry.Rect {
	return geometry.FromPointSize(geometry.Pt(0, 0), c.Bounds().Size())
}

// prevButtonRect returns the Previous nav button bounds in local space.
func (c *CarouselWidget) prevButtonRect() geometry.Rect {
	m := metrics.Carousel
	vp := c.viewportRect()
	if c.orientation == CarouselVertical {
		cx := vp.Center().X
		return geometry.NewRect(
			cx-m.ButtonSize/2,
			m.ButtonOffset,
			m.ButtonSize,
			m.ButtonSize,
		)
	}
	cy := vp.Center().Y
	return geometry.NewRect(
		m.ButtonOffset,
		cy-m.ButtonSize/2,
		m.ButtonSize,
		m.ButtonSize,
	)
}

// nextButtonRect returns the Next nav button bounds in local space.
func (c *CarouselWidget) nextButtonRect() geometry.Rect {
	m := metrics.Carousel
	vp := c.viewportRect()
	if c.orientation == CarouselVertical {
		cx := vp.Center().X
		return geometry.NewRect(
			cx-m.ButtonSize/2,
			vp.Height()-m.ButtonSize+(-m.ButtonOffset),
			m.ButtonSize,
			m.ButtonSize,
		)
	}
	cy := vp.Center().Y
	return geometry.NewRect(
		vp.Width()-m.ButtonSize+(-m.ButtonOffset),
		cy-m.ButtonSize/2,
		m.ButtonSize,
		m.ButtonSize,
	)
}

// Layout sizes the carousel: width fills the constraint, height wraps the
// tallest slide. Each item is laid out at the resolved viewport size so
// all slides share the same dimensions.
func (c *CarouselWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	// Width: fill available or use min constraint.
	w := cons.MaxWidth
	if w >= geometry.Infinity {
		w = cons.MinWidth
	}

	// First pass: measure each item with loose height to find the tallest.
	var maxH float32
	looseCons := geometry.Constraints{
		MinWidth: w, MaxWidth: w,
		MinHeight: 0, MaxHeight: cons.MaxHeight,
	}
	for _, it := range c.items {
		sz := it.Layout(ctx, looseCons)
		if sz.Height > maxH {
			maxH = sz.Height
		}
	}
	if maxH <= 0 {
		maxH = cons.MinHeight
	}

	size := cons.Constrain(geometry.Sz(w, maxH))

	// Second pass: set all items to the resolved viewport size.
	for _, it := range c.items {
		it.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	}

	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw renders the active slide (with animation offset), nav buttons, and
// focus ring. The animation is ticked here (Draw runs every repaint;
// Layout does not).
func (c *CarouselWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.th
	tok := th.Active()
	dark := th.IsDark()
	bounds := c.Bounds()

	// Tick the slide animation.
	c.tickAnimation(ctx)
	if c.animCtrl == nil || !c.animCtrl.HasActive() {
		c.offset = 0
	}

	// Tick auto-play.
	if c.autoPlay && c.slideCount() > 1 {
		dt := ctx.DeltaTime()
		c.autoElapsed += dt
		if c.autoElapsed >= c.autoInterval {
			c.autoElapsed = 0
			nextIdx := c.currentIndex() + 1
			if nextIdx >= c.slideCount() {
				nextIdx = 0
			}
			c.goTo(ctx, nextIdx)
		}
		// Keep repainting while auto-play is active.
		c.SetNeedsRedraw(true)
		ctx.Invalidate()
	}

	canvas.PushTransform(bounds.Min)

	// Clip to viewport.
	vp := c.viewportRect()
	canvas.PushClip(vp)

	// Draw the current slide (and neighbors during animation).
	cur := c.currentIndex()
	n := c.slideCount()
	for i := -1; i <= 1; i++ {
		idx := cur + i
		if idx < 0 || idx >= n {
			continue
		}
		// Should this slide be visible? Only if it has visual overlap.
		slideOffset := float32(i) + c.offset
		var tx, ty float32
		if c.orientation == CarouselVertical {
			ty = slideOffset * (vp.Height() + metrics.Carousel.Gap)
		} else {
			tx = slideOffset * (vp.Width() + metrics.Carousel.Gap)
		}
		canvas.PushTransform(geometry.Pt(tx, ty))
		widget.StampScreenOrigin(c.items[idx], canvas)
		widget.DrawChild(c.items[idx], ctx, canvas)
		canvas.PopTransform()
	}

	canvas.PopClip()

	// Draw nav buttons.
	if c.slideCount() > 1 {
		c.drawNavButton(canvas, tok, dark, c.prevButtonRect(), true)
		c.drawNavButton(canvas, tok, dark, c.nextButtonRect(), false)
	}

	// Focus ring around the viewport.
	if c.focusVisible {
		radius := th.RadiusMD()
		ring := draw.Alpha(tok.Ring, metrics.RingAlpha)
		draw.FocusRing(canvas, vp, radius, ring)
	}

	canvas.PopTransform()
}

// drawNavButton renders a circular outline nav button (prev or next).
func (c *CarouselWidget) drawNavButton(canvas widget.Canvas, tok *theme.Tokens, dark bool, rect geometry.Rect, isPrev bool) {
	m := metrics.Carousel
	radius := m.ButtonSize / 2 // rounded-full

	hovered := (isPrev && c.prevHovered) || (!isPrev && c.nextHovered)
	pressed := (isPrev && c.prevPressed) || (!isPrev && c.nextPressed)
	disabled := (isPrev && !c.canPrev()) || (!isPrev && !c.canNext())

	active := (hovered || pressed) && !disabled

	// Shadow (outline button: shadow-xs).
	if !disabled {
		draw.Shadow(canvas, rect, radius, metrics.ShadowXS)
	}

	// Fill: outline variant logic.
	var bg widget.Color
	if dark {
		bg = draw.MulAlpha(tok.Input, m.DarkButtonBgAlpha)
		if active {
			bg = draw.MulAlpha(tok.Input, 0.5) // dark:hover:bg-input/50
		}
	} else {
		bg = tok.Background
		if active {
			bg = tok.Accent
		}
	}
	canvas.DrawRoundRect(rect, draw.Fade(bg, disabled), radius)

	// Border.
	var borderColor widget.Color
	if dark {
		borderColor = tok.Input
	} else {
		borderColor = tok.Border
	}
	draw.InsideBorder(canvas, rect, radius, draw.Fade(borderColor, disabled), m.ButtonBorderWidth)

	// Icon.
	var ic icon.IconData
	if isPrev {
		if c.orientation == CarouselVertical {
			ic = icons.ChevronUp
		} else {
			ic = icons.ArrowLeft
		}
	} else {
		if c.orientation == CarouselVertical {
			ic = icons.ChevronDown
		} else {
			ic = icons.ArrowRight
		}
	}
	iconRect := geometry.NewRect(
		rect.Min.X+(rect.Width()-m.ButtonIconSize)/2,
		rect.Min.Y+(rect.Height()-m.ButtonIconSize)/2,
		m.ButtonIconSize,
		m.ButtonIconSize,
	)
	var fg widget.Color
	if active {
		fg = tok.AccentForeground
	} else {
		fg = tok.Foreground
	}
	icon.Draw(canvas, ic, iconRect, draw.Fade(fg, disabled))
}

// startAnimation tweens offset from the given start value toward 0 (the
// settled position showing the current slide).
func (c *CarouselWidget) startAnimation(ctx widget.Context, from float32) {
	if c.animCtrl == nil {
		c.offset = 0
		return
	}
	c.animAdpt.ctx = ctx
	animation.To(&c.animAdpt, 0).
		From(from).
		Duration(metrics.Carousel.AnimDuration).
		Ease(animation.CubicBezier(0.4, 0, 0.2, 1)).
		Start(c.animCtrl)
}

// tickAnimation advances the slide animation while active.
func (c *CarouselWidget) tickAnimation(ctx widget.Context) {
	if c.animCtrl == nil || !c.animCtrl.HasActive() {
		return
	}
	dt := ctx.DeltaTime()
	if dt < time.Millisecond {
		dt = time.Millisecond
	}
	if dt > 32*time.Millisecond {
		dt = 32 * time.Millisecond
	}
	c.animCtrl.Tick(dt)
	if c.animCtrl.HasActive() {
		c.SetNeedsRedraw(true)
		ctx.Invalidate()
	}
}

// Event handles mouse interaction with nav buttons, keyboard navigation
// (Left/Right or Up/Down), and focus tracking.
func (c *CarouselWidget) Event(ctx widget.Context, e event.Event) bool {
	if !c.IsVisible() || !c.IsEnabled() {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return c.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		return c.keyEvent(ctx, ev)
	case *event.FocusEvent:
		c.SetFocused(ev.FocusType == event.FocusGained)
		return false
	}
	return false
}

func (c *CarouselWidget) mouseEvent(ctx widget.Context, ev *event.MouseEvent) bool {
	local := ev.Position.Sub(c.Bounds().Min)
	prevRect := c.prevButtonRect()
	nextRect := c.nextButtonRect()
	inPrev := prevRect.Contains(local)
	inNext := nextRect.Contains(local)

	switch ev.MouseType {
	case event.MouseEnter:
		return true
	case event.MouseLeave:
		changed := c.prevHovered || c.nextHovered
		c.prevHovered = false
		c.nextHovered = false
		if changed {
			ctx.SetCursor(widget.CursorDefault)
			c.SetNeedsRedraw(true)
			ctx.InvalidateRect(c.Bounds())
		}
		return true
	case event.MousePress:
		if ev.Button != event.ButtonLeft {
			return false
		}
		c.pointerFocus = true
		ctx.RequestFocus(c)
		c.pointerFocus = false
		if inPrev && c.canPrev() {
			c.prevPressed = true
			c.SetNeedsRedraw(true)
			ctx.InvalidateRect(c.Bounds())
			return true
		}
		if inNext && c.canNext() {
			c.nextPressed = true
			c.SetNeedsRedraw(true)
			ctx.InvalidateRect(c.Bounds())
			return true
		}
		return true
	case event.MouseRelease:
		if c.prevPressed && inPrev {
			c.prev(ctx)
		}
		if c.nextPressed && inNext {
			c.next(ctx)
		}
		c.prevPressed = false
		c.nextPressed = false
		c.SetNeedsRedraw(true)
		ctx.InvalidateRect(c.Bounds())
		return true
	}

	// Handle hover state for nav buttons (mouse move within the widget).
	prevWas, nextWas := c.prevHovered, c.nextHovered
	c.prevHovered = inPrev
	c.nextHovered = inNext
	if c.prevHovered != prevWas || c.nextHovered != nextWas {
		if c.prevHovered || c.nextHovered {
			ctx.SetCursor(widget.CursorPointer)
		} else {
			ctx.SetCursor(widget.CursorDefault)
		}
		c.SetNeedsRedraw(true)
		ctx.InvalidateRect(c.Bounds())
	}
	return false
}

func (c *CarouselWidget) keyEvent(ctx widget.Context, ev *event.KeyEvent) bool {
	if !c.focused {
		return false
	}
	if ev.KeyType != event.KeyPress && ev.KeyType != event.KeyRepeat {
		return false
	}
	if c.orientation == CarouselVertical {
		switch ev.Key {
		case event.KeyUp:
			c.prev(ctx)
			return true
		case event.KeyDown:
			c.next(ctx)
			return true
		}
	} else {
		switch ev.Key {
		case event.KeyLeft:
			c.prev(ctx)
			return true
		case event.KeyRight:
			c.next(ctx)
			return true
		}
	}
	return false
}

// SetFocused tracks focus-visible semantics.
func (c *CarouselWidget) SetFocused(focused bool) {
	c.focused = focused
	c.focusVisible = focused && !c.pointerFocus
	c.WidgetBase.SetFocused(focused)
	c.MarkRedrawLocal()
}

// IsFocusable reports whether the carousel can receive keyboard focus.
func (c *CarouselWidget) IsFocusable() bool {
	return c.IsVisible() && c.IsEnabled()
}

// Children returns the carousel items.
func (c *CarouselWidget) Children() []widget.Widget {
	out := make([]widget.Widget, len(c.items))
	for i, it := range c.items {
		out[i] = it
	}
	return out
}

// Mount registers the bound signal for push invalidation.
func (c *CarouselWidget) Mount(ctx widget.Context) {
	if sched := ctx.Scheduler(); sched != nil && c.sig != nil {
		c.AddBinding(state.BindToScheduler(c.sig, c, sched))
	}
}

// Unmount cancels any running animation.
func (c *CarouselWidget) Unmount() {
	if c.animCtrl != nil {
		c.animCtrl.CancelAll()
	}
}

// carouselProgress adapts the widget's offset field to animation.To's
// signalFloat32 interface, marking the widget dirty on each frame.
type carouselProgress struct {
	w   *CarouselWidget
	ctx widget.Context
}

func (a *carouselProgress) Get() float32 { return a.w.offset }

func (a *carouselProgress) Set(v float32) {
	a.w.offset = v
	a.w.SetNeedsRedraw(true)
	if a.ctx != nil {
		a.ctx.InvalidateRect(a.w.Bounds())
	}
}

// ── CarouselItem ──────────────────────────────────────────────────────

// CarouselItemWidget wraps a single slide's content.
type CarouselItemWidget struct {
	widget.WidgetBase

	content  Widget
	carousel *CarouselWidget
}

// CarouselItem wraps a widget as a carousel slide.
func CarouselItem(content Widget) *CarouselItemWidget {
	it := &CarouselItemWidget{content: content}
	it.SetVisible(true)
	it.SetEnabled(true)
	if ps, ok := content.(interface{ SetParent(widget.Widget) }); ok {
		ps.SetParent(it)
	}
	return it
}

// Layout sizes the item to the given constraints and positions the content.
func (it *CarouselItemWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	sz := it.content.Layout(ctx, c)
	setChildBounds(it.content, geometry.FromPointSize(geometry.Pt(0, 0), sz))
	it.SetBounds(geometry.FromPointSize(it.Position(), sz))
	return sz
}

// Draw renders the item's content.
func (it *CarouselItemWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !it.IsVisible() {
		return
	}
	canvas.PushTransform(it.Bounds().Min)
	widget.StampScreenOrigin(it.content, canvas)
	widget.DrawChild(it.content, ctx, canvas)
	canvas.PopTransform()
}

// Event forwards events to the content.
func (it *CarouselItemWidget) Event(ctx widget.Context, e event.Event) bool {
	if !it.IsVisible() || !it.IsEnabled() {
		return false
	}
	return dispatchToChildren(ctx, e, it.Bounds(), []widget.Widget{it.content})
}

// Children returns the wrapped content.
func (it *CarouselItemWidget) Children() []widget.Widget {
	return []widget.Widget{it.content}
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*CarouselWidget)(nil)
	_ widget.Focusable = (*CarouselWidget)(nil)
	_ widget.Lifecycle = (*CarouselWidget)(nil)
	_ widget.Widget    = (*CarouselItemWidget)(nil)
)

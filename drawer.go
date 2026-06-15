package graft

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// DrawerSide selects the viewport edge the drawer panel anchors to.
type DrawerSide uint8

// Drawer sides. The default is Bottom, matching shadcn/vaul's
// direction="bottom" default (distinct from Sheet, which defaults to right).
const (
	// DrawerBottom anchors the panel to the bottom edge (default). The panel
	// spans the full width, has rounded TOP corners and a grabber handle, and
	// slides up on open.
	DrawerBottom DrawerSide = iota
	// DrawerTop anchors the panel to the top edge (rounded bottom corners).
	DrawerTop
	// DrawerLeft anchors the panel to the left edge.
	DrawerLeft
	// DrawerRight anchors the panel to the right edge.
	DrawerRight
)

// DrawerWidget is the zero-size host that watches an open signal and pushes a
// modal, edge-anchored overlay carrying the DrawerContent
// (docs/research/03-shadcn-pixel-spec.md §Drawer).
//
// Architecture decision: graft-owned widget, modeled on the shipped Sheet
// host. It reuses Sheet's modal chassis (full-window black@50% backdrop, Esc +
// backdrop dismissal, modal swallow, Bind(signal) host, Trigger) but defaults
// to the bottom edge, paints rounded TOP corners with a centered grabber
// handle (the vaul look), and omits the close X (vaul dismisses via drag,
// backdrop, or Esc). On open the panel slides in from its edge over
// DrawerOpenDurationMillis, driven by an animation.Controller ticked in the
// overlay's Draw (the switch.go pattern); the backdrop fades in with it. The
// overlay also implements a basic drag-to-dismiss: a pointer drag on the panel
// toward its edge offsets the panel live, and a release past
// DrawerDismissDragFraction of the panel extent dismisses (otherwise it snaps
// back). Close snaps (immediate overlay removal), matching Sheet. Goldens
// render the SETTLED open state (DrawerPreview, no overlay), so the slide
// offset is zero there.
type DrawerWidget struct {
	widget.WidgetBase

	content *DrawerContentWidget
	open    state.Signal[bool]
	initial bool
	onOpen  func(bool)

	ctx     widget.Context
	overlay *drawerOverlayWidget
	shown   bool
}

// Drawer creates a drawer host for the given content. Bind an open signal with
// Bind, or set an initial state with Open; pair with a DrawerTrigger.
func Drawer(content *DrawerContentWidget) *DrawerWidget {
	d := &DrawerWidget{content: content}
	d.SetVisible(true)
	d.SetEnabled(true)
	return d
}

// Bind controls the open state from a signal (read for render, written on
// Esc/backdrop/drag-dismiss).
func (d *DrawerWidget) Bind(open state.Signal[bool]) *DrawerWidget {
	d.open = open
	return d
}

// Open sets the uncontrolled initial open state (shadcn defaultOpen).
func (d *DrawerWidget) Open(v bool) *DrawerWidget {
	d.initial = v
	return d
}

// OnOpenChange registers an observer invoked whenever the open state changes.
func (d *DrawerWidget) OnOpenChange(fn func(bool)) *DrawerWidget {
	d.onOpen = fn
	return d
}

func (d *DrawerWidget) isOpen() bool {
	if d.open != nil {
		return d.open.Get()
	}
	return d.initial
}

func (d *DrawerWidget) setOpen(v bool) {
	if d.open != nil {
		d.open.Set(v)
	} else {
		d.initial = v
	}
	if d.onOpen != nil {
		d.onOpen(v)
	}
	d.sync()
}

// Mount binds the open signal so external Set calls re-sync the overlay.
func (d *DrawerWidget) Mount(ctx widget.Context) {
	d.ctx = ctx
	if d.open != nil {
		if sched := ctx.Scheduler(); sched != nil {
			d.AddBinding(state.BindToScheduler[bool](d.open, d, sched))
		}
		state.SubscribeForever[bool](d.open, func(bool) { d.sync() })
	}
	d.sync()
}

// Unmount removes any live overlay; bindings are cleaned by WidgetBase.
func (d *DrawerWidget) Unmount() {
	if d.shown {
		d.pop()
	}
}

// Layout reports zero size; the host is invisible chrome.
func (d *DrawerWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	d.ctx = ctx
	d.SetBounds(geometry.FromPointSize(d.Position(), geometry.Sz(0, 0)))
	d.sync()
	return c.Constrain(geometry.Sz(0, 0))
}

// Draw paints nothing; the content lives in the overlay.
func (d *DrawerWidget) Draw(ctx widget.Context, _ widget.Canvas) { d.ctx = ctx }

// Event ignores input; the overlay handles its own.
func (d *DrawerWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; the host is a leaf (content lives in the overlay).
func (d *DrawerWidget) Children() []widget.Widget { return nil }

func (d *DrawerWidget) sync() {
	if d.ctx == nil {
		return
	}
	want := d.isOpen()
	if want && !d.shown {
		d.push()
	} else if !want && d.shown {
		d.pop()
	}
}

func (d *DrawerWidget) push() {
	om := d.ctx.OverlayManager()
	if om == nil {
		return
	}
	d.overlay = newDrawerOverlay(d.content, d.ctx.WindowSize(), func() { d.setOpen(false) })
	om.PushOverlay(d.overlay, func() { d.shown = false; d.overlay = nil })
	d.shown = true
	d.ctx.Invalidate()
}

func (d *DrawerWidget) pop() {
	if om := d.ctx.OverlayManager(); om != nil && d.overlay != nil {
		om.RemoveOverlay(d.overlay)
	}
	d.overlay = nil
	d.shown = false
	d.ctx.Invalidate()
}

// DrawerTrigger wraps a trigger widget so a click opens the drawer by setting
// open to true (mirrors SheetTrigger).
func DrawerTrigger(trigger widget.Widget, open state.Signal[bool]) widget.Widget {
	return &drawerTriggerWidget{trigger: trigger, open: open}
}

type drawerTriggerWidget struct {
	widget.WidgetBase
	trigger widget.Widget
	open    state.Signal[bool]
}

func (t *drawerTriggerWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := t.trigger.Layout(ctx, c)
	setWidgetBounds(t.trigger, geometry.FromPointSize(t.Position(), size))
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

func (t *drawerTriggerWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	setWidgetBounds(t.trigger, t.Bounds())
	t.trigger.Draw(ctx, canvas)
}

func (t *drawerTriggerWidget) Event(ctx widget.Context, e event.Event) bool {
	consumed := t.trigger.Event(ctx, e)
	if me, ok := e.(*event.MouseEvent); ok && me.MouseType == event.MousePress {
		if t.Bounds().Contains(me.Position) && me.Button == event.ButtonLeft {
			if t.open != nil {
				t.open.Set(true)
			}
			return true
		}
	}
	return consumed
}

func (t *drawerTriggerWidget) Children() []widget.Widget {
	return []widget.Widget{t.trigger}
}

// drawerOverlayWidget is the modal chassis: full-window black@50% backdrop with
// the content panel anchored to its side. Esc, backdrop click, and a
// drag-toward-the-edge past the dismiss threshold all dismiss.
type drawerOverlayWidget struct {
	widget.WidgetBase

	content    *DrawerContentWidget
	windowSize geometry.Size
	onDismiss  func()

	// progress is the slide-in position: 0 = panel fully off-screen at its edge,
	// 1 = settled open. It animates 0→1 on open via an animation.Controller
	// ticked in Draw (the switch.go pattern). The backdrop alpha tracks
	// progress so it fades in with the panel.
	progress    float32
	animCtrl    *animation.Controller
	animAdpt    drawerProgress
	animStarted bool

	// dragging tracks an in-flight drag-to-dismiss gesture; dragOffset is the
	// current panel translation along the drag axis (always toward the edge, so
	// >= 0 in the edge direction). dragStart records the pointer position at
	// press for delta computation.
	dragging   bool
	dragStart  geometry.Point
	dragOffset float32
}

func newDrawerOverlay(content *DrawerContentWidget, windowSize geometry.Size, onDismiss func()) *drawerOverlayWidget {
	o := &drawerOverlayWidget{content: content, windowSize: windowSize, onDismiss: onDismiss,
		animCtrl: animation.NewController()}
	o.animAdpt.o = o
	o.SetVisible(true)
	o.SetEnabled(true)
	o.AddChild(content)
	return o
}

func (o *drawerOverlayWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := c.Constrain(o.windowSize)
	o.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	o.content.layoutAnchored(ctx, size)
	return size
}

func (o *drawerOverlayWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	// Kick off the slide-in on the first paint (ctx is available here), then
	// advance it. When idle, snap fully open (progress 1) so a settled drawer is
	// always at its edge regardless of render mode.
	if !o.animStarted {
		o.animStarted = true
		o.startOpen(ctx)
	}
	o.tickAnim(ctx)

	// Backdrop fades in with the panel and dims with the drag-out distance so a
	// drag-to-dismiss visibly lightens the scrim.
	canvas.DrawRect(o.Bounds(), widget.RGBA(0, 0, 0, metrics.OverlayAlpha*o.backdropAlphaFactor()))

	// Panel slides from its edge by the remaining (1-progress) of its extent,
	// plus any live drag offset toward the edge.
	off := o.panelTranslation()
	canvas.PushTransform(off)
	o.content.Draw(ctx, canvas)
	canvas.PopTransform()
}

// panelTranslation returns the panel's current draw translation: the slide-in
// remainder (zero when settled) combined with the live drag offset, both
// directed toward the anchored edge.
func (o *drawerOverlayWidget) panelTranslation() geometry.Point {
	b := o.content.Bounds()
	slide := 1 - o.progress
	if slide < 0 {
		slide = 0
	}
	switch o.content.side {
	case DrawerTop:
		return geometry.Pt(0, -b.Height()*slide-o.dragOffset)
	case DrawerLeft:
		return geometry.Pt(-b.Width()*slide-o.dragOffset, 0)
	case DrawerRight:
		return geometry.Pt(b.Width()*slide+o.dragOffset, 0)
	default: // DrawerBottom
		return geometry.Pt(0, b.Height()*slide+o.dragOffset)
	}
}

// dragExtent returns the panel's extent along the drag/slide axis.
func (o *drawerOverlayWidget) dragExtent() float32 {
	b := o.content.Bounds()
	if o.content.horizontal() {
		return b.Width()
	}
	return b.Height()
}

// backdropAlphaFactor scales the scrim by the slide progress and shrinks it as
// the user drags the panel toward the edge.
func (o *drawerOverlayWidget) backdropAlphaFactor() float32 {
	f := o.progress
	if ext := o.dragExtent(); ext > 0 && o.dragOffset > 0 {
		f *= 1 - clampUnit(o.dragOffset/ext)
	}
	return f
}

func clampUnit(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// startOpen tweens progress 0→1 over DrawerOpenDurationMillis with the shadcn
// ease, driven by the controller ticked in Draw.
func (o *drawerOverlayWidget) startOpen(ctx widget.Context) {
	if o.animCtrl == nil {
		o.progress = 1
		return
	}
	o.animAdpt.ctx = ctx
	animation.To(&o.animAdpt, 1).
		From(0).
		Duration(time.Duration(metrics.DrawerOpenDurationMillis) * time.Millisecond).
		Ease(animation.CubicBezier(0.4, 0, 0.2, 1)).
		Start(o.animCtrl)
}

// tickAnim advances the slide-in while active; snaps to settled-open when idle.
func (o *drawerOverlayWidget) tickAnim(ctx widget.Context) {
	if o.animCtrl == nil || !o.animCtrl.HasActive() {
		o.progress = 1
		return
	}
	dt := ctx.DeltaTime()
	if dt < time.Millisecond {
		dt = time.Millisecond
	}
	if dt > 32*time.Millisecond {
		dt = 32 * time.Millisecond
	}
	o.animCtrl.Tick(dt)
	if o.animCtrl.HasActive() {
		o.SetNeedsRedraw(true)
		ctx.Invalidate()
	}
}

// drawerProgress adapts the overlay's progress field to animation.To's
// signalFloat32 interface, marking the overlay dirty each frame.
type drawerProgress struct {
	o   *drawerOverlayWidget
	ctx widget.Context
}

func (a *drawerProgress) Get() float32 { return a.o.progress }

func (a *drawerProgress) Set(v float32) {
	a.o.progress = v
	a.o.SetNeedsRedraw(true)
	if a.ctx != nil {
		a.ctx.Invalidate()
	}
}

func (o *drawerOverlayWidget) Event(ctx widget.Context, e event.Event) bool {
	// Drag-to-dismiss takes priority over child handling so a press that starts
	// on the panel begins a drag rather than being swallowed by content.
	if o.handleDrag(ctx, e) {
		return true
	}
	if o.content.Event(ctx, e) {
		return true
	}
	if ke, ok := e.(*event.KeyEvent); ok {
		if ke.KeyType == event.KeyPress && ke.Key == event.KeyEscape {
			o.Dismiss()
			return true
		}
	}
	if me, ok := e.(*event.MouseEvent); ok && me.MouseType == event.MousePress {
		if !o.content.Bounds().Contains(me.Position) {
			o.Dismiss()
			return true
		}
	}
	return true
}

// handleDrag implements the basic drag-to-dismiss gesture: press on the panel
// arms a drag; drag toward the edge offsets the panel live; release past the
// dismiss threshold closes, else snaps back. Returns true when it consumed the
// event (and thus a drag is in progress or just ended).
func (o *drawerOverlayWidget) handleDrag(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	switch me.MouseType {
	case event.MousePress:
		if me.Button == event.ButtonLeft && o.content.Bounds().Contains(me.Position) {
			o.dragging = true
			o.dragStart = me.Position
			o.dragOffset = 0
			return false // let content handle the press too (e.g. button arm)
		}
	case event.MouseDrag:
		if o.dragging {
			o.dragOffset = o.edgeDelta(me.Position)
			o.SetNeedsRedraw(true)
			ctx.Invalidate()
			return true
		}
	case event.MouseRelease:
		if o.dragging {
			o.dragging = false
			ext := o.dragExtent()
			if ext > 0 && o.dragOffset >= ext*metrics.DrawerDismissDragFraction {
				o.dragOffset = 0
				o.Dismiss()
			} else {
				o.dragOffset = 0
				o.SetNeedsRedraw(true)
				ctx.Invalidate()
			}
			return true
		}
	}
	return false
}

// edgeDelta returns the pointer drag distance toward the anchored edge from the
// press origin, clamped at zero (dragging away from the edge does not push the
// panel past its settled position).
func (o *drawerOverlayWidget) edgeDelta(pos geometry.Point) float32 {
	var d float32
	switch o.content.side {
	case DrawerTop:
		d = o.dragStart.Y - pos.Y
	case DrawerLeft:
		d = o.dragStart.X - pos.X
	case DrawerRight:
		d = pos.X - o.dragStart.X
	default: // DrawerBottom
		d = pos.Y - o.dragStart.Y
	}
	if d < 0 {
		d = 0
	}
	return d
}

// Dismiss invokes the dismissal callback.
func (o *drawerOverlayWidget) Dismiss() {
	if o.onDismiss != nil {
		o.onDismiss()
	}
}

// Modal reports true — the drawer blocks the tree below.
func (o *drawerOverlayWidget) Modal() bool { return true }

func (o *drawerOverlayWidget) Children() []widget.Widget {
	return []widget.Widget{o.content}
}

// DrawerContentWidget is the edge-anchored panel: bg Background, a single 1px
// Border on the inner edge, shadow-LG, rounded corners on the outward-facing
// edge (bottom drawer → rounded top), a centered grabber handle near the top
// (bottom drawer only), internal gap 16, padding 16. Unlike Sheet there is no
// close X button.
type DrawerContentWidget struct {
	widget.WidgetBase

	children []widget.Widget
	side     DrawerSide
	theme    *theme.Theme
	hideBar  bool
}

// DrawerContent assembles the panel body from header/title/description/footer
// sections (or any widgets). The panel anchors to the bottom edge unless Side
// overrides it; a bottom drawer shows the grabber handle by default.
func DrawerContent(children ...widget.Widget) *DrawerContentWidget {
	c := &DrawerContentWidget{children: children, side: DrawerBottom, theme: CurrentTheme()}
	c.SetVisible(true)
	c.SetEnabled(true)
	for _, ch := range children {
		c.AddChild(ch)
	}
	return c
}

// Side anchors the panel to a viewport edge (default bottom).
func (c *DrawerContentWidget) Side(s DrawerSide) *DrawerContentWidget {
	c.side = s
	return c
}

// HideHandle removes the grabber bar (bottom drawers show it by default).
func (c *DrawerContentWidget) HideHandle() *DrawerContentWidget {
	c.hideBar = true
	return c
}

// Theme pins a specific theme instead of the process-wide current theme.
func (c *DrawerContentWidget) Theme(th *theme.Theme) *DrawerContentWidget {
	c.theme = th
	return c
}

func (c *DrawerContentWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// showsHandle reports whether the grabber bar renders: bottom side, not hidden.
func (c *DrawerContentWidget) showsHandle() bool {
	return c.side == DrawerBottom && !c.hideBar
}

// horizontal reports whether the panel anchors to a left/right edge (spans
// full height, capped width) vs a top/bottom edge (spans full width, capped
// height).
func (c *DrawerContentWidget) horizontal() bool {
	return c.side == DrawerLeft || c.side == DrawerRight
}

// handleReserve is the vertical space the grabber consumes above the content
// column (top inset + bar height), or zero when no handle shows.
func (c *DrawerContentWidget) handleReserve() float32 {
	if !c.showsHandle() {
		return 0
	}
	return metrics.DrawerHandleTopInset + metrics.DrawerHandleHeight
}

// panelSize computes the panel size for a given viewport (left/right = full
// height × min(75% viewport, 384); top/bottom = full width × capped height).
func (c *DrawerContentWidget) panelSize(ctx widget.Context, viewport geometry.Size) geometry.Size {
	if c.horizontal() {
		w := viewport.Width * metrics.DrawerWidthFraction
		if w > metrics.DrawerMaxWidth {
			w = metrics.DrawerMaxWidth
		}
		return geometry.Sz(w, viewport.Height)
	}
	// top/bottom: full width, content-driven height capped at the fraction.
	h := c.contentHeight(ctx, viewport.Width)
	if maxH := viewport.Height * metrics.DrawerHeightFraction; h > maxH {
		h = maxH
	}
	return geometry.Sz(viewport.Width, h)
}

// contentHeight measures the stacked children at the given panel width plus the
// padding and any reserved grabber space.
func (c *DrawerContentWidget) contentHeight(ctx widget.Context, panelWidth float32) float32 {
	innerW := panelWidth - 2*metrics.DrawerPadding
	if innerW < 0 {
		innerW = 0
	}
	childCons := geometry.Loose(geometry.Sz(innerW, 100000))
	total := 2*metrics.DrawerPadding + c.handleReserve()
	for i, ch := range c.children {
		if i > 0 {
			total += metrics.DrawerGap
		}
		total += ch.Layout(ctx, childCons).Height
	}
	return total
}

// layoutAnchored positions the panel against its viewport edge and lays out
// the stacked children inside the padding.
func (c *DrawerContentWidget) layoutAnchored(ctx widget.Context, viewport geometry.Size) {
	size := c.panelSize(ctx, viewport)
	var origin geometry.Point
	switch c.side {
	case DrawerTop:
		origin = geometry.Pt(0, 0)
	case DrawerLeft:
		origin = geometry.Pt(0, 0)
	case DrawerRight:
		origin = geometry.Pt(viewport.Width-size.Width, 0)
	default: // DrawerBottom
		origin = geometry.Pt(0, viewport.Height-size.Height)
	}
	c.SetBounds(geometry.FromPointSize(origin, size))
	c.layoutChildren(ctx, origin, size)
}

// layoutChildren stacks children top-to-bottom inside the padding, below any
// reserved grabber space.
func (c *DrawerContentWidget) layoutChildren(ctx widget.Context, origin geometry.Point, size geometry.Size) {
	innerW := size.Width - 2*metrics.DrawerPadding
	if innerW < 0 {
		innerW = 0
	}
	childCons := geometry.Loose(geometry.Sz(innerW, 100000))
	x := origin.X + metrics.DrawerPadding
	y := origin.Y + metrics.DrawerPadding + c.handleReserve()
	cursorY := y
	for i, ch := range c.children {
		if i > 0 {
			cursorY += metrics.DrawerGap
		}
		// Position the child before layout so container children (header/footer
		// DrawerSectionWidget) position their own descendants relative to the
		// correct origin rather than a stale (0,0). See dialog.go for the same fix.
		setWidgetBounds(ch, geometry.FromPointSize(geometry.Pt(x, cursorY), geometry.Sz(innerW, 0)))
		sz := ch.Layout(ctx, childCons)
		setWidgetBounds(ch, geometry.FromPointSize(geometry.Pt(x, cursorY), geometry.Sz(innerW, sz.Height)))
		cursorY += sz.Height
	}
}

// Layout supports standalone rendering. For a bottom/top drawer it treats the
// available width as the panel width and measures content height; for a
// left/right drawer it uses the max width. Edge anchoring is skipped (the
// panel sits at its current position).
func (c *DrawerContentWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	var size geometry.Size
	if c.horizontal() {
		w := cons.MaxWidth
		if w <= 0 || w > metrics.DrawerMaxWidth {
			w = metrics.DrawerMaxWidth
		}
		h := cons.MaxHeight
		if h <= 0 {
			h = c.contentHeight(ctx, w)
		}
		size = geometry.Sz(w, h)
	} else {
		w := cons.MaxWidth
		if w <= 0 {
			w = metrics.DrawerMaxWidth
		}
		size = geometry.Sz(w, c.contentHeight(ctx, w))
	}
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	c.layoutChildren(ctx, c.Position(), size)
	return size
}

// outerRadius returns the panel corner radius (rounded-t-lg / rounded-b-lg ⇒
// theme RadiusLG). Left/right drawers have no radius.
func (c *DrawerContentWidget) outerRadius(th *theme.Theme) float32 {
	if c.horizontal() {
		return 0
	}
	return th.RadiusLG()
}

// roundedCorners returns the corner mask that stays rounded for the side: the
// edge facing the viewport interior (bottom drawer → top corners).
func (c *DrawerContentWidget) roundedCorners() draw.Corners {
	switch c.side {
	case DrawerTop: // rounded-b: keep the bottom corners.
		return draw.BottomLeft | draw.BottomRight
	case DrawerBottom: // rounded-t: keep the top corners.
		return draw.TopLeft | draw.TopRight
	default:
		return 0
	}
}

// Draw paints shadow, the panel surface (with the outward-edge corner radius),
// the inner-edge border, the grabber handle, and the children.
func (c *DrawerContentWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.resolvedTheme()
	tok := th.Active()
	b := c.Bounds()
	radius := c.outerRadius(th)

	draw.Shadow(canvas, b, radius, metrics.ShadowLG)
	if radius > 0 {
		// Round only the outward-facing corners; square the edge against the
		// viewport (those corners sit off-screen but squaring keeps the fill flush).
		square := draw.AllCorners &^ c.roundedCorners()
		draw.SquareCorners(canvas, b, radius, tok.Background, square)
	} else {
		canvas.DrawRect(b, tok.Background)
	}
	c.drawInnerBorder(canvas, b, tok.Border)

	if c.showsHandle() {
		c.drawHandle(canvas, b, tok)
	}

	for _, ch := range c.children {
		ch.Draw(ctx, canvas)
	}
}

// drawInnerBorder strokes the single 1px border on the edge facing the
// viewport interior (border-t for bottom, border-b for top, border-l for
// right, border-r for left). The border is center-stroked at the panel edge.
func (c *DrawerContentWidget) drawInnerBorder(canvas widget.Canvas, b geometry.Rect, col widget.Color) {
	w := metrics.DrawerBorderWidth
	switch c.side {
	case DrawerTop: // border-b: line on the panel's bottom edge.
		canvas.DrawRect(geometry.NewRect(b.Min.X, b.Max.Y-w, b.Width(), w), col)
	case DrawerLeft: // border-r: line on the panel's right edge.
		canvas.DrawRect(geometry.NewRect(b.Max.X-w, b.Min.Y, w, b.Height()), col)
	case DrawerRight: // border-l: line on the panel's left edge.
		canvas.DrawRect(geometry.NewRect(b.Min.X, b.Min.Y, w, b.Height()), col)
	default: // DrawerBottom — border-t: line on the panel's top edge.
		canvas.DrawRect(geometry.NewRect(b.Min.X, b.Min.Y, b.Width(), w), col)
	}
}

// handleBounds returns the grabber bar rect: centered horizontally, mt-4 below
// the top edge, 100×8px.
func (c *DrawerContentWidget) handleBounds(b geometry.Rect) geometry.Rect {
	x := b.Min.X + (b.Width()-metrics.DrawerHandleWidth)/2
	y := b.Min.Y + metrics.DrawerHandleTopInset
	return geometry.NewRect(x, y, metrics.DrawerHandleWidth, metrics.DrawerHandleHeight)
}

// drawHandle paints the muted, fully-rounded grabber pill.
func (c *DrawerContentWidget) drawHandle(canvas widget.Canvas, b geometry.Rect, tok *theme.Tokens) {
	hb := c.handleBounds(b)
	canvas.DrawRoundRect(hb, tok.Muted, metrics.DrawerHandleHeight/2)
}

// Event forwards to the children (footer buttons, etc.); the overlay handles
// dismissal (Esc, backdrop, drag).
func (c *DrawerContentWidget) Event(ctx widget.Context, e event.Event) bool {
	for _, ch := range c.children {
		if ch.Event(ctx, e) {
			return true
		}
	}
	return false
}

// Children returns the section widgets.
func (c *DrawerContentWidget) Children() []widget.Widget { return c.children }

// DrawerSectionWidget lays out the drawer header (a left-aligned column with
// gap 6) and footer (a column with gap 8). It owns its layout because
// primitives.Box has no main-axis justify.
type DrawerSectionWidget struct {
	widget.WidgetBase
	children []widget.Widget
	gap      float32
}

// DrawerHeader stacks title + description vertically with gap 6 (gap-1.5).
func DrawerHeader(children ...widget.Widget) *DrawerSectionWidget {
	return newDrawerSection(children, metrics.DrawerHeaderGap)
}

// DrawerFooter stacks footer content vertically with gap 8 (gap-2). shadcn
// pushes it to the bottom with mt-auto; in graft the content column already
// orders the footer last, so it renders below the body.
func DrawerFooter(children ...widget.Widget) *DrawerSectionWidget {
	return newDrawerSection(children, metrics.DrawerFooterGap)
}

func newDrawerSection(children []widget.Widget, gap float32) *DrawerSectionWidget {
	s := &DrawerSectionWidget{children: children, gap: gap}
	s.SetVisible(true)
	s.SetEnabled(true)
	for _, ch := range children {
		s.AddChild(ch)
	}
	return s
}

func (s *DrawerSectionWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	width := c.MaxWidth
	childCons := geometry.Loose(geometry.Sz(width, 100000))
	x := s.Position().X
	y := s.Position().Y
	cursorY := y
	var maxW float32
	for i, ch := range s.children {
		if i > 0 {
			cursorY += s.gap
		}
		sz := ch.Layout(ctx, childCons)
		setWidgetBounds(ch, geometry.FromPointSize(geometry.Pt(x, cursorY), geometry.Sz(width, sz.Height)))
		cursorY += sz.Height
		if sz.Width > maxW {
			maxW = sz.Width
		}
	}
	if width <= 0 {
		width = maxW
	}
	size := geometry.Sz(width, cursorY-y)
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

func (s *DrawerSectionWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	for _, ch := range s.children {
		ch.Draw(ctx, canvas)
	}
}

func (s *DrawerSectionWidget) Event(ctx widget.Context, e event.Event) bool {
	for _, ch := range s.children {
		if ch.Event(ctx, e) {
			return true
		}
	}
	return false
}

func (s *DrawerSectionWidget) Children() []widget.Widget { return s.children }

// DrawerTitle renders the drawer title: 18px / 600 / leading-none, foreground.
func DrawerTitle(text string) *TypographyWidget {
	return styled(text, metrics.DrawerTitleFontSize, metrics.DrawerTitleWeight, metrics.DrawerTitleLineHeight)
}

// DrawerDescription renders the drawer description: 14px muted-foreground.
func DrawerDescription(text string) *TypographyWidget {
	return styled(text, metrics.DrawerDescriptionFontSize, 400, metrics.DrawerDescriptionLineHeight).Muted()
}

// DrawerPreview renders the drawer exactly as it appears at runtime: a modal
// frame of windowSize with the black@50% backdrop and the panel anchored to
// its edge (the layoutAnchored path, not the standalone natural-size Layout),
// in the SETTLED open state (slide offset zero). Use it for goldens, docs, and
// screenshots.
func DrawerPreview(content *DrawerContentWidget, windowSize geometry.Size) Widget {
	p := &drawerPreviewWidget{content: content, windowSize: windowSize}
	p.SetVisible(true)
	p.SetEnabled(true)
	p.AddChild(content)
	return p
}

type drawerPreviewWidget struct {
	widget.WidgetBase
	content    *DrawerContentWidget
	windowSize geometry.Size
}

func (p *drawerPreviewWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := c.Constrain(p.windowSize)
	p.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	p.content.layoutAnchored(ctx, size)
	return size
}

func (p *drawerPreviewWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	canvas.DrawRect(p.Bounds(), widget.RGBA(0, 0, 0, metrics.OverlayAlpha))
	p.content.Draw(ctx, canvas)
}

func (p *drawerPreviewWidget) Event(widget.Context, event.Event) bool { return false }

func (p *drawerPreviewWidget) Children() []widget.Widget { return []widget.Widget{p.content} }

// Compile-time interface checks.
var (
	_ widget.Widget    = (*DrawerWidget)(nil)
	_ widget.Lifecycle = (*DrawerWidget)(nil)
	_ widget.Widget    = (*DrawerContentWidget)(nil)
	_ widget.Widget    = (*DrawerSectionWidget)(nil)
	_ widget.Widget    = (*drawerTriggerWidget)(nil)
	_ widget.Widget    = (*drawerOverlayWidget)(nil)
	_ widget.Widget    = (*drawerPreviewWidget)(nil)
)

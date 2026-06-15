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

// SheetSide selects the viewport edge the sheet panel anchors to.
type SheetSide uint8

// Sheet sides (default Right, matching shadcn's side="right" default).
const (
	// SheetRight anchors the panel to the right edge (default).
	SheetRight SheetSide = iota
	// SheetLeft anchors the panel to the left edge.
	SheetLeft
	// SheetTop anchors the panel to the top edge.
	SheetTop
	// SheetBottom anchors the panel to the bottom edge.
	SheetBottom
)

// SheetWidget is the zero-size host that watches an open signal and pushes a
// modal, edge-anchored overlay carrying the SheetContent
// (docs/research/03-shadcn-pixel-spec.md §Sheet).
//
// Architecture decision: graft-owned widget, modeled on the shipped Dialog
// host. It reuses Dialog's modal chassis pattern (full-window black@50%
// backdrop, Esc + backdrop dismissal) but anchors the panel to a viewport
// edge instead of centering it, and the panel spans the full cross-axis with
// a single 1px border on its inner edge. On open the panel slides in from its
// edge over SheetOpenDurationMillis (the shadcn slide-in), driven by an
// animation.Controller ticked in the overlay's Draw (the switch.go pattern);
// the backdrop fades in with it. Close currently snaps (immediate overlay
// removal) — a slide-out would need deferred removal, which conflicts with the
// synchronous overlay lifecycle the host tests rely on. Goldens render the
// SETTLED open state (SheetPreview, no overlay), so the slide offset is zero
// there.
type SheetWidget struct {
	widget.WidgetBase

	content *SheetContentWidget
	open    state.Signal[bool]
	initial bool
	onOpen  func(bool)

	ctx     widget.Context
	overlay *sheetOverlayWidget
	shown   bool
}

// Sheet creates a sheet host for the given content. Bind an open signal with
// Bind, or set an initial state with Open; pair with a SheetTrigger.
func Sheet(content *SheetContentWidget) *SheetWidget {
	s := &SheetWidget{content: content}
	s.SetVisible(true)
	s.SetEnabled(true)
	return s
}

// Bind controls the open state from a signal (read for render, written on
// Esc/backdrop/close-button).
func (s *SheetWidget) Bind(open state.Signal[bool]) *SheetWidget {
	s.open = open
	return s
}

// Open sets the uncontrolled initial open state (shadcn defaultOpen).
func (s *SheetWidget) Open(v bool) *SheetWidget {
	s.initial = v
	return s
}

// OnOpenChange registers an observer invoked whenever the open state changes.
func (s *SheetWidget) OnOpenChange(fn func(bool)) *SheetWidget {
	s.onOpen = fn
	return s
}

func (s *SheetWidget) isOpen() bool {
	if s.open != nil {
		return s.open.Get()
	}
	return s.initial
}

func (s *SheetWidget) setOpen(v bool) {
	if s.open != nil {
		s.open.Set(v)
	} else {
		s.initial = v
	}
	if s.onOpen != nil {
		s.onOpen(v)
	}
	s.sync()
}

// Mount binds the open signal so external Set calls re-sync the overlay.
func (s *SheetWidget) Mount(ctx widget.Context) {
	s.ctx = ctx
	if s.open != nil {
		if sched := ctx.Scheduler(); sched != nil {
			s.AddBinding(state.BindToScheduler[bool](s.open, s, sched))
		}
		state.SubscribeForever[bool](s.open, func(bool) { s.sync() })
	}
	s.sync()
}

// Unmount removes any live overlay; bindings are cleaned by WidgetBase.
func (s *SheetWidget) Unmount() {
	if s.shown {
		s.pop()
	}
}

// Layout reports zero size; the host is invisible chrome.
func (s *SheetWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	s.ctx = ctx
	s.SetBounds(geometry.FromPointSize(s.Position(), geometry.Sz(0, 0)))
	s.sync()
	return c.Constrain(geometry.Sz(0, 0))
}

// Draw paints nothing; the content lives in the overlay.
func (s *SheetWidget) Draw(ctx widget.Context, _ widget.Canvas) { s.ctx = ctx }

// Event ignores input; the overlay handles its own.
func (s *SheetWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; the host is a leaf (content lives in the overlay).
func (s *SheetWidget) Children() []widget.Widget { return nil }

func (s *SheetWidget) sync() {
	if s.ctx == nil {
		return
	}
	want := s.isOpen()
	if want && !s.shown {
		s.push()
	} else if !want && s.shown {
		s.pop()
	}
}

func (s *SheetWidget) push() {
	om := s.ctx.OverlayManager()
	if om == nil {
		return
	}
	s.content.onClose = func() { s.setOpen(false) }
	s.overlay = newSheetOverlay(s.content, s.ctx.WindowSize(), func() { s.setOpen(false) })
	om.PushOverlay(s.overlay, func() { s.shown = false; s.overlay = nil })
	s.shown = true
	s.ctx.Invalidate()
}

func (s *SheetWidget) pop() {
	if om := s.ctx.OverlayManager(); om != nil && s.overlay != nil {
		om.RemoveOverlay(s.overlay)
	}
	s.overlay = nil
	s.shown = false
	s.ctx.Invalidate()
}

// SheetTrigger wraps a trigger widget so a click opens the sheet by setting
// open to true (mirrors DialogTrigger).
func SheetTrigger(trigger widget.Widget, open state.Signal[bool]) widget.Widget {
	return &sheetTriggerWidget{trigger: trigger, open: open}
}

type sheetTriggerWidget struct {
	widget.WidgetBase
	trigger widget.Widget
	open    state.Signal[bool]
}

func (t *sheetTriggerWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := t.trigger.Layout(ctx, c)
	setWidgetBounds(t.trigger, geometry.FromPointSize(t.Position(), size))
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

func (t *sheetTriggerWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	setWidgetBounds(t.trigger, t.Bounds())
	t.trigger.Draw(ctx, canvas)
}

func (t *sheetTriggerWidget) Event(ctx widget.Context, e event.Event) bool {
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

func (t *sheetTriggerWidget) Children() []widget.Widget {
	return []widget.Widget{t.trigger}
}

// sheetOverlayWidget is the modal chassis: full-window black@50% backdrop with
// the content panel anchored to its side. Esc and backdrop click dismiss.
type sheetOverlayWidget struct {
	widget.WidgetBase

	content    *SheetContentWidget
	windowSize geometry.Size
	onDismiss  func()

	// progress is the slide-in position: 0 = panel fully off-screen at its edge,
	// 1 = settled open. It animates 0→1 on open (the shadcn slide-in) via an
	// animation.Controller ticked in Draw (the switch.go pattern). The backdrop
	// alpha tracks progress so it fades in with the panel.
	progress    float32
	animCtrl    *animation.Controller
	animAdpt    sheetProgress
	animStarted bool
}

func newSheetOverlay(content *SheetContentWidget, windowSize geometry.Size, onDismiss func()) *sheetOverlayWidget {
	o := &sheetOverlayWidget{content: content, windowSize: windowSize, onDismiss: onDismiss,
		animCtrl: animation.NewController()}
	o.animAdpt.o = o
	o.SetVisible(true)
	o.SetEnabled(true)
	o.AddChild(content)
	return o
}

func (o *sheetOverlayWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := c.Constrain(o.windowSize)
	o.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	o.content.layoutAnchored(ctx, size)
	return size
}

func (o *sheetOverlayWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	// Kick off the slide-in on the first paint (ctx is available here), then
	// advance it. When idle, snap fully open (progress 1) so a settled sheet is
	// always at its edge regardless of render mode.
	if !o.animStarted {
		o.animStarted = true
		o.startOpen(ctx)
	}
	o.tickAnim(ctx)

	// Backdrop fades in with the panel.
	canvas.DrawRect(o.Bounds(), widget.RGBA(0, 0, 0, metrics.OverlayAlpha*o.progress))

	// Panel slides from its edge by the remaining (1-progress) of its extent.
	off := o.slideOffset()
	canvas.PushTransform(off)
	o.content.Draw(ctx, canvas)
	canvas.PopTransform()
}

// slideOffset returns the panel's current draw translation: zero when settled
// (progress 1), a full panel extent toward the edge when closed (progress 0).
func (o *sheetOverlayWidget) slideOffset() geometry.Point {
	d := 1 - o.progress
	if d <= 0 {
		return geometry.Pt(0, 0)
	}
	b := o.content.Bounds()
	switch o.content.side {
	case SheetLeft:
		return geometry.Pt(-b.Width()*d, 0)
	case SheetTop:
		return geometry.Pt(0, -b.Height()*d)
	case SheetBottom:
		return geometry.Pt(0, b.Height()*d)
	default: // SheetRight
		return geometry.Pt(b.Width()*d, 0)
	}
}

// startOpen tweens progress 0→1 over SheetOpenDurationMillis with the shadcn
// ease, driven by the controller ticked in Draw.
func (o *sheetOverlayWidget) startOpen(ctx widget.Context) {
	if o.animCtrl == nil {
		o.progress = 1
		return
	}
	o.animAdpt.ctx = ctx
	animation.To(&o.animAdpt, 1).
		From(0).
		Duration(time.Duration(metrics.SheetOpenDurationMillis) * time.Millisecond).
		Ease(animation.CubicBezier(0.4, 0, 0.2, 1)).
		Start(o.animCtrl)
}

// tickAnim advances the slide-in while active; snaps to settled-open when idle.
func (o *sheetOverlayWidget) tickAnim(ctx widget.Context) {
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

// sheetProgress adapts the overlay's progress field to animation.To's
// signalFloat32 interface, marking the overlay dirty each frame.
type sheetProgress struct {
	o   *sheetOverlayWidget
	ctx widget.Context
}

func (a *sheetProgress) Get() float32 { return a.o.progress }

func (a *sheetProgress) Set(v float32) {
	a.o.progress = v
	a.o.SetNeedsRedraw(true)
	if a.ctx != nil {
		a.ctx.Invalidate()
	}
}

func (o *sheetOverlayWidget) Event(ctx widget.Context, e event.Event) bool {
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

// Dismiss invokes the dismissal callback.
func (o *sheetOverlayWidget) Dismiss() {
	if o.onDismiss != nil {
		o.onDismiss()
	}
}

// Modal reports true — the sheet blocks the tree below.
func (o *sheetOverlayWidget) Modal() bool { return true }

func (o *sheetOverlayWidget) Children() []widget.Widget {
	return []widget.Widget{o.content}
}

// SheetContentWidget is the edge-anchored panel: bg Background, a single 1px
// Border on the inner edge, shadow-LG, internal gap 16, padding 16, with a
// close X button in the top-right corner (unless hidden).
type SheetContentWidget struct {
	widget.WidgetBase

	children []widget.Widget
	side     SheetSide
	theme    *theme.Theme
	hideX    bool

	onClose func()

	closeHovered bool
	closeFocused bool
}

// SheetContent assembles the panel body from header/title/description/footer
// sections (or any widgets). The close button is shown by default; the panel
// anchors to the right edge unless Side overrides it.
func SheetContent(children ...widget.Widget) *SheetContentWidget {
	c := &SheetContentWidget{children: children, side: SheetRight, theme: CurrentTheme()}
	c.SetVisible(true)
	c.SetEnabled(true)
	for _, ch := range children {
		c.AddChild(ch)
	}
	return c
}

// Side anchors the panel to a viewport edge (default right).
func (c *SheetContentWidget) Side(s SheetSide) *SheetContentWidget {
	c.side = s
	return c
}

// HideClose removes the top-right close button.
func (c *SheetContentWidget) HideClose() *SheetContentWidget {
	c.hideX = true
	return c
}

// OnClose sets the callback invoked when the close button is pressed. The
// Sheet host wires this automatically.
func (c *SheetContentWidget) OnClose(fn func()) *SheetContentWidget {
	c.onClose = fn
	return c
}

// Theme pins a specific theme instead of the process-wide current theme.
func (c *SheetContentWidget) Theme(th *theme.Theme) *SheetContentWidget {
	c.theme = th
	return c
}

func (c *SheetContentWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// horizontal reports whether the panel anchors to a left/right edge (spans
// full height, capped width) vs a top/bottom edge (spans full width, capped
// height).
func (c *SheetContentWidget) horizontal() bool {
	return c.side == SheetLeft || c.side == SheetRight
}

// panelSize computes the panel size for a given viewport (left/right = full
// height × min(75% viewport, 384); top/bottom = full width × capped height).
func (c *SheetContentWidget) panelSize(ctx widget.Context, viewport geometry.Size) geometry.Size {
	if c.horizontal() {
		w := viewport.Width * metrics.SheetWidthFraction
		if w > metrics.SheetMaxWidth {
			w = metrics.SheetMaxWidth
		}
		return geometry.Sz(w, viewport.Height)
	}
	// top/bottom: full width, content-driven height capped at the fraction.
	h := c.contentHeight(ctx, viewport.Width)
	if maxH := viewport.Height * metrics.SheetHeightFraction; h > maxH {
		h = maxH
	}
	return geometry.Sz(viewport.Width, h)
}

// contentHeight measures the stacked children at the given panel width.
func (c *SheetContentWidget) contentHeight(ctx widget.Context, panelWidth float32) float32 {
	innerW := panelWidth - 2*metrics.SheetPadding
	if innerW < 0 {
		innerW = 0
	}
	childCons := geometry.Loose(geometry.Sz(innerW, 100000))
	total := 2 * metrics.SheetPadding
	for i, ch := range c.children {
		if i > 0 {
			total += metrics.SheetGap
		}
		total += ch.Layout(ctx, childCons).Height
	}
	return total
}

// layoutAnchored positions the panel against its viewport edge and lays out
// the stacked children inside the padding.
func (c *SheetContentWidget) layoutAnchored(ctx widget.Context, viewport geometry.Size) {
	size := c.panelSize(ctx, viewport)
	var origin geometry.Point
	switch c.side {
	case SheetRight:
		origin = geometry.Pt(viewport.Width-size.Width, 0)
	case SheetLeft:
		origin = geometry.Pt(0, 0)
	case SheetTop:
		origin = geometry.Pt(0, 0)
	case SheetBottom:
		origin = geometry.Pt(0, viewport.Height-size.Height)
	}
	c.SetBounds(geometry.FromPointSize(origin, size))
	c.layoutChildren(ctx, origin, size)
}

// layoutChildren stacks children top-to-bottom inside the padding.
func (c *SheetContentWidget) layoutChildren(ctx widget.Context, origin geometry.Point, size geometry.Size) {
	innerW := size.Width - 2*metrics.SheetPadding
	if innerW < 0 {
		innerW = 0
	}
	childCons := geometry.Loose(geometry.Sz(innerW, 100000))
	x := origin.X + metrics.SheetPadding
	y := origin.Y + metrics.SheetPadding
	cursorY := y
	for i, ch := range c.children {
		if i > 0 {
			cursorY += metrics.SheetGap
		}
		// Position the child before layout so container children (header/footer
		// SheetSectionWidget) position their own descendants relative to the
		// correct origin rather than a stale (0,0). See dialog.go for the same fix.
		setWidgetBounds(ch, geometry.FromPointSize(geometry.Pt(x, cursorY), geometry.Sz(innerW, 0)))
		sz := ch.Layout(ctx, childCons)
		setWidgetBounds(ch, geometry.FromPointSize(geometry.Pt(x, cursorY), geometry.Sz(innerW, sz.Height)))
		cursorY += sz.Height
	}
}

// Layout supports standalone rendering (goldens). It treats the available
// width/height as the panel size directly (no edge anchoring), so a golden can
// render the panel at a natural size.
func (c *SheetContentWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	w := cons.MaxWidth
	if w <= 0 || w > metrics.SheetMaxWidth {
		w = metrics.SheetMaxWidth
	}
	h := c.contentHeight(ctx, w)
	size := geometry.Sz(w, h)
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	c.layoutChildren(ctx, c.Position(), size)
	return size
}

// Draw paints shadow, the panel surface, the inner-edge border, the children,
// and the close button.
func (c *SheetContentWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.resolvedTheme()
	tok := th.Active()
	b := c.Bounds()

	// Edge-anchored panels have no corner radius; draw a square surface with a
	// soft shadow underneath.
	draw.Shadow(canvas, b, 0, metrics.ShadowLG)
	canvas.DrawRect(b, tok.Background)
	c.drawInnerBorder(canvas, b, tok.Border)

	for _, ch := range c.children {
		ch.Draw(ctx, canvas)
	}

	if !c.hideX {
		c.drawClose(canvas, b, th, tok)
	}
}

// drawInnerBorder strokes the single 1px border on the edge facing the
// viewport interior (border-l for right, border-r for left, border-b for top,
// border-t for bottom). The border is center-stroked at the panel edge.
func (c *SheetContentWidget) drawInnerBorder(canvas widget.Canvas, b geometry.Rect, col widget.Color) {
	w := metrics.SheetBorderWidth
	switch c.side {
	case SheetRight: // border-l: line on the panel's left edge.
		canvas.DrawRect(geometry.NewRect(b.Min.X, b.Min.Y, w, b.Height()), col)
	case SheetLeft: // border-r: line on the panel's right edge.
		canvas.DrawRect(geometry.NewRect(b.Max.X-w, b.Min.Y, w, b.Height()), col)
	case SheetTop: // border-b: line on the panel's bottom edge.
		canvas.DrawRect(geometry.NewRect(b.Min.X, b.Max.Y-w, b.Width(), w), col)
	case SheetBottom: // border-t: line on the panel's top edge.
		canvas.DrawRect(geometry.NewRect(b.Min.X, b.Min.Y, b.Width(), w), col)
	}
}

// closeBounds returns the close button's hit/hover box (centered on the X icon
// at the top-right inset).
func (c *SheetContentWidget) closeBounds(b geometry.Rect) geometry.Rect {
	iconX := b.Max.X - metrics.SheetCloseInset - metrics.SheetCloseIconSize
	iconY := b.Min.Y + metrics.SheetCloseInset
	pad := metrics.SheetCloseHitPad
	return geometry.NewRect(
		iconX-pad, iconY-pad,
		metrics.SheetCloseIconSize+2*pad, metrics.SheetCloseIconSize+2*pad,
	)
}

func (c *SheetContentWidget) drawClose(canvas widget.Canvas, b geometry.Rect, th *theme.Theme, tok *theme.Tokens) {
	hit := c.closeBounds(b)

	if c.closeFocused {
		draw.OffsetRing(canvas, hit, th.RadiusXS(), tok.Ring,
			metrics.LegacyCloseRingWidth, metrics.LegacyCloseRingOffset, tok.Background)
	}

	opacity := metrics.SheetCloseIdleOpacity
	if c.closeHovered {
		opacity = 1
	}
	iconColor := draw.MulAlpha(tok.Foreground, opacity)

	iconRect := geometry.NewRect(
		hit.Min.X+metrics.SheetCloseHitPad, hit.Min.Y+metrics.SheetCloseHitPad,
		metrics.SheetCloseIconSize, metrics.SheetCloseIconSize,
	)
	icon.Draw(canvas, icons.X, iconRect, iconColor)
}

// Event handles close-button hover/click; everything else falls through to the
// children (footer buttons, etc.).
func (c *SheetContentWidget) Event(ctx widget.Context, e event.Event) bool {
	if !c.hideX {
		if me, ok := e.(*event.MouseEvent); ok {
			hit := c.closeBounds(c.Bounds())
			inside := hit.Contains(me.Position)
			switch me.MouseType {
			case event.MouseMove, event.MouseEnter:
				if inside != c.closeHovered {
					c.closeHovered = inside
					if inside {
						ctx.SetCursor(widget.CursorPointer)
					}
					c.SetNeedsRedraw(true)
					ctx.Invalidate()
				}
			case event.MouseLeave:
				if c.closeHovered {
					c.closeHovered = false
					c.SetNeedsRedraw(true)
				}
			case event.MousePress:
				if inside && me.Button == event.ButtonLeft {
					if c.onClose != nil {
						c.onClose()
					}
					return true
				}
			}
		}
	}

	for _, ch := range c.children {
		if ch.Event(ctx, e) {
			return true
		}
	}
	return false
}

// Children returns the section widgets.
func (c *SheetContentWidget) Children() []widget.Widget { return c.children }

// SheetSectionWidget lays out the sheet header (a left-aligned column with gap
// 6) and footer (a column with gap 8 pushed to the bottom). It owns its layout
// because primitives.Box has no main-axis justify.
type SheetSectionWidget struct {
	widget.WidgetBase
	children []widget.Widget
	gap      float32
}

// SheetHeader stacks title + description vertically with gap 6 (gap-1.5).
func SheetHeader(children ...widget.Widget) *SheetSectionWidget {
	return newSheetSection(children, metrics.SheetHeaderGap)
}

// SheetFooter stacks footer content vertically with gap 8 (gap-2). shadcn
// pushes it to the bottom with mt-auto; in graft the content column already
// orders the footer last, so it renders below the body.
func SheetFooter(children ...widget.Widget) *SheetSectionWidget {
	return newSheetSection(children, metrics.SheetFooterGap)
}

func newSheetSection(children []widget.Widget, gap float32) *SheetSectionWidget {
	s := &SheetSectionWidget{children: children, gap: gap}
	s.SetVisible(true)
	s.SetEnabled(true)
	for _, ch := range children {
		s.AddChild(ch)
	}
	return s
}

func (s *SheetSectionWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
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

func (s *SheetSectionWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	for _, ch := range s.children {
		ch.Draw(ctx, canvas)
	}
}

func (s *SheetSectionWidget) Event(ctx widget.Context, e event.Event) bool {
	for _, ch := range s.children {
		if ch.Event(ctx, e) {
			return true
		}
	}
	return false
}

func (s *SheetSectionWidget) Children() []widget.Widget { return s.children }

// SheetTitle renders the sheet title: 18px / 600 / leading-none, foreground.
func SheetTitle(text string) *TypographyWidget {
	return styled(text, metrics.SheetTitleFontSize, metrics.SheetTitleWeight, metrics.SheetTitleLineHeight)
}

// SheetDescription renders the sheet description: 14px muted-foreground.
func SheetDescription(text string) *TypographyWidget {
	return styled(text, metrics.SheetDescriptionFontSize, 400, metrics.SheetDescriptionLineHeight).Muted()
}

// SheetPreview renders the sheet exactly as it appears at runtime: a modal
// frame of windowSize with the black@50% backdrop and the panel anchored
// full-cross-axis to its edge (the layoutAnchored path, not the standalone
// natural-size Layout). Use it for goldens, docs, and screenshots so the
// edge-anchored, full-height appearance is captured rather than a detached
// content box.
func SheetPreview(content *SheetContentWidget, windowSize geometry.Size) Widget {
	p := &sheetPreviewWidget{content: content, windowSize: windowSize}
	p.SetVisible(true)
	p.SetEnabled(true)
	p.AddChild(content)
	return p
}

type sheetPreviewWidget struct {
	widget.WidgetBase
	content    *SheetContentWidget
	windowSize geometry.Size
}

func (p *sheetPreviewWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := c.Constrain(p.windowSize)
	p.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	p.content.layoutAnchored(ctx, size)
	return size
}

func (p *sheetPreviewWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	canvas.DrawRect(p.Bounds(), widget.RGBA(0, 0, 0, metrics.OverlayAlpha))
	p.content.Draw(ctx, canvas)
}

func (p *sheetPreviewWidget) Event(widget.Context, event.Event) bool { return false }

func (p *sheetPreviewWidget) Children() []widget.Widget { return []widget.Widget{p.content} }

// Compile-time interface checks.
var (
	_ widget.Widget    = (*SheetWidget)(nil)
	_ widget.Lifecycle = (*SheetWidget)(nil)
	_ widget.Widget    = (*SheetContentWidget)(nil)
	_ widget.Widget    = (*SheetSectionWidget)(nil)
	_ widget.Widget    = (*sheetTriggerWidget)(nil)
	_ widget.Widget    = (*sheetOverlayWidget)(nil)
	_ widget.Widget    = (*sheetPreviewWidget)(nil)
)

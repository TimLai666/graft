package graft

import (
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

// DialogWidget is the zero-size host that watches an open signal and pushes a
// modal overlay carrying the DialogContent (DESIGN.md §4).
//
// Place it anywhere in the tree; pair it with a DialogTrigger that flips the
// same signal. On open it pushes a full-window modal overlay (pure black @50%
// backdrop in both modes), centers the content, and wires Esc + backdrop click
// to close. AlertDialog reuses this host with backdrop dismissal disabled and
// the close button hidden.
type DialogWidget struct {
	widget.WidgetBase

	content *DialogContentWidget
	open    state.Signal[bool]
	initial bool // uncontrolled initial open state
	onOpen  func(bool)

	// modal dismissal policy: Dialog allows backdrop click, AlertDialog does
	// not (Esc always closes).
	dismissOnBackdrop bool

	ctx     widget.Context // captured at layout/mount for signal-driven pushes
	overlay *dialogOverlayWidget
	shown   bool
}

// Dialog creates a dialog host for the given content. Bind an open signal with
// Bind, or set an initial state with Open; pair with a DialogTrigger.
func Dialog(content *DialogContentWidget) *DialogWidget {
	d := &DialogWidget{
		content:           content,
		dismissOnBackdrop: true,
	}
	d.SetVisible(true)
	d.SetEnabled(true)
	return d
}

// Bind controls the open state from a signal (read for render, written on
// Esc/backdrop/close-button).
func (d *DialogWidget) Bind(open state.Signal[bool]) *DialogWidget {
	d.open = open
	return d
}

// Open sets the uncontrolled initial open state (shadcn defaultOpen).
func (d *DialogWidget) Open(v bool) *DialogWidget {
	d.initial = v
	return d
}

// OnOpenChange registers an observer invoked whenever the open state changes.
func (d *DialogWidget) OnOpenChange(fn func(bool)) *DialogWidget {
	d.onOpen = fn
	return d
}

// isOpen reads the current open state (signal wins over initial).
func (d *DialogWidget) isOpen() bool {
	if d.open != nil {
		return d.open.Get()
	}
	return d.initial
}

// setOpen writes the open state through the signal (if bound) and notifies.
func (d *DialogWidget) setOpen(v bool) {
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
func (d *DialogWidget) Mount(ctx widget.Context) {
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
func (d *DialogWidget) Unmount() {
	if d.shown {
		d.pop()
	}
}

// Layout reports zero size; the host is invisible chrome. It captures ctx so
// signal callbacks can reach the overlay manager.
func (d *DialogWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	d.ctx = ctx
	d.SetBounds(geometry.FromPointSize(d.Position(), geometry.Sz(0, 0)))
	d.sync()
	return c.Constrain(geometry.Sz(0, 0))
}

// Draw paints nothing; the content lives in the overlay.
func (d *DialogWidget) Draw(ctx widget.Context, _ widget.Canvas) { d.ctx = ctx }

// Event ignores input; the overlay handles its own.
func (d *DialogWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; the host is a leaf (content is hosted in the overlay).
func (d *DialogWidget) Children() []widget.Widget { return nil }

// sync reconciles the overlay with the current open state.
func (d *DialogWidget) sync() {
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

func (d *DialogWidget) push() {
	om := d.ctx.OverlayManager()
	if om == nil {
		return
	}
	d.content.onClose = func() { d.setOpen(false) }
	d.overlay = newDialogOverlay(d.content, d.ctx.WindowSize(), d.dismissOnBackdrop, func() {
		d.setOpen(false)
	})
	om.PushOverlay(d.overlay, func() { d.shown = false; d.overlay = nil })
	d.shown = true
	d.ctx.Invalidate()
}

func (d *DialogWidget) pop() {
	if om := d.ctx.OverlayManager(); om != nil && d.overlay != nil {
		om.RemoveOverlay(d.overlay)
	}
	d.overlay = nil
	d.shown = false
	d.ctx.Invalidate()
}

// DialogTrigger wraps a trigger widget so a click opens the dialog by setting
// open to true. The trigger keeps its own appearance (commonly a Button).
func DialogTrigger(trigger widget.Widget, open state.Signal[bool]) widget.Widget {
	return &dialogTriggerWidget{trigger: trigger, open: open}
}

type dialogTriggerWidget struct {
	widget.WidgetBase
	trigger widget.Widget
	open    state.Signal[bool]
}

func (t *dialogTriggerWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := t.trigger.Layout(ctx, c)
	setWidgetBounds(t.trigger, geometry.FromPointSize(t.Position(), size))
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

func (t *dialogTriggerWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	setWidgetBounds(t.trigger, t.Bounds())
	t.trigger.Draw(ctx, canvas)
}

func (t *dialogTriggerWidget) Event(ctx widget.Context, e event.Event) bool {
	// Let the wrapped trigger react first (hover, focus ring, its own OnClick).
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

func (t *dialogTriggerWidget) Children() []widget.Widget {
	return []widget.Widget{t.trigger}
}

// dialogOverlayWidget is the modal chassis: full-window black@50% backdrop with
// the content card centered. It implements overlay.Overlay (Dismiss + Modal)
// so the window honors modal semantics; Esc always dismisses, backdrop click
// dismisses only when allowed.
type dialogOverlayWidget struct {
	widget.WidgetBase

	content           *DialogContentWidget
	windowSize        geometry.Size
	dismissOnBackdrop bool
	onDismiss         func()

	// closeFocused tracks whether keyboard focus currently rests on the close
	// (X) button. Tab moves focus to it; Enter/Space then activates it. Stays
	// false for AlertDialog, whose content has no close button.
	closeFocused bool
}

func newDialogOverlay(content *DialogContentWidget, windowSize geometry.Size, dismissOnBackdrop bool, onDismiss func()) *dialogOverlayWidget {
	o := &dialogOverlayWidget{
		content:           content,
		windowSize:        windowSize,
		dismissOnBackdrop: dismissOnBackdrop,
		onDismiss:         onDismiss,
	}
	o.SetVisible(true)
	o.SetEnabled(true)
	o.AddChild(content)
	return o
}

func (o *dialogOverlayWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := c.Constrain(o.windowSize)
	o.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))

	// Measure content at its capped width, then center it in the window.
	contentSize := o.content.Layout(ctx, geometry.Loose(size))
	x := (size.Width - contentSize.Width) / 2
	y := (size.Height - contentSize.Height) / 2
	if y < 0 {
		y = 0
	}
	// Position the card, then re-layout it: DialogContentWidget.Layout positions
	// its children relative to its own Position(), which was (0,0) on the first
	// pass above — so without this second pass the card background centers but
	// its children (header/content/footer) stay pinned at the window's top-left.
	o.content.SetBounds(geometry.FromPointSize(geometry.Pt(x, y), contentSize))
	o.content.Layout(ctx, geometry.Loose(size))
	return size
}

func (o *dialogOverlayWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	// Pure black @50% in both modes.
	canvas.DrawRect(o.Bounds(), widget.RGBA(0, 0, 0, metrics.OverlayAlpha))
	o.content.Draw(ctx, canvas)
}

func (o *dialogOverlayWidget) Event(ctx widget.Context, e event.Event) bool {
	// Content first (close button, focusable inners).
	if o.content.Event(ctx, e) {
		return true
	}

	if ke, ok := e.(*event.KeyEvent); ok && ke.KeyType == event.KeyPress {
		switch ke.Key {
		case event.KeyEscape:
			o.Dismiss()
			return true
		case event.KeyTab:
			// The close button is the only keyboard-focusable target the
			// overlay manages; Tab toggles focus onto/off it. No close button
			// (AlertDialog) → nothing to focus, fall through.
			if o.content.hasClose() {
				o.setCloseFocused(!o.closeFocused)
				ctx.Invalidate()
				return true
			}
		case event.KeyEnter, event.KeySpace:
			// Activate the focused close button (Enter or Space), mirroring the
			// mouse click. Only when the close button currently holds focus.
			if o.closeFocused && o.content.hasClose() {
				o.content.activateClose()
				return true
			}
		}
	}

	if me, ok := e.(*event.MouseEvent); ok && me.MouseType == event.MousePress {
		if o.dismissOnBackdrop && !o.content.Bounds().Contains(me.Position) {
			o.Dismiss()
			return true
		}
	}

	// Modal: swallow everything else so the tree below is inert.
	return true
}

// setCloseFocused moves keyboard focus onto/off the close button, keeping the
// content's focus ring in sync and requesting a repaint.
func (o *dialogOverlayWidget) setCloseFocused(v bool) {
	if !o.content.hasClose() {
		v = false
	}
	o.closeFocused = v
	o.content.setCloseFocused(v)
}

// Dismiss invokes the dismissal callback.
func (o *dialogOverlayWidget) Dismiss() {
	if o.onDismiss != nil {
		o.onDismiss()
	}
}

// Modal reports true — the dialog blocks the tree below.
func (o *dialogOverlayWidget) Modal() bool { return true }

func (o *dialogOverlayWidget) Children() []widget.Widget {
	return []widget.Widget{o.content}
}

// DialogContentWidget is the centered card: bg Background, 1px Border, radius
// XL (rounded-xl), shadow-LG, padding 24, internal gap 16, with a close X
// button in the top-right corner (unless hidden).
type DialogContentWidget struct {
	widget.WidgetBase

	children []widget.Widget
	theme    *theme.Theme
	hideX    bool

	// onClose is invoked by the close button; wired by the host.
	onClose func()

	closeHovered bool
	closeFocused bool
}

// DialogContent assembles the card body from header/description/footer/content
// sections (or any widgets). The close button is shown by default.
func DialogContent(children ...widget.Widget) *DialogContentWidget {
	c := &DialogContentWidget{
		children: children,
		theme:    CurrentTheme(),
	}
	c.SetVisible(true)
	c.SetEnabled(true)
	for _, ch := range children {
		c.AddChild(ch)
	}
	return c
}

// HideClose removes the top-right close button (used by AlertDialog, or any
// dialog that supplies its own dismissal).
func (c *DialogContentWidget) HideClose() *DialogContentWidget {
	c.hideX = true
	return c
}

// OnClose sets the callback invoked when the close button is pressed. The
// Dialog host wires this automatically; set it directly only when driving a
// DialogContent without a host.
func (c *DialogContentWidget) OnClose(fn func()) *DialogContentWidget {
	c.onClose = fn
	return c
}

// Theme pins a specific theme instead of the process-wide current theme.
func (c *DialogContentWidget) Theme(th *theme.Theme) *DialogContentWidget {
	c.theme = th
	return c
}

func (c *DialogContentWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// Layout stacks children vertically with gap 16 inside padding 24, capping the
// card width at min(512, available−margins).
func (c *DialogContentWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	avail := cons.MaxWidth
	maxW := metrics.DialogMaxWidth
	if avail > 0 {
		if capped := avail - 2*metrics.DialogViewportMargin; capped < maxW {
			maxW = capped
		}
	}
	if maxW < 0 {
		maxW = 0
	}

	innerW := maxW - 2*metrics.DialogPadding
	if innerW < 0 {
		innerW = 0
	}
	childCons := geometry.Loose(geometry.Sz(innerW, 100000))

	x := c.Position().X + metrics.DialogPadding
	y := c.Position().Y + metrics.DialogPadding
	cursorY := y
	for i, ch := range c.children {
		if i > 0 {
			cursorY += metrics.DialogGap
		}
		// Position the child BEFORE laying it out: container children (header/
		// footer DialogSectionWidget) position their own descendants relative to
		// their Position(), so it must be correct when their Layout runs. Setting
		// bounds afterwards (only) leaves nested content pinned at the origin.
		setWidgetBounds(ch, geometry.FromPointSize(geometry.Pt(x, cursorY), geometry.Sz(innerW, 0)))
		sz := ch.Layout(ctx, childCons)
		setWidgetBounds(ch, geometry.FromPointSize(geometry.Pt(x, cursorY), geometry.Sz(innerW, sz.Height)))
		cursorY += sz.Height
	}

	totalH := (cursorY - y) + 2*metrics.DialogPadding
	size := geometry.Sz(maxW, totalH)
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw paints shadow, card surface, border, the children, and the close button.
func (c *DialogContentWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.resolvedTheme()
	tok := th.Active()
	b := c.Bounds()
	radius := th.RadiusXL() // content: rounded-xl

	draw.Shadow(canvas, b, radius, metrics.ShadowLG)
	draw.BorderFill(canvas, b, tok.Background, tok.Border, radius, metrics.DialogBorderWidth)

	for _, ch := range c.children {
		ch.Draw(ctx, canvas)
	}

	if !c.hideX {
		c.drawClose(canvas, b, th, tok)
	}
}

// closeBounds returns the close button's hit/hover box (centered on the X icon
// at the top-right inset).
func (c *DialogContentWidget) closeBounds(b geometry.Rect) geometry.Rect {
	// Icon top-left sits DialogCloseInset from the top and right edges.
	iconX := b.Max.X - metrics.DialogCloseInset - metrics.DialogCloseIconSize
	iconY := b.Min.Y + metrics.DialogCloseInset
	pad := metrics.DialogCloseHitPad
	return geometry.NewRect(
		iconX-pad, iconY-pad,
		metrics.DialogCloseIconSize+2*pad, metrics.DialogCloseIconSize+2*pad,
	)
}

func (c *DialogContentWidget) drawClose(canvas widget.Canvas, b geometry.Rect, th *theme.Theme, tok *theme.Tokens) {
	hit := c.closeBounds(b)

	// Focus ring: legacy ring-2 + ring-offset-2 (gap filled with Background).
	if c.closeFocused {
		draw.OffsetRing(canvas, hit, th.RadiusXS(), tok.Ring,
			metrics.LegacyCloseRingWidth, metrics.LegacyCloseRingOffset, tok.Background)
	}

	opacity := metrics.DialogCloseIdleOpacity
	if c.closeHovered {
		opacity = 1
	}
	iconColor := draw.MulAlpha(tok.Foreground, opacity)

	iconRect := geometry.NewRect(
		hit.Min.X+metrics.DialogCloseHitPad, hit.Min.Y+metrics.DialogCloseHitPad,
		metrics.DialogCloseIconSize, metrics.DialogCloseIconSize,
	)
	icon.Draw(canvas, icons.X, iconRect, iconColor)
}

// Event handles close-button hover/click; everything else falls through to the
// children (footer buttons, etc.).
func (c *DialogContentWidget) Event(ctx widget.Context, e event.Event) bool {
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

// hasClose reports whether this content shows a focusable close button.
func (c *DialogContentWidget) hasClose() bool { return !c.hideX }

// setCloseFocused updates the close button's focus state, repainting the ring
// when it changes. No-op when the close button is hidden (AlertDialog).
func (c *DialogContentWidget) setCloseFocused(v bool) {
	if c.hideX {
		v = false
	}
	if c.closeFocused == v {
		return
	}
	c.closeFocused = v
	c.SetNeedsRedraw(true)
}

// activateClose invokes the close callback (keyboard Enter/Space on the focused
// close button). Returns false when there is no close button to activate.
func (c *DialogContentWidget) activateClose() bool {
	if c.hideX {
		return false
	}
	if c.onClose != nil {
		c.onClose()
	}
	return true
}

// Children returns the section widgets.
func (c *DialogContentWidget) Children() []widget.Widget { return c.children }

// DialogSectionWidget lays out the dialog header (a left-aligned column with
// gap 8) and footer (a right-aligned row with gap 8). It lays itself out rather
// than delegating to primitives.Box because Box has no main-axis justify, which
// the footer's sm:justify-end requires.
type DialogSectionWidget struct {
	widget.WidgetBase
	children []widget.Widget
	row      bool // true = footer (horizontal, right-aligned)
	gap      float32
}

// DialogHeader stacks title + description vertically with gap 8.
func DialogHeader(children ...widget.Widget) *DialogSectionWidget {
	return newDialogSection(children, false, metrics.DialogHeaderGap)
}

// DialogFooter lays buttons in a right-aligned row with gap 8.
func DialogFooter(children ...widget.Widget) *DialogSectionWidget {
	return newDialogSection(children, true, metrics.DialogFooterGap)
}

func newDialogSection(children []widget.Widget, row bool, gap float32) *DialogSectionWidget {
	s := &DialogSectionWidget{children: children, row: row, gap: gap}
	s.SetVisible(true)
	s.SetEnabled(true)
	for _, ch := range children {
		s.AddChild(ch)
	}
	return s
}

func (s *DialogSectionWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if s.row {
		return s.layoutRow(ctx, c)
	}
	return s.layoutColumn(ctx, c)
}

// layoutColumn stacks children top-to-bottom, full-width, gap between.
func (s *DialogSectionWidget) layoutColumn(ctx widget.Context, c geometry.Constraints) geometry.Size {
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

// layoutRow lays children left-to-right at their natural widths, then shifts
// the whole run to the right edge (justify-end), keeping gap between.
func (s *DialogSectionWidget) layoutRow(ctx widget.Context, c geometry.Constraints) geometry.Size {
	width := c.MaxWidth
	childCons := geometry.Loose(geometry.Sz(width, 100000))

	sizes := make([]geometry.Size, len(s.children))
	var runW, maxH float32
	for i, ch := range s.children {
		sizes[i] = ch.Layout(ctx, childCons)
		if i > 0 {
			runW += s.gap
		}
		runW += sizes[i].Width
		if sizes[i].Height > maxH {
			maxH = sizes[i].Height
		}
	}

	if width <= 0 {
		width = runW
	}
	startX := s.Position().X + (width - runW) // right-align
	if startX < s.Position().X {
		startX = s.Position().X
	}
	y := s.Position().Y
	cursorX := startX
	for i, ch := range s.children {
		if i > 0 {
			cursorX += s.gap
		}
		// Vertically center each child in the row.
		cy := y + (maxH-sizes[i].Height)/2
		setWidgetBounds(ch, geometry.FromPointSize(geometry.Pt(cursorX, cy), sizes[i]))
		cursorX += sizes[i].Width
	}

	size := geometry.Sz(width, maxH)
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

func (s *DialogSectionWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	for _, ch := range s.children {
		ch.Draw(ctx, canvas)
	}
}

func (s *DialogSectionWidget) Event(ctx widget.Context, e event.Event) bool {
	for _, ch := range s.children {
		if ch.Event(ctx, e) {
			return true
		}
	}
	return false
}

func (s *DialogSectionWidget) Children() []widget.Widget { return s.children }

// DialogTitle renders the dialog title: 18px / 600 / leading-none.
func DialogTitle(text string) *TypographyWidget {
	return styled(text, metrics.DialogTitleFontSize, metrics.DialogTitleWeight, metrics.DialogTitleFontSize)
}

// DialogDescription renders the dialog description: 14px muted.
func DialogDescription(text string) *TypographyWidget {
	return styled(text, metrics.DialogDescriptionFontSize, 400, metrics.DialogDescriptionLineHeight).Muted()
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*DialogWidget)(nil)
	_ widget.Lifecycle = (*DialogWidget)(nil)
	_ widget.Widget    = (*DialogContentWidget)(nil)
	_ widget.Widget    = (*DialogSectionWidget)(nil)
	_ widget.Widget    = (*dialogTriggerWidget)(nil)
	_ widget.Widget    = (*dialogOverlayWidget)(nil)
)

package graft

import (
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// ButtonWidget is the shadcn Button: 6 variants × 8 sizes, exact pixel
// metrics from metrics.Button.
//
// Architecture decision: graft-owned widget (it draws itself through
// internal/draw + metrics + theme tokens) rather than a painter on
// core/button. The core widget cannot reach shadcn's exact content-driven
// widths (its Layout estimates text width with a 0.55 heuristic and its
// painter cannot affect layout), while graft measures real Geist advances
// via internal/textmetrics, so exact width = text + paddings (+ icon +
// gap) requires owning Layout.
type ButtonWidget struct {
	widget.WidgetBase

	label    string
	children []Widget
	variant  Variant
	size     Size
	icon     *icon.IconData
	iconOnly bool
	onClick  func()
	disabled bool
	full     bool
	width    float32
	style    Style
	theme    *theme.Theme

	disabledSig state.Signal[bool]
	labelSig    state.Signal[string]

	hovered bool
	pressed bool

	// pointerFocus marks that an in-flight mouse press is about to focus
	// the widget, so SetFocused can distinguish pointer focus (no ring)
	// from keyboard focus (focus-visible ring), matching the CSS
	// :focus-visible semantics shadcn relies on.
	pointerFocus bool
	focusVisible bool
}

// Button creates a shadcn button with the given label. Optional children
// render after the label (e.g. a badge), separated by the size's gap.
func Button(label string, children ...Widget) *ButtonWidget {
	b := &ButtonWidget{
		label:    label,
		children: children,
		theme:    CurrentTheme(),
	}
	b.SetVisible(true)
	b.SetEnabled(true)
	type parentSetter interface{ SetParent(widget.Widget) }
	for _, c := range children {
		if ps, ok := c.(parentSetter); ok {
			ps.SetParent(b)
		}
	}
	return b
}

// Variant selects the button variant.
func (b *ButtonWidget) Variant(v Variant) *ButtonWidget { b.variant = v; return b }

// Secondary selects the secondary variant.
func (b *ButtonWidget) Secondary() *ButtonWidget { return b.Variant(VariantSecondary) }

// Destructive selects the destructive variant.
func (b *ButtonWidget) Destructive() *ButtonWidget { return b.Variant(VariantDestructive) }

// Outline selects the outline variant.
func (b *ButtonWidget) Outline() *ButtonWidget { return b.Variant(VariantOutline) }

// Ghost selects the ghost variant.
func (b *ButtonWidget) Ghost() *ButtonWidget { return b.Variant(VariantGhost) }

// Link selects the link variant.
func (b *ButtonWidget) Link() *ButtonWidget { return b.Variant(VariantLink) }

// Size selects the button size.
func (b *ButtonWidget) Size(s Size) *ButtonWidget { b.size = s; return b }

// XS selects the xs size (h 24).
func (b *ButtonWidget) XS() *ButtonWidget { return b.Size(SizeXS) }

// Sm selects the sm size (h 32).
func (b *ButtonWidget) Sm() *ButtonWidget { return b.Size(SizeSM) }

// Lg selects the lg size (h 40).
func (b *ButtonWidget) Lg() *ButtonWidget { return b.Size(SizeLG) }

// Icon adds a leading 16px icon (12px at XS sizes) and switches the
// horizontal padding to the has-[>svg] value.
func (b *ButtonWidget) Icon(ic icon.IconData) *ButtonWidget {
	b.icon = &ic
	return b
}

// IconOnly makes the button a square icon button (36/24/32/40 px per
// size) containing only the icon.
func (b *ButtonWidget) IconOnly(ic icon.IconData) *ButtonWidget {
	b.icon = &ic
	b.iconOnly = true
	return b
}

// OnClick sets the click handler (also fired by Enter/Space when focused).
func (b *ButtonWidget) OnClick(fn func()) *ButtonWidget { b.onClick = fn; return b }

// Disabled sets the disabled state (50% colors, no events, not focusable).
func (b *ButtonWidget) Disabled(v bool) *ButtonWidget { b.disabled = v; return b }

// BindDisabled drives the disabled state from a signal.
func (b *ButtonWidget) BindDisabled(sig state.Signal[bool]) *ButtonWidget {
	b.disabledSig = sig
	return b
}

// BindLabel drives the label text from a signal.
func (b *ButtonWidget) BindLabel(sig state.Signal[string]) *ButtonWidget {
	b.labelSig = sig
	return b
}

// Full makes the button fill the available width.
func (b *ButtonWidget) Full() *ButtonWidget { b.full = true; return b }

// W pins an explicit width in px.
func (b *ButtonWidget) W(px float32) *ButtonWidget { b.width = px; return b }

// Style applies targeted overrides on top of the shadcn look.
func (b *ButtonWidget) Style(fn func(*Style)) *ButtonWidget {
	if fn != nil {
		fn(&b.style)
	}
	return b
}

// Theme pins a specific theme instead of the one snapshotted from
// CurrentTheme at construction.
func (b *ButtonWidget) Theme(th *theme.Theme) *ButtonWidget {
	if th != nil {
		b.theme = th
	}
	return b
}

// resolvedLabel returns the current label text (signal wins).
func (b *ButtonWidget) resolvedLabel() string {
	if b.labelSig != nil {
		return b.labelSig.Get()
	}
	return b.label
}

// resolvedDisabled returns the current disabled state (signal wins).
func (b *ButtonWidget) resolvedDisabled() bool {
	if b.disabledSig != nil {
		return b.disabledSig.Get()
	}
	return b.disabled
}

// sizeMetrics maps the Size enum (and IconOnly) to the metrics table.
func (b *ButtonWidget) sizeMetrics() metrics.ButtonSize {
	s := b.size
	if b.iconOnly {
		switch s {
		case SizeXS:
			s = SizeIconXS
		case SizeSM:
			s = SizeIconSM
		case SizeLG:
			s = SizeIconLG
		case SizeDefault:
			s = SizeIcon
		}
	}
	switch s {
	case SizeXS:
		return metrics.Button.XS
	case SizeSM:
		return metrics.Button.SM
	case SizeLG:
		return metrics.Button.LG
	case SizeIcon:
		return metrics.Button.Icon
	case SizeIconXS:
		return metrics.Button.IconXS
	case SizeIconSM:
		return metrics.Button.IconSM
	case SizeIconLG:
		return metrics.Button.IconLG
	default:
		return metrics.Button.Default
	}
}

// squareSize reports whether the resolved size is one of the square icon
// sizes.
func (b *ButtonWidget) squareSize() bool {
	if b.iconOnly {
		return true
	}
	switch b.size {
	case SizeIcon, SizeIconXS, SizeIconSM, SizeIconLG:
		return true
	}
	return false
}

// fontFamily resolves the registered family for the label weight,
// honoring custom theme fonts.
func (b *ButtonWidget) fontFamily() string {
	if b.theme.FontSans != theme.DefaultFontSans {
		return b.theme.FontSans
	}
	return fonts.Family(metrics.Button.FontWeight)
}

// padX returns the active horizontal padding (has-[>svg] value when a
// leading icon is present).
func (b *ButtonWidget) padX() float32 {
	m := b.sizeMetrics()
	p := m.PadX
	if b.icon != nil && !b.iconOnly {
		p = m.PadXWithIcon
	}
	if b.style.PadX != nil {
		p = *b.style.PadX
	}
	return p
}

// buttonContent is the resolved horizontal content layout in widget-local
// coordinates.
type buttonContent struct {
	iconX, iconSize float32
	textX, textW    float32
	childX          []float32
}

// contentWidth measures the natural content width (icon + gap + label +
// gap-separated children).
func (b *ButtonWidget) contentWidth(ctx widget.Context) float32 {
	m := b.sizeMetrics()
	if b.squareSize() {
		return m.IconSize
	}
	w := textmetrics.Width(b.fontFamily(), m.FontSize, b.resolvedLabel())
	if b.icon != nil {
		w += m.IconSize + m.Gap
	}
	for _, c := range b.children {
		sz := c.Layout(ctx, geometry.Loose(geometry.Sz(geometry.Infinity, m.Height)))
		w += m.Gap + sz.Width
	}
	return w
}

// contentLayout positions the icon, label, and children inside a button of
// final width w (content centered, per justify-center).
func (b *ButtonWidget) contentLayout(ctx widget.Context, w float32) buttonContent {
	m := b.sizeMetrics()
	var c buttonContent
	c.iconSize = m.IconSize

	if b.squareSize() {
		c.iconX = (w - m.IconSize) / 2
		return c
	}

	c.textW = textmetrics.Width(b.fontFamily(), m.FontSize, b.resolvedLabel())
	content := c.textW
	if b.icon != nil {
		content += m.IconSize + m.Gap
	}
	childSizes := make([]geometry.Size, len(b.children))
	for i, ch := range b.children {
		childSizes[i] = ch.Layout(ctx, geometry.Loose(geometry.Sz(geometry.Infinity, m.Height)))
		content += m.Gap + childSizes[i].Width
	}

	x := (w - content) / 2
	if min := b.padX(); x < min {
		x = min
	}
	if b.icon != nil {
		c.iconX = x
		x += m.IconSize + m.Gap
	}
	c.textX = x
	x += c.textW
	c.childX = make([]float32, len(b.children))
	for i := range b.children {
		x += m.Gap
		c.childX[i] = x
		x += childSizes[i].Width
	}
	return c
}

// Layout sizes the button: exact width = text advance + paddings (+ icon
// + gap), fixed shadcn height per size.
func (b *ButtonWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	m := b.sizeMetrics()

	var w float32
	switch {
	case b.width > 0:
		w = b.width
	case b.full && c.MaxWidth < geometry.Infinity:
		w = c.MaxWidth
	case b.squareSize():
		w = m.Height // square icon button
	default:
		w = b.contentWidth(ctx) + 2*b.padX()
	}
	if b.style.MinWidth != nil && w < *b.style.MinWidth {
		w = *b.style.MinWidth
	}
	if b.style.MaxWidth != nil && w > *b.style.MaxWidth {
		w = *b.style.MaxWidth
	}

	h := m.Height
	if b.style.PadY != nil {
		// Escape hatch: vertical padding around the text-sm line box.
		if lh := m.FontSize + 6; 2**b.style.PadY+lh > h {
			h = 2**b.style.PadY + lh
		}
	}

	size := c.Constrain(geometry.Sz(w, h))

	// Position children for drawing/hit-testing.
	cl := b.contentLayout(ctx, size.Width)
	for i, ch := range b.children {
		sz := ch.Layout(ctx, geometry.Loose(geometry.Sz(geometry.Infinity, size.Height)))
		ch.(interface{ SetBounds(geometry.Rect) }).SetBounds(geometry.NewRect(
			cl.childX[i], (size.Height-sz.Height)/2, sz.Width, sz.Height))
	}

	b.SetBounds(geometry.FromPointSize(b.Position(), size))
	return size
}

// buttonVisual is the resolved per-state visual style.
type buttonVisual struct {
	bg        widget.Color
	hasBg     bool
	fg        widget.Color
	border    widget.Color
	borderW   float32
	shadow    bool
	underline bool
}

// visual resolves the variant × state colors from the active token set
// (metrics.Button alphas; DESIGN.md section 5.3 state table).
func (b *ButtonWidget) visual(tok *theme.Tokens, dark, hovered bool) buttonVisual {
	v := buttonVisual{fg: tok.Foreground}
	switch b.variant {
	case VariantSecondary:
		v.fg = tok.SecondaryForeground
		v.bg, v.hasBg = tok.Secondary, true
		if hovered {
			v.bg = draw.Alpha(tok.Secondary, metrics.Button.HoverSecondaryAlpha)
		}
	case VariantDestructive:
		v.fg = tok.DestructiveForeground // text-white (token defaults to white)
		v.bg, v.hasBg = tok.Destructive, true
		if dark {
			v.bg = draw.Alpha(tok.Destructive, metrics.Button.DarkDestructiveBgAlpha)
		}
		if hovered {
			v.bg = draw.Alpha(tok.Destructive, metrics.Button.HoverDestructiveAlpha)
		}
	case VariantOutline:
		v.borderW = metrics.Button.BorderWidth
		v.shadow = true
		if dark {
			v.border = tok.Input // dark:border-input (alpha kept)
			v.bg, v.hasBg = draw.MulAlpha(tok.Input, metrics.Button.DarkOutlineBgAlpha), true
			if hovered {
				v.bg = draw.MulAlpha(tok.Input, metrics.Button.DarkOutlineHoverAlpha)
				v.fg = tok.AccentForeground
			}
		} else {
			v.border = tok.Border
			v.bg, v.hasBg = tok.Background, true
			if hovered {
				v.bg = tok.Accent
				v.fg = tok.AccentForeground
			}
		}
	case VariantGhost:
		if hovered {
			v.fg = tok.AccentForeground
			if dark {
				v.bg, v.hasBg = draw.Alpha(tok.Accent, metrics.Button.DarkGhostHoverAlpha), true
			} else {
				v.bg, v.hasBg = tok.Accent, true
			}
		}
	case VariantLink:
		v.fg = tok.Primary
		v.underline = hovered
	default: // VariantDefault
		v.fg = tok.PrimaryForeground
		v.bg, v.hasBg = tok.Primary, true
		if hovered {
			v.bg = draw.Alpha(tok.Primary, metrics.Button.HoverPrimaryAlpha)
		}
	}
	return v
}

// radius returns the corner radius (rounded-lg via the theme scale).
func (b *ButtonWidget) radius() float32 {
	if b.style.Radius != nil {
		return *b.style.Radius
	}
	return b.theme.RadiusLG()
}

// Draw paints the button: shadow, fill, border, icon, label, underline,
// and focus ring, resolving all colors from the active token set.
func (b *ButtonWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !b.IsVisible() {
		return
	}
	th := b.theme
	tok := th.Active()
	dark := th.IsDark()
	bounds := b.Bounds()
	disabled := b.resolvedDisabled()
	hovered := (b.hovered || b.pressed) && !disabled // pressed == hovered (DESIGN 5.3)
	radius := b.radius()
	m := b.sizeMetrics()

	v := b.visual(tok, dark, hovered)
	if b.style.Background != nil {
		v.bg, v.hasBg = *b.style.Background, true
	}
	if b.style.Foreground != nil {
		v.fg = *b.style.Foreground
	}

	if v.shadow {
		drawButtonShadow(canvas, bounds, radius, disabled)
	}
	if v.hasBg && v.bg.A > 0 {
		canvas.DrawRoundRect(bounds, draw.Fade(v.bg, disabled), radius)
	}

	cl := b.contentLayout(ctx, bounds.Width())
	fg := draw.Fade(v.fg, disabled)

	if b.icon != nil {
		iconRect := geometry.NewRect(
			bounds.Min.X+cl.iconX,
			bounds.Min.Y+(bounds.Height()-cl.iconSize)/2,
			cl.iconSize, cl.iconSize)
		icon.Draw(canvas, *b.icon, iconRect, fg)
	}

	if !b.squareSize() {
		label := b.resolvedLabel()
		textBounds := geometry.NewRect(bounds.Min.X+cl.textX, bounds.Min.Y, cl.textW, bounds.Height())
		family := b.fontFamily()
		if std, ok := canvas.(widget.StyledTextDrawer); ok {
			std.DrawStyledText(label, textBounds, widget.TextStyle{
				FontFamily: family,
				FontSize:   m.FontSize,
				Color:      fg,
				Align:      widget.TextAlignLeft,
			})
		} else {
			canvas.DrawText(label, textBounds, m.FontSize, fg, metrics.Button.FontWeight >= 600, widget.TextAlignLeft)
		}

		if v.underline {
			ascent, descent, _ := textmetrics.LineHeight(family, m.FontSize)
			baseline := bounds.Min.Y + (bounds.Height()-(ascent+descent))/2 + ascent
			y := baseline + metrics.Button.UnderlineOffset
			canvas.DrawLine(
				geometry.Pt(textBounds.Min.X, y),
				geometry.Pt(textBounds.Min.X+cl.textW, y),
				fg, metrics.Button.UnderlineWidth)
		}
	}

	if len(b.children) > 0 {
		canvas.PushTransform(bounds.Min)
		for _, ch := range b.children {
			widget.StampScreenOrigin(ch, canvas)
			widget.DrawChild(ch, ctx, canvas)
		}
		canvas.PopTransform()
	}

	if v.borderW > 0 {
		border := v.border
		if b.focusVisible && !disabled {
			border = tok.Ring // focus-visible:border-ring
		}
		draw.InsideBorder(canvas, bounds, radius, draw.Fade(border, disabled), v.borderW)
	}

	if b.focusVisible && !disabled {
		ring := draw.Alpha(tok.Ring, metrics.RingAlpha)
		if b.variant == VariantDestructive {
			a := metrics.InvalidRingAlphaLight // focus-visible:ring-destructive/20
			if dark {
				a = metrics.InvalidRingAlphaDark // dark:focus-visible:ring-destructive/40
			}
			ring = draw.Alpha(tok.Destructive, a)
		}
		draw.FocusRing(canvas, bounds, radius, ring)
	}
}

// drawButtonShadow draws shadow-xs, fading the layers when disabled.
func drawButtonShadow(canvas widget.Canvas, bounds geometry.Rect, radius float32, disabled bool) {
	if !disabled {
		draw.Shadow(canvas, bounds, radius, metrics.ShadowXS)
		return
	}
	faded := make([]metrics.ShadowLayer, len(metrics.ShadowXS))
	for i, l := range metrics.ShadowXS {
		l.Alpha *= metrics.DisabledOpacity
		faded[i] = l
	}
	draw.Shadow(canvas, bounds, radius, faded)
}

// Event handles hover, press/release click semantics, and keyboard
// activation (Enter/Space when focused).
func (b *ButtonWidget) Event(ctx widget.Context, e event.Event) bool {
	if b.resolvedDisabled() {
		return false // disabled:pointer-events-none
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return b.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		if !b.IsFocused() || ev.KeyType != event.KeyPress {
			return false
		}
		if ev.Key == event.KeyEnter || ev.Key == event.KeySpace {
			b.activate()
			return true
		}
	}
	return false
}

func (b *ButtonWidget) mouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		b.hovered = true
		ctx.SetCursor(widget.CursorPointer) // shadcn base layer: button:not(:disabled){cursor:pointer}
		b.invalidate(ctx)
		return true
	case event.MouseLeave:
		b.hovered = false
		b.pressed = false
		ctx.SetCursor(widget.CursorDefault)
		b.invalidate(ctx)
		return true
	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		b.pressed = true
		b.pointerFocus = true
		ctx.RequestFocus(b)
		b.pointerFocus = false
		b.invalidate(ctx)
		return true
	case event.MouseRelease:
		wasPressed := b.pressed
		b.pressed = false
		b.invalidate(ctx)
		if wasPressed && b.Bounds().Contains(e.Position) {
			b.activate()
		}
		return true
	}
	return false
}

func (b *ButtonWidget) activate() {
	if b.onClick != nil {
		b.onClick()
	}
}

func (b *ButtonWidget) invalidate(ctx widget.Context) {
	b.SetNeedsRedraw(true)
	ctx.InvalidateRect(b.Bounds())
}

// IsFocusable reports whether the button can take keyboard focus.
func (b *ButtonWidget) IsFocusable() bool {
	return b.IsVisible() && b.IsEnabled() && !b.resolvedDisabled()
}

// SetFocused tracks focus-visible: focus gained outside a mouse press
// (i.e. keyboard traversal) shows the ring; pointer focus does not.
func (b *ButtonWidget) SetFocused(focused bool) {
	b.focusVisible = focused && !b.pointerFocus
	b.WidgetBase.SetFocused(focused)
	// MarkRedrawLocal, NOT SetNeedsRedraw: SetFocused runs inside
	// ctx.RequestFocus, which holds the context write lock. SetNeedsRedraw
	// propagates into ctx.RegisterDirtyBoundary (an RLock on that same
	// non-reentrant RWMutex) and deadlocks. MarkRedrawLocal only flips this
	// widget's own dirty flag, which the frame loop still picks up.
	b.MarkRedrawLocal()
}

// Mount registers signal bindings (widget.Lifecycle).
func (b *ButtonWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if b.disabledSig != nil {
		b.AddBinding(state.BindToScheduler(b.disabledSig, b, sched))
	}
	if b.labelSig != nil {
		b.AddBinding(state.BindToScheduler(b.labelSig, b, sched))
	}
}

// Unmount implements widget.Lifecycle; bindings are cleaned automatically.
func (b *ButtonWidget) Unmount() {}

// Children returns extra content widgets passed to Button.
func (b *ButtonWidget) Children() []widget.Widget { return b.children }

// AccessibilityRole returns the button role.
func (b *ButtonWidget) AccessibilityRole() a11y.Role { return a11y.RoleButton }

// AccessibilityLabel returns the button label.
func (b *ButtonWidget) AccessibilityLabel() string { return b.resolvedLabel() }

// AccessibilityHint returns no hint.
func (b *ButtonWidget) AccessibilityHint() string { return "" }

// AccessibilityValue returns no value.
func (b *ButtonWidget) AccessibilityValue() string { return "" }

// AccessibilityState reports the disabled and focused states.
func (b *ButtonWidget) AccessibilityState() a11y.State {
	return a11y.State{Disabled: b.resolvedDisabled(), Focused: b.IsFocused()}
}

// AccessibilityActions returns the click action.
func (b *ButtonWidget) AccessibilityActions() []a11y.Action {
	return []a11y.Action{a11y.ActionClick, a11y.ActionFocus}
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*ButtonWidget)(nil)
	_ widget.Focusable = (*ButtonWidget)(nil)
	_ widget.Lifecycle = (*ButtonWidget)(nil)
	_ a11y.Accessible  = (*ButtonWidget)(nil)
)

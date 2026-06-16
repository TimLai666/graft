package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// BadgeWidget is shadcn's Badge: a near-pill label (rounded-4xl ≈ 26px,
// px 8 / py 2, 12px/500 text, 1px transparent border) with six variants.
// Hover states follow shadcn's [a&] semantics: they apply only when the
// badge is clickable (OnClick set), mirroring badges rendered as anchors.
//
// Architecture: graft-owned widget (no gogpu/ui core widget wrapped) — the
// badge needs no interaction machinery beyond hover/click/focus, and
// drawing directly via metrics + theme tokens gives exact pixel control
// over the corners, border, and text placement.
type BadgeWidget struct {
	widget.WidgetBase

	variant   Variant
	label     *TypographyWidget
	ic        *icon.IconData
	extra     []Widget
	onClick   func()
	theme     *theme.Theme
	hovered   bool
	pressed   bool
	focusRing bool // focus arrived via keyboard (focus-visible)
	noRing    bool // transient: focus arriving from a mouse press

	iconRect geometry.Rect // icon bounds in badge-local coords (zero if no icon)
}

// Badge creates a badge with the given label text. Extra children (small
// widgets) are laid out after the text with the badge gap.
func Badge(text string, children ...Widget) *BadgeWidget {
	b := &BadgeWidget{variant: VariantDefault, extra: children}
	b.label = styled(text, metrics.BadgeFontSize, metrics.BadgeFontWeight, metrics.BadgeLineHeight).
		ColorToken(func(tok *theme.Tokens) widget.Color { return b.textColor(tok) })
	b.label.SetParent(b)
	for _, child := range children {
		if ps, ok := child.(interface{ SetParent(widget.Widget) }); ok {
			ps.SetParent(b)
		}
	}
	b.SetVisible(true)
	b.SetEnabled(true)
	return b
}

// Variant selects the badge variant.
func (b *BadgeWidget) Variant(v Variant) *BadgeWidget {
	b.variant = v
	return b
}

// Secondary selects the secondary variant.
func (b *BadgeWidget) Secondary() *BadgeWidget { return b.Variant(VariantSecondary) }

// Destructive selects the destructive variant.
func (b *BadgeWidget) Destructive() *BadgeWidget { return b.Variant(VariantDestructive) }

// Outline selects the outline variant.
func (b *BadgeWidget) Outline() *BadgeWidget { return b.Variant(VariantOutline) }

// Ghost selects the ghost variant.
func (b *BadgeWidget) Ghost() *BadgeWidget { return b.Variant(VariantGhost) }

// Link selects the link variant.
func (b *BadgeWidget) Link() *BadgeWidget { return b.Variant(VariantLink) }

// Icon adds a leading 12px icon ([&>svg]:size-3).
func (b *BadgeWidget) Icon(ic icon.IconData) *BadgeWidget {
	b.ic = &ic
	return b
}

// OnClick makes the badge clickable, enabling the [a&] hover states, the
// pointer cursor, and keyboard focus/activation.
func (b *BadgeWidget) OnClick(fn func()) *BadgeWidget {
	b.onClick = fn
	return b
}

// Theme pins a specific theme instead of the process-wide current theme.
func (b *BadgeWidget) Theme(th *theme.Theme) *BadgeWidget {
	b.theme = th
	b.label.Theme(th)
	return b
}

func (b *BadgeWidget) resolvedTheme() *theme.Theme {
	if b.theme != nil {
		return b.theme
	}
	return CurrentTheme()
}

// clickable reports whether [a&] semantics apply.
func (b *BadgeWidget) clickable() bool { return b.onClick != nil }

// hoverActive reports whether the [a&]:hover styles should render.
func (b *BadgeWidget) hoverActive() bool {
	return b.clickable() && (b.hovered || b.pressed) && b.IsEnabled()
}

// textColor resolves the label/icon color for the current variant + state.
func (b *BadgeWidget) textColor(tok *theme.Tokens) widget.Color {
	switch b.variant {
	case VariantSecondary:
		return tok.SecondaryForeground
	case VariantDestructive:
		// shadcn uses literal text-white; DestructiveForeground defaults
		// to white in both modes (DESIGN.md §8.16).
		return tok.DestructiveForeground
	case VariantOutline, VariantGhost:
		if b.hoverActive() {
			return tok.AccentForeground // [a&]:hover:text-accent-foreground
		}
		return tok.Foreground
	case VariantLink:
		return tok.Primary // text-primary
	default:
		return tok.PrimaryForeground
	}
}

// fillColor resolves the pill fill for the current variant + state. A zero
// alpha means no fill.
func (b *BadgeWidget) fillColor(tok *theme.Tokens, dark bool) widget.Color {
	hover := b.hoverActive()
	switch b.variant {
	case VariantSecondary:
		if hover {
			return draw.Alpha(tok.Secondary, metrics.BadgeHoverAlpha) // [a&]:hover:bg-secondary/90
		}
		return tok.Secondary
	case VariantDestructive:
		if hover {
			return draw.Alpha(tok.Destructive, metrics.BadgeHoverAlpha) // [a&]:hover:bg-destructive/90
		}
		if dark {
			return draw.Alpha(tok.Destructive, metrics.BadgeDarkDestructiveAlpha) // dark:bg-destructive/60
		}
		return tok.Destructive
	case VariantOutline, VariantGhost:
		if hover {
			return tok.Accent // [a&]:hover:bg-accent
		}
		return widget.Color{}
	case VariantLink:
		return widget.Color{}
	default:
		if hover {
			return draw.Alpha(tok.Primary, metrics.BadgeHoverAlpha) // [a&]:hover:bg-primary/90
		}
		return tok.Primary
	}
}

// borderColor resolves the 1px border color (transparent except outline
// and focus-visible).
func (b *BadgeWidget) borderColor(tok *theme.Tokens) widget.Color {
	if b.focusRing && b.IsFocused() {
		return tok.Ring // focus-visible:border-ring
	}
	if b.variant == VariantOutline {
		return tok.Border // outline: border-border
	}
	return widget.Color{}
}

// ringColor resolves the focus-visible ring color.
func (b *BadgeWidget) ringColor(tok *theme.Tokens, dark bool) widget.Color {
	if b.variant == VariantDestructive {
		// focus-visible:ring-destructive/20 dark:.../40
		a := metrics.InvalidRingAlphaLight
		if dark {
			a = metrics.InvalidRingAlphaDark
		}
		return draw.Alpha(tok.Destructive, a)
	}
	return draw.Alpha(tok.Ring, metrics.RingAlpha) // focus-visible:ring-ring/50
}

// Layout measures icon + text + extra children in a row: the box is the
// content line (16px) plus py 2 and the 1px border on every side.
func (b *BadgeWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	inset := metrics.BadgePadX + metrics.BadgeBorderWidth
	insetY := metrics.BadgePadY + metrics.BadgeBorderWidth
	contentH := metrics.BadgeLineHeight
	loose := geometry.Loose(geometry.Sz(geometry.Infinity, geometry.Infinity))

	x := inset
	first := true
	advance := func(w float32) (start float32) {
		if !first {
			x += metrics.BadgeGap
		}
		first = false
		start = x
		x += w
		return start
	}

	b.iconRect = geometry.Rect{}
	if b.ic != nil {
		ix := advance(metrics.BadgeIconSize)
		iy := insetY + (contentH-metrics.BadgeIconSize)/2
		b.iconRect = geometry.NewRect(ix, iy, metrics.BadgeIconSize, metrics.BadgeIconSize)
	}
	if b.label.Content() != "" {
		size := b.label.Layout(ctx, loose)
		lx := advance(size.Width)
		b.label.SetBounds(geometry.FromPointSize(geometry.Pt(lx, insetY), size))
	}
	for _, child := range b.extra {
		size := child.Layout(ctx, loose)
		cx := advance(size.Width)
		cy := insetY + (contentH-size.Height)/2
		if sb, ok := child.(interface{ SetBounds(geometry.Rect) }); ok {
			sb.SetBounds(geometry.FromPointSize(geometry.Pt(cx, cy), size))
		}
		if size.Height > contentH {
			contentH = size.Height
		}
	}

	w := x + inset
	h := contentH + 2*insetY
	size := c.Constrain(geometry.Sz(w, h))
	b.SetBounds(geometry.FromPointSize(b.Position(), size))
	return size
}

// Draw paints pill fill, focus ring, border, icon, text, and the link
// hover underline, resolving every color from the active token set.
func (b *BadgeWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !b.IsVisible() {
		return
	}
	th := b.resolvedTheme()
	tok := th.Active()
	dark := th.IsDark()
	bounds := b.Bounds()
	radius := th.Radius4XL() // rounded-4xl ≈ 26px

	fill := b.fillColor(tok, dark)
	border := b.borderColor(tok)
	if border.A == 0 && fill.A > 0 {
		// Solid/secondary variants: plain fill, no border.
		canvas.DrawRoundRect(bounds, fill, radius)
	}
	if b.focusRing && b.IsFocused() {
		draw.FocusRing(canvas, bounds, radius, b.ringColor(tok, dark))
	}
	if border.A > 0 {
		// BorderFill (fill + ring) instead of an inside stroke, which renders
		// as a solid gray box on the GPU. Outline variant is transparent
		// (use the page background) unless a hover fill is active.
		bg := tok.Background
		if fill.A > 0 {
			bg = fill
		}
		draw.BorderFill(canvas, bounds, bg, border, radius, metrics.BadgeBorderWidth)
	}

	canvas.PushTransform(bounds.Min)
	if b.ic != nil {
		icon.Draw(canvas, *b.ic, b.iconRect, b.textColor(tok))
	}
	if b.label.Content() != "" {
		widget.StampScreenOrigin(b.label, canvas)
		widget.DrawChild(b.label, ctx, canvas)
		if b.variant == VariantLink && b.hoverActive() {
			b.drawUnderline(canvas, th, tok)
		}
	}
	for _, child := range b.extra {
		widget.StampScreenOrigin(child, canvas)
		widget.DrawChild(child, ctx, canvas)
	}
	canvas.PopTransform()
}

// drawUnderline renders the link-variant hover underline 4px below the
// text baseline ([a&]:hover:underline + underline-offset-4). The baseline
// is recovered from the canvas convention of vertically centering text in
// its bounds via font metrics.
func (b *BadgeWidget) drawUnderline(canvas widget.Canvas, th *theme.Theme, tok *theme.Tokens) {
	lb := b.label.Bounds()
	family := badgeFontFamily(th)
	ascent, descent, _ := textmetrics.LineHeight(family, metrics.BadgeFontSize)
	baseline := lb.Min.Y + lb.Height()/2 + (ascent-descent)/2
	y := baseline + metrics.BadgeUnderlineOffset - metrics.BadgeUnderlineWidth/2
	canvas.DrawRect(
		geometry.NewRect(lb.Min.X, y, lb.Width(), metrics.BadgeUnderlineWidth),
		b.textColor(tok),
	)
}

// badgeFontFamily resolves the badge label family, honoring custom theme
// fonts (custom families register a single face; weight mapping applies
// only to the stock Geist family).
func badgeFontFamily(th *theme.Theme) string {
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(metrics.BadgeFontWeight)
}

// Event implements hover, click, and keyboard activation. Hover/press
// visuals and focus only apply when the badge is clickable.
func (b *BadgeWidget) Event(ctx widget.Context, e event.Event) bool {
	if !b.IsEnabled() {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return b.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		if !b.clickable() || !b.IsFocused() || ev.KeyType != event.KeyPress {
			return false
		}
		if ev.Key == event.KeyEnter || ev.Key == event.KeySpace {
			if b.onClick != nil {
				b.onClick()
			}
			return true
		}
	}
	// NOTE: *event.FocusEvent is window-level (OS focus gained/lost, no target
	// widget) and is broadcast to every widget; consuming it here marked the
	// badge focus-visible whenever the window was focused (focus ring rendered
	// as a solid box, gg#369). Per-widget focus is driven by ctx.RequestFocus →
	// SetFocused (mouseEvent sets noRing around RequestFocus for pointer focus).
	return false
}

func (b *BadgeWidget) mouseEvent(ctx widget.Context, me *event.MouseEvent) bool {
	switch me.MouseType {
	case event.MouseEnter:
		b.hovered = true
		if b.clickable() {
			ctx.SetCursor(widget.CursorPointer)
		}
		b.redraw(ctx)
		return true
	case event.MouseLeave:
		b.hovered = false
		if b.clickable() {
			ctx.SetCursor(widget.CursorDefault)
		}
		b.redraw(ctx)
		return true
	case event.MousePress:
		if !b.clickable() || me.Button != event.ButtonLeft {
			return false
		}
		b.pressed = true
		// Focus from a pointer press must not show the focus-visible ring.
		b.noRing = true
		ctx.RequestFocus(b)
		b.noRing = false
		b.redraw(ctx)
		return true
	case event.MouseRelease:
		if !b.clickable() {
			return false
		}
		wasPressed := b.pressed
		b.pressed = false
		b.redraw(ctx)
		if wasPressed && b.Bounds().Contains(me.Position) && b.onClick != nil {
			b.onClick()
		}
		return wasPressed
	}
	return false
}

func (b *BadgeWidget) redraw(ctx widget.Context) {
	b.SetNeedsRedraw(true)
	if ctx != nil {
		ctx.InvalidateRect(b.Bounds())
	}
}

// SetFocused tracks focus-visible semantics: focus gained while a pointer
// press is in flight suppresses the ring; focus gained any other way
// (Tab traversal, FocusGained events) shows it.
func (b *BadgeWidget) SetFocused(focused bool) {
	b.WidgetBase.SetFocused(focused)
	if focused {
		b.focusRing = !b.noRing
	} else {
		b.focusRing = false
	}
}

// IsFocusable reports keyboard focusability: only clickable badges
// participate in Tab traversal ([a&] semantics).
func (b *BadgeWidget) IsFocusable() bool {
	return b.clickable() && b.IsEnabled() && b.IsVisible()
}

// Children returns the label and any extra children.
func (b *BadgeWidget) Children() []widget.Widget {
	children := make([]widget.Widget, 0, 1+len(b.extra))
	if b.label.Content() != "" {
		children = append(children, b.label)
	}
	return append(children, b.extra...)
}

var (
	_ widget.Widget    = (*BadgeWidget)(nil)
	_ widget.Focusable = (*BadgeWidget)(nil)
)

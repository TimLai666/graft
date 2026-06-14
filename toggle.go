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

// ToggleWidget is the shadcn Toggle: a two-state (on/off) button. Unlike the
// plain Button it carries a pressed/on state that fills with --accent.
//
// Architecture decision: graft-OWNED (same reasoning as Button — exact
// content-driven width from real Geist advances requires owning Layout).
//
// Spec (compare/shadcn-spec.md "toggle"): base rounded-md
// text-sm/500, hover bg-muted text-muted-foreground, on-state bg-accent
// text-accent-foreground; default variant transparent, outline variant
// 1px --input border + shadow-xs (hover bg-accent); sizes default h-7/min-w-7
// px-2.5, sm h-8/min-w-8 px-1.5, lg h-10/min-w-10 px-2.5; icon 16.
//
//	graft.Toggle("Bold").Bind(boldSig)
//	graft.Toggle("").Icon(icons.Bold).Outline().OnChange(func(on bool){ ... })
type ToggleWidget struct {
	widget.WidgetBase

	label    string
	icon     *icon.IconData
	variant  Variant // VariantDefault or VariantOutline
	size     Size
	on       bool
	onSig    state.Signal[bool]
	onChange func(bool)
	disabled bool
	theme    *theme.Theme

	hovered      bool
	pressed      bool
	pointerFocus bool
	focusVisible bool
}

// Toggle creates a toggle button with the given label. Pass an empty label
// and call Icon for an icon-only toggle.
func Toggle(label string) *ToggleWidget {
	t := &ToggleWidget{
		label: label,
		theme: CurrentTheme(),
	}
	t.SetVisible(true)
	t.SetEnabled(true)
	return t
}

// Icon adds a leading 16px icon.
func (t *ToggleWidget) Icon(ic icon.IconData) *ToggleWidget {
	t.icon = &ic
	return t
}

// Outline selects the outline variant (1px --input border + shadow-xs).
func (t *ToggleWidget) Outline() *ToggleWidget { t.variant = VariantOutline; return t }

// Default selects the default variant (transparent).
func (t *ToggleWidget) Default() *ToggleWidget { t.variant = VariantDefault; return t }

// Size selects the toggle size.
func (t *ToggleWidget) Size(s Size) *ToggleWidget { t.size = s; return t }

// Sm selects the sm size (h-8).
func (t *ToggleWidget) Sm() *ToggleWidget { return t.Size(SizeSM) }

// Lg selects the lg size (h-10).
func (t *ToggleWidget) Lg() *ToggleWidget { return t.Size(SizeLG) }

// On sets the initial pressed/on state (uncontrolled).
func (t *ToggleWidget) On(v bool) *ToggleWidget { t.on = v; return t }

// Pressed is an alias for On, matching shadcn's `pressed` prop name.
func (t *ToggleWidget) Pressed(v bool) *ToggleWidget { return t.On(v) }

// Bind makes the on state controlled by a boolean signal.
func (t *ToggleWidget) Bind(sig state.Signal[bool]) *ToggleWidget {
	t.onSig = sig
	return t
}

// OnChange registers an observer fired whenever the on state changes.
func (t *ToggleWidget) OnChange(fn func(bool)) *ToggleWidget {
	t.onChange = fn
	return t
}

// Disabled sets the disabled state (50% colors, no events, not focusable).
func (t *ToggleWidget) Disabled(v bool) *ToggleWidget { t.disabled = v; return t }

// Theme pins a specific theme instead of the snapshotted current theme.
func (t *ToggleWidget) Theme(th *theme.Theme) *ToggleWidget {
	if th != nil {
		t.theme = th
	}
	return t
}

// isOn returns the current on state (bound signal wins).
func (t *ToggleWidget) isOn() bool {
	if t.onSig != nil {
		return t.onSig.Get()
	}
	return t.on
}

// setOn updates the on state (writing through to the bound signal when
// controlled) and fires OnChange.
func (t *ToggleWidget) setOn(v bool) {
	if t.onSig != nil {
		t.onSig.Set(v)
	} else {
		t.on = v
	}
	if t.onChange != nil {
		t.onChange(v)
	}
}

// sizeMetrics maps the Size enum to the metrics table.
func (t *ToggleWidget) sizeMetrics() metrics.ToggleSize {
	switch t.size {
	case SizeSM:
		return metrics.Toggle.SM
	case SizeLG:
		return metrics.Toggle.LG
	default:
		return metrics.Toggle.Default
	}
}

// fontFamily resolves the label family for the toggle weight.
func (t *ToggleWidget) fontFamily() string {
	if t.theme.FontSans != theme.DefaultFontSans {
		return t.theme.FontSans
	}
	return fonts.Family(metrics.Toggle.FontWeight)
}

// contentWidth measures the natural content width (icon + gap + label).
func (t *ToggleWidget) contentWidth() float32 {
	m := metrics.Toggle
	var w float32
	if t.label != "" {
		w += textmetrics.Width(t.fontFamily(), m.FontSize, t.label)
	}
	if t.icon != nil {
		w += m.IconSize
		if t.label != "" {
			w += m.Gap
		}
	}
	return w
}

// Layout sizes the toggle: max(content + 2*padX, minWidth) × fixed height.
func (t *ToggleWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	sm := t.sizeMetrics()
	w := t.contentWidth() + 2*sm.PadX
	if w < sm.MinWidth {
		w = sm.MinWidth
	}
	size := c.Constrain(geometry.Sz(w, sm.Height))
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

// Draw paints shadow (outline), the on/hover fill, the border (outline),
// the icon, the label, and the focus ring, resolving colors at draw time.
func (t *ToggleWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !t.IsVisible() {
		return
	}
	m := metrics.Toggle
	th := t.theme
	tok := th.Active()
	bounds := t.Bounds()
	radius := th.RadiusMD()
	on := t.isOn()
	hovered := (t.hovered || t.pressed) && !t.disabled
	outline := t.variant == VariantOutline

	// Foreground default = --foreground; hover (off) → muted-foreground;
	// on or outline-hover → accent-foreground.
	fg := tok.Foreground
	var fill widget.Color
	hasFill := false

	switch {
	case on:
		fill, hasFill = tok.Accent, true
		fg = tok.AccentForeground
	case hovered && outline:
		fill, hasFill = tok.Accent, true
		fg = tok.AccentForeground
	case hovered:
		fill, hasFill = tok.Muted, true
		fg = tok.MutedForeground
	}

	// shadow-xs on the outline variant.
	if outline {
		drawButtonShadow(canvas, bounds, radius, t.disabled)
	}
	if hasFill && fill.A > 0 {
		canvas.DrawRoundRect(bounds, draw.Fade(fill, t.disabled), radius)
	}

	fg = draw.Fade(fg, t.disabled)

	// Content layout: icon + gap + label, centered (justify-center).
	content := t.contentWidth()
	x := bounds.Min.X + (bounds.Width()-content)/2
	if t.icon != nil {
		iconRect := geometry.NewRect(
			x,
			bounds.Min.Y+(bounds.Height()-m.IconSize)/2,
			m.IconSize, m.IconSize)
		icon.Draw(canvas, *t.icon, iconRect, fg)
		x += m.IconSize
		if t.label != "" {
			x += m.Gap
		}
	}
	if t.label != "" {
		labelW := textmetrics.Width(t.fontFamily(), m.FontSize, t.label)
		labelRect := geometry.NewRect(x, bounds.Min.Y, labelW, bounds.Height())
		family := t.fontFamily()
		if std, ok := canvas.(widget.StyledTextDrawer); ok {
			std.DrawStyledText(t.label, labelRect, widget.TextStyle{
				FontFamily: family,
				FontSize:   m.FontSize,
				Color:      fg,
				Align:      widget.TextAlignLeft,
			})
		} else {
			canvas.DrawText(t.label, labelRect, m.FontSize, fg, m.FontWeight >= 600, widget.TextAlignLeft)
		}
	}

	// Outline border (border-input). When focus-visible the border turns
	// solid --ring (focus-visible:border-ring).
	if outline {
		border := tok.Input
		if t.focusVisible && !t.disabled {
			border = tok.Ring
		}
		draw.InsideBorder(canvas, bounds, radius, draw.Fade(border, t.disabled), m.BorderWidth)
	}

	// Focus ring (focus-visible:ring-[3px] ring-ring/50).
	if t.focusVisible && !t.disabled {
		draw.FocusRing(canvas, bounds, radius, draw.Alpha(tok.Ring, metrics.RingAlpha))
	}
}

// Event handles hover, press/release toggle, and keyboard activation.
func (t *ToggleWidget) Event(ctx widget.Context, e event.Event) bool {
	if t.disabled {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return t.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		if t.IsFocused() && ev.KeyType == event.KeyPress &&
			(ev.Key == event.KeyEnter || ev.Key == event.KeySpace) {
			t.activate(ctx)
			return true
		}
	}
	return false
}

func (t *ToggleWidget) mouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		t.hovered = true
		if ctx != nil {
			ctx.SetCursor(widget.CursorPointer)
			ctx.InvalidateRect(t.Bounds())
		}
		t.SetNeedsRedraw(true)
		return true
	case event.MouseLeave:
		t.hovered = false
		t.pressed = false
		if ctx != nil {
			ctx.SetCursor(widget.CursorDefault)
			ctx.InvalidateRect(t.Bounds())
		}
		t.SetNeedsRedraw(true)
		return true
	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		t.pressed = true
		t.pointerFocus = true
		if ctx != nil {
			ctx.RequestFocus(t)
		}
		t.pointerFocus = false
		t.SetNeedsRedraw(true)
		return true
	case event.MouseRelease:
		wasPressed := t.pressed
		t.pressed = false
		t.SetNeedsRedraw(true)
		if wasPressed && t.Bounds().Contains(e.Position) {
			t.activate(ctx)
		}
		return true
	}
	return false
}

// activate flips the on state and requests a repaint.
func (t *ToggleWidget) activate(ctx widget.Context) {
	t.setOn(!t.isOn())
	t.SetNeedsRedraw(true)
	if ctx != nil {
		ctx.InvalidateRect(t.Bounds())
	}
}

// SetFocused tracks focus-visible: the ring renders only when focus did not
// arrive from a mouse press.
func (t *ToggleWidget) SetFocused(focused bool) {
	t.focusVisible = focused && !t.pointerFocus
	t.WidgetBase.SetFocused(focused)
	t.MarkRedrawLocal() // not SetNeedsRedraw: avoids context-lock re-entry in RequestFocus
}

// IsFocusable reports whether the toggle can take keyboard focus.
func (t *ToggleWidget) IsFocusable() bool {
	return t.IsVisible() && t.IsEnabled() && !t.disabled
}

// Mount binds the controlled-on signal for push invalidation.
func (t *ToggleWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil || t.onSig == nil {
		return
	}
	t.AddBinding(state.BindToScheduler(t.onSig, t, sched))
}

// Unmount implements widget.Lifecycle; bindings clean up automatically.
func (t *ToggleWidget) Unmount() {}

// Children returns nil; ToggleWidget is a leaf.
func (t *ToggleWidget) Children() []widget.Widget { return nil }

// AccessibilityRole returns the button role (Toggle is a pressable button).
func (t *ToggleWidget) AccessibilityRole() a11y.Role { return a11y.RoleButton }

// AccessibilityLabel returns the toggle label.
func (t *ToggleWidget) AccessibilityLabel() string { return t.label }

// AccessibilityHint returns no hint.
func (t *ToggleWidget) AccessibilityHint() string { return "" }

// AccessibilityValue reports the on/off state textually.
func (t *ToggleWidget) AccessibilityValue() string {
	if t.isOn() {
		return "on"
	}
	return "off"
}

// AccessibilityState reports the disabled, focused, and selected (on) states.
func (t *ToggleWidget) AccessibilityState() a11y.State {
	return a11y.State{Disabled: t.disabled, Focused: t.IsFocused(), Selected: t.isOn()}
}

// AccessibilityActions returns the click and focus actions.
func (t *ToggleWidget) AccessibilityActions() []a11y.Action {
	return []a11y.Action{a11y.ActionClick, a11y.ActionFocus}
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*ToggleWidget)(nil)
	_ widget.Focusable = (*ToggleWidget)(nil)
	_ widget.Lifecycle = (*ToggleWidget)(nil)
	_ a11y.Accessible  = (*ToggleWidget)(nil)
)

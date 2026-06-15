package graft

import (
	"regexp"
	"strings"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// InputOTPWidget is shadcn's InputOTP: a one-time password input rendered as a
// row of N separate slots grouped with optional dash separators between groups.
//
// OWNED widget (DESIGN.md S3.2): drawn directly via internal/draw + metrics +
// theme tokens. The whole row is one focusable leaf; keyboard input fills slots
// left-to-right. Backspace removes the last filled character. OnComplete fires
// when all slots are filled.
//
// Usage:
//
//	InputOTP().Length(6).Groups(3, 3).OnComplete(func(code string) { ... })
//	InputOTP().Length(4).Pattern("[0-9]")  // digits only
type InputOTPWidget struct {
	widget.WidgetBase

	theme *theme.Theme

	length  int    // total number of slots (default 6)
	groups  []int  // slot counts per group, e.g. [3, 3]
	pattern string // regexp pattern each character must match

	value    string // the accumulated OTP characters (uncontrolled mode)
	valueSig state.Signal[string]

	onComplete func(string)
	onChange   func(string)

	disabled bool
	hovered  bool

	focusVisible bool

	compiledPattern *regexp.Regexp
}

// InputOTP creates a 6-slot OTP input snapshotting the current theme.
func InputOTP() *InputOTPWidget {
	o := &InputOTPWidget{
		theme:  CurrentTheme(),
		length: 6,
		groups: []int{3, 3},
	}
	o.SetVisible(true)
	o.SetEnabled(true)
	return o
}

// Length sets the total number of input slots.
func (o *InputOTPWidget) Length(n int) *InputOTPWidget {
	if n < 1 {
		n = 1
	}
	o.length = n
	// Reset groups to a single group if they no longer match.
	total := 0
	for _, g := range o.groups {
		total += g
	}
	if total != n {
		o.groups = []int{n}
	}
	return o
}

// Groups sets the slot distribution across groups. The sum of all group sizes
// must equal Length; otherwise Groups adjusts Length to match.
func (o *InputOTPWidget) Groups(sizes ...int) *InputOTPWidget {
	if len(sizes) == 0 {
		return o
	}
	total := 0
	for _, s := range sizes {
		if s < 1 {
			s = 1
		}
		total += s
	}
	o.groups = make([]int, len(sizes))
	copy(o.groups, sizes)
	o.length = total
	return o
}

// Pattern sets a regexp pattern that each typed character must match (e.g.
// "[0-9]" for digits only). An empty pattern accepts all printable runes.
func (o *InputOTPWidget) Pattern(re string) *InputOTPWidget {
	o.pattern = re
	if re != "" {
		o.compiledPattern = regexp.MustCompile(re)
	} else {
		o.compiledPattern = nil
	}
	return o
}

// OnComplete registers a callback fired when all slots are filled.
func (o *InputOTPWidget) OnComplete(fn func(string)) *InputOTPWidget {
	o.onComplete = fn
	return o
}

// OnChange registers a callback fired on every value change.
func (o *InputOTPWidget) OnChange(fn func(string)) *InputOTPWidget {
	o.onChange = fn
	return o
}

// Bind makes the input controlled by sig. The binding is registered in Mount.
func (o *InputOTPWidget) Bind(sig state.Signal[string]) *InputOTPWidget {
	o.valueSig = sig
	return o
}

// Disabled sets the disabled state (faded, not focusable, ignores input).
func (o *InputOTPWidget) Disabled(v bool) *InputOTPWidget {
	o.disabled = v
	return o
}

// Theme pins a specific theme instead of the process-wide current theme.
func (o *InputOTPWidget) Theme(th *theme.Theme) *InputOTPWidget {
	if th != nil {
		o.theme = th
	}
	return o
}

// SetValue sets the value directly (for testing and programmatic use).
func (o *InputOTPWidget) SetValue(s string) *InputOTPWidget {
	runes := []rune(s)
	if len(runes) > o.length {
		runes = runes[:o.length]
	}
	v := string(runes)
	o.value = v
	if o.valueSig != nil {
		o.valueSig.Set(v)
	}
	return o
}

// resolvedTheme returns the pinned theme or the process-wide current theme.
func (o *InputOTPWidget) resolvedTheme() *theme.Theme {
	if o.theme != nil {
		return o.theme
	}
	return CurrentTheme()
}

// resolvedValue returns the current OTP string (bound signal wins).
func (o *InputOTPWidget) resolvedValue() string {
	if o.valueSig != nil {
		return o.valueSig.Get()
	}
	return o.value
}

// setValue writes new content, fires callbacks, and checks completion.
func (o *InputOTPWidget) setValue(s string) {
	runes := []rune(s)
	if len(runes) > o.length {
		runes = runes[:o.length]
	}
	v := string(runes)
	if o.valueSig != nil {
		o.valueSig.Set(v)
	} else {
		o.value = v
	}
	if o.onChange != nil {
		o.onChange(v)
	}
	if len(runes) == o.length && o.onComplete != nil {
		o.onComplete(v)
	}
}

// filledCount returns how many slots are filled.
func (o *InputOTPWidget) filledCount() int {
	return len([]rune(o.resolvedValue()))
}

// IsFocusable reports whether the input can receive keyboard focus.
func (o *InputOTPWidget) IsFocusable() bool {
	return o.IsVisible() && o.IsEnabled() && !o.disabled
}

// SetFocused tracks the focused state. Uses MarkRedrawLocal (not
// SetNeedsRedraw) to avoid context-lock re-entry in RequestFocus.
func (o *InputOTPWidget) SetFocused(focused bool) {
	o.WidgetBase.SetFocused(focused)
	o.focusVisible = focused
	o.MarkRedrawLocal()
}

// Children returns nil; the InputOTP is a leaf widget.
func (o *InputOTPWidget) Children() []widget.Widget { return nil }

// Mount binds the value signal so external writes repaint.
func (o *InputOTPWidget) Mount(ctx widget.Context) {
	if sched := ctx.Scheduler(); sched != nil && o.valueSig != nil {
		o.AddBinding(state.BindToScheduler(o.valueSig, o, sched))
	}
}

// Unmount implements widget.Lifecycle; bindings are cleaned automatically.
func (o *InputOTPWidget) Unmount() {}

// totalWidth computes the full row width including all slots, gaps, and
// separators between groups.
func (o *InputOTPWidget) totalWidth() float32 {
	m := metrics.InputOTP
	var w float32
	for gi, gs := range o.groups {
		// Slots in this group.
		w += float32(gs) * m.SlotSize
		// Gaps between slots within the group.
		if gs > 1 {
			w += float32(gs-1) * m.SlotGap
		}
		// Separator between groups (not after the last group).
		if gi < len(o.groups)-1 {
			w += m.GroupGap
		}
	}
	return w
}

// Layout sizes the widget to fit all slots and separators.
func (o *InputOTPWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	m := metrics.InputOTP
	w := o.totalWidth()
	h := m.SlotSize
	size := constraints.Constrain(geometry.Sz(w, h))
	o.SetBounds(geometry.FromPointSize(o.Position(), size))
	return size
}

// Draw renders the slot boxes, filled characters, active caret, separator
// dashes, and focus ring.
func (o *InputOTPWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !o.IsVisible() {
		return
	}
	th := o.resolvedTheme()
	tok := th.Active()
	dark := th.IsDark()
	m := metrics.InputOTP
	bounds := o.Bounds()
	disabled := o.disabled
	val := []rune(o.resolvedValue())
	activeSlot := len(val) // the slot that will receive the next character
	if activeSlot >= o.length {
		activeSlot = o.length - 1
	}

	x := bounds.Min.X
	slotIdx := 0

	for gi, gs := range o.groups {
		for si := 0; si < gs; si++ {
			slotBounds := geometry.NewRect(x, bounds.Min.Y, m.SlotSize, m.SlotSize)

			// Determine corner radius: first slot in group gets left corners,
			// last slot in group gets right corners. Interior slots are square.
			radius := o.slotRadius(si, gs, m.Radius)

			// Shadow-xs under each slot (skip when disabled).
			if !disabled {
				draw.Shadow(canvas, slotBounds, radius, metrics.ShadowXS)
			}

			// Dark mode fill.
			if dark {
				fill := draw.MulAlpha(tok.Input, m.DarkFillAlpha)
				canvas.DrawRoundRect(slotBounds, draw.Fade(fill, disabled), radius)
			}

			// Determine if this is the active (focused) slot.
			isActive := o.focusVisible && slotIdx == activeSlot

			// Focus ring on the active slot.
			borderColor := tok.Input
			if isActive {
				draw.FocusRing(canvas, slotBounds, radius, draw.Alpha(tok.Ring, metrics.RingAlpha))
				borderColor = tok.Ring
			}

			// Border. BorderFill (fill = page background) instead of an
			// inside stroke, which renders as a solid gray box on the GPU.
			draw.BorderFill(canvas, slotBounds, tok.Background, draw.Fade(borderColor, disabled), radius, m.BorderWidth)

			// Character in the slot.
			if slotIdx < len(val) {
				o.drawSlotChar(canvas, slotBounds, string(val[slotIdx]), draw.Fade(tok.Foreground, disabled))
			} else if isActive {
				// Blinking caret in the active empty slot.
				o.drawCaret(canvas, slotBounds, draw.Fade(tok.Foreground, disabled))
			}

			x += m.SlotSize + m.SlotGap
			slotIdx++
		}
		// Remove the trailing SlotGap after the last slot in the group.
		x -= m.SlotGap

		// Draw separator between groups (except after the last group).
		if gi < len(o.groups)-1 {
			sepX := x + (m.GroupGap-m.SepWidth)/2
			sepY := bounds.Min.Y + (m.SlotSize-m.SepHeight)/2
			sepBounds := geometry.NewRect(sepX, sepY, m.SepWidth, m.SepHeight)
			canvas.DrawRect(sepBounds, draw.Fade(tok.Foreground, disabled))
			x += m.GroupGap
		}
	}
}

// slotRadius returns the effective corner radius for a slot at position si
// within a group of gs slots. Only the first and last slots in a group get
// rounded corners; interior slots are square.
func (o *InputOTPWidget) slotRadius(si, gs int, radius float32) float32 {
	if gs == 1 {
		return radius
	}
	if si == 0 || si == gs-1 {
		return radius
	}
	return 0
}

// drawSlotChar draws a single character centered in the slot bounds.
func (o *InputOTPWidget) drawSlotChar(canvas widget.Canvas, bounds geometry.Rect, ch string, col widget.Color) {
	m := metrics.InputOTP
	family := fonts.Family(m.FontWeight)
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(ch, bounds, widget.TextStyle{
			FontFamily: family,
			FontSize:   m.FontSize,
			Color:      col,
			Align:      widget.TextAlignCenter,
		})
		return
	}
	canvas.DrawText(ch, bounds, m.FontSize, col, m.FontWeight >= 600, widget.TextAlignCenter)
}

// drawCaret draws a thin vertical caret line centered in the slot.
func (o *InputOTPWidget) drawCaret(canvas widget.Canvas, bounds geometry.Rect, col widget.Color) {
	m := metrics.InputOTP
	caretH := m.FontSize * 1.2 // slightly taller than the font
	cx := bounds.Min.X + (bounds.Width()-m.CaretWidth)/2
	cy := bounds.Min.Y + (bounds.Height()-caretH)/2
	caretBounds := geometry.NewRect(cx, cy, m.CaretWidth, caretH)
	canvas.DrawRect(caretBounds, col)
}

// Event handles hover, click, and keyboard input.
func (o *InputOTPWidget) Event(ctx widget.Context, e event.Event) bool {
	if o.disabled {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return o.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		if !o.IsFocused() {
			return false
		}
		return o.keyEvent(ctx, ev)
	}
	return false
}

func (o *InputOTPWidget) mouseEvent(ctx widget.Context, ev *event.MouseEvent) bool {
	switch ev.MouseType {
	case event.MouseEnter:
		o.hovered = true
		ctx.SetCursor(widget.CursorText)
		o.SetNeedsRedraw(true)
		ctx.InvalidateRect(o.Bounds())
		return true
	case event.MouseLeave:
		o.hovered = false
		ctx.SetCursor(widget.CursorDefault)
		o.SetNeedsRedraw(true)
		ctx.InvalidateRect(o.Bounds())
		return true
	case event.MousePress:
		if ev.Button != event.ButtonLeft {
			return false
		}
		ctx.RequestFocus(o)
		return true
	case event.MouseRelease:
		return true
	}
	return false
}

func (o *InputOTPWidget) keyEvent(ctx widget.Context, ev *event.KeyEvent) bool {
	if ev.KeyType != event.KeyPress && ev.KeyType != event.KeyRepeat {
		return false
	}

	switch ev.Key {
	case event.KeyBackspace:
		o.backspace(ctx)
		return true
	case event.KeyDelete:
		o.backspace(ctx)
		return true
	}

	// Printable rune input.
	if ev.Rune != 0 && ev.Rune >= 0x20 {
		o.insertRune(ctx, ev.Rune)
		return true
	}
	return false
}

// insertRune appends a rune to the OTP value if it passes the pattern filter
// and there is room.
func (o *InputOTPWidget) insertRune(ctx widget.Context, r rune) {
	val := o.resolvedValue()
	runes := []rune(val)
	if len(runes) >= o.length {
		return
	}
	ch := string(r)
	if o.compiledPattern != nil && !o.compiledPattern.MatchString(ch) {
		return
	}
	o.setValue(val + ch)
	o.invalidate(ctx)
}

// backspace removes the last character from the OTP value.
func (o *InputOTPWidget) backspace(ctx widget.Context) {
	val := o.resolvedValue()
	runes := []rune(val)
	if len(runes) == 0 {
		return
	}
	o.setValue(string(runes[:len(runes)-1]))
	o.invalidate(ctx)
}

func (o *InputOTPWidget) invalidate(ctx widget.Context) {
	o.SetNeedsRedraw(true)
	ctx.InvalidateRect(o.Bounds())
}

// Value returns the current OTP string.
func (o *InputOTPWidget) Value() string {
	return o.resolvedValue()
}

// IsComplete reports whether all slots are filled.
func (o *InputOTPWidget) IsComplete() bool {
	return o.filledCount() >= o.length
}

// Code returns the filled string (alias for Value, convenient for callers
// checking the final result alongside IsComplete).
func (o *InputOTPWidget) Code() string {
	return o.resolvedValue()
}

// String returns a debug representation.
func (o *InputOTPWidget) String() string {
	v := o.resolvedValue()
	filled := len([]rune(v))
	pad := o.length - filled
	if pad < 0 {
		pad = 0
	}
	return "InputOTP[" + v + strings.Repeat("_", pad) + "]"
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*InputOTPWidget)(nil)
	_ widget.Focusable = (*InputOTPWidget)(nil)
	_ widget.Lifecycle = (*InputOTPWidget)(nil)
)

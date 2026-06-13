package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// CheckboxWidget is the shadcn Checkbox: a 16×16 box with a 4px radius,
// optional adjacent label, and check / minus indicators
// (docs/research/03-shadcn-pixel-spec.md §5 "Checkbox").
//
// OWNED widget (DESIGN.md §3.2): drawn directly via internal/draw + metrics +
// theme tokens. The whole box+label row is one clickable leaf with a pointer
// cursor. Space toggles when focused. Colors are resolved from the active
// token set inside Draw so mode switches repaint without rebuilding.
type CheckboxWidget struct {
	widget.WidgetBase

	label string

	checked  bool
	hovered  bool
	disabled bool

	// indeterminate renders the Minus indicator instead of Check; it does not
	// affect the checked value (matching shadcn's mixed state).
	indeterminate    bool
	indeterminateSig state.ReadonlySignal[bool]

	sig      state.Signal[bool]
	onChange func(bool)

	theme *theme.Theme
}

// Checkbox creates an unchecked checkbox snapshotting the current theme.
func Checkbox() *CheckboxWidget {
	c := &CheckboxWidget{theme: CurrentTheme()}
	c.SetVisible(true)
	c.SetEnabled(true)
	return c
}

// Label sets the text rendered to the right of the box (Label style: 14px /
// weight 500 / leading-none), making the whole row clickable.
func (c *CheckboxWidget) Label(text string) *CheckboxWidget {
	c.label = text
	return c
}

// Checked sets the initial (uncontrolled) checked state.
func (c *CheckboxWidget) Checked(v bool) *CheckboxWidget {
	c.checked = v
	return c
}

// Bind makes the checkbox controlled by sig. The binding is registered in
// Mount; toggles write to sig.
func (c *CheckboxWidget) Bind(sig state.Signal[bool]) *CheckboxWidget {
	c.sig = sig
	c.checked = sig.Get()
	return c
}

// Indeterminate drives the mixed (Minus) indicator from a read-only signal.
func (c *CheckboxWidget) Indeterminate(sig state.ReadonlySignal[bool]) *CheckboxWidget {
	c.indeterminateSig = sig
	return c
}

// SetIndeterminate sets the static mixed state (test/builder convenience).
func (c *CheckboxWidget) SetIndeterminate(v bool) *CheckboxWidget {
	c.indeterminate = v
	return c
}

// OnChange registers a callback fired on every toggle.
func (c *CheckboxWidget) OnChange(fn func(bool)) *CheckboxWidget {
	c.onChange = fn
	return c
}

// Disabled sets the disabled state (faded, not focusable, ignores input).
func (c *CheckboxWidget) Disabled(v bool) *CheckboxWidget {
	c.disabled = v
	return c
}

// Theme pins a specific theme instead of the process-wide current theme.
func (c *CheckboxWidget) Theme(th *theme.Theme) *CheckboxWidget {
	c.theme = th
	return c
}

// IsChecked reports the current checked state (signal wins when bound).
func (c *CheckboxWidget) IsChecked() bool {
	if c.sig != nil {
		return c.sig.Get()
	}
	return c.checked
}

// isIndeterminate reports the effective mixed state (signal wins).
func (c *CheckboxWidget) isIndeterminate() bool {
	if c.indeterminateSig != nil {
		return c.indeterminateSig.Get()
	}
	return c.indeterminate
}

func (c *CheckboxWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// IsFocusable reports whether the checkbox can receive keyboard focus.
func (c *CheckboxWidget) IsFocusable() bool {
	return c.IsVisible() && c.IsEnabled() && !c.disabled
}

// Layout sizes the row: 16px box, plus gap + label width when labeled. Height
// is the 16px box (label is vertically centered against it).
func (c *CheckboxWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	w := metrics.Checkbox.Size
	h := metrics.Checkbox.Size
	if c.label != "" {
		family := fonts.Family(metrics.Checkbox.LabelFontWeight)
		labelW := textmetrics.Width(family, metrics.Checkbox.LabelFontSize, c.label)
		w += metrics.Checkbox.LabelGap + labelW
	}
	size := constraints.Constrain(geometry.Sz(w, h))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw renders the box (shadow, fill, border, indicator, focus ring) and the
// optional label.
func (c *CheckboxWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.resolvedTheme()
	tok := th.Active()
	dark := th.IsDark()
	bounds := c.Bounds()
	disabled := c.disabled

	boxSize := metrics.Checkbox.Size
	box := geometry.NewRect(bounds.Min.X, bounds.Min.Y+(bounds.Height()-boxSize)/2, boxSize, boxSize)
	radius := metrics.Checkbox.Radius
	checked := c.IsChecked()
	indeterminate := c.isIndeterminate()
	marked := checked || indeterminate

	// Shadow-xs under the box (skip when disabled to avoid a faded halo).
	if !disabled {
		draw.Shadow(canvas, box, radius, metrics.ShadowXS)
	}

	// Fill.
	switch {
	case marked:
		canvas.DrawRoundRect(box, draw.Fade(tok.Primary, disabled), radius)
	case dark:
		fill := draw.MulAlpha(tok.Input, metrics.Checkbox.DarkFillAlpha)
		canvas.DrawRoundRect(box, draw.Fade(fill, disabled), radius)
	}

	// Focus ring (keyboard focus) before the border, per the shared recipe.
	borderColor := tok.Input
	if marked {
		borderColor = tok.Primary
	}
	if c.IsFocused() {
		draw.FocusRing(canvas, box, radius, draw.Alpha(tok.Ring, metrics.RingAlpha))
		borderColor = tok.Ring
	}
	draw.InsideBorder(canvas, box, radius, draw.Fade(borderColor, disabled), metrics.Checkbox.BorderWidth)

	// Indicator glyph (Check, or Minus when indeterminate). Icons no-op on
	// canvases without SVGRenderer (the mock); goldens render them.
	if marked {
		ic := icons.Check
		if indeterminate {
			ic = icons.Minus
		}
		c.drawIndicator(canvas, box, ic, draw.Fade(tok.PrimaryForeground, disabled))
	}

	// Label.
	if c.label != "" {
		c.drawLabel(canvas, th, bounds, box, draw.Fade(tok.Foreground, disabled))
	}
}

// drawIndicator centers the 14px check/minus glyph inside the box.
func (c *CheckboxWidget) drawIndicator(canvas widget.Canvas, box geometry.Rect, ic icon.IconData, col widget.Color) {
	sz := metrics.Checkbox.IconSize
	center := box.Center()
	iconBounds := geometry.NewRect(center.X-sz/2, center.Y-sz/2, sz, sz)
	icon.Draw(canvas, ic, iconBounds, col)
}

// drawLabel draws the label text vertically centered against the row.
func (c *CheckboxWidget) drawLabel(canvas widget.Canvas, th *theme.Theme, bounds, box geometry.Rect, col widget.Color) {
	family := fonts.Family(metrics.Checkbox.LabelFontWeight)
	size := metrics.Checkbox.LabelFontSize
	x := box.Max.X + metrics.Checkbox.LabelGap
	lineH := metrics.Checkbox.LabelLineHeight
	textRect := geometry.NewRect(x, bounds.Min.Y+(bounds.Height()-lineH)/2, bounds.Max.X-x, lineH)
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(c.label, textRect, widget.TextStyle{
			FontFamily: family,
			FontSize:   size,
			Color:      col,
			Align:      widget.TextAlignLeft,
		})
		return
	}
	canvas.DrawText(c.label, textRect, size, col, metrics.Checkbox.LabelFontWeight >= 600, widget.TextAlignLeft)
}

// toggle flips the checked state, writing to the bound signal and firing
// OnChange. Indeterminate resolves to checked on the next toggle (shadcn:
// clicking a mixed checkbox checks it).
func (c *CheckboxWidget) toggle(ctx widget.Context) {
	next := !c.IsChecked()
	if c.isIndeterminate() {
		next = true
	}
	c.checked = next
	if c.sig != nil {
		c.sig.Set(next)
	}
	if c.onChange != nil {
		c.onChange(next)
	}
	c.SetNeedsRedraw(true)
	ctx.InvalidateRect(c.Bounds())
}

// Event handles hover, click, and Space-to-toggle.
func (c *CheckboxWidget) Event(ctx widget.Context, e event.Event) bool {
	if c.disabled {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return c.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		if c.IsFocused() && ev.KeyType == event.KeyPress && ev.Key == event.KeySpace {
			c.toggle(ctx)
			return true
		}
	}
	return false
}

func (c *CheckboxWidget) mouseEvent(ctx widget.Context, ev *event.MouseEvent) bool {
	switch ev.MouseType {
	case event.MouseEnter:
		c.hovered = true
		ctx.SetCursor(widget.CursorPointer)
		c.SetNeedsRedraw(true)
		ctx.InvalidateRect(c.Bounds())
		return true
	case event.MouseLeave:
		c.hovered = false
		ctx.SetCursor(widget.CursorDefault)
		c.SetNeedsRedraw(true)
		ctx.InvalidateRect(c.Bounds())
		return true
	case event.MousePress:
		if ev.Button != event.ButtonLeft {
			return false
		}
		ctx.RequestFocus(c)
		return true
	case event.MouseRelease:
		if c.Bounds().Contains(ev.Position) {
			c.toggle(ctx)
		}
		return true
	}
	return false
}

// Children returns nil; the checkbox is a leaf.
func (c *CheckboxWidget) Children() []widget.Widget { return nil }

// Mount binds the controlled and indeterminate signals to the scheduler.
func (c *CheckboxWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if c.sig != nil {
		c.AddBinding(state.BindToScheduler(c.sig, c, sched))
	}
	if c.indeterminateSig != nil {
		c.AddBinding(state.BindToScheduler(c.indeterminateSig, c, sched))
	}
}

// Unmount is a no-op; bindings are cleaned up by WidgetBase.
func (c *CheckboxWidget) Unmount() {}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*CheckboxWidget)(nil)
	_ widget.Focusable = (*CheckboxWidget)(nil)
	_ widget.Lifecycle = (*CheckboxWidget)(nil)
)

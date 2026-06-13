package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// radioGroupCtl is the shared selection controller wired into every item of a
// RadioGroup. Items read Selected() to render their dot and call Select() on
// click; the group reads/writes it for keyboard navigation.
type radioGroupCtl struct {
	value    string
	sig      state.Signal[string]
	onChange func(string)
	items    []*RadioGroupItemWidget
}

func (g *radioGroupCtl) selected() string {
	if g.sig != nil {
		return g.sig.Get()
	}
	return g.value
}

func (g *radioGroupCtl) selectValue(v string) {
	g.value = v
	if g.sig != nil {
		g.sig.Set(v)
	}
	if g.onChange != nil {
		g.onChange(v)
	}
}

// RadioGroupWidget is the shadcn RadioGroup: a vertical (or horizontal) stack
// of radio items with single-selection and arrow-key navigation
// (docs/research/03-shadcn-pixel-spec.md §5 "Radio Group").
//
// OWNED widget (DESIGN.md §3.2): the group embeds a primitives.BoxWidget for
// layout, draw, and mouse dispatch (gap-3 = 12px), and adds arrow-key
// navigation in Event. Items are graft-owned RadioGroupItemWidget leaves that
// draw the circle + dot + label via internal/draw + metrics.
type RadioGroupWidget struct {
	*primitives.BoxWidget

	ctl   *radioGroupCtl
	theme *theme.Theme
}

// RadioGroup creates a vertical radio group from the given items.
func RadioGroup(items ...*RadioGroupItemWidget) *RadioGroupWidget {
	ctl := &radioGroupCtl{items: items}
	children := make([]widget.Widget, len(items))
	for idx, it := range items {
		it.ctl = ctl
		children[idx] = it
	}
	g := &RadioGroupWidget{
		BoxWidget: primitives.VBox(children...).Gap(metrics.RadioGroup.GroupGap),
		ctl:       ctl,
		theme:     CurrentTheme(),
	}
	g.applyTheme()
	return g
}

// Horizontal lays the items out in a row instead of a column.
func (g *RadioGroupWidget) Horizontal() *RadioGroupWidget {
	g.SetDirection(primitives.DirectionHorizontal)
	return g
}

// Value sets the initial (uncontrolled) selected value.
func (g *RadioGroupWidget) Value(v string) *RadioGroupWidget {
	g.ctl.value = v
	return g
}

// Bind makes the group controlled by sig.
func (g *RadioGroupWidget) Bind(sig state.Signal[string]) *RadioGroupWidget {
	g.ctl.sig = sig
	return g
}

// OnChange registers a callback fired when the selection changes.
func (g *RadioGroupWidget) OnChange(fn func(string)) *RadioGroupWidget {
	g.ctl.onChange = fn
	return g
}

// Theme pins a specific theme for the group and all its items.
func (g *RadioGroupWidget) Theme(th *theme.Theme) *RadioGroupWidget {
	g.theme = th
	g.applyTheme()
	return g
}

// Selected returns the currently selected value.
func (g *RadioGroupWidget) Selected() string { return g.ctl.selected() }

func (g *RadioGroupWidget) applyTheme() {
	for _, it := range g.ctl.items {
		it.theme = g.theme
	}
}

// Event intercepts arrow keys for selection movement when an item is focused,
// then delegates everything else to the embedded box (mouse dispatch).
func (g *RadioGroupWidget) Event(ctx widget.Context, e event.Event) bool {
	if ke, ok := e.(*event.KeyEvent); ok && ke.KeyType == event.KeyPress {
		switch ke.Key {
		case event.KeyDown, event.KeyRight:
			if g.moveSelection(ctx, +1) {
				return true
			}
		case event.KeyUp, event.KeyLeft:
			if g.moveSelection(ctx, -1) {
				return true
			}
		}
	}
	return g.BoxWidget.Event(ctx, e)
}

// moveSelection advances the selection to the next/previous item relative to
// the focused (or selected) one, wrapping around, and focuses it.
func (g *RadioGroupWidget) moveSelection(ctx widget.Context, dir int) bool {
	items := g.ctl.items
	n := len(items)
	if n == 0 {
		return false
	}
	cur := -1
	for idx, it := range items {
		if it.IsFocused() {
			cur = idx
			break
		}
	}
	if cur == -1 {
		// Fall back to the selected item.
		sel := g.ctl.selected()
		for idx, it := range items {
			if it.value == sel {
				cur = idx
				break
			}
		}
	}
	if cur == -1 {
		cur = 0
	}
	next := (cur + dir + n) % n
	target := items[next]
	g.ctl.selectValue(target.value)
	// Clear focus on every other item so exactly one item is focused after a
	// move (the framework focus manager normally does this, but keep the
	// invariant locally so navigation is deterministic in any host).
	for _, it := range items {
		if it != target {
			it.SetFocused(false)
		}
	}
	ctx.RequestFocus(target)
	g.SetNeedsRedraw(true)
	ctx.InvalidateRect(g.Bounds())
	return true
}

// RadioGroupItemWidget is one selectable radio item: a 16px circle with a 1px
// Input border, an 8px Primary dot when selected, and a 14px/500 label.
type RadioGroupItemWidget struct {
	widget.WidgetBase

	value string
	label string

	ctl     *radioGroupCtl
	hovered bool
	theme   *theme.Theme
}

// RadioGroupItem creates a radio item with the given value and label.
func RadioGroupItem(value, label string) *RadioGroupItemWidget {
	it := &RadioGroupItemWidget{value: value, label: label, theme: CurrentTheme()}
	it.SetVisible(true)
	it.SetEnabled(true)
	return it
}

// Disabled sets the disabled state for this item.
func (it *RadioGroupItemWidget) Disabled(v bool) *RadioGroupItemWidget {
	it.SetEnabled(!v)
	return it
}

func (it *RadioGroupItemWidget) resolvedTheme() *theme.Theme {
	if it.theme != nil {
		return it.theme
	}
	return CurrentTheme()
}

func (it *RadioGroupItemWidget) isSelected() bool {
	return it.ctl != nil && it.ctl.selected() == it.value
}

// IsFocusable reports whether the item can receive keyboard focus.
func (it *RadioGroupItemWidget) IsFocusable() bool {
	return it.IsVisible() && it.IsEnabled()
}

// Layout sizes the item: 16px circle + gap + label width; height is the
// circle (label vertically centered).
func (it *RadioGroupItemWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	w := metrics.RadioGroup.Size
	h := metrics.RadioGroup.Size
	if it.label != "" {
		family := fonts.Family(metrics.RadioGroup.LabelFontWeight)
		w += metrics.RadioGroup.LabelGap + textmetrics.Width(family, metrics.RadioGroup.LabelFontSize, it.label)
	}
	size := c.Constrain(geometry.Sz(w, h))
	it.SetBounds(geometry.FromPointSize(it.Position(), size))
	return size
}

// Draw renders the circle (shadow, fill, border, dot, focus ring) and label.
func (it *RadioGroupItemWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !it.IsVisible() {
		return
	}
	th := it.resolvedTheme()
	tok := th.Active()
	dark := th.IsDark()
	bounds := it.Bounds()
	disabled := !it.IsEnabled()

	d := metrics.RadioGroup.Size
	circleRect := geometry.NewRect(bounds.Min.X, bounds.Min.Y+(bounds.Height()-d)/2, d, d)
	center := circleRect.Center()
	r := d / 2
	selected := it.isSelected()

	// Shadow-xs under the circle.
	if !disabled {
		draw.Shadow(canvas, circleRect, metrics.RadioGroup.Size, metrics.ShadowXS)
	}

	// Dark unselected fill (Input/30) inside the circle.
	if dark && !selected {
		fill := draw.MulAlpha(tok.Input, metrics.RadioGroup.DarkFillAlpha)
		canvas.DrawCircle(center, r-metrics.RadioGroup.BorderWidth/2, draw.Fade(fill, disabled))
	}

	// Focus ring (circular) before the border.
	borderColor := tok.Input
	if it.IsFocused() {
		ring := draw.Alpha(tok.Ring, metrics.RingAlpha)
		canvas.StrokeCircle(center, r+metrics.RingWidth/2, ring, metrics.RingWidth)
		borderColor = tok.Ring
	}

	// 1px Input border (inside: radius r − w/2).
	bw := metrics.RadioGroup.BorderWidth
	canvas.StrokeCircle(center, r-bw/2, draw.Fade(borderColor, disabled), bw)

	// Selected dot (8px Primary).
	if selected {
		canvas.DrawCircle(center, metrics.RadioGroup.DotSize/2, draw.Fade(tok.Primary, disabled))
	}

	// Label.
	if it.label != "" {
		it.drawLabel(canvas, bounds, circleRect, draw.Fade(tok.Foreground, disabled))
	}
}

func (it *RadioGroupItemWidget) drawLabel(canvas widget.Canvas, bounds, circleRect geometry.Rect, col widget.Color) {
	family := fonts.Family(metrics.RadioGroup.LabelFontWeight)
	size := metrics.RadioGroup.LabelFontSize
	lineH := metrics.RadioGroup.LabelLineHeight
	x := circleRect.Max.X + metrics.RadioGroup.LabelGap
	textRect := geometry.NewRect(x, bounds.Min.Y+(bounds.Height()-lineH)/2, bounds.Max.X-x, lineH)
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(it.label, textRect, widget.TextStyle{
			FontFamily: family,
			FontSize:   size,
			Color:      col,
			Align:      widget.TextAlignLeft,
		})
		return
	}
	canvas.DrawText(it.label, textRect, size, col, false, widget.TextAlignLeft)
}

// Event handles hover and click selection (Space also selects when focused).
func (it *RadioGroupItemWidget) Event(ctx widget.Context, e event.Event) bool {
	if !it.IsEnabled() {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return it.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		if it.IsFocused() && ev.KeyType == event.KeyPress && ev.Key == event.KeySpace {
			it.choose(ctx)
			return true
		}
	}
	return false
}

func (it *RadioGroupItemWidget) mouseEvent(ctx widget.Context, ev *event.MouseEvent) bool {
	switch ev.MouseType {
	case event.MouseEnter:
		it.hovered = true
		ctx.SetCursor(widget.CursorPointer)
		it.SetNeedsRedraw(true)
		ctx.InvalidateRect(it.Bounds())
		return true
	case event.MouseLeave:
		it.hovered = false
		ctx.SetCursor(widget.CursorDefault)
		it.SetNeedsRedraw(true)
		ctx.InvalidateRect(it.Bounds())
		return true
	case event.MousePress:
		if ev.Button != event.ButtonLeft {
			return false
		}
		ctx.RequestFocus(it)
		return true
	case event.MouseRelease:
		if it.Bounds().Contains(ev.Position) {
			it.choose(ctx)
		}
		return true
	}
	return false
}

func (it *RadioGroupItemWidget) choose(ctx widget.Context) {
	if it.ctl != nil {
		it.ctl.selectValue(it.value)
	}
	it.SetNeedsRedraw(true)
	ctx.InvalidateRect(it.Bounds())
}

// Children returns nil; the item is a leaf.
func (it *RadioGroupItemWidget) Children() []widget.Widget { return nil }

// Mount binds the group's signal so external changes repaint the items.
func (g *RadioGroupWidget) Mount(ctx widget.Context) {
	g.BoxWidget.Mount(ctx)
	if sched := ctx.Scheduler(); sched != nil && g.ctl.sig != nil {
		g.AddBinding(state.BindToScheduler(g.ctl.sig, g, sched))
	}
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*RadioGroupWidget)(nil)
	_ widget.Lifecycle = (*RadioGroupWidget)(nil)
	_ widget.Widget    = (*RadioGroupItemWidget)(nil)
	_ widget.Focusable = (*RadioGroupItemWidget)(nil)
)

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

// ToggleGroupWidget is the shadcn ToggleGroup: a fused segmented control of
// ToggleGroupItem buttons sharing one selection (single or multiple).
//
// Architecture decision: graft-OWNED. shadcn renders a w-fit flex of toggles
// at spacing 0 — items butt together, share 1px borders (border-l-0 on
// non-first items), and only the group's outer corners are rounded
// (first:rounded-l-md, last:rounded-r-md). That per-corner-radius fusion is
// exactly what draw.SquareCorners (DESIGN.md §5.4) exists for; no core widget
// provides it, so the group and items are drawn here on internal/draw +
// metrics/togglegroup.go.
//
//	graft.ToggleGroup(
//	    graft.ToggleGroupItem("bold", "B"),
//	    graft.ToggleGroupItem("italic", "I"),
//	    graft.ToggleGroupItem("underline", "U"),
//	).Type("single").Outline().Value("bold")
type ToggleGroupWidget struct {
	widget.WidgetBase

	st    *toggleGroupState
	items []*ToggleGroupItemWidget
}

// toggleGroupState is the selection state shared by the group and its items.
type toggleGroupState struct {
	typ      accordionType // reuse single/multiple enum
	outline  bool
	disabled bool
	value    map[string]bool
	sig      state.Signal[string] // single-mode controlled value
	th       *theme.Theme

	invalidate func()
}

// isOn reports whether the item with the given value is selected.
func (st *toggleGroupState) isOn(value string) bool {
	if st.sig != nil && st.typ == accordionSingle {
		return st.sig.Get() == value && value != ""
	}
	return st.value[value]
}

// toggle flips the item's selection, honoring single-mode exclusivity.
func (st *toggleGroupState) toggle(value string) {
	turningOn := !st.isOn(value)
	if st.typ == accordionSingle {
		if st.sig != nil {
			if turningOn {
				st.sig.Set(value)
			} else {
				st.sig.Set("")
			}
		} else {
			for k := range st.value {
				delete(st.value, k)
			}
			if turningOn {
				st.value[value] = true
			}
		}
	} else {
		if turningOn {
			st.value[value] = true
		} else {
			delete(st.value, value)
		}
	}
	if st.invalidate != nil {
		st.invalidate()
	}
}

// ToggleGroup assembles a fused toggle group from ToggleGroupItem children.
// Defaults to type="single" with nothing selected.
func ToggleGroup(items ...*ToggleGroupItemWidget) *ToggleGroupWidget {
	st := &toggleGroupState{
		typ:   accordionSingle,
		value: make(map[string]bool),
		th:    CurrentTheme(),
	}
	g := &ToggleGroupWidget{st: st, items: items}
	g.SetVisible(true)
	g.SetEnabled(true)
	for i, it := range items {
		it.st = st
		it.first = i == 0
		it.last = i == len(items)-1
		it.SetParent(g)
	}
	return g
}

// Type sets the selection mode ("single" or "multiple").
func (g *ToggleGroupWidget) Type(t string) *ToggleGroupWidget {
	if t == "multiple" {
		g.st.typ = accordionMultiple
	} else {
		g.st.typ = accordionSingle
	}
	return g
}

// Single sets type="single".
func (g *ToggleGroupWidget) Single() *ToggleGroupWidget { g.st.typ = accordionSingle; return g }

// Multiple sets type="multiple".
func (g *ToggleGroupWidget) Multiple() *ToggleGroupWidget { g.st.typ = accordionMultiple; return g }

// Outline selects the outline variant (1px shared --input borders +
// group shadow-xs).
func (g *ToggleGroupWidget) Outline() *ToggleGroupWidget { g.st.outline = true; return g }

// Disabled disables every item.
func (g *ToggleGroupWidget) Disabled(v bool) *ToggleGroupWidget { g.st.disabled = v; return g }

// Value sets the initially selected value(s) (uncontrolled). In single mode
// only the first is used.
func (g *ToggleGroupWidget) Value(values ...string) *ToggleGroupWidget {
	for k := range g.st.value {
		delete(g.st.value, k)
	}
	for _, v := range values {
		if v == "" {
			continue
		}
		g.st.value[v] = true
		if g.st.typ == accordionSingle {
			break
		}
	}
	return g
}

// Bind makes the selection controlled by a signal (single mode; the signal
// holds the selected value, "" for none).
func (g *ToggleGroupWidget) Bind(sig state.Signal[string]) *ToggleGroupWidget {
	g.st.sig = sig
	return g
}

// Theme pins a specific theme instead of the snapshotted current theme.
func (g *ToggleGroupWidget) Theme(th *theme.Theme) *ToggleGroupWidget {
	if th != nil {
		g.st.th = th
	}
	return g
}

// Layout lays the items out fused: each shares the group height; items after
// the first overlap 1px so adjacent borders collapse into one.
func (g *ToggleGroupWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	g.st.invalidate = func() {
		g.SetNeedsRedraw(true)
		if ctx != nil {
			ctx.InvalidateRect(g.Bounds())
		}
	}

	h := metrics.Toggle.Default.Height
	loose := cons.Loosen()
	var x float32
	for i, it := range g.items {
		if i > 0 {
			x -= metrics.ToggleGroup.Overlap
		}
		sz := it.Layout(ctx, loose)
		it.SetBounds(geometry.NewRect(x, 0, sz.Width, h))
		x += sz.Width
	}

	size := cons.Constrain(geometry.Sz(x, h))
	g.SetBounds(geometry.FromPointSize(g.Position(), size))
	return size
}

// Draw paints the group shadow (outline variant), then each item.
func (g *ToggleGroupWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !g.IsVisible() {
		return
	}
	th := g.st.th
	bounds := g.Bounds()

	// shadow-xs on the whole outline group.
	if g.st.outline {
		drawButtonShadow(canvas, bounds, th.RadiusMD(), g.st.disabled)
	}

	canvas.PushTransform(bounds.Min)
	for _, it := range g.items {
		widget.StampScreenOrigin(it, canvas)
		widget.DrawChild(it, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event dispatches to the items.
func (g *ToggleGroupWidget) Event(ctx widget.Context, e event.Event) bool {
	if !g.IsVisible() || !g.IsEnabled() || g.st.disabled {
		return false
	}
	children := make([]widget.Widget, len(g.items))
	for i, it := range g.items {
		children[i] = it
	}
	return dispatchToChildren(ctx, e, g.Bounds(), children)
}

// Children returns the items.
func (g *ToggleGroupWidget) Children() []widget.Widget {
	out := make([]widget.Widget, len(g.items))
	for i, it := range g.items {
		out[i] = it
	}
	return out
}

// Mount binds the controlled-value signal for push invalidation.
func (g *ToggleGroupWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil || g.st.sig == nil {
		return
	}
	g.AddBinding(state.BindToScheduler(g.st.sig, g, sched))
}

// Unmount implements widget.Lifecycle; bindings clean up automatically.
func (g *ToggleGroupWidget) Unmount() {}

// ── ToggleGroupItem ──────────────────────────────────────────────────────

// ToggleGroupItemWidget is one segment of a ToggleGroup.
type ToggleGroupItemWidget struct {
	widget.WidgetBase

	st    *toggleGroupState
	value string
	label string
	icon  *icon.IconData
	first bool
	last  bool

	hovered      bool
	mouseDown    bool
	focusVisible bool
}

// ToggleGroupItem creates a group segment with the given value and label.
func ToggleGroupItem(value, label string) *ToggleGroupItemWidget {
	it := &ToggleGroupItemWidget{value: value, label: label}
	it.SetVisible(true)
	it.SetEnabled(true)
	return it
}

// Icon adds a leading 16px icon to the segment.
func (it *ToggleGroupItemWidget) Icon(ic icon.IconData) *ToggleGroupItemWidget {
	it.icon = &ic
	return it
}

// fontFamily resolves the label family for the item weight.
func (it *ToggleGroupItemWidget) fontFamily() string {
	th := it.st.th
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(metrics.Toggle.FontWeight)
}

// contentWidth measures the segment's natural content width.
func (it *ToggleGroupItemWidget) contentWidth() float32 {
	m := metrics.Toggle
	var w float32
	if it.label != "" {
		w += textmetrics.Width(it.fontFamily(), m.FontSize, it.label)
	}
	if it.icon != nil {
		w += m.IconSize
		if it.label != "" {
			w += m.Gap
		}
	}
	return w
}

// Layout sizes the segment: content + 2*px-3 (min-w-0 — no minimum), fixed
// group height.
func (it *ToggleGroupItemWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	w := it.contentWidth() + 2*metrics.ToggleGroup.ItemPadX
	h := metrics.Toggle.Default.Height
	size := c.Constrain(geometry.Sz(w, h))
	it.SetBounds(geometry.FromPointSize(it.Position(), size))
	return size
}

// on reports whether this segment is selected.
func (it *ToggleGroupItemWidget) on() bool { return it.st.isOn(it.value) }

// cornerMask returns the rounded-corner set for this segment: only the
// group's outer corners are rounded (first → left corners, last → right).
func (it *ToggleGroupItemWidget) cornerMask() draw.Corners {
	var corners draw.Corners
	if it.first {
		corners |= draw.TopLeft | draw.BottomLeft
	}
	if it.last {
		corners |= draw.TopRight | draw.BottomRight
	}
	return corners
}

// Draw paints the segment fill (rounded only at the outer corners), shared
// border (outline), icon, label, and focus ring.
func (it *ToggleGroupItemWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !it.IsVisible() {
		return
	}
	m := metrics.Toggle
	th := it.st.th
	tok := th.Active()
	bounds := it.Bounds()
	radius := th.RadiusMD()
	disabled := it.st.disabled
	on := it.on()
	hovered := (it.hovered || it.mouseDown) && !disabled
	outline := it.st.outline
	corners := it.cornerMask()

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

	// Fill: only the group's outer corners are rounded. draw.SquareCorners
	// fills the round-rect then overpaints the inner (non-outer) corners
	// square. Its corner mask names the corners to SQUARE OFF, so pass the
	// complement of this segment's rounded corners.
	if hasFill && fill.A > 0 {
		square := draw.AllCorners &^ corners
		draw.SquareCorners(canvas, bounds, radius, draw.Fade(fill, disabled), square)
	}

	fg = draw.Fade(fg, disabled)

	// Content centered.
	content := it.contentWidth()
	x := bounds.Min.X + (bounds.Width()-content)/2
	if it.icon != nil {
		iconRect := geometry.NewRect(x, bounds.Min.Y+(bounds.Height()-m.IconSize)/2, m.IconSize, m.IconSize)
		icon.Draw(canvas, *it.icon, iconRect, fg)
		x += m.IconSize
		if it.label != "" {
			x += m.Gap
		}
	}
	if it.label != "" {
		labelW := textmetrics.Width(it.fontFamily(), m.FontSize, it.label)
		labelRect := geometry.NewRect(x, bounds.Min.Y, labelW, bounds.Height())
		family := it.fontFamily()
		if std, ok := canvas.(widget.StyledTextDrawer); ok {
			std.DrawStyledText(it.label, labelRect, widget.TextStyle{
				FontFamily: family,
				FontSize:   m.FontSize,
				Color:      fg,
				Align:      widget.TextAlignLeft,
			})
		} else {
			canvas.DrawText(it.label, labelRect, m.FontSize, fg, m.FontWeight >= 600, widget.TextAlignLeft)
		}
	}

	// Outline shared border. The full per-corner border is drawn for the
	// first item; subsequent items omit their left edge (border-l-0) since
	// they overlap the previous item's right border.
	if outline {
		border := draw.Fade(tok.Input, disabled)
		drawSegmentBorder(canvas, bounds, radius, border, metrics.ToggleGroup.BorderWidth, it.first, it.last)
	}

	// Focus ring on the segment (focus:z-10 keeps it above neighbors).
	if it.focusVisible && !disabled {
		draw.FocusRing(canvas, bounds, radius, draw.Alpha(tok.Ring, metrics.RingAlpha))
	}
}

// Event handles hover, click toggle, keyboard activation, and focus-visible.
func (it *ToggleGroupItemWidget) Event(ctx widget.Context, e event.Event) bool {
	if it.st.disabled {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return it.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		if it.IsFocused() && ev.KeyType == event.KeyPress &&
			(ev.Key == event.KeyEnter || ev.Key == event.KeySpace) {
			it.st.toggle(it.value)
			return true
		}
	case *event.FocusEvent:
		it.SetFocused(ev.FocusType == event.FocusGained)
		return false
	}
	return false
}

func (it *ToggleGroupItemWidget) mouseEvent(ctx widget.Context, ev *event.MouseEvent) bool {
	// The group dispatches with group-local coordinates.
	inside := it.Bounds().Contains(ev.Position)
	switch ev.MouseType {
	case event.MouseEnter:
		it.hovered = true
		if ctx != nil {
			ctx.SetCursor(widget.CursorPointer)
			ctx.InvalidateRect(it.Bounds())
		}
		it.SetNeedsRedraw(true)
		return true
	case event.MouseLeave:
		it.hovered = false
		it.mouseDown = false
		if ctx != nil {
			ctx.SetCursor(widget.CursorDefault)
			ctx.InvalidateRect(it.Bounds())
		}
		it.SetNeedsRedraw(true)
		return true
	case event.MousePress:
		if ev.Button != event.ButtonLeft || !inside {
			return false
		}
		it.mouseDown = true
		if ctx != nil {
			ctx.RequestFocus(it)
		}
		it.SetNeedsRedraw(true)
		return true
	case event.MouseRelease:
		wasDown := it.mouseDown
		it.mouseDown = false
		it.SetNeedsRedraw(true)
		if wasDown && inside {
			it.st.toggle(it.value)
			return true
		}
	}
	return false
}

// SetFocused tracks focus-visible: ring only when not from a mouse press.
func (it *ToggleGroupItemWidget) SetFocused(focused bool) {
	it.WidgetBase.SetFocused(focused)
	if focused {
		it.focusVisible = !it.mouseDown
	} else {
		it.focusVisible = false
	}
	it.SetNeedsRedraw(true)
}

// IsFocusable reports keyboard focusability.
func (it *ToggleGroupItemWidget) IsFocusable() bool {
	return it.IsVisible() && it.IsEnabled() && !it.st.disabled
}

// Children returns nil; the item is a leaf.
func (it *ToggleGroupItemWidget) Children() []widget.Widget { return nil }

// AccessibilityRole returns the button role.
func (it *ToggleGroupItemWidget) AccessibilityRole() a11y.Role { return a11y.RoleButton }

// AccessibilityLabel returns the item label.
func (it *ToggleGroupItemWidget) AccessibilityLabel() string { return it.label }

// AccessibilityHint returns no hint.
func (it *ToggleGroupItemWidget) AccessibilityHint() string { return "" }

// AccessibilityValue returns no value text.
func (it *ToggleGroupItemWidget) AccessibilityValue() string { return "" }

// AccessibilityState reports disabled/focused/selected.
func (it *ToggleGroupItemWidget) AccessibilityState() a11y.State {
	return a11y.State{Disabled: it.st.disabled, Focused: it.IsFocused(), Selected: it.on()}
}

// AccessibilityActions returns the click and focus actions.
func (it *ToggleGroupItemWidget) AccessibilityActions() []a11y.Action {
	return []a11y.Action{a11y.ActionClick, a11y.ActionFocus}
}

// drawSegmentBorder strokes the segment's shared border. The first segment
// gets the full per-corner border (rounded at its left, square at its right);
// subsequent segments omit the left edge (border-l-0) and round only the last
// segment's right corners. Because graft's canvas has no per-side stroke, the
// approximation draws a rounded-rect stroke whose left portion is overlapped
// (and so masked) by the previous segment's fill/border. Inner segments draw
// a square-cornered stroke; the first/last add their outer rounding.
func drawSegmentBorder(canvas widget.Canvas, bounds geometry.Rect, radius float32, col widget.Color, w float32, first, last bool) {
	r := radius
	if !first && !last {
		r = 0
	}
	if first && !last {
		// rounded left, square right: stroke full rect at radius; the square
		// right corners are acceptable since the next segment overlaps them.
		r = radius
	}
	// Draw an inside border at the resolved radius. For non-first segments we
	// extend the stroke 1px to the left (beyond bounds) so it sits under the
	// previous segment, leaving a single visible 1px seam.
	left := bounds.Min.X
	if !first {
		left -= metrics.ToggleGroup.Overlap
	}
	rect := geometry.NewRect(left, bounds.Min.Y, bounds.Max.X-left, bounds.Height())
	draw.InsideBorder(canvas, rect, r, col, w)
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*ToggleGroupWidget)(nil)
	_ widget.Lifecycle = (*ToggleGroupWidget)(nil)
	_ widget.Widget    = (*ToggleGroupItemWidget)(nil)
	_ widget.Focusable = (*ToggleGroupItemWidget)(nil)
	_ a11y.Accessible  = (*ToggleGroupItemWidget)(nil)
)

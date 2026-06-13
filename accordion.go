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

// AccordionWidget is the shadcn Accordion: a vertically stacked set of
// expandable items, each a header trigger with a chevron over collapsible
// content, separated by 1px bottom borders.
//
// Architecture decision (DESIGN.md 3.1/3.2): graft-OWNED. shadcn's Accordion
// is a Radix primitive with a precise anatomy — item border-b, py-4 trigger
// with a justify-between right chevron that rotates 180° on open, hover
// underline, pt-0/pb-4 content, accordion-down/up height animation — that no
// core/* gogpu widget expresses (core/collapsible owns a different styled
// header). So the items, triggers, and content are drawn here directly on
// internal/draw + metrics/accordion.go, reusing the collapsible animation
// pattern (settled state rendered in goldens).
//
// The chevron rotation is rendered by swapping chevron-down (closed) for
// chevron-up (open); graft has no rotation transform (Report 1 §7).
//
// Usage mirrors shadcn:
//
//	graft.Accordion(
//	    graft.AccordionItem("a", "Is it accessible?", graft.Text("Yes.")),
//	    graft.AccordionItem("b", "Is it styled?", graft.Text("Yes.")),
//	).Type("single").Value("a")
type AccordionWidget struct {
	widget.WidgetBase

	st    *accordionState
	items []*AccordionItemWidget
}

// accordionType selects single (one open at a time) or multiple expansion.
type accordionType uint8

const (
	accordionSingle   accordionType = iota // type="single"
	accordionMultiple                      // type="multiple"
)

// accordionState is the open-set state shared by the root and its items.
type accordionState struct {
	typ  accordionType
	open map[string]bool
	sig  state.Signal[string] // single-mode controlled value (one value)
	th   *theme.Theme

	invalidate func()
}

// isOpen reports whether the item with the given value is expanded.
func (st *accordionState) isOpen(value string) bool {
	if st.sig != nil && st.typ == accordionSingle {
		return st.sig.Get() == value && value != ""
	}
	return st.open[value]
}

// toggle flips the item's open state, honoring single-mode exclusivity.
func (st *accordionState) toggle(value string) {
	opening := !st.isOpen(value)
	if st.typ == accordionSingle {
		if st.sig != nil {
			if opening {
				st.sig.Set(value)
			} else {
				st.sig.Set("")
			}
		} else {
			for k := range st.open {
				delete(st.open, k)
			}
			if opening {
				st.open[value] = true
			}
		}
	} else {
		if opening {
			st.open[value] = true
		} else {
			delete(st.open, value)
		}
	}
	if st.invalidate != nil {
		st.invalidate()
	}
}

// Accordion assembles an accordion from AccordionItem children. By default it
// is type="single" with no item open; use Type, Value, or Bind to configure.
func Accordion(items ...*AccordionItemWidget) *AccordionWidget {
	st := &accordionState{
		typ:  accordionSingle,
		open: make(map[string]bool),
		th:   CurrentTheme(),
	}
	a := &AccordionWidget{st: st, items: items}
	a.SetVisible(true)
	a.SetEnabled(true)
	for i, it := range items {
		it.st = st
		it.last = i == len(items)-1
		it.SetParent(a)
	}
	return a
}

// Type sets the accordion expansion mode ("single" or "multiple").
func (a *AccordionWidget) Type(t string) *AccordionWidget {
	if t == "multiple" {
		a.st.typ = accordionMultiple
	} else {
		a.st.typ = accordionSingle
	}
	return a
}

// Single sets type="single" (one item open at a time).
func (a *AccordionWidget) Single() *AccordionWidget { a.st.typ = accordionSingle; return a }

// Multiple sets type="multiple" (any number of items open).
func (a *AccordionWidget) Multiple() *AccordionWidget { a.st.typ = accordionMultiple; return a }

// Value sets the initially open item value(s) (uncontrolled). In single mode
// only the first value is used.
func (a *AccordionWidget) Value(values ...string) *AccordionWidget {
	for k := range a.st.open {
		delete(a.st.open, k)
	}
	for _, v := range values {
		if v == "" {
			continue
		}
		a.st.open[v] = true
		if a.st.typ == accordionSingle {
			break
		}
	}
	return a
}

// Bind makes the open item controlled by a signal (single mode only; the
// signal holds the open value, or "" for none).
func (a *AccordionWidget) Bind(sig state.Signal[string]) *AccordionWidget {
	a.st.sig = sig
	return a
}

// Theme pins a specific theme instead of the snapshotted current theme.
func (a *AccordionWidget) Theme(th *theme.Theme) *AccordionWidget {
	if th != nil {
		a.st.th = th
	}
	return a
}

// Layout stacks the items vertically (each full available width).
func (a *AccordionWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	a.st.invalidate = func() {
		a.SetNeedsRedraw(true)
		if ctx != nil {
			ctx.Invalidate()
		}
	}

	w := cons.MaxWidth
	if w >= geometry.Infinity {
		// w-full: fall back to the widest item's natural width.
		w = 0
		for _, it := range a.items {
			if nw := it.naturalWidth(ctx); nw > w {
				w = nw
			}
		}
	}

	tight := geometry.Constraints{MinWidth: w, MaxWidth: w, MinHeight: 0, MaxHeight: geometry.Infinity}
	var y float32
	for _, it := range a.items {
		sz := it.Layout(ctx, tight)
		it.SetBounds(geometry.NewRect(0, y, sz.Width, sz.Height))
		y += sz.Height
	}

	size := cons.Constrain(geometry.Sz(w, y))
	a.SetBounds(geometry.FromPointSize(a.Position(), size))
	return size
}

// Draw paints the items.
func (a *AccordionWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !a.IsVisible() {
		return
	}
	canvas.PushTransform(a.Bounds().Min)
	for _, it := range a.items {
		widget.StampScreenOrigin(it, canvas)
		widget.DrawChild(it, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event dispatches to the items.
func (a *AccordionWidget) Event(ctx widget.Context, e event.Event) bool {
	if !a.IsVisible() || !a.IsEnabled() {
		return false
	}
	children := make([]widget.Widget, len(a.items))
	for i, it := range a.items {
		children[i] = it
	}
	return dispatchToChildren(ctx, e, a.Bounds(), children)
}

// Children returns the items.
func (a *AccordionWidget) Children() []widget.Widget {
	out := make([]widget.Widget, len(a.items))
	for i, it := range a.items {
		out[i] = it
	}
	return out
}

// Mount binds the controlled-value signal for push invalidation.
func (a *AccordionWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil || a.st.sig == nil {
		return
	}
	a.AddBinding(state.BindToScheduler(a.st.sig, a, sched))
}

// Unmount implements widget.Lifecycle; bindings clean up automatically.
func (a *AccordionWidget) Unmount() {}

// ── AccordionItem ────────────────────────────────────────────────────────

// AccordionItemWidget is one expandable accordion entry: a header trigger
// (label + chevron) with a 1px bottom border, over collapsible content.
type AccordionItemWidget struct {
	widget.WidgetBase

	st      *accordionState
	value   string
	label   string
	content Widget
	last    bool

	hovered      bool
	focusVisible bool
	mouseDown    bool

	// Cached layout geometry (in item-local space) from the last Layout.
	triggerH float32
	contentH float32
}

// AccordionItem creates an accordion entry with the given value (identity),
// header label, and content widget.
func AccordionItem(value, label string, content Widget) *AccordionItemWidget {
	it := &AccordionItemWidget{
		value:   value,
		label:   label,
		content: content,
	}
	it.SetVisible(true)
	it.SetEnabled(true)
	if ps, ok := content.(interface{ SetParent(widget.Widget) }); ok {
		ps.SetParent(it)
	}
	return it
}

// fontFamily resolves the trigger font family, honoring custom theme fonts.
func (it *AccordionItemWidget) fontFamily() string {
	th := it.st.th
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(metrics.Accordion.TriggerFontWeight)
}

// open reports whether this item is expanded.
func (it *AccordionItemWidget) open() bool { return it.st.isOpen(it.value) }

// naturalWidth estimates the item's intrinsic width (label + gap + chevron),
// used when the accordion has no bounded width.
func (it *AccordionItemWidget) naturalWidth(_ widget.Context) float32 {
	m := metrics.Accordion
	lw := textmetrics.Width(it.fontFamily(), m.TriggerFontSize, it.label)
	return lw + m.TriggerGap + m.ChevronSize
}

// Layout sizes the item: a fixed-height trigger row plus, when open, the
// content (pt-0 pb-4) below it. The 1px bottom border is drawn inside the
// item's bottom edge.
func (it *AccordionItemWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Accordion
	w := cons.MaxWidth
	if w >= geometry.Infinity {
		w = it.naturalWidth(ctx)
	}

	it.triggerH = m.TriggerLineHeight + 2*m.TriggerPadY
	it.contentH = 0
	total := it.triggerH

	if it.open() {
		// Content area: full item width, child laid out loosely.
		cc := geometry.Constraints{MinWidth: w, MaxWidth: w, MinHeight: 0, MaxHeight: geometry.Infinity}
		csz := it.content.Layout(ctx, cc)
		setChildBounds(it.content, geometry.NewRect(0, it.triggerH, csz.Width, csz.Height))
		it.contentH = csz.Height + m.ContentPadBottom
		total += it.contentH
	} else {
		setChildBounds(it.content, geometry.NewRect(0, it.triggerH, 0, 0))
	}

	size := cons.Constrain(geometry.Sz(w, total))
	it.SetBounds(geometry.FromPointSize(it.Position(), size))
	return size
}

// Draw paints the trigger row (label, chevron, hover underline, focus ring),
// the open content, and the 1px bottom border.
func (it *AccordionItemWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !it.IsVisible() {
		return
	}
	m := metrics.Accordion
	th := it.st.th
	tok := th.Active()
	bounds := it.Bounds()
	open := it.open()

	// Trigger label: left-aligned, vertically centered in the py-4 row,
	// foreground color, font-medium.
	family := it.fontFamily()
	labelW := textmetrics.Width(family, m.TriggerFontSize, it.label)
	labelRect := geometry.NewRect(
		bounds.Min.X,
		bounds.Min.Y+(it.triggerH-m.TriggerLineHeight)/2,
		bounds.Width()-m.TriggerGap-m.ChevronSize,
		m.TriggerLineHeight,
	)
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(it.label, labelRect, widget.TextStyle{
			FontFamily: family,
			FontSize:   m.TriggerFontSize,
			Color:      tok.Foreground,
			Align:      widget.TextAlignLeft,
		})
	} else {
		canvas.DrawText(it.label, labelRect, m.TriggerFontSize, tok.Foreground,
			m.TriggerFontWeight >= 600, widget.TextAlignLeft)
	}

	// Hover underline (hover:underline) under the label.
	if it.hovered {
		ascent, descent, _ := textmetrics.LineHeight(family, m.TriggerFontSize)
		baseline := labelRect.Min.Y + (labelRect.Height()-(ascent+descent))/2 + ascent
		y := baseline + m.UnderlineOffset
		canvas.DrawLine(
			geometry.Pt(labelRect.Min.X, y),
			geometry.Pt(labelRect.Min.X+labelW, y),
			tok.Foreground, m.UnderlineWidth)
	}

	// Chevron at the right edge, aligned to the first text line
	// (items-start) plus translate-y-0.5. Open swaps down→up.
	chev := icons.ChevronDown
	if open {
		chev = icons.ChevronUp
	}
	chevTop := bounds.Min.Y + m.TriggerPadY + (m.TriggerLineHeight-m.ChevronSize)/2 + m.ChevronDropY
	chevRect := geometry.NewRect(
		bounds.Max.X-m.ChevronSize,
		chevTop,
		m.ChevronSize, m.ChevronSize,
	)
	icon.Draw(canvas, chev, chevRect, tok.MutedForeground)

	// Focus ring on the trigger row (focus-visible:border-ring +
	// ring-[3px] ring-ring/50) hugging the row rect.
	if it.focusVisible {
		rowRect := geometry.NewRect(bounds.Min.X, bounds.Min.Y, bounds.Width(), it.triggerH)
		radius := th.RadiusMD() // rounded-md
		draw.InsideBorder(canvas, rowRect, radius, tok.Ring, 1)
		draw.FocusRing(canvas, rowRect, radius, draw.Alpha(tok.Ring, metrics.RingAlpha))
	}

	// Open content.
	if open {
		canvas.PushTransform(bounds.Min)
		widget.StampScreenOrigin(it.content, canvas)
		widget.DrawChild(it.content, ctx, canvas)
		canvas.PopTransform()
	}

	// Bottom border (border-b; last:border-b-0). Drawn as a 1px rule along
	// the item's bottom edge in the --border token.
	if !it.last {
		y := bounds.Max.Y - m.ItemBorderWidth/2
		canvas.DrawLine(
			geometry.Pt(bounds.Min.X, y),
			geometry.Pt(bounds.Max.X, y),
			tok.Border, m.ItemBorderWidth)
	}
}

// Event handles hover, click-to-toggle on the trigger row, keyboard
// activation, and focus-visible tracking.
func (it *AccordionItemWidget) Event(ctx widget.Context, e event.Event) bool {
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

// triggerRow returns the clickable trigger rect in item-local space.
func (it *AccordionItemWidget) triggerRow() geometry.Rect {
	return geometry.NewRect(0, 0, it.Bounds().Width(), it.triggerH)
}

func (it *AccordionItemWidget) mouseEvent(ctx widget.Context, ev *event.MouseEvent) bool {
	// The root dispatches with the position in accordion-local space (not
	// item-local); translate into item-local before hit-testing the trigger.
	local := ev.Position.Sub(it.Bounds().Min)
	inTrigger := it.triggerRow().Contains(local)
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
		if ev.Button != event.ButtonLeft || !inTrigger {
			return false
		}
		it.mouseDown = true
		if ctx != nil {
			ctx.RequestFocus(it)
		}
		return true
	case event.MouseRelease:
		wasDown := it.mouseDown
		it.mouseDown = false
		if wasDown && inTrigger {
			it.st.toggle(it.value)
			return true
		}
	}
	return false
}

// SetFocused tracks focus-visible: the ring renders only when focus did not
// arrive from a mouse press.
func (it *AccordionItemWidget) SetFocused(focused bool) {
	it.WidgetBase.SetFocused(focused)
	if focused {
		it.focusVisible = !it.mouseDown
	} else {
		it.focusVisible = false
	}
	it.MarkRedrawLocal() // not SetNeedsRedraw: avoids context-lock re-entry in RequestFocus
}

// IsFocusable reports keyboard focusability.
func (it *AccordionItemWidget) IsFocusable() bool {
	return it.IsVisible() && it.IsEnabled()
}

// Children returns the content widget.
func (it *AccordionItemWidget) Children() []widget.Widget {
	return []widget.Widget{it.content}
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*AccordionWidget)(nil)
	_ widget.Lifecycle = (*AccordionWidget)(nil)
	_ widget.Widget    = (*AccordionItemWidget)(nil)
	_ widget.Focusable = (*AccordionItemWidget)(nil)
)

package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// Tabs is the shadcn Tabs component.
//
// Architecture decision (DESIGN.md sections 3.1/7, verified against
// core/tabview): Tabs is graft-OWNED. core/tabview's bar model — a
// full-width bar whose tab bounds are computed by the widget, painted as
// one PaintTabBar call — cannot express shadcn's anatomy (w-fit pill list
// with 3px padding, per-trigger borders/shadows/focus rings, the line
// variant's underline hanging 5px below the list), so the trigger/list/
// content widgets are implemented here directly on internal/draw +
// metrics/tabs.go.
//
// Usage mirrors shadcn:
//
//	graft.Tabs(
//	    graft.TabsList(
//	        graft.TabsTrigger("account", "Account"),
//	        graft.TabsTrigger("password", "Password"),
//	    ),
//	    graft.TabsContent("account", accountPanel),
//	    graft.TabsContent("password", passwordPanel),
//	).Value("account")
//
// Orientation defaults to horizontal (list above content, arrow Left/Right
// navigation). Vertical() lays the list out as a left-hand column with the
// content to its right and arrow Up/Down navigation.

// TabsOrientation selects the Tabs axis (Radix data-[orientation]).
type TabsOrientation int

const (
	// TabsHorizontal stacks the list above the content (default) and uses
	// Left/Right arrows to move between triggers.
	TabsHorizontal TabsOrientation = iota
	// TabsVertical places the list as a left-hand column with the content
	// to its right and uses Up/Down arrows to move between triggers.
	TabsVertical
)

type TabsWidget struct {
	widget.WidgetBase

	st    *tabsState
	items []Widget
}

// tabsState is the selection state shared by the list, triggers, and
// contents of one Tabs tree.
type tabsState struct {
	value       string
	sig         state.Signal[string]
	th          *theme.Theme
	variantLine bool
	orientation TabsOrientation
}

// vertical reports whether the tabs use the vertical orientation.
func (st *tabsState) vertical() bool { return st.orientation == TabsVertical }

// current returns the selected tab value.
func (st *tabsState) current() string {
	if st.sig != nil {
		return st.sig.Get()
	}
	return st.value
}

// set updates the selected tab value (writing through to the bound
// signal when controlled).
func (st *tabsState) set(v string) {
	if st.sig != nil {
		st.sig.Set(v)
		return
	}
	st.value = v
}

// Tabs assembles a tabs root from a TabsList and TabsContent children
// (root: "flex flex-col gap-2"). The first enabled trigger is selected by
// default; override with Value or Bind.
func Tabs(children ...Widget) *TabsWidget {
	t := &TabsWidget{
		st:    &tabsState{th: CurrentTheme()},
		items: children,
	}
	t.SetVisible(true)
	t.SetEnabled(true)

	for _, c := range children {
		switch w := c.(type) {
		case *TabsListWidget:
			w.attach(t)
		case *TabsContentWidget:
			w.st = t.st
		}
		if ps, ok := c.(interface{ SetParent(widget.Widget) }); ok {
			ps.SetParent(t)
		}
	}

	// Default selection: first enabled trigger.
	if t.st.value == "" {
		for _, c := range children {
			if l, ok := c.(*TabsListWidget); ok {
				for _, tr := range l.triggers {
					if !tr.disabled {
						t.st.value = tr.value
						break
					}
				}
				break
			}
		}
	}
	return t
}

// Bind makes the selected value controlled by a signal.
func (t *TabsWidget) Bind(sig state.Signal[string]) *TabsWidget {
	t.st.sig = sig
	return t
}

// Value sets the initially selected tab value (uncontrolled).
func (t *TabsWidget) Value(v string) *TabsWidget {
	t.st.value = v
	return t
}

// Line switches the list to the line variant: transparent list, 4px gap,
// and a 2px foreground underline under the active trigger.
func (t *TabsWidget) Line() *TabsWidget {
	t.st.variantLine = true
	return t
}

// Orientation sets the tabs axis (default TabsHorizontal).
func (t *TabsWidget) Orientation(o TabsOrientation) *TabsWidget {
	t.st.orientation = o
	return t
}

// Vertical lays the list out as a left-hand column with the content to its
// right (shorthand for Orientation(TabsVertical)).
func (t *TabsWidget) Vertical() *TabsWidget {
	t.st.orientation = TabsVertical
	return t
}

// Theme pins a specific theme instead of the snapshotted current theme.
func (t *TabsWidget) Theme(th *theme.Theme) *TabsWidget {
	t.st.th = th
	return t
}

// Layout places the list and the active content. Horizontal stacks them
// vertically with the root gap (gap-2 = 8px); vertical places the list on
// the left and the content to its right with the same gap. Inactive
// contents collapse to zero.
func (t *TabsWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if t.st.vertical() {
		return t.layoutVertical(ctx, c)
	}
	loose := c.Loosen()
	var y, maxW float32
	for _, child := range t.items {
		sz := child.Layout(ctx, loose)
		if sz.Height <= 0 && sz.Width <= 0 {
			setChildBounds(child, geometry.NewRect(0, y, 0, 0))
			continue
		}
		if y > 0 {
			y += metrics.Tabs.RootGap
		}
		setChildBounds(child, geometry.FromPointSize(geometry.Pt(0, y), sz))
		y += sz.Height
		if sz.Width > maxW {
			maxW = sz.Width
		}
	}
	size := c.Constrain(geometry.Sz(maxW, y))
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

// layoutVertical places the list column on the left and the active content
// to its right (root: "flex gap-2", content data-[orientation=vertical]).
// The list sizes to its widest trigger; the content claims the remaining
// width.
func (t *TabsWidget) layoutVertical(ctx widget.Context, c geometry.Constraints) geometry.Size {
	loose := c.Loosen()

	// Measure the list first so the content can claim the remaining width.
	var listW, listH float32
	var list *TabsListWidget
	for _, child := range t.items {
		if l, ok := child.(*TabsListWidget); ok {
			list = l
			sz := l.Layout(ctx, loose)
			setChildBounds(l, geometry.FromPointSize(geometry.Pt(0, 0), sz))
			listW, listH = sz.Width, sz.Height
			break
		}
	}

	contentX := listW
	if list != nil {
		contentX += metrics.Tabs.RootGap
	}

	// Content sits to the right of the list, taking the remaining width.
	contentConstraints := loose
	if loose.MaxWidth != geometry.Infinity {
		rem := loose.MaxWidth - contentX
		if rem < 0 {
			rem = 0
		}
		contentConstraints = geometry.BoxConstraints(0, rem, 0, loose.MaxHeight)
	}

	maxX, maxY := listW, listH
	for _, child := range t.items {
		if _, ok := child.(*TabsListWidget); ok {
			continue
		}
		sz := child.Layout(ctx, contentConstraints)
		if sz.Height <= 0 && sz.Width <= 0 {
			setChildBounds(child, geometry.NewRect(contentX, 0, 0, 0))
			continue
		}
		setChildBounds(child, geometry.FromPointSize(geometry.Pt(contentX, 0), sz))
		if r := contentX + sz.Width; r > maxX {
			maxX = r
		}
		if sz.Height > maxY {
			maxY = sz.Height
		}
	}

	size := c.Constrain(geometry.Sz(maxX, maxY))
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

// setChildBounds applies bounds to any widget exposing SetBounds.
func setChildBounds(w widget.Widget, r geometry.Rect) {
	if s, ok := w.(interface{ SetBounds(geometry.Rect) }); ok {
		s.SetBounds(r)
	}
}

// Draw renders the list and active content.
func (t *TabsWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !t.IsVisible() {
		return
	}
	canvas.PushTransform(t.Bounds().Min)
	for _, child := range t.items {
		widget.StampScreenOrigin(child, canvas)
		widget.DrawChild(child, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event dispatches input to the children (mouse positions translated to
// local space, top-most child first).
func (t *TabsWidget) Event(ctx widget.Context, e event.Event) bool {
	if !t.IsVisible() || !t.IsEnabled() {
		return false
	}
	return dispatchToChildren(ctx, e, t.Bounds(), t.items)
}

// dispatchToChildren routes an event to children following the BoxWidget
// convention: mouse/wheel positions are translated into the container's
// local space and offered to containing children in reverse order;
// other events broadcast.
func dispatchToChildren(ctx widget.Context, e event.Event, bounds geometry.Rect, children []widget.Widget) bool {
	switch ev := e.(type) {
	case *event.MouseEvent:
		local := *ev
		local.Position = ev.Position.Sub(bounds.Min)
		for i := len(children) - 1; i >= 0; i-- {
			child := children[i]
			if bw, ok := child.(interface{ Bounds() geometry.Rect }); ok {
				if !bw.Bounds().Contains(local.Position) {
					continue
				}
			}
			if child.Event(ctx, &local) {
				return true
			}
		}
	case *event.WheelEvent:
		local := *ev
		local.Position = ev.Position.Sub(bounds.Min)
		for i := len(children) - 1; i >= 0; i-- {
			child := children[i]
			if bw, ok := child.(interface{ Bounds() geometry.Rect }); ok {
				if !bw.Bounds().Contains(local.Position) {
					continue
				}
			}
			if child.Event(ctx, &local) {
				return true
			}
		}
	default:
		for i := len(children) - 1; i >= 0; i-- {
			if children[i].Event(ctx, e) {
				return true
			}
		}
	}
	return false
}

// Children returns the list and contents.
func (t *TabsWidget) Children() []widget.Widget { return t.items }

// Mount binds the controlled-value signal for push invalidation.
func (t *TabsWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil || t.st.sig == nil {
		return
	}
	t.AddBinding(state.BindToScheduler(t.st.sig, t, sched))
}

// Unmount implements widget.Lifecycle; bindings clean up automatically.
func (t *TabsWidget) Unmount() {}

// ── TabsList ────────────────────────────────────────────────────────────

// TabsListWidget is the trigger row: inline w-fit, h-8, rounded-lg,
// p-[3px], bg-muted (line variant: transparent, gap-1).
type TabsListWidget struct {
	widget.WidgetBase

	st       *tabsState
	root     *TabsWidget
	triggers []*TabsTriggerWidget
}

// TabsList groups tab triggers.
func TabsList(triggers ...*TabsTriggerWidget) *TabsListWidget {
	l := &TabsListWidget{
		st:       &tabsState{th: CurrentTheme()},
		triggers: triggers,
	}
	l.SetVisible(true)
	l.SetEnabled(true)
	for _, tr := range triggers {
		tr.st = l.st
		tr.list = l
		tr.SetParent(l)
	}
	return l
}

// attach wires the list (and its triggers) to the Tabs root's shared state.
func (l *TabsListWidget) attach(root *TabsWidget) {
	l.root = root
	l.st = root.st
	for _, tr := range l.triggers {
		tr.st = root.st
		tr.root = root
	}
}

// Layout sizes the list to fit its triggers (w-fit) at the fixed 36px
// height, placing each trigger inside the 3px padding. Vertical lists drop
// the fixed height, stack the triggers in a column, and stretch each
// trigger to the widest label so the active pill spans the column.
func (l *TabsListWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if l.st.vertical() {
		return l.layoutVertical(ctx, c)
	}
	m := metrics.Tabs
	innerH := m.ListHeight - 2*m.ListPadding
	trigH := innerH - m.TriggerHeightInset
	var gap float32
	if l.st.variantLine {
		gap = m.LineGap
	}

	x := m.ListPadding
	y := m.ListPadding + (innerH-trigH)/2
	for i, tr := range l.triggers {
		if i > 0 {
			x += gap
		}
		sz := tr.Layout(ctx, geometry.Loose(geometry.Sz(geometry.Infinity, trigH)))
		tr.SetBounds(geometry.NewRect(x, y, sz.Width, trigH))
		x += sz.Width
	}

	size := c.Constrain(geometry.Sz(x+m.ListPadding, m.ListHeight))
	l.SetBounds(geometry.FromPointSize(l.Position(), size))
	return size
}

// layoutVertical stacks the triggers in a column. The list width is the
// widest trigger plus the 3px padding on each side; each trigger is
// stretched to that inner width (flex-1) so the active pill spans the
// column. The list height grows with the triggers (h-fit, no h-8).
func (l *TabsListWidget) layoutVertical(ctx widget.Context, c geometry.Constraints) geometry.Size {
	m := metrics.Tabs
	trigH := m.ListHeight - 2*m.ListPadding - m.TriggerHeightInset
	var gap float32
	if l.st.variantLine {
		gap = m.LineGap
	}

	// First pass: measure the widest trigger to size the column.
	var innerW float32
	for _, tr := range l.triggers {
		sz := tr.Layout(ctx, geometry.Loose(geometry.Sz(geometry.Infinity, trigH)))
		if sz.Width > innerW {
			innerW = sz.Width
		}
	}

	// Second pass: place each trigger stretched to the column width.
	x := m.ListPadding
	y := m.ListPadding
	for i, tr := range l.triggers {
		if i > 0 {
			y += gap
		}
		tr.SetBounds(geometry.NewRect(x, y, innerW, trigH))
		y += trigH
	}

	size := c.Constrain(geometry.Sz(innerW+2*m.ListPadding, y+m.ListPadding))
	l.SetBounds(geometry.FromPointSize(l.Position(), size))
	return size
}

// Draw paints the list pill and the triggers.
func (l *TabsListWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !l.IsVisible() {
		return
	}
	bounds := l.Bounds()
	if !l.st.variantLine {
		tok := l.st.th.Active()
		canvas.DrawRoundRect(bounds, tok.Muted, l.st.th.RadiusLG())
	}
	canvas.PushTransform(bounds.Min)
	for _, tr := range l.triggers {
		widget.StampScreenOrigin(tr, canvas)
		widget.DrawChild(tr, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event dispatches to the triggers.
func (l *TabsListWidget) Event(ctx widget.Context, e event.Event) bool {
	if !l.IsVisible() || !l.IsEnabled() {
		return false
	}
	return dispatchToChildren(ctx, e, l.Bounds(), l.Children())
}

// Children returns the triggers.
func (l *TabsListWidget) Children() []widget.Widget {
	out := make([]widget.Widget, len(l.triggers))
	for i, tr := range l.triggers {
		out[i] = tr
	}
	return out
}

// moveFocus moves selection and focus from a trigger by delta, skipping
// disabled triggers and wrapping around (Radix automatic activation).
func (l *TabsListWidget) moveFocus(ctx widget.Context, from *TabsTriggerWidget, delta int) {
	n := len(l.triggers)
	if n == 0 {
		return
	}
	idx := -1
	for i, tr := range l.triggers {
		if tr == from {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	for step := 0; step < n; step++ {
		idx = (idx + delta + n) % n
		tr := l.triggers[idx]
		if tr.disabled || tr == from {
			continue
		}
		tr.activate(ctx)
		ctx.RequestFocus(tr)
		return
	}
}

// focusEdge selects the first (home) or last (end) enabled trigger.
func (l *TabsListWidget) focusEdge(ctx widget.Context, first bool) {
	idxs := make([]int, 0, len(l.triggers))
	for i := range l.triggers {
		idxs = append(idxs, i)
	}
	if !first {
		for i, j := 0, len(idxs)-1; i < j; i, j = i+1, j-1 {
			idxs[i], idxs[j] = idxs[j], idxs[i]
		}
	}
	for _, i := range idxs {
		tr := l.triggers[i]
		if tr.disabled {
			continue
		}
		tr.activate(ctx)
		ctx.RequestFocus(tr)
		return
	}
}

// ── TabsTrigger ─────────────────────────────────────────────────────────

// TabsTriggerWidget is a single tab trigger.
type TabsTriggerWidget struct {
	widget.WidgetBase

	value    string
	label    string
	disabled bool

	st   *tabsState
	list *TabsListWidget
	root *TabsWidget

	hovered      bool
	mouseDown    bool
	focusVisible bool
}

// TabsTrigger creates a trigger that activates the tab with the given
// value.
func TabsTrigger(value, label string) *TabsTriggerWidget {
	tr := &TabsTriggerWidget{
		value: value,
		label: label,
		st:    &tabsState{th: CurrentTheme()},
	}
	tr.SetVisible(true)
	tr.SetEnabled(true)
	return tr
}

// Disabled sets the disabled state (50% opacity, not selectable).
func (tr *TabsTriggerWidget) Disabled(v bool) *TabsTriggerWidget {
	tr.disabled = v
	return tr
}

// tabsFontFamily resolves the trigger font family, honoring custom theme
// fonts (custom families register a single face, so weight mapping only
// applies to the stock Geist families).
func tabsFontFamily(th *theme.Theme, weight int) string {
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(weight)
}

// Layout measures the trigger: label width + px-1.5 each side, height from
// the list's inner height (25px at default metrics).
func (tr *TabsTriggerWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	m := metrics.Tabs
	w := textmetrics.Width(tabsFontFamily(tr.st.th, m.TriggerFontWeight), m.TriggerFontSize, tr.label) + 2*m.TriggerPadX
	h := m.ListHeight - 2*m.ListPadding - m.TriggerHeightInset
	size := c.Constrain(geometry.Sz(w, h))
	tr.SetBounds(geometry.FromPointSize(tr.Position(), size))
	return size
}

// active reports whether this trigger's tab is selected.
func (tr *TabsTriggerWidget) active() bool {
	return tr.value != "" && tr.st.current() == tr.value
}

// Draw paints the trigger per the shadcn cva (see metrics/tabs.go).
func (tr *TabsTriggerWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !tr.IsVisible() {
		return
	}
	m := metrics.Tabs
	th := tr.st.th
	tok := th.Active()
	dark := th.IsDark()
	bounds := tr.Bounds()
	radius := th.RadiusMD()
	isActive := tr.active()
	line := tr.st.variantLine

	if isActive && !line {
		// data-[state=active]:shadow-sm (default variant).
		draw.Shadow(canvas, bounds, radius, metrics.ShadowSM)
		if dark {
			// dark:data-[state=active]:bg-input/30 + border-input.
			bg := draw.MulAlpha(tok.Input, m.DarkActiveBgOpacity)
			draw.BorderFill(canvas, bounds, draw.Fade(bg, tr.disabled), draw.Fade(tok.Input, tr.disabled), radius, m.TriggerBorderWidth)
		} else {
			// data-[state=active]:bg-background.
			canvas.DrawRoundRect(bounds, draw.Fade(tok.Background, tr.disabled), radius)
		}
	}

	// focus-visible:border-ring focus-visible:ring-[3px] ring-ring/50.
	if tr.focusVisible && !tr.disabled {
		draw.FocusRing(canvas, bounds, radius, draw.Alpha(tok.Ring, metrics.RingAlpha))
		draw.InsideBorder(canvas, bounds, radius, tok.Ring, m.TriggerBorderWidth)
	}

	// Label color: active -> foreground; idle light -> foreground/60;
	// idle dark -> muted-foreground, hover -> foreground.
	var col widget.Color
	switch {
	case isActive:
		col = tok.Foreground
	case dark && tr.hovered:
		col = tok.Foreground
	case dark:
		col = tok.MutedForeground
	default:
		col = draw.MulAlpha(tok.Foreground, m.IdleTextOpacity)
	}
	col = draw.Fade(col, tr.disabled)

	textRect := geometry.NewRect(
		bounds.Min.X,
		bounds.Min.Y+(bounds.Height()-m.TriggerLineHeight)/2,
		bounds.Width(),
		m.TriggerLineHeight,
	)
	family := tabsFontFamily(th, m.TriggerFontWeight)
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(tr.label, textRect, widget.TextStyle{
			FontFamily: family,
			FontSize:   m.TriggerFontSize,
			Color:      col,
			Align:      widget.TextAlignCenter,
		})
	} else {
		canvas.DrawText(tr.label, textRect, m.TriggerFontSize, col, m.TriggerFontWeight >= 600, widget.TextAlignCenter)
	}

	// Line-variant active indicator: 2px bg-foreground. Horizontal hangs
	// 5px below the trigger; vertical hangs 5px to the right of the column
	// (a left-rail indicator along the content edge).
	if isActive && line {
		var indicator geometry.Rect
		if tr.st.vertical() {
			indicator = geometry.NewRect(
				bounds.Max.X+m.UnderlineDrop-m.UnderlineHeight,
				bounds.Min.Y,
				m.UnderlineHeight,
				bounds.Height(),
			)
		} else {
			indicator = geometry.NewRect(
				bounds.Min.X,
				bounds.Max.Y+m.UnderlineDrop-m.UnderlineHeight,
				bounds.Width(),
				m.UnderlineHeight,
			)
		}
		canvas.DrawRect(indicator, draw.Fade(tok.Foreground, tr.disabled))
	}
}

// activate selects this trigger's tab and requests the repaint/relayout.
func (tr *TabsTriggerWidget) activate(ctx widget.Context) {
	if tr.disabled || tr.st.current() == tr.value {
		return
	}
	tr.st.set(tr.value)
	if tr.root != nil {
		tr.root.SetNeedsRedraw(true)
	}
	tr.SetNeedsRedraw(true)
	if ctx != nil {
		ctx.Invalidate() // content switch changes layout
	}
}

// Event handles hover, click activation, keyboard navigation, and
// focus-visible tracking.
func (tr *TabsTriggerWidget) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.MouseEvent:
		if tr.disabled {
			return false
		}
		switch ev.MouseType {
		case event.MouseEnter:
			tr.hovered = true
			tr.SetNeedsRedraw(true)
			if ctx != nil {
				ctx.InvalidateRect(tr.Bounds())
			}
			return true
		case event.MouseLeave:
			tr.hovered = false
			tr.SetNeedsRedraw(true)
			if ctx != nil {
				ctx.InvalidateRect(tr.Bounds())
			}
			return true
		case event.MousePress:
			if ev.Button != event.ButtonLeft {
				return false
			}
			tr.mouseDown = true
			if ctx != nil {
				ctx.RequestFocus(tr)
			}
			tr.activate(ctx)
			return true
		case event.MouseRelease:
			tr.mouseDown = false
			return true
		}
	case *event.KeyEvent:
		if tr.disabled || !tr.IsFocused() || tr.list == nil {
			return false
		}
		if ev.KeyType != event.KeyPress && ev.KeyType != event.KeyRepeat {
			return false
		}
		// Arrow keys map per orientation: horizontal uses Left/Right,
		// vertical uses Up/Down. Home/End jump to the edges in both.
		vertical := tr.st.vertical()
		switch ev.Key {
		case event.KeyLeft:
			if vertical {
				return false
			}
			tr.list.moveFocus(ctx, tr, -1)
			return true
		case event.KeyRight:
			if vertical {
				return false
			}
			tr.list.moveFocus(ctx, tr, +1)
			return true
		case event.KeyUp:
			if !vertical {
				return false
			}
			tr.list.moveFocus(ctx, tr, -1)
			return true
		case event.KeyDown:
			if !vertical {
				return false
			}
			tr.list.moveFocus(ctx, tr, +1)
			return true
		case event.KeyHome:
			tr.list.focusEdge(ctx, true)
			return true
		case event.KeyEnd:
			tr.list.focusEdge(ctx, false)
			return true
		case event.KeyEnter, event.KeySpace:
			tr.activate(ctx)
			return true
		}
	}
	// NOTE: *event.FocusEvent is window-level (OS focus gained/lost, no target
	// widget) and is broadcast to every trigger; consuming it here marked every
	// tab focus-visible whenever the window was focused, drawing a focus ring on
	// each (rendered as a solid box by the GPU stroke path, gg#369 — the faint
	// pill). Per-widget focus is driven by ctx.RequestFocus → SetFocused.
	return false
}

// SetFocused tracks focus-visible semantics: the ring only renders when
// focus did not arrive from a mouse press.
func (tr *TabsTriggerWidget) SetFocused(focused bool) {
	tr.WidgetBase.SetFocused(focused)
	if focused {
		tr.focusVisible = !tr.mouseDown
	} else {
		tr.focusVisible = false
	}
	tr.MarkRedrawLocal() // not SetNeedsRedraw: avoids context-lock re-entry in RequestFocus
}

// IsFocusable reports keyboard focusability (disabled triggers are
// skipped, like disabled:pointer-events-none).
func (tr *TabsTriggerWidget) IsFocusable() bool {
	return tr.IsVisible() && tr.IsEnabled() && !tr.disabled
}

// Children returns nil; the trigger is a leaf.
func (tr *TabsTriggerWidget) Children() []widget.Widget { return nil }

// ── TabsContent ─────────────────────────────────────────────────────────

// TabsContentWidget shows its content when its value is selected.
type TabsContentWidget struct {
	widget.WidgetBase

	st      *tabsState
	value   string
	content Widget
}

// TabsContent associates content with a tab value.
func TabsContent(value string, content Widget) *TabsContentWidget {
	c := &TabsContentWidget{
		st:      &tabsState{th: CurrentTheme()},
		value:   value,
		content: content,
	}
	c.SetVisible(true)
	c.SetEnabled(true)
	if ps, ok := content.(interface{ SetParent(widget.Widget) }); ok {
		ps.SetParent(c)
	}
	return c
}

// isActive reports whether this content's tab is selected.
func (c *TabsContentWidget) isActive() bool {
	return c.st != nil && c.value != "" && c.st.current() == c.value
}

// Layout sizes to the content when active, zero otherwise; the hidden
// content is also marked invisible so focus traversal skips it.
func (c *TabsContentWidget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	active := c.isActive()
	if v, ok := c.content.(interface{ SetVisible(bool) }); ok {
		v.SetVisible(active)
	}
	if !active {
		c.SetBounds(geometry.FromPointSize(c.Position(), geometry.Sz(0, 0)))
		return geometry.Sz(0, 0)
	}
	sz := c.content.Layout(ctx, constraints)
	setChildBounds(c.content, geometry.FromPointSize(geometry.Pt(0, 0), sz))
	c.SetBounds(geometry.FromPointSize(c.Position(), sz))
	return sz
}

// Draw renders the content when active.
func (c *TabsContentWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() || !c.isActive() {
		return
	}
	canvas.PushTransform(c.Bounds().Min)
	widget.StampScreenOrigin(c.content, canvas)
	widget.DrawChild(c.content, ctx, canvas)
	canvas.PopTransform()
}

// Event forwards to the content when active.
func (c *TabsContentWidget) Event(ctx widget.Context, e event.Event) bool {
	if !c.IsVisible() || !c.IsEnabled() || !c.isActive() {
		return false
	}
	return dispatchToChildren(ctx, e, c.Bounds(), []widget.Widget{c.content})
}

// Children returns the wrapped content.
func (c *TabsContentWidget) Children() []widget.Widget { return []widget.Widget{c.content} }

// Compile-time interface checks.
var (
	_ widget.Widget    = (*TabsWidget)(nil)
	_ widget.Lifecycle = (*TabsWidget)(nil)
	_ widget.Widget    = (*TabsListWidget)(nil)
	_ widget.Widget    = (*TabsTriggerWidget)(nil)
	_ widget.Focusable = (*TabsTriggerWidget)(nil)
	_ widget.Widget    = (*TabsContentWidget)(nil)
)

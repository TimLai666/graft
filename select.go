package graft

import (
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/painters"
	"github.com/TimLai666/graft/theme"
)

// Select is graft's shadcn Select: a trigger that opens a floating menu of
// selectable items.
//
// Wrap-vs-own decision (DESIGN.md §3.2, Report 1 §7.6): graft wraps the
// core/dropdown *painter* contract (painters.Dropdown implements
// dropdown.Painter, so raw gogpu/ui users can theme a core dropdown) but
// the SelectWidget itself is graft-OWNED. core/dropdown's Layout hardcodes
// a 48px trigger and 40px menu rows as unexported constants with no
// override hook, so it cannot reach shadcn's h-9/h-8 trigger or 32px item
// rows; and its menu anatomy (left-aligned, no right-side check) does not
// match Select items. The owned widget controls Layout, the overlay, item
// geometry (check on the right), keyboard navigation, and group/separator
// flattening, while delegating the actual drawing to painters.Dropdown so
// the two paths stay pixel-identical.
//
// Groups, separators, and labels are supported by flattening: the menu is a
// flat list of rows, where group labels and separators are non-selectable
// rows the keyboard navigation skips (core/dropdown's flat string model has
// no native group concept — noted as a flattening compromise).
type SelectWidget struct {
	widget.WidgetBase

	entries     []SelectEntry
	rows        []selectRow // flattened render rows
	placeholder string
	disabled    bool
	invalid     bool
	small       bool
	width       float32 // 0 = w-fit

	signal   state.Signal[string]
	initial  string
	value    string // current selection (value string), "" = none
	onChange func(string)

	hovered      bool
	focusVisible bool
	open         bool
	menu         *selectMenu

	theme *theme.Theme
}

// selectRow is one flattened menu row.
type selectRow struct {
	kind      rowKind
	value     string
	label     string
	disabled  bool
	itemIndex int // index among selectable items, -1 for non-items
}

type rowKind uint8

const (
	rowItem rowKind = iota
	rowLabel
	rowSeparator
)

// SelectEntry is one entry passed to Select (item, group, or separator).
type SelectEntry interface {
	appendRows(rows *[]selectRow, itemCount *int)
}

// SelectItemEntry is a selectable Select item.
type SelectItemEntry struct {
	value    string
	label    string
	disabled bool
}

// SelectItem creates a selectable item with the given value and display
// label.
func SelectItem(value, label string) *SelectItemEntry {
	return &SelectItemEntry{value: value, label: label}
}

// Disabled marks the item as non-selectable (rendered faded).
func (e *SelectItemEntry) Disabled(v bool) *SelectItemEntry {
	e.disabled = v
	return e
}

func (e *SelectItemEntry) appendRows(rows *[]selectRow, itemCount *int) {
	*rows = append(*rows, selectRow{
		kind:      rowItem,
		value:     e.value,
		label:     e.label,
		disabled:  e.disabled,
		itemIndex: *itemCount,
	})
	*itemCount++
}

// selectGroupEntry is a labeled group of items.
type selectGroupEntry struct {
	label string
	items []SelectEntry
}

// SelectGroup creates a labeled group; its items render under a 12px muted
// label row (flattening: the label is a non-selectable row).
func SelectGroup(label string, items ...SelectEntry) SelectEntry {
	return &selectGroupEntry{label: label, items: items}
}

func (g *selectGroupEntry) appendRows(rows *[]selectRow, itemCount *int) {
	*rows = append(*rows, selectRow{kind: rowLabel, label: g.label, itemIndex: -1})
	for _, it := range g.items {
		it.appendRows(rows, itemCount)
	}
}

// selectSeparatorEntry is a divider row.
type selectSeparatorEntry struct{}

// SelectSeparator creates a 1px divider row between items.
func SelectSeparator() SelectEntry { return selectSeparatorEntry{} }

func (selectSeparatorEntry) appendRows(rows *[]selectRow, itemCount *int) {
	*rows = append(*rows, selectRow{kind: rowSeparator, itemIndex: -1})
}

// Select builds a Select from item/group/separator entries.
func Select(entries ...SelectEntry) *SelectWidget {
	s := &SelectWidget{entries: entries}
	s.SetVisible(true)
	s.SetEnabled(true)
	s.rebuild()
	return s
}

func (s *SelectWidget) rebuild() {
	s.rows = s.rows[:0]
	itemCount := 0
	for _, e := range s.entries {
		e.appendRows(&s.rows, &itemCount)
	}
}

// Bind binds the selection to a controlled string signal.
func (s *SelectWidget) Bind(sig state.Signal[string]) *SelectWidget {
	s.signal = sig
	return s
}

// Value sets the initial (uncontrolled) selected value.
func (s *SelectWidget) Value(v string) *SelectWidget {
	s.initial = v
	s.value = v
	return s
}

// Placeholder sets the text shown when nothing is selected.
func (s *SelectWidget) Placeholder(text string) *SelectWidget {
	s.placeholder = text
	return s
}

// OnChange registers a selection-change observer.
func (s *SelectWidget) OnChange(fn func(string)) *SelectWidget {
	s.onChange = fn
	return s
}

// Sm switches the trigger to the small height (h-8).
func (s *SelectWidget) Sm() *SelectWidget {
	s.small = true
	return s
}

// Disabled toggles the disabled state.
func (s *SelectWidget) Disabled(v bool) *SelectWidget {
	s.disabled = v
	return s
}

// Invalid toggles the aria-invalid state (destructive ring + border).
func (s *SelectWidget) Invalid(v bool) *SelectWidget {
	s.invalid = v
	return s
}

// W pins the trigger width in px (default is w-fit).
func (s *SelectWidget) W(px float32) *SelectWidget {
	s.width = px
	return s
}

// Theme pins a specific theme instead of the process-wide current theme.
func (s *SelectWidget) Theme(th *theme.Theme) *SelectWidget {
	s.theme = th
	return s
}

func (s *SelectWidget) resolvedTheme() *theme.Theme {
	if s.theme != nil {
		return s.theme
	}
	return CurrentTheme()
}

func (s *SelectWidget) painter() painters.Dropdown {
	return PaintersFor(s.resolvedTheme()).Dropdown
}

// current returns the active selected value, preferring a bound signal.
func (s *SelectWidget) current() string {
	if s.signal != nil {
		return s.signal.Get()
	}
	return s.value
}

// labelFor returns the display label for a value, or "" if not found.
func (s *SelectWidget) labelFor(value string) string {
	for _, r := range s.rows {
		if r.kind == rowItem && r.value == value {
			return r.label
		}
	}
	return ""
}

// triggerText returns the displayed text and whether it is the placeholder.
func (s *SelectWidget) triggerText() (string, bool) {
	cur := s.current()
	if cur != "" {
		if lbl := s.labelFor(cur); lbl != "" {
			return lbl, false
		}
	}
	return s.placeholder, true
}

func (s *SelectWidget) triggerHeight() float32 {
	if s.small {
		return metrics.Select.TriggerHeightSm
	}
	return metrics.Select.TriggerHeight
}

// fitWidth measures the natural (w-fit) trigger width: text + paddings +
// gap + chevron.
func (s *SelectWidget) fitWidth() float32 {
	m := metrics.Select
	text, _ := s.triggerText()
	tw := textmetrics.Width(fonts.Family(m.FontWeight), m.FontSize, text)
	return tw + 2*m.TriggerPadX + m.TriggerGap + m.ChevronSize
}

// Layout sizes the trigger to w-fit (or the pinned width) at h-9/h-8.
func (s *SelectWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	w := s.width
	if w <= 0 {
		w = s.fitWidth()
	}
	size := c.Constrain(geometry.Sz(w, s.triggerHeight()))
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

// Draw renders the trigger via the shared painter.
func (s *SelectWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !s.IsVisible() {
		return
	}
	text, isPlaceholder := s.triggerText()
	st := &dropdown.TriggerPaintState{
		Bounds:        s.Bounds(),
		SelectedText:  text,
		IsPlaceholder: isPlaceholder,
		Open:          s.open,
		Focused:       s.IsFocused(),
		Hovered:       s.hovered,
		Disabled:      s.disabled,
	}
	s.painter().PaintTriggerEx(canvas, st, s.invalid, s.focusVisible)
}

// IsFocusable reports whether the Select can receive focus.
func (s *SelectWidget) IsFocusable() bool {
	return s.IsVisible() && s.IsEnabled() && !s.disabled
}

// Event handles trigger interaction (hover, click to open, keyboard).
func (s *SelectWidget) Event(ctx widget.Context, e event.Event) bool {
	if s.disabled {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return s.handleMouse(ctx, ev)
	case *event.KeyEvent:
		return s.handleKey(ctx, ev)
	case *event.FocusEvent:
		if ev.FocusType == event.FocusGained {
			s.focusVisible = true
			s.SetNeedsRedraw(true)
		} else if ev.FocusType == event.FocusLost {
			s.focusVisible = false
			s.SetNeedsRedraw(true)
		}
		return false
	default:
		return false
	}
}

func (s *SelectWidget) handleMouse(ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		s.hovered = true
		ctx.SetCursor(widget.CursorPointer)
		s.SetNeedsRedraw(true)
		ctx.InvalidateRect(s.Bounds())
		return true
	case event.MouseLeave:
		s.hovered = false
		ctx.SetCursor(widget.CursorDefault)
		s.SetNeedsRedraw(true)
		ctx.InvalidateRect(s.Bounds())
		return true
	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		ctx.RequestFocus(s)
		s.focusVisible = false // mouse focus does not show the ring
		s.toggle(ctx)
		s.SetNeedsRedraw(true)
		return true
	default:
		return false
	}
}

func (s *SelectWidget) handleKey(ctx widget.Context, e *event.KeyEvent) bool {
	if !s.IsFocused() {
		return false
	}
	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}
	switch e.Key {
	case event.KeyEnter, event.KeySpace, event.KeyDown, event.KeyUp:
		s.focusVisible = true
		if !s.open {
			s.openMenu(ctx)
		}
		return true
	case event.KeyEscape:
		if s.open {
			s.closeMenu(ctx)
			return true
		}
		return false
	default:
		return false
	}
}

func (s *SelectWidget) toggle(ctx widget.Context) {
	if s.open {
		s.closeMenu(ctx)
	} else {
		s.openMenu(ctx)
	}
}

// selectableCount returns the number of selectable item rows.
func (s *SelectWidget) selectableCount() int {
	n := 0
	for _, r := range s.rows {
		if r.kind == rowItem {
			n++
		}
	}
	return n
}

// currentItemIndex returns the row's itemIndex of the current selection,
// or -1.
func (s *SelectWidget) currentItemIndex() int {
	cur := s.current()
	for _, r := range s.rows {
		if r.kind == rowItem && r.value == cur {
			return r.itemIndex
		}
	}
	return -1
}

func (s *SelectWidget) openMenu(ctx widget.Context) {
	if s.open {
		return
	}
	om := ctx.OverlayManager()
	if om == nil {
		return
	}
	s.open = true

	th := s.resolvedTheme()
	m := metrics.Select
	anchor := s.ScreenBounds()
	menuW := anchor.Width()
	if menuW < m.ContentMinWidth {
		menuW = m.ContentMinWidth
	}
	menuSize := geometry.Sz(menuW, s.menuHeight())
	pos := overlay.Position(overlay.PlacementBelow, anchor, menuSize, ctx.WindowSize(), m.SideOffset)

	s.menu = newSelectMenu(s, th)
	s.menu.SetBounds(geometry.FromPointSize(pos, menuSize))
	om.PushOverlay(s.menu, func() { s.closeMenu(ctx) })
	s.SetNeedsRedraw(true)
	ctx.Invalidate()
}

// menuHeight returns the total content height for the flattened rows.
func (s *SelectWidget) menuHeight() float32 {
	m := metrics.Select
	h := 2 * m.ContentPad
	for _, r := range s.rows {
		h += s.rowHeight(r)
	}
	return h
}

func (s *SelectWidget) rowHeight(r selectRow) float32 {
	m := metrics.Select
	switch r.kind {
	case rowItem:
		return m.ItemHeight
	case rowLabel:
		return labelRowHeight
	case rowSeparator:
		return separatorRowHeight
	}
	return m.ItemHeight
}

func (s *SelectWidget) closeMenu(ctx widget.Context) {
	if !s.open {
		return
	}
	s.open = false
	if s.menu != nil {
		if om := ctx.OverlayManager(); om != nil {
			om.RemoveOverlay(s.menu)
		}
		s.menu = nil
	}
	s.SetNeedsRedraw(true)
	ctx.Invalidate()
}

// commit applies a new selection value and fires observers.
func (s *SelectWidget) commit(value string) {
	if s.signal != nil {
		s.signal.Set(value)
	} else {
		s.value = value
	}
	if s.onChange != nil {
		s.onChange(value)
	}
}

// Children returns nil; the menu lives in the overlay layer.
func (s *SelectWidget) Children() []widget.Widget { return nil }

// Mount wires the bound signal for push-based invalidation.
func (s *SelectWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil || s.signal == nil {
		return
	}
	b := state.BindToScheduler(s.signal, s, sched)
	s.AddBinding(b)
}

// Unmount is a no-op; bindings are cleaned up by WidgetBase.
func (s *SelectWidget) Unmount() {}

// --- internal overlay menu widget ---------------------------------------

const (
	// labelRowHeight is the height of a group-label row: px-2 py-1.5 around
	// a text-xs (12px) line ⇒ 12 + 12 = 24px. (Select label "px-2 py-1.5
	// text-xs".)
	labelRowHeight float32 = 24
	// separatorRowHeight is the height of a separator row: 1px line + my-1
	// (4px top+bottom = 8px) ⇒ 9px. (Select separator "-mx-1 my-1 h-px".)
	separatorRowHeight float32 = 9
)

// selectMenu is the overlay widget that renders the open Select menu and
// handles keyboard/mouse navigation. It owns the row anatomy so the painter
// only needs to render selectable item rows.
type selectMenu struct {
	widget.WidgetBase

	sel         *SelectWidget
	th          *theme.Theme
	highlighted int // itemIndex of the highlighted selectable item
}

func newSelectMenu(sel *SelectWidget, th *theme.Theme) *selectMenu {
	m := &selectMenu{sel: sel, th: th, highlighted: sel.currentItemIndex()}
	if m.highlighted < 0 {
		m.highlighted = m.firstEnabledItem()
	}
	m.SetVisible(true)
	m.SetEnabled(true)
	return m
}

// Dismiss closes the menu (overlay backdrop/Esc).
func (m *selectMenu) Dismiss() {
	m.sel.open = false
	m.sel.menu = nil
}

// Modal reports whether the menu blocks events below. Select menus are
// non-modal (click-outside dismiss).
func (m *selectMenu) Modal() bool { return false }

func (m *selectMenu) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return m.Bounds().Size()
}

// itemRowByItemIndex returns the flattened-row index for a selectable item
// index, or -1.
func (m *selectMenu) firstEnabledItem() int {
	for _, r := range m.sel.rows {
		if r.kind == rowItem && !r.disabled {
			return r.itemIndex
		}
	}
	return -1
}

func (m *selectMenu) lastEnabledItem() int {
	last := -1
	for _, r := range m.sel.rows {
		if r.kind == rowItem && !r.disabled {
			last = r.itemIndex
		}
	}
	return last
}

// Draw renders the menu surface and rows. The Select item anatomy (label
// left, check right) is owned here so the surface matches what
// painters.Dropdown.PaintMenu would draw for plain item rows while adding
// flattened group labels and separators.
func (m *selectMenu) Draw(_ widget.Context, canvas widget.Canvas) {
	sel := m.sel
	met := metrics.Select
	bounds := m.Bounds()

	// Surface (shadow, popover bg, border) + clip.
	th := m.th
	radius := th.RadiusLG() // content: rounded-lg
	tok := th.Active()
	selectMenuSurface(canvas, bounds, radius, tok.Popover, tok.Border, met.BorderWidth)

	canvas.PushClip(bounds)
	defer canvas.PopClip()

	// Walk flattened rows, tracking Y, drawing each row at its real height.
	y := bounds.Min.Y + met.ContentPad
	innerLeft := bounds.Min.X + met.ContentPad
	innerRight := bounds.Max.X - met.ContentPad
	itemRadius := th.RadiusSM()
	selectedIdx := sel.currentItemIndex()

	for _, r := range sel.rows {
		h := sel.rowHeight(r)
		rowRect := geometry.NewRect(innerLeft, y, innerRight-innerLeft, h)
		switch r.kind {
		case rowItem:
			m.drawItem(canvas, th, rowRect, r, itemRadius, r.itemIndex == m.highlighted, r.itemIndex == selectedIdx)
		case rowLabel:
			m.drawLabel(canvas, th, rowRect, r.label)
		case rowSeparator:
			m.drawSeparator(canvas, th, bounds, y, h)
		}
		y += h
	}
}

func (m *selectMenu) drawItem(canvas widget.Canvas, th *theme.Theme, rect geometry.Rect, r selectRow, itemRadius float32, highlighted, selected bool) {
	met := metrics.Select
	tok := th.Active()
	disabled := r.disabled

	labelColor := tok.PopoverForeground
	if highlighted && !disabled {
		canvas.DrawRoundRect(rect, tok.Accent, itemRadius)
		labelColor = tok.AccentForeground
	}
	labelColor = selectFade(labelColor, disabled)

	textRect := geometry.NewRect(
		rect.Min.X+met.ItemPadLeft,
		rect.Min.Y,
		rect.Width()-met.ItemPadLeft-met.ItemPadRight,
		rect.Height(),
	)
	selectDrawText(canvas, met.FontWeight, met.FontSize, r.label, textRect, labelColor)

	if selected {
		checkX := rect.Max.X - met.CheckRight - met.CheckSize
		checkY := rect.Center().Y - met.CheckSize/2
		selectDrawCheck(canvas, geometry.NewRect(checkX, checkY, met.CheckSize, met.CheckSize), labelColor)
	}
}

func (m *selectMenu) drawLabel(canvas widget.Canvas, th *theme.Theme, rect geometry.Rect, label string) {
	met := metrics.Select
	tok := th.Active()
	textRect := geometry.NewRect(rect.Min.X+met.ItemPadLeft, rect.Min.Y, rect.Width()-2*met.ItemPadLeft, rect.Height())
	selectDrawText(canvas, 400, selectLabelFontSize, label, textRect, tok.MutedForeground)
}

func (m *selectMenu) drawSeparator(canvas widget.Canvas, th *theme.Theme, bounds geometry.Rect, y, h float32) {
	tok := th.Active()
	// my-1 (4px above) centers the 1px line in the 9px row.
	lineY := y + 4
	canvas.DrawRect(geometry.NewRect(bounds.Min.X+selectSeparatorInset, lineY, bounds.Width()-2*selectSeparatorInset, 1), tok.Border)
}

// invalidate marks the menu dirty and requests a repaint frame. In
// event-driven render mode (the default) marking dirty alone never repaints —
// a frame must be requested through the context, otherwise the highlighted
// row change is invisible until the next unrelated event.
func (m *selectMenu) invalidate(ctx widget.Context) {
	m.SetNeedsRedraw(true)
	if ctx != nil {
		ctx.InvalidateRect(m.Bounds())
	}
}

// Event handles keyboard navigation and mouse selection in the menu.
func (m *selectMenu) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.KeyEvent:
		return m.handleKey(ctx, ev)
	case *event.MouseEvent:
		return m.handleMouse(ctx, ev)
	default:
		return false
	}
}

func (m *selectMenu) handleKey(ctx widget.Context, e *event.KeyEvent) bool {
	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}
	switch e.Key {
	case event.KeyDown:
		m.moveHighlight(1)
		m.invalidate(ctx)
		return true
	case event.KeyUp:
		m.moveHighlight(-1)
		m.invalidate(ctx)
		return true
	case event.KeyHome:
		m.highlighted = m.firstEnabledItem()
		m.invalidate(ctx)
		return true
	case event.KeyEnd:
		m.highlighted = m.lastEnabledItem()
		m.invalidate(ctx)
		return true
	case event.KeyEnter, event.KeySpace:
		m.selectHighlighted(ctx)
		return true
	case event.KeyEscape:
		m.sel.closeMenu(ctx)
		return true
	default:
		return false
	}
}

func (m *selectMenu) handleMouse(ctx widget.Context, e *event.MouseEvent) bool {
	bounds := m.Bounds()
	if !bounds.Contains(e.Position) {
		// Click outside dismisses (overlay also handles this, but be safe).
		if e.MouseType == event.MousePress {
			m.sel.closeMenu(ctx)
		}
		return false
	}
	idx, ok := m.itemAt(e.Position)
	switch e.MouseType {
	case event.MouseMove:
		if ok && idx != m.highlighted {
			m.highlighted = idx
			m.invalidate(ctx)
		}
		return true
	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		if ok {
			m.commitItem(ctx, idx)
		}
		return true
	default:
		return true
	}
}

// itemAt returns the selectable itemIndex at position p (true if over an
// enabled item).
func (m *selectMenu) itemAt(p geometry.Point) (int, bool) {
	met := metrics.Select
	bounds := m.Bounds()
	y := bounds.Min.Y + met.ContentPad
	for _, r := range m.sel.rows {
		h := m.sel.rowHeight(r)
		if r.kind == rowItem && p.Y >= y && p.Y < y+h && !r.disabled {
			return r.itemIndex, true
		}
		y += h
	}
	return -1, false
}

func (m *selectMenu) moveHighlight(delta int) {
	count := m.sel.selectableCount()
	if count == 0 {
		return
	}
	next := m.highlighted
	for i := 0; i < count; i++ {
		next += delta
		if next < 0 || next >= count {
			return
		}
		if !m.itemDisabled(next) {
			m.highlighted = next
			return
		}
	}
}

func (m *selectMenu) itemDisabled(itemIndex int) bool {
	for _, r := range m.sel.rows {
		if r.kind == rowItem && r.itemIndex == itemIndex {
			return r.disabled
		}
	}
	return true
}

func (m *selectMenu) valueForItem(itemIndex int) string {
	for _, r := range m.sel.rows {
		if r.kind == rowItem && r.itemIndex == itemIndex {
			return r.value
		}
	}
	return ""
}

func (m *selectMenu) selectHighlighted(ctx widget.Context) {
	if m.highlighted >= 0 && !m.itemDisabled(m.highlighted) {
		m.commitItem(ctx, m.highlighted)
	}
}

func (m *selectMenu) commitItem(ctx widget.Context, itemIndex int) {
	m.sel.commit(m.valueForItem(itemIndex))
	m.sel.closeMenu(ctx)
}

func (m *selectMenu) Children() []widget.Widget { return nil }

// --- shared helpers for the menu (component-prefixed to avoid collisions) -

const (
	// selectLabelFontSize is the group-label font size (text-xs).
	selectLabelFontSize float32 = 12
	// selectSeparatorInset is the separator horizontal inset (-mx-1 makes
	// the line span past p-1 padding, i.e. full width minus 0). We inset by
	// the content padding so the line aligns to the item content edges.
	selectSeparatorInset float32 = 4
)

func selectFade(c widget.Color, disabled bool) widget.Color {
	if disabled {
		c.A *= metrics.DisabledOpacity
		return c
	}
	return c
}

func selectMenuSurface(canvas widget.Canvas, bounds geometry.Rect, radius float32, bg, border widget.Color, borderW float32) {
	// shadow-md then fill then inside border.
	for _, l := range metrics.ShadowMD {
		canvas.DrawRoundRect(bounds.Expand(l.Grow).TranslateXY(0, l.DY), widget.RGBA(0, 0, 0, l.Alpha), radius+l.Grow)
	}
	canvas.DrawRoundRect(bounds, bg, radius)
	r := radius - borderW/2
	if r < 0 {
		r = 0
	}
	canvas.StrokeRoundRect(bounds.Expand(-borderW/2), border, r, borderW)
}

func selectDrawText(canvas widget.Canvas, weight int, size float32, text string, bounds geometry.Rect, color widget.Color) {
	if text == "" {
		return
	}
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(text, bounds, widget.TextStyle{
			FontFamily: fonts.Family(weight),
			FontSize:   size,
			Color:      color,
			Align:      widget.TextAlignLeft,
		})
		return
	}
	canvas.DrawText(text, bounds, size, color, weight >= 600, widget.TextAlignLeft)
}

func selectDrawCheck(canvas widget.Canvas, bounds geometry.Rect, color widget.Color) {
	icon.Draw(canvas, icons.Check, bounds, color)
}

// SelectMenuPreview returns the open Select menu of s rendered as direct
// content (a self-sizing widget), for documentation and golden tests. The
// menu shows the current selection checked and the first enabled item
// highlighted, exactly as it would appear in the overlay.
func SelectMenuPreview(s *SelectWidget) Widget {
	menu := newSelectMenu(s, s.resolvedTheme())
	return &selectMenuPreviewWidget{menu: menu, sel: s}
}

// selectMenuPreviewWidget lays out a selectMenu at its natural size so it
// can be rendered as standalone content (the overlay normally sets bounds).
type selectMenuPreviewWidget struct {
	widget.WidgetBase
	menu *selectMenu
	sel  *SelectWidget
}

func (p *selectMenuPreviewWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	m := metrics.Select
	w := p.sel.width
	if w < m.ContentMinWidth {
		w = m.ContentMinWidth
	}
	h := p.sel.menuHeight()
	// Pad by the shadow-md reach so the golden does not clip it.
	pad := float32(16)
	size := c.Constrain(geometry.Sz(w+2*pad, h+2*pad))
	p.SetBounds(geometry.FromPointSize(p.Position(), size))
	p.menu.SetBounds(geometry.NewRect(p.Position().X+pad, p.Position().Y+pad, w, h))
	return size
}

func (p *selectMenuPreviewWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	p.menu.Draw(ctx, canvas)
}

func (p *selectMenuPreviewWidget) Event(widget.Context, event.Event) bool { return false }
func (p *selectMenuPreviewWidget) Children() []widget.Widget              { return nil }

// Compile-time interface checks.
var (
	_ widget.Widget    = (*SelectWidget)(nil)
	_ widget.Focusable = (*SelectWidget)(nil)
	_ widget.Lifecycle = (*SelectWidget)(nil)
	_ widget.Widget    = (*selectMenu)(nil)
	_ overlay.Overlay  = (*selectMenu)(nil)
	_ widget.Widget    = (*selectMenuPreviewWidget)(nil)
)

package graft

import (
	"strings"

	corepopover "github.com/gogpu/ui/core/popover"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// ComboboxOption is one selectable entry: a stored value and a display label.
type ComboboxOption struct {
	Value string
	Label string
}

// ComboboxItem builds a combobox option from a value and label.
func ComboboxItem(value, label string) ComboboxOption {
	return ComboboxOption{Value: value, Label: label}
}

// ComboboxWidget is the shadcn Combobox: an outline Button trigger
// (selected label or placeholder + trailing chevrons-up-down) that opens a
// Popover containing a search input and a scrollable, case-insensitive
// filtered list. The selected item shows a leading check; an empty filter
// result shows "No results found." (docs/research/03-shadcn-pixel-spec.md §5
// Command + Popover + Button).
//
// Architecture decision: graft-owned composite reusing ButtonWidget for the
// trigger and managing its own popover overlay (like DatePicker), with a
// self-contained search field + list rows drawn from tokens. The list rows
// are owned rather than wrapping core/listview because the per-row check
// indicator and accent selection are pure token painting.
type ComboboxWidget struct {
	widget.WidgetBase

	options     []ComboboxOption
	placeholder string

	value    state.Signal[string]
	onChange func(string)

	width float32

	open  state.Signal[bool]
	shown bool
	om    widget.OverlayManager

	trigger *ButtonWidget
	content *comboboxContent

	theme *theme.Theme
}

// Combobox builds a combobox from a list of options.
func Combobox(options ...ComboboxOption) *ComboboxWidget {
	c := &ComboboxWidget{
		options:     options,
		placeholder: metrics.Combobox.Placeholder,
		value:       state.NewSignal(""),
		width:       metrics.Combobox.TriggerWidth,
		open:        state.NewSignal(false),
		theme:       CurrentTheme(),
	}
	c.SetVisible(true)
	c.SetEnabled(true)
	c.rebuildTrigger()
	c.content = newComboboxContent(c)
	return c
}

// Placeholder sets the trigger's empty-state label.
func (c *ComboboxWidget) Placeholder(s string) *ComboboxWidget {
	c.placeholder = s
	c.rebuildTrigger()
	return c
}

// Bind makes the selected value controlled by sig.
func (c *ComboboxWidget) Bind(sig state.Signal[string]) *ComboboxWidget {
	if sig != nil {
		c.value = sig
		c.rebuildTrigger()
	}
	return c
}

// OnChange registers the selection observer.
func (c *ComboboxWidget) OnChange(fn func(string)) *ComboboxWidget {
	c.onChange = fn
	return c
}

// W sets an explicit trigger/content width.
func (c *ComboboxWidget) W(px float32) *ComboboxWidget {
	c.width = px
	c.rebuildTrigger()
	return c
}

// Theme pins a specific theme.
func (c *ComboboxWidget) Theme(th *theme.Theme) *ComboboxWidget {
	if th != nil {
		c.theme = th
		c.rebuildTrigger()
		if c.content != nil {
			c.content.theme = th
		}
	}
	return c
}

func (c *ComboboxWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// labelFor returns the display label for a stored value.
func (c *ComboboxWidget) labelFor(value string) string {
	for _, o := range c.options {
		if o.Value == value {
			return o.Label
		}
	}
	return ""
}

// triggerLabel returns the selected label or the placeholder.
func (c *ComboboxWidget) triggerLabel() string {
	if v := c.value.Get(); v != "" {
		if l := c.labelFor(v); l != "" {
			return l
		}
	}
	return c.placeholder
}

// rebuildTrigger rebuilds the outline button trigger carrying only the
// label; the host paints the trailing chevrons-up-down at the right edge
// (justify-between).
func (c *ComboboxWidget) rebuildTrigger() {
	btn := Button(c.triggerLabel()).
		Outline().
		W(c.width).
		Theme(c.resolvedTheme()).
		OnClick(func() { c.open.Set(!c.open.Get()) })
	if c.value.Get() == "" {
		tok := c.resolvedTheme().Active()
		mc := tok.MutedForeground
		btn.Style(func(s *Style) { s.Foreground = &mc })
	}
	c.trigger = btn
	ovlSetParent(c.trigger, c)
}

// Layout sizes the host to the trigger.
func (c *ComboboxWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	size := c.trigger.Layout(ctx, cons)
	setWidgetBounds(c.trigger, geometry.NewRect(0, 0, size.Width, size.Height))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw renders the trigger and reconciles the overlay.
func (c *ComboboxWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	bounds := c.Bounds()
	canvas.PushTransform(bounds.Min)
	widget.StampScreenOrigin(c.trigger, canvas)
	widget.DrawChild(c.trigger, ctx, canvas)
	canvas.PopTransform()

	// Trailing chevrons-up-down at the right edge (justify-between), at 50%
	// opacity, inset by the button's horizontal padding.
	m := metrics.Combobox
	tok := c.resolvedTheme().Active()
	chevX := bounds.Max.X - 12 - m.ChevronSize // px-3 outline padding
	chevRect := geometry.NewRect(
		chevX, bounds.Min.Y+(bounds.Height()-m.ChevronSize)/2,
		m.ChevronSize, m.ChevronSize)
	icon.Draw(canvas, icons.ChevronsUpDown, chevRect, draw.MulAlpha(tok.Foreground, m.ChevronOpacity))

	c.syncOverlay(ctx)
}

// Event forwards to the trigger; its OnClick toggles the popover.
func (c *ComboboxWidget) Event(ctx widget.Context, e event.Event) bool {
	consumed := c.trigger.Event(ctx, ovlTranslate(e, c.Bounds().Min))
	c.syncOverlay(ctx)
	return consumed
}

// Children returns the inline trigger.
func (c *ComboboxWidget) Children() []widget.Widget {
	return []widget.Widget{c.trigger}
}

// Mount binds the open and value signals.
func (c *ComboboxWidget) Mount(ctx widget.Context) {
	if sched := ctx.Scheduler(); sched != nil {
		c.AddBinding(state.BindToScheduler(c.open, c, sched))
		c.AddBinding(state.BindToScheduler(c.value, c, sched))
	}
}

// Unmount implements widget.Lifecycle.
func (c *ComboboxWidget) Unmount() {}

// selectValue sets the value, closes the popover, and fires OnChange.
func (c *ComboboxWidget) selectValue(value string) {
	// Toggle off when re-selecting the current value (shadcn behavior).
	if c.value.Get() == value {
		value = ""
	}
	c.value.Set(value)
	c.rebuildTrigger()
	c.open.Set(false)
	c.SetNeedsRedraw(true)
	if c.onChange != nil {
		c.onChange(value)
	}
}

// syncOverlay pushes or removes the list overlay to match the open signal.
func (c *ComboboxWidget) syncOverlay(ctx widget.Context) {
	if ctx == nil || c.content == nil {
		return
	}
	om := ctx.OverlayManager()
	if om == nil {
		return
	}
	want := c.open.Get()
	if want == c.shown {
		return
	}
	if want {
		c.content.resetQuery()
		size := c.content.Layout(ctx, geometry.Loose(ctx.WindowSize()))
		pos := corepopover.CalculatePosition(
			corepopover.Bottom, c.anchorBounds(), size, ctx.WindowSize(),
			metrics.Popover.SideOffset)
		c.content.SetBounds(geometry.FromPointSize(pos, size))
		c.om = om
		om.PushOverlay(c.content, c.handleDismiss)
		c.shown = true
	} else {
		om.RemoveOverlay(c.content)
		c.shown = false
	}
	c.SetNeedsRedraw(true)
}

func (c *ComboboxWidget) handleDismiss() {
	if !c.shown {
		return
	}
	c.shown = false
	if c.om != nil {
		c.om.RemoveOverlay(c.content)
	}
	if c.open.Get() {
		c.open.Set(false)
	}
	c.SetNeedsRedraw(true)
}

func (c *ComboboxWidget) anchorBounds() geometry.Rect {
	if r := c.ScreenBounds(); !r.IsEmpty() {
		return r
	}
	return c.Bounds()
}

// IsOpen reports whether the popover is open.
func (c *ComboboxWidget) IsOpen() bool { return c.open.Get() }

// comboboxContent is the padding-free popover surface hosting the search
// field, the filtered list, and the empty state.
type comboboxContent struct {
	widget.WidgetBase

	owner   *ComboboxWidget
	query   string
	hovered int // 1-based filtered-row index, or 0

	theme *theme.Theme
}

func newComboboxContent(owner *ComboboxWidget) *comboboxContent {
	c := &comboboxContent{owner: owner, theme: owner.resolvedTheme()}
	c.SetVisible(true)
	c.SetEnabled(true)
	return c
}

func (c *comboboxContent) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

func (c *comboboxContent) resetQuery() {
	c.query = ""
	c.hovered = 0
}

// filtered returns the options whose label contains the query (case-
// insensitive).
func (c *comboboxContent) filtered() []ComboboxOption {
	if c.query == "" {
		return c.owner.options
	}
	q := strings.ToLower(c.query)
	var out []ComboboxOption
	for _, o := range c.owner.options {
		if strings.Contains(strings.ToLower(o.Label), q) {
			out = append(out, o)
		}
	}
	return out
}

// listHeight returns the height of the filtered list (or the empty state).
func (c *comboboxContent) listHeight() float32 {
	m := metrics.Combobox
	rows := c.filtered()
	if len(rows) == 0 {
		return m.EmptyPadY*2 + m.EmptyFontSize
	}
	h := m.ListPad*2 + float32(len(rows))*m.ItemHeight
	if h > m.MaxListHeight {
		h = m.MaxListHeight
	}
	return h
}

// Layout sizes the surface to the trigger width and content height.
func (c *comboboxContent) Layout(_ widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Combobox
	w := c.owner.width
	h := m.InputHeight + c.listHeight()
	size := cons.Constrain(geometry.Sz(w, h))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw paints the popover surface, the search row, the filtered list (with
// the selected check and hover accent), and the empty state.
func (c *comboboxContent) Draw(_ widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.resolvedTheme()
	tok := th.Active()
	rMD := th.RadiusMD()
	rSM := th.RadiusSM()
	m := metrics.Combobox
	bounds := c.Bounds()

	// Surface.
	draw.Shadow(canvas, bounds, rMD, metrics.ShadowMD)
	canvas.DrawRoundRect(bounds, tok.Popover, rMD)
	draw.InsideBorder(canvas, bounds, rMD, tok.Border, metrics.Popover.BorderWidth)

	// Search row: search icon + query/placeholder + bottom border.
	searchIconRect := geometry.NewRect(
		bounds.Min.X+m.InputPadX,
		bounds.Min.Y+(m.InputHeight-m.SearchIconSize)/2,
		m.SearchIconSize, m.SearchIconSize)
	icon.Draw(canvas, icons.Search, searchIconRect, draw.MulAlpha(tok.Foreground, m.SearchIconOpacity))

	textX := bounds.Min.X + m.InputPadX + m.SearchIconSize + m.InputGap
	textW := bounds.Width() - (m.InputPadX + m.SearchIconSize + m.InputGap) - m.InputPadX
	searchText := c.query
	searchColor := tok.PopoverForeground
	if searchText == "" {
		searchText = metrics.Combobox.SearchPlaceholder
		searchColor = tok.MutedForeground
	}
	searchRect := geometry.NewRect(textX, bounds.Min.Y, textW, m.InputHeight)
	drawComboText(canvas, searchText, searchRect, m.InputFontSize, 400,
		searchColor, comboFamily(th, 400), widget.TextAlignLeft)

	// Bottom border of the search row.
	by := bounds.Min.Y + m.InputHeight - m.SeparatorWidth/2
	canvas.DrawLine(
		geometry.Pt(bounds.Min.X, by), geometry.Pt(bounds.Max.X, by),
		tok.Border, m.SeparatorWidth)

	// List or empty state.
	rows := c.filtered()
	listTop := bounds.Min.Y + m.InputHeight
	if len(rows) == 0 {
		emptyRect := geometry.NewRect(bounds.Min.X, listTop+m.EmptyPadY, bounds.Width(), m.EmptyFontSize)
		drawComboText(canvas, metrics.Combobox.EmptyText, emptyRect, m.EmptyFontSize, 400,
			tok.MutedForeground, comboFamily(th, 400), widget.TextAlignCenter)
		return
	}

	selected := c.owner.value.Get()
	for i, o := range rows {
		rowY := listTop + m.ListPad + float32(i)*m.ItemHeight
		rowRect := geometry.NewRect(
			bounds.Min.X+m.ListPad, rowY,
			bounds.Width()-2*m.ListPad, m.ItemHeight)

		fg := tok.PopoverForeground
		if i+1 == c.hovered {
			canvas.DrawRoundRect(rowRect, tok.Accent, rSM)
			fg = tok.AccentForeground
		}

		// Leading check (only for the selected value).
		if o.Value == selected {
			checkRect := geometry.NewRect(
				rowRect.Min.X+m.ItemPadX,
				rowRect.Min.Y+(m.ItemHeight-m.CheckSize)/2,
				m.CheckSize, m.CheckSize)
			icon.Draw(canvas, icons.Check, checkRect, fg)
		}

		labelX := rowRect.Min.X + m.ItemPadX + m.CheckSize + m.ItemGap
		labelRect := geometry.NewRect(labelX, rowRect.Min.Y, rowRect.Max.X-labelX-m.ItemPadX, m.ItemHeight)
		drawComboText(canvas, o.Label, labelRect, m.ItemFontSize, 400,
			fg, comboFamily(th, 400), widget.TextAlignLeft)
	}
}

// Event handles typing into the search field, hover, and row selection.
func (c *comboboxContent) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.KeyEvent:
		return c.keyEvent(ev)
	case *event.MouseEvent:
		return c.mouseEvent(ctx, ev)
	}
	return false
}

func (c *comboboxContent) keyEvent(ev *event.KeyEvent) bool {
	if ev.KeyType != event.KeyPress && ev.KeyType != event.KeyRepeat {
		return false
	}
	switch {
	case ev.Key == event.KeyBackspace:
		if c.query != "" {
			c.query = c.query[:len(c.query)-1]
			c.hovered = 0
			c.SetNeedsRedraw(true)
		}
		return true
	case ev.Rune != 0 && ev.Rune >= 0x20:
		c.query += string(ev.Rune)
		c.hovered = 0
		c.SetNeedsRedraw(true)
		return true
	}
	return false
}

func (c *comboboxContent) mouseEvent(ctx widget.Context, ev *event.MouseEvent) bool {
	idx, opt, inRow := c.hitRow(ev.Position)
	switch ev.MouseType {
	case event.MouseMove, event.MouseEnter:
		newHover := 0
		if inRow {
			newHover = idx
			ctx.SetCursor(widget.CursorPointer)
		}
		if newHover != c.hovered {
			c.hovered = newHover
			c.SetNeedsRedraw(true)
		}
		return inRow
	case event.MouseLeave:
		if c.hovered != 0 {
			c.hovered = 0
			c.SetNeedsRedraw(true)
		}
		return false
	case event.MouseRelease:
		if ev.Button == event.ButtonLeft && inRow {
			c.owner.selectValue(opt.Value)
			return true
		}
	}
	return false
}

// hitRow returns the 1-based filtered-row index, its option, and whether the
// point is inside a list row.
func (c *comboboxContent) hitRow(pos geometry.Point) (int, ComboboxOption, bool) {
	m := metrics.Combobox
	rows := c.filtered()
	bounds := c.Bounds()
	listTop := bounds.Min.Y + m.InputHeight + m.ListPad
	local := pos.Sub(geometry.Pt(bounds.Min.X, listTop))
	if local.X < 0 || local.X > bounds.Width() || local.Y < 0 {
		return 0, ComboboxOption{}, false
	}
	row := int(local.Y / m.ItemHeight)
	if row < 0 || row >= len(rows) {
		return 0, ComboboxOption{}, false
	}
	return row + 1, rows[row], true
}

// Children returns nil; the content paints its own rows.
func (c *comboboxContent) Children() []widget.Widget { return nil }

// SetQuery sets the search query (test/preview helper).
func (c *comboboxContent) SetQuery(q string) { c.query = q }

// ComboboxContentPreview renders the open popover surface (search + list)
// as direct content for goldens and docs, with an optional pre-filled query
// and a selected value.
func ComboboxContentPreview(cb *ComboboxWidget, query, selected string, th *theme.Theme) Widget {
	if th == nil {
		th = CurrentTheme()
	}
	cb.value.Set(selected)
	content := newComboboxContent(cb)
	content.theme = th
	content.query = query
	return content
}

// comboFamily resolves the Geist family for a weight, honoring a custom
// theme sans font.
func comboFamily(th *theme.Theme, weight int) string {
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(weight)
}

// drawComboText paints a single line of combobox text.
func drawComboText(canvas widget.Canvas, s string, bounds geometry.Rect,
	size float32, weight int, col widget.Color, family string, align widget.TextAlign) {
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(s, bounds, widget.TextStyle{
			FontFamily: family,
			FontSize:   size,
			Color:      col,
			Align:      align,
		})
		return
	}
	canvas.DrawText(s, bounds, size, col, weight >= 600, align)
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*ComboboxWidget)(nil)
	_ widget.Lifecycle = (*ComboboxWidget)(nil)
	_ widget.Widget    = (*comboboxContent)(nil)
)

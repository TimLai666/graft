package graft

import (
	"strings"

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

// ---------------------------------------------------------------------------
// Data model: items, groups, separators
// ---------------------------------------------------------------------------

// CommandItemDef describes one selectable command entry.
type CommandItemDef struct {
	label    string
	iconData icon.IconData
	hasIcon  bool
	shortcut string
	onSelect func()
}

// CommandItem builds a command item with a label.
func CommandItem(label string) *CommandItemDef {
	return &CommandItemDef{label: label}
}

// Icon attaches a leading icon to the command item.
func (d *CommandItemDef) Icon(ic icon.IconData) *CommandItemDef {
	d.iconData = ic
	d.hasIcon = true
	return d
}

// Shortcut sets the right-aligned shortcut hint text.
func (d *CommandItemDef) Shortcut(s string) *CommandItemDef {
	d.shortcut = s
	return d
}

// OnSelect registers the callback invoked when this item is activated.
func (d *CommandItemDef) OnSelect(fn func()) *CommandItemDef {
	d.onSelect = fn
	return d
}

// CommandGroupDef describes a named group of command items.
type CommandGroupDef struct {
	heading string
	items   []*CommandItemDef
}

// CommandGroup builds a named group of command items.
func CommandGroup(heading string, items ...*CommandItemDef) *CommandGroupDef {
	return &CommandGroupDef{heading: heading, items: items}
}

// commandSeparator is a sentinel in the entry list.
type commandSeparator struct{}

// commandListEntry is a union of groups and separators, in declaration order.
type commandListEntry struct {
	group     *CommandGroupDef
	separator bool
}

// CommandListDef holds the ordered sequence of groups and separators.
type CommandListDef struct {
	entries []commandListEntry
}

// CommandList builds the list body from groups and separators. Each argument
// must be *CommandGroupDef or *commandSeparator (returned by CommandSeparator).
func CommandList(parts ...any) *CommandListDef {
	l := &CommandListDef{}
	for _, p := range parts {
		switch v := p.(type) {
		case *CommandGroupDef:
			l.entries = append(l.entries, commandListEntry{group: v})
		case *commandSeparator:
			_ = v
			l.entries = append(l.entries, commandListEntry{separator: true})
		}
	}
	return l
}

// CommandSeparator returns a separator to place between groups.
func CommandSeparator() *commandSeparator { return &commandSeparator{} }

// CommandInputDef holds configuration for the search input row.
type CommandInputDef struct {
	placeholder string
}

// CommandInput builds the search input descriptor.
func CommandInput() *CommandInputDef {
	return &CommandInputDef{placeholder: metrics.Command.Placeholder}
}

// Placeholder sets the search input placeholder text.
func (d *CommandInputDef) Placeholder(s string) *CommandInputDef {
	d.placeholder = s
	return d
}

// ---------------------------------------------------------------------------
// Flat row model used for rendering and navigation
// ---------------------------------------------------------------------------

// commandRowKind classifies a flat row.
type commandRowKind uint8

const (
	commandRowGroupLabel commandRowKind = iota
	commandRowItem
	commandRowSeparator
)

// commandRow is one row in the flattened, filtered list.
type commandRow struct {
	kind     commandRowKind
	label    string // group heading or item label
	iconData icon.IconData
	hasIcon  bool
	shortcut string
	item     *CommandItemDef // non-nil for item rows
}

// ---------------------------------------------------------------------------
// CommandWidget — the standalone command palette surface
// ---------------------------------------------------------------------------

// CommandWidget is the shadcn Command palette surface: a search input at the
// top and a grouped, filtered list of command items below. It does not manage
// its own overlay; use CommandDialogWidget to wrap it in a dialog overlay.
//
// Architecture decision: graft-owned composite. The widget paints directly
// via theme tokens and metrics, with no child widgets. Children() returns nil
// (leaf drawing). Keyboard navigation, filtering, and selection are
// self-contained.
type CommandWidget struct {
	widget.WidgetBase

	input *CommandInputDef
	list  *CommandListDef

	query    string
	hovered  int // 1-based index into navigable (item) rows, or 0
	onSelect func(string)

	theme *theme.Theme
}

// Command builds a command palette from an input descriptor and a list.
func Command(input *CommandInputDef, list *CommandListDef) *CommandWidget {
	if input == nil {
		input = CommandInput()
	}
	if list == nil {
		list = &CommandListDef{}
	}
	c := &CommandWidget{
		input: input,
		list:  list,
		theme: CurrentTheme(),
	}
	c.SetVisible(true)
	c.SetEnabled(true)
	return c
}

// OnSelect registers the callback invoked when an item is activated.
// The callback receives the item's label.
func (c *CommandWidget) OnSelect(fn func(string)) *CommandWidget {
	c.onSelect = fn
	return c
}

// Theme pins a specific theme.
func (c *CommandWidget) Theme(th *theme.Theme) *CommandWidget {
	if th != nil {
		c.theme = th
	}
	return c
}

func (c *CommandWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// SetQuery sets the search query (test/preview helper).
func (c *CommandWidget) SetQuery(q string) {
	c.query = q
	c.hovered = 0
}

// SetHovered sets the 1-based navigable item highlight (test/preview helper).
func (c *CommandWidget) SetHovered(idx int) { c.hovered = idx }

// flatRows builds the flattened row list, filtering items by the query.
func (c *CommandWidget) flatRows() []commandRow {
	q := strings.ToLower(c.query)
	var rows []commandRow
	for i, entry := range c.list.entries {
		if entry.separator {
			// Only emit a separator between groups that both have visible items.
			if len(rows) > 0 && i+1 < len(c.list.entries) {
				rows = append(rows, commandRow{kind: commandRowSeparator})
			}
			continue
		}
		g := entry.group
		if g == nil {
			continue
		}
		var matched []*CommandItemDef
		for _, it := range g.items {
			if q == "" || strings.Contains(strings.ToLower(it.label), q) {
				matched = append(matched, it)
			}
		}
		if len(matched) == 0 {
			continue
		}
		rows = append(rows, commandRow{
			kind:  commandRowGroupLabel,
			label: g.heading,
		})
		for _, it := range matched {
			rows = append(rows, commandRow{
				kind:     commandRowItem,
				label:    it.label,
				iconData: it.iconData,
				hasIcon:  it.hasIcon,
				shortcut: it.shortcut,
				item:     it,
			})
		}
	}
	// Strip trailing separators.
	for len(rows) > 0 && rows[len(rows)-1].kind == commandRowSeparator {
		rows = rows[:len(rows)-1]
	}
	return rows
}

// navigableItems returns only the item rows from a flat row list.
func navigableItems(rows []commandRow) []*CommandItemDef {
	var out []*CommandItemDef
	for _, r := range rows {
		if r.kind == commandRowItem {
			out = append(out, r.item)
		}
	}
	return out
}


// listHeight returns the pixel height of the flat row list (or the empty state).
func (c *CommandWidget) listHeight(rows []commandRow) float32 {
	if len(rows) == 0 {
		return metrics.Command.EmptyPadY*2 + metrics.Command.EmptyFontSize
	}
	m := metrics.Command
	h := m.ListPad * 2
	for _, r := range rows {
		switch r.kind {
		case commandRowGroupLabel:
			h += m.GroupLabelHeight
		case commandRowItem:
			h += m.ItemHeight
		case commandRowSeparator:
			h += m.SepMarginY*2 + m.SepHeight
		}
	}
	if h > m.DialogMaxHeight {
		h = m.DialogMaxHeight
	}
	return h
}

// Layout sizes the command surface to the dialog width.
func (c *CommandWidget) Layout(_ widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Command
	rows := c.flatRows()
	w := m.DialogWidth
	h := m.InputHeight + c.listHeight(rows)
	size := cons.Constrain(geometry.Sz(w, h))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw paints the command palette surface: shadow, card, search row, list.
func (c *CommandWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.resolvedTheme()
	tok := th.Active()
	rLG := th.RadiusLG() // content: rounded-lg (mirrors combobox/popover surface)
	rSM := th.RadiusSM()
	m := metrics.Command
	bounds := c.Bounds()

	// Surface: shadow + popover bg + border.
	draw.Shadow(canvas, bounds, rLG, metrics.ShadowMD)
	canvas.DrawRoundRect(bounds, tok.Popover, rLG)
	draw.InsideBorder(canvas, bounds, rLG, tok.Border, m.SeparatorWidth)

	// -- Search row --
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
		searchText = c.input.placeholder
		searchColor = tok.MutedForeground
	}
	searchRect := geometry.NewRect(textX, bounds.Min.Y, textW, m.InputHeight)
	cmdDrawText(canvas, searchText, searchRect, m.InputFontSize, 400,
		searchColor, cmdFamily(th, 400), widget.TextAlignLeft)

	// Bottom border of search row.
	by := bounds.Min.Y + m.InputHeight - m.SeparatorWidth/2
	canvas.DrawLine(
		geometry.Pt(bounds.Min.X, by), geometry.Pt(bounds.Max.X, by),
		tok.Border, m.SeparatorWidth)

	// -- List or empty state --
	rows := c.flatRows()
	listTop := bounds.Min.Y + m.InputHeight
	if len(rows) == 0 {
		emptyRect := geometry.NewRect(bounds.Min.X, listTop+m.EmptyPadY, bounds.Width(), m.EmptyFontSize)
		cmdDrawText(canvas, metrics.Command.EmptyText, emptyRect, m.EmptyFontSize, 400,
			tok.MutedForeground, cmdFamily(th, 400), widget.TextAlignCenter)
		return
	}

	cursorY := listTop + m.ListPad
	navIdx := 0 // 1-based navigable counter
	for _, r := range rows {
		switch r.kind {
		case commandRowGroupLabel:
			labelRect := geometry.NewRect(
				bounds.Min.X+m.ListPad+m.GroupLabelPadX,
				cursorY+m.GroupLabelPadY,
				bounds.Width()-2*m.ListPad-2*m.GroupLabelPadX,
				m.GroupLabelSize)
			cmdDrawText(canvas, r.label, labelRect, m.GroupLabelSize, 500,
				tok.MutedForeground, cmdFamily(th, 500), widget.TextAlignLeft)
			cursorY += m.GroupLabelHeight

		case commandRowItem:
			navIdx++
			rowRect := geometry.NewRect(
				bounds.Min.X+m.ListPad,
				cursorY,
				bounds.Width()-2*m.ListPad,
				m.ItemHeight)

			fg := tok.PopoverForeground
			if navIdx == c.hovered {
				canvas.DrawRoundRect(rowRect, tok.Accent, rSM)
				fg = tok.AccentForeground
			}

			contentX := rowRect.Min.X + m.ItemPadX
			// Optional icon.
			if r.hasIcon {
				iconRect := geometry.NewRect(
					contentX,
					rowRect.Min.Y+(m.ItemHeight-m.IconSize)/2,
					m.IconSize, m.IconSize)
				icon.Draw(canvas, r.iconData, iconRect, draw.MulAlpha(fg, m.IconOpacity))
				contentX += m.IconSize + m.ItemGap
			}

			// Shortcut at the right edge.
			shortcutW := float32(0)
			if r.shortcut != "" {
				shortcutW = float32(len(r.shortcut)) * m.ShortcutSize * 0.6
				shortcutRect := geometry.NewRect(
					rowRect.Max.X-m.ItemPadX-shortcutW,
					rowRect.Min.Y,
					shortcutW,
					m.ItemHeight)
				cmdDrawText(canvas, r.shortcut, shortcutRect, m.ShortcutSize, 400,
					tok.MutedForeground, cmdFamily(th, 400), widget.TextAlignRight)
			}

			// Label.
			labelEnd := rowRect.Max.X - m.ItemPadX
			if shortcutW > 0 {
				labelEnd -= shortcutW + m.ItemGap
			}
			labelRect := geometry.NewRect(contentX, rowRect.Min.Y, labelEnd-contentX, m.ItemHeight)
			cmdDrawText(canvas, r.label, labelRect, m.ItemFontSize, 400,
				fg, cmdFamily(th, 400), widget.TextAlignLeft)

			cursorY += m.ItemHeight

		case commandRowSeparator:
			sepY := cursorY + m.SepMarginY
			canvas.DrawLine(
				geometry.Pt(bounds.Min.X+m.ListPad, sepY),
				geometry.Pt(bounds.Max.X-m.ListPad, sepY),
				tok.Border, m.SepHeight)
			cursorY += m.SepMarginY*2 + m.SepHeight
		}
	}
}

// invalidate marks the palette dirty AND requests a frame. SetNeedsRedraw
// alone is not enough in event-driven render mode — without ctx.InvalidateRect
// the typed query / navigation highlight would not visibly update until some
// other event forced a repaint (mirrors internal/widgets/menu Panel.invalidate).
func (c *CommandWidget) invalidate(ctx widget.Context) {
	c.SetNeedsRedraw(true)
	ctx.InvalidateRect(c.Bounds())
}

// Event handles keyboard input for the command palette.
func (c *CommandWidget) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.KeyEvent:
		return c.keyEvent(ctx, ev)
	case *event.MouseEvent:
		return c.mouseEvent(ctx, ev)
	}
	return false
}

func (c *CommandWidget) keyEvent(ctx widget.Context, ev *event.KeyEvent) bool {
	if ev.KeyType != event.KeyPress && ev.KeyType != event.KeyRepeat {
		return false
	}
	navItems := navigableItems(c.flatRows())
	n := len(navItems)
	switch ev.Key {
	case event.KeyBackspace:
		if c.query != "" {
			c.query = c.query[:len(c.query)-1]
			c.hovered = 0
			c.invalidate(ctx)
		}
		return true
	case event.KeyDown:
		if n > 0 {
			c.hovered = c.hovered%n + 1
			c.invalidate(ctx)
		}
		return true
	case event.KeyUp:
		if n > 0 {
			c.hovered--
			if c.hovered <= 0 {
				c.hovered = n
			}
			c.invalidate(ctx)
		}
		return true
	case event.KeyEnter:
		if c.hovered > 0 && c.hovered <= n {
			item := navItems[c.hovered-1]
			if item.onSelect != nil {
				item.onSelect()
			}
			if c.onSelect != nil {
				c.onSelect(item.label)
			}
		}
		return true
	}
	if ev.Rune != 0 && ev.Rune >= 0x20 {
		c.query += string(ev.Rune)
		c.hovered = 0
		c.invalidate(ctx)
		return true
	}
	return false
}

func (c *CommandWidget) mouseEvent(ctx widget.Context, ev *event.MouseEvent) bool {
	idx, _, inRow := c.hitRow(ev.Position)
	switch ev.MouseType {
	case event.MouseMove, event.MouseEnter:
		newHover := 0
		if inRow {
			newHover = idx
			ctx.SetCursor(widget.CursorPointer)
		}
		if newHover != c.hovered {
			c.hovered = newHover
			c.invalidate(ctx)
		}
		return inRow
	case event.MouseLeave:
		if c.hovered != 0 {
			c.hovered = 0
			c.invalidate(ctx)
		}
		return true
	case event.MouseRelease:
		if ev.Button == event.ButtonLeft && inRow {
			navItems := navigableItems(c.flatRows())
			if idx > 0 && idx <= len(navItems) {
				item := navItems[idx-1]
				if item.onSelect != nil {
					item.onSelect()
				}
				if c.onSelect != nil {
					c.onSelect(item.label)
				}
			}
			return true
		}
	}
	return false
}

// hitRow returns the 1-based navigable-item index, its definition, and whether
// the point is inside a selectable row.
func (c *CommandWidget) hitRow(pos geometry.Point) (int, *CommandItemDef, bool) {
	m := metrics.Command
	bounds := c.Bounds()
	rows := c.flatRows()
	cursorY := bounds.Min.Y + m.InputHeight + m.ListPad
	navIdx := 0
	for _, r := range rows {
		var rowH float32
		switch r.kind {
		case commandRowGroupLabel:
			rowH = m.GroupLabelHeight
		case commandRowItem:
			rowH = m.ItemHeight
			navIdx++
			rowRect := geometry.NewRect(bounds.Min.X+m.ListPad, cursorY,
				bounds.Width()-2*m.ListPad, rowH)
			if rowRect.Contains(pos) {
				return navIdx, r.item, true
			}
		case commandRowSeparator:
			rowH = m.SepMarginY*2 + m.SepHeight
		}
		cursorY += rowH
	}
	return 0, nil, false
}

// Children returns nil; the command palette is a leaf widget.
func (c *CommandWidget) Children() []widget.Widget { return nil }

// ---------------------------------------------------------------------------
// CommandDialogWidget — wraps Command in a Dialog overlay
// ---------------------------------------------------------------------------

// CommandDialogWidget hosts a CommandWidget inside a modal dialog overlay,
// following the shadcn CommandDialog pattern. It is a zero-size host that
// watches an open signal and pushes the overlay when open.
type CommandDialogWidget struct {
	widget.WidgetBase

	command *CommandWidget
	open    state.Signal[bool]
	initial bool
	onOpen  func(bool)

	ctx     widget.Context
	overlay *commandDialogOverlay
	shown   bool
}

// CommandDialog creates a command-palette dialog from an input and a list.
func CommandDialog(input *CommandInputDef, list *CommandListDef) *CommandDialogWidget {
	cmd := Command(input, list)
	d := &CommandDialogWidget{
		command: cmd,
	}
	d.SetVisible(true)
	d.SetEnabled(true)
	return d
}

// Bind controls the open state from a signal.
func (d *CommandDialogWidget) Bind(open state.Signal[bool]) *CommandDialogWidget {
	d.open = open
	return d
}

// OnOpenChange registers an observer invoked when the open state changes.
func (d *CommandDialogWidget) OnOpenChange(fn func(bool)) *CommandDialogWidget {
	d.onOpen = fn
	return d
}

// OnSelect registers the callback invoked when a command is activated.
func (d *CommandDialogWidget) OnSelect(fn func(string)) *CommandDialogWidget {
	d.command.OnSelect(fn)
	return d
}

// Theme pins a specific theme on the inner command surface.
func (d *CommandDialogWidget) Theme(th *theme.Theme) *CommandDialogWidget {
	d.command.Theme(th)
	return d
}

// isOpen reads the current open state.
func (d *CommandDialogWidget) isOpen() bool {
	if d.open != nil {
		return d.open.Get()
	}
	return d.initial
}

// setOpen writes the open state and notifies.
func (d *CommandDialogWidget) setOpen(v bool) {
	if d.open != nil {
		d.open.Set(v)
	} else {
		d.initial = v
	}
	if d.onOpen != nil {
		d.onOpen(v)
	}
	d.sync()
}

// Mount binds the open signal.
func (d *CommandDialogWidget) Mount(ctx widget.Context) {
	d.ctx = ctx
	if d.open != nil {
		if sched := ctx.Scheduler(); sched != nil {
			d.AddBinding(state.BindToScheduler(d.open, d, sched))
		}
		state.SubscribeForever(d.open, func(bool) { d.sync() })
	}
	d.sync()
}

// Unmount removes any live overlay.
func (d *CommandDialogWidget) Unmount() {
	if d.shown {
		d.pop()
	}
}

// Layout reports zero size; the host is invisible chrome.
func (d *CommandDialogWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	d.ctx = ctx
	d.SetBounds(geometry.FromPointSize(d.Position(), geometry.Sz(0, 0)))
	d.sync()
	return c.Constrain(geometry.Sz(0, 0))
}

// Draw paints nothing; the content lives in the overlay.
func (d *CommandDialogWidget) Draw(ctx widget.Context, _ widget.Canvas) { d.ctx = ctx }

// Event ignores input; the overlay handles its own.
func (d *CommandDialogWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; the host is a leaf.
func (d *CommandDialogWidget) Children() []widget.Widget { return nil }

// sync reconciles the overlay with the current open state.
func (d *CommandDialogWidget) sync() {
	if d.ctx == nil {
		return
	}
	want := d.isOpen()
	if want && !d.shown {
		d.push()
	} else if !want && d.shown {
		d.pop()
	}
}

func (d *CommandDialogWidget) push() {
	om := d.ctx.OverlayManager()
	if om == nil {
		return
	}
	d.command.query = ""
	d.command.hovered = 0
	d.overlay = newCommandDialogOverlay(d.command, d.ctx.WindowSize(), func() {
		d.setOpen(false)
	})
	om.PushOverlay(d.overlay, func() { d.shown = false; d.overlay = nil })
	d.shown = true
	d.ctx.Invalidate()
}

func (d *CommandDialogWidget) pop() {
	if om := d.ctx.OverlayManager(); om != nil && d.overlay != nil {
		om.RemoveOverlay(d.overlay)
	}
	d.overlay = nil
	d.shown = false
	d.ctx.Invalidate()
}

// ---------------------------------------------------------------------------
// commandDialogOverlay — modal chassis (backdrop + centered command surface)
// ---------------------------------------------------------------------------

type commandDialogOverlay struct {
	widget.WidgetBase

	command    *CommandWidget
	windowSize geometry.Size
	onDismiss  func()
}

func newCommandDialogOverlay(command *CommandWidget, windowSize geometry.Size, onDismiss func()) *commandDialogOverlay {
	o := &commandDialogOverlay{
		command:    command,
		windowSize: windowSize,
		onDismiss:  onDismiss,
	}
	o.SetVisible(true)
	o.SetEnabled(true)
	return o
}

func (o *commandDialogOverlay) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := c.Constrain(o.windowSize)
	o.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))

	contentSize := o.command.Layout(ctx, geometry.Loose(size))
	x := (size.Width - contentSize.Width) / 2
	y := (size.Height - contentSize.Height) / 2
	if y < 0 {
		y = 0
	}
	o.command.SetBounds(geometry.FromPointSize(geometry.Pt(x, y), contentSize))
	return size
}

func (o *commandDialogOverlay) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	canvas.DrawRect(o.Bounds(), widget.RGBA(0, 0, 0, metrics.OverlayAlpha))
	o.command.Draw(ctx, canvas)
}

func (o *commandDialogOverlay) Event(ctx widget.Context, e event.Event) bool {
	// Forward to command surface first.
	if o.command.Event(ctx, e) {
		return true
	}

	// Escape closes.
	if ke, ok := e.(*event.KeyEvent); ok {
		if ke.KeyType == event.KeyPress && ke.Key == event.KeyEscape {
			o.Dismiss()
			return true
		}
	}

	// Backdrop click closes.
	if me, ok := e.(*event.MouseEvent); ok && me.MouseType == event.MousePress {
		if !o.command.Bounds().Contains(me.Position) {
			o.Dismiss()
			return true
		}
	}

	// Modal: swallow everything.
	return true
}

// Dismiss invokes the dismissal callback.
func (o *commandDialogOverlay) Dismiss() {
	if o.onDismiss != nil {
		o.onDismiss()
	}
}

// Modal reports true.
func (o *commandDialogOverlay) Modal() bool { return true }

func (o *commandDialogOverlay) Children() []widget.Widget {
	return []widget.Widget{o.command}
}

// ---------------------------------------------------------------------------
// Preview helper (for goldens and docs)
// ---------------------------------------------------------------------------

// CommandPreview renders the command palette surface (search + grouped list)
// as direct content for goldens and docs, with an optional pre-filled query
// and highlighted item.
func CommandPreview(cmd *CommandWidget, query string, hoveredItem int, th *theme.Theme) widget.Widget {
	if th == nil {
		th = CurrentTheme()
	}
	cmd.theme = th
	cmd.query = query
	cmd.hovered = hoveredItem
	return cmd
}

// ---------------------------------------------------------------------------
// Text helpers
// ---------------------------------------------------------------------------

// cmdFamily resolves the font family for a weight, honoring a custom theme
// sans font.
func cmdFamily(th *theme.Theme, weight int) string {
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(weight)
}

// cmdDrawText paints a single line of command palette text.
func cmdDrawText(canvas widget.Canvas, s string, bounds geometry.Rect,
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

// ---------------------------------------------------------------------------
// Compile-time interface checks
// ---------------------------------------------------------------------------

var (
	_ widget.Widget    = (*CommandWidget)(nil)
	_ widget.Widget    = (*CommandDialogWidget)(nil)
	_ widget.Lifecycle = (*CommandDialogWidget)(nil)
	_ widget.Widget    = (*commandDialogOverlay)(nil)
)

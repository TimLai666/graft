package menu

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// Panel is the shadcn menu surface: a rounded popover with a 1px border and
// shadow-md hosting vertical entry rows. It owns row layout, drawing, and
// keyboard/mouse navigation.
//
// The panel sizes itself to fit its content (width = max(min-w, widest
// row), height = sum of row heights + padding). Callers place it in an
// overlay; for standalone rendering (goldens) call Layout then Draw.
type Panel struct {
	widget.WidgetBase

	entries     []Entry
	th          *theme.Theme
	highlighted int // index into entries of the highlighted row, -1 = none

	// onClose, if set, is invoked when Escape is pressed.
	onClose func()
}

// NewPanel builds a menu panel for the given theme and entries. A nil theme
// resolves to the stock theme.
func NewPanel(th *theme.Theme, entries ...Entry) *Panel {
	p := &Panel{entries: entries, th: th, highlighted: -1}
	p.highlighted = p.firstSelectable()
	p.SetVisible(true)
	p.SetEnabled(true)
	return p
}

// OnClose registers a callback fired when Escape is pressed.
func (p *Panel) OnClose(fn func()) *Panel { p.onClose = fn; return p }

// Highlighted returns the index of the highlighted entry, or -1.
func (p *Panel) Highlighted() int { return p.highlighted }

// SetHighlighted sets the highlighted entry index (clamped to a selectable
// row when possible). Test/host hook.
func (p *Panel) SetHighlighted(i int) {
	if i >= 0 && i < len(p.entries) && p.entries[i].selectable() {
		p.highlighted = i
		p.SetNeedsRedraw(true)
	}
}

func (p *Panel) resolvedTheme() *theme.Theme {
	if p.th != nil {
		return p.th
	}
	return theme.New()
}

// --- layout -------------------------------------------------------------

func (p *Panel) rowHeight(e Entry) float32 {
	if e.kind() == KindSeparator {
		return metrics.Menu.SeparatorHeight
	}
	return metrics.Menu.ItemHeight
}

// ContentSize returns the panel's natural size for the current entries.
func (p *Panel) ContentSize() geometry.Size {
	m := metrics.Menu
	h := 2 * m.Pad
	maxRow := float32(0)
	for _, e := range p.entries {
		h += p.rowHeight(e)
		if w := p.rowNaturalWidth(e); w > maxRow {
			maxRow = w
		}
	}
	w := maxRow + 2*m.Pad
	if w < m.MinWidth {
		w = m.MinWidth
	}
	return geometry.Sz(w, h)
}

// rowNaturalWidth estimates a row's content width (indicator/icon + label +
// gap + shortcut) so the panel can fit the widest row.
func (p *Panel) rowNaturalWidth(e Entry) float32 {
	m := metrics.Menu
	family := fonts.Family(m.ItemFontWeight)
	switch en := e.(type) {
	case *ItemEntry:
		w := 2 * m.ItemPadX
		if en.IsInset {
			w = m.InsetPadLeft + m.ItemPadX
		}
		if en.HasIcon {
			w += m.IconSize + m.ItemGap
		}
		w += textmetrics.Width(family, m.FontSize, en.Label)
		if en.Shortcut != "" {
			w += m.ItemGap*2 + textmetrics.Width(family, m.ShortcutFontSize, en.Shortcut)
		}
		return w
	case *CheckboxEntry:
		return m.InsetPadLeft + textmetrics.Width(family, m.FontSize, en.Label) + m.ItemPadX
	case *RadioEntry:
		return m.InsetPadLeft + textmetrics.Width(family, m.FontSize, en.Label) + m.ItemPadX
	case *LabelEntry:
		left := m.ItemPadX
		if en.IsInset {
			left = m.InsetPadLeft
		}
		return left + textmetrics.Width(fonts.Family(m.LabelFontWeight), m.FontSize, en.Text) + m.ItemPadX
	default:
		return 0
	}
}

// Layout sizes the panel to its content, honoring any min-width from the
// constraints (e.g. a trigger-width floor).
func (p *Panel) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	size := p.ContentSize()
	if c.MinWidth > size.Width {
		size.Width = c.MinWidth
	}
	size = c.Constrain(size)
	p.SetBounds(geometry.FromPointSize(p.Position(), size))
	return size
}

// --- draw ---------------------------------------------------------------

// Draw renders the panel surface and all entry rows.
func (p *Panel) Draw(_ widget.Context, canvas widget.Canvas) {
	if !p.IsVisible() {
		return
	}
	th := p.resolvedTheme()
	tok := th.Active()
	m := metrics.Menu
	bounds := p.Bounds()
	radius := th.RadiusMD()

	// Surface: shadow-md, bg-popover, 1px border.
	draw.Shadow(canvas, bounds, radius, metrics.ShadowMD)
	canvas.DrawRoundRect(bounds, tok.Popover, radius)
	draw.InsideBorder(canvas, bounds, radius, tok.Border, m.BorderWidth)

	canvas.PushClip(bounds)
	defer canvas.PopClip()

	y := bounds.Min.Y + m.Pad
	innerLeft := bounds.Min.X + m.Pad
	innerRight := bounds.Max.X - m.Pad
	itemRadius := th.RadiusSM()

	for i, e := range p.entries {
		h := p.rowHeight(e)
		rect := geometry.NewRect(innerLeft, y, innerRight-innerLeft, h)
		highlighted := i == p.highlighted
		switch en := e.(type) {
		case *ItemEntry:
			p.drawItem(canvas, th, rect, en, itemRadius, highlighted)
		case *CheckboxEntry:
			p.drawCheckbox(canvas, th, rect, en, itemRadius, highlighted)
		case *RadioEntry:
			p.drawRadio(canvas, th, rect, en, itemRadius, highlighted)
		case *LabelEntry:
			p.drawLabel(canvas, th, rect, en)
		case *SeparatorEntry:
			p.drawSeparator(canvas, th, bounds, y)
		}
		y += h
	}
}

func (p *Panel) drawItem(canvas widget.Canvas, th *theme.Theme, rect geometry.Rect, e *ItemEntry, itemRadius float32, highlighted bool) {
	m := metrics.Menu
	tok := th.Active()
	dark := th.IsDark()
	disabled := e.IsDisabled

	textColor := tok.PopoverForeground
	if e.Destructive {
		textColor = tok.Destructive
	}

	if highlighted && !disabled {
		if e.Destructive {
			a := m.DestructiveFocusAlphaLight
			if dark {
				a = m.DestructiveFocusAlphaDark
			}
			canvas.DrawRoundRect(rect, draw.Alpha(tok.Destructive, a), itemRadius)
		} else {
			canvas.DrawRoundRect(rect, tok.Accent, itemRadius)
			textColor = tok.AccentForeground
		}
	}
	textColor = draw.Fade(textColor, disabled)

	left := rect.Min.X + m.ItemPadX
	if e.IsInset {
		left = rect.Min.X + m.InsetPadLeft
	}
	if e.HasIcon {
		iconColor := draw.Fade(tok.MutedForeground, disabled)
		if e.Destructive {
			iconColor = draw.Fade(tok.Destructive, disabled)
		}
		iconRect := geometry.NewRect(left, rect.Center().Y-m.IconSize/2, m.IconSize, m.IconSize)
		icon.Draw(canvas, e.IconData, iconRect, iconColor)
		left += m.IconSize + m.ItemGap
	}

	labelRect := geometry.NewRect(left, rect.Min.Y, rect.Max.X-left-m.ItemPadX, rect.Height())
	drawMenuText(canvas, m.ItemFontWeight, m.FontSize, e.Label, labelRect, textColor)

	if e.Shortcut != "" {
		shortcutColor := draw.Fade(tok.MutedForeground, disabled)
		drawMenuTextRight(canvas, m.ItemFontWeight, m.ShortcutFontSize, e.Shortcut,
			geometry.NewRect(rect.Min.X, rect.Min.Y, rect.Width()-m.ItemPadX, rect.Height()), shortcutColor)
	}
}

func (p *Panel) drawCheckbox(canvas widget.Canvas, th *theme.Theme, rect geometry.Rect, e *CheckboxEntry, itemRadius float32, highlighted bool) {
	m := metrics.Menu
	tok := th.Active()
	disabled := e.IsDisabled

	textColor := tok.PopoverForeground
	if highlighted && !disabled {
		canvas.DrawRoundRect(rect, tok.Accent, itemRadius)
		textColor = tok.AccentForeground
	}
	textColor = draw.Fade(textColor, disabled)

	if e.Checked && e.HasCheck {
		indRect := geometry.NewRect(rect.Min.X+m.IndicatorLeft, rect.Center().Y-m.CheckSize/2, m.CheckSize, m.CheckSize)
		icon.Draw(canvas, e.CheckIcon, indRect, textColor)
	}

	labelRect := geometry.NewRect(rect.Min.X+m.InsetPadLeft, rect.Min.Y, rect.Max.X-(rect.Min.X+m.InsetPadLeft)-m.ItemPadX, rect.Height())
	drawMenuText(canvas, m.ItemFontWeight, m.FontSize, e.Label, labelRect, textColor)
}

func (p *Panel) drawRadio(canvas widget.Canvas, th *theme.Theme, rect geometry.Rect, e *RadioEntry, itemRadius float32, highlighted bool) {
	m := metrics.Menu
	tok := th.Active()
	disabled := e.IsDisabled

	textColor := tok.PopoverForeground
	if highlighted && !disabled {
		canvas.DrawRoundRect(rect, tok.Accent, itemRadius)
		textColor = tok.AccentForeground
	}
	textColor = draw.Fade(textColor, disabled)

	if e.Selected {
		// shadcn uses CircleIcon size-2 fill-current: a FILLED 8px dot. The
		// lucide circle icon is only a stroked outline, so draw a filled
		// circle directly, centered in the 16px indicator slot at left-2.
		slotCenterX := rect.Min.X + m.IndicatorLeft + m.CheckSize/2
		canvas.DrawCircle(geometry.Pt(slotCenterX, rect.Center().Y), m.RadioDotSize/2, textColor)
	}

	labelRect := geometry.NewRect(rect.Min.X+m.InsetPadLeft, rect.Min.Y, rect.Max.X-(rect.Min.X+m.InsetPadLeft)-m.ItemPadX, rect.Height())
	drawMenuText(canvas, m.ItemFontWeight, m.FontSize, e.Label, labelRect, textColor)
}

func (p *Panel) drawLabel(canvas widget.Canvas, th *theme.Theme, rect geometry.Rect, e *LabelEntry) {
	m := metrics.Menu
	tok := th.Active()
	left := rect.Min.X + m.ItemPadX
	if e.IsInset {
		left = rect.Min.X + m.InsetPadLeft
	}
	labelRect := geometry.NewRect(left, rect.Min.Y, rect.Max.X-left-m.ItemPadX, rect.Height())
	drawMenuText(canvas, m.LabelFontWeight, m.FontSize, e.Text, labelRect, tok.Foreground)
}

func (p *Panel) drawSeparator(canvas widget.Canvas, th *theme.Theme, bounds geometry.Rect, y float32) {
	m := metrics.Menu
	tok := th.Active()
	lineY := y + 4 // my-1: 4px above the 1px line.
	canvas.DrawRect(geometry.NewRect(bounds.Min.X+m.SeparatorInset, lineY, bounds.Width()-2*m.SeparatorInset, 1), tok.Border)
}

// --- events -------------------------------------------------------------

// Event handles keyboard navigation and mouse selection.
func (p *Panel) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.KeyEvent:
		return p.handleKey(ctx, ev)
	case *event.MouseEvent:
		return p.handleMouse(ctx, ev)
	default:
		return false
	}
}

func (p *Panel) handleKey(_ widget.Context, e *event.KeyEvent) bool {
	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}
	switch e.Key {
	case event.KeyDown:
		p.move(1)
		p.SetNeedsRedraw(true)
		return true
	case event.KeyUp:
		p.move(-1)
		p.SetNeedsRedraw(true)
		return true
	case event.KeyHome:
		p.highlighted = p.firstSelectable()
		p.SetNeedsRedraw(true)
		return true
	case event.KeyEnd:
		p.highlighted = p.lastSelectable()
		p.SetNeedsRedraw(true)
		return true
	case event.KeyEnter, event.KeySpace:
		p.activate(p.highlighted)
		return true
	case event.KeyEscape:
		if p.onClose != nil {
			p.onClose()
		}
		return true
	default:
		return false
	}
}

func (p *Panel) handleMouse(_ widget.Context, e *event.MouseEvent) bool {
	bounds := p.Bounds()
	if !bounds.Contains(e.Position) {
		return false
	}
	idx := p.rowAt(e.Position)
	switch e.MouseType {
	case event.MouseMove:
		if idx >= 0 && idx != p.highlighted && p.entries[idx].selectable() {
			p.highlighted = idx
			p.SetNeedsRedraw(true)
		}
		return true
	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		if idx >= 0 && p.entries[idx].selectable() {
			p.activate(idx)
		}
		return true
	default:
		return true
	}
}

// rowAt returns the entry index at position pos, or -1.
func (p *Panel) rowAt(pos geometry.Point) int {
	bounds := p.Bounds()
	y := bounds.Min.Y + metrics.Menu.Pad
	for i, e := range p.entries {
		h := p.rowHeight(e)
		if pos.Y >= y && pos.Y < y+h {
			return i
		}
		y += h
	}
	return -1
}

// activate fires the appropriate callback for the entry at index i (and
// flips checkbox/radio state).
func (p *Panel) activate(i int) {
	if i < 0 || i >= len(p.entries) || !p.entries[i].selectable() {
		return
	}
	switch en := p.entries[i].(type) {
	case *ItemEntry:
		if en.OnSelectFn != nil {
			en.OnSelectFn()
		}
	case *CheckboxEntry:
		en.Checked = !en.Checked
		if en.OnChangeFn != nil {
			en.OnChangeFn(en.Checked)
		}
		p.SetNeedsRedraw(true)
	case *RadioEntry:
		if en.OnSelectFn != nil {
			en.OnSelectFn(en.Value)
		}
	}
}

func (p *Panel) move(delta int) {
	if len(p.entries) == 0 {
		return
	}
	i := p.highlighted
	for n := 0; n < len(p.entries); n++ {
		i += delta
		if i < 0 || i >= len(p.entries) {
			return
		}
		if p.entries[i].selectable() {
			p.highlighted = i
			return
		}
	}
}

func (p *Panel) firstSelectable() int {
	for i, e := range p.entries {
		if e.selectable() {
			return i
		}
	}
	return -1
}

func (p *Panel) lastSelectable() int {
	last := -1
	for i, e := range p.entries {
		if e.selectable() {
			last = i
		}
	}
	return last
}

// Children returns nil; the panel draws its rows directly.
func (p *Panel) Children() []widget.Widget { return nil }

// --- text helpers -------------------------------------------------------

func drawMenuText(canvas widget.Canvas, weight int, size float32, text string, bounds geometry.Rect, color widget.Color) {
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

func drawMenuTextRight(canvas widget.Canvas, weight int, size float32, text string, bounds geometry.Rect, color widget.Color) {
	if text == "" {
		return
	}
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(text, bounds, widget.TextStyle{
			FontFamily: fonts.Family(weight),
			FontSize:   size,
			Color:      color,
			Align:      widget.TextAlignRight,
		})
		return
	}
	canvas.DrawText(text, bounds, size, color, weight >= 600, widget.TextAlignRight)
}

// Compile-time interface check.
var _ widget.Widget = (*Panel)(nil)

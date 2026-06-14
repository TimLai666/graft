package graft

import (
	"github.com/gogpu/ui/a11y"
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

// TextareaWidget is shadcn's Textarea: a multi-line text field rendered in the
// shadcn style (docs/research/03-shadcn-pixel-spec.md §5 "Textarea").
//
// WRAP DECISION (DESIGN.md §3.2): graft-OWNED. core/textfield is single-line
// (its cursor model is a flat rune index with no concept of visual rows), so a
// multi-line control with soft-wrap, per-row caret placement, and field-sizing
// auto-grow cannot be expressed by wrapping it. Textarea therefore owns its
// editing model: a plain rune buffer, a caret as a rune offset, soft-wrap into
// visual rows computed from textmetrics measurements, and a content height that
// grows with the wrapped row count (min-h 64).
//
// v1 editing scope: insert runes, Enter inserts a newline, Backspace/Delete,
// Left/Right/Up/Down/Home/End caret motion, basic mouse caret placement, and a
// single contiguous selection (shift+arrows / drag) with replace-on-type. The
// control scrolls vertically when content exceeds the visible area; horizontal
// scroll is unnecessary because text soft-wraps.
type TextareaWidget struct {
	widget.WidgetBase

	theme *theme.Theme

	placeholder string

	// text is the canonical content (uncontrolled mode); valueSig overrides it
	// when bound.
	text     string
	valueSig state.Signal[string]
	onChange func(string)

	invalid  bool
	disabled bool

	width float32 // explicit width in px (.W); 0 = fill available
	rows  int     // requested minimum visible rows (Rows); 0 = min-h 64

	// caret is the rune offset of the insertion point; sel is the other end of
	// the selection (== caret when there is no selection).
	caret int
	sel   int

	// scrollY is the vertical scroll offset in px (top of the viewport).
	scrollY float32

	focusVisible bool
	dragging     bool

	clipboard string

	// wrap caches the last computed soft-wrap layout (rebuilt every Layout).
	wrap wrapLayout
}

// wrapLayout is one soft-wrap pass: rows of rune ranges into the canonical
// content plus the inner content width they were wrapped to.
type wrapLayout struct {
	rows  []wrapRow
	width float32
}

// wrapRow is one visual row: [start,end) rune offsets into the content and
// whether it ends at a hard newline (so the caret can sit past the last glyph).
type wrapRow struct {
	start, end int
	hardBreak  bool
}

// Textarea creates an empty multi-line text field snapshotting the current
// theme.
func Textarea() *TextareaWidget {
	t := &TextareaWidget{theme: CurrentTheme()}
	t.SetVisible(true)
	t.SetEnabled(true)
	return t
}

// Placeholder sets the empty-state placeholder text.
func (t *TextareaWidget) Placeholder(s string) *TextareaWidget {
	t.placeholder = s
	return t
}

// Value sets the initial (uncontrolled) text content.
func (t *TextareaWidget) Value(s string) *TextareaWidget {
	t.text = s
	t.caret = len([]rune(s))
	t.sel = t.caret
	return t
}

// Bind makes the textarea controlled by sig: it renders sig's value and writes
// edits back to it. The binding is registered in Mount.
func (t *TextareaWidget) Bind(sig state.Signal[string]) *TextareaWidget {
	t.valueSig = sig
	return t
}

// OnChange registers a callback fired on every content change.
func (t *TextareaWidget) OnChange(fn func(string)) *TextareaWidget {
	t.onChange = fn
	return t
}

// Disabled sets the disabled state (faded, not focusable, ignores input).
func (t *TextareaWidget) Disabled(v bool) *TextareaWidget {
	t.disabled = v
	return t
}

// Invalid sets the aria-invalid state (destructive border + ring).
func (t *TextareaWidget) Invalid(v bool) *TextareaWidget {
	t.invalid = v
	return t
}

// W sets an explicit width in px (otherwise the textarea fills available
// width).
func (t *TextareaWidget) W(px float32) *TextareaWidget {
	t.width = px
	return t
}

// Rows sets a minimum number of visible text rows; the control still grows
// with content past this. Zero keeps the shadcn min-h 64.
func (t *TextareaWidget) Rows(n int) *TextareaWidget {
	t.rows = n
	return t
}

// Theme pins a specific theme instead of the process-wide current theme.
func (t *TextareaWidget) Theme(th *theme.Theme) *TextareaWidget {
	if th != nil {
		t.theme = th
	}
	return t
}

// resolvedText returns the current content (bound signal wins).
func (t *TextareaWidget) resolvedText() string {
	if t.valueSig != nil {
		return t.valueSig.Get()
	}
	return t.text
}

// setText writes new content, clamps the caret, fires OnChange, and pushes to
// the bound signal.
func (t *TextareaWidget) setText(s string) {
	if t.valueSig != nil {
		t.valueSig.Set(s)
	} else {
		t.text = s
	}
	n := len([]rune(s))
	if t.caret > n {
		t.caret = n
	}
	if t.sel > n {
		t.sel = n
	}
	if t.onChange != nil {
		t.onChange(s)
	}
}

// fontFamily resolves the registered family for the body weight, honoring
// custom theme fonts.
func (t *TextareaWidget) fontFamily() string {
	if t.theme.FontSans != theme.DefaultFontSans {
		return t.theme.FontSans
	}
	return fonts.Family(metrics.Textarea.FontWeight)
}

// minHeight returns the configured minimum control height: min-h 64, raised to
// fit Rows(n) lines when larger.
func (t *TextareaWidget) minHeight() float32 {
	h := metrics.Textarea.MinHeight
	if t.rows > 0 {
		rowsH := 2*metrics.Textarea.PadY + float32(t.rows)*metrics.Textarea.LineHeight
		if rowsH > h {
			h = rowsH
		}
	}
	return h
}

// computeWrap soft-wraps the content into visual rows for an inner content
// width (px between the horizontal paddings). Words longer than the line are
// broken at the rune that overflows.
func (t *TextareaWidget) computeWrap(innerW float32) wrapLayout {
	family := t.fontFamily()
	size := metrics.Textarea.FontSize
	content := []rune(t.resolvedText())

	var rows []wrapRow
	lineStart := 0
	for i := 0; i <= len(content); i++ {
		atEnd := i == len(content)
		if atEnd || content[i] == '\n' {
			// Hard line [lineStart,i): soft-wrap it to innerW.
			rows = append(rows, wrapHardLine(family, size, content, lineStart, i, innerW)...)
			lineStart = i + 1
			if atEnd {
				break
			}
		}
	}
	if len(rows) == 0 {
		rows = []wrapRow{{start: 0, end: 0, hardBreak: true}}
	}
	return wrapLayout{rows: rows, width: innerW}
}

// wrapHardLine soft-wraps a single hard line [start,end) of content into rows
// that fit innerW. The last produced row carries hardBreak=true.
func wrapHardLine(family string, size float32, content []rune, start, end int, innerW float32) []wrapRow {
	if innerW <= 0 || start >= end {
		return []wrapRow{{start: start, end: end, hardBreak: true}}
	}
	var rows []wrapRow
	rowStart := start
	lastSpace := -1 // last whitespace boundary in the current row
	i := start
	for i < end {
		// Measure [rowStart, i+1].
		w := textmetrics.Width(family, size, string(content[rowStart:i+1]))
		if w > innerW && i > rowStart {
			// Overflow: break at the last space if any, else at i.
			brk := i
			if lastSpace > rowStart {
				brk = lastSpace + 1 // include the space on this row
			}
			rows = append(rows, wrapRow{start: rowStart, end: brk})
			rowStart = brk
			lastSpace = -1
			i = brk
			continue
		}
		if content[i] == ' ' {
			lastSpace = i
		}
		i++
	}
	rows = append(rows, wrapRow{start: rowStart, end: end, hardBreak: true})
	return rows
}

// Layout sizes the control: explicit/available width, height = max(min-h, wrap
// rows). field-sizing-content auto-grows with the wrapped line count.
func (t *TextareaWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	w := t.width
	if w <= 0 {
		w = c.MaxWidth
		if w <= 0 || w >= geometry.Infinity {
			w = 200
		}
	}

	innerW := w - 2*metrics.Textarea.PadX - 2*metrics.Textarea.BorderWidth
	if innerW < 0 {
		innerW = 0
	}
	t.wrap = t.computeWrap(innerW)

	contentH := float32(len(t.wrap.rows))*metrics.Textarea.LineHeight + 2*metrics.Textarea.PadY
	h := contentH
	if min := t.minHeight(); h < min {
		h = min
	}

	size := c.Constrain(geometry.Sz(w, h))
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

// rowOfCaret returns the visual row index containing the caret rune offset.
func (t *TextareaWidget) rowOfCaret(caret int) int {
	rows := t.wrap.rows
	for i, r := range rows {
		// Caret belongs to a row if it is within [start, end]; for a non-last
		// soft-wrapped row the end boundary belongs to the next row.
		if caret < r.end {
			return i
		}
		if caret == r.end {
			if r.hardBreak || i == len(rows)-1 {
				return i
			}
			// Soft break: the offset is the start of the next row.
			continue
		}
	}
	if len(rows) == 0 {
		return 0
	}
	return len(rows) - 1
}

// caretXY returns the caret position in inner content coordinates (px from the
// content origin, before padding/scroll).
func (t *TextareaWidget) caretXY(caret int) (x, row float32) {
	content := []rune(t.resolvedText())
	ri := t.rowOfCaret(caret)
	r := t.wrap.rows[ri]
	from := r.start
	if caret < from {
		caret = from
	}
	if caret > r.end {
		caret = r.end
	}
	x = textmetrics.Width(t.fontFamily(), metrics.Textarea.FontSize, string(content[from:caret]))
	return x, float32(ri)
}

// Draw paints shadow, fill, border, focus/invalid ring, the wrapped text (or
// placeholder), the selection, and the caret.
func (t *TextareaWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !t.IsVisible() {
		return
	}
	th := t.theme
	tok := th.Active()
	dark := th.IsDark()
	bounds := t.Bounds()
	radius := th.RadiusLG()
	disabled := t.disabled

	if !disabled {
		draw.Shadow(canvas, bounds, radius, metrics.ShadowXS)
	}
	if dark {
		fill := draw.MulAlpha(tok.Input, metrics.Textarea.DarkFillAlpha)
		canvas.DrawRoundRect(bounds, draw.Fade(fill, disabled), radius)
	}

	borderColor := tok.Input
	switch {
	case t.invalid:
		borderColor = tok.Destructive
		a := metrics.InvalidRingAlphaLight
		if dark {
			a = metrics.InvalidRingAlphaDark
		}
		draw.FocusRing(canvas, bounds, radius, draw.Alpha(tok.Destructive, a))
	case t.focusVisible && !disabled:
		borderColor = tok.Ring
		draw.FocusRing(canvas, bounds, radius, draw.Alpha(tok.Ring, metrics.RingAlpha))
	}
	draw.InsideBorder(canvas, bounds, radius, draw.Fade(borderColor, disabled), metrics.Textarea.BorderWidth)

	content := geometry.NewRect(
		bounds.Min.X+metrics.Textarea.PadX+metrics.Textarea.BorderWidth,
		bounds.Min.Y+metrics.Textarea.PadY+metrics.Textarea.BorderWidth,
		bounds.Width()-2*(metrics.Textarea.PadX+metrics.Textarea.BorderWidth),
		bounds.Height()-2*(metrics.Textarea.PadY+metrics.Textarea.BorderWidth),
	)
	canvas.PushClip(content)
	defer canvas.PopClip()

	family := t.fontFamily()
	size := metrics.Textarea.FontSize
	lineH := metrics.Textarea.LineHeight
	runes := []rune(t.resolvedText())

	// Placeholder when empty.
	if len(runes) == 0 {
		if t.placeholder != "" {
			lineBounds := geometry.NewRect(content.Min.X, content.Min.Y-t.scrollY, content.Width(), lineH)
			drawTextareaText(canvas, t.placeholder, lineBounds, family, size, draw.Fade(tok.MutedForeground, disabled))
		}
		t.drawCaret(canvas, content, tok, disabled)
		return
	}

	// Selection highlight (under text).
	if t.caret != t.sel && !disabled {
		t.drawSelection(canvas, content, runes, family, size, tok)
	}

	// Each wrapped row.
	fg := draw.Fade(tok.Foreground, disabled)
	for i, r := range t.wrap.rows {
		y := content.Min.Y + float32(i)*lineH - t.scrollY
		if y+lineH < content.Min.Y || y > content.Max.Y {
			continue // culled
		}
		seg := string(runes[r.start:r.end])
		lineBounds := geometry.NewRect(content.Min.X, y, content.Width(), lineH)
		drawTextareaText(canvas, seg, lineBounds, family, size, fg)
	}

	t.drawCaret(canvas, content, tok, disabled)
}

// drawSelection paints the selection highlight in Primary across the wrapped
// rows it spans.
func (t *TextareaWidget) drawSelection(canvas widget.Canvas, content geometry.Rect, runes []rune, family string, size float32, tok *theme.Tokens) {
	start, end := t.selRange()
	lineH := metrics.Textarea.LineHeight
	for i, r := range t.wrap.rows {
		rs, re := r.start, r.end
		if re <= start || rs >= end {
			continue
		}
		s := max2(rs, start)
		e := min2(re, end)
		x1 := content.Min.X + textmetrics.Width(family, size, string(runes[rs:s]))
		x2 := content.Min.X + textmetrics.Width(family, size, string(runes[rs:e]))
		y := content.Min.Y + float32(i)*lineH - t.scrollY
		canvas.DrawRect(geometry.NewRect(x1, y, x2-x1, lineH), tok.Primary)
	}
}

// drawCaret paints the 1px Foreground caret at the current insertion point when
// focused with no active selection.
func (t *TextareaWidget) drawCaret(canvas widget.Canvas, content geometry.Rect, tok *theme.Tokens, disabled bool) {
	if !t.focusVisible || disabled || t.caret != t.sel {
		return
	}
	x, row := t.caretXY(t.caret)
	y := content.Min.Y + row*metrics.Textarea.LineHeight - t.scrollY
	caret := geometry.NewRect(content.Min.X+x, y, metrics.Textarea.CaretWidth, metrics.Textarea.LineHeight)
	canvas.DrawRect(caret, tok.Foreground)
}

// drawTextareaText draws one left-aligned line via StyledTextDrawer, falling
// back to DrawText (weight 400 → not bold) on mock canvases.
func drawTextareaText(canvas widget.Canvas, text string, bounds geometry.Rect, family string, size float32, col widget.Color) {
	if text == "" {
		return
	}
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(text, bounds, widget.TextStyle{
			FontFamily: family,
			FontSize:   size,
			Color:      col,
			Align:      widget.TextAlignLeft,
		})
		return
	}
	canvas.DrawText(text, bounds, size, col, false, widget.TextAlignLeft)
}

// selRange returns the ordered [start,end) selection bounds.
func (t *TextareaWidget) selRange() (int, int) {
	if t.caret <= t.sel {
		return t.caret, t.sel
	}
	return t.sel, t.caret
}

// Event handles focus, mouse caret placement/drag, and keyboard editing.
func (t *TextareaWidget) Event(ctx widget.Context, e event.Event) bool {
	if t.disabled {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return t.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		if !t.IsFocused() {
			return false
		}
		return t.keyEvent(ctx, ev)
	}
	return false
}

func (t *TextareaWidget) mouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		ctx.SetCursor(widget.CursorText)
		return true
	case event.MouseLeave:
		ctx.SetCursor(widget.CursorDefault)
		t.dragging = false
		return true
	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		ctx.RequestFocus(t)
		off := t.offsetAt(e.Position)
		t.caret = off
		t.sel = off
		t.dragging = true
		t.invalidate(ctx)
		return true
	case event.MouseMove:
		if !t.dragging {
			return false
		}
		t.caret = t.offsetAt(e.Position)
		t.invalidate(ctx)
		return true
	case event.MouseRelease:
		t.dragging = false
		return true
	}
	return false
}

// offsetAt maps a screen-space point to the nearest rune offset.
func (t *TextareaWidget) offsetAt(p geometry.Point) int {
	bounds := t.Bounds()
	contentX := bounds.Min.X + metrics.Textarea.PadX + metrics.Textarea.BorderWidth
	contentY := bounds.Min.Y + metrics.Textarea.PadY + metrics.Textarea.BorderWidth
	lineH := metrics.Textarea.LineHeight
	rows := t.wrap.rows
	if len(rows) == 0 {
		return 0
	}
	ri := int((p.Y - contentY + t.scrollY) / lineH)
	if ri < 0 {
		ri = 0
	}
	if ri >= len(rows) {
		ri = len(rows) - 1
	}
	r := rows[ri]
	runes := []rune(t.resolvedText())
	family := t.fontFamily()
	size := metrics.Textarea.FontSize
	localX := p.X - contentX
	// Find the rune boundary closest to localX within the row.
	best := r.start
	for i := r.start; i <= r.end; i++ {
		w := textmetrics.Width(family, size, string(runes[r.start:i]))
		if w <= localX {
			best = i
		} else {
			// Choose the nearer of i-1 and i.
			prevW := textmetrics.Width(family, size, string(runes[r.start:i-1]))
			if localX-prevW <= w-localX {
				best = i - 1
			} else {
				best = i
			}
			return clampOffset(best, len(runes))
		}
	}
	// Past the last glyph: clamp to row end, excluding a trailing space on a
	// soft-wrapped row so the caret does not jump to the next row's start.
	if !r.hardBreak && r.end > r.start && runes[r.end-1] == ' ' {
		best = r.end - 1
	} else {
		best = r.end
	}
	return clampOffset(best, len(runes))
}

func (t *TextareaWidget) keyEvent(ctx widget.Context, e *event.KeyEvent) bool {
	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}
	shift := e.Modifiers().Has(event.ModShift)
	ctrl := e.Modifiers().IsCtrl()

	if ctrl {
		switch e.Key {
		case event.KeyA:
			runes := []rune(t.resolvedText())
			t.sel = 0
			t.caret = len(runes)
			t.invalidate(ctx)
			return true
		case event.KeyC:
			t.copySelection()
			return true
		case event.KeyX:
			t.copySelection()
			t.deleteSelection()
			t.scrollToCaret()
			t.invalidate(ctx)
			return true
		case event.KeyV:
			if t.clipboard != "" {
				t.insert(t.clipboard)
				t.scrollToCaret()
				t.invalidate(ctx)
			}
			return true
		}
	}

	switch e.Key {
	case event.KeyBackspace:
		t.backspace()
		t.scrollToCaret()
		t.invalidate(ctx)
		return true
	case event.KeyDelete:
		t.deleteForward()
		t.invalidate(ctx)
		return true
	case event.KeyEnter:
		t.insert("\n")
		t.scrollToCaret()
		t.invalidate(ctx)
		return true
	case event.KeyLeft:
		t.moveCaret(t.caret-1, shift)
		t.scrollToCaret()
		t.invalidate(ctx)
		return true
	case event.KeyRight:
		t.moveCaret(t.caret+1, shift)
		t.scrollToCaret()
		t.invalidate(ctx)
		return true
	case event.KeyUp:
		t.moveCaret(t.verticalOffset(-1), shift)
		t.scrollToCaret()
		t.invalidate(ctx)
		return true
	case event.KeyDown:
		t.moveCaret(t.verticalOffset(1), shift)
		t.scrollToCaret()
		t.invalidate(ctx)
		return true
	case event.KeyHome:
		t.moveCaret(t.rowEdge(false), shift)
		t.scrollToCaret()
		t.invalidate(ctx)
		return true
	case event.KeyEnd:
		t.moveCaret(t.rowEdge(true), shift)
		t.scrollToCaret()
		t.invalidate(ctx)
		return true
	}

	if e.Rune != 0 && e.Rune >= 0x20 {
		t.insert(string(e.Rune))
		t.scrollToCaret()
		t.invalidate(ctx)
		return true
	}
	return false
}

// insert replaces the current selection with s and advances the caret.
func (t *TextareaWidget) insert(s string) {
	runes := []rune(t.resolvedText())
	start, end := t.selRange()
	out := string(runes[:start]) + s + string(runes[end:])
	t.setText(out)
	t.caret = start + len([]rune(s))
	t.sel = t.caret
}

// backspace deletes the selection, or the rune before the caret.
func (t *TextareaWidget) backspace() {
	if t.caret != t.sel {
		t.deleteSelection()
		return
	}
	if t.caret == 0 {
		return
	}
	runes := []rune(t.resolvedText())
	out := string(runes[:t.caret-1]) + string(runes[t.caret:])
	pos := t.caret - 1
	t.setText(out)
	t.caret = pos
	t.sel = pos
}

// deleteForward deletes the selection, or the rune after the caret.
func (t *TextareaWidget) deleteForward() {
	if t.caret != t.sel {
		t.deleteSelection()
		return
	}
	runes := []rune(t.resolvedText())
	if t.caret >= len(runes) {
		return
	}
	out := string(runes[:t.caret]) + string(runes[t.caret+1:])
	t.setText(out)
}

// deleteSelection removes the selected range and collapses the caret to its
// start.
func (t *TextareaWidget) deleteSelection() {
	runes := []rune(t.resolvedText())
	start, end := t.selRange()
	out := string(runes[:start]) + string(runes[end:])
	t.setText(out)
	t.caret = start
	t.sel = start
}

// copySelection stores the selected text in the internal clipboard.
func (t *TextareaWidget) copySelection() {
	if t.caret == t.sel {
		return
	}
	runes := []rune(t.resolvedText())
	start, end := t.selRange()
	t.clipboard = string(runes[start:end])
}

// scrollToCaret adjusts scrollY so the caret row is visible within the
// control's content area. Called after every editing/navigation operation.
func (t *TextareaWidget) scrollToCaret() {
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}
	contentH := bounds.Height() - 2*(metrics.Textarea.PadY+metrics.Textarea.BorderWidth)
	if contentH <= 0 {
		return
	}
	_, row := t.caretXY(t.caret)
	lineH := metrics.Textarea.LineHeight
	caretTop := row * lineH
	caretBot := caretTop + lineH
	if caretTop < t.scrollY {
		t.scrollY = caretTop
	} else if caretBot > t.scrollY+contentH {
		t.scrollY = caretBot - contentH
	}
}

// moveCaret sets the caret, extending the selection when shift is held.
func (t *TextareaWidget) moveCaret(to int, shift bool) {
	to = clampOffset(to, len([]rune(t.resolvedText())))
	t.caret = to
	if !shift {
		t.sel = to
	}
}

// verticalOffset returns the caret offset one visual row up (-1) or down (+1),
// preserving the horizontal column position as closely as possible.
func (t *TextareaWidget) verticalOffset(dir int) int {
	rows := t.wrap.rows
	if len(rows) == 0 {
		return t.caret
	}
	x, rowF := t.caretXY(t.caret)
	target := int(rowF) + dir
	if target < 0 || target >= len(rows) {
		return t.caret
	}
	r := rows[target]
	runes := []rune(t.resolvedText())
	family := t.fontFamily()
	size := metrics.Textarea.FontSize
	best := r.start
	for i := r.start; i <= r.end; i++ {
		w := textmetrics.Width(family, size, string(runes[r.start:i]))
		if w <= x {
			best = i
		} else {
			break
		}
	}
	if !r.hardBreak && r.end > r.start && best == r.end && runes[r.end-1] == ' ' {
		best = r.end - 1
	}
	return clampOffset(best, len(runes))
}

// rowEdge returns the offset at the start (false) or end (true) of the caret's
// visual row.
func (t *TextareaWidget) rowEdge(end bool) int {
	rows := t.wrap.rows
	if len(rows) == 0 {
		return t.caret
	}
	_, rowF := t.caretXY(t.caret)
	r := rows[int(rowF)]
	if end {
		runes := []rune(t.resolvedText())
		if !r.hardBreak && r.end > r.start && runes[r.end-1] == ' ' {
			return r.end - 1
		}
		return r.end
	}
	return r.start
}

func (t *TextareaWidget) invalidate(ctx widget.Context) {
	t.SetNeedsRedraw(true)
	ctx.InvalidateRect(t.Bounds())
}

// IsFocusable reports whether the textarea can take keyboard focus.
func (t *TextareaWidget) IsFocusable() bool {
	return t.IsVisible() && t.IsEnabled() && !t.disabled
}

// SetFocused tracks the focused state. The caret + solid Ring border + focus
// ring appear whenever the control is focused (a text input shows its caret on
// click as well as on keyboard traversal); focusVisible mirrors IsFocused.
func (t *TextareaWidget) SetFocused(focused bool) {
	t.WidgetBase.SetFocused(focused)
	t.focusVisible = focused
	t.MarkRedrawLocal() // not SetNeedsRedraw: avoids context-lock re-entry in RequestFocus
}

// Mount binds the value signal so external writes repaint.
func (t *TextareaWidget) Mount(ctx widget.Context) {
	if sched := ctx.Scheduler(); sched != nil && t.valueSig != nil {
		t.AddBinding(state.BindToScheduler(t.valueSig, t, sched))
	}
}

// Unmount implements widget.Lifecycle; bindings are cleaned automatically.
func (t *TextareaWidget) Unmount() {}

// Children returns nil; the textarea is a leaf.
func (t *TextareaWidget) Children() []widget.Widget { return nil }

// AccessibilityRole returns the text-field role.
func (t *TextareaWidget) AccessibilityRole() a11y.Role { return a11y.RoleTextField }

// AccessibilityLabel returns the placeholder as the accessible label.
func (t *TextareaWidget) AccessibilityLabel() string { return t.placeholder }

// AccessibilityHint returns no hint.
func (t *TextareaWidget) AccessibilityHint() string { return "" }

// AccessibilityValue returns the current text content.
func (t *TextareaWidget) AccessibilityValue() string { return t.resolvedText() }

// AccessibilityState reports the disabled and focused states.
func (t *TextareaWidget) AccessibilityState() a11y.State {
	return a11y.State{Disabled: t.disabled, Focused: t.IsFocused()}
}

// AccessibilityActions returns the focus action.
func (t *TextareaWidget) AccessibilityActions() []a11y.Action {
	return []a11y.Action{a11y.ActionFocus}
}

func clampOffset(v, n int) int {
	if v < 0 {
		return 0
	}
	if v > n {
		return n
	}
	return v
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*TextareaWidget)(nil)
	_ widget.Focusable = (*TextareaWidget)(nil)
	_ widget.Lifecycle = (*TextareaWidget)(nil)
	_ a11y.Accessible  = (*TextareaWidget)(nil)
)

package painters

import (
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// PaintTextField renders a core/textfield.Widget in shadcn Input style
// (docs/research/03-shadcn-pixel-spec.md §5 "Input").
//
// WRAP DECISION: Input wraps core/textfield because the text-editing
// machinery (cursor, selection, insertion, password masking, signal sync) is
// substantial. core/textfield routes all visuals through this Painter and its
// PaintState carries NO ColorScheme — every color is resolved here from the
// painter's Theme via Active() at paint time, so light/dark switches repaint
// without reinstalling the painter. The control height (32px, h-8) is NOT
// reachable through core/textfield's fixed 48px Layout, so the graft
// InputWidget overrides Layout; this painter draws within whatever Bounds it
// is given.
//
// Recipe per the spec: radius LG, 1px Input border via draw.InsideBorder,
// transparent bg (dark: Input/30 via draw.MulAlpha), shadow-xs, font 14px
// Geist 400, placeholder MutedForeground, text Foreground, selection bg
// Primary + fg PrimaryForeground, caret 1px Foreground. Focus: solid Ring
// border + 3px Ring/50 ring. Invalid (HasError): Destructive border + ring
// Destructive/20 (dark /40). Disabled: Fade 0.5 on every color.
func (p TextField) PaintTextField(canvas widget.Canvas, st textfield.PaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	th := p.Theme
	tok := th.Active()
	dark := th.IsDark()
	bounds := st.Bounds
	radius := th.RadiusLG()
	disabled := st.Disabled

	// Shadow-xs sits under the fill (drawn first, only when not disabled —
	// shadcn dims the whole control, and a shadow under a faded control
	// would read as a halo).
	if !disabled {
		draw.Shadow(canvas, bounds, radius, metrics.ShadowXS)
	}

	// Background: transparent in light mode; dark adds Input/30.
	if dark {
		fill := draw.MulAlpha(tok.Input, metrics.Input.DarkFillAlpha)
		canvas.DrawRoundRect(bounds, draw.Fade(fill, disabled), radius)
	}

	// Border + focus/invalid ring.
	borderColor := tok.Input
	switch {
	case st.HasError:
		borderColor = tok.Destructive
		ringAlpha := metrics.InvalidRingAlphaLight
		if dark {
			ringAlpha = metrics.InvalidRingAlphaDark
		}
		draw.FocusRing(canvas, bounds, radius, draw.Alpha(tok.Destructive, ringAlpha))
	case st.Focused:
		borderColor = tok.Ring
		draw.FocusRing(canvas, bounds, radius, draw.Alpha(tok.Ring, metrics.RingAlpha))
	}
	draw.InsideBorder(canvas, bounds, radius, draw.Fade(borderColor, disabled), metrics.Input.BorderWidth)

	// Content area (inside px/py).
	content := geometry.NewRect(
		bounds.Min.X+metrics.Input.PadX,
		bounds.Min.Y+metrics.Input.PadY,
		bounds.Width()-2*metrics.Input.PadX,
		bounds.Height()-2*metrics.Input.PadY,
	)
	canvas.PushClip(content)
	defer canvas.PopClip()

	display := st.Text
	if st.InputType == textfield.TypePassword {
		display = maskPassword(len([]rune(st.Text)))
	}

	family := fonts.Family(metrics.Input.FontWeight)
	fontSize := metrics.Input.FontSize

	// Placeholder when empty and unfocused.
	if display == "" {
		if st.Placeholder != "" {
			drawFieldText(canvas, st.Placeholder, content, family, fontSize,
				draw.Fade(tok.MutedForeground, disabled))
		}
		paintCaret(canvas, st, content, "", family, fontSize, tok, disabled)
		return
	}

	// Selection highlight under the text.
	if st.SelectStart != st.SelectEnd && !disabled {
		paintFieldSelection(canvas, st, content, display, family, fontSize, tok)
	}

	drawFieldText(canvas, display, content, family, fontSize, draw.Fade(tok.Foreground, disabled))
	paintCaret(canvas, st, content, display, family, fontSize, tok, disabled)
}

// drawFieldText draws left-aligned single-line text via StyledTextDrawer,
// falling back to DrawText (weight 400 → not bold) on mock canvases.
func drawFieldText(canvas widget.Canvas, text string, bounds geometry.Rect, family string, size float32, col widget.Color) {
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

// measureFieldText measures a substring width via the canvas's styled or
// plain measurement (whichever the canvas implements).
func measureFieldText(canvas widget.Canvas, text string, family string, size float32) float32 {
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		return std.MeasureStyledText(text, widget.TextStyle{FontFamily: family, FontSize: size})
	}
	return canvas.MeasureText(text, size, false)
}

// paintFieldSelection draws the selection highlight in Primary with the
// selected glyphs re-drawn in PrimaryForeground on top.
func paintFieldSelection(canvas widget.Canvas, st textfield.PaintState, content geometry.Rect, display, family string, size float32, tok *theme.Tokens) {
	runes := []rune(display)
	start, end := st.SelectStart, st.SelectEnd
	if start > end {
		start, end = end, start
	}
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}
	x1 := content.Min.X + measureFieldText(canvas, string(runes[:start]), family, size)
	x2 := content.Min.X + measureFieldText(canvas, string(runes[:end]), family, size)
	if x2 > content.Max.X {
		x2 = content.Max.X
	}
	sel := geometry.NewRect(x1, content.Min.Y, x2-x1, content.Height())
	canvas.DrawRect(sel, tok.Primary)
}

// paintCaret draws a 1px Foreground caret at the cursor position when the
// field is focused, has no active selection, and is not disabled.
func paintCaret(canvas widget.Canvas, st textfield.PaintState, content geometry.Rect, display, family string, size float32, tok *theme.Tokens, disabled bool) {
	if !st.Focused || disabled {
		return
	}
	if st.SelectStart != st.SelectEnd {
		return
	}
	runes := []rune(display)
	pos := st.CursorPos
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}
	caretX := content.Min.X + measureFieldText(canvas, string(runes[:pos]), family, size)
	if caretX > content.Max.X {
		caretX = content.Max.X
	}
	w := metrics.Input.CaretWidth
	caret := geometry.NewRect(caretX, content.Min.Y, w, content.Height())
	canvas.DrawRect(caret, tok.Foreground)
}

// maskPassword returns a run of bullets of the given length.
func maskPassword(n int) string {
	runes := make([]rune, n)
	for i := range runes {
		runes[i] = '•'
	}
	return string(runes)
}

// Compile-time interface check.
var _ textfield.Painter = TextField{}

package painters

import (
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// Dropdown paints the shadcn Select trigger and menu. It implements
// core/dropdown.Painter so raw gogpu/ui users can wire it onto a
// core/dropdown.Widget; the graft Select widget (select.go) is an OWNED
// widget that drives layout/overlay and calls these same methods with the
// exact shadcn px because core/dropdown's Layout hardcodes a 48px trigger
// and 40px item rows that cannot reach shadcn's h-9/h-8 trigger or 32px
// item rows (decision recorded in select.go).
//
// Both methods draw purely relative to the bounds in the PaintState, so
// they stay pixel-faithful regardless of which widget supplies the bounds.
// All colors resolve from Theme.Active() at paint time. The
// DropdownColorScheme on the PaintState is intentionally ignored: graft
// always resolves colors from the theme (DESIGN.md §8 decision 13).

// invalid/extra trigger state is conveyed out-of-band through these flags,
// since core/dropdown.TriggerPaintState has no Invalid field. The graft
// Select widget sets them before calling PaintTrigger.
type dropdownExtra struct {
	invalid      bool
	focusVisible bool
}

// PaintTrigger draws the closed Select trigger in shadcn style. On the raw
// core/dropdown path there is no focus-visible distinction, so st.Focused
// drives the ring.
func (p Dropdown) PaintTrigger(canvas widget.Canvas, st *dropdown.TriggerPaintState) {
	p.paintTrigger(canvas, st, dropdownExtra{focusVisible: st.Focused})
}

// PaintTriggerEx is the graft-internal entry point that carries the extra
// state (invalid, focus-visible) the core PaintState cannot. Trigger height
// (h-9 vs h-8) is encoded in the bounds the caller supplies, so the painter
// does not need a small flag.
func (p Dropdown) PaintTriggerEx(canvas widget.Canvas, st *dropdown.TriggerPaintState, invalid, focusVisible bool) {
	p.paintTrigger(canvas, st, dropdownExtra{invalid: invalid, focusVisible: focusVisible})
}

func (p Dropdown) paintTrigger(canvas widget.Canvas, st *dropdown.TriggerPaintState, ex dropdownExtra) {
	if st.Bounds.IsEmpty() {
		return
	}
	th := dropdownResolveTheme(p.Theme)
	tok := th.Active()
	dark := th.IsDark()
	m := metrics.Select
	bounds := st.Bounds
	radius := th.RadiusLG() // trigger: rounded-lg
	disabled := st.Disabled

	// Shadow (shadow-xs) sits under the fill.
	if !disabled {
		draw.Shadow(canvas, bounds, radius, metrics.ShadowXS)
	}

	// Fill: light transparent; dark input/30, hover input/50.
	if dark {
		fill := draw.MulAlpha(tok.Input, 0.3)
		if st.Hovered && !disabled {
			fill = draw.MulAlpha(tok.Input, 0.5)
		}
		canvas.DrawRoundRect(bounds, draw.Fade(fill, disabled), radius)
	}

	// Border + focus ring.
	switch {
	case ex.invalid:
		ring := draw.Alpha(tok.Destructive, dropdownInvalidRingAlpha(dark))
		if ex.focusVisible {
			draw.FocusRing(canvas, bounds, radius, ring)
		}
		draw.InsideBorder(canvas, bounds, radius, draw.Fade(tok.Destructive, disabled), m.BorderWidth)
	case ex.focusVisible:
		draw.FocusRing(canvas, bounds, radius, draw.Alpha(tok.Ring, metrics.RingAlpha))
		draw.InsideBorder(canvas, bounds, radius, tok.Ring, m.BorderWidth)
	default:
		draw.InsideBorder(canvas, bounds, radius, draw.Fade(tok.Input, disabled), m.BorderWidth)
	}

	// Text (placeholder = muted-foreground, else foreground).
	textColor := tok.Foreground
	if st.IsPlaceholder {
		textColor = tok.MutedForeground
	}
	textColor = draw.Fade(textColor, disabled)
	textRect := geometry.NewRect(
		bounds.Min.X+m.TriggerPadX,
		bounds.Min.Y,
		bounds.Width()-2*m.TriggerPadX-m.ChevronSize-m.TriggerGap,
		bounds.Height(),
	)
	dropdownDrawText(canvas, th, st.SelectedText, textRect, m.FontSize, m.FontWeight, textColor)

	// Chevron-down at the right, 16px, muted-foreground @ 50% opacity.
	chevSize := m.ChevronSize
	chevX := bounds.Max.X - m.TriggerPadX - chevSize
	chevY := bounds.Center().Y - chevSize/2
	chevColor := draw.Fade(draw.MulAlpha(tok.MutedForeground, m.ChevronOpacity), disabled)
	icon.Draw(canvas, icons.ChevronDown, geometry.NewRect(chevX, chevY, chevSize, chevSize), chevColor)
}

// PaintMenu draws the open Select menu. graft takes full control of the row
// anatomy (left text, check on the right) here, so the core menuWidget's
// left-aligned default layout is never used. ItemHeight is honored from the
// PaintState; the graft Select widget sets it to metrics.Select.ItemHeight.
func (p Dropdown) PaintMenu(canvas widget.Canvas, st *dropdown.MenuPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}
	th := dropdownResolveTheme(p.Theme)
	tok := th.Active()
	m := metrics.Select
	bounds := st.Bounds
	radius := th.RadiusMD()

	// Content surface: shadow-md, bg-popover, 1px border.
	draw.Shadow(canvas, bounds, radius, metrics.ShadowMD)
	canvas.DrawRoundRect(bounds, tok.Popover, radius)
	draw.InsideBorder(canvas, bounds, radius, tok.Border, m.BorderWidth)

	canvas.PushClip(bounds)
	defer canvas.PopClip()

	itemH := st.ItemHeight
	if itemH <= 0 {
		itemH = m.ItemHeight
	}
	innerLeft := bounds.Min.X + m.ContentPad
	innerRight := bounds.Max.X - m.ContentPad
	itemRadius := th.RadiusSM()

	endIndex := st.ScrollOffset + st.VisibleCount
	if endIndex > len(st.Items) {
		endIndex = len(st.Items)
	}

	for i := st.ScrollOffset; i < endIndex; i++ {
		item := st.Items[i]
		row := i - st.ScrollOffset
		top := bounds.Min.Y + m.ContentPad + float32(row)*itemH
		itemRect := geometry.NewRect(innerLeft, top, innerRight-innerLeft, itemH)
		disabled := item.Disabled

		// Highlight fill (focus/hover = solid accent).
		if i == st.HighlightedIndex && !disabled {
			canvas.DrawRoundRect(itemRect, tok.Accent, itemRadius)
		}

		// Item label.
		labelColor := tok.PopoverForeground
		if i == st.HighlightedIndex && !disabled {
			labelColor = tok.AccentForeground
		}
		labelColor = draw.Fade(labelColor, disabled)
		textRect := geometry.NewRect(
			itemRect.Min.X+m.ItemPadLeft,
			itemRect.Min.Y,
			itemRect.Width()-m.ItemPadLeft-m.ItemPadRight,
			itemRect.Height(),
		)
		dropdownDrawText(canvas, th, item.DisplayText(), textRect, m.FontSize, m.FontWeight, labelColor)

		// Selected check on the right.
		if i == st.SelectedIndex {
			checkX := itemRect.Max.X - m.CheckRight - m.CheckSize
			checkY := itemRect.Center().Y - m.CheckSize/2
			icon.Draw(canvas, icons.Check,
				geometry.NewRect(checkX, checkY, m.CheckSize, m.CheckSize),
				draw.Fade(labelColor, disabled))
		}
	}
}

// invalidRingAlpha returns the destructive ring alpha for the given mode.
func dropdownInvalidRingAlpha(dark bool) float32 {
	if dark {
		return metrics.InvalidRingAlphaDark
	}
	return metrics.InvalidRingAlphaLight
}

// resolveTheme returns t or the process default if t is nil.
func dropdownResolveTheme(t *theme.Theme) *theme.Theme {
	if t != nil {
		return t
	}
	return theme.New()
}

// drawText renders text through the canvas StyledTextDrawer capability,
// falling back to plain DrawText (bold at weight >= 600) for canvases
// without it (mock canvas).
func dropdownDrawText(canvas widget.Canvas, th *theme.Theme, text string, bounds geometry.Rect, size float32, weight int, color widget.Color) {
	if text == "" {
		return
	}
	family := fonts.Family(weight)
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(text, bounds, widget.TextStyle{
			FontFamily: family,
			FontSize:   size,
			Color:      color,
			Align:      widget.TextAlignLeft,
		})
		return
	}
	canvas.DrawText(text, bounds, size, color, weight >= 600, widget.TextAlignLeft)
}

// Compile-time interface check.
var _ dropdown.Painter = Dropdown{}

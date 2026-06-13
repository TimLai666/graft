package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// KbdWidget is the shadcn Kbd: one or more keyboard-key chips. Each key is a
// 20px-high muted-background pill with 12px/500 muted-foreground text in the
// SANS family (shadcn uses font-sans, not mono); multiple keys render as a
// gap-1 row of chips (the shadcn KbdGroup pattern).
//
// Architecture decision: graft-OWNED leaf. It is inert (pointer-events-none,
// select-none) and draws its chips directly on the canvas via metrics/kbd.go.
//
//	graft.Kbd("⌘")            // single key
//	graft.Kbd("Ctrl", "B")    // combo: two chips
type KbdWidget struct {
	widget.WidgetBase

	keys  []string
	theme *theme.Theme

	// keyW caches each chip's width from Layout (in widget-local order).
	keyW []float32
}

// Kbd creates a key-chip row from the given key labels. A single label
// renders one chip; multiple labels render a gap-1 row.
func Kbd(keys ...string) *KbdWidget {
	k := &KbdWidget{
		keys:  keys,
		theme: CurrentTheme(),
	}
	k.SetVisible(true)
	k.SetEnabled(true)
	return k
}

// Theme pins a specific theme instead of the snapshotted current theme.
func (k *KbdWidget) Theme(th *theme.Theme) *KbdWidget {
	if th != nil {
		k.theme = th
	}
	return k
}

// fontFamily resolves the key-label family (font-sans, honoring theme fonts).
func (k *KbdWidget) fontFamily() string {
	if k.theme.FontSans != theme.DefaultFontSans {
		return k.theme.FontSans
	}
	return fonts.Family(metrics.Kbd.FontWeight)
}

// chipWidth returns the chip width for one key label: text + 2*px-1, clamped
// up to min-w-5.
func (k *KbdWidget) chipWidth(label string) float32 {
	m := metrics.Kbd
	w := textmetrics.Width(k.fontFamily(), m.FontSize, label) + 2*m.PadX
	if w < m.MinWidth {
		w = m.MinWidth
	}
	return w
}

// Layout sizes the chip row: sum of chip widths + gap-1 between them, fixed
// 20px height.
func (k *KbdWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	m := metrics.Kbd
	k.keyW = k.keyW[:0]
	var x float32
	for i, key := range k.keys {
		if i > 0 {
			x += m.Gap
		}
		w := k.chipWidth(key)
		k.keyW = append(k.keyW, w)
		x += w
	}
	size := c.Constrain(geometry.Sz(x, m.Height))
	k.SetBounds(geometry.FromPointSize(k.Position(), size))
	return size
}

// Draw paints each key chip (muted rounded-sm pill) with its centered
// muted-foreground label, resolving colors at draw time.
func (k *KbdWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !k.IsVisible() {
		return
	}
	m := metrics.Kbd
	th := k.theme
	tok := th.Active()
	bounds := k.Bounds()
	radius := th.RadiusSM()
	family := k.fontFamily()

	x := bounds.Min.X
	for i, key := range k.keys {
		if i > 0 {
			x += m.Gap
		}
		w := m.MinWidth
		if i < len(k.keyW) {
			w = k.keyW[i]
		}
		chip := geometry.NewRect(x, bounds.Min.Y, w, m.Height)
		canvas.DrawRoundRect(chip, tok.Muted, radius)

		textRect := geometry.NewRect(
			chip.Min.X,
			chip.Min.Y+(m.Height-m.LineHeight)/2,
			chip.Width(),
			m.LineHeight,
		)
		if std, ok := canvas.(widget.StyledTextDrawer); ok {
			std.DrawStyledText(key, textRect, widget.TextStyle{
				FontFamily: family,
				FontSize:   m.FontSize,
				Color:      tok.MutedForeground,
				Align:      widget.TextAlignCenter,
			})
		} else {
			canvas.DrawText(key, textRect, m.FontSize, tok.MutedForeground,
				m.FontWeight >= 600, widget.TextAlignCenter)
		}
		x += w
	}
}

// Event ignores all input (pointer-events-none).
func (k *KbdWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; KbdWidget is a leaf.
func (k *KbdWidget) Children() []widget.Widget { return nil }

// Compile-time interface check.
var _ widget.Widget = (*KbdWidget)(nil)

package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/theme"
)

// TypographyWidget is graft's themed text leaf: exact font family per
// weight (Geist 400/500/600/700), exact px sizes and line heights, and
// token-based colors resolved from the theme at draw time so mode switches
// repaint without rebuilding the tree.
//
// It is the building block for component text (CardTitle, DialogTitle,
// Label, ...) and for the shadcn typography scale (H1..H4, P, Lead, ...).
//
// v1 renders a single line (shadcn component text is single-line);
// wrapping paragraphs are planned alongside Textarea.
type TypographyWidget struct {
	widget.WidgetBase

	content    string
	size       float32
	weight     int
	lineHeight float32 // line box height in px; 0 means equal to size (leading-none)
	mono       bool
	align      widget.TextAlign
	alpha      float32

	// colorFn picks the text color from the active token set. nil means
	// Foreground. fixed overrides both when non-nil.
	colorFn func(*theme.Tokens) widget.Color
	fixed   *widget.Color

	// Optional chip background (InlineCode, Kbd-style uses).
	bgFn       func(*theme.Tokens) widget.Color
	bgRadius   float32
	padX, padY float32

	theme *theme.Theme
}

// Text creates body text: 14px / weight 400 / 20px line height (text-sm),
// in the foreground color.
func Text(content string) *TypographyWidget {
	t := &TypographyWidget{
		content:    content,
		size:       14,
		weight:     400,
		lineHeight: 20,
		alpha:      1,
		align:      widget.TextAlignLeft,
	}
	t.SetVisible(true)
	t.SetEnabled(true)
	return t
}

// styled is the internal shorthand used by typography constructors.
func styled(content string, size float32, weight int, lineHeight float32) *TypographyWidget {
	t := Text(content)
	t.size = size
	t.weight = weight
	t.lineHeight = lineHeight
	return t
}

// H1 renders shadcn's h1 scale: 36px / extra-bold / tight (48px line).
// Geist ships up to 700 in graft, so extra-bold resolves to Bold.
func H1(content string) *TypographyWidget { return styled(content, 36, 800, 40) }

// H2 renders shadcn's h2 scale: 30px / semibold (36px line).
func H2(content string) *TypographyWidget { return styled(content, 30, 600, 36) }

// H3 renders shadcn's h3 scale: 24px / semibold (32px line).
func H3(content string) *TypographyWidget { return styled(content, 24, 600, 32) }

// H4 renders shadcn's h4 scale: 20px / semibold (28px line).
func H4(content string) *TypographyWidget { return styled(content, 20, 600, 28) }

// P renders paragraph text: 16px / 400 / 28px line (leading-7).
func P(content string) *TypographyWidget { return styled(content, 16, 400, 28) }

// Lead renders intro text: 20px / 400, muted foreground.
func Lead(content string) *TypographyWidget {
	return styled(content, 20, 400, 28).Muted()
}

// Large renders 18px / semibold text.
func Large(content string) *TypographyWidget { return styled(content, 18, 600, 28) }

// Small renders 14px / medium / leading-none text.
func Small(content string) *TypographyWidget { return styled(content, 14, 500, 14) }

// MutedText renders 14px muted-foreground text (shadcn "muted").
func MutedText(content string) *TypographyWidget { return Text(content).Muted() }

// InlineCode renders mono text-sm semibold in a muted chip:
// bg-muted, rounded, px 4.8 / py 3.2 (0.3rem / 0.2rem).
func InlineCode(content string) *TypographyWidget {
	t := styled(content, 14, 600, 20)
	t.mono = true
	t.bgFn = func(tok *theme.Tokens) widget.Color { return tok.Muted }
	t.bgRadius = 4
	t.padX, t.padY = 4.8, 3.2
	return t
}

// FontSize overrides the font size in px.
func (t *TypographyWidget) FontSize(px float32) *TypographyWidget {
	t.size = px
	return t
}

// Weight sets the font weight (400/500/600/700; nearest face wins).
func (t *TypographyWidget) Weight(w int) *TypographyWidget {
	t.weight = w
	return t
}

// LineHeight sets the line box height in px. Zero means leading-none
// (line box equals font size).
func (t *TypographyWidget) LineHeight(px float32) *TypographyWidget {
	t.lineHeight = px
	return t
}

// LeadingNone sets line-height equal to the font size (leading-none).
func (t *TypographyWidget) LeadingNone() *TypographyWidget {
	t.lineHeight = 0
	return t
}

// Muted colors the text with the muted-foreground token.
func (t *TypographyWidget) Muted() *TypographyWidget {
	return t.ColorToken(func(tok *theme.Tokens) widget.Color { return tok.MutedForeground })
}

// Destructive colors the text with the destructive token.
func (t *TypographyWidget) Destructive() *TypographyWidget {
	return t.ColorToken(func(tok *theme.Tokens) widget.Color { return tok.Destructive })
}

// ColorToken selects the text color from the active token set at draw
// time (survives light/dark switches).
func (t *TypographyWidget) ColorToken(fn func(*theme.Tokens) widget.Color) *TypographyWidget {
	t.colorFn = fn
	return t
}

// Color pins a fixed text color (escape hatch; does not adapt to mode).
func (t *TypographyWidget) Color(c widget.Color) *TypographyWidget {
	t.fixed = &c
	return t
}

// Opacity multiplies the resolved color's alpha (e.g. 0.7 for the dialog
// close button at rest).
func (t *TypographyWidget) Opacity(a float32) *TypographyWidget {
	t.alpha = a
	return t
}

// Mono switches to the monospace family.
func (t *TypographyWidget) Mono() *TypographyWidget {
	t.mono = true
	return t
}

// Align sets the horizontal text alignment within the line box.
func (t *TypographyWidget) Align(a widget.TextAlign) *TypographyWidget {
	t.align = a
	return t
}

// Theme pins a specific theme instead of the process-wide current theme.
func (t *TypographyWidget) Theme(th *theme.Theme) *TypographyWidget {
	t.theme = th
	return t
}

// Content returns the current text content.
func (t *TypographyWidget) Content() string { return t.content }

// SetContent replaces the text and requests a redraw.
func (t *TypographyWidget) SetContent(s string) {
	if t.content == s {
		return
	}
	t.content = s
	t.SetNeedsRedraw(true)
}

func (t *TypographyWidget) resolvedTheme() *theme.Theme {
	if t.theme != nil {
		return t.theme
	}
	return CurrentTheme()
}

// family resolves the registered font family for this widget's weight,
// honoring custom theme fonts (custom families register a single face, so
// weight mapping only applies to the stock Geist families).
func (t *TypographyWidget) family(th *theme.Theme) string {
	if t.mono {
		if th.FontMono != theme.DefaultFontMono {
			return th.FontMono
		}
		return fonts.MonoFamily(t.weight)
	}
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(t.weight)
}

func (t *TypographyWidget) lineBox() float32 {
	if t.lineHeight > 0 {
		return t.lineHeight
	}
	return t.size
}

// Layout measures the single-line text via sfnt advances.
func (t *TypographyWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	th := t.resolvedTheme()
	w := textmetrics.Width(t.family(th), t.size, t.content) + 2*t.padX
	h := t.lineBox() + 2*t.padY
	size := c.Constrain(geometry.Sz(w, h))
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

// Draw renders the chip background (if any) and the text through the
// canvas's StyledTextDrawer capability, falling back to plain DrawText
// (bold at weight >= 600) on canvases without it.
func (t *TypographyWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !t.IsVisible() {
		return
	}
	th := t.resolvedTheme()
	tok := th.Active()
	bounds := t.Bounds()

	if t.bgFn != nil {
		canvas.DrawRoundRect(bounds, t.bgFn(tok), t.bgRadius)
	}

	col := tok.Foreground
	switch {
	case t.fixed != nil:
		col = *t.fixed
	case t.colorFn != nil:
		col = t.colorFn(tok)
	}
	if t.alpha < 1 {
		col = draw.MulAlpha(col, t.alpha)
	}

	textBounds := bounds
	if t.padX != 0 || t.padY != 0 {
		textBounds = geometry.NewRect(
			bounds.Min.X+t.padX,
			bounds.Min.Y+t.padY,
			bounds.Width()-2*t.padX,
			bounds.Height()-2*t.padY,
		)
	}

	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(t.content, textBounds, widget.TextStyle{
			FontFamily: t.family(th),
			FontSize:   t.size,
			Color:      col,
			Align:      t.align,
		})
		return
	}
	canvas.DrawText(t.content, textBounds, t.size, col, t.weight >= 600, t.align)
}

// Event ignores all input; text is inert.
func (t *TypographyWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; TypographyWidget is a leaf.
func (t *TypographyWidget) Children() []widget.Widget { return nil }

var _ widget.Widget = (*TypographyWidget)(nil)

package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// AlertWidget is graft's Alert composite: a bordered Card-colored panel with
// an optional leading lucide icon, a title, and a description, matching the
// shadcn new-york-v4 Alert (docs/research/03-shadcn-pixel-spec.md §5).
//
// It is a graft-OWNED widget rather than a primitives.Box composite because
// the chrome (background, 1px inside border, rounded-lg corners) and the
// destructive color routing must live-resolve from the active theme tokens
// inside Draw, and because the two-column grid (fixed 16px icon column +
// content column) is simpler to lay out directly than with nested boxes.
//
// Build it with the content-first constructor and the AlertTitle /
// AlertDescription sub-constructors:
//
//	graft.Alert(
//	    graft.AlertTitle("Heads up!"),
//	    graft.AlertDescription("You can add components to your app."),
//	).Icon(icons.Info)
type AlertWidget struct {
	widget.WidgetBase

	title       *TypographyWidget
	description *TypographyWidget

	hasIcon bool
	ic      icon.IconData

	destructive bool
	theme       *theme.Theme
}

// Alert builds an alert from AlertTitle / AlertDescription children (either,
// both, or neither). Extra non-title/description children are ignored;
// shadcn's Alert holds exactly a title and a description slot.
func Alert(children ...widget.Widget) *AlertWidget {
	a := &AlertWidget{theme: CurrentTheme()}
	a.SetVisible(true)
	a.SetEnabled(true)
	for _, c := range children {
		switch w := c.(type) {
		case *alertTitleWidget:
			a.title = w.TypographyWidget
		case *alertDescriptionWidget:
			a.description = w.TypographyWidget
		}
	}
	return a
}

// alertTitleWidget and alertDescriptionWidget tag a TypographyWidget so the
// Alert constructor can route children to the right slot without exporting
// new types. They embed *TypographyWidget so they remain usable widgets.
type alertTitleWidget struct{ *TypographyWidget }

type alertDescriptionWidget struct{ *TypographyWidget }

// AlertTitle creates the alert title: 14px / weight 500, in the inherited
// foreground (CardForeground via the Alert, Destructive in the destructive
// variant). The Alert resolves the color, so the returned widget is a plain
// medium-weight text leaf.
//
// Note: shadcn's title adds tracking-tight (letter-spacing −0.025em); the
// gogpu/ui text pipeline does not support letter-spacing, so graft renders
// the title at default tracking (see the concerns note).
func AlertTitle(text string) *alertTitleWidget {
	t := styled(text, metrics.Alert.TitleFontSize, metrics.Alert.TitleWeight, metrics.Alert.TitleLineHeight)
	return &alertTitleWidget{t}
}

// AlertDescription creates the alert description: 14px muted-foreground text
// (Destructive at 90% in the destructive variant; the Alert applies that).
func AlertDescription(text string) *alertDescriptionWidget {
	t := styled(text, metrics.Alert.DescFontSize, 400, metrics.Alert.DescLineHeight).Muted()
	return &alertDescriptionWidget{t}
}

// Icon sets the leading lucide icon (16px, top-left, nudged down 2px). Pass
// an icons.* value (e.g. icons.Info, icons.CircleAlert).
func (a *AlertWidget) Icon(ic icon.IconData) *AlertWidget {
	a.ic = ic
	a.hasIcon = true
	return a
}

// Destructive switches the alert to the destructive variant: the title and
// icon use the Destructive token and the description uses Destructive at 90%.
func (a *AlertWidget) Destructive() *AlertWidget {
	a.destructive = true
	return a
}

// Theme pins a specific theme instead of the process-wide current theme.
func (a *AlertWidget) Theme(th *theme.Theme) *AlertWidget {
	a.theme = th
	if a.title != nil {
		a.title.Theme(th)
	}
	if a.description != nil {
		a.description.Theme(th)
	}
	return a
}

func (a *AlertWidget) resolvedTheme() *theme.Theme {
	if a.theme != nil {
		return a.theme
	}
	return CurrentTheme()
}

// contentLeft returns the x offset of the content column relative to the
// alert's inner (padded) origin: 0 with no icon, IconColumn+IconGap with one.
func (a *AlertWidget) contentLeft() float32 {
	if !a.hasIcon {
		return 0
	}
	return metrics.Alert.IconColumn + metrics.Alert.IconGap
}

// Layout sizes the alert: width fills the available space (w-full), height is
// padding + the stacked title/description line boxes + the inter-row gap.
func (a *AlertWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	// Content column width available for the title/description text.
	availW := c.MaxWidth
	if availW <= 0 || availW > 100000 {
		availW = 0
	}
	innerW := availW - 2*metrics.Alert.PadX
	contentW := innerW - a.contentLeft()
	if contentW < 0 {
		contentW = 0
	}

	contentLoose := geometry.Loose(geometry.Sz(contentW, 100000))

	var contentH float32
	rows := 0
	if a.title != nil {
		sz := a.title.Layout(ctx, contentLoose)
		contentH += sz.Height
		rows++
	}
	if a.description != nil {
		sz := a.description.Layout(ctx, contentLoose)
		contentH += sz.Height
		rows++
	}
	if rows == 2 {
		contentH += metrics.Alert.RowGap
	}
	// With an icon and a single short row, the 16px icon can be taller than
	// the content; reserve at least the icon height so it is not clipped.
	if a.hasIcon && contentH < metrics.Alert.IconSize {
		contentH = metrics.Alert.IconSize
	}

	w := availW
	if w <= 0 {
		// Unbounded width: fall back to intrinsic content width.
		var maxRow float32
		if a.title != nil {
			maxRow = a.title.Bounds().Width()
		}
		if a.description != nil && a.description.Bounds().Width() > maxRow {
			maxRow = a.description.Bounds().Width()
		}
		w = maxRow + a.contentLeft() + 2*metrics.Alert.PadX
	}
	h := contentH + 2*metrics.Alert.PadY

	size := c.Constrain(geometry.Sz(w, h))
	a.SetBounds(geometry.FromPointSize(a.Position(), size))

	// Position children in parent-local (alert) coordinates.
	contentX := metrics.Alert.PadX + a.contentLeft()
	y := metrics.Alert.PadY
	if a.title != nil {
		ts := a.title.Bounds().Size()
		a.title.SetBounds(geometry.NewRect(contentX, y, contentW, ts.Height))
		y += ts.Height
		if a.description != nil {
			y += metrics.Alert.RowGap
		}
	}
	if a.description != nil {
		ds := a.description.Bounds().Size()
		a.description.SetBounds(geometry.NewRect(contentX, y, contentW, ds.Height))
	}
	return size
}

// Draw paints the chrome (shadow-free Card-colored fill, 1px inside border,
// rounded-lg corners), the leading icon, and the title/description children.
// All colors resolve from the active token set so mode switches repaint.
func (a *AlertWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !a.IsVisible() {
		return
	}
	th := a.resolvedTheme()
	tok := th.Active()
	bounds := a.Bounds()
	radius := th.RadiusLG()

	// Chrome: Card background + 1px inside border (no shadow on Alert).
	canvas.DrawRoundRect(bounds, tok.Card, radius)
	draw.InsideBorder(canvas, bounds, radius, tok.Border, metrics.Alert.BorderWidth)

	// Title / icon color: CardForeground (default) or Destructive.
	fg := tok.CardForeground
	if a.destructive {
		fg = tok.Destructive
	}

	canvas.PushTransform(bounds.Min)

	// Leading icon: 16px box at the inner top-left, nudged down 2px.
	if a.hasIcon {
		ix := metrics.Alert.PadX
		iy := metrics.Alert.PadY + metrics.Alert.IconNudgeY
		iconRect := geometry.NewRect(ix, iy, metrics.Alert.IconSize, metrics.Alert.IconSize)
		icon.Draw(canvas, a.ic, iconRect, fg)
	}

	if a.title != nil {
		a.title.ColorToken(func(*theme.Tokens) widget.Color { return fg })
		widget.StampScreenOrigin(a.title, canvas)
		widget.DrawChild(a.title, ctx, canvas)
	}
	if a.description != nil {
		if a.destructive {
			a.description.ColorToken(func(t *theme.Tokens) widget.Color {
				return draw.Alpha(t.Destructive, metrics.Alert.DescDestructiveAlpha)
			})
		} else {
			a.description.ColorToken(func(t *theme.Tokens) widget.Color { return t.MutedForeground })
		}
		widget.StampScreenOrigin(a.description, canvas)
		widget.DrawChild(a.description, ctx, canvas)
	}

	canvas.PopTransform()
}

// Event ignores input; the alert is a static surface.
func (a *AlertWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns the title and description leaves for tree traversal.
func (a *AlertWidget) Children() []widget.Widget {
	var out []widget.Widget
	if a.title != nil {
		out = append(out, a.title)
	}
	if a.description != nil {
		out = append(out, a.description)
	}
	return out
}

var _ widget.Widget = (*AlertWidget)(nil)

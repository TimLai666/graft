package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// CardWidget is shadcn's Card: a rounded-xl surface in the card token with
// a 1px border, vertical padding 16, internal gap 16, and shadow-sm.
//
// Architecture: graft-owned widget. Layout composes primitives.Box for the
// section stack, but the card chrome (background, radius, border, shadow)
// is drawn by CardWidget itself so every color resolves from the active
// token set at draw time — primitives.Box caches its Background color and
// would go stale on a light/dark mode switch.
type CardWidget struct {
	widget.WidgetBase

	box   *primitives.BoxWidget
	theme *theme.Theme
}

// Card creates a card around the given sections (CardHeader, CardContent,
// CardFooter, or any widget). Width is natural unless W is set.
func Card(children ...Widget) *CardWidget {
	c := &CardWidget{
		box: primitives.VBox(children...).
			PaddingTop(metrics.CardPadY).
			PaddingBottom(metrics.CardPadY).
			Gap(metrics.CardGap).
			CrossAlign(primitives.CrossAxisStretch),
	}
	c.box.SetParent(c)
	c.SetVisible(true)
	c.SetEnabled(true)
	return c
}

// W pins the card's outer width in px (border included, like CSS
// border-box).
func (c *CardWidget) W(px float32) *CardWidget {
	inner := px - 2*metrics.CardBorderWidth
	if inner < 0 {
		inner = 0
	}
	c.box.Width(inner)
	return c
}

// Theme pins a specific theme instead of the process-wide current theme.
func (c *CardWidget) Theme(th *theme.Theme) *CardWidget {
	c.theme = th
	return c
}

func (c *CardWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// Layout sizes the section stack inside the 1px border (CSS border-box)
// and adds the border back to the outer size.
func (c *CardWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	bw := metrics.CardBorderWidth
	inner := cons.Loosen().Deflate(geometry.UniformInsets(bw))
	size := c.box.Layout(ctx, inner)
	c.box.SetBounds(geometry.FromPointSize(geometry.Pt(bw, bw), size))

	outer := cons.Constrain(geometry.Sz(size.Width+2*bw, size.Height+2*bw))
	c.SetBounds(geometry.FromPointSize(c.Position(), outer))
	return outer
}

// Draw paints shadow-sm, the card fill, the sections, and the 1px inside
// border, resolving all colors from the active token set.
func (c *CardWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.resolvedTheme()
	tok := th.Active()
	bounds := c.Bounds()
	radius := th.RadiusXL() // rounded-xl

	draw.Shadow(canvas, bounds, radius, metrics.ShadowSM) // shadow-sm
	canvas.DrawRoundRect(bounds, tok.Card, radius)        // bg-card

	canvas.PushTransform(bounds.Min)
	widget.StampScreenOrigin(c.box, canvas)
	widget.DrawChild(c.box, ctx, canvas)
	canvas.PopTransform()

	draw.InsideBorder(canvas, bounds, radius, tok.Border, metrics.CardBorderWidth) // border
}

// Event forwards input to the card sections with card-local coordinates.
func (c *CardWidget) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.MouseEvent:
		local := *ev
		local.Position = ev.Position.Sub(c.Bounds().Min)
		if !c.box.Bounds().Contains(local.Position) {
			return false
		}
		return c.box.Event(ctx, &local)
	case *event.WheelEvent:
		local := *ev
		local.Position = ev.Position.Sub(c.Bounds().Min)
		return c.box.Event(ctx, &local)
	default:
		return c.box.Event(ctx, e)
	}
}

// Children returns the internal section stack.
func (c *CardWidget) Children() []widget.Widget { return []widget.Widget{c.box} }

// CardSectionWidget is a card section (header, content, or footer): a thin
// named wrapper over primitives.Box carrying the px-16 shadcn paddings.
// All Box builder methods (Gap, CrossAlign, ...) remain available.
type CardSectionWidget struct {
	*primitives.BoxWidget
}

// CardActionWidget marks a widget as the card header action so CardHeader
// lays it out top-right (shadcn's data-slot="card-action" grid cell).
type CardActionWidget struct {
	*primitives.BoxWidget
}

// CardHeader creates the header section: px 16, vertical gap 4 between
// title and description. If a CardAction child is present it is laid out
// top-right with the remaining children stacked to its left.
func CardHeader(children ...Widget) *CardSectionWidget {
	var action *CardActionWidget
	rest := make([]Widget, 0, len(children))
	for _, ch := range children {
		if a, ok := ch.(*CardActionWidget); ok && action == nil {
			action = a
			continue
		}
		rest = append(rest, ch)
	}

	if action == nil {
		return &CardSectionWidget{
			primitives.VBox(rest...).
				Gap(metrics.CardHeaderGap).
				PaddingLeft(metrics.CardSectionPadX).
				PaddingRight(metrics.CardSectionPadX),
		}
	}

	titles := primitives.VBox(rest...).Gap(metrics.CardHeaderGap)
	return &CardSectionWidget{
		primitives.HBox(primitives.Expanded(titles), action).
			PaddingLeft(metrics.CardSectionPadX).
			PaddingRight(metrics.CardSectionPadX),
	}
}

// CardTitle creates the card title: 16px / weight 500 / leading-snug, in
// the card-foreground token.
func CardTitle(text string) *TypographyWidget {
	return styled(text, metrics.CardTitleFontSize, metrics.CardTitleFontWeight, metrics.CardTitleLineHeight).
		ColorToken(func(tok *theme.Tokens) widget.Color { return tok.CardForeground })
}

// CardDescription creates the card description: 14px muted-foreground.
func CardDescription(text string) *TypographyWidget {
	return styled(text, metrics.CardDescriptionFontSize, 400, metrics.CardDescriptionLineHeight).Muted()
}

// CardAction wraps a widget (button, badge, ...) for top-right placement
// inside CardHeader.
func CardAction(child Widget) *CardActionWidget {
	return &CardActionWidget{primitives.Box(child)}
}

// CardContent creates the content section: px 16.
func CardContent(children ...Widget) *CardSectionWidget {
	return &CardSectionWidget{
		primitives.VBox(children...).
			PaddingLeft(metrics.CardSectionPadX).
			PaddingRight(metrics.CardSectionPadX),
	}
}

// CardFooter creates the footer section: a horizontal row, px 16, items
// centered. Use the inherited Gap method for spacing between actions.
func CardFooter(children ...Widget) *CardSectionWidget {
	return &CardSectionWidget{
		primitives.HBox(children...).
			PaddingLeft(metrics.CardSectionPadX).
			PaddingRight(metrics.CardSectionPadX).
			CrossAlign(primitives.CrossAxisCenter),
	}
}

var (
	_ widget.Widget = (*CardWidget)(nil)
	_ widget.Widget = (*CardSectionWidget)(nil)
	_ widget.Widget = (*CardActionWidget)(nil)
)

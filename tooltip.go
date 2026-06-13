package graft

import (
	"time"

	"github.com/gogpu/ui/core/popover"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// TooltipWidget wraps a target widget and shows a small inverted bubble near
// it on hover (DESIGN.md §4, docs/research/03-shadcn-pixel-spec.md §Tooltip).
//
// Owned widget (not a wrap of core/popover.Tooltip): shadcn's bubble inverts
// colors (bg-foreground / text-background), carries a rotated-diamond arrow,
// and has exact px-3/py-1.5 padding — none of which core/popover.Tooltip's
// painter-driven surface exposes through Layout. The overlay is shown via
// ctx.OverlayManager() positioned with popover.CalculatePosition (Report 2
// §4.4 recipe), placement Top center, flipping when there is no room above.
//
// shadcn ships delayDuration 0; graft adds an ~80ms anti-flicker grace period
// (metrics.TooltipDelayMillis) so brushing past a target does not flash.
type TooltipWidget struct {
	widget.WidgetBase

	target  widget.Widget
	content *TooltipContentWidget
	theme   *theme.Theme

	hovered    bool
	hoverSince time.Time
	overlay    widget.Widget // non-nil while shown
}

// Tooltip wraps target so a hover reveals the given text in a small inverted
// bubble positioned above the target (auto-flipped below when cramped).
func Tooltip(target widget.Widget, text string) *TooltipWidget {
	t := &TooltipWidget{
		target:  target,
		content: newTooltipContent(text),
		theme:   CurrentTheme(),
	}
	t.SetVisible(true)
	t.SetEnabled(true)
	if target != nil {
		t.AddChild(target)
	}
	return t
}

// Theme pins a specific theme instead of the process-wide current theme.
func (t *TooltipWidget) Theme(th *theme.Theme) *TooltipWidget {
	t.theme = th
	t.content.theme = th
	t.content.label.Theme(th)
	return t
}

// Placement overrides the default top-center placement.
func (t *TooltipWidget) Placement(p popover.Placement) *TooltipWidget {
	t.content.placement = p
	return t
}

func (t *TooltipWidget) resolvedTheme() *theme.Theme {
	if t.theme != nil {
		return t.theme
	}
	return CurrentTheme()
}

// Content returns the bubble content widget (rendered directly by goldens).
func (t *TooltipWidget) Content() *TooltipContentWidget { return t.content }

// Layout sizes the wrapper to its target.
func (t *TooltipWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	var size geometry.Size
	if t.target != nil {
		size = t.target.Layout(ctx, c)
		setWidgetBounds(t.target, geometry.FromPointSize(t.Position(), size))
	} else {
		size = c.Constrain(geometry.Sz(0, 0))
	}
	t.SetBounds(geometry.FromPointSize(t.Position(), size))
	return size
}

// Draw renders the target. The bubble paints in its own overlay.
func (t *TooltipWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !t.IsVisible() {
		return
	}
	if t.target != nil {
		setWidgetBounds(t.target, t.Bounds())
		t.target.Draw(ctx, canvas)
	}
}

// Event tracks hover and drives the overlay. MouseEnter/Move inside the target
// schedules the bubble (after the anti-flicker delay); MouseLeave and any press
// hide it.
func (t *TooltipWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}

	switch me.MouseType {
	case event.MouseEnter:
		t.hovered = true
		t.hoverSince = ctx.Now()
		t.maybeShow(ctx)
	case event.MouseMove:
		inside := t.Bounds().Contains(me.Position)
		if inside {
			if !t.hovered {
				t.hovered = true
				t.hoverSince = ctx.Now()
			}
			t.maybeShow(ctx)
		} else if t.hovered {
			t.hovered = false
			t.hide(ctx)
		}
	case event.MouseLeave:
		t.hovered = false
		t.hide(ctx)
	case event.MousePress:
		t.hovered = false
		t.hide(ctx)
	}
	return false
}

// maybeShow opens the overlay once the hover grace period has elapsed.
func (t *TooltipWidget) maybeShow(ctx widget.Context) {
	if t.overlay != nil || !t.hovered {
		return
	}
	if ctx.Now().Sub(t.hoverSince) < metrics.TooltipDelayMillis*time.Millisecond {
		// Keep the frame ticking so the delay can elapse.
		ctx.Invalidate()
		return
	}
	om := ctx.OverlayManager()
	if om == nil {
		return
	}

	t.content.theme = t.resolvedTheme()
	t.content.label.Theme(t.resolvedTheme())
	size := t.content.Layout(ctx, geometry.Loose(ctx.WindowSize()))

	anchor := t.ScreenBounds()
	pos := popover.CalculatePosition(t.content.placement, anchor, size, ctx.WindowSize(), metrics.TooltipGap)
	t.content.anchorCenterX = anchor.Center().X
	t.content.SetBounds(geometry.FromPointSize(pos, size))

	om.PushOverlay(t.content, func() { t.overlay = nil })
	t.overlay = t.content
	ctx.Invalidate()
}

// hide removes the overlay if shown.
func (t *TooltipWidget) hide(ctx widget.Context) {
	if t.overlay == nil {
		return
	}
	if om := ctx.OverlayManager(); om != nil {
		om.RemoveOverlay(t.overlay)
	}
	t.overlay = nil
	ctx.Invalidate()
}

// Children returns the wrapped target.
func (t *TooltipWidget) Children() []widget.Widget {
	if t.target == nil {
		return nil
	}
	return []widget.Widget{t.target}
}

// TooltipContentWidget draws the inverted bubble (bg Foreground, text
// Background), the rotated-diamond arrow, and the text. It is the widget
// rendered directly at natural size by the goldens.
type TooltipContentWidget struct {
	widget.WidgetBase

	label     *TypographyWidget
	placement popover.Placement
	theme     *theme.Theme

	// anchorCenterX is the trigger center in screen space; the arrow is drawn
	// under that X (clamped to the bubble) so it points at the trigger even
	// when the bubble is shifted by viewport clamping. Zero means "centered".
	anchorCenterX float32
}

func newTooltipContent(text string) *TooltipContentWidget {
	c := &TooltipContentWidget{
		label: Text(text).
			FontSize(metrics.TooltipFontSize).
			Weight(metrics.TooltipFontWeight).
			LineHeight(metrics.TooltipLineHeight).
			Align(widget.TextAlignCenter).
			ColorToken(func(tok *theme.Tokens) widget.Color { return tok.Background }),
		placement: popover.Top,
	}
	c.SetVisible(true)
	c.SetEnabled(true)
	return c
}

func (c *TooltipContentWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// Layout measures the bubble: text advance + horizontal padding, line box +
// vertical padding.
func (c *TooltipContentWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	textSize := c.label.Layout(ctx, geometry.Loose(geometry.Sz(10000, 10000)))
	w := textSize.Width + 2*metrics.TooltipPadX
	h := metrics.TooltipLineHeight + 2*metrics.TooltipPadY
	size := cons.Constrain(geometry.Sz(w, h))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw paints the arrow diamond, the bubble fill, then the inverted text.
func (c *TooltipContentWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.resolvedTheme()
	tok := th.Active()
	b := c.Bounds()

	fill := tok.Foreground

	// Arrow first, so the bubble fill overpaints its inner half for a seamless
	// join. The arrow is a square rotated 45° (a diamond) drawn as a filled
	// polygon via the SVG renderer; on canvases without one (mock), it no-ops
	// and the spec test asserts the bubble geometry only.
	c.drawArrow(canvas, b, fill)

	canvas.DrawRoundRect(b, fill, th.RadiusMD())

	textBounds := geometry.NewRect(
		b.Min.X+metrics.TooltipPadX,
		b.Min.Y+metrics.TooltipPadY,
		b.Width()-2*metrics.TooltipPadX,
		metrics.TooltipLineHeight,
	)
	c.label.Theme(th)
	c.label.SetBounds(textBounds)
	c.label.Draw(ctx, canvas)
}

// drawArrow paints the rotated-diamond arrow straddling the bubble edge on the
// side facing the trigger. The diamond's center sits TooltipArrowInset px past
// the bubble edge so its near half overlaps the (later-drawn) bubble fill.
func (c *TooltipContentWidget) drawArrow(canvas widget.Canvas, b geometry.Rect, fill widget.Color) {
	renderer, ok := canvas.(widget.SVGRenderer)
	if !ok {
		return
	}

	half := metrics.TooltipArrowSize / 2
	cx := c.arrowCenterX(b)

	var center geometry.Point
	switch c.placement {
	case popover.Bottom, popover.BottomStart, popover.BottomEnd:
		// Bubble below trigger → arrow on top edge.
		center = geometry.Pt(cx, b.Min.Y-metrics.TooltipArrowInset)
	default:
		// Bubble above trigger (Top*) → arrow on bottom edge.
		center = geometry.Pt(cx, b.Max.Y+metrics.TooltipArrowInset)
	}

	box := geometry.NewRect(center.X-half, center.Y-half, metrics.TooltipArrowSize, metrics.TooltipArrowSize)
	renderer.RenderSVG(arrowDiamondSVG, box, fill)
}

// arrowCenterX returns the X at which to center the arrow: under the trigger
// center, clamped to keep the diamond fully within the bubble width.
func (c *TooltipContentWidget) arrowCenterX(b geometry.Rect) float32 {
	cx := b.Center().X
	if c.anchorCenterX != 0 {
		cx = c.anchorCenterX
	}
	half := metrics.TooltipArrowSize / 2
	if cx < b.Min.X+half {
		cx = b.Min.X + half
	}
	if cx > b.Max.X-half {
		cx = b.Max.X - half
	}
	return cx
}

// arrowDiamondSVG is a 10×10 viewBox square rotated 45° (a diamond), drawn as a
// filled path. fill is overridden at render time via RenderSVG's color arg.
var arrowDiamondSVG = []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">` +
	`<path d="M5 0 L10 5 L5 10 L0 5 Z" fill="currentColor"/></svg>`)

// Event swallows nothing; the bubble is inert (hover lives on the wrapper).
func (c *TooltipContentWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns the label.
func (c *TooltipContentWidget) Children() []widget.Widget {
	return []widget.Widget{c.label}
}

// Compile-time interface checks.
var (
	_ widget.Widget = (*TooltipWidget)(nil)
	_ widget.Widget = (*TooltipContentWidget)(nil)
)

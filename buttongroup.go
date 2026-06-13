package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// ButtonGroupWidget is shadcn's ButtonGroup: a horizontal row of buttons fused
// into one segmented control with shared borders and only the outer corners of
// the first and last segments rounded (DESIGN.md §5.4).
//
// WRAP DECISION (DESIGN.md §3.2): graft-OWNED composite over graft.Button. Each
// child button is forced to square corners and laid out edge-to-edge with a
// −1px overlap so adjacent 1px borders coincide into a single shared divider.
// The group rounds the four outer corners by clipping the whole row to a
// rounded rectangle (per-corner radii do not exist on the canvas), then strokes
// one unified outer border in the Border token so the rounded corners carry a
// border arc that the clipped square button borders would otherwise miss.
type ButtonGroupWidget struct {
	widget.WidgetBase

	buttons []*ButtonWidget
	theme   *theme.Theme
}

// ButtonGroup fuses the given buttons into a segmented control. The buttons are
// reconfigured to square corners; pass already-styled graft.Button values.
func ButtonGroup(buttons ...*ButtonWidget) *ButtonGroupWidget {
	g := &ButtonGroupWidget{buttons: buttons, theme: CurrentTheme()}
	g.SetVisible(true)
	g.SetEnabled(true)
	zero := float32(0)
	for _, b := range buttons {
		if b == nil {
			continue
		}
		b.SetParent(g)
		b.Style(func(s *Style) { s.Radius = &zero })
	}
	return g
}

// Theme pins a specific theme instead of the process-wide current theme.
func (g *ButtonGroupWidget) Theme(th *theme.Theme) *ButtonGroupWidget {
	if th != nil {
		g.theme = th
		for _, b := range g.buttons {
			b.Theme(th)
		}
	}
	return g
}

// radius returns the group's outer corner radius (rounded-md).
func (g *ButtonGroupWidget) radius() float32 { return g.theme.RadiusMD() }

// Layout places the buttons edge-to-edge with a −1px overlap and sizes the
// group to their combined width and shared height.
func (g *ButtonGroupWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	overlap := metrics.ButtonGroup.Overlap
	loose := c.Loosen()

	var x, height float32
	for i, b := range g.buttons {
		if b == nil {
			continue
		}
		sz := b.Layout(ctx, loose)
		if sz.Height > height {
			height = sz.Height
		}
		if i > 0 {
			x -= overlap
		}
		b.SetBounds(geometry.NewRect(x, 0, sz.Width, sz.Height))
		x += sz.Width
	}

	size := c.Constrain(geometry.Sz(x, height))
	// Re-pin every button's vertical bounds to the (uniform) group height.
	for _, b := range g.buttons {
		if b == nil {
			continue
		}
		bb := b.Bounds()
		b.SetBounds(geometry.NewRect(bb.Min.X, 0, bb.Width(), size.Height))
	}
	g.SetBounds(geometry.FromPointSize(g.Position(), size))
	return size
}

// Draw clips the group to a rounded rect, paints each button square inside the
// clip, then strokes the unified outer border in the Border token.
func (g *ButtonGroupWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !g.IsVisible() {
		return
	}
	tok := g.theme.Active()
	bounds := g.Bounds()
	radius := g.radius()

	canvas.PushClipRoundRect(bounds, radius)
	canvas.PushTransform(bounds.Min)
	for _, b := range g.buttons {
		if b == nil {
			continue
		}
		widget.StampScreenOrigin(b, canvas)
		widget.DrawChild(b, ctx, canvas)
	}
	canvas.PopTransform()
	canvas.PopClip()

	draw.InsideBorder(canvas, bounds, radius, tok.Border, metrics.ButtonGroup.OuterBorderWidth)
}

// Event forwards mouse input to the button under the pointer with group-local
// coordinates.
func (g *ButtonGroupWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	local := *me
	local.Position = me.Position.Sub(g.Bounds().Min)
	handled := false
	for _, b := range g.buttons {
		if b == nil {
			continue
		}
		// Translate per-button: button bounds are group-local.
		if b.Event(ctx, &local) {
			handled = true
		}
	}
	return handled
}

// Children returns the button segments.
func (g *ButtonGroupWidget) Children() []widget.Widget {
	out := make([]widget.Widget, 0, len(g.buttons))
	for _, b := range g.buttons {
		if b != nil {
			out = append(out, b)
		}
	}
	return out
}

// Compile-time interface check.
var _ widget.Widget = (*ButtonGroupWidget)(nil)

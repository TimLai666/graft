package graft

import (
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/theme"
)

// ScrollAreaWidget is graft's shadcn ScrollArea.
//
// Architecture decision: ScrollArea WRAPS gogpu/ui's core/scrollview —
// the core widget contributes substantial interaction machinery (wheel
// scrolling, thumb dragging, track paging, keyboard navigation, content
// clipping) — and restyles its scrollbar through painters.Scrollbar
// (transparent track + border-token pill thumb, metrics.ScrollArea).
// The core gutter geometry differs from shadcn by 1px of invisible
// transparent padding; the visible 8px thumb matches exactly.
//
// Note: shadcn's viewport draws a focus-visible ring on keyboard focus.
// core/scrollview also takes focus on thumb press, which would render
// the ring for plain mouse interactions, so graft omits the viewport
// ring entirely (see DESIGN.md section 5.2 focus-visible semantics).
type ScrollAreaWidget struct {
	widget.WidgetBase

	child widget.Widget
	sv    *scrollview.Widget

	w, h  float32
	theme *theme.Theme
}

// ScrollArea wraps child in a scrollable viewport with a shadcn-styled
// scrollbar. Size it with W/H; an unsized axis fills the available space.
func ScrollArea(child Widget) *ScrollAreaWidget {
	s := &ScrollAreaWidget{child: child}
	s.SetVisible(true)
	s.SetEnabled(true)
	s.rebuild()
	return s
}

// rebuild constructs the wrapped scrollview with the painter for the
// resolved theme (painters cannot be swapped after construction).
func (s *ScrollAreaWidget) rebuild() {
	p := PaintersFor(s.theme)
	s.sv = scrollview.New(s.child, scrollview.PainterOpt(p.Scrollbar))
	s.sv.SetParent(s)
}

// W sets the viewport width in px.
func (s *ScrollAreaWidget) W(px float32) *ScrollAreaWidget {
	s.w = px
	return s
}

// H sets the viewport height in px.
func (s *ScrollAreaWidget) H(px float32) *ScrollAreaWidget {
	s.h = px
	return s
}

// Theme pins a specific theme instead of the process-wide current theme.
func (s *ScrollAreaWidget) Theme(th *theme.Theme) *ScrollAreaWidget {
	s.theme = th
	s.rebuild()
	return s
}

// ScrollView exposes the wrapped core/scrollview widget (scroll offsets,
// content size) for advanced use and tests.
func (s *ScrollAreaWidget) ScrollView() *scrollview.Widget { return s.sv }

// fallbackViewport mirrors core/scrollview's default viewport dimension,
// used when an axis is neither sized via W/H nor bounded by constraints.
const fallbackViewport float32 = 200

// Layout sizes the viewport (explicit W/H first, then available space)
// and lays out the wrapped scrollview tightly inside it.
func (s *ScrollAreaWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	w, h := s.w, s.h
	if w <= 0 {
		if c.HasBoundedWidth() {
			w = c.MaxWidth
		} else {
			w = fallbackViewport
		}
	}
	if h <= 0 {
		if c.HasBoundedHeight() {
			h = c.MaxHeight
		} else {
			h = fallbackViewport
		}
	}
	size := c.Constrain(geometry.Sz(w, h))

	s.sv.Layout(ctx, geometry.Tight(size))
	s.sv.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))

	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

// Draw renders the wrapped scrollview (content + scrollbar) at this
// widget's origin.
func (s *ScrollAreaWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !s.IsVisible() {
		return
	}
	bounds := s.Bounds()
	canvas.PushTransform(bounds.Min)
	widget.StampScreenOrigin(s.sv, canvas)
	widget.DrawChild(s.sv, ctx, canvas)
	canvas.PopTransform()
}

// Event forwards input to the wrapped scrollview, translating positional
// events into its coordinate space.
func (s *ScrollAreaWidget) Event(ctx widget.Context, e event.Event) bool {
	offset := s.Bounds().Min
	switch ev := e.(type) {
	case *event.MouseEvent:
		local := *ev
		local.Position = ev.Position.Sub(offset)
		return s.sv.Event(ctx, &local)
	case *event.WheelEvent:
		local := *ev
		local.Position = ev.Position.Sub(offset)
		return s.sv.Event(ctx, &local)
	default:
		return s.sv.Event(ctx, e)
	}
}

// Children returns the wrapped scrollview.
func (s *ScrollAreaWidget) Children() []widget.Widget {
	return []widget.Widget{s.sv}
}

var _ widget.Widget = (*ScrollAreaWidget)(nil)

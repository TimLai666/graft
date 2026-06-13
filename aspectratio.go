package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
)

// AspectRatioWidget is shadcn's AspectRatio: a container that sizes its child
// to a width:height ratio within the available width.
//
// shadcn's aspect-ratio.tsx is a thin wrapper over Radix AspectRatio.Root and
// carries no styling of its own; the box has no padding, border, or background.
// AspectRatioWidget mirrors that: it measures the available width, derives the
// height from the ratio, and lays the child out at that exact box (the child is
// commonly an Image with object-cover, or any widget that should fill the box).
//
// Architecture: graft-owned widget. Pure layout, no painting; Draw simply
// stamps the child. No colors are resolved, so there is nothing theme-sensitive
// here.
type AspectRatioWidget struct {
	widget.WidgetBase

	child widget.Widget
	ratio float32 // width / height
}

// AspectRatio sizes child to a width:height ratio within the available width.
// A ratio of 16.0/9.0 produces a 16:9 box. Non-positive ratios fall back to
// the shadcn demo default (16/9).
func AspectRatio(ratio float32, child widget.Widget) *AspectRatioWidget {
	if ratio <= 0 {
		ratio = metrics.AspectRatioDefault
	}
	a := &AspectRatioWidget{child: child, ratio: ratio}
	a.SetVisible(true)
	a.SetEnabled(true)
	ovlSetParent(child, a)
	return a
}

// Ratio overrides the width:height ratio.
func (a *AspectRatioWidget) Ratio(ratio float32) *AspectRatioWidget {
	if ratio > 0 {
		a.ratio = ratio
	}
	return a
}

// Layout pins the box to the available width and a height of width/ratio, then
// fills the child to that box. When width is unbounded it falls back to a
// 0-height box (callers should constrain the width, as shadcn's CSS does).
func (a *AspectRatioWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	w := cons.MaxWidth
	if w <= 0 || w >= geometry.Infinity {
		w = 0
	}
	h := w / a.ratio
	size := cons.Constrain(geometry.Sz(w, h))

	if a.child != nil {
		a.child.Layout(ctx, geometry.Tight(size))
		ovlSetBounds(a.child, geometry.FromPointSize(geometry.Pt(0, 0), size))
	}
	a.SetBounds(geometry.FromPointSize(a.Position(), size))
	return size
}

// Draw stamps the child at the box origin.
func (a *AspectRatioWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !a.IsVisible() || a.child == nil {
		return
	}
	bounds := a.Bounds()
	canvas.PushTransform(bounds.Min)
	widget.StampScreenOrigin(a.child, canvas)
	widget.DrawChild(a.child, ctx, canvas)
	canvas.PopTransform()
}

// Event forwards input to the child in box-local coordinates. The child fills
// the whole box, so any positional event reaching the host is inside the child;
// only the coordinate translation is needed.
func (a *AspectRatioWidget) Event(ctx widget.Context, e event.Event) bool {
	if a.child == nil {
		return false
	}
	return a.child.Event(ctx, ovlTranslate(e, a.Bounds().Min))
}

// Children returns the single child.
func (a *AspectRatioWidget) Children() []widget.Widget {
	if a.child == nil {
		return nil
	}
	return []widget.Widget{a.child}
}

var _ widget.Widget = (*AspectRatioWidget)(nil)

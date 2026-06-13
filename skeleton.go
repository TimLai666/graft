package graft

import (
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// SkeletonWidget is graft's loading placeholder: an Accent-filled rounded-md
// box that pulses its opacity between 1 and 0.5 on a 2s loop, matching the
// shadcn Skeleton ("animate-pulse rounded-md bg-accent",
// docs/research/03-shadcn-pixel-spec.md §5).
//
// Size it explicitly with W/H; there is no intrinsic content. For a circular
// avatar-style skeleton, set a square size and Rounded(th.RadiusFull()) (or
// use the .Circle helper).
//
// The pulse is widget-driven (Report 2 §8): elapsed time accumulates from
// ctx.DeltaTime() in Layout and the widget asks the framework to keep ticking
// while it animates. A single-frame render (the offscreen golden path) sees
// DeltaTime() == 0 on the first frame, so the golden is deterministic at the
// pulse start (full opacity).
type SkeletonWidget struct {
	widget.WidgetBase

	w, h      float32
	radius    float32 // 0 means rounded-md (RadiusMD); use Rounded to override
	hasRadius bool

	elapsed time.Duration // accumulated animation time within the pulse loop

	theme *theme.Theme
}

// Skeleton creates a loading placeholder. Set its size with W and H (default
// 0×0 — always size it). The radius defaults to rounded-md.
func Skeleton() *SkeletonWidget {
	s := &SkeletonWidget{theme: CurrentTheme()}
	s.SetVisible(true)
	s.SetEnabled(true)
	return s
}

// W sets the skeleton width in px.
func (s *SkeletonWidget) W(px float32) *SkeletonWidget {
	s.w = px
	return s
}

// H sets the skeleton height in px.
func (s *SkeletonWidget) H(px float32) *SkeletonWidget {
	s.h = px
	return s
}

// Size sets both width and height in px.
func (s *SkeletonWidget) Size(w, h float32) *SkeletonWidget {
	s.w, s.h = w, h
	return s
}

// Rounded overrides the corner radius in px (default rounded-md). Pass a
// theme radius (e.g. th.RadiusFull()) for a circular skeleton.
func (s *SkeletonWidget) Rounded(px float32) *SkeletonWidget {
	s.radius = px
	s.hasRadius = true
	return s
}

// Circle makes the skeleton a full-radius pill/circle (rounded-full), the
// shape used for avatar placeholders.
func (s *SkeletonWidget) Circle() *SkeletonWidget {
	return s.Rounded(s.resolvedTheme().RadiusFull())
}

// Theme pins a specific theme instead of the process-wide current theme.
func (s *SkeletonWidget) Theme(th *theme.Theme) *SkeletonWidget {
	s.theme = th
	return s
}

func (s *SkeletonWidget) resolvedTheme() *theme.Theme {
	if s.theme != nil {
		return s.theme
	}
	return CurrentTheme()
}

func (s *SkeletonWidget) cornerRadius(th *theme.Theme) float32 {
	if s.hasRadius {
		return s.radius
	}
	return th.RadiusMD()
}

// Layout sizes the skeleton to its explicit dimensions and advances the pulse
// clock by the frame delta, keeping the animation pumping while mounted.
func (s *SkeletonWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if ctx != nil {
		dt := ctx.DeltaTime()
		if dt > 0 {
			s.elapsed = (s.elapsed + dt) % metrics.Skeleton.PulsePeriod
			s.SetNeedsRedraw(true)
		}
		// Keep the render loop ticking so the pulse continues.
		if sched, ok := ctx.(widget.AnimationScheduler); ok {
			sched.ScheduleAnimationFrame()
		}
	}
	size := c.Constrain(geometry.Sz(s.w, s.h))
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

// pulseOpacity returns the current pulse opacity (1 at the loop start/end,
// metrics.Skeleton.PulseMinOpacity at the midpoint), easing the linear loop
// phase through the Tailwind cubic-bezier(0.4, 0, 0.6, 1) curve.
func (s *SkeletonWidget) pulseOpacity() float32 {
	period := float32(metrics.Skeleton.PulsePeriod)
	t := float32(s.elapsed) / period // linear phase in [0, 1)

	// Triangle wave: 0→1 over the first half (toward the trough), 1→0 back.
	var tri float32
	if t < 0.5 {
		tri = t * 2
	} else {
		tri = (1 - t) * 2
	}
	eased := cubicBezierY(tri, 0.4, 0, 0.6, 1)
	return 1 - eased*(1-metrics.Skeleton.PulseMinOpacity)
}

// Draw paints the Accent fill at the current pulse opacity, rounded-md.
func (s *SkeletonWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !s.IsVisible() {
		return
	}
	th := s.resolvedTheme()
	tok := th.Active()
	fill := draw.MulAlpha(tok.Accent, s.pulseOpacity())
	canvas.DrawRoundRect(s.Bounds(), fill, s.cornerRadius(th))
}

// Event ignores all input; the skeleton is inert.
func (s *SkeletonWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; the skeleton is a leaf.
func (s *SkeletonWidget) Children() []widget.Widget { return nil }

// cubicBezierY evaluates the y value of a CSS cubic-bezier(x1, y1, x2, y2)
// timing curve at the given progress x in [0, 1]. The endpoints are fixed at
// (0,0) and (1,1). It solves for the bezier parameter t such that Bx(t) == x
// (Newton's method with a bisection fallback), then returns By(t).
func cubicBezierY(x, x1, y1, x2, y2 float32) float32 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}
	bezier := func(c1, c2, t float32) float32 {
		mt := 1 - t
		return 3*mt*mt*t*c1 + 3*mt*t*t*c2 + t*t*t
	}
	// Solve Bx(t) = x for t.
	t := x
	for i := 0; i < 8; i++ {
		xEst := bezier(x1, x2, t)
		// Derivative of Bx wrt t.
		mt := 1 - t
		d := 3*mt*mt*x1 + 6*mt*t*(x2-x1) + 3*t*t*(1-x2)
		if d == 0 {
			break
		}
		t -= (xEst - x) / d
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
	}
	return bezier(y1, y2, t)
}

var (
	_ widget.Widget = (*SkeletonWidget)(nil)
)

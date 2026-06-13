package graft

import (
	"math"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// SpinnerWidget is graft's loading spinner: a ~270° arc with round caps that
// rotates once per second, matching the shadcn Spinner (a spinning lucide
// Loader2, "size-4 animate-spin", docs/research/03-shadcn-pixel-spec.md §5).
//
// The gogpu/ui canvas has no rotation transform, so rather than rotate an
// icon the spinner draws its arc directly and advances the start angle from
// the accumulated frame time. A single-frame render (the offscreen golden
// path) sees DeltaTime() == 0, so the golden is deterministic at rotation
// phase 0 (leading cap at the top).
//
// Default size is 16px; override with Size. The color is currentColor →
// Foreground by default; override the token with ColorToken.
type SpinnerWidget struct {
	widget.WidgetBase

	size    float32
	elapsed time.Duration // accumulated rotation time within the spin loop

	colorFn func(*theme.Tokens) widget.Color // nil → Foreground
	theme   *theme.Theme
}

// Spinner creates a 16px loading spinner in the foreground color.
func Spinner() *SpinnerWidget {
	s := &SpinnerWidget{size: metrics.Spinner.DefaultSize, theme: CurrentTheme()}
	s.SetVisible(true)
	s.SetEnabled(true)
	return s
}

// Size sets the spinner side length in px (default 16).
func (s *SpinnerWidget) Size(px float32) *SpinnerWidget {
	s.size = px
	return s
}

// ColorToken selects the arc color from the active token set at draw time
// (e.g. func(t *theme.Tokens) widget.Color { return t.MutedForeground }).
func (s *SpinnerWidget) ColorToken(fn func(*theme.Tokens) widget.Color) *SpinnerWidget {
	s.colorFn = fn
	return s
}

// Theme pins a specific theme instead of the process-wide current theme.
func (s *SpinnerWidget) Theme(th *theme.Theme) *SpinnerWidget {
	s.theme = th
	return s
}

func (s *SpinnerWidget) resolvedTheme() *theme.Theme {
	if s.theme != nil {
		return s.theme
	}
	return CurrentTheme()
}

// Layout sizes the spinner to a square and advances the rotation clock by the
// frame delta, keeping the animation pumping while mounted.
func (s *SpinnerWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if ctx != nil {
		dt := ctx.DeltaTime()
		if dt > 0 {
			s.elapsed = (s.elapsed + dt) % metrics.Spinner.SpinPeriod
			s.SetNeedsRedraw(true)
		}
		if sched, ok := ctx.(widget.AnimationScheduler); ok {
			sched.ScheduleAnimationFrame()
		}
	}
	size := c.Constrain(geometry.Sz(s.size, s.size))
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

// rotation returns the current rotation offset in radians (0 at phase start).
func (s *SpinnerWidget) rotation() float64 {
	frac := float64(s.elapsed) / float64(metrics.Spinner.SpinPeriod)
	return frac * 2 * math.Pi
}

// Draw paints the rotating arc with round caps in the resolved color.
func (s *SpinnerWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !s.IsVisible() {
		return
	}
	th := s.resolvedTheme()
	tok := th.Active()

	col := tok.Foreground
	if s.colorFn != nil {
		col = s.colorFn(tok)
	}

	bounds := s.Bounds()
	stroke := bounds.Width() * metrics.Spinner.StrokeRatio
	center := bounds.Center()
	// Inset the arc radius by half the stroke so the stroke stays inside the
	// nominal size box (matching the lucide icon's bounding box).
	radius := bounds.Width()/2 - stroke/2
	if radius < 0 {
		radius = 0
	}

	start := metrics.Spinner.StartAngle + s.rotation()
	strokeArc(canvas, center, radius, start, metrics.Spinner.SweepAngle, col, stroke)
}

// strokeArc draws an arc with round caps via the ArcStroker capability,
// falling back to a plain (butt-cap) StrokeArc on canvases without it (the
// mock canvas records the StrokeArc form).
func strokeArc(canvas widget.Canvas, center geometry.Point, radius float32, start, sweep float64, col widget.Color, stroke float32) {
	if as, ok := canvas.(widget.ArcStroker); ok {
		as.StrokeArcStyled(center, radius, start, sweep, col, stroke, widget.LineCapRound)
		return
	}
	canvas.StrokeArc(center, radius, start, sweep, col, stroke)
}

// Event ignores all input; the spinner is inert.
func (s *SpinnerWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; the spinner is a leaf.
func (s *SpinnerWidget) Children() []widget.Widget { return nil }

var (
	_ widget.Widget = (*SpinnerWidget)(nil)
)

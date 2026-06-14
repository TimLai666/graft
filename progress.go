package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// ProgressWidget is graft's determinate progress bar: a full-width 4px pill
// whose track is the Primary token at 20% and whose indicator fills the pill
// left→right by the current value (0..100), matching the shadcn Progress
// (docs/research/03-shadcn-pixel-spec.md §5).
//
// The indicator is clipped to the rounded-full pill so its leading edge keeps
// the pill's rounded ends (mirroring shadcn's overflow-hidden rounded-full
// root). Drive the value statically with Value, or reactively with Bind.
type ProgressWidget struct {
	widget.WidgetBase

	value float64                       // static value when sig is nil
	sig   state.ReadonlySignal[float64] // bound value source (controlled)

	theme *theme.Theme
}

// Progress creates a progress bar at 0%.
func Progress() *ProgressWidget {
	p := &ProgressWidget{theme: CurrentTheme()}
	p.SetVisible(true)
	p.SetEnabled(true)
	return p
}

// Value sets the initial/static progress value, clamped to 0..100.
func (p *ProgressWidget) Value(v float64) *ProgressWidget {
	p.value = clampProgress(v)
	return p
}

// Bind drives the value from a signal (controlled mode). The widget repaints
// whenever the signal changes; the binding is created in Mount and torn down
// automatically on unmount.
func (p *ProgressWidget) Bind(sig state.Signal[float64]) *ProgressWidget {
	p.sig = sig
	return p
}

// Theme pins a specific theme instead of the process-wide current theme.
func (p *ProgressWidget) Theme(th *theme.Theme) *ProgressWidget {
	p.theme = th
	return p
}

func (p *ProgressWidget) resolvedTheme() *theme.Theme {
	if p.theme != nil {
		return p.theme
	}
	return CurrentTheme()
}

// currentValue returns the live value: the bound signal if set, else the
// static value.
func (p *ProgressWidget) currentValue() float64 {
	if p.sig != nil {
		return clampProgress(p.sig.Get())
	}
	return p.value
}

// Mount binds the value signal to the scheduler so signal writes repaint just
// this widget (Report 2 §6).
func (p *ProgressWidget) Mount(ctx widget.Context) {
	if p.sig == nil {
		return
	}
	sched := ctx.Scheduler()
	if sched == nil {
		return // headless
	}
	p.AddBinding(state.BindToScheduler[float64](p.sig, p, sched))
}

// Unmount is a no-op; CleanupBindings runs automatically before it.
func (p *ProgressWidget) Unmount() {}

// Layout makes the bar full-width and 4px tall.
func (p *ProgressWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	w := c.MaxWidth
	if w <= 0 || w > 100000 {
		w = 0
	}
	size := c.Constrain(geometry.Sz(w, metrics.Progress.Height))
	p.SetBounds(geometry.FromPointSize(p.Position(), size))
	return size
}

// Draw paints the Primary@20% track, then the Primary indicator clipped to the
// pill so the fill keeps rounded ends. Both colors resolve from the active
// token set at draw time.
func (p *ProgressWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !p.IsVisible() {
		return
	}
	th := p.resolvedTheme()
	tok := th.Active()
	bounds := p.Bounds()
	radius := th.RadiusFull()

	// Track: Primary at 20%.
	canvas.DrawRoundRect(bounds, draw.Alpha(tok.Primary, metrics.Progress.TrackAlpha), radius)

	// Indicator: Primary, width = value fraction, clipped to the pill so the
	// leading edge is rounded along with the trailing one.
	frac := float32((p.currentValue() - metrics.Progress.Min) / (metrics.Progress.Max - metrics.Progress.Min))
	if frac <= 0 {
		return
	}
	if frac > 1 {
		frac = 1
	}
	canvas.PushClipRoundRect(bounds, radius)
	fillW := bounds.Width() * frac
	fill := geometry.NewRect(bounds.Min.X, bounds.Min.Y, fillW, bounds.Height())
	canvas.DrawRoundRect(fill, tok.Primary, radius)
	canvas.PopClip()
}

// Event ignores all input; the progress bar is a display element.
func (p *ProgressWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; the progress bar is a leaf.
func (p *ProgressWidget) Children() []widget.Widget { return nil }

// clampProgress clamps v to the 0..100 value range.
func clampProgress(v float64) float64 {
	if v < metrics.Progress.Min {
		return metrics.Progress.Min
	}
	if v > metrics.Progress.Max {
		return metrics.Progress.Max
	}
	return v
}

var (
	_ widget.Widget    = (*ProgressWidget)(nil)
	_ widget.Lifecycle = (*ProgressWidget)(nil)
)

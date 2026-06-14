package graft

import (
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// SliderWidget is the shadcn Slider.
//
// Architecture decision (DESIGN.md sections 3.1/7): Slider WRAPS the
// gogpu/ui core/slider widget — it contributes the drag, pointer-capture,
// and keyboard machinery — and styles it with painters.Slider. The core
// Layout is fully constrainable via tight constraints, so shadcn's exact
// pixels (12px-high root, full-width track) are reachable without a
// graft-owned reimplementation. The wrapper owns sizing (.W / w-full) and
// theme injection; all visuals live in painters/slider.go + metrics/slider.go.
//
// Like shadcn, the root is w-full: inside a width-bounded container the
// slider stretches; use W for an explicit width.
type SliderWidget struct {
	widget.WidgetBase

	th *theme.Theme

	min, max, step float32
	value          float32
	hasValue       bool
	sig            state.Signal[float32]
	onChange       func(float32)
	disabled       bool
	width          float32

	core *slider.Widget
}

// Slider creates a slider with range [0, 100], like shadcn's defaults.
func Slider() *SliderWidget {
	s := &SliderWidget{th: CurrentTheme(), max: 100}
	s.SetVisible(true)
	s.SetEnabled(true)
	return s
}

// Min sets the minimum value (default 0).
func (s *SliderWidget) Min(v float32) *SliderWidget { s.min = v; return s.invalidateCore() }

// Max sets the maximum value (default 100).
func (s *SliderWidget) Max(v float32) *SliderWidget { s.max = v; return s.invalidateCore() }

// Step sets the snap increment; 0 (default) is continuous.
func (s *SliderWidget) Step(v float32) *SliderWidget { s.step = v; return s.invalidateCore() }

// Value sets the initial (uncontrolled) value.
func (s *SliderWidget) Value(v float32) *SliderWidget {
	s.value = v
	s.hasValue = true
	return s.invalidateCore()
}

// Bind makes the slider controlled by a signal: the thumb reflects the
// signal and dragging writes back to it.
func (s *SliderWidget) Bind(sig state.Signal[float32]) *SliderWidget {
	s.sig = sig
	return s.invalidateCore()
}

// OnChange registers an observer for value changes.
func (s *SliderWidget) OnChange(fn func(float32)) *SliderWidget {
	s.onChange = fn
	return s.invalidateCore()
}

// Disabled sets the disabled state (50% opacity, no interaction).
func (s *SliderWidget) Disabled(v bool) *SliderWidget { s.disabled = v; return s.invalidateCore() }

// W pins an explicit width in px instead of stretching to the container.
func (s *SliderWidget) W(px float32) *SliderWidget { s.width = px; return s }

// Theme pins a specific theme instead of the snapshotted current theme.
func (s *SliderWidget) Theme(th *theme.Theme) *SliderWidget { s.th = th; return s.invalidateCore() }

// invalidateCore drops the built core widget so the next layout rebuilds
// it with the updated configuration. Config methods run before the tree
// is mounted, so this never discards live state.
func (s *SliderWidget) invalidateCore() *SliderWidget {
	s.core = nil
	return s
}

// ensureCore builds the wrapped core/slider on first use.
func (s *SliderWidget) ensureCore() *slider.Widget {
	if s.core != nil {
		return s.core
	}
	opts := []slider.Option{
		slider.Min(s.min),
		slider.Max(s.max),
		slider.PainterOpt(PaintersFor(s.th).Slider),
	}
	if s.step > 0 {
		opts = append(opts, slider.Step(s.step))
	}
	switch {
	case s.sig != nil:
		opts = append(opts, slider.ValueSignal(s.sig))
	case s.hasValue:
		opts = append(opts, slider.Value(s.value))
	}
	if s.onChange != nil {
		opts = append(opts, slider.OnChange(s.onChange))
	}
	if s.disabled {
		opts = append(opts, slider.Disabled(true))
	}
	c := slider.New(opts...)
	c.SetParent(s)
	s.core = c
	return c
}

// Layout sizes the slider: explicit .W, else the bounded container width
// (w-full), else metrics.Slider.DefaultWidth; height is the 12px thumb.
func (s *SliderWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	core := s.ensureCore()
	w := s.width
	if w <= 0 {
		if c.MaxWidth > 0 && c.MaxWidth < geometry.Infinity {
			w = c.MaxWidth
		} else {
			w = metrics.Slider.DefaultWidth
		}
	}
	size := c.Constrain(geometry.Sz(w, metrics.Slider.Height))
	core.Layout(ctx, geometry.Tight(size))
	core.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

// Draw delegates to the wrapped core slider (which paints through
// painters.Slider).
func (s *SliderWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !s.IsVisible() {
		return
	}
	core := s.ensureCore()
	canvas.PushTransform(s.Bounds().Min)
	widget.StampScreenOrigin(core, canvas)
	widget.DrawChild(core, ctx, canvas)
	canvas.PopTransform()
}

// Event forwards input to the core slider, translating pointer positions
// into its coordinate space.
func (s *SliderWidget) Event(ctx widget.Context, e event.Event) bool {
	if !s.IsVisible() || !s.IsEnabled() {
		return false
	}
	core := s.ensureCore()
	switch ev := e.(type) {
	case *event.MouseEvent:
		local := *ev
		local.Position = ev.Position.Sub(s.Bounds().Min)
		return core.Event(ctx, &local)
	case *event.WheelEvent:
		local := *ev
		local.Position = ev.Position.Sub(s.Bounds().Min)
		return core.Event(ctx, &local)
	default:
		return core.Event(ctx, e)
	}
}

// Children exposes the core slider so mounting, focus traversal, and
// hover hit-testing reach it.
func (s *SliderWidget) Children() []widget.Widget {
	return []widget.Widget{s.ensureCore()}
}

var _ widget.Widget = (*SliderWidget)(nil)

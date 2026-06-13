package graft

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// SwitchWidget is the shadcn Switch: a pill track with a sliding thumb
// (docs/research/03-shadcn-pixel-spec.md §5 "Switch").
//
// OWNED widget (DESIGN.md §3.2 / §5 decision 5): Switch is a new widget, not a
// checkbox painter — a painter cannot change layout, and the 32×18.4 track +
// 16px thumb with 14px travel need their own Layout. Drawn directly via
// internal/draw + metrics + theme tokens; colors resolve from the active token
// set inside Draw so mode switches repaint without rebuilding.
//
// The thumb translation animates 150ms with cubic-bezier(0.4,0,0.2,1), driven
// by an animation.Controller ticked in Layout (the collapsible §8.2 pattern).
// Goldens render the settled state (progress pinned to the checked value).
type SwitchWidget struct {
	widget.WidgetBase

	checked  bool
	hovered  bool
	disabled bool
	sm       bool

	sig      state.Signal[bool]
	onChange func(bool)

	theme *theme.Theme

	// progress is the thumb position: 0 = off (left), 1 = on (right). It is
	// pinned to the checked state at construction and tweened on toggle.
	progress float32
	animCtrl *animation.Controller
}

// Switch creates an unchecked switch snapshotting the current theme.
func Switch() *SwitchWidget {
	s := &SwitchWidget{theme: CurrentTheme(), animCtrl: animation.NewController()}
	s.SetVisible(true)
	s.SetEnabled(true)
	return s
}

// Sm selects the small size (24×14 track / 12px thumb / 10px travel).
func (s *SwitchWidget) Sm() *SwitchWidget {
	s.sm = true
	return s
}

// Checked sets the initial (uncontrolled) checked state.
func (s *SwitchWidget) Checked(v bool) *SwitchWidget {
	s.checked = v
	s.progress = boolToProgress(v)
	return s
}

// Bind makes the switch controlled by sig. The binding is registered in Mount.
func (s *SwitchWidget) Bind(sig state.Signal[bool]) *SwitchWidget {
	s.sig = sig
	s.checked = sig.Get()
	s.progress = boolToProgress(s.checked)
	return s
}

// OnChange registers a callback fired on every toggle.
func (s *SwitchWidget) OnChange(fn func(bool)) *SwitchWidget {
	s.onChange = fn
	return s
}

// Disabled sets the disabled state (faded, not focusable, ignores input).
func (s *SwitchWidget) Disabled(v bool) *SwitchWidget {
	s.disabled = v
	return s
}

// Theme pins a specific theme instead of the process-wide current theme.
func (s *SwitchWidget) Theme(th *theme.Theme) *SwitchWidget {
	s.theme = th
	return s
}

// IsChecked reports the current checked state (signal wins when bound).
func (s *SwitchWidget) IsChecked() bool {
	if s.sig != nil {
		return s.sig.Get()
	}
	return s.checked
}

func (s *SwitchWidget) size() metrics.SwitchSize {
	if s.sm {
		return metrics.Switch.SM
	}
	return metrics.Switch.Default
}

func (s *SwitchWidget) resolvedTheme() *theme.Theme {
	if s.theme != nil {
		return s.theme
	}
	return CurrentTheme()
}

// IsFocusable reports whether the switch can receive keyboard focus.
func (s *SwitchWidget) IsFocusable() bool {
	return s.IsVisible() && s.IsEnabled() && !s.disabled
}

// Layout sizes the track and ticks the thumb animation.
func (s *SwitchWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	s.tickAnimation(ctx)
	sz := s.size()
	size := c.Constrain(geometry.Sz(sz.TrackWidth, sz.TrackHeight))
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

// Draw renders the track (shadow, fill, transparent border) and the thumb.
func (s *SwitchWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !s.IsVisible() {
		return
	}
	th := s.resolvedTheme()
	tok := th.Active()
	dark := th.IsDark()
	bounds := s.Bounds()
	disabled := s.disabled
	sz := s.size()
	checked := s.IsChecked()

	// Track geometry centered in bounds.
	track := geometry.NewRect(
		bounds.Min.X+(bounds.Width()-sz.TrackWidth)/2,
		bounds.Min.Y+(bounds.Height()-sz.TrackHeight)/2,
		sz.TrackWidth, sz.TrackHeight,
	)
	radius := sz.TrackHeight / 2 // rounded-full pill

	// Shadow-xs under the track.
	if !disabled {
		draw.Shadow(canvas, track, radius, metrics.ShadowXS)
	}

	// Track fill: on → Primary; off → Input (dark off → Input/80).
	trackColor := tok.Input
	if checked {
		trackColor = tok.Primary
	} else if dark {
		trackColor = draw.MulAlpha(tok.Input, metrics.Switch.DarkUncheckedTrackAlpha)
	}
	canvas.DrawRoundRect(track, draw.Fade(trackColor, disabled), radius)

	// Focus ring on the track (radius full).
	if s.IsFocused() {
		draw.FocusRing(canvas, track, radius, draw.Alpha(tok.Ring, metrics.RingAlpha))
	}

	// Thumb: animated horizontal position from progress.
	inset := (sz.TrackHeight - sz.ThumbSize) / 2
	thumbX := track.Min.X + inset + sz.Travel*s.progress
	thumbCenter := geometry.Pt(thumbX+sz.ThumbSize/2, track.Center().Y)
	canvas.DrawCircle(thumbCenter, sz.ThumbSize/2, draw.Fade(s.thumbColor(tok, dark, checked), disabled))
}

// thumbColor resolves the thumb fill per mode/state: light always Background;
// dark on → PrimaryForeground, dark off → Foreground.
func (s *SwitchWidget) thumbColor(tok *theme.Tokens, dark, checked bool) widget.Color {
	if !dark {
		return tok.Background
	}
	if checked {
		return tok.PrimaryForeground
	}
	return tok.Foreground
}

// toggle flips the checked state, animates the thumb, writes the bound signal,
// and fires OnChange.
func (s *SwitchWidget) toggle(ctx widget.Context) {
	next := !s.IsChecked()
	s.checked = next
	if s.sig != nil {
		s.sig.Set(next)
	}
	if s.onChange != nil {
		s.onChange(next)
	}
	s.startAnimation(ctx, next)
	s.SetNeedsRedraw(true)
	ctx.InvalidateRect(s.Bounds())
}

// startAnimation tweens progress toward the target (1 on, 0 off) over 150ms
// with the shadcn cubic-bezier easing.
func (s *SwitchWidget) startAnimation(ctx widget.Context, on bool) {
	if s.animCtrl == nil {
		s.progress = boolToProgress(on)
		return
	}
	adapter := &switchProgress{w: s, ctx: ctx}
	animation.To(adapter, boolToProgress(on)).
		From(s.progress).
		Duration(metrics.Switch.AnimDuration).
		Ease(animation.CubicBezier(0.4, 0, 0.2, 1)).
		Start(s.animCtrl)
}

// tickAnimation advances the thumb animation while active.
func (s *SwitchWidget) tickAnimation(ctx widget.Context) {
	if s.animCtrl == nil || !s.animCtrl.HasActive() {
		return
	}
	dt := ctx.DeltaTime()
	if dt < time.Millisecond {
		dt = time.Millisecond
	}
	if dt > 32*time.Millisecond {
		dt = 32 * time.Millisecond
	}
	s.animCtrl.Tick(dt)
	if s.animCtrl.HasActive() {
		s.SetNeedsRedraw(true)
		ctx.Invalidate()
	}
}

// Event handles hover, click, and Space-to-toggle.
func (s *SwitchWidget) Event(ctx widget.Context, e event.Event) bool {
	if s.disabled {
		return false
	}
	switch ev := e.(type) {
	case *event.MouseEvent:
		return s.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		if s.IsFocused() && ev.KeyType == event.KeyPress && ev.Key == event.KeySpace {
			s.toggle(ctx)
			return true
		}
	}
	return false
}

func (s *SwitchWidget) mouseEvent(ctx widget.Context, ev *event.MouseEvent) bool {
	switch ev.MouseType {
	case event.MouseEnter:
		s.hovered = true
		ctx.SetCursor(widget.CursorPointer)
		s.SetNeedsRedraw(true)
		ctx.InvalidateRect(s.Bounds())
		return true
	case event.MouseLeave:
		s.hovered = false
		ctx.SetCursor(widget.CursorDefault)
		s.SetNeedsRedraw(true)
		ctx.InvalidateRect(s.Bounds())
		return true
	case event.MousePress:
		if ev.Button != event.ButtonLeft {
			return false
		}
		ctx.RequestFocus(s)
		return true
	case event.MouseRelease:
		if s.Bounds().Contains(ev.Position) {
			s.toggle(ctx)
		}
		return true
	}
	return false
}

// Children returns nil; the switch is a leaf.
func (s *SwitchWidget) Children() []widget.Widget { return nil }

// Mount binds the controlled signal and syncs the settled thumb position.
func (s *SwitchWidget) Mount(ctx widget.Context) {
	if sched := ctx.Scheduler(); sched != nil && s.sig != nil {
		s.AddBinding(state.BindToScheduler(s.sig, s, sched))
	}
	s.progress = boolToProgress(s.IsChecked())
}

// Unmount cancels any running animation; bindings are cleaned up by WidgetBase.
func (s *SwitchWidget) Unmount() {
	if s.animCtrl != nil {
		s.animCtrl.CancelAll()
	}
}

// boolToProgress maps the checked state to the settled thumb progress.
func boolToProgress(checked bool) float32 {
	if checked {
		return 1
	}
	return 0
}

// switchProgress adapts the widget's progress field to animation.To's
// signalFloat32 interface, marking the widget dirty on each frame.
type switchProgress struct {
	w   *SwitchWidget
	ctx widget.Context
}

func (a *switchProgress) Get() float32 { return a.w.progress }

func (a *switchProgress) Set(v float32) {
	a.w.progress = v
	a.w.SetNeedsRedraw(true)
	if a.ctx != nil {
		a.ctx.InvalidateRect(a.w.Bounds())
	}
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*SwitchWidget)(nil)
	_ widget.Focusable = (*SwitchWidget)(nil)
	_ widget.Lifecycle = (*SwitchWidget)(nil)
)

package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// SeparatorWidget is shadcn's Separator: a 1px rule in the border token,
// horizontal full-width by default or vertical full-height via Vertical.
//
// Architecture: graft-owned widget (no gogpu/ui core widget wrapped) — a
// separator is a static rule with no interaction machinery, so drawing
// directly via metrics + theme tokens gives full pixel control.
type SeparatorWidget struct {
	widget.WidgetBase

	vertical bool
	theme    *theme.Theme
}

// Separator creates a horizontal 1px divider that spans the available
// width (data-[orientation=horizontal]:h-px w-full, bg-border).
func Separator() *SeparatorWidget {
	s := &SeparatorWidget{}
	s.SetVisible(true)
	s.SetEnabled(true)
	return s
}

// Vertical switches to the vertical orientation: a 1px-wide rule spanning
// the available height (data-[orientation=vertical]:w-px h-full).
func (s *SeparatorWidget) Vertical() *SeparatorWidget {
	s.vertical = true
	return s
}

// Theme pins a specific theme instead of the process-wide current theme.
func (s *SeparatorWidget) Theme(th *theme.Theme) *SeparatorWidget {
	s.theme = th
	return s
}

func (s *SeparatorWidget) resolvedTheme() *theme.Theme {
	if s.theme != nil {
		return s.theme
	}
	return CurrentTheme()
}

// Layout sizes the rule: 1px thick along its cross axis, filling the
// bounded constraint along its main axis (w-full / h-full).
func (s *SeparatorWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	var want geometry.Size
	if s.vertical {
		var h float32
		if c.HasBoundedHeight() {
			h = c.MaxHeight
		}
		want = geometry.Sz(metrics.SeparatorThickness, h)
	} else {
		var w float32
		if c.HasBoundedWidth() {
			w = c.MaxWidth
		}
		want = geometry.Sz(w, metrics.SeparatorThickness)
	}
	size := c.Constrain(want)
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

// Draw fills the rule with the active Border token (bg-border), resolved
// at draw time so mode switches repaint without rebuilding the tree.
func (s *SeparatorWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !s.IsVisible() {
		return
	}
	tok := s.resolvedTheme().Active()
	canvas.DrawRect(s.Bounds(), tok.Border)
}

// Event ignores all input; separators are inert.
func (s *SeparatorWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; SeparatorWidget is a leaf.
func (s *SeparatorWidget) Children() []widget.Widget { return nil }

var _ widget.Widget = (*SeparatorWidget)(nil)

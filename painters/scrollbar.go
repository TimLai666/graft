package painters

import (
	"sync"

	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// PaintScrollbar renders shadcn's scroll-area scrollbar: a fully
// transparent track ("border-l-transparent" gutter) and a pill-shaped
// thumb filled with the border token ("relative flex-1 rounded-full
// bg-border"). Hover and drag keep the same colors — shadcn does not
// restyle the thumb per interaction state.
//
// Geometry note: core/scrollview computes the thumb/track rects itself
// (painters cannot affect layout). Its thumb is 8px thick inset 2px in a
// 12px gutter, while shadcn uses an 8px thumb inset 1px in a 10px gutter
// (metrics.ScrollArea). The thumb thickness — the only visible part —
// matches shadcn's 8px exactly; the extra gutter pixel is invisible
// because the track never paints.
//
// A non-zero PaintState.ColorScheme overrides the theme colors per the
// gogpu/ui painter convention.
func (s Scrollbar) PaintScrollbar(canvas widget.Canvas, ps scrollview.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	var track, thumb widget.Color
	if ps.ColorScheme != (scrollview.ScrollbarColorScheme{}) {
		cs := ps.ColorScheme
		track = cs.Track
		thumb = cs.Thumb
		switch {
		case ps.Dragging && cs.ThumbDrag.A > 0:
			thumb = cs.ThumbDrag
		case ps.Hovered && cs.ThumbHover.A > 0:
			thumb = cs.ThumbHover
		}
	} else {
		// Track stays transparent; thumb is bg-border in every state.
		thumb = s.activeTheme().Active().Border
	}

	if ps.VScrollVisible {
		if track.A > 0 {
			canvas.DrawRoundRect(ps.VTrackRect, track, metrics.ScrollArea.ThumbRadius)
		}
		if !ps.VThumbRect.IsEmpty() {
			canvas.DrawRoundRect(ps.VThumbRect, thumb, metrics.ScrollArea.ThumbRadius)
		}
	}
	if ps.HScrollVisible {
		if track.A > 0 {
			canvas.DrawRoundRect(ps.HTrackRect, track, metrics.ScrollArea.ThumbRadius)
		}
		if !ps.HThumbRect.IsEmpty() {
			canvas.DrawRoundRect(ps.HThumbRect, thumb, metrics.ScrollArea.ThumbRadius)
		}
	}
}

// activeTheme resolves the painter's theme, falling back to the stock
// neutral theme when the bundle was built with a nil theme.
func (s Scrollbar) activeTheme() *theme.Theme {
	if s.Theme != nil {
		return s.Theme
	}
	scrollbarFallbackOnce.Do(func() { scrollbarFallbackTheme = theme.New() })
	return scrollbarFallbackTheme
}

var (
	scrollbarFallbackOnce  sync.Once
	scrollbarFallbackTheme *theme.Theme
)

// Compile-time interface check.
var _ scrollview.Painter = Scrollbar{}

package graft

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// setWidgetBounds positions a child widget when the parent only holds the
// widget.Widget interface (which has no SetBounds). All graft and primitives
// widgets embed widget.WidgetBase, which provides SetBounds; the assertion
// fails safe for any exotic widget lacking it.
//
// Shared by the overlay-hosting components (Tooltip, Dialog) that lay out a
// single content/target child by hand.
func setWidgetBounds(w widget.Widget, b geometry.Rect) {
	if s, ok := w.(interface{ SetBounds(geometry.Rect) }); ok {
		s.SetBounds(b)
	}
}

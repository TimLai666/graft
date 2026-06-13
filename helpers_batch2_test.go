package graft_test

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// fixedWidthLoose returns constraints that pin the width to w but leave the
// height free, the natural input for measuring a w-full component's height.
func fixedWidthLoose(w float32) geometry.Constraints {
	return geometry.Constraints{MinWidth: w, MaxWidth: w, MinHeight: 0, MaxHeight: 100000}
}

// widthBox wraps a w-full widget in a fixed-width, zero-padding Box so golden
// renders of width-filling components (Alert) get a stable, bounded width.
// The default CrossAxisStretch passes the box width straight to the child.
//
// Named with a batch-2 file prefix to avoid colliding with helpers defined by
// other batches at merge time.
func widthBox(child widget.Widget, width float32) widget.Widget {
	return primitives.VBox(child).Width(width).Padding(0)
}

package graft_test

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
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

// zeroDeltaContext returns a MockContext whose DeltaTime is 0, matching the
// deterministic offscreen golden path (widget.NewContext before any BeginFrame).
// Use it when drawing time-driven widgets (skeleton pulse, spinner spin) so the
// animation stays pinned at phase 0 for spec assertions; the default
// uitest.NewMockContext advances 16ms per frame.
func zeroDeltaContext() *uitest.MockContext {
	ctx := uitest.NewMockContext()
	ctx.DeltaVal = 0
	return ctx
}

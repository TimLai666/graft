package graft

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
)

// TestSheetOpenSlideAnimation verifies the open slide-in: progress starts at 0
// (panel fully off its edge), advances monotonically as the controller is
// ticked in Draw, and snaps to a settled 1 (zero offset) once the tween
// completes. Internal test — it reaches the unexported overlay/animation state.
func TestSheetOpenSlideAnimation(t *testing.T) {
	if err := LoadAssets(); err != nil {
		t.Fatal(err)
	}
	content := SheetContent(SheetTitle("Hi")) // default side = right
	win := geometry.Sz(800, 600)
	o := newSheetOverlay(content, win, func() {})

	ctx := uitest.NewMockContext()
	o.Layout(ctx, geometry.Tight(win)) // sets content.Bounds() via layoutAnchored

	o.startOpen(ctx)
	if o.progress != 0 {
		t.Fatalf("progress before any tick = %v, want 0 (panel closed)", o.progress)
	}
	// Closed: a right-anchored panel sits a full width off to the right.
	if off := o.slideOffset(); off.X <= 0 {
		t.Fatalf("closed slide offset.X = %v, want > 0", off.X)
	}

	// Ticking (16ms/frame) must advance progress monotonically toward 1.
	prev := o.progress
	for i := 0; i < 5; i++ {
		o.tickAnim(ctx)
		if o.progress < prev {
			t.Fatalf("progress went backwards: %v -> %v", prev, o.progress)
		}
		prev = o.progress
	}
	if o.progress <= 0 {
		t.Fatalf("progress did not advance after 5 ticks: %v", o.progress)
	}

	// Well past the 500ms duration → settled fully open, zero offset.
	for i := 0; i < 60; i++ {
		o.tickAnim(ctx)
	}
	if o.progress != 1 {
		t.Fatalf("settled progress = %v, want 1", o.progress)
	}
	if off := o.slideOffset(); off != (geometry.Point{}) {
		t.Fatalf("settled slide offset = %v, want (0,0)", off)
	}
}

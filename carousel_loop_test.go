package graft

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
)

// TestCarouselLoopWraps verifies prev/next wrap around the ends when Loop is on.
func TestCarouselLoopWraps(t *testing.T) {
	sig := state.NewSignal(0)
	c := Carousel(
		CarouselItem(primitives.Box()),
		CarouselItem(primitives.Box()),
		CarouselItem(primitives.Box()),
	).Loop(true).Bind(sig)
	ctx := uitest.NewMockContext()

	c.prev(ctx) // at 0 → wrap to last (2)
	if got := sig.Get(); got != 2 {
		t.Fatalf("loop prev from 0 = %d, want 2", got)
	}
	c.next(ctx) // at 2 → wrap to first (0)
	if got := sig.Get(); got != 0 {
		t.Fatalf("loop next from 2 = %d, want 0", got)
	}
}

// TestCarouselNoLoopClamps verifies prev at the start is a no-op without Loop.
func TestCarouselNoLoopClamps(t *testing.T) {
	sig := state.NewSignal(0)
	c := Carousel(
		CarouselItem(primitives.Box()),
		CarouselItem(primitives.Box()),
	).Bind(sig)
	ctx := uitest.NewMockContext()

	c.prev(ctx) // at 0, no loop → stays 0
	if got := sig.Get(); got != 0 {
		t.Fatalf("no-loop prev at start = %d, want 0", got)
	}
}

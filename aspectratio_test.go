package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
)

// aspectRatioChild returns a plain colored box used as the AspectRatio child
// in tests (uniquely named to avoid cross-batch helper collisions).
func aspectRatioChild() *primitives.BoxWidget {
	return primitives.Box().Background(widget.Hex(0x3b82f6)).Rounded(8)
}

// TestAspectRatio16x9 verifies the box derives height = width/ratio.
func TestAspectRatio16x9(t *testing.T) {
	forceLightMode(t)

	ar := graft.AspectRatio(16.0/9.0, aspectRatioChild())
	size := uitest.LayoutWidget(ar, 320, 1000)
	if size.Width != 320 {
		t.Fatalf("width = %v, want 320", size.Width)
	}
	wantH := float32(320) / (16.0 / 9.0)
	if size.Height != wantH {
		t.Fatalf("height = %v, want %v", size.Height, wantH)
	}
}

// TestAspectRatioSquare verifies a 1:1 ratio produces a square.
func TestAspectRatioSquare(t *testing.T) {
	forceLightMode(t)

	ar := graft.AspectRatio(1, aspectRatioChild())
	size := uitest.LayoutWidget(ar, 200, 1000)
	if size != geometry.Sz(200, 200) {
		t.Fatalf("square box = %v, want 200x200", size)
	}
}

// TestAspectRatioDefaultsOnNonPositive verifies a non-positive ratio falls
// back to 16:9.
func TestAspectRatioDefaultsOnNonPositive(t *testing.T) {
	forceLightMode(t)

	ar := graft.AspectRatio(0, aspectRatioChild())
	size := uitest.LayoutWidget(ar, 320, 1000)
	wantH := float32(320) / (16.0 / 9.0)
	if size.Height != wantH {
		t.Fatalf("default ratio height = %v, want %v", size.Height, wantH)
	}
}

// TestAspectRatioFillsChild verifies the child is sized to fill the box.
func TestAspectRatioFillsChild(t *testing.T) {
	forceLightMode(t)

	child := aspectRatioChild()
	ar := graft.AspectRatio(2, child)
	uitest.LayoutWidget(ar, 400, 1000)
	cb := child.Bounds()
	if cb.Width() != 400 || cb.Height() != 200 {
		t.Fatalf("child bounds = %vx%v, want 400x200", cb.Width(), cb.Height())
	}
}

func TestGoldenAspectRatio(t *testing.T) {
	gtest.GoldenLightDark(t, "aspect-ratio-16x9", func() widget.Widget {
		ar := graft.AspectRatio(16.0/9.0, aspectRatioChild()).Ratio(16.0 / 9.0)
		return primitives.Box(ar).Width(320).Padding(24)
	})
}

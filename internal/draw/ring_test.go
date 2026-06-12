package draw

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

func TestFocusRing(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	bounds := geometry.NewRect(10, 20, 100, 36)
	ring := widget.RGBA(0.2, 0.4, 0.9, 0.5)

	FocusRing(canvas, bounds, 8, ring)

	if len(canvas.StrokeRoundRects) != 1 {
		t.Fatalf("got %d StrokeRoundRect calls, want 1", len(canvas.StrokeRoundRects))
	}
	want := uitest.StrokeRoundRectCall{
		Bounds:      bounds.Expand(1.5),
		Color:       ring,
		Radius:      9.5,
		StrokeWidth: 3,
	}
	if got := canvas.StrokeRoundRects[0]; got != want {
		t.Errorf("FocusRing call = %+v, want %+v", got, want)
	}
}

func TestFocusRingWidth(t *testing.T) {
	bounds := geometry.NewRect(0, 0, 16, 16)
	ring := widget.RGBA(1, 0, 0, 0.2)

	tests := []struct {
		name   string
		radius float32
		width  float32
		want   uitest.StrokeRoundRectCall
	}{
		{
			"slider thumb width 4",
			8, 4,
			uitest.StrokeRoundRectCall{Bounds: bounds.Expand(2), Color: ring, Radius: 10, StrokeWidth: 4},
		},
		{
			"width 2",
			4, 2,
			uitest.StrokeRoundRectCall{Bounds: bounds.Expand(1), Color: ring, Radius: 5, StrokeWidth: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &uitest.MockCanvas{}
			FocusRingWidth(canvas, bounds, tt.radius, ring, tt.width)
			if len(canvas.StrokeRoundRects) != 1 {
				t.Fatalf("got %d StrokeRoundRect calls, want 1", len(canvas.StrokeRoundRects))
			}
			if got := canvas.StrokeRoundRects[0]; got != tt.want {
				t.Errorf("call = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFocusRingWidthZeroIsNoop(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	FocusRingWidth(canvas, geometry.NewRect(0, 0, 10, 10), 4, widget.ColorRed, 0)
	if n := canvas.TotalDrawCalls(); n != 0 {
		t.Errorf("zero-width ring drew %d calls, want 0", n)
	}
}

func TestOffsetRing(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	bounds := geometry.NewRect(16, 16, 16, 16)
	ring := widget.RGBA(0.2, 0.4, 0.9, 1)
	gap := widget.RGBA(1, 1, 1, 1)

	// Legacy close button recipe: ring-2 ring-offset-2.
	OffsetRing(canvas, bounds, 4, ring, 2, 2, gap)

	if len(canvas.StrokeRoundRects) != 2 {
		t.Fatalf("got %d StrokeRoundRect calls, want 2", len(canvas.StrokeRoundRects))
	}
	wantGap := uitest.StrokeRoundRectCall{
		Bounds:      bounds.Expand(1),
		Color:       gap,
		Radius:      5,
		StrokeWidth: 2,
	}
	if got := canvas.StrokeRoundRects[0]; got != wantGap {
		t.Errorf("gap stroke = %+v, want %+v", got, wantGap)
	}
	wantRing := uitest.StrokeRoundRectCall{
		Bounds:      bounds.Expand(3),
		Color:       ring,
		Radius:      7,
		StrokeWidth: 2,
	}
	if got := canvas.StrokeRoundRects[1]; got != wantRing {
		t.Errorf("ring stroke = %+v, want %+v", got, wantRing)
	}
}

func TestOffsetRingNoGapFill(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	bounds := geometry.NewRect(0, 0, 10, 10)
	ring := widget.ColorBlue

	// Transparent gap fill suppresses the gap stroke.
	OffsetRing(canvas, bounds, 2, ring, 2, 2, widget.ColorTransparent)

	if len(canvas.StrokeRoundRects) != 1 {
		t.Fatalf("got %d StrokeRoundRect calls, want 1", len(canvas.StrokeRoundRects))
	}
	want := uitest.StrokeRoundRectCall{
		Bounds:      bounds.Expand(3),
		Color:       ring,
		Radius:      5,
		StrokeWidth: 2,
	}
	if got := canvas.StrokeRoundRects[0]; got != want {
		t.Errorf("ring stroke = %+v, want %+v", got, want)
	}
}

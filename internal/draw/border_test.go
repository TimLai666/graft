package draw

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

func TestInsideBorder(t *testing.T) {
	bounds := geometry.NewRect(10, 10, 100, 40)
	col := widget.RGBA(0.5, 0.5, 0.5, 1)

	tests := []struct {
		name   string
		radius float32
		w      float32
		want   uitest.StrokeRoundRectCall
	}{
		{
			"1px border radius 8",
			8, 1,
			uitest.StrokeRoundRectCall{Bounds: bounds.Expand(-0.5), Color: col, Radius: 7.5, StrokeWidth: 1},
		},
		{
			"2px border radius 6",
			6, 2,
			uitest.StrokeRoundRectCall{Bounds: bounds.Expand(-1), Color: col, Radius: 5, StrokeWidth: 2},
		},
		{
			"radius smaller than half width clamps to zero",
			0.25, 1,
			uitest.StrokeRoundRectCall{Bounds: bounds.Expand(-0.5), Color: col, Radius: 0, StrokeWidth: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &uitest.MockCanvas{}
			InsideBorder(canvas, bounds, tt.radius, col, tt.w)
			if len(canvas.StrokeRoundRects) != 1 {
				t.Fatalf("got %d StrokeRoundRect calls, want 1", len(canvas.StrokeRoundRects))
			}
			if got := canvas.StrokeRoundRects[0]; got != tt.want {
				t.Errorf("call = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestInsideBorderZeroWidthIsNoop(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	InsideBorder(canvas, geometry.NewRect(0, 0, 10, 10), 4, widget.ColorBlack, 0)
	if n := canvas.TotalDrawCalls(); n != 0 {
		t.Errorf("zero-width border drew %d calls, want 0", n)
	}
}

func TestCornersHas(t *testing.T) {
	tests := []struct {
		name string
		mask Corners
		c    Corners
		want bool
	}{
		{"single set", TopLeft, TopLeft, true},
		{"single unset", TopLeft, TopRight, false},
		{"combined contains member", TopLeft | BottomRight, BottomRight, true},
		{"combined missing member", TopLeft | BottomRight, BottomLeft, false},
		{"all contains pair", AllCorners, TopLeft | BottomLeft, true},
		{"partial does not contain pair", TopLeft, TopLeft | TopRight, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mask.Has(tt.c); got != tt.want {
				t.Errorf("(%b).Has(%b) = %v, want %v", tt.mask, tt.c, got, tt.want)
			}
		})
	}
}

func TestSquareCorners(t *testing.T) {
	bounds := geometry.NewRect(0, 0, 100, 40)
	fill := widget.RGBA(0.9, 0.9, 0.9, 1)

	canvas := &uitest.MockCanvas{}
	SquareCorners(canvas, bounds, 8, fill, TopLeft|BottomRight)

	if len(canvas.RoundRects) != 1 {
		t.Fatalf("got %d DrawRoundRect calls, want 1", len(canvas.RoundRects))
	}
	wantFill := uitest.DrawRoundRectCall{Bounds: bounds, Color: fill, Radius: 8}
	if got := canvas.RoundRects[0]; got != wantFill {
		t.Errorf("round-rect fill = %+v, want %+v", got, wantFill)
	}

	if len(canvas.Rects) != 2 {
		t.Fatalf("got %d DrawRect calls, want 2", len(canvas.Rects))
	}
	wantTL := uitest.DrawRectCall{Bounds: geometry.NewRect(0, 0, 8, 8), Color: fill}
	if got := canvas.Rects[0]; got != wantTL {
		t.Errorf("top-left overpaint = %+v, want %+v", got, wantTL)
	}
	wantBR := uitest.DrawRectCall{Bounds: geometry.NewRect(92, 32, 8, 8), Color: fill}
	if got := canvas.Rects[1]; got != wantBR {
		t.Errorf("bottom-right overpaint = %+v, want %+v", got, wantBR)
	}
}

func TestSquareCornersAll(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	SquareCorners(canvas, geometry.NewRect(0, 0, 40, 40), 6, widget.ColorWhite, AllCorners)
	if len(canvas.RoundRects) != 1 || len(canvas.Rects) != 4 {
		t.Errorf("got %d round-rects and %d rects, want 1 and 4",
			len(canvas.RoundRects), len(canvas.Rects))
	}
}

func TestSquareCornersNoneOrZeroRadius(t *testing.T) {
	tests := []struct {
		name    string
		radius  float32
		corners Corners
	}{
		{"no corners selected", 8, 0},
		{"zero radius", 0, AllCorners},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &uitest.MockCanvas{}
			SquareCorners(canvas, geometry.NewRect(0, 0, 40, 40), tt.radius, widget.ColorWhite, tt.corners)
			if len(canvas.RoundRects) != 1 {
				t.Fatalf("got %d DrawRoundRect calls, want 1", len(canvas.RoundRects))
			}
			if len(canvas.Rects) != 0 {
				t.Errorf("got %d DrawRect overpaints, want 0", len(canvas.Rects))
			}
		})
	}
}

func TestSquareCornersClampsQuadrantToHalfExtent(t *testing.T) {
	// Pill radius (9999) on a 40x20 rect: the renderer clamps the radius to
	// half the smaller dimension (10), so the quadrant must clamp the same.
	canvas := &uitest.MockCanvas{}
	bounds := geometry.NewRect(0, 0, 40, 20)
	fill := widget.ColorWhite
	SquareCorners(canvas, bounds, 9999, fill, TopRight)

	if len(canvas.Rects) != 1 {
		t.Fatalf("got %d DrawRect calls, want 1", len(canvas.Rects))
	}
	want := uitest.DrawRectCall{Bounds: geometry.NewRect(30, 0, 10, 10), Color: fill}
	if got := canvas.Rects[0]; got != want {
		t.Errorf("clamped overpaint = %+v, want %+v", got, want)
	}
}

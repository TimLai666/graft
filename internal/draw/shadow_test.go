package draw

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
)

func TestShadowEmitsOneRoundRectPerLayer(t *testing.T) {
	tests := []struct {
		name   string
		layers []metrics.ShadowLayer
	}{
		{"XS", metrics.ShadowXS},
		{"SM", metrics.ShadowSM},
		{"MD", metrics.ShadowMD},
		{"LG", metrics.ShadowLG},
	}
	bounds := geometry.NewRect(20, 30, 200, 100)
	const radius float32 = 10
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &uitest.MockCanvas{}
			Shadow(canvas, bounds, radius, tt.layers)
			if len(canvas.RoundRects) != len(tt.layers) {
				t.Fatalf("got %d DrawRoundRect calls, want %d", len(canvas.RoundRects), len(tt.layers))
			}
			for i, l := range tt.layers {
				want := uitest.DrawRoundRectCall{
					Bounds: bounds.Expand(l.Grow).TranslateXY(0, l.DY),
					Color:  widget.RGBA(0, 0, 0, l.Alpha),
					Radius: radius + l.Grow,
				}
				if got := canvas.RoundRects[i]; got != want {
					t.Errorf("layer %d = %+v, want %+v", i, got, want)
				}
			}
		})
	}
}

func TestShadowSMExactCalls(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	bounds := geometry.NewRect(0, 0, 100, 50)
	Shadow(canvas, bounds, 14, metrics.ShadowSM)

	want := []uitest.DrawRoundRectCall{
		{Bounds: geometry.NewRect(0, 1, 100, 50), Color: widget.RGBA(0, 0, 0, 0.06), Radius: 14},
		{Bounds: geometry.NewRect(-1, 0, 102, 52), Color: widget.RGBA(0, 0, 0, 0.05), Radius: 15},
		{Bounds: geometry.NewRect(-2, 0, 104, 54), Color: widget.RGBA(0, 0, 0, 0.03), Radius: 16},
	}
	if len(canvas.RoundRects) != len(want) {
		t.Fatalf("got %d calls, want %d", len(canvas.RoundRects), len(want))
	}
	for i := range want {
		if canvas.RoundRects[i] != want[i] {
			t.Errorf("call %d = %+v, want %+v", i, canvas.RoundRects[i], want[i])
		}
	}
}

func TestShadowEmptyLayersIsNoop(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	Shadow(canvas, geometry.NewRect(0, 0, 10, 10), 4, nil)
	if n := canvas.TotalDrawCalls(); n != 0 {
		t.Errorf("nil layers drew %d calls, want 0", n)
	}
}

package draw

import (
	"testing"

	"github.com/gogpu/ui/widget"
)

func TestAlpha(t *testing.T) {
	tests := []struct {
		name string
		in   widget.Color
		a    float32
		want widget.Color
	}{
		{"replace on opaque", widget.RGBA(0.2, 0.4, 0.6, 1), 0.5, widget.RGBA(0.2, 0.4, 0.6, 0.5)},
		{"replace on translucent", widget.RGBA(1, 0, 0, 0.1), 0.9, widget.RGBA(1, 0, 0, 0.9)},
		{"zero alpha", widget.RGBA(1, 1, 1, 1), 0, widget.RGBA(1, 1, 1, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Alpha(tt.in, tt.a); got != tt.want {
				t.Errorf("Alpha(%v, %v) = %v, want %v", tt.in, tt.a, got, tt.want)
			}
		})
	}
}

func TestMulAlpha(t *testing.T) {
	tests := []struct {
		name string
		in   widget.Color
		f    float32
		want widget.Color
	}{
		{"halve opaque", widget.RGBA(0.2, 0.4, 0.6, 1), 0.5, widget.RGBA(0.2, 0.4, 0.6, 0.5)},
		{"halve dark border token", widget.RGBA(1, 1, 1, 0.1), 0.5, widget.RGBA(1, 1, 1, 0.05)},
		{"identity", widget.RGBA(0, 1, 0, 0.8), 1, widget.RGBA(0, 1, 0, 0.8)},
		{"to zero", widget.RGBA(0, 1, 0, 0.8), 0, widget.RGBA(0, 1, 0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MulAlpha(tt.in, tt.f); got != tt.want {
				t.Errorf("MulAlpha(%v, %v) = %v, want %v", tt.in, tt.f, got, tt.want)
			}
		})
	}
}

func TestFade(t *testing.T) {
	c := widget.RGBA(0.1, 0.2, 0.3, 0.8)
	if got := Fade(c, false); got != c {
		t.Errorf("Fade(c, false) = %v, want unchanged %v", got, c)
	}
	want := widget.RGBA(0.1, 0.2, 0.3, 0.4)
	if got := Fade(c, true); got != want {
		t.Errorf("Fade(c, true) = %v, want %v", got, want)
	}
}

// colorNear reports whether two colors match within eps per channel.
func colorNear(a, b widget.Color, eps float32) bool {
	abs := func(v float32) float32 {
		if v < 0 {
			return -v
		}
		return v
	}
	return abs(a.R-b.R) <= eps && abs(a.G-b.G) <= eps &&
		abs(a.B-b.B) <= eps && abs(a.A-b.A) <= eps
}

func TestOver(t *testing.T) {
	tests := []struct {
		name string
		bg   widget.Color
		fg   widget.Color
		want widget.Color
	}{
		{
			"opaque fg wins",
			widget.RGBA(0, 0, 1, 1),
			widget.RGBA(1, 0, 0, 1),
			widget.RGBA(1, 0, 0, 1),
		},
		{
			"transparent fg keeps bg",
			widget.RGBA(0, 0, 1, 1),
			widget.RGBA(1, 0, 0, 0),
			widget.RGBA(0, 0, 1, 1),
		},
		{
			"half fg over opaque bg averages",
			widget.RGBA(0, 0, 1, 1),
			widget.RGBA(1, 0, 0, 0.5),
			widget.RGBA(0.5, 0, 0.5, 1),
		},
		{
			"primary/90 over white",
			widget.RGBA(1, 1, 1, 1),
			widget.RGBA(0.2, 0.2, 0.2, 0.9),
			widget.RGBA(0.28, 0.28, 0.28, 1),
		},
		{
			"both translucent",
			widget.RGBA(0, 0, 1, 0.5),
			widget.RGBA(1, 0, 0, 0.5),
			// aOut = 0.5 + 0.5*0.5 = 0.75; R = 0.5/0.75, B = 0.25/0.75.
			widget.RGBA(2.0/3.0, 0, 1.0/3.0, 0.75),
		},
		{
			"both transparent",
			widget.RGBA(0, 0, 1, 0),
			widget.RGBA(1, 0, 0, 0),
			widget.Color{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Over(tt.bg, tt.fg)
			if !colorNear(got, tt.want, 1e-6) {
				t.Errorf("Over(%v, %v) = %v, want %v", tt.bg, tt.fg, got, tt.want)
			}
		})
	}
}

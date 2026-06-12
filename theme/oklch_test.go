package theme

import (
	"math"
	"testing"

	"github.com/gogpu/ui/widget"
)

// Reference values from the Tailwind CSS v4 palette, which is authored in
// OKLCH with published sRGB hex fallbacks (gamut handling = linear-channel
// clipping, same as browsers on sRGB displays). shadcn/ui tokens draw from
// this palette, so matching these means matching shadcn pixel-for-pixel.
func TestOKLCHAgainstTailwindV4Palette(t *testing.T) {
	cases := []struct {
		name    string
		l, c, h float64
		hex     uint32
	}{
		{"white", 1, 0, 0, 0xffffff},
		{"black", 0, 0, 0, 0x000000},
		{"neutral-50", 0.985, 0, 0, 0xfafafa},
		{"neutral-100", 0.97, 0, 0, 0xf5f5f5},
		{"neutral-200", 0.922, 0, 0, 0xe5e5e5},
		{"neutral-400", 0.708, 0, 0, 0xa1a1a1},
		{"neutral-500", 0.556, 0, 0, 0x737373},
		{"neutral-900", 0.205, 0, 0, 0x171717},
		{"neutral-950", 0.145, 0, 0, 0x0a0a0a},
		// In-gamut chromatic.
		{"red-500", 0.637, 0.237, 25.331, 0xfb2c36},
		{"blue-600", 0.546, 0.245, 262.881, 0x155dfc},
		// Out-of-gamut: exercises the clipping path. This is shadcn's
		// --destructive token; #e7000b is Tailwind's published fallback.
		{"red-600", 0.577, 0.245, 27.325, 0xe7000b},
	}

	const tolerance = 1.5 / 255 // off-by-one from palette rounding

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := OKLCH(tc.l, tc.c, tc.h)
			want := widget.Hex(tc.hex)
			for _, ch := range []struct {
				name      string
				got, want float32
			}{
				{"R", got.R, want.R},
				{"G", got.G, want.G},
				{"B", got.B, want.B},
			} {
				if diff := math.Abs(float64(ch.got - ch.want)); diff > tolerance {
					t.Errorf("%s: got %.4f want %.4f (diff %.4f = %.1f/255)",
						ch.name, ch.got, ch.want, diff, diff*255)
				}
			}
			if got.A != 1 {
				t.Errorf("alpha: got %v want 1", got.A)
			}
		})
	}
}

// srgbToOKLCH is an independent inverse transform (sRGB → OKLab → LCH) used
// to verify the forward path by round-tripping.
func srgbToOKLCH(r, g, b float64) (l, c, h float64) {
	toLin := func(c float64) float64 {
		if c <= 0.04045 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	rl, gl, bl := toLin(r), toLin(g), toLin(b)

	lm := 0.4122214708*rl + 0.5363325363*gl + 0.0514459929*bl
	mm := 0.2119034982*rl + 0.6806995451*gl + 0.1073969566*bl
	sm := 0.0883024619*rl + 0.2817188376*gl + 0.6299787005*bl

	l_, m_, s_ := math.Cbrt(lm), math.Cbrt(mm), math.Cbrt(sm)

	L := 0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_
	a := 1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_
	bb := 0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_

	c = math.Hypot(a, bb)
	h = math.Atan2(bb, a) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return L, c, h
}

func TestOKLCHRoundTrip(t *testing.T) {
	// A grid of in-gamut sRGB colors must survive sRGB → OKLCH → sRGB.
	steps := []float64{0, 0.13, 0.35, 0.5, 0.68, 0.87, 1}
	for _, r := range steps {
		for _, g := range steps {
			for _, b := range steps {
				l, c, h := srgbToOKLCH(r, g, b)
				got := OKLCH(l, c, h)
				const tol = 1e-3
				if math.Abs(float64(got.R)-r) > tol ||
					math.Abs(float64(got.G)-g) > tol ||
					math.Abs(float64(got.B)-b) > tol {
					t.Fatalf("round-trip (%.2f,%.2f,%.2f) → oklch(%.4f %.4f %.2f) → (%.4f,%.4f,%.4f)",
						r, g, b, l, c, h, got.R, got.G, got.B)
				}
			}
		}
	}
}

func TestParseOKLCH(t *testing.T) {
	cases := []struct {
		in        string
		want      widget.Color
		tolerance float64
	}{
		{"oklch(1 0 0)", widget.RGBA(1, 1, 1, 1), 0.002},
		{"oklch(0.145 0 0)", widget.Hex(0x0a0a0a), 0.01},
		{"oklch(1 0 0 / 10%)", widget.RGBA(1, 1, 1, 0.1), 0.002},
		{"oklch(0.708 0 0 / 0.5)", widget.Hex(0xa1a1a1).WithAlpha(0.5), 0.01},
		{"  oklch(0.577 0.245 27.325)  ", widget.Hex(0xe7000b), 0.01},
		{"oklch(63.7% 0.237 25.331)", widget.Hex(0xfb2c36), 0.01},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseOKLCH(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range []float64{
				math.Abs(float64(got.R - tc.want.R)),
				math.Abs(float64(got.G - tc.want.G)),
				math.Abs(float64(got.B - tc.want.B)),
				math.Abs(float64(got.A - tc.want.A)),
			} {
				if d > tc.tolerance {
					t.Fatalf("got %+v want %+v (channel diff %.4f)", got, tc.want, d)
				}
			}
		})
	}

	for _, bad := range []string{"", "rgb(1 2 3)", "oklch(1 0)", "oklch(1 0 0", "oklch(x 0 0)"} {
		if _, err := ParseOKLCH(bad); err == nil {
			t.Errorf("ParseOKLCH(%q): expected error, got nil", bad)
		}
	}
}

func TestOutOfGamutClipsInRange(t *testing.T) {
	// Wildly out-of-gamut inputs must still produce valid colors.
	for _, tc := range []struct{ l, c, h float64 }{
		{0.8, 0.4, 145},  // hyper-saturated green
		{0.5, 0.45, 300}, // hyper-saturated purple
		{1.2, 0.1, 0},    // lightness above 1
		{-0.1, 0.1, 0},   // lightness below 0
	} {
		got := OKLCH(tc.l, tc.c, tc.h)
		for _, v := range []float32{got.R, got.G, got.B, got.A} {
			if v < 0 || v > 1 {
				t.Fatalf("OKLCH(%v, %v, %v) out of range: %+v", tc.l, tc.c, tc.h, got)
			}
		}
	}
}

package theme

import (
	"testing"

	"github.com/gogpu/ui/widget"
)

// nonZero reports whether a color is set to something visible: any channel
// non-zero with a non-zero alpha (dark border oklch(1 0 0 / 10%) counts).
func nonZero(c widget.Color) bool {
	return c.A > 0 && (c.R > 0 || c.G > 0 || c.B > 0 || c.A < 1)
}

func TestAllPresetsParse(t *testing.T) {
	bases := []Base{Neutral, Stone, Zinc, Gray, Slate}
	for _, b := range bases {
		b := b
		t.Run(b.String(), func(t *testing.T) {
			th := New(BaseColor(b))

			if th.Light.Background == th.Dark.Background {
				t.Errorf("light Background == dark Background (%+v)", th.Light.Background)
			}
			for _, mode := range []struct {
				name string
				tok  *Tokens
			}{
				{"light", &th.Light},
				{"dark", &th.Dark},
			} {
				if !nonZero(mode.tok.Primary) {
					t.Errorf("%s Primary is zero", mode.name)
				}
				if !nonZero(mode.tok.Ring) {
					t.Errorf("%s Ring is zero", mode.name)
				}
				if !nonZero(mode.tok.Border) {
					t.Errorf("%s Border is zero", mode.name)
				}
				if !nonZero(mode.tok.Foreground) {
					t.Errorf("%s Foreground is zero", mode.name)
				}
				if !nonZero(mode.tok.DestructiveForeground) {
					t.Errorf("%s DestructiveForeground is zero", mode.name)
				}
				for i, c := range mode.tok.Chart {
					if !nonZero(c) {
						t.Errorf("%s Chart[%d] is zero", mode.name, i)
					}
				}
			}
			if th.Radius != 10 {
				t.Errorf("Radius = %v, want 10", th.Radius)
			}
			// All presets ship the dark alpha border/input tokens.
			if got := th.Dark.Border; got != MustOKLCH("oklch(1 0 0 / 10%)") {
				t.Errorf("dark Border = %+v, want oklch(1 0 0 / 10%%)", got)
			}
			if got := th.Dark.Input; got != MustOKLCH("oklch(1 0 0 / 15%)") {
				t.Errorf("dark Input = %+v, want oklch(1 0 0 / 15%%)", got)
			}
		})
	}
}

func TestPresetSpotValues(t *testing.T) {
	// One distinctive value per non-neutral preset, straight from the
	// upstream cssVarsV4 blocks (see presets/*.css headers).
	cases := []struct {
		base  Base
		pick  func(*Theme) widget.Color
		name  string
		oklch string
	}{
		{Stone, func(t *Theme) widget.Color { return t.Light.Foreground }, "stone Light.Foreground", "oklch(0.147 0.004 49.25)"},
		{Stone, func(t *Theme) widget.Color { return t.Dark.Primary }, "stone Dark.Primary", "oklch(0.923 0.003 48.717)"},
		{Zinc, func(t *Theme) widget.Color { return t.Light.Primary }, "zinc Light.Primary", "oklch(0.21 0.006 285.885)"},
		{Zinc, func(t *Theme) widget.Color { return t.Dark.Ring }, "zinc Dark.Ring", "oklch(0.552 0.016 285.938)"},
		{Gray, func(t *Theme) widget.Color { return t.Light.Foreground }, "gray Light.Foreground", "oklch(0.13 0.028 261.692)"},
		{Gray, func(t *Theme) widget.Color { return t.Light.Border }, "gray Light.Border", "oklch(0.928 0.006 264.531)"},
		{Slate, func(t *Theme) widget.Color { return t.Light.Primary }, "slate Light.Primary", "oklch(0.208 0.042 265.755)"},
		{Slate, func(t *Theme) widget.Color { return t.Dark.Secondary }, "slate Dark.Secondary", "oklch(0.279 0.041 260.031)"},
	}
	for _, tc := range cases {
		th := New(BaseColor(tc.base))
		if got, want := tc.pick(th), MustOKLCH(tc.oklch); got != want {
			t.Errorf("%s = %+v, want %s = %+v", tc.name, got, tc.oklch, want)
		}
	}
}

func TestPresetsDistinct(t *testing.T) {
	// Integration check: the five embedded presets must yield pairwise
	// distinct palettes in both modes. Tokens is comparable (colors and a
	// fixed-size array only), so whole-struct equality is exact.
	bases := []Base{Neutral, Stone, Zinc, Gray, Slate}
	themes := make([]*Theme, len(bases))
	for i, b := range bases {
		themes[i] = New(BaseColor(b))
	}
	for i := 0; i < len(bases); i++ {
		for j := i + 1; j < len(bases); j++ {
			if themes[i].Light == themes[j].Light {
				t.Errorf("%s and %s produce identical Light palettes", bases[i], bases[j])
			}
			if themes[i].Dark == themes[j].Dark {
				t.Errorf("%s and %s produce identical Dark palettes", bases[i], bases[j])
			}
		}
	}
}

func TestPresetIsolation(t *testing.T) {
	// Mutating one theme must not bleed into the preset cache.
	a := New()
	a.Light.Primary = OKLCH(0.5, 0.1, 100)
	b := New()
	if b.Light.Primary == a.Light.Primary {
		t.Error("mutating a theme leaked into the preset cache")
	}
}

func TestBaseString(t *testing.T) {
	cases := map[Base]string{
		Neutral:  "neutral",
		Stone:    "stone",
		Zinc:     "zinc",
		Gray:     "gray",
		Slate:    "slate",
		Base(42): "unknown",
	}
	for b, want := range cases {
		if got := b.String(); got != want {
			t.Errorf("Base(%d).String() = %q, want %q", b, got, want)
		}
	}
}

func TestPresetUnknownBasePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("preset(unknown) must panic")
		}
	}()
	preset(Base(42))
}

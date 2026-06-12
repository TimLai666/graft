package theme

import (
	"math"
	"testing"

	"github.com/gogpu/ui/widget"
)

// colorEq compares colors with a small epsilon (parsing paths may differ
// from the reference computation by float rounding only).
func colorEq(a, b widget.Color, eps float32) bool {
	abs := func(v float32) float32 { return float32(math.Abs(float64(v))) }
	return abs(a.R-b.R) <= eps && abs(a.G-b.G) <= eps &&
		abs(a.B-b.B) <= eps && abs(a.A-b.A) <= eps
}

func TestNeutralPresetSpotValues(t *testing.T) {
	th := New() // Neutral default

	cases := []struct {
		name string
		got  widget.Color
		want widget.Color
	}{
		{"Light.Background", th.Light.Background, MustOKLCH("oklch(1 0 0)")},
		{"Light.Foreground", th.Light.Foreground, MustOKLCH("oklch(0.145 0 0)")},
		{"Light.Primary", th.Light.Primary, MustOKLCH("oklch(0.205 0 0)")},
		{"Light.Destructive", th.Light.Destructive, MustOKLCH("oklch(0.577 0.245 27.325)")},
		{"Light.Ring", th.Light.Ring, MustOKLCH("oklch(0.708 0 0)")},
		{"Light.Chart[0]", th.Light.Chart[0], MustOKLCH("oklch(0.646 0.222 41.116)")},
		{"Dark.Background", th.Dark.Background, MustOKLCH("oklch(0.145 0 0)")},
		{"Dark.Border", th.Dark.Border, MustOKLCH("oklch(1 0 0 / 10%)")},
		{"Dark.Input", th.Dark.Input, MustOKLCH("oklch(1 0 0 / 15%)")},
		{"Dark.Chart[4]", th.Dark.Chart[4], MustOKLCH("oklch(0.645 0.246 16.439)")},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = %+v, want %+v", tc.name, tc.got, tc.want)
		}
	}

	if th.Radius != 10 {
		t.Errorf("Radius = %v, want 10 (--radius: 0.625rem)", th.Radius)
	}
	// Canonical theme omits --destructive-foreground; default is white.
	if want := widget.RGB(1, 1, 1); th.Light.DestructiveForeground != want ||
		th.Dark.DestructiveForeground != want {
		t.Errorf("DestructiveForeground = %+v / %+v, want white",
			th.Light.DestructiveForeground, th.Dark.DestructiveForeground)
	}
}

func TestApplyCSSPartialOverride(t *testing.T) {
	th := New()
	wantSecondary := th.Light.Secondary // must survive the override
	wantDarkPrimary := th.Dark.Primary

	css := `
		/* tweakcn-style partial export */
		:root {
			--primary: oklch(0.55 0.2 260);
			--radius: 0.5rem;
		}
		.dark {
			--ring: #ff0000;
		}
	`
	if err := th.ApplyCSS(css); err != nil {
		t.Fatalf("ApplyCSS: %v", err)
	}

	if want := MustOKLCH("oklch(0.55 0.2 260)"); th.Light.Primary != want {
		t.Errorf("Light.Primary = %+v, want %+v", th.Light.Primary, want)
	}
	if want := (widget.Color{R: 1, G: 0, B: 0, A: 1}); th.Dark.Ring != want {
		t.Errorf("Dark.Ring = %+v, want %+v", th.Dark.Ring, want)
	}
	if th.Radius != 8 {
		t.Errorf("Radius = %v, want 8 (0.5rem)", th.Radius)
	}
	// Untouched values stay put: the :root block must not leak into Dark,
	// and unlisted tokens keep their preset values.
	if th.Light.Secondary != wantSecondary {
		t.Errorf("Light.Secondary changed: %+v", th.Light.Secondary)
	}
	if th.Dark.Primary != wantDarkPrimary {
		t.Errorf("Dark.Primary changed by :root block: %+v", th.Dark.Primary)
	}
}

func TestApplyCSSSelectorVariants(t *testing.T) {
	th := New()
	css := `:root, .light { --background: #102030; }
	        .dark, [data-theme="dark"] { --background: #405060; }
	        .unrelated { --background: #ffffff; }`
	if err := th.ApplyCSS(css); err != nil {
		t.Fatalf("ApplyCSS: %v", err)
	}
	lightWant, _ := parseHexColor("#102030")
	darkWant, _ := parseHexColor("#405060")
	if th.Light.Background != lightWant {
		t.Errorf("Light.Background = %+v, want %+v", th.Light.Background, lightWant)
	}
	if th.Dark.Background != darkWant {
		t.Errorf("Dark.Background = %+v, want %+v", th.Dark.Background, darkWant)
	}
}

func TestApplyCSSUnknownVariablesIgnored(t *testing.T) {
	th := New()
	before := th.Light
	css := `:root { --surface: oklch(0.98 0 0); --font-sans: Inter; --shadow-2xs: 0 1px 3px 0px hsl(0 0% 0% / 0.05); }`
	if err := th.ApplyCSS(css); err != nil {
		t.Fatalf("ApplyCSS should ignore unknown variables, got %v", err)
	}
	if th.Light != before {
		t.Error("unknown variables must not modify tokens")
	}
}

func TestApplyCSSBadValueErrors(t *testing.T) {
	th := New()
	if err := th.ApplyCSS(`:root { --primary: not-a-color; }`); err == nil {
		t.Error("expected error for unparseable known variable")
	}
	if err := th.ApplyCSS(`:root { --radius: banana; }`); err == nil {
		t.Error("expected error for unparseable --radius")
	}
}

func TestParseThemeCSSStartsFromNeutral(t *testing.T) {
	th, err := ParseThemeCSS(`:root { --primary: #336699; }`)
	if err != nil {
		t.Fatalf("ParseThemeCSS: %v", err)
	}
	want, _ := parseHexColor("#336699")
	if th.Light.Primary != want {
		t.Errorf("Light.Primary = %+v, want %+v", th.Light.Primary, want)
	}
	// Everything else is stock neutral.
	neutral := New()
	if th.Light.Background != neutral.Light.Background {
		t.Errorf("Light.Background = %+v, want neutral %+v",
			th.Light.Background, neutral.Light.Background)
	}
	if th.Dark.Primary != neutral.Dark.Primary {
		t.Errorf("Dark.Primary = %+v, want neutral %+v",
			th.Dark.Primary, neutral.Dark.Primary)
	}
}

func TestParseCSSColor(t *testing.T) {
	const eps = 1e-6
	cases := []struct {
		in   string
		want widget.Color
	}{
		// oklch
		{"oklch(1 0 0)", widget.Color{R: 1, G: 1, B: 1, A: 1}},
		{"oklch(0.708 0 0 / 50%)", OKLCHA(0.708, 0, 0, 0.5)},
		// hsl: modern space syntax, legacy commas, hsla, alpha slash
		{"hsl(0 0% 100%)", widget.Color{R: 1, G: 1, B: 1, A: 1}},
		{"hsl(0, 100%, 50%)", widget.Color{R: 1, G: 0, B: 0, A: 1}},
		{"hsla(120, 100%, 25%, 0.5)", widget.Color{R: 0, G: 0.5, B: 0, A: 0.5}},
		{"hsl(240 100% 50% / 25%)", widget.Color{R: 0, G: 0, B: 1, A: 0.25}},
		// legacy shadcn bare triplet
		{"0 0% 100%", widget.Color{R: 1, G: 1, B: 1, A: 1}},
		// rgb
		{"rgb(255, 0, 0)", widget.Color{R: 1, G: 0, B: 0, A: 1}},
		{"rgba(0, 0, 0, 0.5)", widget.Color{R: 0, G: 0, B: 0, A: 0.5}},
		{"rgb(255 0 0 / 50%)", widget.Color{R: 1, G: 0, B: 0, A: 0.5}},
		{"rgb(100% 0% 0%)", widget.Color{R: 1, G: 0, B: 0, A: 1}},
		// hex
		{"#fff", widget.Color{R: 1, G: 1, B: 1, A: 1}},
		{"#f003", widget.Color{R: 1, G: 0, B: 0, A: 0.2}},
		{"#e7000b", widget.Color{R: 231.0 / 255, G: 0, B: 11.0 / 255, A: 1}},
		{"#0a0a0a80", widget.Color{R: 10.0 / 255, G: 10.0 / 255, B: 10.0 / 255, A: 128.0 / 255}},
	}
	for _, tc := range cases {
		got, err := parseCSSColor(tc.in)
		if err != nil {
			t.Errorf("parseCSSColor(%q): %v", tc.in, err)
			continue
		}
		if !colorEq(got, tc.want, eps) {
			t.Errorf("parseCSSColor(%q) = %+v, want %+v", tc.in, got, tc.want)
		}
	}

	// Bare-triplet equivalence with the wrapped form.
	bare, err1 := parseCSSColor("224 71.4% 4.1%")
	wrapped, err2 := parseCSSColor("hsl(224 71.4% 4.1%)")
	if err1 != nil || err2 != nil {
		t.Fatalf("triplet parse errors: %v / %v", err1, err2)
	}
	if bare != wrapped {
		t.Errorf("bare triplet %+v != hsl() form %+v", bare, wrapped)
	}

	for _, bad := range []string{"", "blue", "oklch(", "#12", "rgb(1,2)", "hsl(x 0% 0%)"} {
		if _, err := parseCSSColor(bad); err == nil {
			t.Errorf("parseCSSColor(%q): expected error", bad)
		}
	}
}

func TestParseCSSLength(t *testing.T) {
	cases := []struct {
		in   string
		want float32
	}{
		{"0.625rem", 10},
		{"0.5rem", 8},
		{"1rem", 16},
		{"10px", 10},
		{"12px", 12},
		{"7", 7},
		{"0", 0},
	}
	for _, tc := range cases {
		got, err := parseCSSLength(tc.in)
		if err != nil {
			t.Errorf("parseCSSLength(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseCSSLength(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
	if _, err := parseCSSLength("2em"); err == nil {
		t.Error("parseCSSLength(\"2em\"): expected error (unit not supported)")
	}
}

func TestApplyCSSCommentsAndUnbalanced(t *testing.T) {
	th := New()
	if err := th.ApplyCSS(`/* header */ :root { /* inline */ --primary: #112233; }`); err != nil {
		t.Fatalf("ApplyCSS with comments: %v", err)
	}
	want, _ := parseHexColor("#112233")
	if th.Light.Primary != want {
		t.Errorf("Light.Primary = %+v, want %+v", th.Light.Primary, want)
	}
	if err := th.ApplyCSS(`:root { --primary: #112233;`); err == nil {
		t.Error("expected error for unbalanced braces")
	}
}

package theme

import (
	"testing"

	uitheme "github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

// swapSystemDarkHook overrides SystemDarkHook for one test.
func swapSystemDarkHook(t *testing.T, hook func() bool) {
	t.Helper()
	old := SystemDarkHook
	SystemDarkHook = hook
	t.Cleanup(func() { SystemDarkHook = old })
}

func TestActiveRespectsSetMode(t *testing.T) {
	th := New()

	th.SetMode(ModeLight)
	if th.IsDark() {
		t.Error("IsDark() = true in ModeLight")
	}
	if th.Active() != &th.Light {
		t.Error("Active() != &Light in ModeLight")
	}
	if got := th.OnSurface(); got != th.Light.Foreground {
		t.Errorf("OnSurface() = %+v, want light Foreground %+v", got, th.Light.Foreground)
	}

	th.SetMode(ModeDark)
	if !th.IsDark() {
		t.Error("IsDark() = false in ModeDark")
	}
	if th.Active() != &th.Dark {
		t.Error("Active() != &Dark in ModeDark")
	}
	if got := th.OnSurface(); got != th.Dark.Foreground {
		t.Errorf("OnSurface() = %+v, want dark Foreground %+v", got, th.Dark.Foreground)
	}
	if th.Mode() != ModeDark {
		t.Errorf("Mode() = %v, want ModeDark", th.Mode())
	}
}

func TestModeSystemUsesHook(t *testing.T) {
	th := New()
	th.SetMode(ModeSystem)

	swapSystemDarkHook(t, func() bool { return false })
	if th.IsDark() {
		t.Error("ModeSystem with light hook: IsDark() = true")
	}

	swapSystemDarkHook(t, func() bool { return true })
	if !th.IsDark() {
		t.Error("ModeSystem with dark hook: IsDark() = false")
	}
	if th.Active() != &th.Dark {
		t.Error("ModeSystem with dark hook: Active() != &Dark")
	}

	swapSystemDarkHook(t, nil)
	if th.IsDark() {
		t.Error("nil SystemDarkHook must fall back to light")
	}
}

func TestOnModeChange(t *testing.T) {
	th := New()
	th.SetMode(ModeLight)

	var got []Mode
	th.OnModeChange(func(m Mode) { got = append(got, m) })
	th.OnModeChange(nil) // must be a no-op, not a panic

	th.SetMode(ModeDark)
	th.SetMode(ModeDark) // same mode: no notification
	th.SetMode(ModeLight)

	want := []Mode{ModeDark, ModeLight}
	if len(got) != len(want) {
		t.Fatalf("callbacks fired %d times (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("callback %d got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestAsUIThemeMapping(t *testing.T) {
	th := New()
	th.SetMode(ModeLight)

	ut := th.AsUITheme()
	tok := &th.Light
	checks := []struct {
		name      string
		got, want widget.Color
	}{
		{"Background", ut.Colors.Background, tok.Background},
		{"Surface", ut.Colors.Surface, tok.Card},
		{"SurfaceVariant", ut.Colors.SurfaceVariant, tok.Muted},
		{"Primary", ut.Colors.Primary, tok.Primary},
		{"OnPrimary", ut.Colors.OnPrimary, tok.PrimaryForeground},
		{"Secondary", ut.Colors.Secondary, tok.Secondary},
		{"OnBackground", ut.Colors.OnBackground, tok.Foreground},
		{"OnSurface", ut.Colors.OnSurface, tok.Foreground},
		{"Error", ut.Colors.Error, tok.Destructive},
		{"OnError", ut.Colors.OnError, tok.DestructiveForeground},
		{"Outline", ut.Colors.Outline, tok.Border},
		{"Divider", ut.Colors.Divider, tok.Border},
		{"Shadow", ut.Colors.Shadow, widget.RGBA(0, 0, 0, 0.10)},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("Colors.%s = %+v, want %+v", c.name, c.got, c.want)
		}
	}
	if ut.Mode != uitheme.ModeLight {
		t.Errorf("Mode = %v, want ModeLight", ut.Mode)
	}
	if ut.Name != "graft-light" {
		t.Errorf("Name = %q, want graft-light", ut.Name)
	}
	if ut.Typography.FontFamily != DefaultFontSans {
		t.Errorf("Typography.FontFamily = %q, want %q", ut.Typography.FontFamily, DefaultFontSans)
	}
	if ut.Radii.M != th.RadiusMD() || ut.Radii.L != th.RadiusLG() {
		t.Errorf("Radii M/L = %v/%v, want %v/%v", ut.Radii.M, ut.Radii.L, th.RadiusMD(), th.RadiusLG())
	}

	// The dark variant maps the dark tokens.
	th.SetMode(ModeDark)
	utd := th.AsUITheme()
	if utd.Name != "graft-dark" || utd.Mode != uitheme.ModeDark {
		t.Errorf("dark theme Name/Mode = %q/%v", utd.Name, utd.Mode)
	}
	if utd.Colors.Background != th.Dark.Background {
		t.Errorf("dark Background = %+v, want %+v", utd.Colors.Background, th.Dark.Background)
	}

	// Extension interop (§2.6): tokens recoverable from the ui theme.
	if got, ok := uitheme.ExtensionAs[*Theme](ut, "graft"); !ok || got != th {
		t.Errorf("ExtensionAs = %v/%v, want the source theme", got, ok)
	}
	// Registry registration.
	if _, ok := uitheme.Get("graft-light"); !ok {
		t.Error("graft-light not registered in the gogpu/ui theme registry")
	}
	if _, ok := uitheme.Get("graft-dark"); !ok {
		t.Error("graft-dark not registered in the gogpu/ui theme registry")
	}
}

func TestRadiusScale(t *testing.T) {
	cases := []struct {
		radius                            float32
		xs, sm, md, lg, xl, xxl, x3l, x4l float32
	}{
		// Default 10px: shadcn's published 2/6/8/10/14 plus 18/22/26.
		{10, 2, 6, 8, 10, 14, 18, 22, 26},
		// Radius 0 ("sharp" themes): everything except XS collapses to 0.
		{0, 2, 0, 0, 0, 0, 0, 0, 0},
	}
	for _, tc := range cases {
		th := New(Radius(tc.radius))
		got := []struct {
			name string
			v    float32
			want float32
		}{
			{"RadiusXS", th.RadiusXS(), tc.xs},
			{"RadiusSM", th.RadiusSM(), tc.sm},
			{"RadiusMD", th.RadiusMD(), tc.md},
			{"RadiusLG", th.RadiusLG(), tc.lg},
			{"RadiusXL", th.RadiusXL(), tc.xl},
			{"Radius2XL", th.Radius2XL(), tc.xxl},
			{"Radius3XL", th.Radius3XL(), tc.x3l},
			{"Radius4XL", th.Radius4XL(), tc.x4l},
		}
		for _, g := range got {
			if g.v != g.want {
				t.Errorf("radius %v: %s() = %v, want %v", tc.radius, g.name, g.v, g.want)
			}
		}
		if th.RadiusFull() != 9999 {
			t.Errorf("RadiusFull() = %v, want 9999", th.RadiusFull())
		}
	}
}

func TestNewOptions(t *testing.T) {
	th := New(BaseColor(Zinc), Radius(6), Fonts("Inter", "JetBrains Mono"))
	if th.Radius != 6 {
		t.Errorf("Radius = %v, want 6", th.Radius)
	}
	if th.FontSans != "Inter" || th.FontMono != "JetBrains Mono" {
		t.Errorf("fonts = %q/%q", th.FontSans, th.FontMono)
	}
	zinc := preset(Zinc)
	if th.Light.Primary != zinc.light.Primary {
		t.Errorf("Light.Primary = %+v, want zinc preset %+v", th.Light.Primary, zinc.light.Primary)
	}

	// Defaults.
	def := New()
	if def.FontSans != DefaultFontSans || def.FontMono != DefaultFontMono {
		t.Errorf("default fonts = %q/%q", def.FontSans, def.FontMono)
	}
	if def.Mode() != ModeSystem {
		t.Errorf("default Mode = %v, want ModeSystem", def.Mode())
	}
}

func TestModeString(t *testing.T) {
	cases := map[Mode]string{
		ModeLight:  "light",
		ModeDark:   "dark",
		ModeSystem: "system",
		Mode(99):   "unknown",
	}
	for m, want := range cases {
		if got := m.String(); got != want {
			t.Errorf("Mode(%d).String() = %q, want %q", m, got, want)
		}
	}
}

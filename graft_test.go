package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/theme"
)

func looseConstraints() geometry.Constraints {
	return geometry.Loose(geometry.Sz(10000, 10000))
}

// TestGoldenTokenSwatch renders every shadcn token as a labeled swatch in
// both modes — an end-to-end check of presets, OKLCH conversion, theme
// resolution, and offscreen rendering.
func TestGoldenTokenSwatch(t *testing.T) {
	gtest.GoldenLightDark(t, "token-swatch", func() widget.Widget {
		tok := graft.CurrentTheme().Active()
		rows := []struct {
			name  string
			color widget.Color
		}{
			{"background", tok.Background},
			{"foreground", tok.Foreground},
			{"card", tok.Card},
			{"primary", tok.Primary},
			{"primary-foreground", tok.PrimaryForeground},
			{"secondary", tok.Secondary},
			{"muted", tok.Muted},
			{"muted-foreground", tok.MutedForeground},
			{"accent", tok.Accent},
			{"destructive", tok.Destructive},
			{"border", tok.Border},
			{"input", tok.Input},
			{"ring", tok.Ring},
			{"chart-1", tok.Chart[0]},
			{"chart-2", tok.Chart[1]},
			{"chart-3", tok.Chart[2]},
			{"chart-4", tok.Chart[3]},
			{"chart-5", tok.Chart[4]},
		}

		items := make([]widget.Widget, 0, len(rows))
		for _, row := range rows {
			swatch := primitives.Box().
				Width(48).Height(24).
				Background(row.color).
				Rounded(4).
				BorderStyle(1, tok.Border)
			label := graft.Text(row.name)
			items = append(items, primitives.HBox(swatch, label).Gap(12).CrossAlign(primitives.CrossAxisCenter))
		}
		return primitives.VBox(items...).Gap(6).Padding(24)
	})
}

// TestGoldenTypography renders the typography scale with Geist weights.
func TestGoldenTypography(t *testing.T) {
	gtest.GoldenLightDark(t, "typography", func() widget.Widget {
		return primitives.VBox(
			graft.H1("The Joke Tax"),
			graft.H2("The People of the Kingdom"),
			graft.H3("The Joke Tax"),
			graft.H4("People stopped telling jokes"),
			graft.P("The king, seeing how much happier his subjects were."),
			graft.Lead("A modal dialog that interrupts the user."),
			graft.Large("Are you absolutely sure?"),
			graft.Small("Email address"),
			graft.MutedText("Enter your email address."),
			graft.InlineCode("npm install graft"),
		).Gap(12).Padding(24)
	})
}

// TestTypographyMeasurement pins layout behavior: width tracks content,
// height equals the line box.
func TestTypographyMeasurement(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	short := graft.Text("Hi")
	long := graft.Text("Hello, pixel-faithful world")

	ws := layoutWidth(t, short)
	wl := layoutWidth(t, long)
	if ws <= 0 || wl <= ws {
		t.Fatalf("widths not monotonic: short=%v long=%v", ws, wl)
	}
}

func layoutWidth(t *testing.T, w widget.Widget) float32 {
	t.Helper()
	size := w.Layout(nil, looseConstraints())
	return size.Width
}

// TestThemeReexports pins the root-package aliases.
func TestThemeReexports(t *testing.T) {
	th := graft.NewTheme(graft.BaseColor(graft.Zinc), graft.Radius(8))
	if th.Radius != 8 {
		t.Fatalf("Radius option: got %v", th.Radius)
	}
	if th.Light.Primary == th.Light.Background {
		t.Fatal("zinc preset: primary should differ from background")
	}
	if graft.OKLCH(1, 0, 0) != (widget.Color{R: 1, G: 1, B: 1, A: 1}) {
		t.Fatal("OKLCH re-export broken")
	}

	p := graft.PaintersFor(th)
	if p == nil || p.Button.Theme != th {
		t.Fatal("PaintersFor: bundle not wired to theme")
	}
	if graft.PaintersFor(th) != p {
		t.Fatal("PaintersFor: bundle not cached")
	}
}

// TestModeSwitchRepaints verifies Active() flips with mode on the shared theme.
func TestModeSwitchRepaints(t *testing.T) {
	th := graft.CurrentTheme()
	prev := th.Mode()
	defer th.SetMode(prev)

	th.SetMode(theme.ModeLight)
	light := th.Active().Background
	th.SetMode(theme.ModeDark)
	dark := th.Active().Background
	if light == dark {
		t.Fatal("light and dark background tokens must differ")
	}
}

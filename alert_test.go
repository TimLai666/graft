package graft_test

import (
	"testing"

	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// alertForceLight pins the shared theme to light mode for spec-test color
// assertions and restores the previous mode on cleanup. It returns the active
// light token set.
func alertForceLight(t *testing.T) *graft.Theme {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(graft.ModeLight)
	t.Cleanup(func() { th.SetMode(prev) })
	return th
}

// TestAlertSpecChrome verifies the alert paints a Card-colored rounded-lg
// panel with a 1px inside border in the Border token.
func TestAlertSpecChrome(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()

	a := graft.Alert(
		graft.AlertTitle("Heads up!"),
		graft.AlertDescription("You can add components to your app."),
	)
	a.Layout(nil, fixedWidthLoose(360))
	canvas := uitest.DrawWidget(a)

	// Chrome is now BorderFill (two convex fills): an outer round-rect in
	// --border at full bounds, then an inner Card-colored round-rect inset by
	// the 1px border. No StrokeRoundRect for the border anymore.
	if len(canvas.StrokeRoundRects) != 0 {
		t.Fatalf("alert drew %d strokes, want 0 (border is now BorderFill)", len(canvas.StrokeRoundRects))
	}
	if len(canvas.RoundRects) < 2 {
		t.Fatalf("alert drew %d round-rects, want >= 2 (border + card)", len(canvas.RoundRects))
	}

	border := canvas.RoundRects[0]
	if border.Color != tok.Border {
		t.Errorf("alert border color = %v, want Border %v", border.Color, tok.Border)
	}
	if border.Bounds != a.Bounds() {
		t.Errorf("alert border bounds = %v, want full bounds %v", border.Bounds, a.Bounds())
	}
	if border.Radius != th.RadiusLG() {
		t.Errorf("alert border radius = %v, want RadiusLG %v", border.Radius, th.RadiusLG())
	}

	bg := canvas.RoundRects[1]
	if bg.Color != tok.Card {
		t.Errorf("alert bg color = %v, want Card %v", bg.Color, tok.Card)
	}
	if want := a.Bounds().Expand(-1); bg.Bounds != want {
		t.Errorf("alert bg bounds = %v, want inset %v", bg.Bounds, want)
	}
	if want := th.RadiusLG() - 1; bg.Radius != want {
		t.Errorf("alert bg radius = %v, want RadiusLG-1 %v", bg.Radius, want)
	}
}

// TestAlertSpecHeight pins the alert height: py-2 padding (8px top+bottom =
// 16px) plus the stacked title (20px) + row gap (2px) + description (20px) =
// 58px for the two-row default.
func TestAlertSpecHeight(t *testing.T) {
	alertForceLight(t)
	a := graft.Alert(
		graft.AlertTitle("Heads up!"),
		graft.AlertDescription("You can add components to your app."),
	)
	size := a.Layout(nil, fixedWidthLoose(360))
	want := 2*metrics.Alert.PadY + metrics.Alert.TitleLineHeight + metrics.Alert.RowGap + metrics.Alert.DescLineHeight
	if size.Height != want {
		t.Errorf("alert height = %v, want %v", size.Height, want)
	}
	if size.Width != 360 {
		t.Errorf("alert width = %v, want 360 (w-full)", size.Width)
	}
}

// TestAlertSpecDestructiveColors verifies the destructive variant colors the
// title with the Destructive token and the description with Destructive@90%.
func TestAlertSpecDestructiveColors(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()

	a := graft.Alert(
		graft.AlertTitle("Error"),
		graft.AlertDescription("Something went wrong."),
	).Destructive()
	a.Layout(nil, fixedWidthLoose(360))
	canvas := uitest.DrawWidget(a)

	// TypographyWidget draws through the StyledTextDrawer path on the mock.
	if len(canvas.StyledTexts) < 2 {
		t.Fatalf("alert drew %d styled-text runs, want >= 2", len(canvas.StyledTexts))
	}
	title := canvas.StyledTexts[0]
	if title.Style.Color != tok.Destructive {
		t.Errorf("destructive title color = %v, want Destructive %v", title.Style.Color, tok.Destructive)
	}
	desc := canvas.StyledTexts[1]
	wantA := tok.Destructive
	wantA.A = 0.9
	if desc.Style.Color != wantA {
		t.Errorf("destructive description color = %v, want Destructive@90%% %v", desc.Style.Color, wantA)
	}
}

// TestAlertSpecIconColumn verifies the content column shifts right by the icon
// column (16px) + gap (12px) when an icon is set.
func TestAlertSpecIconColumn(t *testing.T) {
	alertForceLight(t)

	noIcon := graft.Alert(graft.AlertTitle("Title"))
	noIcon.Layout(nil, fixedWidthLoose(360))
	cNo := uitest.DrawWidget(noIcon)

	withIcon := graft.Alert(graft.AlertTitle("Title")).Icon(icons.Info)
	withIcon.Layout(nil, fixedWidthLoose(360))
	cYes := uitest.DrawWidget(withIcon)

	if len(cNo.StyledTexts) == 0 || len(cYes.StyledTexts) == 0 {
		t.Fatal("expected title text in both alerts")
	}
	// Text bounds are in alert-local coordinates (drawn under PushTransform).
	noX := cNo.StyledTexts[0].Bounds.Min.X
	yesX := cYes.StyledTexts[0].Bounds.Min.X
	if got := yesX - noX; got != 16+12 {
		t.Errorf("icon shifts content by %v, want 28 (16 column + 12 gap)", got)
	}
}

// TestGoldenAlert renders the alert variants in light and dark modes.
func TestGoldenAlert(t *testing.T) {
	build := func(icon bool, destructive bool) func() widget.Widget {
		return func() widget.Widget {
			a := graft.Alert(
				graft.AlertTitle("Heads up!"),
				graft.AlertDescription("You can add components to your app using the cli."),
			)
			if destructive {
				a = graft.Alert(
					graft.AlertTitle("Unable to process your payment."),
					graft.AlertDescription("Please verify your billing information."),
				).Destructive()
			}
			if icon {
				if destructive {
					a.Icon(icons.CircleAlert)
				} else {
					a.Icon(icons.Info)
				}
			}
			// Constrain width to a fixed 400px panel for a stable golden.
			return widthBox(a, 400)
		}
	}

	gtest.GoldenLightDark(t, "alert-default-icon", build(true, false))
	gtest.GoldenLightDark(t, "alert-destructive-icon", build(true, true))
	gtest.GoldenLightDark(t, "alert-no-icon", build(false, false))
}

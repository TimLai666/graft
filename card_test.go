package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

func TestCardSpec(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)
	tok := th.Active()

	card := graft.Card(graft.CardContent(graft.Text("Hello"))).W(360)
	size := uitest.LayoutWidget(card, 800, 600)
	if size.Width != 360 {
		t.Errorf("card width: got %v, want 360", size.Width)
	}
	// h = border 2*1 + py 2*24 + one 20px text line = 70.
	if size.Height != 70 {
		t.Errorf("card height: got %v, want 70", size.Height)
	}

	cv := uitest.DrawWidget(card)
	// shadow-sm = 3 layers + 1 card fill.
	if len(cv.RoundRects) != len(metrics.ShadowSM)+1 {
		t.Fatalf("round rects: got %d, want %d", len(cv.RoundRects), len(metrics.ShadowSM)+1)
	}
	fill := cv.RoundRects[len(metrics.ShadowSM)]
	if fill.Color != tok.Card {
		t.Errorf("card fill: got %v, want card token %v", fill.Color, tok.Card)
	}
	if fill.Radius != th.RadiusXL() {
		t.Errorf("card radius: got %v, want rounded-xl %v", fill.Radius, th.RadiusXL())
	}
	if fill.Bounds != geometry.NewRect(0, 0, 360, 70) {
		t.Errorf("card fill bounds: got %v", fill.Bounds)
	}

	if len(cv.StrokeRoundRects) != 1 {
		t.Fatalf("strokes: got %d, want 1 (border)", len(cv.StrokeRoundRects))
	}
	border := cv.StrokeRoundRects[0]
	if border.Color != tok.Border || border.StrokeWidth != 1 {
		t.Errorf("border: got %+v, want 1px border token", border)
	}
	if border.Bounds != geometry.NewRect(0, 0, 360, 70).Expand(-0.5) {
		t.Errorf("border bounds: got %v (inside border expected)", border.Bounds)
	}
	if border.Radius != th.RadiusXL()-0.5 {
		t.Errorf("border radius: got %v, want %v", border.Radius, th.RadiusXL()-0.5)
	}
}

func TestCardTitleTypography(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	card := graft.Card(graft.CardHeader(
		graft.CardTitle("Title"),
		graft.CardDescription("Description"),
	)).W(360)
	uitest.LayoutWidget(card, 800, 600)
	cv := uitest.DrawWidget(card)

	if len(cv.StyledTexts) != 2 {
		t.Fatalf("texts: got %d, want 2", len(cv.StyledTexts))
	}
	title, desc := cv.StyledTexts[0], cv.StyledTexts[1]
	if title.Style.FontSize != 16 || title.Style.FontFamily != fonts.Family(600) {
		t.Errorf("title: got %v/%q, want 16px %q (font-semibold)",
			title.Style.FontSize, title.Style.FontFamily, fonts.Family(600))
	}
	if title.Bounds.Height() != 16 {
		t.Errorf("title line box: got %v, want 16 (leading-none)", title.Bounds.Height())
	}
	if desc.Style.FontSize != 14 || desc.Style.FontFamily != fonts.Family(400) {
		t.Errorf("description: got %v/%q, want 14px regular", desc.Style.FontSize, desc.Style.FontFamily)
	}
	if desc.Style.Color != graft.CurrentTheme().Active().MutedForeground {
		t.Errorf("description color: got %v, want muted-foreground", desc.Style.Color)
	}
	// gap-2 between title (16px line) and description.
	if got := desc.Bounds.Min.Y - title.Bounds.Min.Y; got != 16+metrics.CardHeaderGap {
		t.Errorf("title/description gap: got %v, want %v", got, 16+metrics.CardHeaderGap)
	}
}

func TestCardActionTopRight(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	badge := graft.Badge("New")
	action := graft.CardAction(badge)
	header := graft.CardHeader(
		graft.CardTitle("Title"),
		graft.CardDescription("Desc"),
		action,
	)
	card := graft.Card(header).W(360)
	uitest.LayoutWidget(card, 800, 600)

	// Header section stretches to the card inner width (360 - 2px border).
	if got := header.Bounds().Width(); got != 358 {
		t.Errorf("header width: got %v, want 358", got)
	}
	// Action sits top-right inside the header (justify-self-end,
	// self-start): right edge at header width - px 24, top edge at 0.
	if got := action.Bounds().Max.X; got != 358-metrics.CardSectionPadX {
		t.Errorf("action right edge: got %v, want %v", got, 358-metrics.CardSectionPadX)
	}
	if got := action.Bounds().Min.Y; got != 0 {
		t.Errorf("action top edge: got %v, want 0", got)
	}
	// The badge fills the action cell.
	if badge.Bounds().Min != geometry.Pt(0, 0) {
		t.Errorf("badge position in action: got %v, want (0,0)", badge.Bounds().Min)
	}
}

func TestGoldenCard(t *testing.T) {
	gtest.GoldenLightDark(t, "card-basic", func() widget.Widget {
		return pad16(graft.Card(
			graft.CardHeader(
				graft.CardTitle("Login to your account"),
				graft.CardDescription("Enter your email below to login"),
			),
			graft.CardContent(
				graft.Text("Card content goes here."),
			),
			graft.CardFooter(
				graft.MutedText("Don't have an account?"),
			),
		).W(360))
	})

	gtest.GoldenLightDark(t, "card-action", func() widget.Widget {
		return pad16(graft.Card(
			graft.CardHeader(
				graft.CardTitle("Notifications"),
				graft.CardDescription("You have 3 unread messages"),
				graft.CardAction(graft.Badge("3 new")),
			),
			graft.CardContent(
				graft.Text("Review them in your inbox."),
			),
		).W(360))
	})
}

// pad16 wraps a golden subject with breathing room so outside-painted
// chrome (shadows, focus rings) is visible in the capture.
func pad16(w widget.Widget) widget.Widget {
	return primitives.Box(w).Padding(16)
}

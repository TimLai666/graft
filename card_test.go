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
	// h = border 2*1 + py 2*16 + one 20px text line = 54.
	wantH := 2*metrics.CardBorderWidth + 2*metrics.CardPadY + 20
	if size.Height != wantH {
		t.Errorf("card height: got %v, want %v", size.Height, wantH)
	}

	cv := uitest.DrawWidget(card)
	// shadow-sm = 3 layers + BorderFill (outer border round-rect + inner
	// card fill round-rect).
	if len(cv.RoundRects) != len(metrics.ShadowSM)+2 {
		t.Fatalf("round rects: got %d, want %d", len(cv.RoundRects), len(metrics.ShadowSM)+2)
	}

	// Outer round-rect = border color at full bounds / full radius.
	border := cv.RoundRects[len(metrics.ShadowSM)]
	if border.Color != tok.Border {
		t.Errorf("border fill: got %v, want border token %v", border.Color, tok.Border)
	}
	if border.Radius != th.RadiusXL() {
		t.Errorf("border radius: got %v, want rounded-xl %v", border.Radius, th.RadiusXL())
	}
	if border.Bounds != geometry.NewRect(0, 0, 360, wantH) {
		t.Errorf("border fill bounds: got %v", border.Bounds)
	}

	// Inner round-rect = card fill, inset by 1px, radius clamped by 1.
	fill := cv.RoundRects[len(metrics.ShadowSM)+1]
	if fill.Color != tok.Card {
		t.Errorf("card fill: got %v, want card token %v", fill.Color, tok.Card)
	}
	if fill.Radius != th.RadiusXL()-1 {
		t.Errorf("card radius: got %v, want rounded-xl-1 %v", fill.Radius, th.RadiusXL()-1)
	}
	if fill.Bounds != geometry.NewRect(0, 0, 360, wantH).Expand(-1) {
		t.Errorf("card fill bounds: got %v", fill.Bounds)
	}

	if len(cv.StrokeRoundRects) != 0 {
		t.Fatalf("strokes: got %d, want 0 (border now a fill)", len(cv.StrokeRoundRects))
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
	if title.Style.FontSize != metrics.CardTitleFontSize || title.Style.FontFamily != fonts.Family(metrics.CardTitleFontWeight) {
		t.Errorf("title: got %v/%q, want %vpx %q (font-medium)",
			title.Style.FontSize, title.Style.FontFamily, metrics.CardTitleFontSize, fonts.Family(metrics.CardTitleFontWeight))
	}
	if title.Bounds.Height() != metrics.CardTitleLineHeight {
		t.Errorf("title line box: got %v, want %v (leading-snug)", title.Bounds.Height(), metrics.CardTitleLineHeight)
	}
	if desc.Style.FontSize != 14 || desc.Style.FontFamily != fonts.Family(400) {
		t.Errorf("description: got %v/%q, want 14px regular", desc.Style.FontSize, desc.Style.FontFamily)
	}
	if desc.Style.Color != graft.CurrentTheme().Active().MutedForeground {
		t.Errorf("description color: got %v, want muted-foreground", desc.Style.Color)
	}
	// gap-1 between title (22px leading-snug line) and description.
	if got := desc.Bounds.Min.Y - title.Bounds.Min.Y; got != metrics.CardTitleLineHeight+metrics.CardHeaderGap {
		t.Errorf("title/description gap: got %v, want %v", got, metrics.CardTitleLineHeight+metrics.CardHeaderGap)
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

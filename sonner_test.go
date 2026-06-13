package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// TestToastCardGeometry pins the Sonner toast card: width 356, padding 16,
// rounded-lg, 1px border, bg-popover, shadow-lg, and the title/description
// line layout.
func TestToastCardGeometry(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	card := graft.ToastCard("Event has been created",
		graft.ToastDescription("Sunday, December 03, 2023 at 9:00 AM"))
	size := uitest.LayoutWidget(card, 900, 900)

	if size.Width != metrics.Sonner.Width {
		t.Errorf("width = %v, want %v", size.Width, metrics.Sonner.Width)
	}
	// p-16 + title(20) + gap(4) + desc(18) + p-16.
	wantH := metrics.Sonner.Padding*2 + metrics.Sonner.TitleLineHeight +
		metrics.Sonner.TextGap + metrics.Sonner.DescriptionLineHeight
	if size.Height != wantH {
		t.Errorf("height = %v, want %v", size.Height, wantH)
	}

	mc := uitest.DrawWidget(card)
	th := graft.CurrentTheme()
	tok := th.Active()

	// bg-popover fill at rounded-lg, full bounds.
	bg := -1
	for i, c := range mc.RoundRects {
		if c.Color == tok.Popover {
			bg = i
			break
		}
	}
	if bg < 0 {
		t.Fatalf("no bg-popover fill: %+v", mc.RoundRects)
	}
	if got := mc.RoundRects[bg]; got.Radius != th.RadiusLG() ||
		got.Bounds != geometry.NewRect(0, 0, size.Width, size.Height) {
		t.Errorf("bg fill = %+v, want radius %v full bounds", got, th.RadiusLG())
	}
	// shadow-lg layers paint before the fill.
	if bg != len(metrics.ShadowLG) {
		t.Errorf("fills before bg = %d, want %d shadow layers", bg, len(metrics.ShadowLG))
	}

	// 1px inside border in the border token.
	if len(mc.StrokeRoundRects) != 1 {
		t.Fatalf("strokes = %d, want 1 border", len(mc.StrokeRoundRects))
	}
	if st := mc.StrokeRoundRects[0]; st.Color != tok.Border || st.StrokeWidth != metrics.Sonner.BorderWidth {
		t.Errorf("border = %+v, want 1px border token", st)
	}
}

// TestToastCardStatusIconShiftsText verifies a variant icon pushes the
// text column right by icon + gap.
func TestToastCardStatusIconShiftsText(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	plain := graft.ToastCard("Saved")
	withIcon := graft.ToastCard("Saved", graft.ToastSuccessOpt())

	uitest.LayoutWidget(plain, 900, 900)
	uitest.LayoutWidget(withIcon, 900, 900)

	mcPlain := uitest.DrawWidget(plain)
	mcIcon := uitest.DrawWidget(withIcon)

	plainX := titleTextX(t, mcPlain)
	iconX := titleTextX(t, mcIcon)
	if want := metrics.Sonner.IconSize + metrics.Sonner.IconGap; iconX-plainX != want {
		t.Errorf("status-icon text shift = %v, want %v", iconX-plainX, want)
	}
}

// titleTextX returns the x origin of the first drawn title text.
func titleTextX(t *testing.T, mc *uitest.MockCanvas) float32 {
	t.Helper()
	if len(mc.StyledTexts) > 0 {
		return mc.StyledTexts[0].Bounds.Min.X
	}
	if len(mc.Texts) > 0 {
		return mc.Texts[0].Bounds.Min.X
	}
	t.Fatal("no title text drawn")
	return 0
}

// TestToastQueueDrains verifies the imperative Toast API feeds a mounted
// Toaster, which lays cards out bottom-right.
func TestToastQueueDrains(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	graft.Toast("First")
	graft.Toast("Second", graft.ToastSuccessOpt())

	region := graft.Toaster()
	ctx := uitest.NewMockContext()
	ctx.WindowSizeVal = geometry.Sz(800, 600)

	uitest.LayoutWidget(region, 800, 600)
	uitest.DrawWidgetWithContext(region, ctx)

	cards := region.Children()
	if len(cards) != 2 {
		t.Fatalf("toaster drained %d cards, want 2", len(cards))
	}
	// Newest (Second) renders nearest the corner: lowest on screen.
	type bounder interface{ Bounds() geometry.Rect }
	b0 := cards[0].(bounder).Bounds()
	b1 := cards[1].(bounder).Bounds()
	if b0.Min.Y <= b1.Min.Y {
		t.Errorf("newest card not nearest corner: y0=%v y1=%v", b0.Min.Y, b1.Min.Y)
	}
	// Right-anchored with viewport offset.
	wantRight := float32(800) - metrics.Sonner.ViewportOffset
	if b0.Max.X != wantRight {
		t.Errorf("card right edge = %v, want %v", b0.Max.X, wantRight)
	}
}

func TestGoldenToast(t *testing.T) {
	gtest.GoldenLightDark(t, "toast-card", func() widget.Widget {
		card := graft.ToastCard("Event has been created",
			graft.ToastDescription("Sunday, December 03, 2023 at 9:00 AM"))
		return primitives.Box(card).Padding(32)
	})

	gtest.GoldenLightDark(t, "toast-stack", func() widget.Widget {
		stack := graft.ToastStack(
			graft.ToastCard("Your message was sent", graft.ToastSuccessOpt()),
			graft.ToastCard("Could not save changes", graft.ToastErrorOpt(),
				graft.ToastDescription("There was a problem with your request.")),
			graft.ToastCard("Scheduled maintenance tonight", graft.ToastWarningOpt()),
		)
		return primitives.Box(stack).Padding(32)
	})
}

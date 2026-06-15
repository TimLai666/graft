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

	// BorderFill: outer border round-rect (full bounds) then inner bg-popover
	// fill (inset by the 1px border). shadow-lg paints before both.
	bw := metrics.Sonner.BorderWidth

	border := mc.RoundRects[len(metrics.ShadowLG)]
	if border.Color != tok.Border {
		t.Errorf("border fill = %+v, want border token %v", border, tok.Border)
	}
	if border.Radius != th.RadiusLG() ||
		border.Bounds != geometry.NewRect(0, 0, size.Width, size.Height) {
		t.Errorf("border fill = %+v, want radius %v full bounds", border, th.RadiusLG())
	}

	bg := mc.RoundRects[len(metrics.ShadowLG)+1]
	if bg.Color != tok.Popover {
		t.Errorf("bg fill = %+v, want bg-popover %v", bg, tok.Popover)
	}
	if bg.Radius != th.RadiusLG()-bw ||
		bg.Bounds != geometry.NewRect(bw, bw, size.Width-2*bw, size.Height-2*bw) {
		t.Errorf("bg fill = %+v, want radius %v inset by %v", bg, th.RadiusLG()-bw, bw)
	}

	// No border stroke any more (border is now a fill).
	if len(mc.StrokeRoundRects) != 0 {
		t.Fatalf("strokes = %d, want 0 (border now a fill)", len(mc.StrokeRoundRects))
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

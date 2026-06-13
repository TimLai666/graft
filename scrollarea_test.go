package graft_test

import (
	"math"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

func approxEq(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.01
}

// buildScrollArea returns a 180x200 scroll area over 600px-tall content.
func buildScrollArea() *graft.ScrollAreaWidget {
	content := primitives.Box().Height(600)
	return graft.ScrollArea(content).W(180).H(200)
}

// findPillRects returns every round-rect drawn with the scroll thumb's
// rounded-full radius.
func findPillRects(mc *uitest.MockCanvas) []uitest.DrawRoundRectCall {
	var out []uitest.DrawRoundRectCall
	for _, c := range mc.RoundRects {
		if c.Radius == metrics.ScrollArea.ThumbRadius {
			out = append(out, c)
		}
	}
	return out
}

// TestScrollAreaThumbGeometry pins the thumb rect and color: an 8px-wide
// pill in the border token, inset in the right gutter (shadcn scroll-area
// thumb "rounded-full bg-border", metrics.ScrollArea).
func TestScrollAreaThumbGeometry(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	sa := buildScrollArea()
	size := uitest.LayoutWidget(sa, 800, 600)
	if size.Width != 180 || size.Height != 200 {
		t.Fatalf("scroll area size = %v, want 180x200", size)
	}

	mc := uitest.DrawWidget(sa)
	pills := findPillRects(mc)
	if len(pills) != 1 {
		t.Fatalf("thumb round-rects = %d, want 1 (track must stay transparent)", len(pills))
	}
	thumb := pills[0]

	tok := graft.CurrentTheme().Active()
	if thumb.Color != tok.Border {
		t.Errorf("thumb color = %v, want border token %v", thumb.Color, tok.Border)
	}

	// core/scrollview geometry: 12px gutter, 2px padding, 8px thumb.
	// Thumb thickness must equal shadcn's 8px (metrics.ScrollArea.ThumbWidth).
	if got := thumb.Bounds.Width(); !approxEq(got, metrics.ScrollArea.ThumbWidth) {
		t.Errorf("thumb width = %v, want %v", got, metrics.ScrollArea.ThumbWidth)
	}
	if got := thumb.Bounds.Min.X; !approxEq(got, 170) {
		t.Errorf("thumb x = %v, want 170", got)
	}
	if got := thumb.Bounds.Min.Y; !approxEq(got, 2) {
		t.Errorf("thumb y = %v, want 2 (top, unscrolled)", got)
	}
	// Proportional thumb: viewport/content * trackLen = 200/600*196.
	if got := thumb.Bounds.Height(); !approxEq(got, 200.0/600.0*196.0) {
		t.Errorf("thumb height = %v, want %v", got, 200.0/600.0*196.0)
	}

	// No other fills: the shadcn track is transparent and the viewport
	// itself paints no background.
	for _, c := range mc.RoundRects {
		if c.Radius != metrics.ScrollArea.ThumbRadius {
			t.Errorf("unexpected round-rect %+v", c)
		}
	}
}

// TestScrollAreaHoverKeepsColors verifies hover does not restyle the thumb
// (shadcn has no hover styles on scroll-area).
func TestScrollAreaHoverKeepsColors(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	sa := buildScrollArea()
	uitest.LayoutWidget(sa, 800, 600)

	before := findPillRects(uitest.DrawWidget(sa))
	ctx := uitest.NewMockContext()
	sa.Event(ctx, uitest.MouseEnter(90, 100))
	after := findPillRects(uitest.DrawWidget(sa))

	if len(before) != 1 || len(after) != 1 {
		t.Fatalf("thumb counts = %d/%d, want 1/1", len(before), len(after))
	}
	if before[0].Color != after[0].Color {
		t.Errorf("hover changed thumb color: %v -> %v", before[0].Color, after[0].Color)
	}
}

// TestScrollAreaWheelMovesThumb verifies wheel scrolling moves the thumb
// proportionally.
func TestScrollAreaWheelMovesThumb(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	sa := buildScrollArea()
	uitest.LayoutWidget(sa, 800, 600)

	ctx := uitest.NewMockContext()
	if !sa.Event(ctx, uitest.WheelScroll(90, 100, 1)) {
		t.Fatal("wheel event not consumed")
	}
	if _, y := sa.ScrollView().ScrollOffset(); y != 40 {
		t.Fatalf("scrollY = %v, want 40 (one default wheel step)", y)
	}

	pills := findPillRects(uitest.DrawWidget(sa))
	if len(pills) != 1 {
		t.Fatalf("thumb count = %d, want 1", len(pills))
	}
	// thumbPos = scrollY/maxScroll * (trackLen - thumbLen).
	trackLen := float32(196.0)
	thumbLen := float32(200.0 / 600.0 * 196.0)
	want := 2 + 40.0/400.0*(trackLen-thumbLen)
	if got := pills[0].Bounds.Min.Y; !approxEq(got, want) {
		t.Errorf("scrolled thumb y = %v, want %v", got, want)
	}
}

// TestScrollAreaClipsContent verifies the viewport clips its content.
func TestScrollAreaClipsContent(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	sa := buildScrollArea()
	uitest.LayoutWidget(sa, 800, 600)
	mc := uitest.DrawWidget(sa)
	found := false
	for _, c := range mc.Clips {
		if c == geometry.NewRect(0, 0, 180, 200) {
			found = true
		}
	}
	if !found {
		t.Errorf("viewport clip rect not found in %v", mc.Clips)
	}
}

func scrollAreaGoldenDemo() *graft.ScrollAreaWidget {
	versions := []string{
		"v1.2.0-beta.50", "v1.2.0-beta.48", "v1.2.0-beta.47",
		"v1.2.0-beta.46", "v1.2.0-beta.45", "v1.2.0-beta.44",
		"v1.2.0-beta.43", "v1.2.0-beta.41", "v1.2.0-beta.40",
		"v1.2.0-beta.39",
	}
	rows := make([]widget.Widget, 0, 2*len(versions))
	rows = append(rows, graft.Small("Tags"))
	for _, v := range versions {
		rows = append(rows, graft.Text(v))
	}
	content := primitives.VBox(rows...).Gap(8).Padding(16)
	return graft.ScrollArea(content).W(200).H(160)
}

func TestGoldenScrollArea(t *testing.T) {
	gtest.GoldenLightDark(t, "scrollarea-default", func() widget.Widget {
		return primitives.Box(scrollAreaGoldenDemo()).Padding(24)
	})

	gtest.GoldenLightDark(t, "scrollarea-scrolled", func() widget.Widget {
		sa := scrollAreaGoldenDemo()
		uitest.LayoutWidget(sa, 200, 160)
		sa.Event(uitest.NewMockContext(), uitest.WheelScroll(100, 80, 2))
		return primitives.Box(sa).Padding(24)
	})
}

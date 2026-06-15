package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// TestItemOutlineSpec checks the outline variant's outer size, the 1px inside
// border in the Border token, rounded-md corners, and that no Item fill is
// drawn.
func TestItemOutlineSpec(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)
	tok := th.Active()

	item := graft.Item(
		graft.ItemMedia(icons.Circle).Icon(),
		graft.ItemContent(
			graft.ItemTitle("Notifications"),
			graft.ItemDescription("Configure how you receive notifications."),
		),
		graft.ItemActions(graft.Switch()),
	).Outline().W(400)

	size := uitest.LayoutWidget(item, 800, 600)
	if size.Width != 400 {
		t.Errorf("item width: got %v, want 400", size.Width)
	}
	// h = border 2*1 + p 2*16 + content (title 20 + gap 4 + desc 20 = 44) = 78.
	wantH := 2*metrics.ItemBorderWidth + 2*metrics.ItemPadDefault +
		2*metrics.ItemTitleLineHeight + metrics.ItemContentGap
	if size.Height != wantH {
		t.Errorf("item height: got %v, want %v", size.Height, wantH)
	}

	cv := uitest.DrawWidget(item)

	bw := metrics.ItemBorderWidth
	full := geometry.NewRect(0, 0, 400, wantH)

	// The outline variant's border is now a BorderFill: the full-bounds
	// round-rect is the Border-colored outer ring (no stroke).
	var border *uitest.DrawRoundRectCall
	for n := range cv.RoundRects {
		if cv.RoundRects[n].Bounds == full {
			border = &cv.RoundRects[n]
		}
	}
	if border == nil {
		t.Fatalf("no full-bounds outline border round-rect %v found", full)
	}
	if border.Color != tok.Border {
		t.Errorf("border fill: got %+v, want Border token %v", border.Color, tok.Border)
	}
	if border.Radius != th.RadiusMD() {
		t.Errorf("border radius: got %v, want rounded-md %v", border.Radius, th.RadiusMD())
	}

	// Inner fill = page Background, inset by the border width.
	inside := full.Expand(-bw)
	var inner *uitest.DrawRoundRectCall
	for n := range cv.RoundRects {
		if cv.RoundRects[n].Bounds == inside && cv.RoundRects[n].Color == tok.Background {
			inner = &cv.RoundRects[n]
		}
	}
	if inner == nil {
		t.Fatalf("no inset Background fill round-rect %v found", inside)
	}
	if inner.Radius != th.RadiusMD()-bw {
		t.Errorf("inner fill radius: got %v, want %v", inner.Radius, th.RadiusMD()-bw)
	}

	// No 1px border stroke any more (border is now a fill).
	for _, s := range cv.StrokeRoundRects {
		if s.StrokeWidth == metrics.ItemBorderWidth {
			t.Fatalf("unexpected 1px border stroke: %+v", s)
		}
	}
}

// TestItemMutedSpec checks the muted variant fills the full bounds with the
// Muted token at 50% alpha, rounded-md, and draws no border.
func TestItemMutedSpec(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)
	tok := th.Active()

	item := graft.Item(graft.ItemContent(graft.ItemTitle("Muted row"))).Muted().W(300)
	size := uitest.LayoutWidget(item, 800, 600)
	// h = border 2 + p 32 + title 20 = 54.
	wantH := 2*metrics.ItemBorderWidth + 2*metrics.ItemPadDefault + metrics.ItemTitleLineHeight
	if size != geometry.Sz(300, wantH) {
		t.Fatalf("muted item size: got %v, want 300x%v", size, wantH)
	}

	cv := uitest.DrawWidget(item)
	if len(cv.RoundRects) != 1 {
		t.Fatalf("muted round rects: got %d, want 1 (bg-muted/50)", len(cv.RoundRects))
	}
	fill := cv.RoundRects[0]
	wantColor := draw.MulAlpha(tok.Muted, metrics.ItemMutedBgAlpha)
	if fill.Color != wantColor {
		t.Errorf("muted fill: got %v, want Muted/50 %v", fill.Color, wantColor)
	}
	if fill.Radius != th.RadiusMD() {
		t.Errorf("muted radius: got %v, want rounded-md %v", fill.Radius, th.RadiusMD())
	}
	if fill.Bounds != geometry.NewRect(0, 0, 300, wantH) {
		t.Errorf("muted fill bounds: got %v", fill.Bounds)
	}
	if len(cv.StrokeRoundRects) != 0 {
		t.Errorf("muted strokes: got %d, want 0 (border-transparent)", len(cv.StrokeRoundRects))
	}
}

// TestItemDefaultNoChrome checks the default variant paints neither a fill nor
// a border (border-transparent, no background).
func TestItemDefaultNoChrome(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	item := graft.Item(graft.ItemContent(graft.ItemTitle("Plain row"))).W(300)
	uitest.LayoutWidget(item, 800, 600)
	cv := uitest.DrawWidget(item)
	if len(cv.RoundRects) != 0 {
		t.Errorf("default round rects: got %d, want 0", len(cv.RoundRects))
	}
	if len(cv.StrokeRoundRects) != 0 {
		t.Errorf("default strokes: got %d, want 0", len(cv.StrokeRoundRects))
	}
}

// TestItemSmSize checks the small size uses py-3 px-4 padding and a gap-2.5
// row gap.
func TestItemSmSize(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	item := graft.Item(graft.ItemContent(graft.ItemTitle("T"), graft.ItemDescription("D"))).Sm().W(400)
	size := uitest.LayoutWidget(item, 800, 600)
	// h = border 2 + py 2*12 + content (title 20 + gap 4 + desc 20 = 44) = 70.
	wantH := 2*metrics.ItemBorderWidth + 2*metrics.ItemPadYSm +
		2*metrics.ItemTitleLineHeight + metrics.ItemContentGap
	if size.Height != wantH {
		t.Errorf("sm item height: got %v, want %v", size.Height, wantH)
	}
}

// TestItemMediaIconChip checks the icon-media chip: a fixed 32px square with a
// Muted fill, a 1px Border inside border, and rounded-sm corners.
func TestItemMediaIconChip(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)
	tok := th.Active()

	media := graft.ItemMedia(icons.Circle).Icon()
	size := uitest.LayoutWidget(media, 200, 200)
	if size != geometry.Sz(metrics.ItemIconChipSize, metrics.ItemIconChipSize) {
		t.Fatalf("icon chip size: got %v, want %vx%v", size, metrics.ItemIconChipSize, metrics.ItemIconChipSize)
	}

	cv := uitest.DrawWidget(media)
	bw := metrics.ItemBorderWidth

	// BorderFill: outer Border round-rect at full chip bounds, inner Muted
	// fill inset by the border width. No stroke.
	if len(cv.RoundRects) != 2 {
		t.Fatalf("icon chip round rects: got %d, want 2 (Border border + Muted fill)", len(cv.RoundRects))
	}
	border := cv.RoundRects[0]
	if border.Color != tok.Border {
		t.Errorf("icon chip border fill: got %v, want Border %v", border.Color, tok.Border)
	}
	if border.Radius != th.RadiusSM() {
		t.Errorf("icon chip border radius: got %v, want rounded-sm %v", border.Radius, th.RadiusSM())
	}
	chip := cv.RoundRects[1]
	if chip.Color != tok.Muted {
		t.Errorf("icon chip fill: got %v, want Muted %v", chip.Color, tok.Muted)
	}
	if chip.Radius != th.RadiusSM()-bw {
		t.Errorf("icon chip fill radius: got %v, want %v", chip.Radius, th.RadiusSM()-bw)
	}
	if len(cv.StrokeRoundRects) != 0 {
		t.Fatalf("icon chip strokes: got %d, want 0 (border now a fill)", len(cv.StrokeRoundRects))
	}
}

// TestItemMediaImage checks the image-media variant: a fixed 40px square with
// a Muted fill and rounded-sm corners.
func TestItemMediaImage(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)

	media := graft.ItemMediaChild(graft.Text("")).Image()
	size := uitest.LayoutWidget(media, 200, 200)
	if size != geometry.Sz(metrics.ItemImageSize, metrics.ItemImageSize) {
		t.Fatalf("image media size: got %v, want %vx%v", size, metrics.ItemImageSize, metrics.ItemImageSize)
	}
	cv := uitest.DrawWidget(media)
	if len(cv.RoundRects) != 1 {
		t.Fatalf("image media round rects: got %d, want 1", len(cv.RoundRects))
	}
	if cv.RoundRects[0].Radius != th.RadiusSM() {
		t.Errorf("image media radius: got %v, want rounded-sm %v", cv.RoundRects[0].Radius, th.RadiusSM())
	}
}

// TestItemTypography checks the title and description font families, sizes,
// colors, and the gap-1 spacing between them.
func TestItemTypography(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)
	tok := th.Active()

	item := graft.Item(graft.ItemContent(
		graft.ItemTitle("Title"),
		graft.ItemDescription("Description"),
	)).W(300)
	uitest.LayoutWidget(item, 800, 600)
	cv := uitest.DrawWidget(item)

	if len(cv.StyledTexts) != 2 {
		t.Fatalf("texts: got %d, want 2", len(cv.StyledTexts))
	}
	title, desc := cv.StyledTexts[0], cv.StyledTexts[1]
	if title.Style.FontSize != metrics.ItemTitleFontSize ||
		title.Style.FontFamily != fonts.Family(metrics.ItemTitleFontWeight) {
		t.Errorf("title: got %v/%q, want %vpx %q (font-medium)",
			title.Style.FontSize, title.Style.FontFamily,
			metrics.ItemTitleFontSize, fonts.Family(metrics.ItemTitleFontWeight))
	}
	if title.Style.Color != tok.Foreground {
		t.Errorf("title color: got %v, want Foreground %v", title.Style.Color, tok.Foreground)
	}
	if title.Bounds.Height() != metrics.ItemTitleLineHeight {
		t.Errorf("title line box: got %v, want %v (leading-snug)", title.Bounds.Height(), metrics.ItemTitleLineHeight)
	}
	if desc.Style.FontSize != metrics.ItemDescriptionFontSize ||
		desc.Style.FontFamily != fonts.Family(metrics.ItemDescriptionFontWeight) {
		t.Errorf("description: got %v/%q, want %vpx regular",
			desc.Style.FontSize, desc.Style.FontFamily, metrics.ItemDescriptionFontSize)
	}
	if desc.Style.Color != tok.MutedForeground {
		t.Errorf("description color: got %v, want MutedForeground %v", desc.Style.Color, tok.MutedForeground)
	}
	// gap-1 between the title line box and the description.
	if got := desc.Bounds.Min.Y - title.Bounds.Min.Y; got != metrics.ItemTitleLineHeight+metrics.ItemContentGap {
		t.Errorf("title/description gap: got %v, want %v", got, metrics.ItemTitleLineHeight+metrics.ItemContentGap)
	}
}

// TestItemGroupSeparator checks the vertical stack height across an Item, an
// ItemSeparator (1px Border rule), and a small Item.
func TestItemGroupSeparator(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)
	tok := th.Active()

	group := graft.ItemGroup(
		graft.Item(graft.ItemContent(graft.ItemTitle("A"))),
		graft.ItemSeparator(),
		graft.Item(graft.ItemContent(graft.ItemTitle("B"))).Sm(),
	)
	size := uitest.LayoutWidget(group, 400, 600)

	// First item: border 2 + p 32 + title 20 = 54.
	firstH := 2*metrics.ItemBorderWidth + 2*metrics.ItemPadDefault + metrics.ItemTitleLineHeight
	// Small item: border 2 + py 24 + title 20 = 46.
	smH := 2*metrics.ItemBorderWidth + 2*metrics.ItemPadYSm + metrics.ItemTitleLineHeight
	wantH := firstH + metrics.SeparatorThickness + smH
	if size.Height != wantH {
		t.Errorf("group height: got %v, want %v", size.Height, wantH)
	}

	cv := uitest.DrawWidget(group)
	// The separator is the single plain Rect drawn in the Border token.
	var sep *uitest.DrawRectCall
	for n := range cv.Rects {
		if cv.Rects[n].Color == tok.Border {
			sep = &cv.Rects[n]
		}
	}
	if sep == nil {
		t.Fatal("no separator rect in Border token found")
	}
	if sep.Bounds.Height() != metrics.SeparatorThickness {
		t.Errorf("separator thickness: got %v, want %v", sep.Bounds.Height(), metrics.SeparatorThickness)
	}
}

func TestGoldenItem(t *testing.T) {
	gtest.GoldenLightDark(t, "item-variants", func() widget.Widget {
		return pad16(graft.ItemGroup(
			graft.Item(
				graft.ItemMedia(icons.Info).Icon(),
				graft.ItemContent(
					graft.ItemTitle("Notifications"),
					graft.ItemDescription("Configure how you receive notifications."),
				),
				graft.ItemActions(graft.Switch()),
			).W(420),
			graft.Item(
				graft.ItemMedia(icons.CircleCheck).Icon(),
				graft.ItemContent(
					graft.ItemTitle("Profile"),
					graft.ItemDescription("Manage your public profile and avatar."),
				),
				graft.ItemActions(graft.Badge("Pro")),
			).Outline().W(420),
			graft.Item(
				graft.ItemMedia(icons.Calendar).Icon(),
				graft.ItemContent(
					graft.ItemTitle("Preferences"),
					graft.ItemDescription("Adjust theme, language, and shortcuts."),
				),
			).Muted().W(420),
		).Gap(12))
	})

	gtest.GoldenLightDark(t, "item-group", func() widget.Widget {
		return pad16(graft.Item(
			graft.ItemContent(
				graft.ItemGroup(
					graft.Item(
						graft.ItemMedia(icons.Circle).Icon(),
						graft.ItemContent(
							graft.ItemTitle("Inbox"),
							graft.ItemDescription("Unread messages and mentions."),
						),
					),
					graft.ItemSeparator(),
					graft.Item(
						graft.ItemMedia(icons.Search).Icon(),
						graft.ItemContent(
							graft.ItemTitle("Search"),
							graft.ItemDescription("Find anything across your workspace."),
						),
					).Sm(),
				),
			),
		).Outline().W(420))
	})
}

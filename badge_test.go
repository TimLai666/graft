package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/internal/textmetrics"
)

func TestBadgeDefaultSpec(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)
	tok := th.Active()

	b := graft.Badge("New")
	size := uitest.LayoutWidget(b, 500, 500)

	// h = line 16 + py 2*2 + border 2*1 = 22 (py-0.5 + border).
	if size.Height != 22 {
		t.Errorf("badge height: got %v, want 22", size.Height)
	}
	// w = text + px 2*8 + border 2*1 (px-2 + border).
	wantW := textmetrics.Width(fonts.Family(500), 12, "New") + 18
	if size.Width != wantW {
		t.Errorf("badge width: got %v, want %v", size.Width, wantW)
	}

	cv := uitest.DrawWidget(b)
	if len(cv.RoundRects) != 1 {
		t.Fatalf("round rects: got %d, want 1 (pill fill)", len(cv.RoundRects))
	}
	pill := cv.RoundRects[0]
	if pill.Color != tok.Primary {
		t.Errorf("pill fill: got %v, want primary %v", pill.Color, tok.Primary)
	}
	if pill.Radius != th.RadiusFull() {
		t.Errorf("pill radius: got %v, want full %v", pill.Radius, th.RadiusFull())
	}
	if pill.Bounds != geometry.FromPointSize(geometry.Pt(0, 0), size) {
		t.Errorf("pill bounds: got %v", pill.Bounds)
	}
	if len(cv.StrokeRoundRects) != 0 {
		t.Errorf("default badge must not stroke a border, got %d strokes", len(cv.StrokeRoundRects))
	}
	if len(cv.StyledTexts) != 1 {
		t.Fatalf("texts: got %d, want 1", len(cv.StyledTexts))
	}
	txt := cv.StyledTexts[0]
	if txt.Style.Color != tok.PrimaryForeground {
		t.Errorf("text color: got %v, want primary-foreground %v", txt.Style.Color, tok.PrimaryForeground)
	}
	if txt.Style.FontSize != 12 {
		t.Errorf("font size: got %v, want 12 (text-xs)", txt.Style.FontSize)
	}
	if txt.Style.FontFamily != fonts.Family(500) {
		t.Errorf("font family: got %q, want %q (font-medium)", txt.Style.FontFamily, fonts.Family(500))
	}
	// Text line box sits inside padding + border: y = 3, h = 16.
	if txt.Bounds.Min.Y != 3 || txt.Bounds.Height() != 16 {
		t.Errorf("text line box: got %v, want y=3 h=16", txt.Bounds)
	}
}

func TestBadgeOutlineSpec(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)
	tok := th.Active()

	b := graft.Badge("Outline").Outline()
	size := uitest.LayoutWidget(b, 500, 500)
	cv := uitest.DrawWidget(b)

	if len(cv.RoundRects) != 0 {
		t.Errorf("outline badge has no fill, got %d round rects", len(cv.RoundRects))
	}
	if len(cv.StrokeRoundRects) != 1 {
		t.Fatalf("strokes: got %d, want 1 (inside border)", len(cv.StrokeRoundRects))
	}
	st := cv.StrokeRoundRects[0]
	if st.Color != tok.Border {
		t.Errorf("border color: got %v, want border %v", st.Color, tok.Border)
	}
	if st.StrokeWidth != 1 {
		t.Errorf("border width: got %v, want 1", st.StrokeWidth)
	}
	wantBounds := geometry.FromPointSize(geometry.Pt(0, 0), size).Expand(-0.5)
	if st.Bounds != wantBounds {
		t.Errorf("border bounds: got %v, want inset %v", st.Bounds, wantBounds)
	}
	if cv.StyledTexts[0].Style.Color != tok.Foreground {
		t.Errorf("outline text color: got %v, want foreground", cv.StyledTexts[0].Style.Color)
	}
}

func TestBadgeHoverOnlyWhenClickable(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)
	tok := th.Active()

	// Non-clickable badge: hover must NOT change the fill ([a&] semantics).
	plain := graft.Badge("Plain")
	uitest.LayoutWidget(plain, 500, 500)
	ctx := uitest.NewMockContext()
	plain.Event(ctx, uitest.MouseEnter(5, 5))
	cv := uitest.DrawWidget(plain)
	if cv.RoundRects[0].Color != tok.Primary {
		t.Errorf("non-clickable hover fill: got %v, want plain primary", cv.RoundRects[0].Color)
	}

	// Clickable badge: hover dims to primary/90.
	clicked := false
	link := graft.Badge("Link").OnClick(func() { clicked = true })
	uitest.LayoutWidget(link, 500, 500)
	link.Event(ctx, uitest.MouseEnter(5, 5))
	cv = uitest.DrawWidget(link)
	want := tok.Primary
	want.A = 0.9
	if cv.RoundRects[0].Color != want {
		t.Errorf("clickable hover fill: got %v, want primary/90 %v", cv.RoundRects[0].Color, want)
	}

	if !uitest.SimulateClick(link, 5, 5) {
		t.Fatal("click not consumed")
	}
	if !clicked {
		t.Fatal("OnClick not fired")
	}
}

func TestBadgeFocusRing(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)
	tok := th.Active()

	b := graft.Badge("Focus").OnClick(func() {})
	size := uitest.LayoutWidget(b, 500, 500)
	ctx := uitest.NewMockContext()
	b.Event(ctx, uitest.FocusGained())

	cv := uitest.DrawWidget(b)
	// Focus-visible: 3px ring band + border turning solid ring color.
	var ring, border int
	for _, st := range cv.StrokeRoundRects {
		switch st.StrokeWidth {
		case 3:
			ring++
			wantRing := tok.Ring
			wantRing.A = 0.5
			if st.Color != wantRing {
				t.Errorf("ring color: got %v, want ring/50 %v", st.Color, wantRing)
			}
			wantBounds := geometry.FromPointSize(geometry.Pt(0, 0), size).Expand(1.5)
			if st.Bounds != wantBounds {
				t.Errorf("ring bounds: got %v, want %v", st.Bounds, wantBounds)
			}
			if st.Radius != th.RadiusFull()+1.5 {
				t.Errorf("ring radius: got %v, want full+1.5", st.Radius)
			}
		case 1:
			border++
			if st.Color != tok.Ring {
				t.Errorf("focused border color: got %v, want solid ring %v", st.Color, tok.Ring)
			}
		}
	}
	if ring != 1 || border != 1 {
		t.Fatalf("focus strokes: ring=%d border=%d, want 1 and 1", ring, border)
	}

	// Click focus must not show the ring.
	b2 := graft.Badge("Click").OnClick(func() {})
	uitest.LayoutWidget(b2, 500, 500)
	uitest.SimulateClick(b2, 5, 5)
	cv = uitest.DrawWidget(b2)
	for _, st := range cv.StrokeRoundRects {
		if st.StrokeWidth == 3 {
			t.Error("pointer focus must not draw the focus-visible ring")
		}
	}
}

func TestGoldenBadge(t *testing.T) {
	hover := func(b *graft.BadgeWidget) *graft.BadgeWidget {
		uitest.LayoutWidget(b, 500, 500)
		b.Event(uitest.NewMockContext(), uitest.MouseEnter(5, 5))
		return b
	}

	gtest.GoldenLightDark(t, "badge-variants", func() widget.Widget {
		return primitives.VBox(
			primitives.HBox(
				graft.Badge("Badge"),
				graft.Badge("Secondary").Secondary(),
				graft.Badge("Destructive").Destructive(),
				graft.Badge("Outline").Outline(),
				graft.Badge("Ghost").Ghost(),
				graft.Badge("Link").Link(),
			).Gap(8),
			primitives.HBox(
				graft.Badge("Verified").Icon(icons.Check).Secondary(),
				graft.Badge("Alert").Icon(icons.CircleAlert).Destructive(),
				graft.Badge("8"),
			).Gap(8),
		).Gap(8).Padding(16)
	})

	gtest.GoldenLightDark(t, "badge-states", func() widget.Widget {
		focused := graft.Badge("Focused").OnClick(func() {})
		uitest.LayoutWidget(focused, 500, 500)
		focused.Event(uitest.NewMockContext(), uitest.FocusGained())
		return primitives.HBox(
			hover(graft.Badge("Hover").OnClick(func() {})),
			hover(graft.Badge("Hover").Secondary().OnClick(func() {})),
			hover(graft.Badge("Hover").Destructive().OnClick(func() {})),
			hover(graft.Badge("Hover").Outline().OnClick(func() {})),
			hover(graft.Badge("Hover").Link().OnClick(func() {})),
			focused,
		).Gap(12).Padding(16)
	})
}

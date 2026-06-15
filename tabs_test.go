package graft_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// buildTabs returns a standard three-trigger tabs tree for tests.
func buildTabs(line bool) (*graft.TabsWidget, []*graft.TabsTriggerWidget) {
	trs := []*graft.TabsTriggerWidget{
		graft.TabsTrigger("account", "Account"),
		graft.TabsTrigger("password", "Password"),
		graft.TabsTrigger("more", "More").Disabled(true),
	}
	t := graft.Tabs(
		graft.TabsList(trs...),
		graft.TabsContent("account", graft.Text("Account content")),
		graft.TabsContent("password", graft.Text("Password content")),
	)
	if line {
		t.Line()
	}
	return t, trs
}

func TestTabsSpecLayout(t *testing.T) {
	lightTokens(t)
	tabs, trs := buildTabs(false)
	uitest.LayoutWidget(tabs, 800, 600)

	m := metrics.Tabs
	list := tabs.Children()[0].(*graft.TabsListWidget)
	if got := list.Bounds().Height(); got != m.ListHeight {
		t.Fatalf("list height = %v, want %v", got, m.ListHeight)
	}

	// Trigger height = 36 - 2*3 - 1 = 29, vertically centered at y 3.5.
	wantH := m.ListHeight - 2*m.ListPadding - m.TriggerHeightInset
	b0 := trs[0].Bounds()
	if b0.Height() != wantH {
		t.Fatalf("trigger height = %v, want %v", b0.Height(), wantH)
	}
	if !approx(b0.Min.Y, m.ListPadding+0.5) {
		t.Fatalf("trigger y = %v, want %v", b0.Min.Y, m.ListPadding+0.5)
	}
	if !approx(b0.Min.X, m.ListPadding) {
		t.Fatalf("trigger x = %v, want %v", b0.Min.X, m.ListPadding)
	}
	// Default variant: triggers adjacent, no gap.
	if !approx(trs[1].Bounds().Min.X, b0.Max.X) {
		t.Fatalf("second trigger x = %v, want %v", trs[1].Bounds().Min.X, b0.Max.X)
	}
	// Content sits below the list with the 8px root gap.
	content := tabs.Children()[1].(*graft.TabsContentWidget)
	if !approx(content.Bounds().Min.Y, m.ListHeight+m.RootGap) {
		t.Fatalf("content y = %v, want %v", content.Bounds().Min.Y, m.ListHeight+m.RootGap)
	}
}

func TestTabsSpecDefaultVariantColors(t *testing.T) {
	tok := lightTokens(t)
	th := graft.CurrentTheme()
	tabs, trs := buildTabs(false)
	uitest.LayoutWidget(tabs, 800, 600)
	mc := uitest.DrawWidget(tabs)

	// List pill: muted, radius-lg.
	foundList := false
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Muted && approx(rr.Radius, th.RadiusLG()) &&
			approx(rr.Bounds.Height(), metrics.Tabs.ListHeight) {
			foundList = true
		}
	}
	if !foundList {
		t.Fatalf("muted list pill missing; roundrects: %+v", mc.RoundRects)
	}

	// Active trigger: bg-background at radius-md over the trigger bounds.
	foundActive := false
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Background && approx(rr.Radius, th.RadiusMD()) &&
			approx(rr.Bounds.Width(), trs[0].Bounds().Width()) {
			foundActive = true
		}
	}
	if !foundActive {
		t.Fatalf("active trigger background missing; roundrects: %+v", mc.RoundRects)
	}

	// Labels: active = foreground; idle = foreground at 60%; disabled =
	// idle faded 50%.
	idle := tok.Foreground
	idle.A *= metrics.Tabs.IdleTextOpacity
	disabled := idle
	disabled.A *= metrics.DisabledOpacity
	var gotActive, gotIdle, gotDisabled bool
	for _, st := range mc.StyledTexts {
		switch st.Text {
		case "Account":
			gotActive = st.Style.Color == tok.Foreground
			if st.Style.FontSize != metrics.Tabs.TriggerFontSize {
				t.Fatalf("trigger font size = %v", st.Style.FontSize)
			}
			if st.Style.FontFamily != "Geist Medium" {
				t.Fatalf("trigger family = %q, want Geist Medium (font-medium 500)", st.Style.FontFamily)
			}
		case "Password":
			gotIdle = st.Style.Color == idle
		case "More":
			gotDisabled = st.Style.Color == disabled
		}
	}
	if !gotActive || !gotIdle || !gotDisabled {
		t.Fatalf("label colors wrong: active=%v idle=%v disabled=%v", gotActive, gotIdle, gotDisabled)
	}

	// Only the active content is rendered.
	for _, st := range mc.StyledTexts {
		if st.Text == "Password content" {
			t.Fatal("inactive content must not draw")
		}
	}
}

func TestTabsSpecDarkActive(t *testing.T) {
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeDark)
	t.Cleanup(func() { th.SetMode(prev) })
	tok := th.Active()

	tabs, _ := buildTabs(false)
	uitest.LayoutWidget(tabs, 800, 600)
	mc := uitest.DrawWidget(tabs)

	// dark active: bg input/30 (alpha multiplied) + 1px input border.
	wantBg := tok.Input
	wantBg.A *= metrics.Tabs.DarkActiveBgOpacity
	foundBg := false
	for _, rr := range mc.RoundRects {
		if rr.Color == wantBg {
			foundBg = true
		}
	}
	if !foundBg {
		t.Fatalf("dark active trigger fill (input/30) missing; roundrects: %+v", mc.RoundRects)
	}
	// The border is now the outer BorderFill round-rect in the full Input
	// token (the inner fill is the alpha-multiplied input/30 checked above).
	foundBorder := false
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Input {
			foundBorder = true
		}
	}
	if !foundBorder {
		t.Fatalf("dark active trigger border (input) missing; roundrects: %+v", mc.RoundRects)
	}
}

func TestTabsSpecLineVariant(t *testing.T) {
	tok := lightTokens(t)
	tabs, trs := buildTabs(true)
	uitest.LayoutWidget(tabs, 800, 600)
	mc := uitest.DrawWidget(tabs)

	// No muted list pill, no active background.
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Muted || rr.Color == tok.Background {
			t.Fatalf("line variant must not fill list or trigger: %+v", rr)
		}
	}

	// Triggers separated by the 4px line gap.
	if !approx(trs[1].Bounds().Min.X, trs[0].Bounds().Max.X+metrics.Tabs.LineGap) {
		t.Fatalf("line gap: trigger2 x = %v want %v", trs[1].Bounds().Min.X, trs[0].Bounds().Max.X+metrics.Tabs.LineGap)
	}

	// Active underline: 2px foreground rect whose bottom is 5px below the
	// trigger, spanning the trigger width.
	b := trs[0].Bounds()
	found := false
	for _, r := range mc.Rects {
		if r.Color == tok.Foreground &&
			approx(r.Bounds.Height(), metrics.Tabs.UnderlineHeight) &&
			approx(r.Bounds.Width(), b.Width()) &&
			approx(r.Bounds.Max.Y, b.Max.Y+metrics.Tabs.UnderlineDrop) {
			found = true
		}
	}
	if !found {
		t.Fatalf("line underline missing; rects: %+v", mc.Rects)
	}
}

func TestTabsSpecFocusRing(t *testing.T) {
	tok := lightTokens(t)
	tabs, trs := buildTabs(false)
	uitest.LayoutWidget(tabs, 800, 600)

	// Keyboard focus (FocusGained without mouse press) draws the ring.
	trs[1].Event(uitest.NewMockContext(), uitest.FocusGained())
	mc := uitest.DrawWidget(tabs)

	want := tok.Ring
	want.A = metrics.RingAlpha
	foundRing := false
	for _, sr := range mc.StrokeRoundRects {
		if sr.Color == want && approx(sr.StrokeWidth, metrics.RingWidth) {
			foundRing = true
		}
	}
	if !foundRing {
		t.Fatalf("focus ring missing; strokes: %+v", mc.StrokeRoundRects)
	}
	foundBorder := false
	for _, sr := range mc.StrokeRoundRects {
		if sr.Color == tok.Ring && approx(sr.StrokeWidth, 1) {
			foundBorder = true
		}
	}
	if !foundBorder {
		t.Fatal("focused trigger must draw a solid ring border")
	}
}

func TestTabsInteraction(t *testing.T) {
	lightTokens(t)
	sig := state.NewSignal("account")
	trs := []*graft.TabsTriggerWidget{
		graft.TabsTrigger("account", "Account"),
		graft.TabsTrigger("password", "Password"),
		graft.TabsTrigger("more", "More").Disabled(true),
	}
	tabs := graft.Tabs(
		graft.TabsList(trs...),
		graft.TabsContent("account", graft.Text("A")),
		graft.TabsContent("password", graft.Text("B")),
	).Bind(sig)
	uitest.LayoutWidget(tabs, 800, 600)

	// Click the second trigger (translate into window coords: list at 0,0).
	ctx := uitest.NewMockContext()
	b := trs[1].Bounds() // list-local; list at origin of tabs
	pt := b.Center()
	tabs.Event(ctx, uitest.Click(pt.X, pt.Y))
	if sig.Get() != "password" {
		t.Fatalf("click: signal = %q, want password", sig.Get())
	}

	// Arrow navigation from password: Right skips the disabled trigger
	// and wraps to account.
	if !trs[1].IsFocused() {
		t.Fatal("clicked trigger should hold focus")
	}
	trs[1].Event(ctx, uitest.KeyPress(event.KeyLeft, 0))
	if sig.Get() != "account" {
		t.Fatalf("ArrowLeft: signal = %q, want account", sig.Get())
	}
	if !trs[0].IsFocused() {
		t.Fatal("ArrowLeft should move focus to the first trigger")
	}
	trs[0].Event(ctx, uitest.KeyPress(event.KeyRight, 0))
	if sig.Get() != "password" {
		t.Fatalf("ArrowRight: signal = %q, want password", sig.Get())
	}
	trs[1].Event(ctx, uitest.KeyPress(event.KeyRight, 0))
	if sig.Get() != "account" {
		t.Fatalf("ArrowRight wrap: signal = %q, want account (disabled skipped)", sig.Get())
	}
}

// buildTabsVertical returns a vertical three-trigger tabs tree.
func buildTabsVertical() (*graft.TabsWidget, []*graft.TabsTriggerWidget) {
	trs := []*graft.TabsTriggerWidget{
		graft.TabsTrigger("account", "Account"),
		graft.TabsTrigger("password", "Password"),
		graft.TabsTrigger("more", "More").Disabled(true),
	}
	t := graft.Tabs(
		graft.TabsList(trs...),
		graft.TabsContent("account", graft.Text("Account content")),
		graft.TabsContent("password", graft.Text("Password content")),
	).Vertical()
	return t, trs
}

func TestTabsVerticalLayout(t *testing.T) {
	lightTokens(t)
	tabs, trs := buildTabsVertical()
	uitest.LayoutWidget(tabs, 800, 600)

	m := metrics.Tabs
	list := tabs.Children()[0].(*graft.TabsListWidget)
	content := tabs.Children()[1].(*graft.TabsContentWidget)

	// List sits at the left edge; content sits to its right past the gap.
	if !approx(list.Bounds().Min.X, 0) {
		t.Fatalf("list x = %v, want 0", list.Bounds().Min.X)
	}
	if !approx(content.Bounds().Min.X, list.Bounds().Max.X+m.RootGap) {
		t.Fatalf("content x = %v, want %v (list right + gap)", content.Bounds().Min.X, list.Bounds().Max.X+m.RootGap)
	}
	// Content shares the top edge with the list (row layout, not stacked).
	if !approx(content.Bounds().Min.Y, 0) {
		t.Fatalf("content y = %v, want 0 (beside the list, not below)", content.Bounds().Min.Y)
	}

	// Triggers stack vertically: same x, increasing y, equal (stretched) width.
	b0, b1 := trs[0].Bounds(), trs[1].Bounds()
	if !approx(b0.Min.X, m.ListPadding) || !approx(b1.Min.X, m.ListPadding) {
		t.Fatalf("triggers must share x %v: got %v, %v", m.ListPadding, b0.Min.X, b1.Min.X)
	}
	if !approx(b1.Min.Y, b0.Max.Y) {
		t.Fatalf("second trigger y = %v, want %v (stacked under first)", b1.Min.Y, b0.Max.Y)
	}
	if !approx(b0.Width(), b1.Width()) {
		t.Fatalf("triggers must stretch to equal width: %v vs %v", b0.Width(), b1.Width())
	}
	// Column width = widest trigger ("Password") + padding both sides.
	if !approx(list.Bounds().Width(), b0.Width()+2*m.ListPadding) {
		t.Fatalf("list width = %v, want trigger width + 2*padding (%v)", list.Bounds().Width(), b0.Width()+2*m.ListPadding)
	}
}

func TestTabsVerticalKeyboard(t *testing.T) {
	lightTokens(t)
	sig := state.NewSignal("account")
	trs := []*graft.TabsTriggerWidget{
		graft.TabsTrigger("account", "Account"),
		graft.TabsTrigger("password", "Password"),
		graft.TabsTrigger("more", "More").Disabled(true),
	}
	tabs := graft.Tabs(
		graft.TabsList(trs...),
		graft.TabsContent("account", graft.Text("A")),
		graft.TabsContent("password", graft.Text("B")),
	).Bind(sig).Vertical()
	uitest.LayoutWidget(tabs, 800, 600)

	ctx := uitest.NewMockContext()
	ctx.RequestFocus(trs[0]) // focus the first trigger

	// Down moves to the next enabled trigger and activates it.
	trs[0].Event(ctx, uitest.KeyPress(event.KeyDown, 0))
	if sig.Get() != "password" {
		t.Fatalf("ArrowDown: signal = %q, want password", sig.Get())
	}
	if !trs[1].IsFocused() {
		t.Fatal("ArrowDown should move focus to the second trigger")
	}

	// Up moves back to the first.
	trs[1].Event(ctx, uitest.KeyPress(event.KeyUp, 0))
	if sig.Get() != "account" {
		t.Fatalf("ArrowUp: signal = %q, want account", sig.Get())
	}
	if !trs[0].IsFocused() {
		t.Fatal("ArrowUp should move focus to the first trigger")
	}

	// Down again, then Down wraps past the disabled trigger back to account.
	trs[0].Event(ctx, uitest.KeyPress(event.KeyDown, 0))
	if sig.Get() != "password" {
		t.Fatalf("ArrowDown (2): signal = %q, want password", sig.Get())
	}
	trs[1].Event(ctx, uitest.KeyPress(event.KeyDown, 0))
	if sig.Get() != "account" {
		t.Fatalf("ArrowDown wrap: signal = %q, want account (disabled skipped)", sig.Get())
	}

	// Left/Right are no-ops in vertical orientation.
	ctx.RequestFocus(trs[0])
	if handled := trs[0].Event(ctx, uitest.KeyPress(event.KeyRight, 0)); handled {
		t.Fatal("ArrowRight must not be handled in vertical orientation")
	}
	if sig.Get() != "account" {
		t.Fatalf("ArrowRight in vertical changed selection to %q", sig.Get())
	}

	// Home/End still jump to the edges.
	ctx.RequestFocus(trs[1])
	trs[1].Event(ctx, uitest.KeyPress(event.KeyHome, 0))
	if sig.Get() != "account" || !trs[0].IsFocused() {
		t.Fatalf("Home: signal = %q focused0 = %v", sig.Get(), trs[0].IsFocused())
	}
}

// TestTabsHorizontalKeyboardUnchanged guards that horizontal arrow keys keep
// working (Left/Right move; Up/Down are no-ops).
func TestTabsHorizontalKeyboardUnchanged(t *testing.T) {
	lightTokens(t)
	sig := state.NewSignal("account")
	trs := []*graft.TabsTriggerWidget{
		graft.TabsTrigger("account", "Account"),
		graft.TabsTrigger("password", "Password"),
	}
	tabs := graft.Tabs(
		graft.TabsList(trs...),
		graft.TabsContent("account", graft.Text("A")),
		graft.TabsContent("password", graft.Text("B")),
	).Bind(sig)
	uitest.LayoutWidget(tabs, 800, 600)

	ctx := uitest.NewMockContext()
	ctx.RequestFocus(trs[0])

	// Right moves; Up/Down are no-ops in horizontal orientation.
	trs[0].Event(ctx, uitest.KeyPress(event.KeyRight, 0))
	if sig.Get() != "password" {
		t.Fatalf("ArrowRight: signal = %q, want password", sig.Get())
	}
	if handled := trs[1].Event(ctx, uitest.KeyPress(event.KeyDown, 0)); handled {
		t.Fatal("ArrowDown must not be handled in horizontal orientation")
	}
	if sig.Get() != "password" {
		t.Fatalf("ArrowDown in horizontal changed selection to %q", sig.Get())
	}
	trs[1].Event(ctx, uitest.KeyPress(event.KeyLeft, 0))
	if sig.Get() != "account" {
		t.Fatalf("ArrowLeft: signal = %q, want account", sig.Get())
	}
}

func TestGoldenTabs(t *testing.T) {
	build := func(line bool) func() widget.Widget {
		return func() widget.Widget {
			tabs, _ := buildTabs(line)
			return primitives.Box(tabs).Padding(16)
		}
	}
	gtest.GoldenLightDark(t, "tabs-default", build(false))
	gtest.GoldenLightDark(t, "tabs-line", build(true))

	gtest.GoldenLightDark(t, "tabs-focus", func() widget.Widget {
		tabs, trs := buildTabs(false)
		uitest.LayoutWidget(tabs, 800, 600)
		trs[1].Event(uitest.NewMockContext(), uitest.FocusGained())
		return primitives.Box(tabs).Padding(16)
	})
}

func TestGoldenTabsVertical(t *testing.T) {
	gtest.GoldenLightDark(t, "tabs-vertical", func() widget.Widget {
		tabs, _ := buildTabsVertical()
		return primitives.Box(tabs).Padding(16)
	})
}

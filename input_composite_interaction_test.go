package graft_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/metrics"
)

// loadAssetsT is the shared per-test asset loader for the interaction tests in
// this file.
func loadAssetsT(t *testing.T) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
}

// TestTabsUncontrolledClickSwitches verifies that a Tabs configured with an
// initial .Value (uncontrolled, no Bind) switches its active content when a
// trigger is clicked — the existing TestTabsInteraction only exercises the
// signal-bound path.
func TestTabsUncontrolledClickSwitches(t *testing.T) {
	loadAssetsT(t)

	trs := []*graft.TabsTriggerWidget{
		graft.TabsTrigger("a", "Account"),
		graft.TabsTrigger("b", "Password"),
	}
	tabs := graft.Tabs(
		graft.TabsList(trs...),
		graft.TabsContent("a", graft.Text("A content")),
		graft.TabsContent("b", graft.Text("B content")),
	).Value("a")
	uitest.LayoutWidget(tabs, 800, 600)

	// Initially only A content shows.
	mc0 := uitest.DrawWidget(tabs)
	if !drewAnyText(mc0, "A content") || drewAnyText(mc0, "B content") {
		t.Fatalf("initial uncontrolled state wrong; texts=%v", interactionTexts(mc0))
	}

	ctx := uitest.NewMockContext()
	pt := trs[1].Bounds().Center()
	tabs.Event(ctx, uitest.Click(pt.X, pt.Y))
	tabs.Event(ctx, uitest.Release(pt.X, pt.Y))

	uitest.LayoutWidget(tabs, 800, 600)
	mc := uitest.DrawWidget(tabs)
	if drewAnyText(mc, "A content") {
		t.Fatal("uncontrolled tabs: A content still drawn after clicking B")
	}
	if !drewAnyText(mc, "B content") {
		t.Fatal("uncontrolled tabs: B content not drawn after clicking B")
	}
}

// TestCalendarMonthNavigation verifies the prev/next ghost buttons advance and
// rewind the displayed month and repaint the grid.
func TestCalendarMonthNavigation(t *testing.T) {
	loadAssetsT(t)

	cal := graft.Calendar().Month(time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC))
	uitest.LayoutWidget(cal, 900, 900)
	if mc := uitest.DrawWidget(cal); !drewAnyText(mc, "June 2026") {
		t.Fatalf("caption not June 2026; texts=%v", interactionTexts(mc))
	}

	m := metrics.Calendar
	ctx := uitest.NewMockContext()

	// Next button: center of the last header-row cell.
	nx, ny := cal.Bounds().Width()-m.CellSize/2, m.CellSize/2
	cal.Event(ctx, uitest.Click(nx, ny))
	cal.Event(ctx, uitest.Release(nx, ny))
	uitest.LayoutWidget(cal, 900, 900)
	if mc := uitest.DrawWidget(cal); !drewAnyText(mc, "July 2026") {
		t.Fatalf("next nav did not advance to July 2026; texts=%v", interactionTexts(mc))
	}

	// Prev button twice: July -> June -> May.
	px, py := m.CellSize/2, m.CellSize/2
	for i := 0; i < 2; i++ {
		cal.Event(ctx, uitest.Click(px, py))
		cal.Event(ctx, uitest.Release(px, py))
		uitest.LayoutWidget(cal, 900, 900)
	}
	if mc := uitest.DrawWidget(cal); !drewAnyText(mc, "May 2026") {
		t.Fatalf("two prev navs did not reach May 2026; texts=%v", interactionTexts(mc))
	}
}

// TestCalendarOutsideDayNavigates verifies selecting a day from an adjacent
// month both fires OnSelect for that date and navigates the grid to its month.
func TestCalendarOutsideDayNavigates(t *testing.T) {
	loadAssetsT(t)

	var got time.Time
	cal := graft.Calendar().
		Month(time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC)).
		OnSelect(func(d time.Time) { got = d })
	uitest.LayoutWidget(cal, 900, 900)

	// June 1 2026 is a Monday, so the Sunday-start grid's first cell (row0,col0)
	// is May 31 (outside the displayed month).
	m := metrics.Calendar
	firstWeekY := m.CellSize + m.CaptionGap + m.CellSize + m.HeaderGap
	cx, cy := m.CellSize/2, firstWeekY+m.CellSize/2

	ctx := uitest.NewMockContext()
	cal.Event(ctx, uitest.Click(cx, cy))
	cal.Event(ctx, uitest.Release(cx, cy))

	if got.IsZero() {
		t.Fatal("outside-month day click did not fire OnSelect")
	}
	if got.Month() != time.May || got.Day() != 31 {
		t.Fatalf("selected = %v, want 2026-05-31", got.Format("2006-01-02"))
	}
	uitest.LayoutWidget(cal, 900, 900)
	if mc := uitest.DrawWidget(cal); !drewAnyText(mc, "May 2026") {
		t.Fatalf("selecting an outside day did not navigate to May 2026; texts=%v", interactionTexts(mc))
	}
}

// TestComboboxUncontrolledSelectClosesAndChanges drives the open -> click-row
// path on an uncontrolled combobox and asserts OnChange fires with the row's
// value and the popover closes.
func TestComboboxUncontrolledSelectClosesAndChanges(t *testing.T) {
	loadAssetsT(t)

	var got string
	cb := graft.Combobox(
		graft.ComboboxItem("next", "Next.js"),
		graft.ComboboxItem("remix", "Remix"),
	).OnChange(func(v string) { got = v })

	ctx := uitest.NewMockContext()
	om := newFakeOverlayManager()
	ctx.OverlayVal = om
	uitest.LayoutWidget(cb, 800, 600)
	uitest.DrawWidgetWithContext(cb, ctx)

	uitest.SimulateClickWithContext(cb, ctx, 10, 10)
	if !cb.IsOpen() || len(om.pushed) != 1 {
		t.Fatalf("combobox did not open (open=%v pushed=%d)", cb.IsOpen(), len(om.pushed))
	}
	content := om.pushed[0]
	content.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))
	b := content.(interface{ Bounds() geometry.Rect }).Bounds()

	m := metrics.Combobox
	rowX := b.Min.X + b.Width()/2
	rowY := b.Min.Y + m.InputHeight + m.ListPad + m.ItemHeight + m.ItemHeight/2 // second row
	content.Event(ctx, uitest.Click(rowX, rowY))
	content.Event(ctx, uitest.Release(rowX, rowY))

	if got != "remix" {
		t.Fatalf("OnChange got %q, want remix", got)
	}
	if cb.IsOpen() {
		t.Fatal("combobox should close after a row selection")
	}
}

// TestInputInsertsCJKRunes verifies a focused Input inserts non-ASCII runes
// (the graft wrapper must not drop runes >= 0x20 of any language; OS IME
// delivery is out of scope).
func TestInputInsertsCJKRunes(t *testing.T) {
	loadAssetsT(t)

	in := graft.Input()
	in.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(300, 100)))
	ctx := uitest.NewMockContext()
	in.Event(ctx, uitest.Click(10, 18)) // focus via mouse press
	for _, r := range "中文字" {
		in.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeyUnknown, r, 0))
	}
	if got := in.Text(); got != "中文字" {
		t.Fatalf("Input CJK insertion: got %q want 中文字", got)
	}
}

// TestInputGroupInsertsCJKAndButtonClick verifies the InputGroup's borderless
// field inserts CJK runes and fires OnChange, and that a trailing button addon
// receives its click.
func TestInputGroupInsertsCJKAndButtonClick(t *testing.T) {
	loadAssetsT(t)

	var changed string
	g := graft.InputGroup().OnChange(func(s string) { changed = s })
	g.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(300, 100)))
	ctx := uitest.NewMockContext()
	g.Event(ctx, uitest.Click(40, 18))
	for _, r := range "你好" {
		g.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeyUnknown, r, 0))
	}
	if g.Text() != "你好" || changed != "你好" {
		t.Fatalf("InputGroup CJK: text=%q onChange=%q want 你好", g.Text(), changed)
	}

	clicked := false
	g2 := graft.InputGroup().
		Value("x").
		Trailing(graft.InputGroupButton(graft.Button("Copy").Outline().OnClick(func() { clicked = true }))).
		W(300)
	g2.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(300, 100)))
	var btn *graft.ButtonWidget
	for _, ch := range g2.Children() {
		if b, ok := ch.(*graft.ButtonWidget); ok {
			btn = b
		}
	}
	if btn == nil {
		t.Fatal("trailing button addon missing from Children()")
	}
	pt := btn.Bounds().Center()
	ctx2 := uitest.NewMockContext()
	g2.Event(ctx2, uitest.Click(pt.X, pt.Y))
	g2.Event(ctx2, uitest.Release(pt.X, pt.Y))
	if !clicked {
		t.Fatal("trailing button addon did not fire OnClick")
	}
}

// TestSliderArrowKeysAdjustValue verifies keyboard adjustment (the existing
// slider tests cover drag and hover only).
func TestSliderArrowKeysAdjustValue(t *testing.T) {
	loadAssetsT(t)

	var observed float32 = -1
	s := graft.Slider().Value(50).Step(1).W(200).OnChange(func(v float32) { observed = v })
	uitest.LayoutWidget(s, 200, 16)
	ctx := uitest.NewMockContext()
	s.Event(ctx, uitest.Click(100, 8)) // focus the core slider
	s.Event(ctx, uitest.Release(100, 8))
	s.Event(ctx, uitest.KeyPress(event.KeyRight, 0))
	if observed != 51 {
		t.Fatalf("ArrowRight: OnChange observed %v, want 51", observed)
	}
	s.Event(ctx, uitest.KeyPress(event.KeyLeft, 0))
	if observed != 50 {
		t.Fatalf("ArrowLeft: OnChange observed %v, want 50", observed)
	}
}

// interactionTexts collects all text draws for diagnostics.
func interactionTexts(mc *uitest.MockCanvas) []string {
	var out []string
	for _, st := range mc.StyledTexts {
		out = append(out, st.Text)
	}
	for _, dt := range mc.Texts {
		out = append(out, dt.Text)
	}
	return out
}

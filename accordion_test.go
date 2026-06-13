package graft_test

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// buildAccordion returns a standard three-item accordion with the first item
// open by default.
func buildAccordion() *graft.AccordionWidget {
	return graft.Accordion(
		graft.AccordionItem("item-1", "Is it accessible?",
			graft.Text("Yes. It adheres to the WAI-ARIA design pattern.")),
		graft.AccordionItem("item-2", "Is it styled?",
			graft.Text("Yes. It comes with default styles.")),
		graft.AccordionItem("item-3", "Is it animated?",
			graft.Text("Yes. It is animated by default.")),
	).Type("single").Value("item-1")
}

func TestAccordionSpecLayout(t *testing.T) {
	lightTokens(t)
	a := buildAccordion()
	uitest.LayoutWidget(a, 400, 600)

	m := metrics.Accordion
	items := a.Children()

	// Trigger height = line-height 20 + 2*py16 = 52.
	wantTrig := m.TriggerLineHeight + 2*m.TriggerPadY
	it0 := items[0].(bounded)
	it1 := items[1].(bounded)

	// item-1 is open: its height exceeds the bare trigger by content + pb-4.
	if it0.Bounds().Height() <= wantTrig {
		t.Fatalf("open item height = %v, want > trigger %v", it0.Bounds().Height(), wantTrig)
	}
	// item-2 is closed: exactly the trigger height.
	if !approx(it1.Bounds().Height(), wantTrig) {
		t.Fatalf("closed item height = %v, want %v", it1.Bounds().Height(), wantTrig)
	}
	// Items stack with no gap.
	if !approx(it1.Bounds().Min.Y, it0.Bounds().Max.Y) {
		t.Fatalf("items not adjacent: item2 y=%v item1 maxY=%v", it1.Bounds().Min.Y, it0.Bounds().Max.Y)
	}
}

func TestAccordionSpecChevronAndBorders(t *testing.T) {
	tok := lightTokens(t)
	a := buildAccordion()
	uitest.LayoutWidget(a, 400, 600)
	mc := uitest.DrawWidget(a)

	m := metrics.Accordion

	// Three items, last has no bottom border → two 1px border-token rules.
	borders := 0
	for _, ln := range mc.Lines {
		if ln.Color == tok.Border && approx(ln.StrokeWidth, m.ItemBorderWidth) {
			borders++
		}
	}
	if borders != 2 {
		t.Fatalf("border-b rules = %d, want 2 (last:border-b-0); lines: %+v", borders, mc.Lines)
	}

	// Open content of item-1 is drawn; closed content is not.
	var openDrawn, closedDrawn bool
	for _, st := range mc.StyledTexts {
		switch st.Text {
		case "Yes. It adheres to the WAI-ARIA design pattern.":
			openDrawn = true
		case "Yes. It comes with default styles.":
			closedDrawn = true
		}
	}
	if !openDrawn {
		t.Fatal("open item content must draw")
	}
	if closedDrawn {
		t.Fatal("closed item content must not draw")
	}

	// Trigger label is foreground / Geist Medium / 14px.
	for _, st := range mc.StyledTexts {
		if st.Text == "Is it accessible?" {
			if st.Style.Color != tok.Foreground {
				t.Fatalf("trigger label color = %+v, want foreground", st.Style.Color)
			}
			if st.Style.FontSize != m.TriggerFontSize {
				t.Fatalf("trigger font size = %v, want %v", st.Style.FontSize, m.TriggerFontSize)
			}
			if st.Style.FontFamily != "Geist Medium" {
				t.Fatalf("trigger family = %q, want Geist Medium", st.Style.FontFamily)
			}
		}
	}
}

func TestAccordionSingleExclusive(t *testing.T) {
	lightTokens(t)
	a := graft.Accordion(
		graft.AccordionItem("a", "A", graft.Text("body a")),
		graft.AccordionItem("b", "B", graft.Text("body b")),
	).Type("single").Value("a")
	uitest.LayoutWidget(a, 400, 600)

	ctx := uitest.NewMockContext()
	itB := a.Children()[1].(bounded)
	// Click the second trigger (top half of its row).
	pt := itB.Bounds().Center()
	pt.Y = itB.Bounds().Min.Y + 10
	a.Event(ctx, uitest.Click(pt.X, pt.Y))
	a.Event(ctx, uitest.Release(pt.X, pt.Y))

	uitest.LayoutWidget(a, 400, 600)
	mc := uitest.DrawWidget(a)
	var aOpen, bOpen bool
	for _, st := range mc.StyledTexts {
		switch st.Text {
		case "body a":
			aOpen = true
		case "body b":
			bOpen = true
		}
	}
	if aOpen {
		t.Fatal("single mode: opening B must close A")
	}
	if !bOpen {
		t.Fatal("single mode: B should be open after click")
	}
}

func TestAccordionMultiple(t *testing.T) {
	lightTokens(t)
	a := graft.Accordion(
		graft.AccordionItem("a", "A", graft.Text("body a")),
		graft.AccordionItem("b", "B", graft.Text("body b")),
	).Type("multiple").Value("a")
	uitest.LayoutWidget(a, 400, 600)

	ctx := uitest.NewMockContext()
	itB := a.Children()[1].(bounded)
	pt := itB.Bounds().Center()
	pt.Y = itB.Bounds().Min.Y + 10
	a.Event(ctx, uitest.Click(pt.X, pt.Y))
	a.Event(ctx, uitest.Release(pt.X, pt.Y))

	uitest.LayoutWidget(a, 400, 600)
	mc := uitest.DrawWidget(a)
	var aOpen, bOpen bool
	for _, st := range mc.StyledTexts {
		switch st.Text {
		case "body a":
			aOpen = true
		case "body b":
			bOpen = true
		}
	}
	if !aOpen || !bOpen {
		t.Fatalf("multiple mode: both should be open; a=%v b=%v", aOpen, bOpen)
	}
}

func TestGoldenAccordion(t *testing.T) {
	gtest.GoldenLightDark(t, "accordion-single", func() widget.Widget {
		a := buildAccordion()
		return primitives.Box(a).Padding(16).Width(420)
	})
}

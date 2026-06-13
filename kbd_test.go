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

func TestKbdSpecSingleChip(t *testing.T) {
	tok := lightTokens(t)
	th := graft.CurrentTheme()
	k := graft.Kbd("K")
	uitest.LayoutWidget(k, 200, 100)
	mc := uitest.DrawWidget(k)

	m := metrics.Kbd
	// Height = 20, single-glyph chip clamps to min-w-5 = 20 (square).
	if !approx(k.Bounds().Height(), m.Height) {
		t.Fatalf("kbd height = %v, want %v", k.Bounds().Height(), m.Height)
	}
	if !approx(k.Bounds().Width(), m.MinWidth) {
		t.Fatalf("single-glyph kbd width = %v, want min-w %v", k.Bounds().Width(), m.MinWidth)
	}

	// Chip: muted fill at radius-sm.
	found := false
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Muted && approx(rr.Radius, th.RadiusSM()) &&
			approx(rr.Bounds.Height(), m.Height) {
			found = true
		}
	}
	if !found {
		t.Fatalf("kbd chip (muted, radius-sm) missing; roundrects %+v", mc.RoundRects)
	}

	// Label: muted-foreground, font-sans (Geist Medium), 12px.
	for _, st := range mc.StyledTexts {
		if st.Text == "K" {
			if st.Style.Color != tok.MutedForeground {
				t.Fatalf("kbd label color = %+v, want muted-foreground", st.Style.Color)
			}
			if st.Style.FontSize != m.FontSize {
				t.Fatalf("kbd font size = %v, want %v", st.Style.FontSize, m.FontSize)
			}
			if st.Style.FontFamily != "Geist Medium" {
				t.Fatalf("kbd family = %q, want Geist Medium (font-sans 500, NOT mono)", st.Style.FontFamily)
			}
		}
	}
}

func TestKbdSpecComboRow(t *testing.T) {
	tok := lightTokens(t)
	k := graft.Kbd("Ctrl", "B")
	uitest.LayoutWidget(k, 200, 100)
	mc := uitest.DrawWidget(k)

	m := metrics.Kbd
	// Two chips → two muted roundrects.
	chips := 0
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Muted {
			chips++
		}
	}
	if chips != 2 {
		t.Fatalf("combo kbd should draw 2 chips; got %d", chips)
	}

	// Combo is wider than a single chip (chip1 + gap + chip2).
	if k.Bounds().Width() <= m.MinWidth {
		t.Fatalf("combo kbd width = %v, want > single chip", k.Bounds().Width())
	}
	// Both labels rendered.
	var gotCtrl, gotB bool
	for _, st := range mc.StyledTexts {
		switch st.Text {
		case "Ctrl":
			gotCtrl = true
		case "B":
			gotB = true
		}
	}
	if !gotCtrl || !gotB {
		t.Fatalf("combo kbd labels: ctrl=%v b=%v", gotCtrl, gotB)
	}
}

func TestGoldenKbd(t *testing.T) {
	gtest.GoldenLightDark(t, "kbd", func() widget.Widget {
		return primitives.VBox(
			graft.Kbd("K"),
			graft.Kbd("Cmd", "K"),
			graft.Kbd("Ctrl", "Shift", "P"),
		).Gap(12).Padding(16).CrossAlign(primitives.CrossAxisStart)
	})
}

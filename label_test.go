package graft_test

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/internal/textmetrics"
)

func TestLabelMetrics(t *testing.T) {
	tok := lightTokens(t)
	l := graft.Label("Email address")
	size := uitest.LayoutWidget(l, 800, 600)

	// leading-none: line box equals the 14px font size.
	if size.Height != 14 {
		t.Errorf("height = %v, want 14 (text-sm leading-none)", size.Height)
	}
	if want := textmetrics.Width("Geist Medium", 14, "Email address"); size.Width != want {
		t.Errorf("width = %v, want %v", size.Width, want)
	}

	c := uitest.DrawWidget(l)
	if len(c.StyledTexts) != 1 {
		t.Fatalf("styled texts = %d, want 1", len(c.StyledTexts))
	}
	txt := c.StyledTexts[0]
	if txt.Style.FontFamily != "Geist Medium" {
		t.Errorf("family = %q, want Geist Medium (font-medium)", txt.Style.FontFamily)
	}
	if txt.Style.FontSize != 14 {
		t.Errorf("size = %v, want 14", txt.Style.FontSize)
	}
	if txt.Style.Color != tok.Foreground {
		t.Errorf("color = %v, want foreground", txt.Style.Color)
	}
}

func TestLabelDisabled(t *testing.T) {
	tok := lightTokens(t)
	l := graft.Label("Email").Disabled(true)
	uitest.LayoutWidget(l, 800, 600)
	c := uitest.DrawWidget(l)
	if want := mulAlpha(tok.Foreground, 0.5); c.StyledTexts[0].Style.Color != want {
		t.Errorf("disabled color = %v, want foreground at 50%% %v", c.StyledTexts[0].Style.Color, want)
	}
}

func TestGoldenLabel(t *testing.T) {
	gtest.GoldenLightDark(t, "label", func() widget.Widget {
		return primitives.VBox(
			graft.Label("Email address"),
			graft.Label("Disabled label").Disabled(true),
		).Gap(12).Padding(24)
	})
}

package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/theme"
)

// forceLightMode pins the shared theme to light for token assertions and
// restores the previous mode on cleanup.
func forceLightMode(t *testing.T) *theme.Theme {
	t.Helper()
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	t.Cleanup(func() { th.SetMode(prev) })
	return th
}

func TestSeparatorHorizontalSpec(t *testing.T) {
	th := forceLightMode(t)

	sep := graft.Separator()
	size := uitest.LayoutWidget(sep, 300, 100)
	if size != geometry.Sz(300, 1) {
		t.Fatalf("horizontal separator size: got %v, want 300x1", size)
	}

	cv := uitest.DrawWidget(sep)
	if len(cv.Rects) != 1 {
		t.Fatalf("draw calls: got %d rects, want 1", len(cv.Rects))
	}
	r := cv.Rects[0]
	if r.Bounds != geometry.NewRect(0, 0, 300, 1) {
		t.Errorf("rect bounds: got %v", r.Bounds)
	}
	if r.Color != th.Active().Border {
		t.Errorf("rect color: got %v, want border token %v", r.Color, th.Active().Border)
	}
}

func TestSeparatorVerticalSpec(t *testing.T) {
	forceLightMode(t)

	sep := graft.Separator().Vertical()
	size := uitest.LayoutWidget(sep, 300, 48)
	if size != geometry.Sz(1, 48) {
		t.Fatalf("vertical separator size: got %v, want 1x48", size)
	}

	cv := uitest.DrawWidget(sep)
	if len(cv.Rects) != 1 {
		t.Fatalf("draw calls: got %d rects, want 1", len(cv.Rects))
	}
	if cv.Rects[0].Bounds != geometry.NewRect(0, 0, 1, 48) {
		t.Errorf("rect bounds: got %v", cv.Rects[0].Bounds)
	}
}

func TestGoldenSeparator(t *testing.T) {
	gtest.GoldenLightDark(t, "separator-horizontal", func() widget.Widget {
		return primitives.VBox(
			graft.Text("An open-source UI component library."),
			graft.Separator(),
			graft.Text("Install and customize at will."),
		).Gap(12).Padding(16).Width(300)
	})

	gtest.GoldenLightDark(t, "separator-vertical", func() widget.Widget {
		return primitives.HBox(
			graft.Text("Blog"),
			graft.Separator().Vertical(),
			graft.Text("Docs"),
			graft.Separator().Vertical(),
			graft.Text("Source"),
		).Gap(12).Padding(16).Height(52).CrossAlign(primitives.CrossAxisCenter)
	})
}

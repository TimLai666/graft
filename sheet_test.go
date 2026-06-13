package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// sheetDemoContent builds the canonical demo body used by spec tests and
// goldens: a header (title + description) and a footer button.
func sheetDemoContent() *graft.SheetContentWidget {
	return graft.SheetContent(
		graft.SheetHeader(
			graft.SheetTitle("Edit profile"),
			graft.SheetDescription("Make changes to your profile here."),
		),
		graft.SheetFooter(
			ovButton("Save changes", true),
		),
	)
}

// TestSheetContentSurface pins the panel surface: bg Background, square (no
// radius), shadow-LG, and a single 1px Border on the inner edge.
func TestSheetContentSurface(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	tok := graft.CurrentTheme().Active()

	content := sheetDemoContent()
	size := content.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(900, 900)))
	if size.Width != metrics.SheetMaxWidth {
		t.Fatalf("panel width: got %v want %v (max-w-sm)", size.Width, metrics.SheetMaxWidth)
	}

	content.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	canvas := uitest.DrawWidget(content)

	// Surface fill: bg Background, full bounds, square (DrawRect, not RoundRect).
	var surface *uitest.DrawRectCall
	for i := range canvas.Rects {
		if canvas.Rects[i].Bounds.Size() == size && canvas.Rects[i].Color == tok.Background {
			surface = &canvas.Rects[i]
			break
		}
	}
	if surface == nil {
		t.Fatalf("no background surface DrawRect at size %v", size)
	}

	// Inner-edge border: a 1px-wide Border-colored vertical line on the left
	// edge (default side = right ⇒ border-l).
	var border *uitest.DrawRectCall
	for i := range canvas.Rects {
		r := &canvas.Rects[i]
		if r.Color == tok.Border && r.Bounds.Width() == metrics.SheetBorderWidth &&
			r.Bounds.Height() == size.Height {
			border = r
			break
		}
	}
	if border == nil {
		t.Fatal("no 1px inner-edge border line")
	}
	if border.Bounds.Min.X != 0 {
		t.Errorf("right sheet border-l should sit at x=0, got %v", border.Bounds.Min.X)
	}
}

// TestSheetSideAnchoring pins the panel origin and shape for each side in a
// known viewport.
func TestSheetSideAnchoring(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	ctx := uitest.NewMockContext()
	viewport := geometry.Sz(1000, 800)

	cases := []struct {
		side       graft.SheetSide
		wantOrigin geometry.Point
		fullHeight bool
		fullWidth  bool
	}{
		{graft.SheetRight, geometry.Pt(1000-metrics.SheetMaxWidth, 0), true, false},
		{graft.SheetLeft, geometry.Pt(0, 0), true, false},
		{graft.SheetTop, geometry.Pt(0, 0), false, true},
	}
	for _, tc := range cases {
		content := sheetDemoContent().Side(tc.side)
		host := graft.Sheet(content).Open(true)

		om := &ovFakeOverlayManager{}
		ctx.OverlayVal = om
		host.Layout(ctx, looseConstraints())
		// The pushed overlay lays out the content against the window size.
		if len(om.pushed) != 1 {
			t.Fatalf("side %d: overlay not pushed", tc.side)
		}
		ov := om.pushed[0]
		ov.Layout(ctx, geometry.Tight(viewport))

		b := content.Bounds()
		if b.Min != tc.wantOrigin {
			t.Errorf("side %d: origin got %v want %v", tc.side, b.Min, tc.wantOrigin)
		}
		if tc.fullHeight && b.Height() != viewport.Height {
			t.Errorf("side %d: height got %v want full %v", tc.side, b.Height(), viewport.Height)
		}
		if tc.fullWidth && b.Width() != viewport.Width {
			t.Errorf("side %d: width got %v want full %v", tc.side, b.Width(), viewport.Width)
		}
		if !tc.fullWidth && b.Width() != metrics.SheetMaxWidth {
			t.Errorf("side %d: width got %v want %v", tc.side, b.Width(), metrics.SheetMaxWidth)
		}
	}
}

// TestSheetHostShowHide drives the host's open signal against a fake overlay
// manager: setting open pushes a modal overlay; clearing it removes it.
func TestSheetHostShowHide(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	open := state.NewSignal(false)
	host := graft.Sheet(sheetDemoContent()).Bind(open)

	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om

	host.Layout(ctx, looseConstraints())
	if om.liveCount() != 0 {
		t.Fatalf("sheet shown while closed: live=%d", om.liveCount())
	}

	open.Set(true)
	host.Layout(ctx, looseConstraints())
	if om.liveCount() != 1 {
		t.Fatalf("sheet not shown after open: live=%d", om.liveCount())
	}

	open.Set(false)
	host.Layout(ctx, looseConstraints())
	if om.liveCount() != 0 {
		t.Fatalf("sheet not hidden after close: live=%d", om.liveCount())
	}
}

// TestSheetTriggerOpens verifies the trigger flips the open signal on click.
func TestSheetTriggerOpens(t *testing.T) {
	defer ovForceLightMode(t)()
	open := state.NewSignal(false)
	trig := graft.SheetTrigger(primitives.Box().Width(80).Height(36), open)
	trig.Layout(uitest.NewMockContext(), looseConstraints())

	if uitest.SimulateClick(trig, 10, 10); !open.Get() {
		t.Fatal("trigger click did not open the sheet")
	}
}

// TestGoldenSheet renders the sheet exactly as it appears at runtime — a
// modal frame with the backdrop and the panel anchored full-height to its
// edge — via SheetPreview. Goldens render the SETTLED open state (slide
// offset zero), light + dark.
func TestGoldenSheet(t *testing.T) {
	build := func(side graft.SheetSide) func() widget.Widget {
		return func() widget.Widget {
			return graft.SheetPreview(sheetDemoContent().Side(side), geometry.Sz(640, 420))
		}
	}
	gtest.GoldenLightDark(t, "sheet-right", build(graft.SheetRight))
	gtest.GoldenLightDark(t, "sheet-left", build(graft.SheetLeft))
}

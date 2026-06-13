package graft_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// resizablePanelBox builds a colored panel body for goldens.
func resizablePanelBox(label string) widget.Widget {
	return primitives.Box(
		graft.Text(label).Align(widget.TextAlignCenter),
	).Padding(16)
}

// TestResizableSpecLayout pins the ratio split and the handle hit-band
// geometry for a two-panel horizontal group.
func TestResizableSpecLayout(t *testing.T) {
	lightTokens(t)

	g := graft.ResizablePanelGroup(graft.ResizableHorizontal,
		graft.ResizablePanel(primitives.Box().Width(10).Height(10)),
		graft.ResizableHandle(),
		graft.ResizablePanel(primitives.Box().Width(10).Height(10)),
	)
	size := uitest.LayoutWidgetTight(g, 405, 200)
	if size.Width != 405 || size.Height != 200 {
		t.Fatalf("group size: got %vx%v want 405x200", size.Width, size.Height)
	}

	ratios := g.Ratios()
	if len(ratios) != 2 || !approx(ratios[0], 0.5) || !approx(ratios[1], 0.5) {
		t.Fatalf("default ratios: got %v want [0.5 0.5]", ratios)
	}

	// With a single 4px handle, panels split 405-4 = 401 -> 200.5 each.
	children := g.Children()
	leftPanel := children[0].(interface{ Bounds() geometry.Rect })
	handle := children[1].(interface{ Bounds() geometry.Rect })
	rightPanel := children[2].(interface{ Bounds() geometry.Rect })

	if !approx(handle.Bounds().Width(), metrics.Resizable.HitWidth) {
		t.Fatalf("handle hit width: got %v want %v", handle.Bounds().Width(), metrics.Resizable.HitWidth)
	}
	wantHalf := (405 - metrics.Resizable.HitWidth) / 2
	if !approx(leftPanel.Bounds().Width(), wantHalf) {
		t.Fatalf("left panel width: got %v want %v", leftPanel.Bounds().Width(), wantHalf)
	}
	if !approx(rightPanel.Bounds().Width(), wantHalf) {
		t.Fatalf("right panel width: got %v want %v", rightPanel.Bounds().Width(), wantHalf)
	}
	// Handle sits between the two panels.
	if !approx(handle.Bounds().Min.X, wantHalf) {
		t.Fatalf("handle x: got %v want %v", handle.Bounds().Min.X, wantHalf)
	}
	if handle.Bounds().Height() != 200 {
		t.Fatalf("handle should span full cross axis: got %v want 200", handle.Bounds().Height())
	}
}

// TestResizableHandleLine asserts the divider draws a 1px border line, and
// the grip chip + glyph when WithHandle is set.
func TestResizableHandleLine(t *testing.T) {
	tok := lightTokens(t)

	g := graft.ResizablePanelGroup(graft.ResizableHorizontal,
		graft.ResizablePanel(primitives.Box().Width(10).Height(10)),
		graft.ResizableHandle().WithHandle(),
		graft.ResizablePanel(primitives.Box().Width(10).Height(10)),
	)
	uitest.LayoutWidgetTight(g, 405, 200)
	mc := uitest.DrawWidget(g)

	// 1px border line spanning the full height.
	foundLine := false
	for _, r := range mc.Rects {
		if r.Color == tok.Border && approx(r.Bounds.Width(), metrics.Resizable.LineWidth) &&
			approx(r.Bounds.Height(), 200) {
			foundLine = true
		}
	}
	if !foundLine {
		t.Fatalf("1px border divider line not drawn; rects: %+v", mc.Rects)
	}

	// Grip chip: a rounded rect 12 wide x 16 tall in --border.
	foundChip := false
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Border && approx(rr.Bounds.Width(), metrics.Resizable.GripShortSide) &&
			approx(rr.Bounds.Height(), metrics.Resizable.GripLongSide) {
			foundChip = true
		}
	}
	if !foundChip {
		t.Fatalf("grip chip not drawn; roundrects: %+v", mc.RoundRects)
	}
}

// TestResizableDrag drives a drag on the divider and checks the adjacent
// panel ratios shift while the others (none here) are preserved.
func TestResizableDrag(t *testing.T) {
	lightTokens(t)

	g := graft.ResizablePanelGroup(graft.ResizableHorizontal,
		graft.ResizablePanel(primitives.Box().Width(10).Height(10)),
		graft.ResizableHandle(),
		graft.ResizablePanel(primitives.Box().Width(10).Height(10)),
	)
	uitest.LayoutWidgetTight(g, 404, 200)
	// avail = 404 - 4 = 400; half = 200 -> handle at x=200.
	ctx := uitest.NewMockContext()

	// Press on the handle (hit band x in [200,204)).
	if !g.Event(ctx, uitest.Click(201, 100)) {
		t.Fatal("press on divider not consumed")
	}
	// Drag 80px to the right: dRatio = 80/400 = 0.2.
	g.Event(ctx, uitest.MouseDrag(281, 100))

	ratios := g.Ratios()
	if !approx(ratios[0], 0.7) || !approx(ratios[1], 0.3) {
		t.Fatalf("after drag: ratios = %v want [0.7 0.3]", ratios)
	}

	g.Event(ctx, uitest.Release(281, 100))
	// A second layout reflects the new ratios.
	uitest.LayoutWidgetTight(g, 404, 200)
	left := g.Children()[0].(interface{ Bounds() geometry.Rect })
	if !approx(left.Bounds().Width(), 0.7*400) {
		t.Fatalf("left panel width after drag: got %v want %v", left.Bounds().Width(), 0.7*400)
	}
}

// TestResizableDragClamp ensures dragging past the edge clamps without
// producing negative panel sizes.
func TestResizableDragClamp(t *testing.T) {
	lightTokens(t)

	g := graft.ResizablePanelGroup(graft.ResizableHorizontal,
		graft.ResizablePanel(primitives.Box().Width(10).Height(10)),
		graft.ResizableHandle(),
		graft.ResizablePanel(primitives.Box().Width(10).Height(10)),
	)
	uitest.LayoutWidgetTight(g, 404, 200)
	ctx := uitest.NewMockContext()

	g.Event(ctx, uitest.Click(201, 100))
	// Drag far left, well past the start.
	g.Event(ctx, uitest.MouseDrag(-500, 100))
	g.Event(ctx, uitest.Release(-500, 100))

	ratios := g.Ratios()
	if ratios[0] < 0 || ratios[1] < 0 {
		t.Fatalf("clamp failed: ratios = %v", ratios)
	}
	if !approx(ratios[0]+ratios[1], 1) {
		t.Fatalf("ratio pair total not preserved: %v", ratios)
	}
}

// TestGoldenResizable renders settled two-panel horizontal, vertical, and
// with-grip variants in light + dark.
func TestGoldenResizable(t *testing.T) {
	gtest.GoldenLightDark(t, "resizable-horizontal", func() widget.Widget {
		g := graft.ResizablePanelGroup(graft.ResizableHorizontal,
			graft.ResizablePanel(resizablePanelBox("One")),
			graft.ResizableHandle(),
			graft.ResizablePanel(resizablePanelBox("Two")),
		)
		uitest.LayoutWidgetTight(g, 360, 120)
		return primitives.Box(g).Width(360).Height(120)
	})

	gtest.GoldenLightDark(t, "resizable-vertical", func() widget.Widget {
		g := graft.ResizablePanelGroup(graft.ResizableVertical,
			graft.ResizablePanel(resizablePanelBox("Header")),
			graft.ResizableHandle(),
			graft.ResizablePanel(resizablePanelBox("Body")),
		)
		uitest.LayoutWidgetTight(g, 280, 200)
		return primitives.Box(g).Width(280).Height(200)
	})

	gtest.GoldenLightDark(t, "resizable-with-grip", func() widget.Widget {
		g := graft.ResizablePanelGroup(graft.ResizableHorizontal,
			graft.ResizablePanel(resizablePanelBox("Left")),
			graft.ResizableHandle().WithHandle(),
			graft.ResizablePanel(resizablePanelBox("Right")),
		)
		uitest.LayoutWidgetTight(g, 360, 120)
		return primitives.Box(g).Width(360).Height(120)
	})
}

package graft_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// drawerDemoContent builds the canonical demo body used by spec tests and
// goldens: a header (title + description) and a footer button.
func drawerDemoContent() *graft.DrawerContentWidget {
	return graft.DrawerContent(
		graft.DrawerHeader(
			graft.DrawerTitle("Move Goal"),
			graft.DrawerDescription("Set your daily activity goal."),
		),
		graft.DrawerFooter(
			ovButton("Submit", true),
		),
	)
}

// TestDrawerContentSurface pins the bottom panel surface: bg Background,
// rounded TOP corners (RadiusLG), shadow-LG, a single 1px Border on the top
// edge, and the grabber handle.
func TestDrawerContentSurface(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	tok := th.Active()

	content := drawerDemoContent() // default side = bottom
	size := content.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(640, 900)))
	if size.Width != 640 {
		t.Fatalf("bottom panel width: got %v want full 640", size.Width)
	}

	content.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	canvas := uitest.DrawWidget(content)

	// Surface fill: bg Background, rounded with RadiusLG (the rounded-t look is
	// composited via DrawRoundRect + square overpaint, so the first round-rect
	// fill is the panel surface).
	var surface *uitest.DrawRoundRectCall
	for i := range canvas.RoundRects {
		if canvas.RoundRects[i].Color == tok.Background {
			surface = &canvas.RoundRects[i]
			break
		}
	}
	if surface == nil {
		t.Fatal("no rounded Background surface for bottom drawer")
	}
	if surface.Radius != th.RadiusLG() {
		t.Errorf("panel radius: got %v want RadiusLG %v", surface.Radius, th.RadiusLG())
	}

	// Inner-edge border: a 1px-tall Border-colored line spanning the full width
	// at the top edge (border-t).
	var border *uitest.DrawRectCall
	for i := range canvas.Rects {
		r := &canvas.Rects[i]
		if r.Color == tok.Border && r.Bounds.Height() == metrics.DrawerBorderWidth &&
			r.Bounds.Width() == size.Width {
			border = r
			break
		}
	}
	if border == nil {
		t.Fatal("no 1px top inner-edge border line")
	}
	if border.Bounds.Min.Y != 0 {
		t.Errorf("bottom drawer border-t should sit at y=0, got %v", border.Bounds.Min.Y)
	}

	// Grabber handle: a muted, fully-rounded pill of the spec'd size, centered.
	var handle *uitest.DrawRoundRectCall
	for i := range canvas.RoundRects {
		r := &canvas.RoundRects[i]
		if r.Color == tok.Muted && r.Bounds.Width() == metrics.DrawerHandleWidth &&
			r.Bounds.Height() == metrics.DrawerHandleHeight {
			handle = r
			break
		}
	}
	if handle == nil {
		t.Fatal("no grabber handle pill")
	}
	wantX := (size.Width - metrics.DrawerHandleWidth) / 2
	if handle.Bounds.Min.X != wantX {
		t.Errorf("handle X: got %v want centered %v", handle.Bounds.Min.X, wantX)
	}
	if handle.Bounds.Min.Y != metrics.DrawerHandleTopInset {
		t.Errorf("handle Y: got %v want mt-4 %v", handle.Bounds.Min.Y, metrics.DrawerHandleTopInset)
	}
}

// TestDrawerDefaultSideIsBottom verifies an unconfigured DrawerContent anchors
// to the bottom edge (distinct from Sheet's right default).
func TestDrawerDefaultSideIsBottom(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	ctx := uitest.NewMockContext()
	viewport := geometry.Sz(1000, 800)

	content := drawerDemoContent()
	host := graft.Drawer(content).Open(true)

	om := &ovFakeOverlayManager{}
	ctx.OverlayVal = om
	host.Layout(ctx, looseConstraints())
	if len(om.pushed) != 1 {
		t.Fatalf("overlay not pushed")
	}
	om.pushed[0].Layout(ctx, geometry.Tight(viewport))

	b := content.Bounds()
	if b.Width() != viewport.Width {
		t.Errorf("bottom drawer width: got %v want full %v", b.Width(), viewport.Width)
	}
	if b.Max.Y != viewport.Height {
		t.Errorf("bottom drawer should sit flush with the viewport bottom: max.Y %v want %v",
			b.Max.Y, viewport.Height)
	}
}

// TestDrawerSideAnchoring pins the panel origin/shape for each side.
func TestDrawerSideAnchoring(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	ctx := uitest.NewMockContext()
	viewport := geometry.Sz(1000, 800)

	cases := []struct {
		side       graft.DrawerSide
		wantOrigin geometry.Point
		fullHeight bool
		fullWidth  bool
	}{
		{graft.DrawerLeft, geometry.Pt(0, 0), true, false},
		{graft.DrawerRight, geometry.Pt(1000 - metrics.DrawerMaxWidth, 0), true, false},
		{graft.DrawerTop, geometry.Pt(0, 0), false, true},
	}
	for _, tc := range cases {
		content := drawerDemoContent().Side(tc.side)
		host := graft.Drawer(content).Open(true)

		om := &ovFakeOverlayManager{}
		ctx.OverlayVal = om
		host.Layout(ctx, looseConstraints())
		if len(om.pushed) != 1 {
			t.Fatalf("side %d: overlay not pushed", tc.side)
		}
		om.pushed[0].Layout(ctx, geometry.Tight(viewport))

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
		if !tc.fullWidth && b.Width() != metrics.DrawerMaxWidth {
			t.Errorf("side %d: width got %v want %v", tc.side, b.Width(), metrics.DrawerMaxWidth)
		}
	}
}

// TestDrawerNonBottomHidesHandle verifies the grabber is bottom-only.
func TestDrawerNonBottomHidesHandle(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	tok := graft.CurrentTheme().Active()

	content := drawerDemoContent().Side(graft.DrawerRight)
	size := content.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(384, 800)))
	content.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	canvas := uitest.DrawWidget(content)

	for _, r := range canvas.RoundRects {
		if r.Color == tok.Muted && r.Bounds.Width() == metrics.DrawerHandleWidth {
			t.Fatal("right drawer should not draw a grabber handle")
		}
	}
}

// TestDrawerHostShowHide drives the host's open signal against a fake overlay
// manager: setting open pushes a modal overlay; clearing it removes it.
func TestDrawerHostShowHide(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	open := state.NewSignal(false)
	host := graft.Drawer(drawerDemoContent()).Bind(open)

	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om

	host.Layout(ctx, looseConstraints())
	if om.liveCount() != 0 {
		t.Fatalf("drawer shown while closed: live=%d", om.liveCount())
	}

	open.Set(true)
	host.Layout(ctx, looseConstraints())
	if om.liveCount() != 1 {
		t.Fatalf("drawer not shown after open: live=%d", om.liveCount())
	}

	open.Set(false)
	host.Layout(ctx, looseConstraints())
	if om.liveCount() != 0 {
		t.Fatalf("drawer not hidden after close: live=%d", om.liveCount())
	}
}

// TestDrawerTriggerOpens verifies the trigger flips the open signal on click.
func TestDrawerTriggerOpens(t *testing.T) {
	defer ovForceLightMode(t)()
	open := state.NewSignal(false)
	trig := graft.DrawerTrigger(primitives.Box().Width(80).Height(36), open)
	trig.Layout(uitest.NewMockContext(), looseConstraints())

	if uitest.SimulateClick(trig, 10, 10); !open.Get() {
		t.Fatal("trigger click did not open the drawer")
	}
}

// TestDrawerBackdropDismiss verifies a press outside the panel closes it.
func TestDrawerBackdropDismiss(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	open := state.NewSignal(true)
	host := graft.Drawer(drawerDemoContent()).Bind(open)

	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om
	host.Layout(ctx, looseConstraints())
	ov := om.pushed[0]
	ov.Layout(ctx, geometry.Tight(geometry.Sz(800, 600)))

	// Press near the top of the viewport, well above a bottom drawer panel.
	press := &event.MouseEvent{MouseType: event.MousePress, Button: event.ButtonLeft,
		Position: geometry.Pt(400, 20)}
	ov.Event(ctx, press)

	if open.Get() {
		t.Fatal("backdrop press did not dismiss the drawer")
	}
}

// TestDrawerDragDismiss verifies a downward drag past the threshold on a bottom
// drawer dismisses it, while a short drag snaps back (stays open).
func TestDrawerDragDismiss(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	drag := func(dy float32) bool {
		open := state.NewSignal(true)
		host := graft.Drawer(drawerDemoContent()).Bind(open)
		om := &ovFakeOverlayManager{}
		ctx := uitest.NewMockContext()
		ctx.OverlayVal = om
		host.Layout(ctx, looseConstraints())
		ov := om.pushed[0]
		ov.Layout(ctx, geometry.Tight(geometry.Sz(800, 600)))

		start := geometry.Pt(400, 560) // inside the bottom panel near its top
		ov.Event(ctx, &event.MouseEvent{MouseType: event.MousePress, Button: event.ButtonLeft, Position: start})
		ov.Event(ctx, &event.MouseEvent{MouseType: event.MouseDrag, Button: event.ButtonLeft,
			Position: geometry.Pt(start.X, start.Y+dy)})
		ov.Event(ctx, &event.MouseEvent{MouseType: event.MouseRelease, Button: event.ButtonLeft,
			Position: geometry.Pt(start.X, start.Y+dy)})
		return open.Get()
	}

	// A tiny drag snaps back: still open.
	if !drag(5) {
		t.Error("short drag should snap back and keep the drawer open")
	}
	// A long downward drag past the panel-fraction threshold dismisses.
	if drag(400) {
		t.Error("long downward drag should dismiss the drawer")
	}
}

// TestDrawerEscDismiss verifies Esc closes the drawer.
func TestDrawerEscDismiss(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	open := state.NewSignal(true)
	host := graft.Drawer(drawerDemoContent()).Bind(open)

	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om
	host.Layout(ctx, looseConstraints())
	ov := om.pushed[0]
	ov.Layout(ctx, geometry.Tight(geometry.Sz(800, 600)))

	ov.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEscape})
	if open.Get() {
		t.Fatal("Esc did not dismiss the drawer")
	}
}

// TestGoldenDrawer renders the drawer exactly as it appears at runtime — a
// modal frame with the backdrop and the panel anchored to its edge — via
// DrawerPreview. Goldens render the SETTLED open state (slide offset zero),
// light + dark.
func TestGoldenDrawer(t *testing.T) {
	build := func(side graft.DrawerSide) func() widget.Widget {
		return func() widget.Widget {
			return graft.DrawerPreview(drawerDemoContent().Side(side), geometry.Sz(640, 420))
		}
	}
	gtest.GoldenLightDark(t, "drawer-bottom", build(graft.DrawerBottom))
	gtest.GoldenLightDark(t, "drawer-right", build(graft.DrawerRight))
}

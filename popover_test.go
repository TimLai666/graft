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

// fakeOverlayManager records overlay pushes/removals for spec tests
// (uitest.MockContext.OverlayVal is nil by default; this fills the gap).
type fakeOverlayManager struct {
	pushed    []widget.Widget
	onDismiss map[widget.Widget]func()
	removed   []widget.Widget
}

func newFakeOverlayManager() *fakeOverlayManager {
	return &fakeOverlayManager{onDismiss: map[widget.Widget]func(){}}
}

func (f *fakeOverlayManager) PushOverlay(w widget.Widget, onDismiss func()) {
	f.pushed = append(f.pushed, w)
	f.onDismiss[w] = onDismiss
}

func (f *fakeOverlayManager) PopOverlay() {
	if len(f.pushed) == 0 {
		return
	}
	f.removed = append(f.removed, f.pushed[len(f.pushed)-1])
	f.pushed = f.pushed[:len(f.pushed)-1]
}

func (f *fakeOverlayManager) RemoveOverlay(w widget.Widget) {
	for i, p := range f.pushed {
		if p == w {
			f.pushed = append(f.pushed[:i], f.pushed[i+1:]...)
			f.removed = append(f.removed, w)
			return
		}
	}
}

func (f *fakeOverlayManager) has(w widget.Widget) bool {
	for _, p := range f.pushed {
		if p == w {
			return true
		}
	}
	return false
}

var _ widget.OverlayManager = (*fakeOverlayManager)(nil)

// TestPopoverContentGeometry pins the shadcn popover surface: w-72 (288),
// p-4 (16), rounded-md (8), 1px border, bg-popover, shadow-md.
func TestPopoverContentGeometry(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	content := graft.PopoverContent(
		graft.H4("Dimensions"),
		graft.MutedText("Set the dimensions for the layer."),
	)
	size := uitest.LayoutWidget(content, 900, 900)
	if size.Width != metrics.Popover.Width {
		t.Errorf("width = %v, want %v (w-72)", size.Width, metrics.Popover.Width)
	}
	// p-4 + H4 line (28) + muted line (20) + p-4.
	if want := float32(16 + 28 + 20 + 16); size.Height != want {
		t.Errorf("height = %v, want %v", size.Height, want)
	}

	mc := uitest.DrawWidget(content)
	tok := graft.CurrentTheme().Active()
	th := graft.CurrentTheme()

	// BorderFill: outer border round-rect (full bounds) then inner bg-popover
	// fill (inset by the 1px border). shadow-md paints before both.
	bw := metrics.Popover.BorderWidth

	// Border round-rect = full bounds / full radius, in the border token,
	// right after the shadow layers.
	border := mc.RoundRects[len(metrics.ShadowMD)]
	if border.Color != tok.Border {
		t.Errorf("border fill = %+v, want border token %v", border, tok.Border)
	}
	if border.Radius != th.RadiusLG() ||
		border.Bounds != geometry.NewRect(0, 0, size.Width, size.Height) {
		t.Errorf("border fill = %+v, want radius %v full bounds", border, th.RadiusLG())
	}

	// Inner bg-popover fill = inset by the border width, radius clamped by it.
	bg := mc.RoundRects[len(metrics.ShadowMD)+1]
	if bg.Color != tok.Popover {
		t.Errorf("bg fill = %+v, want bg-popover %v", bg, tok.Popover)
	}
	if bg.Radius != th.RadiusLG()-bw ||
		bg.Bounds != geometry.NewRect(bw, bw, size.Width-2*bw, size.Height-2*bw) {
		t.Errorf("bg fill = %+v, want radius %v inset by %v", bg, th.RadiusLG()-bw, bw)
	}

	// No border stroke any more (border is now a fill).
	if len(mc.StrokeRoundRects) != 0 {
		t.Fatalf("strokes = %d, want 0 (border now a fill)", len(mc.StrokeRoundRects))
	}

	// First child starts at the p-4 inset.
	first := content.Children()[0]
	if b, ok := first.(interface{ Bounds() geometry.Rect }); ok {
		if b.Bounds().Min != geometry.Pt(16, 16) {
			t.Errorf("first child at %v, want (16,16)", b.Bounds().Min)
		}
	}
}

// TestPopoverOpenCloseMachine drives the trigger -> overlay -> dismiss
// cycle against a fake overlay manager.
func TestPopoverOpenCloseMachine(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	open := state.NewSignal(false)
	trigger := primitives.Box().Width(300).Height(36)
	content := graft.PopoverContent(graft.Text("Place content for the popover here."))

	var changes []bool
	pop := graft.Popover(graft.PopoverTrigger(trigger, open), content).
		Bind(open).
		OnOpenChange(func(v bool) { changes = append(changes, v) })

	ctx := uitest.NewMockContext()
	om := newFakeOverlayManager()
	ctx.OverlayVal = om

	uitest.LayoutWidget(pop, 800, 600)
	uitest.DrawWidgetWithContext(pop, ctx) // stamps screen origins; closed
	if len(om.pushed) != 0 {
		t.Fatal("popover pushed an overlay while closed")
	}

	// Click the trigger: the signal flips and the content is pushed,
	// positioned bottom-center with sideOffset 4.
	uitest.SimulateClickWithContext(pop, ctx, 10, 10)
	if !open.Get() {
		t.Fatal("trigger click did not open the popover")
	}
	if !om.has(content) {
		t.Fatal("open popover did not push its content overlay")
	}
	wantPos := geometry.Pt(150-metrics.Popover.Width/2, 36+metrics.Popover.SideOffset)
	if content.Bounds().Min != wantPos {
		t.Errorf("content position = %v, want %v", content.Bounds().Min, wantPos)
	}
	if content.Bounds().Width() != metrics.Popover.Width {
		t.Errorf("content width = %v, want %v", content.Bounds().Width(), metrics.Popover.Width)
	}

	// Simulate the overlay container dismissing (click outside / Escape).
	om.onDismiss[content]()
	if open.Get() {
		t.Fatal("dismiss did not close the open signal")
	}
	if om.has(content) {
		t.Fatal("dismiss did not remove the overlay")
	}

	// Controlled close: signal write removes the overlay on next draw.
	open.Set(true)
	uitest.DrawWidgetWithContext(pop, ctx)
	if !om.has(content) {
		t.Fatal("signal open did not push the overlay")
	}
	open.Set(false)
	uitest.DrawWidgetWithContext(pop, ctx)
	if om.has(content) {
		t.Fatal("signal close did not remove the overlay")
	}

	// OnOpenChange fired for the interaction transitions only (click
	// open + dismiss close); controlled writes do not refire it.
	if len(changes) != 2 || changes[0] != true || changes[1] != false {
		t.Errorf("OnOpenChange calls = %v, want [true false]", changes)
	}
}

func TestPopoverContentWidthOverride(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	content := graft.PopoverContent(graft.Text("hi")).W(200)
	size := uitest.LayoutWidget(content, 900, 900)
	if size.Width != 200 {
		t.Errorf("width = %v, want 200", size.Width)
	}
}

func TestGoldenPopover(t *testing.T) {
	gtest.GoldenLightDark(t, "popover-content", func() widget.Widget {
		content := graft.PopoverContent(
			primitives.VBox(
				graft.Small("Dimensions"),
				graft.MutedText("Set the dimensions for the layer."),
			).Gap(8),
		)
		return primitives.Box(content).Padding(32)
	})
}

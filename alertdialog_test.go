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
)

// dispatchEsc sends an Escape key press to the widget.
func dispatchEsc(w widget.Widget, ctx widget.Context) bool {
	return w.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, event.ModNone))
}

// TestAlertDialogNoBackdropDismiss verifies that the alert overlay ignores a
// backdrop click but still closes on Escape.
func TestAlertDialogNoBackdropDismiss(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	open := state.NewSignal(false)
	content := graft.AlertDialogContent(
		graft.AlertDialogHeader(
			graft.AlertDialogTitle("Are you absolutely sure?"),
			graft.AlertDialogDescription("This action cannot be undone."),
		),
	)
	host := graft.AlertDialog(content).Bind(open)

	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om
	ctx.WindowSizeVal = geometry.Sz(800, 600)

	host.Layout(ctx, looseConstraints())
	open.Set(true)
	host.Layout(ctx, looseConstraints())

	if len(om.pushed) != 1 {
		t.Fatalf("alert overlay not pushed: pushed=%d", len(om.pushed))
	}
	overlay := om.pushed[0]
	overlay.Layout(ctx, geometry.Tight(geometry.Sz(800, 600)))

	// Backdrop click far from the centered card → must NOT close.
	overlay.Event(ctx, event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(5, 5), geometry.Pt(5, 5), event.ModNone))
	if !open.Get() {
		t.Fatal("alert dialog closed on backdrop click (must not)")
	}

	// Escape → must close.
	dispatchEsc(overlay, ctx)
	if open.Get() {
		t.Fatal("alert dialog did not close on Escape")
	}
}

// TestDialogBackdropDismiss verifies the plain Dialog DOES close on backdrop
// click (the contrast with AlertDialog).
func TestDialogBackdropDismiss(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	open := state.NewSignal(false)
	host := graft.Dialog(graft.DialogContent(graft.DialogTitle("Hi"))).Bind(open)

	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om
	ctx.WindowSizeVal = geometry.Sz(800, 600)

	host.Layout(ctx, looseConstraints())
	open.Set(true)
	host.Layout(ctx, looseConstraints())

	overlay := om.pushed[0]
	overlay.Layout(ctx, geometry.Tight(geometry.Sz(800, 600)))
	overlay.Event(ctx, event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(5, 5), geometry.Pt(5, 5), event.ModNone))
	if open.Get() {
		t.Fatal("dialog did not close on backdrop click")
	}
}

// TestAlertDialogButtons verifies Action/Cancel fire their callbacks on click.
func TestAlertDialogButtons(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	var actioned, cancelled bool
	action := graft.AlertDialogAction("Continue", func() { actioned = true })
	cancel := graft.AlertDialogCancel("Cancel", func() { cancelled = true })

	action.Layout(uitest.NewMockContext(), looseConstraints())
	cancel.Layout(uitest.NewMockContext(), looseConstraints())

	uitest.SimulateClick(action, 10, 10)
	uitest.SimulateClick(cancel, 10, 10)
	if !actioned {
		t.Error("AlertDialogAction did not fire onClick")
	}
	if !cancelled {
		t.Error("AlertDialogCancel did not fire onClick")
	}
}

// TestGoldenAlertDialog renders the alert card with action+cancel footer,
// light + dark.
func TestGoldenAlertDialog(t *testing.T) {
	gtest.GoldenLightDark(t, "alert-dialog-content", func() widget.Widget {
		content := graft.AlertDialogContent(
			graft.AlertDialogHeader(
				graft.AlertDialogTitle("Are you absolutely sure?"),
				graft.AlertDialogDescription("This action cannot be undone. This will permanently delete your account."),
			),
			graft.AlertDialogFooter(
				graft.AlertDialogCancel("Cancel", nil),
				graft.AlertDialogAction("Continue", nil),
			),
		)
		return primitives.VBox(content).Padding(24)
	})
}

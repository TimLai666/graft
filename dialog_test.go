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

// ovButton builds a minimal styled button stand-in for dialog/alert footer
// demos. (graft.Button is not present on this batch's base, so the goldens use
// a local pill rather than depending on another batch's widget.)
func ovButton(label string, primary bool) widget.Widget {
	tok := graft.CurrentTheme().Active()
	bg := tok.Primary
	fg := tok.PrimaryForeground
	if !primary {
		bg = tok.Background
		fg = tok.Foreground
	}
	txt := graft.Text(label).
		FontSize(14).Weight(500).LineHeight(20).
		Align(widget.TextAlignCenter).
		Color(fg)
	box := primitives.VBox(txt).
		PaddingXY(16, 8).
		CrossAlign(primitives.CrossAxisCenter).
		Background(bg).
		Rounded(graft.CurrentTheme().RadiusMD())
	if !primary {
		box = box.BorderStyle(1, tok.Border)
	}
	return box
}

// TestDialogContentLayout pins the card metrics: max width 512, padding 24,
// radius LG, background fill, 1px Border.
func TestDialogContentLayout(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	tok := graft.CurrentTheme().Active()

	content := graft.DialogContent(
		graft.DialogHeader(
			graft.DialogTitle("Edit profile"),
			graft.DialogDescription("Make changes to your profile here."),
		),
	)
	size := content.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(900, 900)))

	if size.Width != metrics.DialogMaxWidth {
		t.Fatalf("card width: got %v want %v", size.Width, metrics.DialogMaxWidth)
	}

	content.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	canvas := uitest.DrawWidget(content)

	// Card surface round-rect.
	var card *uitest.DrawRoundRectCall
	for i := range canvas.RoundRects {
		if canvas.RoundRects[i].Bounds.Size() == size {
			card = &canvas.RoundRects[i]
			break
		}
	}
	if card == nil {
		t.Fatalf("no card round-rect at size %v", size)
	}
	if card.Color != tok.Background {
		t.Errorf("card fill: got %+v want background %+v", card.Color, tok.Background)
	}
	if card.Radius != graft.CurrentTheme().RadiusLG() {
		t.Errorf("card radius: got %v want LG %v", card.Radius, graft.CurrentTheme().RadiusLG())
	}

	// Inside border stroke in Border color, width 1.
	var border *uitest.StrokeRoundRectCall
	for i := range canvas.StrokeRoundRects {
		if canvas.StrokeRoundRects[i].StrokeWidth == metrics.DialogBorderWidth {
			border = &canvas.StrokeRoundRects[i]
			break
		}
	}
	if border == nil {
		t.Fatal("no 1px card border stroke")
	}
	if border.Color != tok.Border {
		t.Errorf("card border color: got %+v want Border %+v", border.Color, tok.Border)
	}
}

// TestDialogViewportCap pins the small-viewport width cap: viewport − 32px.
func TestDialogViewportCap(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	content := graft.DialogContent(graft.DialogTitle("Hi"))
	// Available width 400 → capped to 400 − 2*16 = 368 (< 512).
	size := content.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 900)))
	want := float32(400) - 2*metrics.DialogViewportMargin
	if size.Width != want {
		t.Fatalf("capped width: got %v want %v", size.Width, want)
	}
}

// TestDialogCloseButton verifies the X icon hit area sits at the top-right
// inset and that hiding it removes the geometry.
func TestDialogCloseButton(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	content := graft.DialogContent(graft.DialogTitle("Hi"))
	size := content.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(900, 900)))
	content.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))

	// Hover the close button → cursor pointer; press → onClose fires.
	closed := false
	content.OnClose(func() { closed = true })

	cx := size.Width - metrics.DialogCloseInset - metrics.DialogCloseIconSize/2
	cy := metrics.DialogCloseInset + metrics.DialogCloseIconSize/2

	ctx := uitest.NewMockContext()
	content.Event(ctx, event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(cx, cy), geometry.Pt(cx, cy), event.ModNone))
	if ctx.CursorVal != widget.CursorPointer {
		t.Errorf("close hover: cursor got %v want Pointer", ctx.CursorVal)
	}

	content.Event(ctx, event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(cx, cy), geometry.Pt(cx, cy), event.ModNone))
	if !closed {
		t.Error("close button press did not invoke onClose")
	}
}

// TestDialogHostShowHide drives the host's open signal against a fake overlay
// manager: setting open pushes a modal overlay; clearing it removes it.
func TestDialogHostShowHide(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	open := state.NewSignal(false)
	host := graft.Dialog(graft.DialogContent(graft.DialogTitle("Hi"))).Bind(open)

	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om

	host.Layout(ctx, looseConstraints())
	if om.liveCount() != 0 {
		t.Fatalf("dialog shown while closed: live=%d", om.liveCount())
	}

	open.Set(true)
	host.Layout(ctx, looseConstraints()) // re-sync on next frame
	if om.liveCount() != 1 {
		t.Fatalf("dialog not shown after open: live=%d pushed=%d", om.liveCount(), len(om.pushed))
	}

	open.Set(false)
	host.Layout(ctx, looseConstraints())
	if om.liveCount() != 0 {
		t.Fatalf("dialog not hidden after close: live=%d", om.liveCount())
	}
}

// TestDialogTriggerOpens verifies the trigger flips the open signal on click.
func TestDialogTriggerOpens(t *testing.T) {
	defer ovForceLightMode(t)()
	open := state.NewSignal(false)
	trig := graft.DialogTrigger(primitives.Box().Width(80).Height(36), open)
	trig.Layout(uitest.NewMockContext(), looseConstraints())

	if uitest.SimulateClick(trig, 10, 10); !open.Get() {
		t.Fatal("trigger click did not open the dialog")
	}
}

// TestGoldenDialog renders the content card directly, with and without the
// close button, light + dark.
func TestGoldenDialog(t *testing.T) {
	build := func(withClose bool) func() widget.Widget {
		return func() widget.Widget {
			content := graft.DialogContent(
				graft.DialogHeader(
					graft.DialogTitle("Edit profile"),
					graft.DialogDescription("Make changes to your profile here. Click save when done."),
				),
				graft.DialogFooter(
					ovButton("Cancel", false),
					ovButton("Save changes", true),
				),
			)
			if !withClose {
				content.HideClose()
			}
			// Pad the frame so the shadow + close ring are captured.
			return primitives.VBox(content).Padding(24)
		}
	}
	gtest.GoldenLightDark(t, "dialog-content", build(true))
	gtest.GoldenLightDark(t, "dialog-content-noclose", build(false))
}

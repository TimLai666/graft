package graft_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
)

// TestClickReachesButtonDirect: a button in a plain VBox must fire OnClick.
func TestClickReachesButtonDirect(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	clicked := false
	btn := graft.Button("Hit me").OnClick(func() { clicked = true })
	root := primitives.VBox(btn).Padding(20)

	uitest.LayoutWidget(root, 400, 200)
	uitest.SimulateClick(root, 60, 38)

	if !clicked {
		t.Fatal("OnClick did NOT fire through plain VBox")
	}
}

// TestClickReachesButtonThroughScrollArea: same button wrapped in ScrollArea
// (the kitchensink root structure).
func TestClickReachesButtonThroughScrollArea(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	clicked := false
	btn := graft.Button("Hit me").OnClick(func() { clicked = true })
	root := graft.ScrollArea(primitives.VBox(btn).Padding(20))

	uitest.LayoutWidget(root, 400, 200)
	uitest.SimulateClick(root, 60, 38)

	if !clicked {
		t.Fatal("OnClick did NOT fire through ScrollArea wrapper")
	}
}

// TestToggleClickThenDrawNoPanic: clicking a Toggle then re-rendering the
// on-state must not panic.
func TestToggleClickThenDrawNoPanic(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	tg := graft.Toggle("Bold")
	root := primitives.VBox(tg).Padding(20)
	uitest.LayoutWidget(root, 400, 200)
	uitest.SimulateClick(root, 50, 38)
	uitest.LayoutWidget(root, 400, 200)
	_ = uitest.DrawWidget(root)
}

// reentrantFocusWidget's SetFocused re-enters the context lock via
// RegisterDirtyBoundary — the exact pattern that deadlocked the live window
// (focusing a widget whose SetFocused called SetNeedsRedraw, which propagates
// to ctx.RegisterDirtyBoundary's RLock while RequestFocus holds the write
// Lock). This documents the failure mode and proves the timeout harness
// detects it; the graft fix is to use MarkRedrawLocal in SetFocused, which
// never touches the context mutex (see button.go / toggle.go / accordion.go
// etc. SetFocused). Verified end-to-end against the real window.
type reentrantFocusWidget struct {
	widget.WidgetBase
	ctx *widget.ContextImpl
}

func (w *reentrantFocusWidget) SetFocused(focused bool) {
	w.WidgetBase.SetFocused(focused)
	w.ctx.RegisterDirtyBoundary(1)
}
func (w *reentrantFocusWidget) IsFocusable() bool { return true }
func (w *reentrantFocusWidget) Layout(widget.Context, geometry.Constraints) geometry.Size {
	return geometry.Sz(0, 0)
}
func (w *reentrantFocusWidget) Draw(widget.Context, widget.Canvas)     {}
func (w *reentrantFocusWidget) Event(widget.Context, event.Event) bool { return false }
func (w *reentrantFocusWidget) Children() []widget.Widget              { return nil }

// TestContextLockReentryDeadlocks documents and pins the deadlock mechanism:
// calling a context-locking method from inside SetFocused (which RequestFocus
// invokes under its write lock) hangs. graft widgets must therefore never call
// SetNeedsRedraw from SetFocused — they use MarkRedrawLocal instead.
func TestContextLockReentryDeadlocks(t *testing.T) {
	ctx := widget.NewContext()
	ctx.SetOnRegisterDirtyBoundary(func(uint64) {})
	w := &reentrantFocusWidget{ctx: ctx}

	done := make(chan struct{})
	go func() {
		ctx.RequestFocus(w) // hangs: RLock under held write Lock
		close(done)
	}()
	select {
	case <-done:
		t.Fatal("expected re-entrant SetFocused to deadlock; the harness is not exercising the lock")
	case <-time.After(time.Second):
		// Deadlocked as expected. The leaked goroutine is acceptable here.
	}
}

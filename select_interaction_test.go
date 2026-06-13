package graft_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/metrics"
)

// recordingOverlayManager is a headless widget.OverlayManager that captures the
// pushed overlay so tests can dispatch events to it (the MockContext's default
// OverlayManager is nil, so the open-menu path is otherwise unreachable).
type recordingOverlayManager struct {
	top       widget.Widget
	onDismiss func()
}

func (o *recordingOverlayManager) PushOverlay(w widget.Widget, onDismiss func()) {
	o.top = w
	o.onDismiss = onDismiss
}

func (o *recordingOverlayManager) PopOverlay() { o.top = nil; o.onDismiss = nil }

func (o *recordingOverlayManager) RemoveOverlay(w widget.Widget) {
	if o.top == w {
		o.top = nil
		o.onDismiss = nil
	}
}

// TestSelectOpenPushesOverlayAndInvalidates verifies clicking the trigger opens
// the menu (pushes an overlay) and requests a repaint frame.
func TestSelectOpenPushesOverlayAndInvalidates(t *testing.T) {
	selectLightTheme(t)
	s := graft.Select(
		graft.SelectItem("apple", "Apple"),
		graft.SelectItem("banana", "Banana"),
	).Placeholder("Pick").W(220)

	om := &recordingOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om

	size := s.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))
	s.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))

	s.Event(ctx, uitest.Click(size.Width/2, size.Height/2))
	if om.top == nil {
		t.Fatal("trigger click did not push the menu overlay")
	}
	if !ctx.Invalidated {
		t.Error("opening the menu did not request a repaint frame")
	}
}

// TestSelectMenuKeyboardNavInvalidates verifies arrow keys in the open menu
// request a repaint frame (the highlight-desync repaint bug).
func TestSelectMenuKeyboardNavInvalidates(t *testing.T) {
	selectLightTheme(t)
	s := graft.Select(
		graft.SelectItem("apple", "Apple"),
		graft.SelectItem("banana", "Banana"),
	).Placeholder("Pick").W(220)

	om := &recordingOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om

	size := s.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))
	s.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	s.Event(ctx, uitest.Click(size.Width/2, size.Height/2))
	if om.top == nil {
		t.Fatal("menu not open")
	}

	ctx.Reset()
	if !om.top.Event(ctx, uitest.KeyPress(event.KeyDown, event.ModNone)) {
		t.Fatal("menu did not consume Down")
	}
	if len(ctx.InvalidatedRects) == 0 && !ctx.Invalidated {
		t.Error("menu Down keypress did not request a repaint frame")
	}
}

// TestSelectMenuClickCommitsAndCloses verifies clicking an item commits the
// bound signal, fires OnChange, and closes the menu (removes the overlay).
func TestSelectMenuClickCommitsAndCloses(t *testing.T) {
	selectLightTheme(t)
	sig := state.NewSignal("")
	var changed string
	s := graft.Select(
		graft.SelectItem("apple", "Apple"),
		graft.SelectItem("banana", "Banana"),
	).Placeholder("Pick").W(220).Bind(sig).OnChange(func(v string) { changed = v })

	om := &recordingOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om

	size := s.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))
	s.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	s.Event(ctx, uitest.Click(size.Width/2, size.Height/2))
	menu := om.top
	if menu == nil {
		t.Fatal("menu not open")
	}

	// Click the second item (Banana). The menu overlay bounds were set by
	// openMenu; rows start at ContentPad and each item is ItemHeight tall.
	mb := menu.(interface{ Bounds() geometry.Rect }).Bounds()
	clickY := mb.Min.Y + metrics.Select.ContentPad + metrics.Select.ItemHeight + metrics.Select.ItemHeight/2
	clickX := mb.Min.X + mb.Width()/2
	menu.Event(ctx, uitest.Click(clickX, clickY))

	if sig.Get() != "banana" {
		t.Errorf("bound signal = %q, want banana", sig.Get())
	}
	if changed != "banana" {
		t.Errorf("OnChange value = %q, want banana", changed)
	}
	if om.top != nil {
		t.Error("menu did not close after committing a selection")
	}
}

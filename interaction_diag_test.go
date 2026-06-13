package graft_test

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"

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
	// Click roughly in the button center (20px padding + ~half button).
	uitest.SimulateClick(root, 60, 38)

	if !clicked {
		t.Fatal("OnClick did NOT fire through plain VBox")
	}
}

// TestClickReachesButtonThroughScrollArea: the same button wrapped in a
// graft.ScrollArea (the kitchensink root structure) must also fire OnClick.
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

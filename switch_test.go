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
	"github.com/TimLai666/graft/theme"
)

// switchForceLight pins the current theme to light for a spec test.
func switchForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

func laidOutSwitch(w widget.Widget) {
	w.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
}

// TestSwitchLayout pins the default and sm track dimensions.
func TestSwitchLayout(t *testing.T) {
	_, restore := switchForceLight(t)
	defer restore()

	def := graft.Switch()
	size := def.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	if size.Width != metrics.Switch.Default.TrackWidth || size.Height != metrics.Switch.Default.TrackHeight {
		t.Fatalf("default size: got %v want %v×%v", size, metrics.Switch.Default.TrackWidth, metrics.Switch.Default.TrackHeight)
	}

	sm := graft.Switch().Sm()
	ssize := sm.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	if ssize.Width != metrics.Switch.SM.TrackWidth || ssize.Height != metrics.Switch.SM.TrackHeight {
		t.Fatalf("sm size: got %v want %v×%v", ssize, metrics.Switch.SM.TrackWidth, metrics.Switch.SM.TrackHeight)
	}
}

// TestSwitchSpecOffLight asserts an off switch has an Input track and a
// Background thumb at the left.
func TestSwitchSpecOffLight(t *testing.T) {
	tok, restore := switchForceLight(t)
	defer restore()

	s := graft.Switch()
	laidOutSwitch(s)
	canvas := uitest.DrawWidget(s)

	var track *uitest.DrawRoundRectCall
	for idx := range canvas.RoundRects {
		track = &canvas.RoundRects[idx] // last round rect = track (after shadow layers)
	}
	if track == nil || track.Color != tok.Input {
		t.Fatalf("off track should be Input, got %+v", track)
	}
	if track.Radius != metrics.Switch.Default.TrackHeight/2 {
		t.Fatalf("track radius: got %v want %v", track.Radius, metrics.Switch.Default.TrackHeight/2)
	}

	if len(canvas.Circles) != 1 {
		t.Fatalf("want 1 thumb circle, got %d", len(canvas.Circles))
	}
	thumb := canvas.Circles[0]
	if thumb.Color != tok.Background {
		t.Fatalf("thumb color: got %+v want Background %+v", thumb.Color, tok.Background)
	}
	if thumb.Radius != metrics.Switch.Default.ThumbSize/2 {
		t.Fatalf("thumb radius: got %v want %v", thumb.Radius, metrics.Switch.Default.ThumbSize/2)
	}
	// Off: thumb near the left edge.
	inset := (metrics.Switch.Default.TrackHeight - metrics.Switch.Default.ThumbSize) / 2
	wantX := track.Bounds.Min.X + inset + metrics.Switch.Default.ThumbSize/2
	if thumb.Center.X != wantX {
		t.Fatalf("off thumb x: got %v want %v", thumb.Center.X, wantX)
	}
}

// TestSwitchSpecOnLight asserts an on switch has a Primary track and the thumb
// translated right by Travel.
func TestSwitchSpecOnLight(t *testing.T) {
	tok, restore := switchForceLight(t)
	defer restore()

	s := graft.Switch().Checked(true)
	laidOutSwitch(s)
	canvas := uitest.DrawWidget(s)

	var track *uitest.DrawRoundRectCall
	for idx := range canvas.RoundRects {
		track = &canvas.RoundRects[idx]
	}
	if track == nil || track.Color != tok.Primary {
		t.Fatalf("on track should be Primary, got %+v", track)
	}

	thumb := canvas.Circles[0]
	inset := (metrics.Switch.Default.TrackHeight - metrics.Switch.Default.ThumbSize) / 2
	wantX := track.Bounds.Min.X + inset + metrics.Switch.Default.Travel + metrics.Switch.Default.ThumbSize/2
	if thumb.Center.X != wantX {
		t.Fatalf("on thumb x: got %v want %v (travel %v)", thumb.Center.X, wantX, metrics.Switch.Default.Travel)
	}
}

// TestSwitchSpecFocusRing asserts the focus ring on the track (radius full).
func TestSwitchSpecFocusRing(t *testing.T) {
	tok, restore := switchForceLight(t)
	defer restore()

	s := graft.Switch()
	laidOutSwitch(s)
	s.SetFocused(true)
	canvas := uitest.DrawWidget(s)

	var ring *uitest.StrokeRoundRectCall
	for idx := range canvas.StrokeRoundRects {
		if canvas.StrokeRoundRects[idx].StrokeWidth == metrics.RingWidth {
			ring = &canvas.StrokeRoundRects[idx]
		}
	}
	if ring == nil {
		t.Fatal("no focus ring stroke (width 3)")
	}
	wantRing := tok.Ring
	wantRing.A = metrics.RingAlpha
	if ring.Color != wantRing {
		t.Fatalf("ring color: got %+v want Ring@0.5 %+v", ring.Color, wantRing)
	}
}

// TestSwitchToggleClick verifies a click toggles and fires OnChange.
func TestSwitchToggleClick(t *testing.T) {
	_, restore := switchForceLight(t)
	defer restore()

	var got []bool
	s := graft.Switch().OnChange(func(v bool) { got = append(got, v) })
	laidOutSwitch(s)
	uitest.SimulateClick(s, 16, 9)
	if !s.IsChecked() {
		t.Fatal("click should turn the switch on")
	}
	uitest.SimulateClick(s, 16, 9)
	if s.IsChecked() {
		t.Fatal("second click should turn it off")
	}
	if len(got) != 2 || got[0] != true || got[1] != false {
		t.Fatalf("OnChange: got %v want [true false]", got)
	}
}

// TestSwitchSpaceToggle verifies Space toggles when focused.
func TestSwitchSpaceToggle(t *testing.T) {
	_, restore := switchForceLight(t)
	defer restore()

	s := graft.Switch()
	laidOutSwitch(s)
	s.SetFocused(true)
	ctx := uitest.NewMockContext()
	if !s.Event(ctx, uitest.KeyPress(event.KeySpace, event.ModNone)) {
		t.Fatal("Space should be consumed")
	}
	if !s.IsChecked() {
		t.Fatal("Space should turn the switch on")
	}
}

// TestSwitchBindControlled verifies controlled state.
func TestSwitchBindControlled(t *testing.T) {
	_, restore := switchForceLight(t)
	defer restore()

	sig := state.NewSignal(false)
	s := graft.Switch().Bind(sig)
	laidOutSwitch(s)
	uitest.SimulateClick(s, 16, 9)
	if !sig.Get() {
		t.Fatal("click should write true to bound signal")
	}
}

// TestSwitchDisabled pins disabled switches out of focus + interaction.
func TestSwitchDisabled(t *testing.T) {
	_, restore := switchForceLight(t)
	defer restore()

	s := graft.Switch().Disabled(true)
	laidOutSwitch(s)
	if s.IsFocusable() {
		t.Fatal("disabled switch should not be focusable")
	}
	uitest.SimulateClick(s, 16, 9)
	if s.IsChecked() {
		t.Fatal("disabled switch should ignore clicks")
	}
}

// TestGoldenSwitch renders the switch states in light and dark.
func TestGoldenSwitch(t *testing.T) {
	gtest.GoldenLightDark(t, "switch-off", func() widget.Widget {
		return switchWrap(graft.Switch())
	})
	gtest.GoldenLightDark(t, "switch-on", func() widget.Widget {
		return switchWrap(graft.Switch().Checked(true))
	})
	gtest.GoldenLightDark(t, "switch-sm-off", func() widget.Widget {
		return switchWrap(graft.Switch().Sm())
	})
	gtest.GoldenLightDark(t, "switch-sm-on", func() widget.Widget {
		return switchWrap(graft.Switch().Sm().Checked(true))
	})
	gtest.GoldenLightDark(t, "switch-disabled", func() widget.Widget {
		return switchWrap(graft.Switch().Checked(true).Disabled(true))
	})
	gtest.GoldenLightDark(t, "switch-focused", func() widget.Widget {
		s := graft.Switch().Checked(true)
		s.SetFocused(true)
		return switchWrap(s)
	})
}

func switchWrap(w widget.Widget) widget.Widget {
	return primitives.VBox(w).Padding(12)
}

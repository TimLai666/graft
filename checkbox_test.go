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

// checkboxForceLight pins the current theme to light for a spec test.
func checkboxForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

func laidOutCheckbox(w widget.Widget) {
	w.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
}

// TestCheckboxLayout pins the 16×16 box and label row width.
func TestCheckboxLayout(t *testing.T) {
	_, restore := checkboxForceLight(t)
	defer restore()

	plain := graft.Checkbox()
	size := plain.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	if size.Width != metrics.Checkbox.Size || size.Height != metrics.Checkbox.Size {
		t.Fatalf("plain size: got %v want %v×%v", size, metrics.Checkbox.Size, metrics.Checkbox.Size)
	}

	labeled := graft.Checkbox().Label("Accept terms")
	lsize := labeled.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 400)))
	if lsize.Height != metrics.Checkbox.Size {
		t.Fatalf("labeled height: got %v want %v", lsize.Height, metrics.Checkbox.Size)
	}
	if lsize.Width <= metrics.Checkbox.Size+metrics.Checkbox.LabelGap {
		t.Fatalf("labeled width should include label text: got %v", lsize.Width)
	}
}

// TestCheckboxSpecUncheckedLight asserts a 16×16 box, radius 4, 1px Input
// border via InsideBorder, and no fill in light mode.
func TestCheckboxSpecUncheckedLight(t *testing.T) {
	tok, restore := checkboxForceLight(t)
	defer restore()

	c := graft.Checkbox()
	laidOutCheckbox(c)
	canvas := uitest.DrawWidget(c)

	for _, rr := range canvas.RoundRects {
		if rr.Color.A == 1 {
			t.Fatalf("unexpected opaque fill in light unchecked: %+v", rr)
		}
	}

	var border *uitest.StrokeRoundRectCall
	for idx := range canvas.StrokeRoundRects {
		if canvas.StrokeRoundRects[idx].StrokeWidth == metrics.Checkbox.BorderWidth {
			border = &canvas.StrokeRoundRects[idx]
		}
	}
	if border == nil {
		t.Fatal("no box border stroke found")
	}
	if border.Color != tok.Input {
		t.Fatalf("border color: got %+v want Input %+v", border.Color, tok.Input)
	}
	wantR := metrics.Checkbox.Radius - metrics.Checkbox.BorderWidth/2
	if border.Radius != wantR {
		t.Fatalf("border radius: got %v want %v", border.Radius, wantR)
	}
	// Box is 16×16 (stroke bounds are inset by w/2).
	bw := border.Bounds.Width() + metrics.Checkbox.BorderWidth
	if bw != metrics.Checkbox.Size {
		t.Fatalf("box width: got %v want %v", bw, metrics.Checkbox.Size)
	}
}

// TestCheckboxSpecCheckedLight asserts checked draws a Primary fill + Primary
// border (and the indicator is requested via icon.Draw, which no-ops on mock).
func TestCheckboxSpecCheckedLight(t *testing.T) {
	tok, restore := checkboxForceLight(t)
	defer restore()

	c := graft.Checkbox().Checked(true)
	laidOutCheckbox(c)
	canvas := uitest.DrawWidget(c)

	var fill *uitest.DrawRoundRectCall
	for idx := range canvas.RoundRects {
		if canvas.RoundRects[idx].Color == tok.Primary {
			fill = &canvas.RoundRects[idx]
		}
	}
	if fill == nil {
		t.Fatal("checked box should have a Primary fill")
	}
	if fill.Radius != metrics.Checkbox.Radius {
		t.Fatalf("fill radius: got %v want %v", fill.Radius, metrics.Checkbox.Radius)
	}

	var border *uitest.StrokeRoundRectCall
	for idx := range canvas.StrokeRoundRects {
		if canvas.StrokeRoundRects[idx].StrokeWidth == metrics.Checkbox.BorderWidth {
			border = &canvas.StrokeRoundRects[idx]
		}
	}
	if border == nil || border.Color != tok.Primary {
		t.Fatalf("checked border should be Primary, got %+v", border)
	}
}

// TestCheckboxSpecFocusRing asserts the focus state draws a 3px Ring/50 ring.
func TestCheckboxSpecFocusRing(t *testing.T) {
	tok, restore := checkboxForceLight(t)
	defer restore()

	c := graft.Checkbox()
	laidOutCheckbox(c)
	c.SetFocused(true)
	canvas := uitest.DrawWidget(c)

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

// TestCheckboxToggleClick verifies a click toggles and fires OnChange.
func TestCheckboxToggleClick(t *testing.T) {
	_, restore := checkboxForceLight(t)
	defer restore()

	var got []bool
	c := graft.Checkbox().OnChange(func(v bool) { got = append(got, v) })
	laidOutCheckbox(c)

	uitest.SimulateClick(c, 8, 8)
	if !c.IsChecked() {
		t.Fatal("click should check the box")
	}
	uitest.SimulateClick(c, 8, 8)
	if c.IsChecked() {
		t.Fatal("second click should uncheck the box")
	}
	if len(got) != 2 || got[0] != true || got[1] != false {
		t.Fatalf("OnChange sequence: got %v want [true false]", got)
	}
}

// TestCheckboxSpaceToggle verifies Space toggles when focused.
func TestCheckboxSpaceToggle(t *testing.T) {
	_, restore := checkboxForceLight(t)
	defer restore()

	c := graft.Checkbox()
	laidOutCheckbox(c)
	c.SetFocused(true)
	ctx := uitest.NewMockContext()
	ev := uitest.KeyPress(event.KeySpace, event.ModNone)
	if !c.Event(ctx, ev) {
		t.Fatal("Space should be consumed when focused")
	}
	if !c.IsChecked() {
		t.Fatal("Space should check the box")
	}
}

// TestCheckboxBindControlled verifies a bound signal drives the state.
func TestCheckboxBindControlled(t *testing.T) {
	_, restore := checkboxForceLight(t)
	defer restore()

	sig := state.NewSignal(false)
	c := graft.Checkbox().Bind(sig)
	laidOutCheckbox(c)
	uitest.SimulateClick(c, 8, 8)
	if !sig.Get() {
		t.Fatal("click should write true to bound signal")
	}
	sig.Set(false)
	if c.IsChecked() {
		t.Fatal("checkbox should reflect external signal change")
	}
}

// TestCheckboxDisabled pins disabled checkboxes out of focus + interaction.
func TestCheckboxDisabled(t *testing.T) {
	_, restore := checkboxForceLight(t)
	defer restore()

	c := graft.Checkbox().Disabled(true)
	laidOutCheckbox(c)
	if c.IsFocusable() {
		t.Fatal("disabled checkbox should not be focusable")
	}
	uitest.SimulateClick(c, 8, 8)
	if c.IsChecked() {
		t.Fatal("disabled checkbox should ignore clicks")
	}
}

// TestGoldenCheckbox renders the checkbox states in light and dark.
func TestGoldenCheckbox(t *testing.T) {
	gtest.GoldenLightDark(t, "checkbox-unchecked", func() widget.Widget {
		return checkboxRow(graft.Checkbox())
	})
	gtest.GoldenLightDark(t, "checkbox-checked", func() widget.Widget {
		return checkboxRow(graft.Checkbox().Checked(true))
	})
	gtest.GoldenLightDark(t, "checkbox-indeterminate", func() widget.Widget {
		return checkboxRow(graft.Checkbox().SetIndeterminate(true))
	})
	gtest.GoldenLightDark(t, "checkbox-with-label", func() widget.Widget {
		return checkboxRow(graft.Checkbox().Checked(true).Label("Accept terms and conditions"))
	})
	gtest.GoldenLightDark(t, "checkbox-disabled", func() widget.Widget {
		return checkboxRow(graft.Checkbox().Checked(true).Label("Disabled option").Disabled(true))
	})
	gtest.GoldenLightDark(t, "checkbox-focused", func() widget.Widget {
		c := graft.Checkbox().Checked(true)
		c.SetFocused(true)
		return checkboxRow(c)
	})
}

func checkboxRow(w widget.Widget) widget.Widget {
	return primitives.VBox(w).Padding(12)
}

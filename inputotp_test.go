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

// inputOTPForceLight pins the current theme to light for a spec test.
func inputOTPForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

func laidOutInputOTP(w widget.Widget) {
	w.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 100)))
}

// TestInputOTPLayout pins the total width from slot sizes, gaps, and group
// separators.
func TestInputOTPLayout(t *testing.T) {
	_, restore := inputOTPForceLight(t)
	defer restore()

	// Default: 6 slots in groups of [3,3].
	o := graft.InputOTP()
	size := o.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 100)))

	m := metrics.InputOTP
	// 3 slots + 2 gaps + separator + 3 slots + 2 gaps
	wantW := float32(6)*m.SlotSize + float32(4)*m.SlotGap + m.GroupGap
	if size.Width != wantW {
		t.Fatalf("width: got %v want %v", size.Width, wantW)
	}
	if size.Height != m.SlotSize {
		t.Fatalf("height: got %v want %v", size.Height, m.SlotSize)
	}
}

// TestInputOTPLayoutSingleGroup verifies layout with a single group.
func TestInputOTPLayoutSingleGroup(t *testing.T) {
	_, restore := inputOTPForceLight(t)
	defer restore()

	o := graft.InputOTP().Length(4).Groups(4)
	size := o.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(400, 100)))

	m := metrics.InputOTP
	wantW := float32(4)*m.SlotSize + float32(3)*m.SlotGap
	if size.Width != wantW {
		t.Fatalf("width: got %v want %v", size.Width, wantW)
	}
}

// TestInputOTPTyping verifies rune insertion fills slots left-to-right.
func TestInputOTPTyping(t *testing.T) {
	_, restore := inputOTPForceLight(t)
	defer restore()

	var completed string
	o := graft.InputOTP().Length(4).Groups(4).OnComplete(func(code string) {
		completed = code
	})
	laidOutInputOTP(o)
	o.SetFocused(true)
	ctx := uitest.NewMockContext()

	// Type "1234".
	for _, r := range "1234" {
		o.Event(ctx, event.NewKeyEvent(event.KeyPress, 0, r, 0))
	}
	if got := o.Value(); got != "1234" {
		t.Fatalf("after typing: got %q want %q", got, "1234")
	}
	if completed != "1234" {
		t.Fatalf("OnComplete: got %q want %q", completed, "1234")
	}
	if !o.IsComplete() {
		t.Fatal("should be complete after filling all slots")
	}
}

// TestInputOTPBackspace verifies backspace removes the last character.
func TestInputOTPBackspace(t *testing.T) {
	_, restore := inputOTPForceLight(t)
	defer restore()

	o := graft.InputOTP().Length(4).Groups(4).SetValue("123")
	laidOutInputOTP(o)
	o.SetFocused(true)
	ctx := uitest.NewMockContext()

	o.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeyBackspace, 0, 0))
	if got := o.Value(); got != "12" {
		t.Fatalf("after backspace: got %q want %q", got, "12")
	}

	// Backspace on empty is a no-op.
	o.SetValue("")
	o.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeyBackspace, 0, 0))
	if got := o.Value(); got != "" {
		t.Fatalf("backspace on empty: got %q want %q", got, "")
	}
}

// TestInputOTPPattern verifies the pattern filter rejects non-matching input.
func TestInputOTPPattern(t *testing.T) {
	_, restore := inputOTPForceLight(t)
	defer restore()

	o := graft.InputOTP().Length(4).Groups(4).Pattern("[0-9]")
	laidOutInputOTP(o)
	o.SetFocused(true)
	ctx := uitest.NewMockContext()

	// Letters should be rejected.
	o.Event(ctx, event.NewKeyEvent(event.KeyPress, 0, 'a', 0))
	if got := o.Value(); got != "" {
		t.Fatalf("letter should be rejected: got %q", got)
	}

	// Digits should be accepted.
	o.Event(ctx, event.NewKeyEvent(event.KeyPress, 0, '5', 0))
	if got := o.Value(); got != "5" {
		t.Fatalf("digit should be accepted: got %q", got)
	}
}

// TestInputOTPOverflow verifies typing beyond length is ignored.
func TestInputOTPOverflow(t *testing.T) {
	_, restore := inputOTPForceLight(t)
	defer restore()

	o := graft.InputOTP().Length(2).Groups(2)
	laidOutInputOTP(o)
	o.SetFocused(true)
	ctx := uitest.NewMockContext()

	for _, r := range "abc" {
		o.Event(ctx, event.NewKeyEvent(event.KeyPress, 0, r, 0))
	}
	if got := o.Value(); got != "ab" {
		t.Fatalf("overflow: got %q want %q", got, "ab")
	}
}

// TestInputOTPBindControlled verifies a bound signal drives the state.
func TestInputOTPBindControlled(t *testing.T) {
	_, restore := inputOTPForceLight(t)
	defer restore()

	sig := state.NewSignal("")
	o := graft.InputOTP().Length(4).Groups(4).Bind(sig)
	laidOutInputOTP(o)
	o.SetFocused(true)
	ctx := uitest.NewMockContext()

	o.Event(ctx, event.NewKeyEvent(event.KeyPress, 0, 'A', 0))
	if sig.Get() != "A" {
		t.Fatalf("bound signal should reflect typed input: got %q", sig.Get())
	}

	sig.Set("XY")
	if o.Value() != "XY" {
		t.Fatalf("widget should reflect external signal change: got %q", o.Value())
	}
}

// TestInputOTPOnChange verifies the onChange callback fires on every change.
func TestInputOTPOnChange(t *testing.T) {
	_, restore := inputOTPForceLight(t)
	defer restore()

	var changes []string
	o := graft.InputOTP().Length(3).Groups(3).OnChange(func(v string) {
		changes = append(changes, v)
	})
	laidOutInputOTP(o)
	o.SetFocused(true)
	ctx := uitest.NewMockContext()

	o.Event(ctx, event.NewKeyEvent(event.KeyPress, 0, 'a', 0))
	o.Event(ctx, event.NewKeyEvent(event.KeyPress, 0, 'b', 0))
	o.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeyBackspace, 0, 0))

	want := []string{"a", "ab", "a"}
	if len(changes) != len(want) {
		t.Fatalf("onChange count: got %d want %d", len(changes), len(want))
	}
	for i, w := range want {
		if changes[i] != w {
			t.Fatalf("onChange[%d]: got %q want %q", i, changes[i], w)
		}
	}
}

// TestInputOTPDisabled verifies disabled InputOTP ignores input and is not
// focusable.
func TestInputOTPDisabled(t *testing.T) {
	_, restore := inputOTPForceLight(t)
	defer restore()

	o := graft.InputOTP().Length(4).Groups(4).Disabled(true)
	laidOutInputOTP(o)
	if o.IsFocusable() {
		t.Fatal("disabled InputOTP should not be focusable")
	}
}

// TestInputOTPSpecEmptyLight asserts slot borders use the Input token color.
func TestInputOTPSpecEmptyLight(t *testing.T) {
	tok, restore := inputOTPForceLight(t)
	defer restore()

	o := graft.InputOTP().Length(4).Groups(2, 2)
	laidOutInputOTP(o)
	canvas := uitest.DrawWidget(o)

	// Should have slot borders using Input color.
	var borderCount int
	for _, s := range canvas.StrokeRoundRects {
		if s.StrokeWidth == metrics.InputOTP.BorderWidth && s.Color == tok.Input {
			borderCount++
		}
	}
	if borderCount < 4 {
		t.Fatalf("expected at least 4 Input-colored borders, got %d", borderCount)
	}
}

// TestInputOTPSpecFocusRing asserts the focused slot gets a ring.
func TestInputOTPSpecFocusRing(t *testing.T) {
	_, restore := inputOTPForceLight(t)
	defer restore()

	o := graft.InputOTP().Length(4).Groups(4)
	laidOutInputOTP(o)
	o.SetFocused(true)
	canvas := uitest.DrawWidget(o)

	// The focused slot should have a focus ring (width 3).
	var ring *uitest.StrokeRoundRectCall
	for idx := range canvas.StrokeRoundRects {
		if canvas.StrokeRoundRects[idx].StrokeWidth == metrics.RingWidth {
			ring = &canvas.StrokeRoundRects[idx]
			break
		}
	}
	if ring == nil {
		t.Fatal("no focus ring stroke (width 3) on active slot")
	}
}

// TestInputOTPSpecSeparator asserts the separator dash is drawn between groups.
func TestInputOTPSpecSeparator(t *testing.T) {
	tok, restore := inputOTPForceLight(t)
	defer restore()

	o := graft.InputOTP().Length(6).Groups(3, 3)
	laidOutInputOTP(o)
	canvas := uitest.DrawWidget(o)

	// Should have at least one separator rect (a Foreground-colored flat rect).
	var sepCount int
	for _, r := range canvas.Rects {
		if r.Color == tok.Foreground {
			w := r.Bounds.Width()
			h := r.Bounds.Height()
			if w == metrics.InputOTP.SepWidth && h == metrics.InputOTP.SepHeight {
				sepCount++
			}
		}
	}
	if sepCount < 1 {
		t.Fatalf("expected at least 1 separator dash, got %d", sepCount)
	}
}

// TestGoldenInputOTP renders the InputOTP states in light and dark.
func TestGoldenInputOTP(t *testing.T) {
	gtest.GoldenLightDark(t, "inputotp-empty", func() widget.Widget {
		return inputOTPRow(graft.InputOTP().Length(6).Groups(3, 3))
	})
	gtest.GoldenLightDark(t, "inputotp-partial", func() widget.Widget {
		return inputOTPRow(graft.InputOTP().Length(6).Groups(3, 3).SetValue("123"))
	})
	gtest.GoldenLightDark(t, "inputotp-full", func() widget.Widget {
		return inputOTPRow(graft.InputOTP().Length(6).Groups(3, 3).SetValue("123456"))
	})
	gtest.GoldenLightDark(t, "inputotp-focused", func() widget.Widget {
		o := graft.InputOTP().Length(6).Groups(3, 3).SetValue("12")
		o.SetFocused(true)
		return inputOTPRow(o)
	})
	gtest.GoldenLightDark(t, "inputotp-disabled", func() widget.Widget {
		return inputOTPRow(graft.InputOTP().Length(6).Groups(3, 3).SetValue("123").Disabled(true))
	})
	gtest.GoldenLightDark(t, "inputotp-4digit", func() widget.Widget {
		return inputOTPRow(graft.InputOTP().Length(4).Groups(4).SetValue("90"))
	})
}

func inputOTPRow(w widget.Widget) widget.Widget {
	return primitives.VBox(w).Padding(12)
}

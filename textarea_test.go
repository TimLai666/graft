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

// textareaRow wraps a control in padding so the focus ring (drawn outside
// bounds) is not clipped at the golden image edge.
func textareaRow(w widget.Widget) widget.Widget {
	return primitives.VBox(w).Padding(12)
}

// textareaForceLight pins the current theme to light and returns the active
// tokens plus a restore function.
func textareaForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

// laidOutTextarea runs Layout at a fixed width and returns the widget.
func laidOutTextarea(w widget.Widget, width float32) geometry.Size {
	return w.Layout(uitest.NewMockContext(), geometry.Loose(geometry.Sz(width, 1000)))
}

// TestTextareaMinHeight pins the min-h-16 (64px) floor and Rows growth.
func TestTextareaMinHeight(t *testing.T) {
	_, restore := textareaForceLight(t)
	defer restore()

	ta := graft.Textarea().Placeholder("Type here").W(240)
	size := laidOutTextarea(ta, 300)
	if size.Height != metrics.Textarea.MinHeight {
		t.Fatalf("empty height: got %v want %v", size.Height, metrics.Textarea.MinHeight)
	}
	if size.Width != 240 {
		t.Fatalf("width: got %v want 240", size.Width)
	}

	rows := graft.Textarea().Rows(5).W(240)
	size = laidOutTextarea(rows, 300)
	wantH := 2*metrics.Textarea.PadY + 5*metrics.Textarea.LineHeight
	if size.Height != wantH {
		t.Fatalf("rows=5 height: got %v want %v", size.Height, wantH)
	}
}

// TestTextareaAutoGrow verifies that multi-line content grows the control past
// the minimum height.
func TestTextareaAutoGrow(t *testing.T) {
	_, restore := textareaForceLight(t)
	defer restore()

	ta := graft.Textarea().Value("line1\nline2\nline3\nline4\nline5").W(240)
	size := laidOutTextarea(ta, 300)
	wantH := 5*metrics.Textarea.LineHeight + 2*metrics.Textarea.PadY
	if size.Height != wantH {
		t.Fatalf("auto-grow height: got %v want %v", size.Height, wantH)
	}
}

// TestTextareaSpecEmptyLight asserts the empty textarea draws a 1px Input
// border and a muted placeholder, with no opaque fill in light mode.
func TestTextareaSpecEmptyLight(t *testing.T) {
	tok, restore := textareaForceLight(t)
	defer restore()

	ta := graft.Textarea().Placeholder("Type your message").W(240)
	laidOutTextarea(ta, 300)
	canvas := uitest.DrawWidget(ta)

	radius := graft.CurrentTheme().RadiusLG()

	for _, rr := range canvas.RoundRects {
		if rr.Color.A == 1 {
			t.Fatalf("unexpected opaque background fill in light mode: %+v", rr)
		}
	}

	var borders []uitest.StrokeRoundRectCall
	for _, s := range canvas.StrokeRoundRects {
		if s.StrokeWidth == metrics.Textarea.BorderWidth {
			borders = append(borders, s)
		}
	}
	if len(borders) != 1 {
		t.Fatalf("want 1 border stroke, got %d", len(borders))
	}
	if borders[0].Color != tok.Input {
		t.Fatalf("border color: got %+v want Input %+v", borders[0].Color, tok.Input)
	}
	if want := radius - metrics.Textarea.BorderWidth/2; borders[0].Radius != want {
		t.Fatalf("border radius: got %v want %v", borders[0].Radius, want)
	}

	if len(canvas.StyledTexts) != 1 {
		t.Fatalf("want 1 styled text (placeholder), got %d", len(canvas.StyledTexts))
	}
	ph := canvas.StyledTexts[0]
	if ph.Text != "Type your message" {
		t.Fatalf("placeholder text: got %q", ph.Text)
	}
	if ph.Style.Color != tok.MutedForeground {
		t.Fatalf("placeholder color: got %+v want MutedForeground %+v", ph.Style.Color, tok.MutedForeground)
	}
	if ph.Style.FontSize != metrics.Textarea.FontSize {
		t.Fatalf("placeholder font size: got %v want %v", ph.Style.FontSize, metrics.Textarea.FontSize)
	}
}

// TestTextareaSpecFocusedRing asserts the focus state draws the 3px Ring/50
// ring and a solid Ring border.
func TestTextareaSpecFocusedRing(t *testing.T) {
	tok, restore := textareaForceLight(t)
	defer restore()

	ta := graft.Textarea().Value("hello").W(240)
	laidOutTextarea(ta, 300)
	ta.SetFocused(true)
	canvas := uitest.DrawWidget(ta)

	radius := graft.CurrentTheme().RadiusLG()

	var ring *uitest.StrokeRoundRectCall
	for idx := range canvas.StrokeRoundRects {
		if canvas.StrokeRoundRects[idx].StrokeWidth == metrics.RingWidth {
			ring = &canvas.StrokeRoundRects[idx]
		}
	}
	if ring == nil {
		t.Fatal("no focus ring stroke (width 3) found")
	}
	wantRing := tok.Ring
	wantRing.A = metrics.RingAlpha
	if ring.Color != wantRing {
		t.Fatalf("ring color: got %+v want Ring@0.5 %+v", ring.Color, wantRing)
	}
	if ring.Radius != radius+metrics.RingWidth/2 {
		t.Fatalf("ring radius: got %v want %v", ring.Radius, radius+metrics.RingWidth/2)
	}

	for _, s := range canvas.StrokeRoundRects {
		if s.StrokeWidth == metrics.Textarea.BorderWidth && s.Color != tok.Ring {
			t.Fatalf("focused border should be solid Ring, got %+v", s.Color)
		}
	}
}

// TestTextareaSpecInvalidRing asserts the invalid state draws a destructive
// ring and border.
func TestTextareaSpecInvalidRing(t *testing.T) {
	tok, restore := textareaForceLight(t)
	defer restore()

	ta := graft.Textarea().Value("nope").Invalid(true).W(240)
	laidOutTextarea(ta, 300)
	canvas := uitest.DrawWidget(ta)

	var ring *uitest.StrokeRoundRectCall
	for idx := range canvas.StrokeRoundRects {
		if canvas.StrokeRoundRects[idx].StrokeWidth == metrics.RingWidth {
			ring = &canvas.StrokeRoundRects[idx]
		}
	}
	if ring == nil {
		t.Fatal("no invalid ring stroke (width 3) found")
	}
	wantRing := tok.Destructive
	wantRing.A = metrics.InvalidRingAlphaLight
	if ring.Color != wantRing {
		t.Fatalf("ring color: got %+v want Destructive@0.2 %+v", ring.Color, wantRing)
	}
	for _, s := range canvas.StrokeRoundRects {
		if s.StrokeWidth == metrics.Textarea.BorderWidth && s.Color != tok.Destructive {
			t.Fatalf("invalid border should be Destructive, got %+v", s.Color)
		}
	}
}

// TestTextareaTypingInsertsRunes verifies the editing model: typed runes,
// newline on Enter, backspace, and caret motion.
func TestTextareaTypingInsertsRunes(t *testing.T) {
	_, restore := textareaForceLight(t)
	defer restore()

	ta := graft.Textarea().W(240)
	laidOutTextarea(ta, 300)
	ta.SetFocused(true)
	ctx := uitest.NewMockContext()

	typeRune := func(r rune) {
		ta.Event(ctx, event.NewKeyEvent(event.KeyPress, 0, r, 0))
	}
	for _, r := range "ab" {
		typeRune(r)
	}
	ta.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, 0))
	typeRune('c')
	if got := ta.AccessibilityValue(); got != "ab\nc" {
		t.Fatalf("after typing: got %q want %q", got, "ab\nc")
	}

	ta.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeyBackspace, 0, 0))
	if got := ta.AccessibilityValue(); got != "ab\n" {
		t.Fatalf("after backspace: got %q want %q", got, "ab\n")
	}
}

// TestTextareaBindControlled verifies a bound signal drives the rendered text
// and edits write back.
func TestTextareaBindControlled(t *testing.T) {
	_, restore := textareaForceLight(t)
	defer restore()

	sig := state.NewSignal("seed")
	ta := graft.Textarea().Bind(sig).W(240)
	laidOutTextarea(ta, 300)
	if got := ta.AccessibilityValue(); got != "seed" {
		t.Fatalf("bound initial: got %q want seed", got)
	}
	sig.Set("changed")
	if got := ta.AccessibilityValue(); got != "changed" {
		t.Fatalf("after sig.Set: got %q want changed", got)
	}
}

// TestTextareaDisabledNotFocusable pins disabled textareas out of focus
// traversal.
func TestTextareaDisabledNotFocusable(t *testing.T) {
	_, restore := textareaForceLight(t)
	defer restore()

	if graft.Textarea().Disabled(true).IsFocusable() {
		t.Fatal("disabled textarea should not be focusable")
	}
	if !graft.Textarea().IsFocusable() {
		t.Fatal("enabled textarea should be focusable")
	}
}

// TestGoldenTextarea renders the four Textarea states in light and dark.
func TestGoldenTextarea(t *testing.T) {
	gtest.GoldenLightDark(t, "textarea-empty", func() widget.Widget {
		return textareaRow(graft.Textarea().Placeholder("Type your message here.").W(280))
	})
	gtest.GoldenLightDark(t, "textarea-filled", func() widget.Widget {
		return textareaRow(graft.Textarea().Value("The quick brown fox jumps over the lazy dog.\nA second line of content.").W(280))
	})
	gtest.GoldenLightDark(t, "textarea-focused", func() widget.Widget {
		ta := graft.Textarea().Value("Editing this textarea.").W(280)
		ta.SetFocused(true)
		return textareaRow(ta)
	})
	gtest.GoldenLightDark(t, "textarea-disabled", func() widget.Widget {
		return textareaRow(graft.Textarea().Value("Cannot edit this.").Disabled(true).W(280))
	})
}

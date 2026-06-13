package graft_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// ovForceLightMode sets the process theme to light and returns a restore func.
// Overlay-batch-local helper (prefixed to avoid merge collisions with the
// shared forceLightMode that batch 1 defines).
func ovForceLightMode(t *testing.T) func() {
	t.Helper()
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return func() { th.SetMode(prev) }
}

// ovFakeOverlayManager records overlay pushes/removes so spec tests can verify
// the show/hide state machine without a real window.
type ovFakeOverlayManager struct {
	pushed    []widget.Widget
	removed   []widget.Widget
	onDismiss []func()
}

func (m *ovFakeOverlayManager) PushOverlay(w widget.Widget, onDismiss func()) {
	m.pushed = append(m.pushed, w)
	m.onDismiss = append(m.onDismiss, onDismiss)
}
func (m *ovFakeOverlayManager) PopOverlay() {}
func (m *ovFakeOverlayManager) RemoveOverlay(w widget.Widget) {
	m.removed = append(m.removed, w)
}

func (m *ovFakeOverlayManager) liveCount() int { return len(m.pushed) - len(m.removed) }

// TestTooltipContentGeometry pins the bubble fill: foreground color, radius MD,
// height = lineHeight + 2*padY, width = textWidth + 2*padX.
func TestTooltipContentGeometry(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	tok := graft.CurrentTheme().Active()

	tip := graft.Tooltip(primitives.Box().Width(40).Height(20), "Add to library")
	content := tip.Content()
	size := content.Layout(uitest.NewMockContext(), looseConstraints())

	wantH := metrics.TooltipLineHeight + 2*metrics.TooltipPadY
	if size.Height != wantH {
		t.Fatalf("bubble height: got %v want %v", size.Height, wantH)
	}
	if size.Width <= 2*metrics.TooltipPadX {
		t.Fatalf("bubble width too small: got %v", size.Width)
	}

	content.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	canvas := uitest.DrawWidget(content)

	var bubble *uitest.DrawRoundRectCall
	for i := range canvas.RoundRects {
		c := &canvas.RoundRects[i]
		if c.Bounds.Size() == size {
			bubble = c
			break
		}
	}
	if bubble == nil {
		t.Fatalf("no bubble round-rect matching size %v; got %d round-rects", size, len(canvas.RoundRects))
	}
	if bubble.Color != tok.Foreground {
		t.Errorf("bubble fill: got %+v want foreground %+v", bubble.Color, tok.Foreground)
	}
	wantRadius := graft.CurrentTheme().RadiusMD()
	if bubble.Radius != wantRadius {
		t.Errorf("bubble radius: got %v want MD %v", bubble.Radius, wantRadius)
	}
}

// TestTooltipTextInverted pins the inverted label: drawn in the Background token.
func TestTooltipTextInverted(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	tok := graft.CurrentTheme().Active()

	tip := graft.Tooltip(primitives.Box().Width(40).Height(20), "Inverted")
	content := tip.Content()
	size := content.Layout(uitest.NewMockContext(), looseConstraints())
	content.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	canvas := uitest.DrawWidget(content)

	if len(canvas.StyledTexts) == 0 {
		t.Fatal("no styled text drawn for tooltip label")
	}
	st := canvas.StyledTexts[0]
	if st.Text != "Inverted" {
		t.Errorf("label text: got %q", st.Text)
	}
	if st.Style.FontSize != metrics.TooltipFontSize {
		t.Errorf("label font size: got %v want %v", st.Style.FontSize, metrics.TooltipFontSize)
	}
	if st.Style.Color != tok.Background {
		t.Errorf("label color: got %+v want background %+v (inverted)", st.Style.Color, tok.Background)
	}
}

// TestTooltipShowHide drives the hover state machine against a fake overlay
// manager: hover past the delay pushes the bubble; leave removes it.
func TestTooltipShowHide(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	tip := graft.Tooltip(primitives.Box().Width(40).Height(20), "Hi")
	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om

	tip.Layout(ctx, looseConstraints())

	// Enter at t0: within the grace period, nothing shown yet.
	t0 := ctx.TimeVal
	tip.Event(ctx, uitest.MouseEnter(5, 5))
	if om.liveCount() != 0 {
		t.Fatalf("tooltip shown before delay elapsed: live=%d", om.liveCount())
	}

	// Advance past the delay and send a move inside: bubble pushed.
	ctx.TimeVal = t0.Add(metrics.TooltipDelayMillis*time.Millisecond + time.Millisecond)
	tip.Event(ctx, uitest.MouseMove(5, 5))
	if om.liveCount() != 1 {
		t.Fatalf("tooltip not shown after delay: live=%d pushed=%d", om.liveCount(), len(om.pushed))
	}

	// Leave: bubble removed.
	tip.Event(ctx, uitest.MouseLeave(100, 100))
	if om.liveCount() != 0 {
		t.Fatalf("tooltip not hidden on leave: live=%d", om.liveCount())
	}
}

// TestGoldenTooltip renders the bubble (with arrow) directly, light + dark.
func TestGoldenTooltip(t *testing.T) {
	gtest.GoldenLightDark(t, "tooltip-content", func() widget.Widget {
		tip := graft.Tooltip(primitives.Box().Width(0).Height(0), "Add to library")
		c := tip.Content()
		// Pad the frame so the arrow (drawn outside the bubble) is captured.
		return primitives.VBox(c).Padding(16)
	})
}

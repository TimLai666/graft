package graft_test

import (
	"math"
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
)

// TestProgressSpecTrack verifies the track: full width, 8px tall, Primary@20%
// fill at rounded-full.
func TestProgressSpecTrack(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()

	p := graft.Progress().Value(0)
	size := p.Layout(nil, fixedWidthLoose(300))
	if size.Width != 300 || size.Height != 8 {
		t.Errorf("progress size = %vx%v, want 300x8", size.Width, size.Height)
	}

	canvas := uitest.DrawWidget(p)
	if len(canvas.RoundRects) != 1 {
		t.Fatalf("0%% progress drew %d round-rects, want 1 (track only)", len(canvas.RoundRects))
	}
	track := canvas.RoundRects[0]
	wantTrack := tok.Primary
	wantTrack.A = 0.2
	if track.Color != wantTrack {
		t.Errorf("track color = %v, want Primary@20%% %v", track.Color, wantTrack)
	}
	if track.Radius != th.RadiusFull() {
		t.Errorf("track radius = %v, want RadiusFull %v", track.Radius, th.RadiusFull())
	}
}

// TestProgressSpecIndicator verifies the indicator at 50%: a Primary fill
// half the track width, clipped to the rounded-full pill.
func TestProgressSpecIndicator(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()

	p := graft.Progress().Value(50)
	p.Layout(nil, fixedWidthLoose(300))
	canvas := uitest.DrawWidget(p)

	if len(canvas.RoundRects) != 2 {
		t.Fatalf("50%% progress drew %d round-rects, want 2 (track + indicator)", len(canvas.RoundRects))
	}
	if len(canvas.ClipRoundRects) != 1 {
		t.Fatalf("indicator drew %d round-rect clips, want 1", len(canvas.ClipRoundRects))
	}
	ind := canvas.RoundRects[1]
	if ind.Color != tok.Primary {
		t.Errorf("indicator color = %v, want Primary %v", ind.Color, tok.Primary)
	}
	if math.Abs(float64(ind.Bounds.Width()-150)) > 1e-3 {
		t.Errorf("indicator width = %v, want 150 (50%% of 300)", ind.Bounds.Width())
	}
	// Clip is the full pill so the leading edge stays rounded.
	if canvas.ClipRoundRects[0].Radius != th.RadiusFull() {
		t.Errorf("clip radius = %v, want RadiusFull", canvas.ClipRoundRects[0].Radius)
	}
}

// TestProgressSpecFull verifies 100% fills the entire track width.
func TestProgressSpecFull(t *testing.T) {
	alertForceLight(t)
	p := graft.Progress().Value(100)
	p.Layout(nil, fixedWidthLoose(300))
	canvas := uitest.DrawWidget(p)
	if len(canvas.RoundRects) != 2 {
		t.Fatalf("100%% progress drew %d round-rects, want 2", len(canvas.RoundRects))
	}
	if math.Abs(float64(canvas.RoundRects[1].Bounds.Width()-300)) > 1e-3 {
		t.Errorf("indicator width = %v, want 300 (full)", canvas.RoundRects[1].Bounds.Width())
	}
}

// TestProgressClampsAndValue checks value clamping at the edges.
func TestProgressClampsAndValue(t *testing.T) {
	alertForceLight(t)

	over := graft.Progress().Value(150)
	over.Layout(nil, fixedWidthLoose(300))
	cOver := uitest.DrawWidget(over)
	if len(cOver.RoundRects) != 2 || math.Abs(float64(cOver.RoundRects[1].Bounds.Width()-300)) > 1e-3 {
		t.Error("value > 100 should clamp to full")
	}

	under := graft.Progress().Value(-20)
	under.Layout(nil, fixedWidthLoose(300))
	cUnder := uitest.DrawWidget(under)
	if len(cUnder.RoundRects) != 1 {
		t.Error("value < 0 should clamp to empty (track only)")
	}
}

// TestGoldenProgress renders the four canonical fill levels in light and dark.
func TestGoldenProgress(t *testing.T) {
	gtest.GoldenLightDark(t, "progress-levels", func() widget.Widget {
		bar := func(v float64) widget.Widget {
			return primitives.VBox(graft.Progress().Value(v)).Width(320)
		}
		return primitives.VBox(
			bar(0), bar(33), bar(66), bar(100),
		).Gap(16).Padding(24)
	})
}

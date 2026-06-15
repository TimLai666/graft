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
)

// hoverCardDemoContent builds the canonical demo body: a heading and two
// muted lines, the shadcn @nextjs profile example shape.
func hoverCardDemoContent() *graft.HoverCardContentWidget {
	return graft.HoverCardContent(
		primitives.VBox(
			graft.Small("@nextjs"),
			graft.MutedText("The React Framework — created and maintained by @vercel."),
		).Gap(8),
	)
}

// TestHoverCardContentGeometry pins the shadcn hover-card surface: w-64 (256),
// p-4 (16), rounded-md, 1px border, bg-popover, shadow-md.
func TestHoverCardContentGeometry(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	tok := graft.CurrentTheme().Active()
	th := graft.CurrentTheme()

	content := hoverCardDemoContent()
	size := uitest.LayoutWidget(content, 900, 900)
	if size.Width != metrics.HoverCardWidth {
		t.Errorf("width = %v, want %v (w-64)", size.Width, metrics.HoverCardWidth)
	}

	content.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	canvas := uitest.DrawWidget(content)

	// BorderFill: outer border round-rect (full bounds) then inner bg-popover
	// fill (inset by the 1px border). shadow-md paints before both.
	bw := metrics.HoverCardBorderWidth

	border := canvas.RoundRects[len(metrics.ShadowMD)]
	if border.Color != tok.Border {
		t.Errorf("border fill = %+v, want border token %v", border, tok.Border)
	}
	if border.Radius != th.RadiusMD() ||
		border.Bounds != geometry.NewRect(0, 0, size.Width, size.Height) {
		t.Errorf("border fill = %+v, want radius %v full bounds", border, th.RadiusMD())
	}

	bg := canvas.RoundRects[len(metrics.ShadowMD)+1]
	if bg.Color != tok.Popover {
		t.Errorf("bg fill = %+v, want bg-popover %v", bg, tok.Popover)
	}
	if bg.Radius != th.RadiusMD()-bw ||
		bg.Bounds != geometry.NewRect(bw, bw, size.Width-2*bw, size.Height-2*bw) {
		t.Errorf("bg fill = %+v, want radius %v inset by %v", bg, th.RadiusMD()-bw, bw)
	}

	// No border stroke any more (border is now a fill).
	if len(canvas.StrokeRoundRects) != 0 {
		t.Fatalf("strokes = %d, want 0 (border now a fill)", len(canvas.StrokeRoundRects))
	}
}

// TestHoverCardWidthOverride pins the W override.
func TestHoverCardWidthOverride(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	content := graft.HoverCardContent(graft.Text("hi")).W(320)
	size := uitest.LayoutWidget(content, 900, 900)
	if size.Width != 320 {
		t.Errorf("width = %v, want 320", size.Width)
	}
}

// TestHoverCardOpenDelay drives the hover state machine: hover within the
// open-delay shows nothing; advancing past it pushes the card; leaving and
// advancing past the close delay removes it.
func TestHoverCardOpenDelay(t *testing.T) {
	defer ovForceLightMode(t)()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}

	hc := graft.HoverCard(primitives.Box().Width(60).Height(20), hoverCardDemoContent())
	om := &ovFakeOverlayManager{}
	ctx := uitest.NewMockContext()
	ctx.OverlayVal = om
	hc.Layout(ctx, looseConstraints())

	// Enter at t0: within the open delay, nothing shown.
	t0 := ctx.TimeVal
	hc.Event(ctx, uitest.MouseEnter(5, 5))
	if om.liveCount() != 0 {
		t.Fatalf("card shown before open delay: live=%d", om.liveCount())
	}

	// Advance past the open delay, move inside: card pushed.
	ctx.TimeVal = t0.Add(metrics.HoverCardOpenDelayMillis*time.Millisecond + time.Millisecond)
	hc.Event(ctx, uitest.MouseMove(5, 5))
	if om.liveCount() != 1 {
		t.Fatalf("card not shown after open delay: live=%d", om.liveCount())
	}

	// Leave at t1: within the close delay the card stays.
	t1 := ctx.TimeVal
	hc.Event(ctx, uitest.MouseLeave(500, 500))
	if om.liveCount() != 1 {
		t.Fatalf("card removed before close delay: live=%d", om.liveCount())
	}

	// Advance past the close delay, send another outside move: card removed.
	ctx.TimeVal = t1.Add(metrics.HoverCardCloseDelayMillis*time.Millisecond + time.Millisecond)
	hc.Event(ctx, uitest.MouseMove(500, 500))
	if om.liveCount() != 0 {
		t.Fatalf("card not removed after close delay: live=%d", om.liveCount())
	}
}

// TestGoldenHoverCard renders the card content directly, light + dark.
func TestGoldenHoverCard(t *testing.T) {
	gtest.GoldenLightDark(t, "hover-card-content", func() widget.Widget {
		content := hoverCardDemoContent()
		// Pad the frame so the shadow is captured.
		return primitives.VBox(content).Padding(24)
	})
}

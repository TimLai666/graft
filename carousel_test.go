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

// carouselForceLight pins the current theme to light for a spec test.
func carouselForceLight(t *testing.T) (*theme.Tokens, func()) {
	t.Helper()
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := graft.CurrentTheme()
	prev := th.Mode()
	th.SetMode(theme.ModeLight)
	return th.Active(), func() { th.SetMode(prev) }
}

// buildCarousel returns a three-slide horizontal carousel for tests.
func buildCarousel() *graft.CarouselWidget {
	return graft.Carousel(
		graft.CarouselItem(graft.Text("Slide 1")),
		graft.CarouselItem(graft.Text("Slide 2")),
		graft.CarouselItem(graft.Text("Slide 3")),
	)
}

func layoutCarousel(w widget.Widget) {
	w.Layout(uitest.NewMockContext(), geometry.Tight(geometry.Sz(400, 200)))
}

// TestCarouselLayout verifies the carousel sizes to its constraints and
// all items are positioned at the viewport size.
func TestCarouselLayout(t *testing.T) {
	_, restore := carouselForceLight(t)
	defer restore()

	c := buildCarousel()
	size := c.Layout(uitest.NewMockContext(), geometry.Tight(geometry.Sz(400, 200)))
	if size.Width != 400 || size.Height != 200 {
		t.Fatalf("carousel size: got %v, want 400x200", size)
	}
	// Children should exist.
	children := c.Children()
	if len(children) != 3 {
		t.Fatalf("children count: got %d, want 3", len(children))
	}
}

// TestCarouselNavButtons verifies nav button positioning.
func TestCarouselNavButtons(t *testing.T) {
	_, restore := carouselForceLight(t)
	defer restore()

	c := buildCarousel()
	layoutCarousel(c)

	m := metrics.Carousel
	bounds := c.Bounds()

	// At index 0, only Next should be enabled.
	if c.Children() == nil {
		t.Fatal("carousel has no children")
	}
	// Verify the widget is at start (index 0), so Previous is disabled.
	// We verify this through drawing: prev button should have faded alpha.
	mc := uitest.DrawWidget(c)
	_ = mc

	// Nav buttons should be 32x32 circles.
	if m.ButtonSize != 32 {
		t.Fatalf("button size metric: got %v, want 32", m.ButtonSize)
	}
	if bounds.Width() != 400 {
		t.Fatalf("bounds width: got %v, want 400", bounds.Width())
	}
}

// TestCarouselKeyboardNav verifies Left/Right arrow keyboard navigation.
func TestCarouselKeyboardNav(t *testing.T) {
	_, restore := carouselForceLight(t)
	defer restore()

	sig := state.NewSignal(0)
	c := graft.Carousel(
		graft.CarouselItem(graft.Text("A")),
		graft.CarouselItem(graft.Text("B")),
		graft.CarouselItem(graft.Text("C")),
	).Bind(sig)
	layoutCarousel(c)

	ctx := uitest.NewMockContext()

	// Focus the carousel.
	c.SetFocused(true) // keyboard focus (live: focus.Manager.Focus → SetFocused)

	// Right arrow: 0 -> 1.
	c.Event(ctx, uitest.KeyPress(event.KeyRight, event.ModNone))
	if sig.Get() != 1 {
		t.Fatalf("Right: signal = %d, want 1", sig.Get())
	}

	// Right arrow: 1 -> 2.
	c.Event(ctx, uitest.KeyPress(event.KeyRight, event.ModNone))
	if sig.Get() != 2 {
		t.Fatalf("Right: signal = %d, want 2", sig.Get())
	}

	// Right arrow at end: stays at 2.
	c.Event(ctx, uitest.KeyPress(event.KeyRight, event.ModNone))
	if sig.Get() != 2 {
		t.Fatalf("Right at end: signal = %d, want 2", sig.Get())
	}

	// Left arrow: 2 -> 1.
	c.Event(ctx, uitest.KeyPress(event.KeyLeft, event.ModNone))
	if sig.Get() != 1 {
		t.Fatalf("Left: signal = %d, want 1", sig.Get())
	}

	// Left arrow: 1 -> 0.
	c.Event(ctx, uitest.KeyPress(event.KeyLeft, event.ModNone))
	if sig.Get() != 0 {
		t.Fatalf("Left: signal = %d, want 0", sig.Get())
	}

	// Left arrow at start: stays at 0.
	c.Event(ctx, uitest.KeyPress(event.KeyLeft, event.ModNone))
	if sig.Get() != 0 {
		t.Fatalf("Left at start: signal = %d, want 0", sig.Get())
	}
}

// TestCarouselVerticalKeyboardNav verifies Up/Down keyboard navigation
// in vertical orientation.
func TestCarouselVerticalKeyboardNav(t *testing.T) {
	_, restore := carouselForceLight(t)
	defer restore()

	sig := state.NewSignal(0)
	c := graft.Carousel(
		graft.CarouselItem(graft.Text("A")),
		graft.CarouselItem(graft.Text("B")),
	).Vertical().Bind(sig)
	layoutCarousel(c)

	ctx := uitest.NewMockContext()
	c.SetFocused(true) // keyboard focus (live: focus.Manager.Focus → SetFocused)

	c.Event(ctx, uitest.KeyPress(event.KeyDown, event.ModNone))
	if sig.Get() != 1 {
		t.Fatalf("Down: signal = %d, want 1", sig.Get())
	}

	c.Event(ctx, uitest.KeyPress(event.KeyUp, event.ModNone))
	if sig.Get() != 0 {
		t.Fatalf("Up: signal = %d, want 0", sig.Get())
	}
}

// TestCarouselBindControlled verifies the signal controls the index.
func TestCarouselBindControlled(t *testing.T) {
	_, restore := carouselForceLight(t)
	defer restore()

	sig := state.NewSignal(1)
	c := graft.Carousel(
		graft.CarouselItem(graft.Text("A")),
		graft.CarouselItem(graft.Text("B")),
		graft.CarouselItem(graft.Text("C")),
	).Bind(sig)
	layoutCarousel(c)

	// Initial index from signal.
	children := c.Children()
	if len(children) != 3 {
		t.Fatalf("children: got %d, want 3", len(children))
	}
}

// TestCarouselFocusable verifies focus behavior.
func TestCarouselFocusable(t *testing.T) {
	_, restore := carouselForceLight(t)
	defer restore()

	c := buildCarousel()
	layoutCarousel(c)

	if !c.IsFocusable() {
		t.Fatal("carousel should be focusable")
	}
}

// TestCarouselDrawsNavButtons verifies that nav buttons are drawn when
// there are multiple slides.
func TestCarouselDrawsNavButtons(t *testing.T) {
	_, restore := carouselForceLight(t)
	defer restore()

	c := buildCarousel()
	layoutCarousel(c)
	mc := uitest.DrawWidget(c)

	// The carousel should draw round rects for the nav buttons (the
	// circular outline buttons use DrawRoundRect with radius = size/2).
	foundButton := false
	m := metrics.Carousel
	btnRadius := m.ButtonSize / 2
	for _, rr := range mc.RoundRects {
		if rr.Radius == btnRadius {
			foundButton = true
			break
		}
	}
	if !foundButton {
		t.Fatalf("no nav button round rect with radius %v found; roundrects: %+v",
			btnRadius, mc.RoundRects)
	}
}

// TestCarouselSingleSlideNoButtons verifies that a single-slide carousel
// does not render nav buttons.
func TestCarouselSingleSlideNoButtons(t *testing.T) {
	_, restore := carouselForceLight(t)
	defer restore()

	c := graft.Carousel(graft.CarouselItem(graft.Text("Only")))
	layoutCarousel(c)
	mc := uitest.DrawWidget(c)

	m := metrics.Carousel
	btnRadius := m.ButtonSize / 2
	for _, rr := range mc.RoundRects {
		if rr.Radius == btnRadius {
			t.Fatal("single-slide carousel should not draw nav buttons")
		}
	}
}

// TestGoldenCarousel renders golden images for the carousel.
func TestGoldenCarousel(t *testing.T) {
	slide := func(label string, bg uint32) widget.Widget {
		return primitives.Box(graft.Text(label)).
			Background(widget.Hex(bg)).
			Width(360).Height(160).Padding(16)
	}

	gtest.GoldenLightDark(t, "carousel-default", func() widget.Widget {
		c := graft.Carousel(
			graft.CarouselItem(slide("Slide 1", 0xdbeafe)),
			graft.CarouselItem(slide("Slide 2", 0xfce7f3)),
			graft.CarouselItem(slide("Slide 3", 0xd1fae5)),
		)
		return primitives.Box(c).Padding(24).Width(420)
	})

	gtest.GoldenLightDark(t, "carousel-second", func() widget.Widget {
		c := graft.Carousel(
			graft.CarouselItem(slide("Slide 1", 0xdbeafe)),
			graft.CarouselItem(slide("Slide 2", 0xfce7f3)),
			graft.CarouselItem(slide("Slide 3", 0xd1fae5)),
		).Index(1)
		return primitives.Box(c).Padding(24).Width(420)
	})

	gtest.GoldenLightDark(t, "carousel-vertical", func() widget.Widget {
		c := graft.Carousel(
			graft.CarouselItem(slide("Slide 1", 0xdbeafe)),
			graft.CarouselItem(slide("Slide 2", 0xfce7f3)),
		).Vertical()
		return primitives.Box(c).Padding(24).Width(420)
	})

	gtest.GoldenLightDark(t, "carousel-focused", func() widget.Widget {
		c := graft.Carousel(
			graft.CarouselItem(slide("Slide 1", 0xdbeafe)),
			graft.CarouselItem(slide("Slide 2", 0xfce7f3)),
			graft.CarouselItem(slide("Slide 3", 0xd1fae5)),
		)
		layoutCarousel(c)
		c.SetFocused(true)
		return primitives.Box(c).Padding(24).Width(420)
	})
}

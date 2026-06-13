package graft_test

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// emptyDemo builds the canonical empty state used across Empty tests.
func emptyDemo() *graft.EmptyWidget {
	return graft.Empty(
		graft.EmptyHeader(
			graft.EmptyMedia(icons.Search),
			graft.EmptyTitle("No results found"),
			graft.EmptyDescription("Try adjusting your search to find what you are looking for."),
		),
		graft.EmptyContent(
			graft.Button("Clear search").Outline(),
		),
	)
}

// TestEmptyMediaTileSpec pins the size-10 tile and centered size-6 icon.
func TestEmptyMediaTileSpec(t *testing.T) {
	th := forceLightMode(t)
	tok := th.Active()

	m := graft.EmptyMedia(icons.Search)
	size := uitest.LayoutWidget(m, 200, 200)
	if size.Width != metrics.EmptyMediaSize {
		t.Fatalf("media width = %v, want %v", size.Width, metrics.EmptyMediaSize)
	}
	if size.Height != metrics.EmptyMediaSize+metrics.EmptyMediaMarginBottom {
		t.Fatalf("media height = %v, want %v", size.Height, metrics.EmptyMediaSize+metrics.EmptyMediaMarginBottom)
	}

	mc := uitest.DrawWidget(m)
	// The muted tile is a rounded rect filled with the Muted token.
	var foundTile bool
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Muted && rr.Bounds.Width() == metrics.EmptyMediaSize {
			foundTile = true
		}
	}
	if !foundTile {
		t.Errorf("media tile not drawn as a %vpx muted rounded rect", metrics.EmptyMediaSize)
	}
}

// TestEmptyRendersTexts verifies the title and description render.
func TestEmptyRendersTexts(t *testing.T) {
	forceLightMode(t)

	e := emptyDemo()
	uitest.LayoutWidget(e, 600, 600)
	mc := uitest.DrawWidget(e)
	if !drewAnyText(mc, "No results found") {
		t.Errorf("title not drawn")
	}
	if !drewAnyText(mc, "Try adjusting your search to find what you are looking for.") {
		t.Errorf("description not drawn")
	}
}

// TestEmptyDrawsDashedBorder verifies the dashed border emits line segments in
// the Border token.
func TestEmptyDrawsDashedBorder(t *testing.T) {
	th := forceLightMode(t)
	tok := th.Active()

	e := emptyDemo()
	uitest.LayoutWidget(e, 600, 600)
	mc := uitest.DrawWidget(e)
	var dashes int
	for _, ln := range mc.Lines {
		if ln.Color == tok.Border {
			dashes++
		}
	}
	if dashes < 4 {
		t.Errorf("dashed border drew %d border-colored segments, want many", dashes)
	}
}

func TestGoldenEmpty(t *testing.T) {
	gtest.GoldenLightDark(t, "empty-state", func() widget.Widget {
		return primitives.Box(emptyDemo()).Width(440).Padding(24)
	})
}

package graft_test

import (
	"math"
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// chartApproxEq reports whether a and b are within eps.
func chartApproxEq(a, b float32) bool {
	return math.Abs(float64(a-b)) <= 1e-3
}

// TestChartLineSeriesColors verifies each line series strokes in its positional
// chart-palette token at the 2px Recharts stroke width.
func TestChartLineSeriesColors(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()

	c := graft.LineChart(
		graft.LineSeries("desktop", 100, 200, 150),
		graft.LineSeries("mobile", 50, 80, 120),
	).Legend(false).Grid(false).YAxis(false)
	c.Layout(nil, fixedWidthLoose(400))
	canvas := uitest.DrawWidget(c)

	if len(canvas.Lines) == 0 {
		t.Fatal("line chart drew no lines")
	}
	// Each series has 3 points -> 2 segments. Collect colors used.
	seen := map[widget.Color]int{}
	for _, l := range canvas.Lines {
		if !chartApproxEq(l.StrokeWidth, metrics.Chart.LineWidth) {
			t.Errorf("line stroke width = %v, want %v", l.StrokeWidth, metrics.Chart.LineWidth)
		}
		seen[l.Color]++
	}
	if seen[tok.Chart[0]] != 2 {
		t.Errorf("series 0 drew %d segments in Chart[0], want 2", seen[tok.Chart[0]])
	}
	if seen[tok.Chart[1]] != 2 {
		t.Errorf("series 1 drew %d segments in Chart[1], want 2", seen[tok.Chart[1]])
	}
}

// TestChartGridlinesUseBorderToken verifies horizontal gridlines use the Border
// token at 50% alpha (stroke-border/50) and there are GridDivisions+1 of them.
func TestChartGridlinesUseBorderToken(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()
	wantGrid := tok.Border
	wantGrid.A *= metrics.Chart.GridAlpha

	c := graft.LineChart(graft.LineSeries("a", 1, 2, 3)).Legend(false).YAxis(false)
	c.Layout(nil, fixedWidthLoose(400))
	canvas := uitest.DrawWidget(c)

	gridCount := 0
	for _, l := range canvas.Lines {
		if l.Color == wantGrid && chartApproxEq(l.StrokeWidth, metrics.Chart.GridLineWidth) {
			// Gridlines are horizontal (same Y at both ends).
			if chartApproxEq(l.From.Y, l.To.Y) {
				gridCount++
			}
		}
	}
	if gridCount != metrics.Chart.GridDivisions+1 {
		t.Errorf("drew %d gridlines, want %d", gridCount, metrics.Chart.GridDivisions+1)
	}
}

// TestChartBarsUseChartTokensAndRadius verifies bars fill in their chart token
// at the rounded radius and sit on the zero baseline.
func TestChartBarsUseChartTokensAndRadius(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()

	c := graft.BarChart(graft.BarSeries("desktop", 100, 200, 50)).
		Categories("a", "b", "c").Legend(false)
	c.Layout(nil, fixedWidthLoose(400))
	canvas := uitest.DrawWidget(c)

	bars := 0
	for _, r := range canvas.RoundRects {
		if r.Color == tok.Chart[0] {
			bars++
			if r.Radius > metrics.Chart.BarRadius+1e-3 {
				t.Errorf("bar radius = %v, want <= %v", r.Radius, metrics.Chart.BarRadius)
			}
		}
	}
	if bars != 3 {
		t.Errorf("drew %d bars in Chart[0], want 3", bars)
	}

	// The tallest bar (value 200) should be taller than the shortest (50).
	var tall, short float32
	for _, r := range canvas.RoundRects {
		if r.Color != tok.Chart[0] {
			continue
		}
		if r.Bounds.Height() > tall {
			tall = r.Bounds.Height()
		}
	}
	short = tall
	for _, r := range canvas.RoundRects {
		if r.Color != tok.Chart[0] {
			continue
		}
		if r.Bounds.Height() < short {
			short = r.Bounds.Height()
		}
	}
	if !(tall > short) {
		t.Errorf("expected varying bar heights, got tall=%v short=%v", tall, short)
	}
}

// TestChartAreaFillAndLine verifies an area series draws both a translucent fill
// (DrawRect strips at AreaFillAlpha of the chart token) and the line stroke.
func TestChartAreaFillAndLine(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()
	wantFill := tok.Chart[0]
	wantFill.A *= metrics.Chart.AreaFillAlpha

	c := graft.AreaChart(graft.AreaSeries("desktop", 100, 200, 150, 180)).
		Legend(false).Grid(false).YAxis(false)
	c.Layout(nil, fixedWidthLoose(400))
	canvas := uitest.DrawWidget(c)

	fillStrips := 0
	for _, r := range canvas.Rects {
		if r.Color == wantFill {
			fillStrips++
		}
	}
	if fillStrips == 0 {
		t.Error("area chart drew no translucent fill strips")
	}
	lineSegs := 0
	for _, l := range canvas.Lines {
		if l.Color == tok.Chart[0] {
			lineSegs++
		}
	}
	if lineSegs != 3 { // 4 points -> 3 segments
		t.Errorf("area chart drew %d line segments, want 3", lineSegs)
	}
}

// TestChartLegendSwatches verifies the legend draws one rounded swatch per
// series in its chart token at the legend swatch size.
func TestChartLegendSwatches(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()

	c := graft.LineChart(
		graft.LineSeries("desktop", 1, 2),
		graft.LineSeries("mobile", 3, 4),
	).Grid(false).YAxis(false)
	c.Layout(nil, fixedWidthLoose(400))
	canvas := uitest.DrawWidget(c)

	swatches := map[widget.Color]int{}
	for _, r := range canvas.RoundRects {
		if chartApproxEq(r.Bounds.Width(), metrics.Chart.LegendSwatch) &&
			chartApproxEq(r.Bounds.Height(), metrics.Chart.LegendSwatch) &&
			chartApproxEq(r.Radius, metrics.Chart.LegendSwatchRadius) {
			swatches[r.Color]++
		}
	}
	if swatches[tok.Chart[0]] != 1 || swatches[tok.Chart[1]] != 1 {
		t.Errorf("legend swatches = %v, want one each for Chart[0] and Chart[1]", swatches)
	}
}

// TestChartContainerConfigOverrides verifies ChartContainer applies labels and
// color indices from a ChartConfig onto the chart series.
func TestChartContainerConfigOverrides(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()

	idx := 3
	cfg := graft.ChartConfig{
		"desktop": {Label: "Desktop", ColorIndex: &idx},
	}
	c := graft.ChartContainer(cfg, graft.LineChart(graft.LineSeries("desktop", 1, 2, 3)).
		Grid(false).YAxis(false).Legend(false))
	c.Layout(nil, fixedWidthLoose(400))
	canvas := uitest.DrawWidget(c)

	got := 0
	for _, l := range canvas.Lines {
		if l.Color == tok.Chart[3] {
			got++
		}
	}
	if got != 2 {
		t.Errorf("config ColorIndex=3 drew %d segments in Chart[3], want 2", got)
	}
}

// TestChartLayoutAspectVideo verifies the default chart height follows the
// aspect-video ratio plus the legend band.
func TestChartLayoutAspectVideo(t *testing.T) {
	alertForceLight(t)
	c := graft.LineChart(graft.LineSeries("a", 1, 2, 3))
	size := c.Layout(nil, fixedWidthLoose(450))
	wantH := 450/metrics.Chart.AspectRatio + metrics.Chart.LegendTopGap + metrics.Chart.LegendHeight
	if !chartApproxEq(size.Height, wantH) {
		t.Errorf("chart height = %v, want %v (aspect-video + legend)", size.Height, wantH)
	}
}

// TestGoldenChartLine renders a two-series line chart in light and dark.
func TestGoldenChartLine(t *testing.T) {
	gtest.GoldenLightDark(t, "chart-line", func() widget.Widget {
		return primitives.VBox(
			graft.LineChart(
				graft.LineSeries("desktop", 186, 305, 237, 273, 209, 214).Dots(true),
				graft.LineSeries("mobile", 80, 200, 120, 190, 130, 140).Dots(true),
			).Categories("Jan", "Feb", "Mar", "Apr", "May", "Jun").W(460),
		).Padding(24)
	})
}

// TestGoldenChartBar renders a single-series bar chart in light and dark.
func TestGoldenChartBar(t *testing.T) {
	gtest.GoldenLightDark(t, "chart-bar", func() widget.Widget {
		return primitives.VBox(
			graft.BarChart(
				graft.BarSeries("desktop", 186, 305, 237, 273, 209, 214),
			).Categories("Jan", "Feb", "Mar", "Apr", "May", "Jun").W(460),
		).Padding(24)
	})
}

// TestGoldenChartArea renders a single-series area chart in light and dark.
func TestGoldenChartArea(t *testing.T) {
	gtest.GoldenLightDark(t, "chart-area", func() widget.Widget {
		return primitives.VBox(
			graft.AreaChart(
				graft.AreaSeries("desktop", 186, 305, 237, 273, 209, 214),
			).Categories("Jan", "Feb", "Mar", "Apr", "May", "Jun").W(460),
		).Padding(24)
	})
}

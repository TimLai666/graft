package graft_test

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
)

// TestGoldenChartPie renders a donut pie chart (light + dark).
func TestGoldenChartPie(t *testing.T) {
	gtest.GoldenLightDark(t, "chart-pie", func() widget.Widget {
		return primitives.VBox(
			graft.PieChart(
				graft.PieSlice("chrome", 275),
				graft.PieSlice("safari", 200),
				graft.PieSlice("firefox", 187),
				graft.PieSlice("edge", 173),
				graft.PieSlice("other", 90),
			).Donut().CenterLabel("925").W(360),
		).Padding(24)
	})
}

// TestGoldenChartRadial renders a multi-ring radial-bar chart (light + dark).
func TestGoldenChartRadial(t *testing.T) {
	gtest.GoldenLightDark(t, "chart-radial", func() widget.Widget {
		return primitives.VBox(
			graft.RadialChart(
				graft.RadialValue("desktop", 1260),
				graft.RadialValue("mobile", 570),
				graft.RadialValue("tablet", 320),
			).W(360),
		).Padding(24)
	})
}

// TestGoldenChartRadar renders a radar (spider) chart (light + dark).
func TestGoldenChartRadar(t *testing.T) {
	gtest.GoldenLightDark(t, "chart-radar", func() widget.Widget {
		return primitives.VBox(
			graft.RadarChart(
				graft.RadarSeries("desktop", 186, 305, 237, 273, 209, 214),
				graft.RadarSeries("mobile", 80, 200, 120, 190, 130, 140),
			).Categories("Jan", "Feb", "Mar", "Apr", "May", "Jun").W(360),
		).Padding(24)
	})
}

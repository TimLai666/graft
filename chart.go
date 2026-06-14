package graft

import (
	"fmt"
	"math"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// ChartKind selects how a chart series is rendered.
type ChartKind uint8

// Chart series render kinds.
const (
	// ChartLine renders a series as a polyline (optionally with dots).
	ChartLine ChartKind = iota
	// ChartBar renders a series as vertical bars.
	ChartBar
	// ChartArea renders a series as a polyline with a translucent fill below.
	ChartArea
)

// ChartSeries is one named data series. It mirrors a shadcn ChartConfig entry
// (a label + a chart color) plus the data points themselves. The color is NOT
// stored as a token: series take their color from the active theme's
// Chart[0..4] tokens by their position in the chart, so light/dark switches
// repaint correctly. Override with Color/ColorIndex.
type ChartSeries struct {
	key    string
	label  string
	values []float64
	kind   ChartKind
	dots   bool

	// colorIdx, when >= 0, pins the chart-token index (0..4). -1 means use the
	// series' positional index in the chart.
	colorIdx int
	// fixed, when non-nil, pins an explicit color (escape hatch).
	fixed *widget.Color
}

// LineSeries creates a line series named key with the given Y values. The
// label defaults to key; override with Label.
func LineSeries(key string, values ...float64) ChartSeries {
	return ChartSeries{key: key, label: key, values: values, kind: ChartLine, colorIdx: -1}
}

// BarSeries creates a vertical-bar series named key with the given Y values.
func BarSeries(key string, values ...float64) ChartSeries {
	return ChartSeries{key: key, label: key, values: values, kind: ChartBar, colorIdx: -1}
}

// AreaSeries creates an area series (line + translucent fill) named key.
func AreaSeries(key string, values ...float64) ChartSeries {
	return ChartSeries{key: key, label: key, values: values, kind: ChartArea, colorIdx: -1}
}

// Label sets the human-readable series label shown in the legend.
func (s ChartSeries) Label(label string) ChartSeries {
	s.label = label
	return s
}

// Dots toggles point markers on a line/area series.
func (s ChartSeries) Dots(v bool) ChartSeries {
	s.dots = v
	return s
}

// ColorIndex pins the chart-token index (0..4) for this series instead of its
// positional index.
func (s ChartSeries) ColorIndex(i int) ChartSeries {
	s.colorIdx = i
	return s
}

// Color pins an explicit color for this series (escape hatch; does not adapt
// to the chart palette).
func (s ChartSeries) Color(c widget.Color) ChartSeries {
	s.fixed = &c
	return s
}

// resolveColor returns the series color for the active token set, honoring an
// explicit override, then a pinned index, then the positional index.
func (s ChartSeries) resolveColor(tok *theme.Tokens, pos int) widget.Color {
	if s.fixed != nil {
		return *s.fixed
	}
	idx := s.colorIdx
	if idx < 0 {
		idx = pos
	}
	if idx < 0 {
		idx = 0
	}
	return tok.Chart[idx%len(tok.Chart)]
}

// ChartWidget is graft's native chart: axes (muted gridlines + value/category
// labels), a legend, and one rendering per series in its chart-palette color.
// It stands in for shadcn's Recharts <ChartContainer> (DESIGN.md §4; the
// pixel constants live in metrics/chart.go).
//
// OWNED widget: there is no fitting gogpu/ui core widget. core/linechart is a
// right-aligned streaming chart (MaxPoints-based scroll) and cannot express
// the static, multi-type, category-axis shadcn look, so the painters.LineChart
// slot is intentionally left unused and the chart is drawn directly on the
// canvas. All colors resolve from the active token set inside Draw, so mode
// switches repaint without rebuilding the tree.
type ChartWidget struct {
	widget.WidgetBase

	series     []ChartSeries
	categories []string // X-axis category labels (one per data index)
	showGrid   bool
	showYAxis  bool
	showLegend bool

	// Polar shapes (pie/radial/radar) — zero value shapeCartesian keeps the
	// line/bar/area renderer. See chart_polar.go.
	shape       chartShape
	slices      []ChartSlice // pie/radial values
	donutInner  float32      // pie donut inner-radius fraction; <0 = full pie
	polarMax    float64      // radial full-sweep value; NaN = max(values)
	centerLabel string       // pie/donut/radial center text

	width  float32
	height float32 // 0 = derive from width via aspect-video

	theme *theme.Theme
}

// Chart creates a chart from one or more series. Mixed kinds are allowed
// (e.g. bars + a line). Width defaults to a card-friendly size and height
// follows the aspect-video ratio; pin them with W/H.
//
//	graft.Chart(
//	    graft.LineSeries("desktop", 186, 305, 237, 273, 209, 214),
//	    graft.LineSeries("mobile", 80, 200, 120, 190, 130, 140),
//	).Categories("Jan", "Feb", "Mar", "Apr", "May", "Jun")
func Chart(series ...ChartSeries) *ChartWidget {
	c := &ChartWidget{
		series:     series,
		showGrid:   true,
		showYAxis:  true,
		showLegend: true,
		width:      metrics.Chart.DefaultWidth,
		theme:      CurrentTheme(),
	}
	c.SetVisible(true)
	c.SetEnabled(true)
	return c
}

// LineChart creates a line chart from the given series, forcing each to the
// line kind. categories is an optional X-axis label list passed inline.
func LineChart(series ...ChartSeries) *ChartWidget {
	for i := range series {
		series[i].kind = ChartLine
	}
	return Chart(series...)
}

// BarChart creates a bar chart, forcing each series to the bar kind.
func BarChart(series ...ChartSeries) *ChartWidget {
	for i := range series {
		series[i].kind = ChartBar
	}
	return Chart(series...)
}

// AreaChart creates an area chart, forcing each series to the area kind.
func AreaChart(series ...ChartSeries) *ChartWidget {
	for i := range series {
		series[i].kind = ChartArea
	}
	return Chart(series...)
}

// ChartContainer establishes a chart from a ChartConfig (series colors/labels)
// and a chart widget, matching shadcn's <ChartContainer config>. The config
// supplies labels and color indices; the child carries the data. It returns
// the configured chart so it composes directly into a layout.
func ChartContainer(config ChartConfig, chart *ChartWidget) *ChartWidget {
	if chart == nil {
		return Chart()
	}
	for i := range chart.series {
		if entry, ok := config[chart.series[i].key]; ok {
			if entry.Label != "" {
				chart.series[i].label = entry.Label
			}
			if entry.ColorIndex != nil {
				chart.series[i].colorIdx = *entry.ColorIndex
			}
			if entry.Color != nil {
				c := *entry.Color
				chart.series[i].fixed = &c
			}
		}
	}
	return chart
}

// ChartConfigEntry is one ChartConfig entry: a legend label and a color source
// (a chart-token index or an explicit color). Mirrors shadcn's
// ChartConfig[key] = { label, color }.
type ChartConfigEntry struct {
	// Label is the human-readable series label.
	Label string
	// ColorIndex, when set, pins the chart-token index (0..4).
	ColorIndex *int
	// Color, when set, pins an explicit color (escape hatch).
	Color *widget.Color
}

// ChartConfig maps a series key to its display config, mirroring shadcn's
// ChartConfig object.
type ChartConfig map[string]ChartConfigEntry

// Categories sets the X-axis category labels (one per data index).
func (c *ChartWidget) Categories(labels ...string) *ChartWidget {
	c.categories = labels
	return c
}

// W pins the chart width in px.
func (c *ChartWidget) W(px float32) *ChartWidget {
	c.width = px
	return c
}

// H pins the chart height in px (otherwise derived from the aspect ratio).
func (c *ChartWidget) H(px float32) *ChartWidget {
	c.height = px
	return c
}

// Grid toggles the horizontal gridlines.
func (c *ChartWidget) Grid(v bool) *ChartWidget {
	c.showGrid = v
	return c
}

// YAxis toggles the Y-axis value labels (and the left inset for them).
func (c *ChartWidget) YAxis(v bool) *ChartWidget {
	c.showYAxis = v
	return c
}

// Legend toggles the series legend row under the chart.
func (c *ChartWidget) Legend(v bool) *ChartWidget {
	c.showLegend = v
	return c
}

// Dots enables point markers on every line/area series.
func (c *ChartWidget) Dots(v bool) *ChartWidget {
	for i := range c.series {
		if c.series[i].kind != ChartBar {
			c.series[i].dots = v
		}
	}
	return c
}

// Theme pins a specific theme instead of the process-wide current theme.
func (c *ChartWidget) Theme(th *theme.Theme) *ChartWidget {
	c.theme = th
	return c
}

func (c *ChartWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// plotHeight returns the chart's drawn height (the aspect-video height plus the
// legend band).
func (c *ChartWidget) chartHeight(w float32) float32 {
	h := c.height
	if h <= 0 {
		h = w / metrics.Chart.AspectRatio
	}
	if c.showLegend {
		h += metrics.Chart.LegendTopGap + metrics.Chart.LegendHeight
	}
	return h
}

// Layout sizes the chart to its pinned width (or the constraint width) and the
// derived height.
func (c *ChartWidget) Layout(_ widget.Context, cons geometry.Constraints) geometry.Size {
	w := c.width
	if cons.MaxWidth > 0 && cons.MaxWidth < 100000 && cons.MaxWidth < w {
		w = cons.MaxWidth
	}
	if w <= 0 {
		w = metrics.Chart.DefaultWidth
	}
	size := cons.Constrain(geometry.Sz(w, c.chartHeight(w)))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// dataRange returns the number of data points (the longest series) and the
// [min, max] Y range, with the min floored at 0 so bars/areas sit on a zero
// baseline and a small headroom added to the max.
func (c *ChartWidget) dataRange() (points int, yMin, yMax float64) {
	yMin, yMax = 0, math.Inf(-1)
	for _, s := range c.series {
		if len(s.values) > points {
			points = len(s.values)
		}
		for _, v := range s.values {
			if v > yMax {
				yMax = v
			}
			if v < yMin {
				yMin = v
			}
		}
	}
	if math.IsInf(yMax, -1) || yMax <= yMin {
		yMax = yMin + 1
	}
	// Round the top up to a clean gridline step for tidy Y labels.
	yMax = niceCeil(yMax, metrics.Chart.GridDivisions)
	return points, yMin, yMax
}

// plotArea returns the rectangle where series are drawn, inset for axis labels.
func (c *ChartWidget) plotArea(bounds geometry.Rect) geometry.Rect {
	m := metrics.Chart
	left := m.PadLeft
	if !c.showYAxis {
		left = m.PadRight
	}
	plotH := bounds.Height()
	if c.showLegend {
		plotH -= m.LegendTopGap + m.LegendHeight
	}
	return geometry.NewRect(
		bounds.Min.X+left,
		bounds.Min.Y+m.PadTop,
		bounds.Width()-left-m.PadRight,
		plotH-m.PadTop-m.PadBottom,
	)
}

// Draw paints the gridlines, axis labels, every series, and the legend,
// resolving all colors from the active token set.
func (c *ChartWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.resolvedTheme()
	tok := th.Active()
	bounds := c.Bounds()
	m := metrics.Chart

	// Polar shapes have their own renderers (chart_polar.go); the cartesian
	// grid/axis/series path below is skipped for them.
	if c.shape != shapeCartesian {
		switch c.shape {
		case shapePie:
			c.drawPie(canvas, th, tok, bounds)
		case shapeRadial:
			c.drawRadial(canvas, th, tok, bounds)
		case shapeRadar:
			c.drawRadar(canvas, th, tok, bounds)
		}
		if c.showLegend {
			c.drawLegend(canvas, th, tok, bounds)
		}
		return
	}

	plot := c.plotArea(bounds)
	if plot.Width() <= 0 || plot.Height() <= 0 {
		return
	}
	points, yMin, yMax := c.dataRange()
	yRange := yMax - yMin

	gridColor := draw.Alpha(tok.Border, m.GridAlpha)

	// Horizontal gridlines + Y-axis value labels.
	if c.showGrid || c.showYAxis {
		for i := 0; i <= m.GridDivisions; i++ {
			t := float32(i) / float32(m.GridDivisions)
			y := plot.Max.Y - t*plot.Height()
			if c.showGrid {
				canvas.DrawLine(
					geometry.Pt(plot.Min.X, y),
					geometry.Pt(plot.Max.X, y),
					gridColor, m.GridLineWidth,
				)
			}
			if c.showYAxis {
				value := yMin + float64(t)*yRange
				labelRect := geometry.NewRect(
					bounds.Min.X,
					y-m.AxisLabelLineHeight/2,
					m.PadLeft-6,
					m.AxisLabelLineHeight,
				)
				c.drawLabel(canvas, th, formatAxisValue(value), labelRect,
					tok.MutedForeground, widget.TextAlignRight)
			}
		}
	}

	// Series. Bars are grouped per category; lines/areas are drawn at the
	// category centers so they align with the bars and X labels.
	barSeries := c.barSeriesCount()
	barIdx := 0
	for pos, s := range c.series {
		col := s.resolveColor(tok, pos)
		switch s.kind {
		case ChartBar:
			c.drawBars(canvas, plot, s, col, points, yMin, yRange, barIdx, barSeries)
			barIdx++
		case ChartArea:
			c.drawArea(canvas, plot, s, col, points, yMin, yRange, tok)
			c.drawLine(canvas, plot, s, col, points, yMin, yRange, tok)
		default: // ChartLine
			c.drawLine(canvas, plot, s, col, points, yMin, yRange, tok)
		}
	}

	// X-axis category labels under the plot.
	if len(c.categories) > 0 && points > 0 {
		for i := 0; i < points && i < len(c.categories); i++ {
			x := c.categoryCenterX(plot, i, points)
			labelRect := geometry.NewRect(
				x-plot.Width()/float32(points)/2,
				plot.Max.Y+m.PadBottom-m.AxisLabelLineHeight-2,
				plot.Width()/float32(points),
				m.AxisLabelLineHeight,
			)
			c.drawLabel(canvas, th, c.categories[i], labelRect,
				tok.MutedForeground, widget.TextAlignCenter)
		}
	}

	// Legend row.
	if c.showLegend {
		c.drawLegend(canvas, th, tok, bounds)
	}
}

// categoryCenterX returns the horizontal center of category i across points
// equally spaced category slots.
func (c *ChartWidget) categoryCenterX(plot geometry.Rect, i, points int) float32 {
	if points <= 0 {
		return plot.Center().X
	}
	if points == 1 {
		return plot.Center().X
	}
	slot := plot.Width() / float32(points)
	return plot.Min.X + slot*(float32(i)+0.5)
}

// pointX returns the X position of data index i for line/area series. With
// categories the points sit at category centers; without, they span the full
// plot width edge-to-edge (Recharts' default line layout).
func (c *ChartWidget) pointX(plot geometry.Rect, i, points int) float32 {
	if len(c.categories) > 0 {
		return c.categoryCenterX(plot, i, points)
	}
	if points <= 1 {
		return plot.Min.X
	}
	return plot.Min.X + plot.Width()*float32(i)/float32(points-1)
}

func (c *ChartWidget) barSeriesCount() int {
	n := 0
	for _, s := range c.series {
		if s.kind == ChartBar {
			n++
		}
	}
	return n
}

// yForValue maps a data value to a Y pixel inside the plot (max at top).
func yForValue(value, yMin, yRange float64, plot geometry.Rect) float32 {
	if yRange <= 0 {
		return plot.Max.Y
	}
	t := (value - yMin) / yRange
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return plot.Max.Y - float32(t)*plot.Height()
}

// drawLine strokes a series polyline and optional dots.
func (c *ChartWidget) drawLine(canvas widget.Canvas, plot geometry.Rect, s ChartSeries, col widget.Color, points int, yMin, yRange float64, tok *theme.Tokens) {
	m := metrics.Chart
	n := len(s.values)
	if n == 0 {
		return
	}
	xs := make([]float32, n)
	ys := make([]float32, n)
	for i, v := range s.values {
		xs[i] = c.pointX(plot, i, points)
		ys[i] = yForValue(v, yMin, yRange, plot)
	}
	for i := 1; i < n; i++ {
		canvas.DrawLine(geometry.Pt(xs[i-1], ys[i-1]), geometry.Pt(xs[i], ys[i]), col, m.LineWidth)
	}
	if s.dots {
		for i := 0; i < n; i++ {
			center := geometry.Pt(xs[i], ys[i])
			// White (Background-token) halo, then the colored dot.
			canvas.DrawCircle(center, m.DotRadius+m.DotStrokeWidth/2, tok.Background)
			canvas.DrawCircle(center, m.DotRadius, col)
		}
	}
}

// drawArea fills the band between the line and the baseline with the series
// color at reduced alpha. With no polygon-fill primitive on the base canvas,
// it sweeps contiguous ~1px columns across the plot, interpolating the line
// height at each column. Columns are stepped (not overlapped) so the
// translucent fill stays a single flat alpha with no double-darkened seams.
func (c *ChartWidget) drawArea(canvas widget.Canvas, plot geometry.Rect, s ChartSeries, col widget.Color, points int, yMin, yRange float64, tok *theme.Tokens) {
	n := len(s.values)
	if n < 2 {
		return
	}
	fill := draw.Alpha(col, metrics.Chart.AreaFillAlpha)
	baseY := plot.Max.Y

	xStart := c.pointX(plot, 0, points)
	xEnd := c.pointX(plot, n-1, points)
	span := xEnd - xStart
	if span <= 0 {
		return
	}
	// One column per device pixel; +1 width per column makes adjacent columns
	// abut exactly (no gaps) without overlapping into the next.
	const colW float32 = 1
	cols := int(math.Ceil(float64(span / colW)))
	for k := 0; k < cols; k++ {
		x := xStart + float32(k)*colW
		topY := c.areaTopAt(plot, s, points, yMin, yRange, x)
		h := baseY - topY
		if h <= 0 {
			continue
		}
		canvas.DrawRect(geometry.NewRect(x, topY, colW+1, h), fill)
	}
}

// areaTopAt returns the interpolated line Y at pixel x for an area series.
func (c *ChartWidget) areaTopAt(plot geometry.Rect, s ChartSeries, points int, yMin, yRange float64, x float32) float32 {
	n := len(s.values)
	for i := 1; i < n; i++ {
		x0 := c.pointX(plot, i-1, points)
		x1 := c.pointX(plot, i, points)
		if x < x0 || x > x1 || x1 == x0 {
			continue
		}
		y0 := yForValue(s.values[i-1], yMin, yRange, plot)
		y1 := yForValue(s.values[i], yMin, yRange, plot)
		t := (x - x0) / (x1 - x0)
		return y0 + (y1-y0)*t
	}
	return yForValue(s.values[n-1], yMin, yRange, plot)
}

// drawBars draws one series' bars, offset within each category group so
// multiple bar series sit side by side.
func (c *ChartWidget) drawBars(canvas widget.Canvas, plot geometry.Rect, s ChartSeries, col widget.Color, points int, yMin, yRange float64, barIdx, barSeries int) {
	if points <= 0 || barSeries <= 0 {
		return
	}
	m := metrics.Chart
	slot := plot.Width() / float32(points)
	groupW := slot * (1 - m.BarGroupGap)
	totalInner := m.BarInnerGap * float32(barSeries-1)
	barW := (groupW - totalInner) / float32(barSeries)
	if barW < 1 {
		barW = 1
	}
	radius := m.BarRadius
	if barW/2 < radius {
		radius = barW / 2
	}
	for i, v := range s.values {
		groupCenter := plot.Min.X + slot*(float32(i)+0.5)
		groupLeft := groupCenter - groupW/2
		x := groupLeft + float32(barIdx)*(barW+m.BarInnerGap)
		topY := yForValue(v, yMin, yRange, plot)
		h := plot.Max.Y - topY
		if h <= 0 {
			continue
		}
		canvas.DrawRoundRect(geometry.NewRect(x, topY, barW, h), col, radius)
	}
}

// drawLegend renders the swatch+label row centered under the chart. Entries
// come from the series (cartesian/radar) or the slices (pie/radial).
func (c *ChartWidget) drawLegend(canvas widget.Canvas, th *theme.Theme, tok *theme.Tokens, bounds geometry.Rect) {
	m := metrics.Chart
	family := fonts.Family(m.AxisLabelWeight)

	// Unify legend entries across cartesian series and polar slices.
	var labels []string
	var cols []widget.Color
	if len(c.slices) > 0 {
		for pos, s := range c.slices {
			labels = append(labels, s.label)
			cols = append(cols, s.resolveColor(tok, pos))
		}
	} else {
		for pos, s := range c.series {
			labels = append(labels, s.label)
			cols = append(cols, s.resolveColor(tok, pos))
		}
	}
	if len(labels) == 0 {
		return
	}

	// Measure total width to center the row.
	var total float32
	widths := make([]float32, len(labels))
	for i, lbl := range labels {
		lw := canvas.MeasureText(lbl, m.AxisLabelFontSize, false)
		widths[i] = lw
		total += m.LegendSwatch + m.LegendSwatchTextGap + lw
	}
	if len(labels) > 1 {
		total += m.LegendGap * float32(len(labels)-1)
	}

	rowY := bounds.Max.Y - m.LegendHeight
	x := bounds.Min.X + (bounds.Width()-total)/2
	if x < bounds.Min.X {
		x = bounds.Min.X
	}
	cy := rowY + m.LegendHeight/2

	for i, lbl := range labels {
		sw := geometry.NewRect(x, cy-m.LegendSwatch/2, m.LegendSwatch, m.LegendSwatch)
		canvas.DrawRoundRect(sw, cols[i], m.LegendSwatchRadius)
		x += m.LegendSwatch + m.LegendSwatchTextGap
		labelRect := geometry.NewRect(x, rowY, widths[i], m.LegendHeight)
		c.drawLabelFamily(canvas, family, lbl, labelRect, tok.MutedForeground, widget.TextAlignLeft)
		x += widths[i] + m.LegendGap
	}
}

// drawLabel draws an axis/tick label resolving the family for the configured
// weight (text-xs, normal weight).
func (c *ChartWidget) drawLabel(canvas widget.Canvas, th *theme.Theme, text string, bounds geometry.Rect, col widget.Color, align widget.TextAlign) {
	c.drawLabelFamily(canvas, fonts.Family(metrics.Chart.AxisLabelWeight), text, bounds, col, align)
}

func (c *ChartWidget) drawLabelFamily(canvas widget.Canvas, family, text string, bounds geometry.Rect, col widget.Color, align widget.TextAlign) {
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(text, bounds, widget.TextStyle{
			FontFamily: family,
			FontSize:   metrics.Chart.AxisLabelFontSize,
			Color:      col,
			Align:      align,
		})
		return
	}
	canvas.DrawText(text, bounds, metrics.Chart.AxisLabelFontSize, col, false, align)
}

// Event ignores all input; the chart is a display element.
func (c *ChartWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; the chart is a leaf that draws everything itself.
func (c *ChartWidget) Children() []widget.Widget { return nil }

// niceCeil rounds top up to a clean multiple so the divisions land on round
// Y-axis values (e.g. 273 with 4 divisions -> 280 -> ticks 0,70,140,210,280).
func niceCeil(top float64, divisions int) float64 {
	if top <= 0 || divisions <= 0 {
		return top
	}
	rough := top / float64(divisions)
	mag := math.Pow(10, math.Floor(math.Log10(rough)))
	if mag <= 0 {
		mag = 1
	}
	norm := rough / mag
	var step float64
	switch {
	case norm <= 1:
		step = 1
	case norm <= 2:
		step = 2
	case norm <= 2.5:
		step = 2.5
	case norm <= 5:
		step = 5
	default:
		step = 10
	}
	step *= mag
	return math.Ceil(top/step) * step
}

// formatAxisValue formats a Y-axis tick value without trailing zeros.
func formatAxisValue(v float64) string {
	if v == math.Trunc(v) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%g", v)
}

// Compile-time interface checks.
var _ widget.Widget = (*ChartWidget)(nil)

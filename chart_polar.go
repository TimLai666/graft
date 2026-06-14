package graft

import (
	"math"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// chartShape selects the chart's coordinate system. Cartesian charts (line,
// bar, area) share one renderer; the polar shapes (pie, radial, radar) each
// have their own. The zero value is cartesian so the existing Chart/LineChart/
// BarChart/AreaChart constructors keep their behavior unchanged.
type chartShape uint8

const (
	shapeCartesian chartShape = iota
	shapePie
	shapeRadial
	shapeRadar
)

// ChartSlice is one labelled value in a pie/donut chart. Each slice takes its
// color from the active theme's Chart[0..4] tokens by its position (so light/
// dark switches repaint), or from an explicit override.
type ChartSlice struct {
	label string
	value float64

	// colorIdx >= 0 pins the chart-token index; -1 uses the positional index.
	colorIdx int
	// fixed, when non-nil, pins an explicit color (escape hatch).
	fixed *widget.Color
}

// PieSlice creates a pie/donut slice with the given label and value. Values are
// rendered proportional to their share of the total.
func PieSlice(label string, value float64) ChartSlice {
	return ChartSlice{label: label, value: value, colorIdx: -1}
}

// ColorIndex pins the chart-token index (0..4) for this slice instead of its
// positional index.
func (s ChartSlice) ColorIndex(i int) ChartSlice {
	s.colorIdx = i
	return s
}

// Color pins an explicit color for this slice (escape hatch).
func (s ChartSlice) Color(c widget.Color) ChartSlice {
	s.fixed = &c
	return s
}

// resolveColor returns the slice color for the active token set.
func (s ChartSlice) resolveColor(tok *theme.Tokens, pos int) widget.Color {
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

// PieChart creates a pie chart from labelled slices. Each slice is an arc filled
// in its positional chart-token color, proportional to value/total. Use Donut to
// cut an inner radius (a donut) and CenterLabel to print text in the hole.
//
//	graft.PieChart(
//	    graft.PieSlice("chrome", 275),
//	    graft.PieSlice("safari", 200),
//	    graft.PieSlice("firefox", 187),
//	).Donut()
func PieChart(slices ...ChartSlice) *ChartWidget {
	c := newPolarChart(shapePie)
	c.slices = slices
	return c
}

// RadialChart creates a radial-bar chart: one concentric arc per value, drawn
// around a common center over a faint background track, each in its positional
// chart-token color. A single value reads as a ring gauge; multiple values stack
// as nested rings (shadcn's "radial" chart). Values map 0..max to a near-full
// sweep, where max defaults to the largest value (override with Max).
//
//	graft.RadialChart(
//	    graft.RadialValue("desktop", 1260),
//	    graft.RadialValue("mobile", 570),
//	)
func RadialChart(values ...ChartSlice) *ChartWidget {
	c := newPolarChart(shapeRadial)
	c.slices = values
	return c
}

// RadialValue creates one ring in a radial-bar chart. It is an alias for
// PieSlice with a name that reads naturally at the call site.
func RadialValue(label string, value float64) ChartSlice {
	return PieSlice(label, value)
}

// RadarChart creates a radar (spider) chart from one or more series, one polygon
// per series connecting its per-category values, stroked and translucently
// filled in its positional chart-token color over a grid of axis spokes and
// concentric rings. Pass the per-axis labels with Categories.
//
//	graft.RadarChart(
//	    graft.RadarSeries("desktop", 186, 305, 237, 273, 209, 214),
//	).Categories("Jan", "Feb", "Mar", "Apr", "May", "Jun")
func RadarChart(series ...ChartSeries) *ChartWidget {
	c := newPolarChart(shapeRadar)
	for i := range series {
		series[i].kind = ChartLine
	}
	c.series = series
	return c
}

// RadarSeries creates a radar series named key with one value per category. It
// is an alias for LineSeries; the radar renderer closes the polygon itself.
func RadarSeries(key string, values ...float64) ChartSeries {
	return LineSeries(key, values...)
}

// newPolarChart builds a ChartWidget configured for a polar shape with the
// shared defaults (no cartesian grid/axis; legend on).
func newPolarChart(shape chartShape) *ChartWidget {
	c := &ChartWidget{
		shape:      shape,
		showGrid:   true, // for radar: the concentric rings + spokes
		showYAxis:  false,
		showLegend: true,
		width:      metrics.Chart.DefaultWidth,
		theme:      CurrentTheme(),
		donutInner: -1, // -1 = full pie unless Donut() is called
		polarMax:   math.NaN(),
	}
	c.SetVisible(true)
	c.SetEnabled(true)
	return c
}

// Donut cuts an inner radius into a pie chart, turning it into a donut. With no
// argument it uses the default inner fraction; pass a fraction (0..1) of the
// outer radius to override.
func (c *ChartWidget) Donut(frac ...float32) *ChartWidget {
	if len(frac) > 0 {
		c.donutInner = frac[0]
	} else {
		c.donutInner = metrics.Chart.PieDonutInner
	}
	return c
}

// CenterLabel sets text drawn in the center of a pie/donut or radial chart
// (typically a total). It only shows when there is a hole to draw it in.
func (c *ChartWidget) CenterLabel(text string) *ChartWidget {
	c.centerLabel = text
	return c
}

// Max pins the value mapped to a full radial-bar sweep (default: the largest
// value). Use it to put several radial charts on a shared scale.
func (c *ChartWidget) Max(v float64) *ChartWidget {
	c.polarMax = v
	return c
}

// polarPlot returns the centered square plot rectangle for a polar chart, its
// center, and the outer radius, inset for the legend band.
func (c *ChartWidget) polarPlot(bounds geometry.Rect) (center geometry.Point, radius float32) {
	m := metrics.Chart
	h := bounds.Height()
	if c.showLegend {
		h -= m.LegendTopGap + m.LegendHeight
	}
	side := bounds.Width()
	if h < side {
		side = h
	}
	cx := bounds.Min.X + bounds.Width()/2
	cy := bounds.Min.Y + h/2
	radius = side/2 - m.PolarPad
	if radius < 0 {
		radius = 0
	}
	return geometry.Pt(cx, cy), radius
}

// fillWedge fills an annular wedge [inner, outer] over the angular span
// [start, start+sweep] using a thick StrokeArc band. The base canvas has no
// filled-wedge primitive, so a stroke of width (outer-inner) centered on the
// mean radius covers the annulus exactly; for inner=0 this fills a full disk
// sector. Butt caps give the wedge straight radial edges.
func fillWedge(canvas widget.Canvas, center geometry.Point, inner, outer float32, start, sweep float64, col widget.Color) {
	if sweep == 0 || outer <= 0 {
		return
	}
	if inner < 0 {
		inner = 0
	}
	mid := (inner + outer) / 2
	width := outer - inner
	if width <= 0 {
		return
	}
	canvas.StrokeArc(center, mid, start, sweep, col, width)
}

// drawPie paints the slices (and an optional center label) of a pie/donut.
func (c *ChartWidget) drawPie(canvas widget.Canvas, th *theme.Theme, tok *theme.Tokens, bounds geometry.Rect) {
	center, outer := c.polarPlot(bounds)
	if outer <= 0 {
		return
	}
	m := metrics.Chart

	var total float64
	for _, s := range c.slices {
		if s.value > 0 {
			total += s.value
		}
	}
	if total <= 0 {
		return
	}

	inner := float32(0)
	if c.donutInner >= 0 {
		inner = outer * c.donutInner
	}

	// Start at the top (12 o'clock) and sweep clockwise, matching Recharts.
	const startAngle = -math.Pi / 2
	gap := m.PieGapAngle
	// Don't apply a gap when there is effectively a single slice.
	visible := 0
	for _, s := range c.slices {
		if s.value > 0 {
			visible++
		}
	}
	if visible <= 1 {
		gap = 0
	}

	angle := startAngle
	for pos, s := range c.slices {
		if s.value <= 0 {
			continue
		}
		frac := s.value / total
		sweep := frac * 2 * math.Pi
		drawSweep := sweep - gap
		if drawSweep <= 0 {
			drawSweep = sweep
		}
		col := s.resolveColor(tok, pos)
		fillWedge(canvas, center, inner, outer, angle+gap/2, drawSweep, col)
		angle += sweep
	}

	// Center label sits in the donut hole.
	if c.centerLabel != "" && inner > 0 {
		labelRect := geometry.NewRect(
			center.X-inner,
			center.Y-m.AxisLabelLineHeight/2,
			inner*2,
			m.AxisLabelLineHeight,
		)
		c.drawLabel(canvas, th, c.centerLabel, labelRect, tok.Foreground, widget.TextAlignCenter)
	}
}

// drawRadial paints concentric radial bars over background track rings.
func (c *ChartWidget) drawRadial(canvas widget.Canvas, th *theme.Theme, tok *theme.Tokens, bounds geometry.Rect) {
	center, polarR := c.polarPlot(bounds)
	if polarR <= 0 {
		return
	}
	m := metrics.Chart

	maxVal := c.polarMax
	if math.IsNaN(maxVal) || maxVal <= 0 {
		maxVal = 0
		for _, s := range c.slices {
			if s.value > maxVal {
				maxVal = s.value
			}
		}
	}
	if maxVal <= 0 {
		return
	}

	track := draw.Alpha(tok.Muted, m.RadialTrackAlpha)
	thickness := m.RadialBarThickness
	step := thickness + m.RadialBarGap

	// Outermost ring radius (centerline of the band).
	ringR := polarR*m.RadialBarRadiusFrac - thickness/2
	// Sweep nearly the full circle (leave a hair so a full value still reads as
	// a ring rather than overlapping its own start cap). Start at the top.
	const startAngle = -math.Pi / 2
	const fullSweep = 2 * math.Pi

	for pos, s := range c.slices {
		if ringR-thickness/2 <= 0 {
			break
		}
		// Background track for this ring.
		canvas.StrokeArc(center, ringR, startAngle, fullSweep, track, thickness)
		// Value arc, clockwise from the top.
		frac := s.value / maxVal
		if frac > 1 {
			frac = 1
		}
		if frac < 0 {
			frac = 0
		}
		if frac > 0 {
			col := s.resolveColor(tok, pos)
			canvas.StrokeArc(center, ringR, startAngle, frac*fullSweep, col, thickness)
		}
		ringR -= step
	}

	if c.centerLabel != "" {
		labelRect := geometry.NewRect(
			center.X-polarR,
			center.Y-m.AxisLabelLineHeight/2,
			polarR*2,
			m.AxisLabelLineHeight,
		)
		c.drawLabel(canvas, th, c.centerLabel, labelRect, tok.Foreground, widget.TextAlignCenter)
	}
}

// radarPoint returns the screen point at the given axis index for value v
// scaled to [0, maxVal] over the radius polarR. Axes are spaced evenly with
// axis 0 at the top, going clockwise.
func radarPoint(center geometry.Point, polarR float32, axis, axes int, v, maxVal float64) geometry.Point {
	if axes <= 0 || maxVal <= 0 {
		return center
	}
	frac := v / maxVal
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	r := polarR * float32(frac)
	angle := -math.Pi/2 + 2*math.Pi*float64(axis)/float64(axes)
	return geometry.Pt(
		center.X+r*float32(math.Cos(angle)),
		center.Y+r*float32(math.Sin(angle)),
	)
}

// drawRadar paints the spider grid (concentric rings + axis spokes), each
// series' closed polygon (translucent fill + stroke + vertex dots), and the
// per-axis category labels.
func (c *ChartWidget) drawRadar(canvas widget.Canvas, th *theme.Theme, tok *theme.Tokens, bounds geometry.Rect) {
	center, polarR := c.polarPlot(bounds)
	if polarR <= 0 {
		return
	}
	m := metrics.Chart

	axes := 0
	for _, s := range c.series {
		if len(s.values) > axes {
			axes = len(s.values)
		}
	}
	if len(c.categories) > axes {
		axes = len(c.categories)
	}
	if axes < 3 {
		return
	}

	var maxVal float64
	for _, s := range c.series {
		for _, v := range s.values {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if maxVal <= 0 {
		maxVal = 1
	}
	maxVal = niceCeil(maxVal, m.RadarRings)

	gridColor := draw.Alpha(tok.Border, m.GridAlpha)

	// Concentric grid rings, drawn as closed polygons through the axis points so
	// they read as the spider web (Recharts' PolarGrid gridType="polygon").
	if c.showGrid {
		for ring := 1; ring <= m.RadarRings; ring++ {
			rv := maxVal * float64(ring) / float64(m.RadarRings)
			prev := radarPoint(center, polarR, 0, axes, rv, maxVal)
			first := prev
			for a := 1; a < axes; a++ {
				p := radarPoint(center, polarR, a, axes, rv, maxVal)
				canvas.DrawLine(prev, p, gridColor, m.RadarSpokeWidth)
				prev = p
			}
			canvas.DrawLine(prev, first, gridColor, m.RadarSpokeWidth)
		}
		// Axis spokes from center to each outer vertex.
		for a := 0; a < axes; a++ {
			outer := radarPoint(center, polarR, a, axes, maxVal, maxVal)
			canvas.DrawLine(center, outer, gridColor, m.RadarSpokeWidth)
		}
	}

	// Series polygons.
	for pos, s := range c.series {
		col := s.resolveColor(tok, pos)
		c.drawRadarSeries(canvas, center, polarR, axes, maxVal, s, col)
	}

	// Per-axis category labels just outside the outer vertices.
	if len(c.categories) > 0 {
		for a := 0; a < axes && a < len(c.categories); a++ {
			lp := radarPoint(center, polarR+m.PolarLabelGap, a, axes, maxVal, maxVal)
			align := widget.TextAlignCenter
			if lp.X < center.X-1 {
				align = widget.TextAlignRight
			} else if lp.X > center.X+1 {
				align = widget.TextAlignLeft
			}
			lw := canvas.MeasureText(c.categories[a], m.AxisLabelFontSize, false)
			rx := lp.X
			switch align {
			case widget.TextAlignRight:
				rx = lp.X - lw
			case widget.TextAlignCenter:
				rx = lp.X - lw/2
			}
			labelRect := geometry.NewRect(rx, lp.Y-m.AxisLabelLineHeight/2, lw, m.AxisLabelLineHeight)
			c.drawLabel(canvas, th, c.categories[a], labelRect, tok.MutedForeground, align)
		}
	}
}

// drawRadarSeries fills, strokes, and dots one radar polygon.
func (c *ChartWidget) drawRadarSeries(canvas widget.Canvas, center geometry.Point, polarR float32, axes int, maxVal float64, s ChartSeries, col widget.Color) {
	m := metrics.Chart
	n := len(s.values)
	if n < 1 {
		return
	}
	pts := make([]geometry.Point, axes)
	for a := 0; a < axes; a++ {
		v := 0.0
		if a < n {
			v = s.values[a]
		}
		pts[a] = radarPoint(center, polarR, a, axes, v, maxVal)
	}

	// Translucent fill: a triangle fan from the center to each polygon edge,
	// approximated as thin radial sweeps so the flat alpha stays single-coverage
	// (no polygon-fill primitive on the base canvas), mirroring drawArea.
	fill := draw.Alpha(col, m.RadarFillAlpha)
	for a := 0; a < axes; a++ {
		fillRadarTriangle(canvas, center, pts[a], pts[(a+1)%axes], fill)
	}

	// Stroke the closed polygon edges.
	for a := 0; a < axes; a++ {
		canvas.DrawLine(pts[a], pts[(a+1)%axes], col, m.LineWidth)
	}
	// Vertex dots.
	for a := 0; a < axes; a++ {
		canvas.DrawCircle(pts[a], m.RadarDotRadius, col)
	}
}

// fillRadarTriangle fills the triangle (center, a, b) by sweeping thin segments
// from the center across the a->b edge. Each step draws a 1px-wide line from the
// center to a point interpolated along the edge; consecutive steps abut to cover
// the triangle with a single flat alpha (no overlap double-darkening), the same
// technique drawArea uses for cartesian fills.
func fillRadarTriangle(canvas widget.Canvas, center, a, b geometry.Point, fill widget.Color) {
	edge := b.Sub(a).Length()
	if edge <= 0 {
		return
	}
	steps := int(math.Ceil(float64(edge)))
	if steps < 1 {
		steps = 1
	}
	for k := 0; k <= steps; k++ {
		t := float32(k) / float32(steps)
		p := a.Lerp(b, t)
		canvas.DrawLine(center, p, fill, 1.5)
	}
}

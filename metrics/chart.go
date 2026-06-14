package metrics

// Chart holds the pixel constants for graft's native charts, which stand in
// for shadcn's Recharts-based <ChartContainer> (ui.shadcn.com/docs/components/chart).
//
// graft cannot embed Recharts, so the look is reproduced from the shadcn
// ChartContainer/ChartConfig conventions. The styling literals translated here
// come from the new-york-v4 chart.tsx wrapper classes:
//
//	ChartContainer: "flex aspect-video justify-center text-xs
//	    [&_.recharts-cartesian-axis-tick_text]:fill-muted-foreground
//	    [&_.recharts-cartesian-grid_line[stroke='#ccc']]:stroke-border/50 ..."
//	CartesianGrid:  vertical={false}                  -> horizontal gridlines only
//	XAxis:          tickLine={false} axisLine={false} tickMargin={10}
//	Bar:            radius={4}
//	Legend square:  "h-2 w-2 shrink-0 rounded-[2px]"  -> 8px swatch, 2px radius
//	Tooltip/legend label text: text-xs (12px), muted-foreground for ticks
//
// Series colors are NEVER stored here; they come from the active theme's
// Chart[0..4] tokens at draw time so light/dark mode switches repaint.
var Chart = struct {
	// AspectRatio is the default width:height ratio (aspect-video = 16/9).
	AspectRatio float32

	// DefaultWidth is the natural width used when the chart is laid out with
	// loose/unbounded constraints (a sensible shadcn card-width default).
	DefaultWidth float32

	// PadTop, PadRight, PadBottom, PadLeft inset the plot area from the
	// chart bounds. The bottom inset holds the X-axis category labels and the
	// left inset holds the Y-axis value labels. Recharts' XAxis tickMargin is
	// 10px; the bottom band accounts for that plus the 12px label line.
	PadTop, PadRight, PadBottom, PadLeft float32

	// GridDivisions is the number of horizontal gridlines / Y-axis ticks
	// drawn across the plot area (lines = GridDivisions+1).
	GridDivisions int

	// GridLineWidth is the stroke width of the muted gridlines (1px hairline).
	GridLineWidth float32

	// GridAlpha is the alpha applied to the Border token for gridlines
	// (stroke-border/50 = 50%).
	GridAlpha float32

	// AxisLabelFontSize is the tick-label font size (text-xs = 12px).
	AxisLabelFontSize float32

	// AxisLabelWeight is the tick-label font weight (normal).
	AxisLabelWeight int

	// AxisLabelLineHeight is the tick-label line box in px.
	AxisLabelLineHeight float32

	// LineWidth is the polyline stroke width for line/area series (2px,
	// Recharts' default strokeWidth).
	LineWidth float32

	// DotRadius is the radius of the optional point markers on line series.
	DotRadius float32

	// DotStrokeWidth is the white halo stroke around dots (Recharts draws a
	// 1px #fff ring, mapped to the chart Background token here).
	DotStrokeWidth float32

	// BarRadius is the rounded-corner radius on bars (Recharts radius={4}).
	BarRadius float32

	// BarGroupGap is the gap between category groups as a fraction of the
	// category slot width (0..1).
	BarGroupGap float32

	// BarInnerGap is the gap between bars within one category group, in px.
	BarInnerGap float32

	// AreaFillAlpha is the alpha applied to a series color for the area fill
	// beneath the line (shadcn area charts use a translucent fill).
	AreaFillAlpha float32

	// LegendHeight is the height of the legend row drawn under the chart.
	LegendHeight float32

	// LegendSwatch is the side length of the legend color square (h-2 w-2).
	LegendSwatch float32

	// LegendSwatchRadius is the legend swatch corner radius (rounded-[2px]).
	LegendSwatchRadius float32

	// LegendGap is the horizontal gap between legend items (gap-4 = 16px).
	LegendGap float32

	// LegendSwatchTextGap is the gap between a legend swatch and its label
	// (gap-1.5 = 6px).
	LegendSwatchTextGap float32

	// LegendTopGap is the vertical gap between the plot/X-axis band and the
	// legend row (pt-3 = 12px).
	LegendTopGap float32

	// --- Polar charts (pie / donut / radial / radar) ---
	//
	// Pie, radial, and radar are drawn in a square polar plot inset by
	// PolarPad on every side. There is no filled-wedge primitive on the base
	// canvas, so wedges and rings are painted as thick StrokeArc bands (a
	// stroke of width W centered on radius r fills the annulus [r-W/2, r+W/2];
	// a full pie is W=outer centered on outer/2). Angles follow gg's screen
	// convention used by the spinner: 0 at 3 o'clock, increasing clockwise,
	// with the first slice/value starting at the top (-pi/2).

	// PolarPad insets the square polar plot from the chart bounds on each side.
	PolarPad float32

	// PieDonutInner is the default donut inner-radius fraction (0 = full pie).
	// shadcn's donut uses innerRadius ~60% of the outer radius.
	PieDonutInner float32

	// PieGapAngle is the angular gap between adjacent slices, in radians, so
	// slices read as distinct wedges (shadcn pies have a thin separator).
	PieGapAngle float64

	// RadialTrackAlpha is the alpha applied to the Muted token for a radial
	// bar's background track ring.
	RadialTrackAlpha float32

	// RadialBarThickness is the thickness of each radial bar ring, in px.
	RadialBarThickness float32

	// RadialBarGap is the gap between concentric radial rings, in px.
	RadialBarGap float32

	// RadialBarRadiusFrac is the fraction of the polar radius at which the
	// outermost radial ring sits (the rings step inward from there).
	RadialBarRadiusFrac float32

	// RadarRings is the number of concentric grid rings drawn behind a radar
	// (matching GridDivisions for visual consistency).
	RadarRings int

	// RadarFillAlpha is the alpha applied to a radar series color for its
	// translucent polygon fill.
	RadarFillAlpha float32

	// RadarSpokeWidth is the stroke width of the radial axis spokes/grid rings.
	RadarSpokeWidth float32

	// RadarDotRadius is the radius of the vertex markers on a radar polygon.
	RadarDotRadius float32

	// PolarLabelGap is the gap between the outer radius and a polar axis label.
	PolarLabelGap float32
}{
	AspectRatio:  16.0 / 9.0,
	DefaultWidth: 460,

	PadTop:    8,
	PadRight:  12,
	PadBottom: 28,
	PadLeft:   36,

	GridDivisions: 4,
	GridLineWidth: 1,
	GridAlpha:     0.5,

	AxisLabelFontSize:   12,
	AxisLabelWeight:     400,
	AxisLabelLineHeight: 16,

	LineWidth:      2,
	DotRadius:      3,
	DotStrokeWidth: 2,

	BarRadius:   4,
	BarGroupGap: 0.3,
	BarInnerGap: 4,

	AreaFillAlpha: 0.25,

	LegendHeight:        20,
	LegendSwatch:        8,
	LegendSwatchRadius:  2,
	LegendGap:           16,
	LegendSwatchTextGap: 6,
	LegendTopGap:        12,

	PolarPad:      16,
	PieDonutInner: 0.6,
	PieGapAngle:   0.012,

	RadialTrackAlpha:    1,
	RadialBarThickness:  16,
	RadialBarGap:        4,
	RadialBarRadiusFrac: 0.95,

	RadarRings:      4,
	RadarFillAlpha:  0.25,
	RadarSpokeWidth: 1,
	RadarDotRadius:  3,

	PolarLabelGap: 10,
}

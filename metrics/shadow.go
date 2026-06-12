package metrics

// ShadowLayer is one stacked low-alpha round-rect approximating a CSS
// box-shadow level (DESIGN.md section 5.5). gogpu/ui has no blur primitive,
// so each shadcn shadow is rendered as a few of these layers drawn before
// the element fill.
//
// DY is the vertical offset in px, Grow is the outward expansion in px
// (the Rect.Expand amount), and Alpha is the layer opacity. The layer color
// is always black.
type ShadowLayer struct {
	// DY is the vertical translation of the layer in px.
	DY float32

	// Grow is the outward expansion of the layer in px on every edge.
	Grow float32

	// Alpha is the opacity of the black layer fill, in [0, 1].
	Alpha float32
}

// Shadow layer tables for the shadcn shadow levels used by graft
// (DESIGN.md section 5.5). Starting values are tuned once against the
// kitchen-sink example during Phase 0 and then frozen by golden tests.
// Alphas stay at or below 0.08 to avoid hard edges.
//
// Usage map (Report 3 section 4): XS for outline button, input, textarea,
// checkbox, radio, switch and select trigger; SM for card, slider thumb and
// active tab; MD for popover and select/dropdown content; LG for dialog,
// sheet and sub-menus.
var (
	// ShadowXS approximates "shadow-xs" (0 1px 2px 0 rgb(0 0 0 / 0.05)).
	ShadowXS = []ShadowLayer{
		{DY: 1, Grow: 0, Alpha: 0.04},
		{DY: 1, Grow: 1, Alpha: 0.03},
	}

	// ShadowSM approximates "shadow-sm".
	ShadowSM = []ShadowLayer{
		{DY: 1, Grow: 0, Alpha: 0.06},
		{DY: 1, Grow: 1, Alpha: 0.05},
		{DY: 2, Grow: 2, Alpha: 0.03},
	}

	// ShadowMD approximates "shadow-md".
	ShadowMD = []ShadowLayer{
		{DY: 2, Grow: 0, Alpha: 0.06},
		{DY: 4, Grow: 2, Alpha: 0.05},
		{DY: 6, Grow: 4, Alpha: 0.03},
	}

	// ShadowLG approximates "shadow-lg".
	ShadowLG = []ShadowLayer{
		{DY: 4, Grow: 2, Alpha: 0.05},
		{DY: 8, Grow: 5, Alpha: 0.04},
		{DY: 12, Grow: 9, Alpha: 0.03},
	}
)

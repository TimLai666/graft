package metrics

// Tooltip holds the exact pixel constants for the shadcn Tooltip content and
// arrow (docs/research/03-shadcn-pixel-spec.md §"Tooltip").
//
// Content (verbatim):
//
//	"z-50 w-fit ... rounded-md bg-foreground px-3 py-1.5 text-xs text-balance
//	 text-background ..."
//
// Arrow (verbatim):
//
//	"z-50 size-2.5 translate-y-[calc(-50%_-_2px)] rotate-45 rounded-[2px]
//	 bg-foreground fill-foreground"
//
// The content auto-inverts: fill is the theme Foreground token, text is the
// Background token. Radius routes through the theme RadiusMD(); only literals
// (the arrow's rounded-[2px] and its 2px straddle offset) live here.
const (
	// TooltipPadX is the horizontal padding in px (px-3 = 0.75rem).
	TooltipPadX float32 = 12

	// TooltipPadY is the vertical padding in px (py-1.5 = 0.375rem).
	TooltipPadY float32 = 6

	// TooltipFontSize is the content text size in px (text-xs).
	TooltipFontSize float32 = 12

	// TooltipLineHeight is the content line box in px (text-xs line-height).
	TooltipLineHeight float32 = 16

	// TooltipFontWeight is the content text weight (default 400).
	TooltipFontWeight int = 400

	// TooltipGap is the spacing between trigger and content in px. shadcn
	// places the arrow flush against the trigger (sideOffset 0); the visible
	// gap is the arrow's half-diagonal. Position uses 0 so the arrow tip
	// touches the trigger edge.
	TooltipGap float32 = 0

	// TooltipArrowSize is the arrow square side length in px (size-2.5 =
	// 0.625rem); rotated 45° it reads as a diamond.
	TooltipArrowSize float32 = 10

	// TooltipArrowRadius is the arrow corner radius in px (rounded-[2px]).
	TooltipArrowRadius float32 = 2

	// TooltipArrowInset is how far the arrow's center sits past the content
	// edge in px. shadcn's translate-y-[calc(-50%-2px)] pulls the rotated
	// 10px square so its center is 2px outside the content box, leaving the
	// near half overlapping the content (seamless) and the far half forming
	// the visible point.
	TooltipArrowInset float32 = 2

	// TooltipDelayMillis is the anti-flicker hover delay in ms. shadcn ships
	// delayDuration 0; graft adds a small grace period so brushing past a
	// target does not flash a tooltip. See tooltip.go.
	TooltipDelayMillis = 80
)

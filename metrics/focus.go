package metrics

// Focus-ring constants shared by every focusable component
// (DESIGN.md section 5.1 "Shared" and section 5.2).
//
// The shadcn recipe is "focus-visible:border-ring focus-visible:ring-[3px]
// focus-visible:ring-ring/50": a 3px no-offset box-shadow band outside the
// border box plus the element border turning solid --ring.
const (
	// RingWidth is the width in px of the standard focus ring band drawn
	// outside the element bounds (ring-[3px]).
	RingWidth float32 = 3

	// RingAlpha is the alpha applied to the theme Ring color for the
	// standard focus ring (ring-ring/50).
	RingAlpha float32 = 0.5

	// InvalidRingAlphaLight is the alpha applied to the theme Destructive
	// color for the aria-invalid focus ring in light mode
	// (aria-invalid:ring-destructive/20).
	InvalidRingAlphaLight float32 = 0.2

	// InvalidRingAlphaDark is the alpha applied to the theme Destructive
	// color for the aria-invalid focus ring in dark mode
	// (dark:aria-invalid:ring-destructive/40).
	InvalidRingAlphaDark float32 = 0.4

	// SliderRingWidth is the focus/hover ring width in px for the slider
	// thumb (ring-4).
	SliderRingWidth float32 = 4

	// LegacyCloseRingWidth is the ring width in px for the Dialog/Sheet
	// close button, which keeps the legacy "ring-2 ring-offset-2" recipe.
	LegacyCloseRingWidth float32 = 2

	// LegacyCloseRingOffset is the ring offset in px for the Dialog/Sheet
	// close button (ring-offset-2); the offset gap is filled with the
	// theme Background color.
	LegacyCloseRingOffset float32 = 2

	// DisabledOpacity is the opacity multiplier applied to every color of
	// a disabled control (disabled:opacity-50, DESIGN.md section 5.3).
	DisabledOpacity float32 = 0.5
)

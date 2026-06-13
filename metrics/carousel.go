package metrics

import "time"

// Carousel holds the exact pixel constants for the shadcn Carousel component.
//
// shadcn/ui carousel anatomy:
//
//	root:    "relative" (contains the viewport and nav buttons)
//	content: "flex" (horizontal or vertical slide container)
//	item:    "min-w-0 shrink-0 grow-0 basis-full" (per-slide wrapper)
//	prev/next buttons: "absolute h-8 w-8 rounded-full" positioned centered
//	         on the left/right edges (top/bottom for vertical)
//
// The nav buttons use the outline button variant (border, bg-background,
// shadow-xs) at 32x32 with a 16px icon.
var Carousel = struct {
	// ButtonSize is the nav button diameter in px (h-8 w-8 = 32).
	ButtonSize float32

	// ButtonOffset is the negative offset from the carousel edge to the
	// button center in px. The button is centered on the edge, so the
	// offset is -ButtonSize/2 = -16.
	ButtonOffset float32

	// ButtonIconSize is the icon size inside the nav button in px
	// (size-4 = 16).
	ButtonIconSize float32

	// Gap is the spacing between carousel items in px (gap-4 = 16,
	// applied via padding-left on each item in shadcn's Embla-based
	// implementation).
	Gap float32

	// AnimDuration is the slide transition animation duration.
	AnimDuration time.Duration

	// ButtonBorderWidth is the outline button border in px (border = 1).
	ButtonBorderWidth float32

	// DarkButtonBgAlpha is the dark-mode outline button rest fill alpha
	// on the Input token (dark:bg-input/30).
	DarkButtonBgAlpha float32
}{
	ButtonSize:        32,
	ButtonOffset:      -16,
	ButtonIconSize:    16,
	Gap:               16,
	AnimDuration:      300 * time.Millisecond,
	ButtonBorderWidth: 1,
	DarkButtonBgAlpha: 0.3,
}

package metrics

// InputGroup holds the pixel constants for the shadcn InputGroup composite
// (shadcn/ui input-group), which shares the Input chrome but hosts a borderless
// inner field plus leading/trailing addons:
//
//	root  "flex h-9 w-full rounded-md border border-input shadow-xs
//	       dark:bg-input/30" (same focus/invalid recipe as Input)
//	addon "flex items-center gap-2 text-muted-foreground [&_svg]:size-4
//	       [&_svg]:text-muted-foreground" with px-3 edge padding
//	input "borderless bg-transparent" filling the remaining space.
//
// Radius routes through the theme (rounded-md → t.RadiusMD()); only fixed
// literals live here. The package is pure data with no imports.
var InputGroup = struct {
	// Height is the container height in px (h-9 → 36).
	Height float32

	// PadX is the horizontal edge padding before/after addons and text in px
	// (px-3 → 12).
	PadX float32

	// Gap is the spacing in px between an addon and the text (gap-2 → 8).
	Gap float32

	// IconSize is the addon icon box in px ([&_svg]:size-4 → 16).
	IconSize float32

	// FontSize is the inner text size in px (text-sm → 14).
	FontSize float32

	// FontWeight is the inner text weight (normal → 400).
	FontWeight int

	// BorderWidth is the container border width in px (border → 1).
	BorderWidth float32

	// CaretWidth is the inner caret width in px (1px Foreground).
	CaretWidth float32

	// DarkFillAlpha is the dark-mode background fill alpha on the Input token
	// (dark:bg-input/30 → 0.30).
	DarkFillAlpha float32
}{
	Height:        36,
	PadX:          12,
	Gap:           8,
	IconSize:      16,
	FontSize:      14,
	FontWeight:    400,
	BorderWidth:   1,
	CaretWidth:    1,
	DarkFillAlpha: 0.30,
}

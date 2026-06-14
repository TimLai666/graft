package metrics

// Input holds the exact pixel constants for the shadcn Input control
// (current ui.shadcn.com live style), transcribed from the registry class
// string:
//
//	"h-8 w-full min-w-0 rounded-lg border border-input bg-transparent px-2.5 py-1
//	 text-base shadow-xs ... placeholder:text-muted-foreground ... md:text-sm
//	 dark:bg-input/30" + focus/invalid recipes.
//
// Radius routes through the theme (rounded-lg → t.RadiusLG()); only fixed
// literals live here. The package is pure data with no imports.
var Input = struct {
	// Height is the control height in px (h-8 → 32).
	Height float32

	// PadX is the horizontal content padding in px (px-2.5 → 10).
	PadX float32

	// PadY is the vertical content padding in px (py-1 → 4).
	PadY float32

	// FontSize is the desktop text size in px (md:text-sm → 14).
	FontSize float32

	// FontWeight is the text weight (text-base/text-sm default → 400).
	FontWeight int

	// BorderWidth is the control border width in px (border → 1).
	BorderWidth float32

	// CaretWidth is the text caret width in px (1px Foreground).
	CaretWidth float32

	// DarkFillAlpha is the dark-mode background fill alpha applied to the
	// Input token (dark:bg-input/30 → 0.30).
	DarkFillAlpha float32
}{
	Height:        32,
	PadX:          10,
	PadY:          4,
	FontSize:      14,
	FontWeight:    400,
	BorderWidth:   1,
	CaretWidth:    1,
	DarkFillAlpha: 0.30,
}

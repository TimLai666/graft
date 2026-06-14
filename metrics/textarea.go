package metrics

// Textarea holds the exact pixel constants for the shadcn Textarea control
// (current ui.shadcn.com live style), transcribed from the registry class
// string:
//
//	"flex field-sizing-content min-h-16 w-full rounded-lg border border-input
//	 bg-transparent px-2.5 py-2 text-base shadow-xs ... md:text-sm dark:bg-input/30"
//	+ "focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
//	+ "aria-invalid:border-destructive aria-invalid:ring-destructive/20
//	   dark:aria-invalid:ring-destructive/40".
//
// Radius routes through the theme (rounded-lg → t.RadiusLG()); only fixed
// literals live here. The package is pure data with no imports.
var Textarea = struct {
	// MinHeight is the minimum control height in px (min-h-16 → 64).
	MinHeight float32

	// PadX is the horizontal content padding in px (px-2.5 → 10).
	PadX float32

	// PadY is the vertical content padding in px (py-2 → 8).
	PadY float32

	// FontSize is the desktop text size in px (md:text-sm → 14).
	FontSize float32

	// FontWeight is the text weight (text-base/text-sm default → 400).
	FontWeight int

	// LineHeight is the per-line box height in px. shadcn text areas inherit
	// the document line-height; at 14px text the rendered line box is 20px
	// (leading-5 / 1.43 ratio used by the Geist body scale).
	LineHeight float32

	// BorderWidth is the control border width in px (border → 1).
	BorderWidth float32

	// CaretWidth is the text caret width in px (1px Foreground).
	CaretWidth float32

	// DarkFillAlpha is the dark-mode background fill alpha applied to the
	// Input token (dark:bg-input/30 → 0.30).
	DarkFillAlpha float32
}{
	MinHeight:     64,
	PadX:          10,
	PadY:          8,
	FontSize:      14,
	FontWeight:    400,
	LineHeight:    20,
	BorderWidth:   1,
	CaretWidth:    1,
	DarkFillAlpha: 0.30,
}

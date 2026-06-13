package metrics

// Checkbox holds the exact pixel constants for the shadcn Checkbox control
// (docs/research/03-shadcn-pixel-spec.md §5 "Checkbox"), transcribed from the
// new-york-v4 registry:
//
//	root: "peer size-4 shrink-0 rounded-[4px] border border-input shadow-xs ...
//	       data-[state=checked]:border-primary data-[state=checked]:bg-primary
//	       data-[state=checked]:text-primary-foreground dark:bg-input/30
//	       dark:data-[state=checked]:bg-primary"
//	indicator: CheckIcon "size-3.5" (14px) / Minus for indeterminate.
//
// The box radius is a literal rounded-[4px] (NOT routed through theme).
var Checkbox = struct {
	// Size is the box width/height in px (size-4 → 16).
	Size float32

	// Radius is the box corner radius in px (rounded-[4px] → 4, literal).
	Radius float32

	// BorderWidth is the box border width in px (border → 1).
	BorderWidth float32

	// IconSize is the check / minus glyph size in px (size-3.5 → 14).
	IconSize float32

	// LabelGap is the gap between box and label in px (the row's gap-2 → 8).
	LabelGap float32

	// LabelFontSize is the label text size in px (Label text-sm → 14).
	LabelFontSize float32

	// LabelFontWeight is the label text weight (Label font-medium → 500).
	LabelFontWeight int

	// LabelLineHeight is the label line box in px (leading-none → font size).
	LabelLineHeight float32

	// DarkFillAlpha is the dark unchecked background fill alpha applied to
	// the Input token (dark:bg-input/30 → 0.30).
	DarkFillAlpha float32
}{
	Size:            16,
	Radius:          4,
	BorderWidth:     1,
	IconSize:        14,
	LabelGap:        8,
	LabelFontSize:   14,
	LabelFontWeight: 500,
	LabelLineHeight: 14,
	DarkFillAlpha:   0.30,
}

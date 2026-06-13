package metrics

// RadioGroup holds the exact pixel constants for the shadcn RadioGroup
// control (docs/research/03-shadcn-pixel-spec.md §5 "Radio Group"),
// transcribed from the new-york-v4 registry:
//
//	group: "grid gap-3" (12px gap)
//	item:  "aspect-square size-4 shrink-0 rounded-full border border-input
//	        text-primary shadow-xs ... dark:bg-input/30"
//	indicator: centered CircleIcon "size-2 ... fill-primary" (8px dot)
var RadioGroup = struct {
	// Size is the circle diameter in px (size-4 → 16).
	Size float32

	// BorderWidth is the circle border width in px (border → 1).
	BorderWidth float32

	// DotSize is the inner selected dot diameter in px (size-2 → 8).
	DotSize float32

	// GroupGap is the spacing between items in px (gap-3 → 12).
	GroupGap float32

	// LabelGap is the gap between circle and label in px (8).
	LabelGap float32

	// LabelFontSize is the label text size in px (Label text-sm → 14).
	LabelFontSize float32

	// LabelFontWeight is the label text weight (Label font-medium → 500).
	LabelFontWeight int

	// LabelLineHeight is the label line box in px (leading-none → font size).
	LabelLineHeight float32

	// DarkFillAlpha is the dark unselected background fill alpha applied to
	// the Input token (dark:bg-input/30 → 0.30).
	DarkFillAlpha float32
}{
	Size:            16,
	BorderWidth:     1,
	DotSize:         8,
	GroupGap:        12,
	LabelGap:        8,
	LabelFontSize:   14,
	LabelFontWeight: 500,
	LabelLineHeight: 14,
	DarkFillAlpha:   0.30,
}

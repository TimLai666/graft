package metrics

// InputOTP holds the exact pixel constants for the shadcn InputOTP control
// (docs/research/03-shadcn-pixel-spec.md), transcribed from the new-york-v4
// registry:
//
//	slot: "relative flex h-10 w-10 items-center justify-center border-y
//	       border-r border-input text-lg font-semibold ...
//	       first:rounded-l-md first:border-l last:rounded-r-md"
//	separator: "flex items-center" (a 16×2 dash)
//
// Groups of slots are separated by a horizontal dash (gap-x-4 between groups).
// Individual slots within a group are butted together (gap 0, shared borders).
// The radius routes through the theme (rounded-md -> t.RadiusMD()).
var InputOTP = struct {
	// SlotSize is the slot width and height in px (w-10 h-10 -> 40).
	SlotSize float32

	// SlotGap is the horizontal gap between slots within the same group in
	// px. Shadcn butts slots together with shared borders (0), but a small
	// visual gap is used here for clarity between distinct boxes.
	SlotGap float32

	// GroupGap is the horizontal gap between slot groups in px (gap-x-4 ->
	// 16, including room for the separator dash).
	GroupGap float32

	// FontSize is the slot digit text size in px (text-lg -> 20 on mobile,
	// but we use the desktop size).
	FontSize float32

	// FontWeight is the slot digit text weight (font-semibold -> 600).
	FontWeight int

	// BorderWidth is the slot border width in px (border -> 1).
	BorderWidth float32

	// Radius is the corner radius in px for the first/last slot in each
	// group (rounded-md -> 8). Interior slots have square corners.
	Radius float32

	// SepWidth is the separator dash width in px (the Dot/Minus icon area).
	SepWidth float32

	// SepHeight is the separator dash height in px.
	SepHeight float32

	// CaretWidth is the blinking caret width in px.
	CaretWidth float32

	// DarkFillAlpha is the dark-mode background fill alpha applied to the
	// Input token (dark:bg-input/30 -> 0.30).
	DarkFillAlpha float32
}{
	SlotSize:      40,
	SlotGap:       0,
	GroupGap:      16,
	FontSize:      20,
	FontWeight:    600,
	BorderWidth:   1,
	Radius:        8,
	SepWidth:      16,
	SepHeight:     2,
	CaretWidth:    1,
	DarkFillAlpha: 0.30,
}

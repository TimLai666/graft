package metrics

// Kbd metrics transcribed from the shadcn new-york-v4 Kbd component
// (docs/research/03-shadcn-pixel-spec.md "Kbd", quoted verbatim there).
//
// Kbd: "pointer-events-none inline-flex h-5 w-fit min-w-5 items-center
//
//	justify-center gap-1 rounded-sm bg-muted px-1 font-sans text-xs
//	font-medium text-muted-foreground select-none".
//
// Note: font-sans, NOT mono. Inside tooltips the colors invert
// (bg-background/20 text-background); graft renders the default surface.
//
// Radius routes through the theme (RadiusSM = 6px at the default --radius).
var Kbd = struct {
	// Height is the chip height in px (h-5 = 20).
	Height float32

	// MinWidth is the chip minimum width in px (min-w-5 = 20), making a
	// single-glyph key a square.
	MinWidth float32

	// PadX is the chip horizontal padding in px (px-1 = 4).
	PadX float32

	// FontSize is the key label size in px (text-xs = 12).
	FontSize float32

	// LineHeight is the key label line box in px (text-xs = 16).
	LineHeight float32

	// FontWeight is the key label weight (font-medium = 500).
	FontWeight int

	// Gap is the spacing between adjacent key chips in a combo in px
	// (gap-1 = 4); also the in-chip gap for an icon+label key.
	Gap float32

	// IconSize is the in-key icon box in px (size-3 = 12 per spec note).
	IconSize float32
}{
	Height:     20,  // h-5
	MinWidth:   20,  // min-w-5
	PadX:       4,   // px-1
	FontSize:   12,  // text-xs
	LineHeight: 16,  // text-xs line height
	FontWeight: 500, // font-medium
	Gap:        4,   // gap-1
	IconSize:   12,  // size-3 (icons inside Kbd)
}

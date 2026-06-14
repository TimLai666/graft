package metrics

// Toggle metrics transcribed from the current shadcn (radix-era) toggleVariants
// cva (compare/shadcn-spec.md "toggle", measured 2026-06-14 from ui.shadcn.com).
//
// Base: "inline-flex items-center justify-center gap-2 rounded-md text-sm
//        font-medium whitespace-nowrap transition-[color,box-shadow]
//        outline-none hover:bg-muted hover:text-muted-foreground ...
//        data-[state=on]:bg-accent data-[state=on]:text-accent-foreground
//        ... [&_svg:not([class*='size-'])]:size-4".
// Variants: default "bg-transparent";
//           outline "border border-input bg-transparent shadow-xs
//                    hover:bg-accent hover:text-accent-foreground".
// Sizes: default "h-7 min-w-7 px-2.5", sm "h-8 min-w-8 px-1.5",
//        lg "h-10 min-w-10 px-2.5".
// On-state fill is --accent.
//
// Radius routes through the theme (RadiusMD = 8px at the default --radius).

// ToggleSize holds the px metrics of one toggleVariants size variant.
type ToggleSize struct {
	// Height is the fixed control height (h-7 = 28, h-8 = 32, h-10 = 40).
	Height float32

	// MinWidth is the minimum width (min-w-7 = 28, min-w-8 = 32,
	// min-w-10 = 40), making an icon-only toggle a square.
	MinWidth float32

	// PadX is the horizontal padding (px-2.5 = 10, px-1.5 = 6).
	PadX float32
}

// Toggle collects every px constant of the shadcn Toggle component.
var Toggle = struct {
	// Default is size "default": h-7 min-w-7 px-2.5.
	Default ToggleSize

	// SM is size "sm": h-8 min-w-8 px-1.5.
	SM ToggleSize

	// LG is size "lg": h-10 min-w-10 px-2.5.
	LG ToggleSize

	// FontSize is the label size in px (text-sm = 14).
	FontSize float32

	// LineHeight is the label line box in px (text-sm = 20).
	LineHeight float32

	// FontWeight is the label weight (font-medium = 500).
	FontWeight int

	// Gap is the spacing between icon and label in px (gap-2 = 8).
	Gap float32

	// IconSize is the svg icon box in px (size-4 = 16).
	IconSize float32

	// BorderWidth is the outline-variant border width in px (border = 1px),
	// in the --input token.
	BorderWidth float32
}{
	Default: ToggleSize{Height: 28, MinWidth: 28, PadX: 10}, // h-7 min-w-7 px-2.5
	SM:      ToggleSize{Height: 32, MinWidth: 32, PadX: 6},  // h-8 min-w-8 px-1.5
	LG:      ToggleSize{Height: 40, MinWidth: 40, PadX: 10}, // h-10 min-w-10 px-2.5

	FontSize:    14,  // text-sm
	LineHeight:  20,  // text-sm line height
	FontWeight:  500, // font-medium
	Gap:         8,   // gap-2
	IconSize:    16,  // size-4
	BorderWidth: 1,   // border (outline)
}

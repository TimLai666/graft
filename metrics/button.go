package metrics

// Button metrics transcribed from the shadcn new-york-v4 buttonVariants cva
// (docs/research/03-shadcn-pixel-spec.md §5 "Button", quoted verbatim there).
//
// Base classes: "inline-flex shrink-0 items-center justify-center gap-1.5
// rounded-lg text-sm font-medium ... [&_svg:not([class*='size-'])]:size-4".
// The corner radius is rounded-lg and routes through the theme
// (Theme.RadiusLG, 10px at the default --radius of 10).

// ButtonSize holds the pixel metrics of one buttonVariants size variant.
type ButtonSize struct {
	// Height is the fixed control height (h-9 = 36, h-6 = 24, h-8 = 32,
	// h-10 = 40; icon sizes are squares of the same value).
	Height float32

	// PadX is the horizontal padding without an icon (px-4 = 16, px-2 = 8,
	// px-3 = 12, px-6 = 24).
	PadX float32

	// PadXWithIcon is the horizontal padding when the button directly
	// contains an svg (has-[>svg]:px-3 = 12, px-1.5 = 6, px-2.5 = 10,
	// px-4 = 16).
	PadXWithIcon float32

	// PadY is the vertical padding (py-2 = 8). The fixed Height already
	// pins the control; PadY is recorded for completeness.
	PadY float32

	// Gap is the spacing between icon and label (gap-2 = 8, gap-1 = 4,
	// gap-1.5 = 6).
	Gap float32

	// FontSize is the label size (text-sm = 14, xs: text-xs = 12).
	FontSize float32

	// IconSize is the svg icon box ([&_svg:...]:size-4 = 16, xs sizes:
	// size-3 = 12).
	IconSize float32
}

// Button collects every pixel constant of the shadcn Button component.
var Button = struct {
	// Default is size "default": h-8 px-2.5 py-2 gap-1.5 has-[>svg]:px-2
	// (current ui.shadcn.com live style; rounded-lg via the theme).
	Default ButtonSize

	// XS is size "xs": h-6 gap-1 rounded-md px-2 text-xs
	// has-[>svg]:px-1.5 [&_svg:not([class*='size-'])]:size-3.
	XS ButtonSize

	// SM is size "sm": h-8 gap-1.5 rounded-md px-3 has-[>svg]:px-2.5.
	SM ButtonSize

	// LG is size "lg": h-10 rounded-md px-6 has-[>svg]:px-4.
	LG ButtonSize

	// Icon is size "icon": size-9 (36×36 square).
	Icon ButtonSize

	// IconXS is size "icon-xs": size-6 rounded-md
	// [&_svg:not([class*='size-'])]:size-3 (24×24, 12px icon).
	IconXS ButtonSize

	// IconSM is size "icon-sm": size-8 (32×32).
	IconSM ButtonSize

	// IconLG is size "icon-lg": size-10 (40×40).
	IconLG ButtonSize

	// FontWeight is the label weight (font-medium = 500).
	FontWeight int

	// BorderWidth is the outline-variant border width in px (border = 1px).
	BorderWidth float32

	// UnderlineOffset is the link-variant hover underline distance below
	// the text baseline (underline-offset-4 = 4px).
	UnderlineOffset float32

	// UnderlineWidth is the link-variant hover underline thickness
	// (CSS text-decoration default ≈ 1px).
	UnderlineWidth float32

	// HoverPrimaryAlpha is the default-variant hover fill alpha
	// (hover:bg-primary/90).
	HoverPrimaryAlpha float32

	// HoverSecondaryAlpha is the secondary-variant hover fill alpha
	// (hover:bg-secondary/80).
	HoverSecondaryAlpha float32

	// HoverDestructiveAlpha is the destructive-variant hover fill alpha
	// (hover:bg-destructive/90).
	HoverDestructiveAlpha float32

	// DarkDestructiveBgAlpha is the destructive-variant dark-mode rest
	// fill alpha (dark:bg-destructive/60).
	DarkDestructiveBgAlpha float32

	// DarkGhostHoverAlpha is the ghost-variant dark-mode hover fill alpha
	// (dark:hover:bg-accent/50).
	DarkGhostHoverAlpha float32

	// DarkOutlineBgAlpha is the outline-variant dark-mode rest fill alpha
	// multiplier on the Input token (dark:bg-input/30).
	DarkOutlineBgAlpha float32

	// DarkOutlineHoverAlpha is the outline-variant dark-mode hover fill
	// alpha multiplier on the Input token (dark:hover:bg-input/50).
	DarkOutlineHoverAlpha float32
}{
	Default: ButtonSize{Height: 32, PadX: 10, PadXWithIcon: 8, PadY: 8, Gap: 6, FontSize: 14, IconSize: 16},
	XS:      ButtonSize{Height: 24, PadX: 8, PadXWithIcon: 6, Gap: 4, FontSize: 12, IconSize: 12},
	SM:      ButtonSize{Height: 32, PadX: 12, PadXWithIcon: 10, Gap: 6, FontSize: 14, IconSize: 16},
	LG:      ButtonSize{Height: 40, PadX: 24, PadXWithIcon: 16, Gap: 8, FontSize: 14, IconSize: 16},
	Icon:    ButtonSize{Height: 36, Gap: 8, FontSize: 14, IconSize: 16},
	IconXS:  ButtonSize{Height: 24, Gap: 4, FontSize: 12, IconSize: 12},
	IconSM:  ButtonSize{Height: 32, Gap: 6, FontSize: 14, IconSize: 16},
	IconLG:  ButtonSize{Height: 40, Gap: 8, FontSize: 14, IconSize: 16},

	FontWeight:      500, // font-medium
	BorderWidth:     1,   // border (outline variant)
	UnderlineOffset: 4,   // underline-offset-4
	UnderlineWidth:  1,

	HoverPrimaryAlpha:      0.9, // hover:bg-primary/90
	HoverSecondaryAlpha:    0.8, // hover:bg-secondary/80
	HoverDestructiveAlpha:  0.9, // hover:bg-destructive/90
	DarkDestructiveBgAlpha: 0.6, // dark:bg-destructive/60
	DarkGhostHoverAlpha:    0.5, // dark:hover:bg-accent/50
	DarkOutlineBgAlpha:     0.3, // dark:bg-input/30
	DarkOutlineHoverAlpha:  0.5, // dark:hover:bg-input/50
}

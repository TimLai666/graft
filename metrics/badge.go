package metrics

// Badge metrics.
//
// Source: current ui.shadcn.com live badge.tsx badgeVariants cva base
// (radix era, newer than new-york-v4):
//
//	"inline-flex w-fit shrink-0 items-center justify-center gap-1
//	 overflow-hidden rounded-4xl border border-transparent px-2 py-0.5
//	 text-xs font-medium whitespace-nowrap transition-[color,box-shadow]
//	 focus-visible:border-ring focus-visible:ring-[3px]
//	 focus-visible:ring-ring/50 aria-invalid:border-destructive
//	 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40
//	 [&>svg]:pointer-events-none [&>svg]:size-3"
//
// Variants:
//
//	default:     "bg-primary text-primary-foreground [a&]:hover:bg-primary/90"
//	secondary:   "bg-secondary text-secondary-foreground [a&]:hover:bg-secondary/90"
//	destructive: "bg-destructive text-white focus-visible:ring-destructive/20
//	              dark:bg-destructive/60 dark:focus-visible:ring-destructive/40
//	              [a&]:hover:bg-destructive/90"
//	outline:     "border-border text-foreground [a&]:hover:bg-accent
//	              [a&]:hover:text-accent-foreground"
//	ghost:       "[a&]:hover:bg-accent [a&]:hover:text-accent-foreground"
//	link:        "text-primary underline-offset-4 [a&]:hover:underline"
//
// The radius is rounded-4xl (near-pill ≈ 26px at the default --radius of
// 10), drawn with the theme Radius4XL.
const (
	// BadgePadX is the horizontal padding in px (px-2).
	BadgePadX float32 = 8

	// BadgePadY is the vertical padding in px (py-0.5).
	BadgePadY float32 = 2

	// BadgeGap is the gap between icon, text, and extra children in px
	// (gap-1).
	BadgeGap float32 = 4

	// BadgeFontSize is the label font size in px (text-xs).
	BadgeFontSize float32 = 12

	// BadgeLineHeight is the label line height in px (text-xs = 12px/16px).
	BadgeLineHeight float32 = 16

	// BadgeFontWeight is the label font weight (font-medium).
	BadgeFontWeight = 500

	// BadgeIconSize is the svg icon size in px ([&>svg]:size-3).
	BadgeIconSize float32 = 12

	// BadgeBorderWidth is the border width in px (border); the border is
	// transparent except for the outline variant (border-border) and the
	// focus-visible state (border-ring), but always contributes to the
	// box size like a CSS border.
	BadgeBorderWidth float32 = 1

	// BadgeHoverAlpha is the hover fill alpha for the default, secondary,
	// and destructive variants when rendered as a link
	// ([a&]:hover:bg-primary/90, .../90).
	BadgeHoverAlpha float32 = 0.9

	// BadgeDarkDestructiveAlpha is the dark-mode base fill alpha for the
	// destructive variant (dark:bg-destructive/60).
	BadgeDarkDestructiveAlpha float32 = 0.6

	// BadgeUnderlineOffset is the link-variant hover underline offset in
	// px below the text baseline (underline-offset-4).
	BadgeUnderlineOffset float32 = 4

	// BadgeUnderlineWidth is the link-variant hover underline thickness
	// in px (CSS text-decoration default ≈ 1px at 12px font).
	BadgeUnderlineWidth float32 = 1
)

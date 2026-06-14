package metrics

// NavigationMenu holds the exact pixel constants for the shadcn Navigation
// Menu — a horizontal bar whose items are either plain links or triggers
// (label + small chevron) that open a floating content panel ("viewport")
// below the bar.
//
// Values are transcribed from the shadcn new-york-v4 navigation-menu registry
// entry. Radii route through the theme (RadiusMD for triggers/links, RadiusLG
// for the content panel, RadiusSM for content link rows); only literals live
// here.
//
// Root list (verbatim): "group flex flex-1 list-none items-center
// justify-center gap-1" → gap-1 = 4px between items; the root has no surface
// chrome (transparent).
//
// Trigger / link (verbatim navigationMenuTriggerStyle): "group inline-flex h-9
// w-max items-center justify-center rounded-md bg-background px-4 py-2 text-sm
// font-medium ... hover:bg-accent hover:text-accent-foreground
// focus:bg-accent focus:text-accent-foreground
// data-[state=open]:bg-accent/50 data-[state=open]:text-accent-foreground" →
// h-9 = 36px, px-4 = 16px, text-sm = 14px, font-medium = 500; hover/open fill
// --accent. The trigger appends a chevron "size-3" (12px) with "ml-1" (4px).
//
// Content panel (verbatim viewport): "... rounded-md border bg-popover
// text-popover-foreground shadow ... p-1 ..." with content "p-2"; the live v4
// style uses rounded-lg + shadow-md (matches the menu/popover family). Content
// link rows (verbatim): "block select-none space-y-1 rounded-sm p-3 ...
// hover:bg-accent hover:text-accent-foreground" → p-3 = 12px, rounded-sm; a
// row has a text-sm font-medium title (leading-none) and an optional
// text-sm text-muted-foreground description (leading-snug, line-clamp-2).
const (
	// NavigationMenuHeight is the bar/trigger height in px (h-9).
	NavigationMenuHeight float32 = 36

	// NavigationMenuGap is the gap between top-level items in px (gap-1).
	NavigationMenuGap float32 = 4

	// NavigationMenuTriggerPadX is the trigger/link horizontal padding in px
	// (px-4).
	NavigationMenuTriggerPadX float32 = 16

	// NavigationMenuTriggerFontSize is the trigger/link label size in px
	// (text-sm).
	NavigationMenuTriggerFontSize float32 = 14

	// NavigationMenuTriggerFontWeight is the trigger/link label weight
	// (font-medium).
	NavigationMenuTriggerFontWeight int = 500

	// NavigationMenuChevronSize is the trigger chevron box in px (size-3).
	NavigationMenuChevronSize float32 = 12

	// NavigationMenuChevronGap is the gap between label and chevron in px
	// (ml-1).
	NavigationMenuChevronGap float32 = 4

	// NavigationMenuBorderWidth is the content panel border width in px
	// (border).
	NavigationMenuBorderWidth float32 = 1

	// NavigationMenuContentPad is the content panel inner padding in px (the
	// viewport p-1 plus content padding settle to a p-2 = 8px inset around
	// the link list).
	NavigationMenuContentPad float32 = 8

	// NavigationMenuContentMinWidth is the content panel minimum width in px.
	// shadcn nav-menu content has no fixed min on the viewport, but the demo
	// link grids settle around 14rem; pin a comfortable floor.
	NavigationMenuContentMinWidth float32 = 224

	// NavigationMenuSideOffset is the gap in px between the trigger and the
	// content panel opened below it (sideOffset = 4, matching the menu
	// family).
	NavigationMenuSideOffset float32 = 4

	// NavigationMenuLinkPad is the content link row padding in px (p-3).
	NavigationMenuLinkPad float32 = 12

	// NavigationMenuLinkTitleSize is the content link title size in px
	// (text-sm).
	NavigationMenuLinkTitleSize float32 = 14

	// NavigationMenuLinkTitleWeight is the content link title weight
	// (font-medium).
	NavigationMenuLinkTitleWeight int = 500

	// NavigationMenuLinkTitleLineHeight is the content link title line box in
	// px (leading-none ≈ text-sm cap height; use 20px text-sm leading for the
	// single-title case).
	NavigationMenuLinkTitleLineHeight float32 = 20

	// NavigationMenuLinkDescSize is the content link description size in px
	// (text-sm).
	NavigationMenuLinkDescSize float32 = 14

	// NavigationMenuLinkDescLineHeight is the content link description line
	// box in px (leading-snug ≈ 1.375 × 14 ≈ 20).
	NavigationMenuLinkDescLineHeight float32 = 20

	// NavigationMenuLinkDescGap is the vertical gap between a link title and
	// its description in px (space-y-1 = 4px).
	NavigationMenuLinkDescGap float32 = 4
)

package metrics

// Menubar holds the exact pixel constants for the shadcn Menubar — a
// horizontal bar of menu triggers, each opening a dropdown panel below it.
//
// Menubar is not enumerated in docs/research/03-shadcn-pixel-spec.md §5; the
// bar/trigger values are transcribed from the shadcn new-york-v4 menubar
// registry entry, and the drop-down panels reuse the shared menu engine
// (metrics.Menu — same class strings as DropdownMenu).
//
// Root (verbatim): "flex h-9 items-center gap-1 rounded-md border bg-background
// p-1 shadow-xs" → h-9 = 36px, gap-1 = 4px, rounded-md, 1px border,
// bg --background, p-1 = 4px, shadow-xs.
//
// Trigger (verbatim): "flex items-center rounded-sm px-2 py-1 text-sm
// font-medium outline-hidden select-none focus:bg-accent
// focus:text-accent-foreground data-[state=open]:bg-accent
// data-[state=open]:text-accent-foreground" → px-2 = 8px, py-1 = 4px,
// rounded-sm, text-sm = 14px, font-medium = 500; hover/open fill --accent with
// text --accent-foreground.
//
// Content (verbatim): "... min-w-[12rem] ... rounded-md border bg-popover p-1
// text-popover-foreground shadow-md ..." → min-w-[12rem] = 192px (wider than
// the dropdown-menu default 128px); other panel/row metrics route through
// metrics.Menu.
//
// Radii route through the theme (RadiusMD for bar/panel, RadiusSM for the
// triggers); only literals live here.
const (
	// MenubarHeight is the bar height in px (h-9).
	MenubarHeight float32 = 36

	// MenubarGap is the gap between triggers in px (gap-1).
	MenubarGap float32 = 4

	// MenubarPadding is the bar inner padding in px (p-1).
	MenubarPadding float32 = 4

	// MenubarBorderWidth is the bar border width in px (border).
	MenubarBorderWidth float32 = 1

	// MenubarTriggerPadX is the trigger horizontal padding in px (px-2).
	MenubarTriggerPadX float32 = 8

	// MenubarTriggerPadY is the trigger vertical padding in px (py-1).
	MenubarTriggerPadY float32 = 4

	// MenubarTriggerFontSize is the trigger label size in px (text-sm).
	MenubarTriggerFontSize float32 = 14

	// MenubarTriggerFontWeight is the trigger label weight (font-medium).
	MenubarTriggerFontWeight int = 500

	// MenubarTriggerLineHeight is the trigger label line box in px (text-sm
	// leading = 20px).
	MenubarTriggerLineHeight float32 = 20

	// MenubarMenuMinWidth is the drop-down panel minimum width in px
	// (min-w-[12rem]). Wider than the dropdown-menu default (metrics.Menu's
	// 128px) per the menubar registry entry.
	MenubarMenuMinWidth float32 = 192

	// MenubarSideOffset is the gap in px between the trigger and the panel
	// opened below it (sideOffset = 4, matching the menu family).
	MenubarSideOffset float32 = 4
)

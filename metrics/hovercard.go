package metrics

// HoverCard holds the exact pixel constants for the shadcn HoverCard content
// (the hover-triggered cousin of Popover). HoverCard is not enumerated in
// docs/research/03-shadcn-pixel-spec.md §5; these values are transcribed from
// the shadcn new-york-v4 hover-card registry entry, whose content class is:
//
//	"z-50 w-64 origin-(--radix-hover-card-content-transform-origin)
//	 rounded-md border bg-popover p-4 text-popover-foreground shadow-md
//	 outline-hidden ..."  (sideOffset 4, align center)
//
// → width w-64 = 256px, padding p-4 = 16px, radius rounded-md (RadiusMD),
// 1px border, bg --popover, shadow-md, side offset 4px. The HoverCard root
// ships openDelay 700ms / closeDelay 300ms; graft honors those via ctx.Now().
//
// Radius routes through Theme.RadiusMD(); shadow is metrics.ShadowMD. Only
// literals live here.
const (
	// HoverCardWidth is the content width in px (w-64).
	HoverCardWidth float32 = 256

	// HoverCardPadding is the content padding in px (p-4).
	HoverCardPadding float32 = 16

	// HoverCardBorderWidth is the content border width in px (border).
	HoverCardBorderWidth float32 = 1

	// HoverCardSideOffset is the gap between trigger and content in px
	// (sideOffset = 4).
	HoverCardSideOffset float32 = 4

	// HoverCardOpenDelayMillis is the hover-open delay in ms (openDelay 700).
	HoverCardOpenDelayMillis = 700

	// HoverCardCloseDelayMillis is the close delay after the pointer leaves in
	// ms (closeDelay 300).
	HoverCardCloseDelayMillis = 300
)

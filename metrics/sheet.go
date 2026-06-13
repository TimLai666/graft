package metrics

// Sheet holds the exact pixel constants for the shadcn Sheet — an
// edge-anchored modal panel (docs/research/03-shadcn-pixel-spec.md §Sheet).
//
// Overlay (verbatim, identical to Dialog): "fixed inset-0 z-50 bg-black/50"
// → pure black @50% in BOTH modes (reuses metrics.OverlayAlpha).
//
// Content base (verbatim): "fixed z-50 flex flex-col gap-4 bg-background
// shadow-lg transition ease-in-out data-[state=closed]:animate-out
// data-[state=closed]:duration-300 data-[state=open]:animate-in
// data-[state=open]:duration-500" → bg --background, shadow-lg, internal
// gap-4 (16px), open anim 500ms / close 300ms ease-in-out slide.
//
// Sides (verbatim):
//
//	right/left: "inset-y-0 ... h-full w-3/4 border-l|border-r ... sm:max-w-sm"
//	            → full height, width 75% of viewport, max-w-sm = 384px, 1px
//	              border on the inner edge.
//	top/bottom: "inset-x-0 ... h-auto border-b|border-t" → full width, auto
//	            height, 1px border on the inner edge.
//
// Header (verbatim): "flex flex-col gap-1.5 p-4" → padding 16px, gap 6px.
// Footer (verbatim): "mt-auto flex flex-col gap-2 p-4" → padding 16px, gap 8px.
// Title: "font-semibold text-foreground" (text-lg in shadcn ⇒ 18px/600).
// Description: "text-sm text-muted-foreground" (14px muted).
// Close button: same as Dialog (top/right 16px, X icon 16px, opacity 70→100).
const (
	// SheetMaxWidth is the left/right panel max width in px (sm:max-w-sm).
	SheetMaxWidth float32 = 384

	// SheetWidthFraction is the left/right panel width as a fraction of the
	// viewport width (w-3/4 = 75%).
	SheetWidthFraction float32 = 0.75

	// SheetHeightFraction is the top/bottom panel height as a fraction of the
	// viewport height. shadcn uses h-auto (content-driven); for a deterministic
	// settled render graft caps it so a content-light panel does not span the
	// whole viewport. This is graft-specific (noted): the shadcn class is
	// h-auto, which content already drives, so this acts only as a max.
	SheetHeightFraction float32 = 0.5

	// SheetBorderWidth is the panel border width in px (border = 1px). Only the
	// inner edge facing the viewport is drawn.
	SheetBorderWidth float32 = 1

	// SheetGap is the vertical gap between top-level content sections in px
	// (flex flex-col gap-4 = 16px).
	SheetGap float32 = 16

	// SheetPadding is the content padding applied around the column in px. The
	// header/footer carry p-4 (16px); graft applies the same 16px padding to
	// the whole content column so bare children align with the sections.
	SheetPadding float32 = 16

	// SheetHeaderGap is the gap between title and description in px
	// (flex flex-col gap-1.5 = 6px).
	SheetHeaderGap float32 = 6

	// SheetFooterGap is the gap between footer buttons in px
	// (flex flex-col gap-2 = 8px).
	SheetFooterGap float32 = 8

	// SheetTitleFontSize is the title size in px (text-lg).
	SheetTitleFontSize float32 = 18

	// SheetTitleWeight is the title weight (font-semibold).
	SheetTitleWeight int = 600

	// SheetTitleLineHeight is the title line box in px (leading-none ⇒ size).
	SheetTitleLineHeight float32 = 18

	// SheetDescriptionFontSize is the description size in px (text-sm).
	SheetDescriptionFontSize float32 = 14

	// SheetDescriptionLineHeight is the description line box in px (text-sm
	// leading = 20px).
	SheetDescriptionLineHeight float32 = 20

	// SheetCloseInset is the distance from the top-right corner to the close
	// button in px (top-4 right-4).
	SheetCloseInset float32 = 16

	// SheetCloseIconSize is the X icon size in px (size-4).
	SheetCloseIconSize float32 = 16

	// SheetCloseHitPad is the padding around the X icon forming the hit/hover
	// area in px (the button's rounded-xs box).
	SheetCloseHitPad float32 = 2

	// SheetCloseIdleOpacity is the close button's resting opacity (opacity-70).
	SheetCloseIdleOpacity float32 = 0.7

	// SheetOpenDurationMillis is the slide-in animation duration in ms
	// (data-[state=open]:duration-500). Goldens render the SETTLED open state;
	// this documents the spec timing.
	SheetOpenDurationMillis = 500

	// SheetCloseDurationMillis is the slide-out animation duration in ms
	// (data-[state=closed]:duration-300).
	SheetCloseDurationMillis = 300
)

package metrics

// Drawer holds the exact pixel constants for the shadcn Drawer — a vaul-based
// modal panel that slides in from a viewport edge, defaulting to the bottom
// (docs/research/03-shadcn-pixel-spec.md §Drawer).
//
// Overlay (verbatim, identical to Dialog/Sheet): "fixed inset-0 z-50
// bg-black/50" → pure black @50% in BOTH modes (reuses metrics.OverlayAlpha).
//
// Content base (verbatim): "group/drawer-content bg-background fixed z-50 flex
// h-auto flex-col" with side variants. Bottom (default):
// "data-[vaul-drawer-direction=bottom]:inset-x-0
// data-[vaul-drawer-direction=bottom]:bottom-0
// data-[vaul-drawer-direction=bottom]:mt-24
// data-[vaul-drawer-direction=bottom]:max-h-[80vh]
// data-[vaul-drawer-direction=bottom]:rounded-t-lg
// data-[vaul-drawer-direction=bottom]:border-t" → full width, anchored to the
// bottom, max height 80% of the viewport, rounded TOP corners (rounded-t-lg →
// theme RadiusLG) and a 1px top border. Top mirrors it (rounded-b-lg,
// border-b). Left/right span full height, width up to 3/4 / sm:max-w-sm with a
// single inner-edge border and no corner radius.
//
// Handle / grabber (verbatim, bottom only): "bg-muted mx-auto mt-4 hidden h-2
// w-[100px] shrink-0 rounded-full
// group-data-[vaul-drawer-direction=bottom]/drawer-content:block" → a centered
// muted pill 100×8px, 16px below the top edge, fully rounded.
//
// Header (verbatim): "flex flex-col gap-1.5 p-4" — but vaul's drawer header
// centers text on mobile and left-aligns on desktop ("group-data-[...]:text-left
// md:text-left"); graft left-aligns. Footer (verbatim): "mt-auto flex flex-col
// gap-2 p-4". Title: "text-foreground font-semibold". Description:
// "text-muted-foreground text-sm".
//
// Unlike Sheet, the Drawer has NO close X button (vaul dismisses via the
// grabber drag-down, backdrop click, or Esc).
const (
	// DrawerMaxWidth is the left/right panel max width in px (sm:max-w-sm).
	DrawerMaxWidth float32 = 384

	// DrawerWidthFraction is the left/right panel width as a fraction of the
	// viewport width (w-3/4 = 75%).
	DrawerWidthFraction float32 = 0.75

	// DrawerHeightFraction is the top/bottom panel max height as a fraction of
	// the viewport height (max-h-[80vh]). vaul uses content-driven h-auto; this
	// caps it so a deterministic settled render does not span the whole
	// viewport.
	DrawerHeightFraction float32 = 0.8

	// DrawerBorderWidth is the panel border width in px (border = 1px). Only the
	// inner edge facing the viewport is drawn.
	DrawerBorderWidth float32 = 1

	// DrawerGap is the vertical gap between top-level content sections in px
	// (flex flex-col gap-4 → 16px on the inner column; vaul stacks header/body/
	// footer; graft uses the same 16px Sheet rhythm).
	DrawerGap float32 = 16

	// DrawerPadding is the content padding applied around the column in px. The
	// header/footer carry p-4 (16px); graft applies the same 16px padding to the
	// whole content column so bare children align with the sections.
	DrawerPadding float32 = 16

	// DrawerHandleTopInset is the gap from the panel's top edge to the grabber
	// in px (mt-4 = 16px). Bottom drawers only.
	DrawerHandleTopInset float32 = 16

	// DrawerHandleWidth is the grabber bar width in px (w-[100px]).
	DrawerHandleWidth float32 = 100

	// DrawerHandleHeight is the grabber bar height in px (h-2 = 8px).
	DrawerHandleHeight float32 = 8

	// DrawerHeaderGap is the gap between title and description in px
	// (flex flex-col gap-1.5 = 6px).
	DrawerHeaderGap float32 = 6

	// DrawerFooterGap is the gap between footer buttons in px
	// (flex flex-col gap-2 = 8px).
	DrawerFooterGap float32 = 8

	// DrawerTitleFontSize is the title size in px. vaul's drawer title is the
	// shadcn DialogTitle scale (text-lg ⇒ 18px); graft matches Sheet at 18px.
	DrawerTitleFontSize float32 = 18

	// DrawerTitleWeight is the title weight (font-semibold).
	DrawerTitleWeight int = 600

	// DrawerTitleLineHeight is the title line box in px (leading-none ⇒ size).
	DrawerTitleLineHeight float32 = 18

	// DrawerDescriptionFontSize is the description size in px (text-sm).
	DrawerDescriptionFontSize float32 = 14

	// DrawerDescriptionLineHeight is the description line box in px (text-sm
	// leading = 20px).
	DrawerDescriptionLineHeight float32 = 20

	// DrawerOpenDurationMillis is the slide-in animation duration in ms. vaul's
	// default open transition is ~500ms with an ease-out curve; goldens render
	// the SETTLED open state, so this only documents the spec timing.
	DrawerOpenDurationMillis = 500

	// DrawerCloseDurationMillis is the slide-out animation duration in ms.
	DrawerCloseDurationMillis = 300

	// DrawerDismissDragFraction is the fraction of the panel's drag-axis extent
	// a pointer drag must travel before release dismisses the drawer. A shorter
	// drag snaps the panel back to its settled position. vaul uses a velocity +
	// distance heuristic; graft uses a simple distance threshold.
	DrawerDismissDragFraction float32 = 0.45
)

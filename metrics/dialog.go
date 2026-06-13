package metrics

// Dialog holds the exact pixel constants for the shadcn Dialog (and the
// AlertDialog, which shares them) — overlay, content card, sections, and the
// close button (docs/research/03-shadcn-pixel-spec.md §Dialog).
//
// Content (verbatim): "fixed top-[50%] left-[50%] z-50 grid w-full
// max-w-[calc(100%-2rem)] ... gap-4 rounded-lg border bg-background p-6
// shadow-lg ... sm:max-w-lg" → centered, max-w 512 (mobile: viewport−32),
// p-6 24px, internal gap-4 16px, rounded-lg, 1px border, bg --background,
// shadow-lg.
//
// Close button (verbatim): "absolute top-4 right-4 rounded-xs opacity-70 ...
// hover:opacity-100 focus:ring-2 focus:ring-ring focus:ring-offset-2 ...
// [&_svg:not([class*='size-'])]:size-4" → 16px from corner, rounded-xs, X icon
// 16px, opacity 70→100, focus = 2px ring + 2px offset (gap = --background).
//
// Header "flex flex-col gap-2"; Title "text-lg leading-none font-semibold"
// (18/18, 600); Description "text-sm text-muted-foreground" (14, muted);
// Footer "flex flex-col-reverse gap-2 sm:flex-row sm:justify-end" (row, gap 8,
// right-aligned). Radii route through theme (RadiusLG/RadiusXS); only literals
// live here.
const (
	// DialogMaxWidth is the content card max width in px (sm:max-w-lg = 32rem).
	DialogMaxWidth float32 = 512

	// DialogViewportMargin is the gap kept on each side on small viewports in
	// px (max-w-[calc(100%-2rem)] → 2rem total = 16px each side).
	DialogViewportMargin float32 = 16

	// DialogPadding is the content card padding in px (p-6 = 1.5rem).
	DialogPadding float32 = 24

	// DialogGap is the vertical gap between content sections in px (gap-4).
	DialogGap float32 = 16

	// DialogBorderWidth is the content card border width in px (border).
	DialogBorderWidth float32 = 1

	// OverlayAlpha is the modal backdrop opacity — pure black @50% in BOTH
	// modes (bg-black/50).
	OverlayAlpha float32 = 0.5

	// DialogHeaderGap is the gap between title and description in px (gap-2).
	DialogHeaderGap float32 = 8

	// DialogFooterGap is the gap between footer buttons in px (gap-2).
	DialogFooterGap float32 = 8

	// DialogTitleFontSize is the title size in px (text-lg).
	DialogTitleFontSize float32 = 18

	// DialogTitleWeight is the title weight (font-semibold).
	DialogTitleWeight int = 600

	// DialogDescriptionFontSize is the description size in px (text-sm).
	DialogDescriptionFontSize float32 = 14

	// DialogDescriptionLineHeight is the description line box in px (leading
	// of text-sm = 20px).
	DialogDescriptionLineHeight float32 = 20

	// DialogCloseInset is the distance from the top-right corner to the close
	// button in px (top-4 right-4).
	DialogCloseInset float32 = 16

	// DialogCloseIconSize is the X icon size in px (size-4).
	DialogCloseIconSize float32 = 16

	// DialogCloseHitPad is the padding around the X icon forming the hit/hover
	// area in px (the button's rounded-xs box).
	DialogCloseHitPad float32 = 2

	// DialogCloseIdleOpacity is the close button's resting opacity (opacity-70).
	DialogCloseIdleOpacity float32 = 0.7
)

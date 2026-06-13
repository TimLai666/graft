package metrics

// Empty metrics.
//
// Source: shadcn new-york-v4 empty.tsx.
//
//	Empty (root): "flex min-w-0 flex-1 flex-col items-center justify-center
//	               gap-6 rounded-lg border-dashed p-6 text-center text-balance
//	               md:p-12"
//	EmptyHeader:  "flex max-w-sm flex-col items-center gap-2 text-center"
//	EmptyMedia base:  "mb-2 flex shrink-0 items-center justify-center
//	                   [&_svg]:pointer-events-none [&_svg]:shrink-0"
//	EmptyMedia icon:  "flex size-10 shrink-0 items-center justify-center
//	                   rounded-lg bg-muted text-foreground
//	                   [&_svg:not([class*='size-'])]:size-6"
//	EmptyTitle:   "text-lg font-medium tracking-tight"
//	EmptyDescription: "text-sm/relaxed text-muted-foreground ..."
//	EmptyContent: "flex w-full max-w-sm min-w-0 flex-col items-center gap-4
//	               text-sm text-balance"
//
// The root radius is rounded-lg (Theme.RadiusLG). The media icon tile radius is
// rounded-lg as well. graft renders the base (non-md) padding (p-6 = 24px).
const (
	// EmptyGap is the gap between header / media / content sections in px
	// (gap-6).
	EmptyGap float32 = 24

	// EmptyPadding is the root padding in px (p-6; the md:p-12 breakpoint is
	// not modeled).
	EmptyPadding float32 = 24

	// EmptyMaxWidth caps the header and content width in px (max-w-sm).
	EmptyMaxWidth float32 = 384

	// EmptyHeaderGap is the gap between title and description in px (gap-2).
	EmptyHeaderGap float32 = 8

	// EmptyContentGap is the gap between content children in px (gap-4).
	EmptyContentGap float32 = 16

	// EmptyMediaSize is the icon tile size in px (size-10).
	EmptyMediaSize float32 = 40

	// EmptyMediaIconSize is the icon glyph size inside the tile in px
	// (size-6).
	EmptyMediaIconSize float32 = 24

	// EmptyMediaMarginBottom is the margin below the media tile in px (mb-2).
	EmptyMediaMarginBottom float32 = 8

	// EmptyTitleFontSize is the title font size in px (text-lg).
	EmptyTitleFontSize float32 = 18

	// EmptyTitleLineHeight is the title line height in px (text-lg =
	// 18px/28px).
	EmptyTitleLineHeight float32 = 28

	// EmptyTitleFontWeight is the title weight (font-medium).
	EmptyTitleFontWeight = 500

	// EmptyDescriptionFontSize is the description font size in px (text-sm).
	EmptyDescriptionFontSize float32 = 14

	// EmptyDescriptionLineHeight is the description line height in px
	// (text-sm/relaxed = 14px * leading-relaxed 1.625 = 22.75, rounded).
	EmptyDescriptionLineHeight float32 = 23
)

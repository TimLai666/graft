package metrics

// Breadcrumb holds the pixel constants for the shadcn Breadcrumb component
// (docs/research/03-shadcn-pixel-spec.md §5 "Breadcrumb"), transcribed from the
// new-york-v4 registry class strings:
//
//	List      "flex flex-wrap items-center gap-1.5 text-sm break-words
//	           text-muted-foreground sm:gap-2.5"
//	Item      "inline-flex items-center gap-1.5"
//	Link      "transition-colors hover:text-foreground"
//	Page      "font-normal text-foreground"
//	Separator "[&>svg]:size-3.5"  (ChevronRight 14px)
//	Ellipsis  "flex size-9 items-center justify-center"  (16px icon)
//
// The package is pure data with no imports.
var Breadcrumb = struct {
	// Gap is the horizontal spacing in px between list elements (gap-1.5 → 6;
	// sm:gap-2.5 → 10, not modeled — graft uses the base 6).
	Gap float32

	// FontSize is the breadcrumb text size in px (text-sm → 14).
	FontSize float32

	// LinkFontWeight is the link/separator text weight (normal → 400).
	LinkFontWeight int

	// PageFontWeight is the current-page weight (font-normal → 400).
	PageFontWeight int

	// SeparatorSize is the chevron-right separator icon box in px
	// ([&>svg]:size-3.5 → 14).
	SeparatorSize float32

	// EllipsisBox is the ellipsis hit-area box in px (size-9 → 36).
	EllipsisBox float32

	// EllipsisIcon is the ellipsis glyph size in px (16).
	EllipsisIcon float32
}{
	Gap:            6,
	FontSize:       14,
	LinkFontWeight: 400,
	PageFontWeight: 400,
	SeparatorSize:  14,
	EllipsisBox:    36,
	EllipsisIcon:   16,
}

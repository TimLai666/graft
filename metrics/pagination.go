package metrics

// Pagination holds the pixel constants for the shadcn Pagination component
// (docs/research/03-shadcn-pixel-spec.md §5 "Pagination"), transcribed from the
// new-york-v4 registry class strings:
//
//	Content  "flex flex-row items-center gap-1"
//	Link     buttonVariants{ variant: isActive ? "outline" : "ghost",
//	          size: "icon" }  → 36×36
//	Prev/Next buttonVariants{ size: "default" } + "gap-1 px-2.5" with 16px
//	          chevrons
//	Ellipsis "flex size-9 items-center justify-center"  (36px box, 16px icon)
//
// All button metrics route through metrics.Button; only the layout gap and
// ellipsis box live here. The package is pure data with no imports.
var Pagination = struct {
	// Gap is the horizontal spacing in px between pagination items (gap-1 → 4).
	Gap float32

	// EllipsisBox is the ellipsis box in px (size-9 → 36).
	EllipsisBox float32

	// EllipsisIcon is the ellipsis glyph size in px (16).
	EllipsisIcon float32
}{
	Gap:          4,
	EllipsisBox:  36,
	EllipsisIcon: 16,
}

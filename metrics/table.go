package metrics

// Table holds the exact pixel constants for the shadcn Table component
// (docs/research/03-shadcn-pixel-spec.md §5 "Table"), transcribed from the
// new-york-v4 registry class strings:
//
//	Table   "w-full caption-bottom text-sm"
//	Header  "[&_tr]:border-b"
//	Body    "[&_tr:last-child]:border-0"
//	Footer  "border-t bg-muted/50 font-medium [&>tr]:last:border-b-0"
//	Row     "border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted"
//	Head    "h-10 px-2 text-left align-middle font-medium whitespace-nowrap text-foreground"
//	Cell    "p-2 align-middle whitespace-nowrap"
//	Caption "mt-4 text-sm text-muted-foreground"
//
// The package is pure data with no imports; alphas route through draw helpers
// and colors through the active token set at draw time.
var Table = struct {
	// FontSize is the table text size in px (text-sm → 14).
	FontSize float32

	// HeadHeight is the header cell height in px (h-10 → 40).
	HeadHeight float32

	// HeadFontWeight is the header cell weight (font-medium → 500).
	HeadFontWeight int

	// HeadPadX is the header cell horizontal padding in px (px-2 → 8).
	HeadPadX float32

	// CellPad is the body/footer cell uniform padding in px (p-2 → 8).
	CellPad float32

	// CellFontWeight is the body cell weight (inherited normal → 400).
	CellFontWeight int

	// FooterFontWeight is the footer cell weight (font-medium → 500).
	FooterFontWeight int

	// BorderWidth is the row separator width in px (border-b / border-t → 1).
	BorderWidth float32

	// CaptionMarginTop is the gap above the caption in px (mt-4 → 16).
	CaptionMarginTop float32

	// CaptionFontSize is the caption text size in px (text-sm → 14).
	CaptionFontSize float32

	// RowHoverAlpha is the alpha applied to the Muted token for a hovered
	// body row (hover:bg-muted/50 → 0.50).
	RowHoverAlpha float32

	// FooterBgAlpha is the alpha applied to the Muted token for the footer
	// background (bg-muted/50 → 0.50).
	FooterBgAlpha float32
}{
	FontSize:         14,
	HeadHeight:       40,
	HeadFontWeight:   500,
	HeadPadX:         8,
	CellPad:          8,
	CellFontWeight:   400,
	FooterFontWeight: 500,
	BorderWidth:      1,
	CaptionMarginTop: 16,
	CaptionFontSize:  14,
	RowHoverAlpha:    0.5,
	FooterBgAlpha:    0.5,
}

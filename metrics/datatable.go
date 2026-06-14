package metrics

// DataTable holds the exact pixel constants for shadcn's DataTable recipe
// (docs/research/03-shadcn-pixel-spec.md §5 "Data Table"). The DataTable is a
// recipe layered over the presentational Table: it reuses every Table metric
// (head h-10, cell p-2, text-sm, 1px row separators — see metrics.Table) and
// adds the chrome the recipe introduces: a sortable-header arrow indicator, a
// leading checkbox selection column, and a pagination footer.
//
// In the React source the sort control is a ghost Button wrapping the header
// label plus an ArrowUpDown/ChevronsUpDown icon, the selection column is a
// Checkbox in both the head and each cell, and the footer is a flex row with a
// "N of M row(s) selected." caption on the left and Previous/Next buttons on
// the right (the live demo also shows a rows-per-page select; graft renders a
// fixed page size). The package is pure data with no imports.
var DataTable = struct {
	// SortIconSize is the sort-indicator arrow box in px. shadcn renders the
	// header sort control's icon at size-4 (16); the DataTable demo's header
	// uses the same 14px text-sm label as a Table head, so a 14px arrow keeps
	// the glyph optically matched to the label cap height.
	SortIconSize float32

	// SortIconGap is the gap between the header label and the sort arrow in px
	// (the header sort Button's gap-2 → 8).
	SortIconGap float32

	// CheckboxColWidth is the width of the leading selection column in px. The
	// 16px checkbox sits in a p-2 (8px) cell, so 16 + 2*8 = 32; the recipe pads
	// the column slightly wider (w-12 in the demo header) — 40px gives the box
	// comfortable breathing room while staying narrow.
	CheckboxColWidth float32

	// FooterHeight is the pagination footer band height in px. The footer is a
	// flex row of default-size (h-8 → 32) buttons with py-4 (16) vertical
	// padding around them: 32 + 2*16 = 64.
	FooterHeight float32

	// FooterPadY is the footer vertical padding in px (py-4 → 16).
	FooterPadY float32

	// FooterGap is the gap between the Previous and Next buttons in px
	// (space-x-2 → 8).
	FooterGap float32

	// FooterFontSize is the "N of M row(s) selected." caption size in px
	// (text-sm → 14).
	FooterFontSize float32

	// SelectedRowAlpha is the alpha applied to the Muted token for a selected
	// row background (data-[state=selected]:bg-muted → 1.0, but the Muted
	// token is already a low-contrast fill so it reads as a tint).
	SelectedRowAlpha float32
}{
	SortIconSize:     14,
	SortIconGap:      8,
	CheckboxColWidth: 40,
	FooterHeight:     64,
	FooterPadY:       16,
	FooterGap:        8,
	FooterFontSize:   14,
	SelectedRowAlpha: 1.0,
}

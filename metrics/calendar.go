package metrics

// Calendar holds the pixel metrics for the shadcn Calendar
// (react-day-picker v9 wrapper; docs/research/03-shadcn-pixel-spec.md §5
// "Calendar"). Sizes are driven by --cell-size (default --spacing(8) =
// 32px). Verbatim source class strings:
//
//	nav: "absolute inset-x-0 top-0 flex w-full items-center
//	      justify-between gap-1"
//	button_previous/next: "size-(--cell-size) p-0 ..."  (ghost icon, 32px)
//	month_caption: "flex h-(--cell-size) ... justify-center"
//	weekday: "flex-1 rounded-md text-[0.8rem] font-normal
//	          text-muted-foreground select-none"   (12.8px)
//	week: "mt-2 flex w-full"
//	day: "group/day relative aspect-square h-full w-full p-0
//	      text-center select-none"
//	DayButton: "... data-[selected-single=true]:bg-primary
//	            data-[selected-single=true]:text-primary-foreground
//	            data-[range-middle=true]:bg-accent
//	            data-[range-start=true]:rounded-md
//	            data-[range-start=true]:bg-primary ..."
//
// Selected day = --primary pill (rounded-md = 8px). Today = outline (1px
// --border) on an --accent tint. Outside-month days = muted. Day text is
// text-sm (14px). The grid is 6 rows × 7 columns of cell-size squares; the
// weekday header is one cell-size row, with mt-2 (8px) above each week row.
var Calendar = struct {
	// CellSize is the day/nav/weekday cell side (--cell-size = 32px).
	CellSize float32
	// Columns is the number of columns (7 weekdays).
	Columns int
	// Rows is the number of week rows rendered (always 6 for a stable grid).
	Rows int

	// WeekdayFontSize is the weekday header size (text-[0.8rem] = 12.8px).
	WeekdayFontSize float32
	// WeekdayFontWeight is the weekday weight (font-normal = 400).
	WeekdayFontWeight int

	// DayFontSize is the day-number size (text-sm = 14px).
	DayFontSize float32
	// DayFontWeight is the day weight (font-normal = 400).
	DayFontWeight int

	// WeekGap is the vertical gap above each week row (mt-2 = 8px).
	WeekGap float32
	// HeaderGap is the gap below the weekday header before the first week.
	HeaderGap float32

	// CaptionFontSize is the month/year caption size (text-sm = 14px).
	CaptionFontSize float32
	// CaptionFontWeight is the caption weight (font-medium = 500).
	CaptionFontWeight int
	// CaptionGap is the gap below the caption/nav row before the weekdays.
	CaptionGap float32

	// NavIconSize is the chevron icon size in the nav buttons (size-4 =
	// 16px).
	NavIconSize float32

	// TodayBorderWidth is the outline width on the "today" cell (1px).
	TodayBorderWidth float32
}{
	CellSize: 32, // --cell-size
	Columns:  7,
	Rows:     6,

	WeekdayFontSize:   12.8, // text-[0.8rem]
	WeekdayFontWeight: 400,  // font-normal

	DayFontSize:   14, // text-sm
	DayFontWeight: 400,

	WeekGap:   8, // mt-2
	HeaderGap: 8,

	CaptionFontSize:   14,  // text-sm
	CaptionFontWeight: 500, // font-medium
	CaptionGap:        8,

	NavIconSize: 16, // size-4

	TodayBorderWidth: 1,
}

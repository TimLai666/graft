package metrics

// Alert holds the exact pixel constants for the shadcn Alert composite
// (docs/research/03-shadcn-pixel-spec.md §5 "Alert").
//
// Source (new-york-v4 alertVariants base):
//
//	"relative grid w-full grid-cols-[0_1fr] items-start gap-y-0.5 rounded-lg
//	 border px-4 py-3 text-sm has-[>svg]:grid-cols-[calc(var(--spacing)*4)_1fr]
//	 has-[>svg]:gap-x-3 [&>svg]:size-4 [&>svg]:translate-y-0.5 [&>svg]:text-current"
//
// Title: "col-start-2 line-clamp-1 min-h-4 font-medium tracking-tight"
// Description: "col-start-2 grid justify-items-start gap-1 text-sm
//
//	text-muted-foreground [&_p]:leading-relaxed"
//
// The radius routes through the theme (rounded-lg → t.RadiusLG()); only the
// literal px values live here.
var Alert = struct {
	// PadX is the horizontal padding in px (px-4 = 16px).
	PadX float32

	// PadY is the vertical padding in px (py-3 = 12px).
	PadY float32

	// BorderWidth is the alert border width in px (border = 1px).
	BorderWidth float32

	// RowGap is the vertical gap between title and description in px
	// (gap-y-0.5 = 2px).
	RowGap float32

	// IconSize is the lucide icon side length in px ([&>svg]:size-4 = 16px).
	IconSize float32

	// IconColumn is the width of the icon grid column in px
	// (has-[>svg]:grid-cols-[calc(var(--spacing)*4)_1fr] → --spacing*4 = 16px).
	IconColumn float32

	// IconGap is the horizontal gap between the icon column and content
	// column in px (has-[>svg]:gap-x-3 = 12px).
	IconGap float32

	// IconNudgeY is the vertical icon offset in px ([&>svg]:translate-y-0.5
	// = +2px), aligning the 16px icon optically with the title baseline.
	IconNudgeY float32

	// TitleFontSize is the title font size in px (text-sm = 14px).
	TitleFontSize float32

	// TitleLineHeight is the title line box height in px. shadcn uses the
	// text-sm line height (leading-5 = 20px); min-h-4 (16px) only sets a
	// floor, which a 20px line already clears.
	TitleLineHeight float32

	// TitleWeight is the title font weight (font-medium = 500).
	TitleWeight int

	// DescFontSize is the description font size in px (text-sm = 14px).
	DescFontSize float32

	// DescLineHeight is the description line box height in px (leading-5 =
	// 20px; [&_p]:leading-relaxed only applies to nested <p>, not the
	// single-line graft description).
	DescLineHeight float32

	// DescDestructiveAlpha is the alpha applied to the Destructive token for
	// the description text in the destructive variant
	// (*:data-[slot=alert-description]:text-destructive/90 = 90%).
	DescDestructiveAlpha float32
}{
	PadX:                 16,
	PadY:                 12,
	BorderWidth:          1,
	RowGap:               2,
	IconSize:             16,
	IconColumn:           16,
	IconGap:              12,
	IconNudgeY:           2,
	TitleFontSize:        14,
	TitleLineHeight:      20,
	TitleWeight:          500,
	DescFontSize:         14,
	DescLineHeight:       20,
	DescDestructiveAlpha: 0.9,
}

package metrics

// Field metrics transcribed from the shadcn new-york-v4 field.tsx registry
// entry. The Field family is composition over primitives (Box stacks) plus
// the existing Label; only the vertical/horizontal gaps, the legend margin,
// and the text sizes/weights/line-heights live here. Radius routes through
// the theme where a sub-control needs one. The package is pure data.
//
// Source class strings (verbatim from field.tsx):
//
//	FieldSet:     "flex flex-col gap-6"
//	FieldLegend:  "mb-3 font-medium" + data-[variant=legend]:text-base
//	                                 + data-[variant=label]:text-sm
//	FieldGroup:   "flex w-full flex-col gap-7"
//	Field:        "flex w-full gap-3"  (vertical: flex-col; horizontal: flex-row items-center)
//	FieldContent: "flex flex-1 flex-col gap-1.5 leading-snug"
//	FieldLabel:   Label + "flex w-fit gap-2 leading-snug"
//	FieldTitle:   "text-sm leading-snug font-medium"
//	FieldDescription: "text-sm leading-normal font-normal text-muted-foreground"
//	FieldError:   "text-sm font-normal text-destructive"
//	FieldSeparator: "-my-2 h-5 text-sm" with an absolute Separator at top-1/2
var Field = struct {
	// SetGap is the vertical gap between fields inside a FieldSet in px
	// (gap-6 = 24).
	SetGap float32

	// GroupGap is the vertical gap between fields inside a FieldGroup in px
	// (gap-7 = 28).
	GroupGap float32

	// Gap is the gap between a field's own parts (label, control,
	// description, error) in px (gap-3 = 12).
	Gap float32

	// ContentGap is the vertical gap inside a FieldContent column in px
	// (gap-1.5 = 6).
	ContentGap float32

	// LabelGap is the gap between a field label's text and an adjacent icon
	// in px (gap-2 = 8).
	LabelGap float32

	// LegendMarginBottom is the space below a FieldLegend in px (mb-3 = 12).
	LegendMarginBottom float32

	// LegendFontSize is the legend text size in px for the default "legend"
	// variant (text-base = 16).
	LegendFontSize float32

	// LegendLabelFontSize is the legend text size in px for the "label"
	// variant (text-sm = 14).
	LegendLabelFontSize float32

	// LegendFontWeight is the legend weight (font-medium = 500).
	LegendFontWeight int

	// LabelFontSize is the field label text size in px (text-sm = 14).
	LabelFontSize float32

	// LabelFontWeight is the field label weight (font-medium = 500).
	LabelFontWeight int

	// LabelLineHeight is the field label line box in px (leading-snug at
	// 14px ≈ 14 × 1.375 = 19.25).
	LabelLineHeight float32

	// DescriptionFontSize is the field description size in px (text-sm = 14).
	DescriptionFontSize float32

	// DescriptionFontWeight is the description weight (font-normal = 400).
	DescriptionFontWeight int

	// DescriptionLineHeight is the description line box in px (leading-normal
	// at 14px = 14 × 1.5 = 21).
	DescriptionLineHeight float32

	// ErrorFontSize is the field error size in px (text-sm = 14).
	ErrorFontSize float32

	// ErrorFontWeight is the error weight (font-normal = 400).
	ErrorFontWeight int

	// ErrorLineHeight is the error line box in px (text-sm default = 20).
	ErrorLineHeight float32

	// SeparatorHeight is the FieldSeparator row height in px (h-5 = 20).
	SeparatorHeight float32

	// SeparatorTextGap is the horizontal padding around centered separator
	// text in px (px-2 = 8).
	SeparatorTextGap float32

	// SeparatorThickness is the rule thickness in px (Separator h-px = 1).
	SeparatorThickness float32
}{
	SetGap:                24,
	GroupGap:              28,
	Gap:                   12,
	ContentGap:            6,
	LabelGap:              8,
	LegendMarginBottom:    12,
	LegendFontSize:        16,
	LegendLabelFontSize:   14,
	LegendFontWeight:      500,
	LabelFontSize:         14,
	LabelFontWeight:       500,
	LabelLineHeight:       19.25,
	DescriptionFontSize:   14,
	DescriptionFontWeight: 400,
	DescriptionLineHeight: 21,
	ErrorFontSize:         14,
	ErrorFontWeight:       400,
	ErrorLineHeight:       20,
	SeparatorHeight:       20,
	SeparatorTextGap:      8,
	SeparatorThickness:    1,
}

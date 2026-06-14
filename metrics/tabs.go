package metrics

// Tabs holds the exact px metrics for the shadcn Tabs component
// (docs/research/03-shadcn-pixel-spec.md section 5 "Tabs").
//
// Root: "group/tabs flex gap-2 data-[orientation=horizontal]:flex-col".
// List: "inline-flex w-fit items-center justify-center rounded-lg p-[3px]
// text-muted-foreground group-data-[orientation=horizontal]/tabs:h-8"
// with variants default:"bg-muted" / line:"gap-1 bg-transparent".
// Trigger: "relative inline-flex h-[calc(100%-1px)] flex-1 items-center
// justify-center gap-1.5 rounded-md border border-transparent px-1.5 py-0.5
// text-sm font-medium whitespace-nowrap text-foreground/60 ...
// dark:text-muted-foreground dark:hover:text-foreground
// data-[state=active]:bg-background data-[state=active]:text-foreground
// group-data-[variant=default]/tabs-list:data-[state=active]:shadow-sm
// dark:data-[state=active]:border-input dark:data-[state=active]:bg-input/30"
// plus the line-variant underline "after:absolute after:inset-x-0
// after:bottom-[-5px] after:h-0.5 after:bg-foreground".
//
// Radii route through the theme: list = RadiusLG, trigger = RadiusMD.
var Tabs = struct {
	// RootGap is the gap between the list and the content (root gap-2).
	RootGap float32

	// ListHeight is the tab list height in px (h-8 = 32px, horizontal).
	ListHeight float32

	// ListPadding is the list inner padding in px (p-[3px]).
	ListPadding float32

	// LineGap is the gap between triggers in the line variant (gap-1).
	// The default variant has no gap.
	LineGap float32

	// TriggerHeightInset is subtracted from the list's inner height to
	// get the trigger height (h-[calc(100%-1px)] -> 32-6-1 = 25px).
	TriggerHeightInset float32

	// TriggerPadX is the trigger horizontal padding in px (px-1.5).
	TriggerPadX float32

	// TriggerPadY is the trigger vertical padding in px (py-0.5).
	TriggerPadY float32

	// TriggerFontSize is the trigger label size in px (text-sm).
	TriggerFontSize float32

	// TriggerLineHeight is the trigger label line box in px (text-sm ->
	// 20px line height).
	TriggerLineHeight float32

	// TriggerFontWeight is the trigger label weight (font-medium = 500).
	TriggerFontWeight int

	// TriggerBorderWidth is the trigger border width in px (border;
	// transparent except dark-mode active = --input, focus = --ring).
	TriggerBorderWidth float32

	// IdleTextOpacity is the idle trigger text alpha in light mode
	// (text-foreground/60).
	IdleTextOpacity float32

	// DarkActiveBgOpacity is the alpha multiplier on the --input token
	// for the dark-mode active trigger fill (dark:bg-input/30; --input
	// already carries alpha in dark mode, so this multiplies).
	DarkActiveBgOpacity float32

	// UnderlineHeight is the line-variant active underline thickness
	// (after:h-0.5 = 2px) in --foreground.
	UnderlineHeight float32

	// UnderlineDrop is how far the underline's bottom edge sits below
	// the trigger's bottom edge (after:bottom-[-5px]).
	UnderlineDrop float32
}{
	RootGap:             8,   // gap-2
	ListHeight:          32,  // h-8
	ListPadding:         3,   // p-[3px]
	LineGap:             4,   // gap-1 (line variant)
	TriggerHeightInset:  1,   // h-[calc(100%-1px)]
	TriggerPadX:         6,   // px-1.5
	TriggerPadY:         2,   // py-0.5
	TriggerFontSize:     14,  // text-sm
	TriggerLineHeight:   20,  // text-sm line height
	TriggerFontWeight:   500, // font-medium
	TriggerBorderWidth:  1,   // border
	IdleTextOpacity:     0.6, // text-foreground/60
	DarkActiveBgOpacity: 0.3, // dark:bg-input/30
	UnderlineHeight:     2,   // after:h-0.5
	UnderlineDrop:       5,   // after:bottom-[-5px]
}

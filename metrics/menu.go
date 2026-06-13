package metrics

// Menu holds the pixel metrics for the shadcn menu family (DropdownMenu,
// ContextMenu, Menubar drop-downs), shared by the internal menu engine and
// the dropdownmenu.go composite. Every value is annotated with its source
// Tailwind class string (Report 3 §5 "Dropdown Menu"). Radii route through
// the theme (RadiusMD for the panel, RadiusSM for items).
//
// Source content:
//
//	"... min-w-[8rem] ... rounded-md border bg-popover p-1
//	 text-popover-foreground shadow-md ..."  (SubContent: shadow-lg)
//
// Source item:
//
//	"relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5
//	 text-sm ... focus:bg-accent focus:text-accent-foreground
//	 data-[disabled]:opacity-50 data-[inset]:pl-8
//	 data-[variant=destructive]:text-destructive
//	 data-[variant=destructive]:focus:bg-destructive/10 ...
//	 dark:data-[variant=destructive]:focus:bg-destructive/20 ...
//	 [&_svg]:size-4 [&_svg]:text-muted-foreground"
//
// Source checkbox/radio item: "... py-1.5 pr-2 pl-8 ..." with indicator at
// left-2 in a 14px box; check 16px, radio dot CircleIcon size-2 (8px).
//
// Source label: "px-2 py-1.5 text-sm font-medium data-[inset]:pl-8".
// Source separator: "-mx-1 my-1 h-px bg-border".
// Source shortcut: "ml-auto text-xs tracking-widest text-muted-foreground".
var Menu = struct {
	// MinWidth is the panel minimum width (min-w-[8rem] = 128px).
	MinWidth float32
	// Pad is the panel inner padding (p-1 = 4px).
	Pad float32
	// BorderWidth is the panel border width (border = 1px).
	BorderWidth float32
	// SideOffset is the placement side offset toward the trigger (4px).
	SideOffset float32

	// ItemHeight is the rendered item/checkbox/radio/label row height:
	// py-1.5 (12px) around a text-sm line box (leading 20px) ⇒ 32px.
	ItemHeight float32
	// ItemPadX is the item horizontal padding (px-2 = 8px).
	ItemPadX float32
	// ItemPadY is the item vertical padding (py-1.5 = 6px).
	ItemPadY float32
	// ItemGap is the gap between an item's icon/text (gap-2 = 8px).
	ItemGap float32
	// InsetPadLeft is the left padding for inset items, checkbox items, and
	// radio items (data-[inset]:pl-8 / pl-8 = 32px).
	InsetPadLeft float32
	// IconSize is the leading icon size (size-4 = 16px).
	IconSize float32
	// IndicatorLeft is the checkbox/radio indicator left offset (left-2 =
	// 8px).
	IndicatorLeft float32
	// CheckSize is the checkbox check icon size (size-4 = 16px).
	CheckSize float32
	// RadioDotSize is the radio dot size (CircleIcon size-2 = 8px).
	RadioDotSize float32

	// FontSize is the item/label font size (text-sm = 14px).
	FontSize float32
	// LabelFontWeight is the label weight (font-medium = 500).
	LabelFontWeight int
	// ItemFontWeight is the item weight (text-sm, no font-medium = 400).
	ItemFontWeight int

	// ShortcutFontSize is the shortcut font size (text-xs = 12px).
	ShortcutFontSize float32

	// SeparatorHeight is the divider row height: 1px line + my-1 (8px) ⇒
	// 9px.
	SeparatorHeight float32
	// SeparatorInset is the divider horizontal inset; -mx-1 makes the line
	// span past the p-1 padding so it reaches the panel edges. We inset by
	// the panel padding so it aligns to item content edges.
	SeparatorInset float32

	// DestructiveFocusAlphaLight is the destructive item focus fill alpha
	// in light mode (focus:bg-destructive/10).
	DestructiveFocusAlphaLight float32
	// DestructiveFocusAlphaDark is the destructive item focus fill alpha in
	// dark mode (dark:...:bg-destructive/20).
	DestructiveFocusAlphaDark float32
}{
	MinWidth:    128, // min-w-[8rem]
	Pad:         4,   // p-1
	BorderWidth: 1,   // border
	SideOffset:  4,

	ItemHeight:    32, // py-1.5 (12) + leading-5 (20)
	ItemPadX:      8,  // px-2
	ItemPadY:      6,  // py-1.5
	ItemGap:       8,  // gap-2
	InsetPadLeft:  32, // pl-8
	IconSize:      16, // size-4
	IndicatorLeft: 8,  // left-2
	CheckSize:     16, // size-4
	RadioDotSize:  8,  // size-2

	FontSize:        14,  // text-sm
	LabelFontWeight: 500, // font-medium
	ItemFontWeight:  400,

	ShortcutFontSize: 12, // text-xs

	SeparatorHeight: 9, // 1px + my-1 (8)
	SeparatorInset:  4, // align to p-1 content edge

	DestructiveFocusAlphaLight: 0.10, // destructive/10
	DestructiveFocusAlphaDark:  0.20, // destructive/20
}

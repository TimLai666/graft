package metrics

// Select holds the pixel metrics for the shadcn Select component
// (trigger + content + item). Values are annotated with their source
// Tailwind class string (current ui.shadcn.com live style). Radii route
// through the theme (RadiusLG for the trigger, RadiusMD for the content,
// RadiusSM for items) so the user's --radius knob propagates.
//
// Source trigger:
//
//	"flex w-fit items-center justify-between gap-1.5 rounded-lg border
//	 border-input bg-transparent px-2.5 py-2 text-sm whitespace-nowrap
//	 shadow-xs ... data-[size=default]:h-8 data-[size=sm]:h-8 ...
//	 dark:bg-input/30 dark:hover:bg-input/50 ... [&_svg]:size-4
//	 [&_svg]:text-muted-foreground" + chevron "size-4 opacity-50"
//
// Source content:
//
//	"... min-w-[8rem] ... rounded-md border bg-popover
//	 text-popover-foreground shadow-md ..." + viewport "p-1"
//
// Source item:
//
//	"relative flex w-full cursor-default items-center gap-2 rounded-sm
//	 py-1.5 pr-8 pl-2 text-sm ... focus:bg-accent
//	 focus:text-accent-foreground data-[disabled]:opacity-50 ...
//	 [&_svg]:size-4"; check indicator "absolute right-2 ... size-3.5"
//	 with CheckIcon size-4.
var Select = struct {
	// TriggerHeight is the default trigger height (data-[size=default]:h-8).
	TriggerHeight float32
	// TriggerHeightSm is the small trigger height (data-[size=sm]:h-8).
	TriggerHeightSm float32
	// TriggerPadX is the trigger horizontal padding (px-2.5).
	TriggerPadX float32
	// TriggerGap is the gap between trigger text and chevron (gap-1.5).
	TriggerGap float32
	// BorderWidth is the trigger/content border width (border = 1px).
	BorderWidth float32
	// FontSize is the trigger and item font size (text-sm).
	FontSize float32
	// ChevronSize is the trigger chevron icon size (size-4).
	ChevronSize float32
	// ChevronOpacity is the trigger chevron opacity (chevron "opacity-50").
	ChevronOpacity float32

	// ContentMinWidth is the menu minimum width (min-w-[8rem] = 128px).
	ContentMinWidth float32
	// ContentPad is the menu viewport inner padding (viewport "p-1").
	ContentPad float32
	// SideOffset is the popper side offset toward the trigger (4px).
	SideOffset float32

	// ItemHeight is the rendered item row height. shadcn items are
	// py-1.5 (6px top+bottom = 12px) around a text-sm line box
	// (leading 20px) ⇒ 12 + 20 = 32px.
	ItemHeight float32
	// ItemPadY is the item vertical padding (py-1.5 = 6px).
	ItemPadY float32
	// ItemPadLeft is the item left padding (pl-2 = 8px).
	ItemPadLeft float32
	// ItemPadRight is the item right padding (pr-8 = 32px), reserving
	// room for the check indicator.
	ItemPadRight float32
	// ItemGap is the gap inside an item (gap-2 = 8px).
	ItemGap float32
	// CheckSize is the selected-item check icon size (CheckIcon size-4).
	CheckSize float32
	// CheckRight is the check indicator distance from the item right
	// edge (indicator "right-2" = 8px).
	CheckRight float32

	// FontWeight is the trigger/item text weight (text-sm, no font-medium
	// ⇒ 400).
	FontWeight int
}{
	TriggerHeight:   32, // h-8
	TriggerHeightSm: 32, // h-8
	TriggerPadX:     10, // px-2.5
	TriggerGap:      6,  // gap-1.5
	BorderWidth:     1,  // border
	FontSize:        14, // text-sm
	ChevronSize:     16, // size-4
	ChevronOpacity:  0.5,

	ContentMinWidth: 128, // min-w-[8rem]
	ContentPad:      4,   // p-1
	SideOffset:      4,

	ItemHeight:   32, // py-1.5 (12) + leading-5 (20)
	ItemPadY:     6,  // py-1.5
	ItemPadLeft:  8,  // pl-2
	ItemPadRight: 32, // pr-8
	ItemGap:      8,  // gap-2
	CheckSize:    16, // size-4
	CheckRight:   8,  // right-2

	FontWeight: 400,
}

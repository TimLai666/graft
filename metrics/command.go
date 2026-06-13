package metrics

// Command holds the pixel metrics for the shadcn Command palette
// (cmdk-style) — a dialog overlay with a search input at the top and a
// grouped, filtered command list below.
//
// Source markup (shadcn/ui new-york-v4):
//
//	<Command className="rounded-lg border shadow-md">
//	  <CommandInput placeholder="Type a command or search..." />
//	  <CommandList>
//	    <CommandEmpty>No results found.</CommandEmpty>
//	    <CommandGroup heading="Suggestions">
//	      <CommandItem><CalendarIcon /> Calendar <CommandShortcut>Ctrl K</CommandShortcut></CommandItem>
//	    </CommandGroup>
//	    <CommandSeparator />
//	  </CommandList>
//	</Command>
//
// CommandInput: "flex h-12 items-center gap-2 border-b px-3" → h-12 = 48px,
// px-3 = 12px, gap-2 = 8px; search icon 16px @50%.
//
// CommandItem: "px-2 py-1.5 rounded-sm text-sm" → height ~36px (py-1.5=12px
// + line ~24px), px-2=8px, rounded-sm=6px, text-sm=14px.
//
// CommandGroup heading: "text-xs font-medium text-muted-foreground px-2 py-1.5"
// → 12px, px-2=8px, py-1.5=6px top/bottom.
//
// CommandShortcut: "text-xs text-muted-foreground" → 12px.
//
// CommandSeparator: "h-px bg-border" → 1px.
//
// CommandDialog wraps Command in a Dialog: max-w ~460px, max-h ~320px for the
// list area.
var Command = struct {
	// InputHeight is the search-input wrapper height (h-12 = 48px).
	InputHeight float32

	// InputFontSize is the search field font size (text-sm = 14px).
	InputFontSize float32

	// InputPadX is the search-input wrapper horizontal padding (px-3 = 12px).
	InputPadX float32

	// InputGap is the gap between the search icon and the text field (gap-2 = 8px).
	InputGap float32

	// SearchIconSize is the leading search icon size (size-4 = 16px).
	SearchIconSize float32

	// SearchIconOpacity is the search icon opacity (opacity-50).
	SearchIconOpacity float32

	// SeparatorWidth is the input-area bottom border and command separators (1px).
	SeparatorWidth float32

	// ItemHeight is the command item row height (py-1.5 + ~24px line = 36px).
	ItemHeight float32

	// ItemPadX is the item horizontal padding (px-2 = 8px).
	ItemPadX float32

	// ItemRadius is the item hover/selection rounded-sm radius (6px).
	ItemRadius float32

	// ItemFontSize is the item label font size (text-sm = 14px).
	ItemFontSize float32

	// ItemGap is the gap between the icon and the label (gap-2 = 8px).
	ItemGap float32

	// IconSize is the leading icon size in a command item (size-4 = 16px).
	IconSize float32

	// IconOpacity is the leading icon idle opacity (muted-foreground style).
	IconOpacity float32

	// ShortcutSize is the right-aligned shortcut text size (text-xs = 12px).
	ShortcutSize float32

	// GroupLabelSize is the group heading font size (text-xs = 12px).
	GroupLabelSize float32

	// GroupLabelPadX is the group heading horizontal padding (px-2 = 8px).
	GroupLabelPadX float32

	// GroupLabelPadY is the group heading vertical padding (py-1.5 = 6px each).
	GroupLabelPadY float32

	// GroupLabelHeight is the total group heading row height (6+12+6 = 24px).
	GroupLabelHeight float32

	// ListPad is the group inner padding (p-1 = 4px).
	ListPad float32

	// SepHeight is the separator row height (h-px = 1px + my-1 = 4+1+4 = 9px).
	SepHeight float32

	// SepMarginY is the vertical margin around the separator (my-1 = 4px).
	SepMarginY float32

	// DialogWidth is the command dialog content width (~460px).
	DialogWidth float32

	// DialogMaxHeight is the command list max-height (max-h-[300px] + padding).
	DialogMaxHeight float32

	// EmptyPadY is the empty-state vertical padding (py-6 = 24px).
	EmptyPadY float32

	// EmptyFontSize is the empty-state font size (text-sm = 14px).
	EmptyFontSize float32

	// Placeholder is the default search-input placeholder.
	Placeholder string

	// EmptyText is the no-results label.
	EmptyText string
}{
	InputHeight:       48,
	InputFontSize:     14,
	InputPadX:         12,
	InputGap:          8,
	SearchIconSize:    16,
	SearchIconOpacity: 0.50,
	SeparatorWidth:    1,

	ItemHeight:   36,
	ItemPadX:     8,
	ItemRadius:   6,
	ItemFontSize: 14,
	ItemGap:      8,

	IconSize:    16,
	IconOpacity: 0.50,

	ShortcutSize: 12,

	GroupLabelSize:   12,
	GroupLabelPadX:   8,
	GroupLabelPadY:   6,
	GroupLabelHeight: 24,

	ListPad: 4,

	SepHeight:  1,
	SepMarginY: 4,

	DialogWidth:     460,
	DialogMaxHeight: 320,

	EmptyPadY:     24,
	EmptyFontSize: 14,

	Placeholder: "Type a command or search...",
	EmptyText:   "No results found.",
}

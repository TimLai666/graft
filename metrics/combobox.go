package metrics

// Combobox holds the pixel metrics for the shadcn Combobox composition
// (docs/research/03-shadcn-pixel-spec.md §5 "Command" + Popover + Button).
// Source markup:
//
//	<Button variant="outline" role="combobox"
//	        className="w-[200px] justify-between">
//	  {value ? label : placeholder}
//	  <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
//	</Button>
//	<PopoverContent className="w-[200px] p-0">
//	  <Command>
//	    <CommandInput placeholder="Search ..." />     // wrapper h-9, border-b
//	    <CommandEmpty>No results found.</CommandEmpty> // py-6 text-center
//	    <CommandGroup>
//	      <CommandItem ...>                            // px-2 py-1.5 rounded-sm
//	        <Check className="mr-2 h-4 w-4 opacity-..." />
//	        {label}
//	      </CommandItem>
//	    </CommandGroup>
//	  </Command>
//	</PopoverContent>
//
// Command input wrapper: "flex h-9 items-center gap-2 border-b px-3" with a
// 16px search icon @50% opacity; item: "px-2 py-1.5 rounded-sm text-sm",
// selected = bg-accent, leading 16px check; empty: "py-6 text-center
// text-sm".
var Combobox = struct {
	// TriggerWidth is the default trigger/content width (w-[200px] = 200px).
	TriggerWidth float32
	// ChevronSize is the trailing chevrons-up-down size (h-4 w-4 = 16px).
	ChevronSize float32
	// ChevronOpacity is the trailing chevron opacity (opacity-50).
	ChevronOpacity float32

	// ListPad is the list/group inner padding (Command group p-1 = 4px).
	ListPad float32

	// InputHeight is the search-input wrapper height (h-9 = 36px).
	InputHeight float32
	// InputPadX is the search-input wrapper horizontal padding (px-3 = 12px).
	InputPadX float32
	// InputGap is the gap between the search icon and the field (gap-2 = 8px).
	InputGap float32
	// SearchIconSize is the leading search icon size (size-4 = 16px).
	SearchIconSize float32
	// SearchIconOpacity is the search icon opacity (opacity-50).
	SearchIconOpacity float32
	// InputFontSize is the search field font size (text-sm = 14px).
	InputFontSize float32
	// SeparatorWidth is the input wrapper bottom border (border-b = 1px).
	SeparatorWidth float32

	// ItemHeight is the item row height: py-1.5 (12px) + leading-5 (20px)
	// = 32px.
	ItemHeight float32
	// ItemPadX is the item horizontal padding (px-2 = 8px).
	ItemPadX float32
	// ItemGap is the gap between the check and the label (mr-2 = 8px).
	ItemGap float32
	// CheckSize is the leading check icon size (h-4 w-4 = 16px).
	CheckSize float32
	// ItemFontSize is the item font size (text-sm = 14px).
	ItemFontSize float32
	// ItemRadius routes through Theme.RadiusSM (rounded-sm = 6px).

	// EmptyPadY is the empty-state vertical padding (py-6 = 24px).
	EmptyPadY float32
	// EmptyFontSize is the empty-state font size (text-sm = 14px).
	EmptyFontSize float32

	// MaxListHeight caps the scrollable list height (Command list
	// max-h-[300px]).
	MaxListHeight float32

	// Placeholder is the default trigger placeholder.
	Placeholder string
	// SearchPlaceholder is the default search-input placeholder.
	SearchPlaceholder string
	// EmptyText is the no-results label.
	EmptyText string
}{
	TriggerWidth:   200,
	ChevronSize:    16,
	ChevronOpacity: 0.50,

	ListPad: 4, // p-1

	InputHeight:       36, // h-9
	InputPadX:         12, // px-3
	InputGap:          8,  // gap-2
	SearchIconSize:    16, // size-4
	SearchIconOpacity: 0.50,
	InputFontSize:     14, // text-sm
	SeparatorWidth:    1,  // border-b

	ItemHeight:   32, // py-1.5 + leading-5
	ItemPadX:     8,  // px-2
	ItemGap:      8,  // mr-2
	CheckSize:    16, // size-4
	ItemFontSize: 14, // text-sm

	EmptyPadY:     24, // py-6
	EmptyFontSize: 14, // text-sm

	MaxListHeight: 300,

	Placeholder:       "Select...",
	SearchPlaceholder: "Search...",
	EmptyText:         "No results found.",
}

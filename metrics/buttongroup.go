package metrics

// ButtonGroup holds the pixel constants for the shadcn ButtonGroup composite
// (docs/research/03-shadcn-pixel-spec.md §5 "Toggle Group" fused-segment rules,
// which ButtonGroup reuses):
//
//	ToggleGroup root "flex w-fit items-center rounded-md ... shadow-xs"; items at
//	spacing 0 "rounded-none first:rounded-l-md last:rounded-r-md border-l-0
//	first:border-l" → a fused segmented control: shared 1px borders by
//	overlapping −1px, only the first/last outer corners rounded (rounded-md =
//	8px via the theme), divider lines in the Border token.
//
// Radius routes through the theme; the group's outer corners follow the child
// buttons, now rounded-lg (→ t.RadiusLG()). Only fixed literals live here.
var ButtonGroup = struct {
	// Overlap is the negative horizontal gap in px applied between adjacent
	// buttons so their 1px borders coincide into one shared divider.
	Overlap float32

	// DividerWidth is the width in px of the divider drawn between segments
	// (border-l → 1).
	DividerWidth float32

	// OuterBorderWidth is the width in px of the unified outer border drawn
	// around the whole group (border → 1).
	OuterBorderWidth float32
}{
	Overlap:          1,
	DividerWidth:     1,
	OuterBorderWidth: 1,
}

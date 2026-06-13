package metrics

// ToggleGroup metrics transcribed from the shadcn new-york-v4 ToggleGroup
// (docs/research/03-shadcn-pixel-spec.md "Toggle + Toggle Group", quoted
// verbatim there).
//
// Root: "group/toggle-group flex w-fit items-center
//
//	gap-[--spacing(var(--gap))] rounded-md
//	data-[spacing=default]:data-[variant=outline]:shadow-xs".
//
// Items add: "w-auto min-w-0 shrink-0 px-3 focus:z-10 focus-visible:z-10" and
//
//	at spacing 0:
//	"data-[spacing=0]:rounded-none data-[spacing=0]:shadow-none
//	 data-[spacing=0]:first:rounded-l-md data-[spacing=0]:last:rounded-r-md
//	 data-[spacing=0]:data-[variant=outline]:border-l-0
//	 data-[spacing=0]:data-[variant=outline]:first:border-l".
//
// graft renders the spacing-0 fused segmented control: items butt together
// (no gap), share 1px borders (each item suppresses its left border except
// the first, so adjacent borders don't double), and only the group's outer
// corners are rounded (first item rounds left, last rounds right) via
// draw.SquareCorners. The outline variant carries shadow-xs on the whole
// group. Item height/colors inherit the Toggle metrics (h-9, etc.); the only
// override is the px-3 horizontal padding.
var ToggleGroup = struct {
	// ItemPadX is the item horizontal padding in px (px-3 = 12), overriding
	// the standalone Toggle padding.
	ItemPadX float32

	// BorderWidth is the outline-variant shared border width in px (1px),
	// in the --input token.
	BorderWidth float32

	// Overlap is how far each item after the first shifts left so its left
	// edge sits on top of the previous item's right border, collapsing the
	// two adjacent 1px borders into one (border-l-0 on non-first items).
	Overlap float32
}{
	ItemPadX:    12, // px-3
	BorderWidth: 1,  // border (outline)
	Overlap:     1,  // adjacent borders collapse to 1px
}

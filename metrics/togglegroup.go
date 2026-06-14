package metrics

// ToggleGroup metrics transcribed from the current shadcn (radix-era)
// ToggleGroup (compare/shadcn-spec.md "toggle-group-item", measured 2026-06-14
// from ui.shadcn.com).
//
// Root: "group/toggle-group flex w-fit items-center
//
//	gap-[--spacing(var(--gap))] rounded-lg
//	data-[spacing=default]:data-[variant=outline]:shadow-xs".
//
// Items: "h-8 min-w-8 px-2.5 rounded-lg w-auto min-w-0 shrink-0 focus:z-10
//	focus-visible:z-10" and
//
//	at spacing 0:
//	"data-[spacing=0]:rounded-none data-[spacing=0]:shadow-none
//	 data-[spacing=0]:first:rounded-l-lg data-[spacing=0]:last:rounded-r-lg
//	 data-[spacing=0]:data-[variant=outline]:border-l-0
//	 data-[spacing=0]:data-[variant=outline]:first:border-l".
//
// graft renders the spacing-0 fused segmented control: items butt together
// (no gap), share 1px borders (each item suppresses its left border except
// the first, so adjacent borders don't double), and only the group's outer
// corners are rounded (first item rounds left, last rounds right) via
// draw.SquareCorners. The outline variant carries shadow-xs on the whole
// group. Items carry their own h-8 height (32) and px-2.5 padding (10), and
// round at rounded-lg (RadiusLG = 10px) — distinct from the standalone Toggle.
var ToggleGroup = struct {
	// ItemHeight is the fixed item height in px (h-8 = 32), distinct from the
	// standalone Toggle's default height.
	ItemHeight float32

	// ItemPadX is the item horizontal padding in px (px-2.5 = 10).
	ItemPadX float32

	// BorderWidth is the outline-variant shared border width in px (1px),
	// in the --input token.
	BorderWidth float32

	// Overlap is how far each item after the first shifts left so its left
	// edge sits on top of the previous item's right border, collapsing the
	// two adjacent 1px borders into one (border-l-0 on non-first items).
	Overlap float32
}{
	ItemHeight:  32, // h-8
	ItemPadX:    10, // px-2.5
	BorderWidth: 1,  // border (outline)
	Overlap:     1,  // adjacent borders collapse to 1px
}

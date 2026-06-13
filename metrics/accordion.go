package metrics

// Accordion holds the exact px metrics for the shadcn Accordion component
// (docs/research/03-shadcn-pixel-spec.md "Accordion", quoted verbatim there).
//
// Item:    "border-b last:border-b-0".
// Trigger: "flex flex-1 items-start justify-between gap-4 rounded-md py-4
//
//	text-left text-sm font-medium transition-all outline-none
//	hover:underline focus-visible:border-ring focus-visible:ring-[3px]
//	focus-visible:ring-ring/50 disabled:pointer-events-none
//	disabled:opacity-50 [&[data-state=open]>svg]:rotate-180" plus a
//	chevron "pointer-events-none size-4 shrink-0 translate-y-0.5
//	text-muted-foreground transition-transform duration-200".
//
// Content: "overflow-hidden text-sm" with inner "pt-0 pb-4".
//
// The chevron rotates 180° when open. graft has no rotation transform
// (Report 1 §7), so the chevron-down (closed) and chevron-up (open) icons are
// swapped instead — visually identical to the rotated glyph.
//
// The content collapse is the accordion-down/up keyframe (height 0 ↔ natural,
// 0.2s ease-out). Goldens render the settled state.
var Accordion = struct {
	// TriggerPadY is the trigger vertical padding in px (py-4 = 16).
	TriggerPadY float32

	// TriggerGap is the gap between the trigger label and the chevron in px
	// (gap-4 = 16).
	TriggerGap float32

	// TriggerFontSize is the trigger label size in px (text-sm = 14).
	TriggerFontSize float32

	// TriggerLineHeight is the trigger label line box in px (text-sm = 20).
	TriggerLineHeight float32

	// TriggerFontWeight is the trigger label weight (font-medium = 500).
	TriggerFontWeight int

	// ChevronSize is the chevron icon box in px (size-4 = 16).
	ChevronSize float32

	// ChevronDropY shifts the chevron down by 2px (translate-y-0.5) so it
	// aligns with the first text line (items-start).
	ChevronDropY float32

	// ItemBorderWidth is the item bottom-border width in px (border-b = 1px),
	// in the --border token; the last item has none (last:border-b-0).
	ItemBorderWidth float32

	// ContentPadBottom is the content inner bottom padding in px (pb-4 = 16);
	// the top padding is zero (pt-0).
	ContentPadBottom float32

	// ContentFontSize is the content text size in px (text-sm = 14).
	ContentFontSize float32

	// UnderlineWidth is the hover underline thickness in px (CSS default ≈ 1).
	UnderlineWidth float32

	// UnderlineOffset is the hover underline distance below the text baseline
	// in px (text-decoration default underline-offset ≈ 2-3px; shadcn uses
	// the Tailwind default, no underline-offset utility on the trigger).
	UnderlineOffset float32

	// DurationMS is the content height + chevron animation duration in ms
	// (accordion-down/up 0.2s ease-out; chevron transition-transform 200ms).
	DurationMS float32
}{
	TriggerPadY:       16,  // py-4
	TriggerGap:        16,  // gap-4
	TriggerFontSize:   14,  // text-sm
	TriggerLineHeight: 20,  // text-sm line height
	TriggerFontWeight: 500, // font-medium
	ChevronSize:       16,  // size-4
	ChevronDropY:      2,   // translate-y-0.5
	ItemBorderWidth:   1,   // border-b
	ContentPadBottom:  16,  // pb-4
	ContentFontSize:   14,  // text-sm
	UnderlineWidth:    1,
	UnderlineOffset:   2,
	DurationMS:        200, // 0.2s ease-out
}

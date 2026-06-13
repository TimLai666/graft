package metrics

// Label metrics transcribed from the shadcn new-york-v4 label registry
// entry (docs/research/03-shadcn-pixel-spec.md §5 "Label"):
//
//	"flex items-center gap-2 text-sm leading-none font-medium select-none
//	 group-data-[disabled=true]:pointer-events-none
//	 group-data-[disabled=true]:opacity-50
//	 peer-disabled:cursor-not-allowed peer-disabled:opacity-50"
var Label = struct {
	// FontSize is the label text size (text-sm = 14px). leading-none means
	// the line box equals the font size.
	FontSize float32

	// FontWeight is the label weight (font-medium = 500).
	FontWeight int

	// Gap is the spacing between the label text and a leading icon or
	// adjacent control (gap-2 = 8px).
	Gap float32
}{
	FontSize:   14,  // text-sm
	FontWeight: 500, // font-medium
	Gap:        8,   // gap-2
}

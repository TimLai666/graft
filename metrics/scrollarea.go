package metrics

// ScrollArea metrics, transcribed from the shadcn new-york-v4 scroll-area
// registry entry (docs/research/03-shadcn-pixel-spec.md, section 5
// "Scroll Area").
//
// Scrollbar: "flex touch-none p-px transition-colors select-none" +
// vertical "h-full w-2.5 border-l border-l-transparent" / horizontal
// "h-2.5 flex-col border-t border-t-transparent".
// Thumb: "relative flex-1 rounded-full bg-border".
var ScrollArea = struct {
	// GutterWidth is the scrollbar gutter thickness: "w-2.5" (vertical) /
	// "h-2.5" (horizontal) = 10px.
	GutterWidth float32

	// GutterPadding is the padding between gutter edge and thumb:
	// "p-px" = 1px.
	GutterPadding float32

	// ThumbWidth is the resulting thumb thickness:
	// GutterWidth - 2*GutterPadding = 8px.
	ThumbWidth float32

	// ThumbRadius is the thumb corner radius: "rounded-full" = 9999
	// (pill; the renderer clamps to half the thumb thickness).
	ThumbRadius float32
}{
	GutterWidth:   10,
	GutterPadding: 1,
	ThumbWidth:    8,
	ThumbRadius:   9999,
}

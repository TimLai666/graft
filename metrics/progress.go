package metrics

// Progress holds the exact pixel constants for the shadcn Progress bar
// (docs/research/03-shadcn-pixel-spec.md §5 "Progress").
//
// Source (new-york-v4):
//
//	Root:      "relative h-2 w-full overflow-hidden rounded-full bg-primary/20"
//	Indicator: "h-full w-full flex-1 bg-primary transition-all"
//	           transform: translateX(-(100 - value)%)
//
// The track and indicator both use the Primary token (track at 20% alpha),
// the radius is rounded-full, and the bar fills left→right by value. All of
// those route through the theme; only the literals live here.
var Progress = struct {
	// Height is the bar height in px (h-2 = 8px).
	Height float32

	// TrackAlpha is the alpha applied to the Primary token for the track
	// (bg-primary/20 = 20%).
	TrackAlpha float32

	// Min and Max bound the value range (shadcn uses 0..100).
	Min, Max float64
}{
	Height:     8,
	TrackAlpha: 0.2,
	Min:        0,
	Max:        100,
}

package metrics

// Slider holds the exact px metrics for the shadcn Slider component
// (docs/research/03-shadcn-pixel-spec.md section 5 "Slider").
//
// Track: "relative grow overflow-hidden rounded-full bg-muted
// data-[orientation=horizontal]:h-1 ..." — 4px thick pill in --muted.
// Range: "absolute bg-primary data-[orientation=horizontal]:h-full".
// Thumb: "block size-3 shrink-0 rounded-full border border-primary
// bg-white shadow-sm ring-ring/50 ... hover:ring-4 focus-visible:ring-4
// ... disabled:opacity-50" — 12px circle, always-white fill, 1px
// --primary border, shadow-sm, 4px ring (--ring at 50%) on hover AND
// keyboard focus. Root: "disabled:opacity-50" — whole control at 50%.
//
// The 4px ring width itself is the shared SliderRingWidth constant in
// focus.go; the ring alpha is the shared RingAlpha (ring-ring/50).
var Slider = struct {
	// TrackThickness is the track cross-axis size in px (h-1 = 4px).
	TrackThickness float32

	// TrackRadius is the track corner radius (rounded-full on a 4px
	// track clamps to half the thickness).
	TrackRadius float32

	// ThumbSize is the thumb diameter in px (size-3 = 12px).
	ThumbSize float32

	// ThumbBorderWidth is the thumb border width in px (border = 1px,
	// in --primary).
	ThumbBorderWidth float32

	// Height is the slider root's layout height in px. shadcn's root has
	// no explicit height; the 12px thumb defines it. The hover/focus ring
	// and shadow paint outside these bounds, exactly like the CSS
	// box-shadow ring.
	Height float32

	// DefaultWidth is the fallback width in px when the slider is laid
	// out without a bounded width and without an explicit .W. shadcn's
	// slider is w-full; inside bounded containers graft stretches the
	// same way, so this only applies in unbounded contexts.
	DefaultWidth float32
}{
	TrackThickness:   4,  // data-[orientation=horizontal]:h-1
	TrackRadius:      2,  // rounded-full, clamped to TrackThickness/2
	ThumbSize:        12, // size-3
	ThumbBorderWidth: 1,  // border border-primary
	Height:           12, // root height = thumb diameter
	DefaultWidth:     200,
}

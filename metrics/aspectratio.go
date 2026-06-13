package metrics

// AspectRatio metrics.
//
// Source: shadcn new-york-v4 aspect-ratio.tsx is a thin wrapper over Radix
// AspectRatio.Root with no styling of its own — it simply sizes a slot to a
// width:height ratio (via the CSS aspect-ratio property) within the available
// width. There are no Tailwind classes to transcribe; the only "metric" is
// the default ratio shadcn's docs demo with (16/9).
const (
	// AspectRatioDefault is the fallback ratio (width/height) used when a
	// non-positive ratio is supplied. shadcn's demo uses 16/9.
	AspectRatioDefault float32 = 16.0 / 9.0
)

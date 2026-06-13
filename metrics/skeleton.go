package metrics

import "time"

// Skeleton holds the exact constants for the shadcn Skeleton placeholder
// (docs/research/03-shadcn-pixel-spec.md §5 "Skeleton").
//
// Source (new-york-v4):
//
//	"animate-pulse rounded-md bg-accent"
//
// with the Tailwind pulse keyframes:
//
//	@keyframes pulse { 50% { opacity: .5 } }
//	animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite
//
// The fill token (Accent) and radius (rounded-md → t.RadiusMD()) route
// through the theme; only the animation timing and the trough opacity are
// literals here.
var Skeleton = struct {
	// PulsePeriod is the full pulse cycle duration (animate-pulse = 2s).
	PulsePeriod time.Duration

	// PulseMinOpacity is the opacity at the 50% keyframe trough
	// (@keyframes pulse { 50% { opacity: .5 } }).
	PulseMinOpacity float32
}{
	PulsePeriod:     2 * time.Second,
	PulseMinOpacity: 0.5,
}

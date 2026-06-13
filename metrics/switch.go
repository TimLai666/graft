package metrics

import "time"

// SwitchSize holds the geometry for one Switch size variant.
type SwitchSize struct {
	// TrackWidth is the track width in px.
	TrackWidth float32

	// TrackHeight is the track height in px.
	TrackHeight float32

	// ThumbSize is the thumb circle diameter in px.
	ThumbSize float32

	// Travel is the thumb horizontal translation when checked, in px
	// (translateX = thumb width − 2px).
	Travel float32
}

// Switch holds the exact pixel constants for the shadcn Switch control
// (docs/research/03-shadcn-pixel-spec.md §5 "Switch"), transcribed from the
// new-york-v4 registry:
//
//	track: "... rounded-full border border-transparent shadow-xs ...
//	        data-[size=default]:h-[1.15rem] data-[size=default]:w-8
//	        data-[size=sm]:h-3.5 data-[size=sm]:w-6
//	        data-[state=checked]:bg-primary data-[state=unchecked]:bg-input
//	        dark:data-[state=unchecked]:bg-input/80"
//	thumb: "... rounded-full bg-background ... size-4 / size-3
//	        data-[state=checked]:translate-x-[calc(100%-2px)]
//	        dark:data-[state=checked]:bg-primary-foreground
//	        dark:data-[state=unchecked]:bg-foreground"
//
// Track radius is rounded-full (pill); thumb is a full circle. The track
// border is 1px transparent (it reserves the 1px the thumb inset honors).
var Switch = struct {
	// Default is the default size: track 32×18.4 (h = 1.15rem = 18.4px),
	// thumb 16px, travel 14px.
	Default SwitchSize

	// SM is the small size: track 24×14, thumb 12px, travel 10px.
	SM SwitchSize

	// BorderWidth is the track border width in px (border → 1, transparent).
	BorderWidth float32

	// DarkUncheckedTrackAlpha is the dark unchecked track alpha applied to
	// the Input token (dark:data-[state=unchecked]:bg-input/80 → 0.80).
	DarkUncheckedTrackAlpha float32

	// AnimDuration is the thumb translation animation duration
	// (Tailwind transition-transform default 150ms).
	AnimDuration time.Duration
}{
	Default: SwitchSize{TrackWidth: 32, TrackHeight: 18.4, ThumbSize: 16, Travel: 14},
	SM:      SwitchSize{TrackWidth: 24, TrackHeight: 14, ThumbSize: 12, Travel: 10},

	BorderWidth:             1,
	DarkUncheckedTrackAlpha: 0.80,
	AnimDuration:            150 * time.Millisecond,
}

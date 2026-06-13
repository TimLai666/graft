package metrics

// Avatar holds the exact pixel constants for the shadcn Avatar
// (docs/research/03-shadcn-pixel-spec.md §5 "Avatar (sized)").
//
// Source (new-york-v4):
//
//	Root:     "relative flex size-8 shrink-0 overflow-hidden rounded-full
//	           data-[size=lg]:size-10 data-[size=sm]:size-6"
//	Image:    "aspect-square size-full"
//	Fallback: "bg-muted flex size-full items-center justify-center
//	           rounded-full text-sm" (12px text at sm)
//
// The fill (Muted) and text color (MutedForeground) route through the theme;
// only the sizes and font sizes live here. The radius is rounded-full.
var Avatar = struct {
	// SizeSM / SizeDefault / SizeLG are the avatar diameters in px
	// (size-6 = 24, size-8 = 32, size-10 = 40).
	SizeSM, SizeDefault, SizeLG float32

	// FallbackFontSize is the fallback initials font size in px at the
	// default/lg sizes (text-sm = 14px).
	FallbackFontSize float32

	// FallbackFontSizeSM is the fallback font size in px at the sm size
	// (12px, per Report 3 §5).
	FallbackFontSizeSM float32

	// FallbackWeight is the fallback initials font weight. shadcn uses the
	// inherited 400; graft keeps it at 400.
	FallbackWeight int
}{
	SizeSM:             24,
	SizeDefault:        32,
	SizeLG:             40,
	FallbackFontSize:   14,
	FallbackFontSizeSM: 12,
	FallbackWeight:     400,
}

package metrics

// Separator metrics.
//
// Source: shadcn new-york-v4 separator.tsx:
//
//	"shrink-0 bg-border data-[orientation=horizontal]:h-px
//	 data-[orientation=horizontal]:w-full data-[orientation=vertical]:h-full
//	 data-[orientation=vertical]:w-px"
const (
	// SeparatorThickness is the rule thickness in px (h-px / w-px = 1px).
	// The fill color is the theme Border token (bg-border).
	SeparatorThickness float32 = 1
)

package metrics

// Sonner holds the pixel metrics for the shadcn Sonner toaster
// (docs/research/03-shadcn-pixel-spec.md §5 "Sonner (Toaster)").
//
// The shadcn wrapper maps Sonner's CSS variables onto the design tokens:
//
//	--normal-bg: var(--popover)
//	--normal-text: var(--popover-foreground)
//	--normal-border: var(--border)
//	--border-radius: var(--radius)   (10px → routes through Theme.RadiusLG)
//
// Sonner's own defaults supply the rest: toast width 356px, padding 16px,
// shadow ≈ shadow-lg, status icons 16px lucide. Title is text-sm/500;
// description is the smaller muted line (13px). The toaster region anchors
// to a window corner (default bottom-right) with a 16px viewport offset and
// stacks toasts with a small gap, newest nearest the corner (LIFO).
var Sonner = struct {
	// Width is the toast card width (Sonner default width 356px).
	Width float32
	// Padding is the toast card inner padding (Sonner default 16px).
	Padding float32
	// BorderWidth is the toast card border (--normal-border = 1px).
	BorderWidth float32

	// Gap is the vertical gap between stacked toasts.
	Gap float32
	// ViewportOffset is the inset from the window edge to the toaster
	// region (Sonner default offset 16px → 24px? Sonner uses 32px on
	// desktop via --offset, but the shadcn demo sits 16px in; we use 16).
	ViewportOffset float32

	// IconSize is the status-icon size (16px lucide).
	IconSize float32
	// IconGap is the gap between the status icon and the text column.
	IconGap float32

	// TitleFontSize is the title size (text-sm = 14px).
	TitleFontSize float32
	// TitleFontWeight is the title weight (font-medium = 500).
	TitleFontWeight int
	// TitleLineHeight is the title line box (leading-5 = 20px).
	TitleLineHeight float32

	// DescriptionFontSize is the description size (13px).
	DescriptionFontSize float32
	// DescriptionLineHeight is the description line box.
	DescriptionLineHeight float32
	// TextGap is the gap between the title and description rows.
	TextGap float32

	// ActionGap is the gap above an action button row.
	ActionGap float32

	// AutoDismiss is the default auto-dismiss duration, in milliseconds
	// (Sonner default 4000ms).
	AutoDismiss int
}{
	Width:       356, // Sonner width
	Padding:     16,  // Sonner padding
	BorderWidth: 1,   // --normal-border

	Gap:            14, // Sonner gap between toasts
	ViewportOffset: 16,

	IconSize: 16, // size-4
	IconGap:  8,  // gap-2

	TitleFontSize:   14,  // text-sm
	TitleFontWeight: 500, // font-medium
	TitleLineHeight: 20,  // leading-5

	DescriptionFontSize:   13,
	DescriptionLineHeight: 18,
	TextGap:               4,

	ActionGap: 12,

	AutoDismiss: 4000,
}

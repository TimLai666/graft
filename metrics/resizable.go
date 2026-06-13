package metrics

// Resizable holds the exact px metrics for the shadcn Resizable component
// (react-resizable-panels wrapper; no entry in
// docs/research/03-shadcn-pixel-spec.md, transcribed from the registry
// resizable.tsx Tailwind classes).
//
// ResizablePanelGroup: "flex h-full w-full
// aria-[orientation=vertical]:flex-col" — a flex row (horizontal) or column
// (vertical) of panels separated by handles.
//
// ResizablePanel: no class (the panel is just its content, sized by ratio).
//
// ResizableHandle: "relative flex w-px items-center justify-center bg-border
// after:absolute after:inset-y-0 after:left-1/2 after:w-1
// after:-translate-x-1/2 ... aria-[orientation=horizontal]:h-px
// aria-[orientation=horizontal]:w-full ... after:h-1 ..." — a 1px line in
// --border with a 4px-wide invisible hit/drag area centered on it
// (after:w-1 horizontal split; after:h-1 vertical split).
//
// Grip handle (withHandle): "z-10 flex h-4 w-3 items-center justify-center
// rounded-xs border bg-border" containing a "size-2.5" GripVertical icon —
// a 16×12 chip in --border with a 1px --border outline and a 10px grip
// glyph. Radius routes through the theme (rounded-xs = RadiusXS).
var Resizable = struct {
	// LineWidth is the divider line thickness in px (w-px / h-px = 1px,
	// drawn in --border).
	LineWidth float32

	// HitWidth is the invisible drag/hit band thickness in px centered on
	// the line (after:w-1 / after:h-1 = 4px). It is the draggable target,
	// larger than the visible 1px line.
	HitWidth float32

	// GripLongSide is the grip chip's size along the divider axis in px
	// (h-4 = 16 for a vertical divider, the long edge of the 16×12 chip).
	GripLongSide float32

	// GripShortSide is the grip chip's size across the divider axis in px
	// (w-3 = 12 for a vertical divider).
	GripShortSide float32

	// GripIconSize is the grip glyph size in px (size-2.5 = 10px).
	GripIconSize float32

	// GripBorderWidth is the grip chip outline width in px (border = 1px).
	GripBorderWidth float32

	// DefaultWidth and DefaultHeight are fallback group dimensions in px
	// when the group is laid out without a bounded axis (h-full/w-full have
	// no intrinsic size).
	DefaultWidth  float32
	DefaultHeight float32
}{
	LineWidth:       1,   // w-px / h-px
	HitWidth:        4,   // after:w-1 / after:h-1
	GripLongSide:    16,  // h-4
	GripShortSide:   12,  // w-3
	GripIconSize:    10,  // size-2.5
	GripBorderWidth: 1,   // border
	DefaultWidth:    400, // fallback flex row width
	DefaultHeight:   200, // fallback flex column height
}

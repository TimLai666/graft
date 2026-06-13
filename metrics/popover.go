package metrics

// Popover metrics, transcribed from the shadcn new-york-v4 popover
// registry entry (docs/research/03-shadcn-pixel-spec.md, section 5
// "Popover").
//
// Content: "z-50 w-72 rounded-md border bg-popover p-4
// text-popover-foreground shadow-md outline-hidden ...", sideOffset 4.
//
// Radius: "rounded-md" routes through Theme.RadiusMD (8px @ default
// radius). Shadow: "shadow-md" = ShadowMD.
var Popover = struct {
	// Width is the content width: "w-72" = 288px.
	Width float32

	// Padding is the content padding: "p-4" = 16px.
	Padding float32

	// SideOffset is the gap between trigger and content: sideOffset = 4.
	SideOffset float32

	// BorderWidth is the content border: "border" = 1px.
	BorderWidth float32
}{
	Width:       288,
	Padding:     16,
	SideOffset:  4,
	BorderWidth: 1,
}

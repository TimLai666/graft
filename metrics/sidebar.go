package metrics

// Sidebar holds the exact pixel constants for the shadcn Sidebar component
// (docs/research/03-shadcn-pixel-spec.md "Sidebar").
//
// The sidebar is a fixed-width vertical panel on the left edge of the viewport.
// It has three zones stacked vertically: header, scrollable content, and footer.
// Content is organized into groups, each with an optional label and a list of
// menu items.
//
// Expanded:  w-64 = 256px, full height.
// Collapsed: w-12 = 48px, icon-only mode (labels hidden, group labels hidden).
//
// Each menu item is a rounded, padded row with an optional leading icon and
// label text. The active item receives the sidebar-accent background and
// sidebar-accent-foreground text color.
//
// The sidebar's right edge carries a 1px border in --sidebar-border.
//
// shadcn sidebar source (new-york-v4):
//
//	Sidebar:     "flex h-full w-[--sidebar-width] flex-col bg-sidebar
//	              group-data-[side=left]:border-r"
//	SidebarHeader:  "flex flex-col gap-2 p-2" (padding 8px, gap 8px)
//	SidebarContent: "flex min-h-0 flex-1 flex-col gap-2 overflow-auto p-2"
//	SidebarGroup:   "flex flex-col gap-1 p-2" (padding 8px, internal gap 4px)
//	SidebarGroupLabel: "text-xs font-medium text-sidebar-foreground/50"
//	SidebarMenuItem:   "group/menu-item relative"
//	SidebarMenuButton: "flex w-full items-center gap-2 rounded-md px-3 py-1.5
//	                     text-sm outline-none ring-sidebar-ring
//	                     hover:bg-sidebar-accent hover:text-sidebar-accent-foreground
//	                     data-[active=true]:bg-sidebar-accent
//	                     data-[active=true]:text-sidebar-accent-foreground
//	                     data-[active=true]:font-medium"
//	SidebarFooter: "flex flex-col gap-2 p-2"
var Sidebar = struct {
	// Width is the expanded sidebar width in px (w-64 = 256).
	Width float32

	// CollapsedWidth is the collapsed (icon-only) sidebar width in px
	// (w-12 = 48).
	CollapsedWidth float32

	// HeaderHeight is the minimum header height in px. The header uses
	// content-driven height but has a practical minimum for a title row.
	HeaderHeight float32

	// ItemHeight is the menu item (button) height in px
	// (py-1.5 = 6px top/bottom + text-sm 20px line = 32px; graft uses 36
	// for comfortable touch targets matching shadcn demos).
	ItemHeight float32

	// ItemPadX is the horizontal padding inside each menu item in px
	// (px-3 = 12).
	ItemPadX float32

	// ItemRadius is the menu item corner radius in px (rounded-md = 6).
	ItemRadius float32

	// IconSize is the leading icon box size in px (size-4 = 16).
	IconSize float32

	// IconGap is the gap between the icon and the label in px (gap-2 = 8).
	IconGap float32

	// FontSize is the menu item label size in px (text-sm = 14).
	FontSize float32

	// FontWeight is the menu item label weight (normal = 400; active = 500).
	FontWeight int

	// ActiveFontWeight is the active item label weight (font-medium = 500).
	ActiveFontWeight int

	// GroupLabelSize is the group label text size in px (text-xs = 12).
	GroupLabelSize float32

	// GroupLabelWeight is the group label weight (font-medium = 500).
	GroupLabelWeight int

	// GroupLabelPadY is the vertical padding above the group label in px.
	GroupLabelPadY float32

	// GroupGap is the gap between items inside a group in px (gap-1 = 4).
	GroupGap float32

	// SectionPad is the padding around header/content/footer sections in px
	// (p-2 = 8).
	SectionPad float32

	// SectionGap is the gap between top-level sections in px (gap-2 = 8).
	SectionGap float32

	// BorderWidth is the right-edge border width in px (border-r = 1).
	BorderWidth float32
}{
	Width:            256, // w-64
	CollapsedWidth:   48,  // w-12
	HeaderHeight:     56,
	ItemHeight:       36,
	ItemPadX:         12,  // px-3
	ItemRadius:       6,   // rounded-md
	IconSize:         16,  // size-4
	IconGap:          8,   // gap-2
	FontSize:         14,  // text-sm
	FontWeight:       400, // normal
	ActiveFontWeight: 500, // font-medium
	GroupLabelSize:    12,  // text-xs
	GroupLabelWeight:  500, // font-medium
	GroupLabelPadY:    4,
	GroupGap:          4,  // gap-1
	SectionPad:       8,  // p-2
	SectionGap:       8,  // gap-2
	BorderWidth:      1,  // border-r
}

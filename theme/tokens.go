package theme

import "github.com/gogpu/ui/widget"

// Tokens mirrors every CSS variable in shadcn's :root/.dark blocks
// (see docs/research/03-shadcn-pixel-spec.md §1). One Tokens value holds
// the colors for a single mode; Theme carries one for light and one for
// dark.
//
// Naming follows the shadcn convention: every surface token pairs with a
// "-Foreground" token holding the text/icon color drawn on that surface
// (Primary fills a button, PrimaryForeground colors its label).
//
// Dark-mode Border and Input intentionally carry alpha (shadcn uses
// oklch(1 0 0 / 10%) and / 15%); they must be alpha-blended at draw time,
// never precomposited.
type Tokens struct {
	// Background and Foreground are the page surface and default text color
	// (--background / --foreground).
	Background, Foreground widget.Color

	// Card and CardForeground style card surfaces (--card /
	// --card-foreground).
	Card, CardForeground widget.Color

	// Popover and PopoverForeground style floating surfaces such as
	// popovers, menus, and tooltip inversions (--popover /
	// --popover-foreground).
	Popover, PopoverForeground widget.Color

	// Primary and PrimaryForeground style primary actions (--primary /
	// --primary-foreground).
	Primary, PrimaryForeground widget.Color

	// Secondary and SecondaryForeground style secondary actions
	// (--secondary / --secondary-foreground).
	Secondary, SecondaryForeground widget.Color

	// Muted and MutedForeground style subdued surfaces and de-emphasized
	// text (--muted / --muted-foreground).
	Muted, MutedForeground widget.Color

	// Accent and AccentForeground style hover/selection fills (--accent /
	// --accent-foreground).
	Accent, AccentForeground widget.Color

	// Destructive is the error/danger color (--destructive).
	Destructive widget.Color

	// DestructiveForeground is the text color on destructive surfaces.
	// The canonical shadcn theme omits this variable and uses literal
	// text-white instead, so it defaults to white in both modes; CSS
	// imports that define --destructive-foreground override it.
	DestructiveForeground widget.Color

	// Border, Input, and Ring are the stroke colors for general borders,
	// form-control borders, and focus rings (--border / --input / --ring).
	Border, Input, Ring widget.Color

	// Chart holds the five chart series colors --chart-1 .. --chart-5
	// (Chart[0] = --chart-1).
	Chart [5]widget.Color

	// Sidebar and SidebarForeground style the sidebar surface (--sidebar /
	// --sidebar-foreground).
	Sidebar, SidebarForeground widget.Color

	// SidebarPrimary and SidebarPrimaryForeground style primary elements
	// inside the sidebar (--sidebar-primary / --sidebar-primary-foreground).
	SidebarPrimary, SidebarPrimaryForeground widget.Color

	// SidebarAccent and SidebarAccentForeground style hover/selection fills
	// inside the sidebar (--sidebar-accent / --sidebar-accent-foreground).
	SidebarAccent, SidebarAccentForeground widget.Color

	// SidebarBorder and SidebarRing are the sidebar stroke colors
	// (--sidebar-border / --sidebar-ring).
	SidebarBorder, SidebarRing widget.Color
}

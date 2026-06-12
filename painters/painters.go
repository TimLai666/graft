// Package painters provides shadcn-styled painter implementations for
// gogpu/ui core widgets, following the devtools bundle pattern.
//
// Each painter struct holds a *theme.Theme and resolves colors via
// Theme.Active() at paint time, so light/dark mode switches repaint
// correctly without rebuilding widget trees or reinstalling painters.
//
// All painter struct types are pre-declared here so that adding a painter
// implementation is purely additive (one new file with the Paint method
// plus a compile-time interface check) and never edits this shared file.
package painters

import "github.com/TimLai666/graft/theme"

// Button paints core/button widgets in shadcn style.
type Button struct{ Theme *theme.Theme }

// Checkbox paints core/checkbox widgets in shadcn style.
type Checkbox struct{ Theme *theme.Theme }

// Radio paints core/radio widgets in shadcn style.
type Radio struct{ Theme *theme.Theme }

// TextField paints core/textfield widgets in shadcn style.
type TextField struct{ Theme *theme.Theme }

// Dropdown paints core/dropdown widgets (Select trigger + menu) in shadcn style.
type Dropdown struct{ Theme *theme.Theme }

// Slider paints core/slider widgets in shadcn style.
type Slider struct{ Theme *theme.Theme }

// Dialog paints core/dialog widgets in shadcn style.
type Dialog struct{ Theme *theme.Theme }

// Scrollbar paints core/scrollview scrollbars in shadcn style.
type Scrollbar struct{ Theme *theme.Theme }

// TabView paints core/tabview tab bars in shadcn style.
type TabView struct{ Theme *theme.Theme }

// Popover paints core/popover surfaces and tooltips in shadcn style.
type Popover struct{ Theme *theme.Theme }

// Collapsible paints core/collapsible headers in shadcn style.
type Collapsible struct{ Theme *theme.Theme }

// ProgressBar paints core/progressbar widgets in shadcn style.
type ProgressBar struct{ Theme *theme.Theme }

// SplitView paints core/splitview dividers in shadcn style.
type SplitView struct{ Theme *theme.Theme }

// DataTable paints core/datatable widgets in shadcn style.
type DataTable struct{ Theme *theme.Theme }

// ListView paints core/listview widgets in shadcn style.
type ListView struct{ Theme *theme.Theme }

// Menu paints core/menu bars and menus in shadcn style.
type Menu struct{ Theme *theme.Theme }

// LineChart paints core/linechart widgets in shadcn style.
type LineChart struct{ Theme *theme.Theme }

// Painters bundles one painter per themed gogpu/ui core widget, all sharing
// a single *theme.Theme pointer (the devtools.NewPainters pattern).
type Painters struct {
	Button      Button
	Checkbox    Checkbox
	Radio       Radio
	TextField   TextField
	Dropdown    Dropdown
	Slider      Slider
	Dialog      Dialog
	Scrollbar   Scrollbar
	TabView     TabView
	Popover     Popover
	Collapsible Collapsible
	ProgressBar ProgressBar
	SplitView   SplitView
	DataTable   DataTable
	ListView    ListView
	Menu        Menu
	LineChart   LineChart
}

// New returns the full painter bundle for the given theme.
// A nil theme is tolerated: painters fall back to the neutral light palette.
func New(t *theme.Theme) *Painters {
	return &Painters{
		Button:      Button{Theme: t},
		Checkbox:    Checkbox{Theme: t},
		Radio:       Radio{Theme: t},
		TextField:   TextField{Theme: t},
		Dropdown:    Dropdown{Theme: t},
		Slider:      Slider{Theme: t},
		Dialog:      Dialog{Theme: t},
		Scrollbar:   Scrollbar{Theme: t},
		TabView:     TabView{Theme: t},
		Popover:     Popover{Theme: t},
		Collapsible: Collapsible{Theme: t},
		ProgressBar: ProgressBar{Theme: t},
		SplitView:   SplitView{Theme: t},
		DataTable:   DataTable{Theme: t},
		ListView:    ListView{Theme: t},
		Menu:        Menu{Theme: t},
		LineChart:   LineChart{Theme: t},
	}
}

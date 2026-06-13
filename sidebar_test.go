package graft_test

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// buildSidebar returns a sidebar with header, two groups, and footer for tests.
func buildSidebar() *graft.SidebarWidget {
	return graft.Sidebar(
		graft.SidebarHeader(graft.H4("App")),
		graft.SidebarContent(
			graft.SidebarGroup("Navigation",
				graft.SidebarMenuItem("Dashboard").Icon(icons.PanelLeft).Active(true),
				graft.SidebarMenuItem("Search").Icon(icons.Search),
			),
			graft.SidebarGroup("Settings",
				graft.SidebarMenuItem("General").Icon(icons.Info),
			),
		),
		graft.SidebarFooter(graft.MutedText("v1.0")),
	)
}

func TestSidebarSpecLayout(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	sb := buildSidebar()
	size := uitest.LayoutWidget(sb, 256, 400)

	m := metrics.Sidebar

	// Sidebar should take the full constrained width.
	if size.Width != m.Width {
		t.Errorf("sidebar width: got %v, want %v", size.Width, m.Width)
	}
	// Sidebar should take the full constrained height.
	if size.Height != 400 {
		t.Errorf("sidebar height: got %v, want 400", size.Height)
	}
}

func TestSidebarCollapsedLayout(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	sb := buildSidebar().Collapsed(true)
	size := uitest.LayoutWidget(sb, 256, 400)

	m := metrics.Sidebar

	// Collapsed sidebar should be CollapsedWidth wide.
	if size.Width != m.CollapsedWidth {
		t.Errorf("collapsed sidebar width: got %v, want %v", size.Width, m.CollapsedWidth)
	}
}

func TestSidebarLayoutHorizontalSplit(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	sb := buildSidebar()
	main := graft.Text("Main content")

	layout := graft.SidebarLayout(sb, main)
	size := uitest.LayoutWidget(layout, 800, 600)

	m := metrics.Sidebar

	// Layout fills the available space.
	if size.Width != 800 {
		t.Errorf("layout width: got %v, want 800", size.Width)
	}
	if size.Height != 600 {
		t.Errorf("layout height: got %v, want 600", size.Height)
	}

	// Sidebar occupies the left side.
	sbBounds := sb.Bounds()
	if sbBounds.Min.X != 0 {
		t.Errorf("sidebar x: got %v, want 0", sbBounds.Min.X)
	}
	if sbBounds.Width() != m.Width {
		t.Errorf("sidebar width in layout: got %v, want %v", sbBounds.Width(), m.Width)
	}

	// Main content starts after the sidebar.
	mainBounds := main.Bounds()
	if mainBounds.Min.X != m.Width {
		t.Errorf("main x: got %v, want %v", mainBounds.Min.X, m.Width)
	}
	wantMainW := float32(800) - m.Width
	if mainBounds.Width() != wantMainW {
		t.Errorf("main width: got %v, want %v", mainBounds.Width(), wantMainW)
	}
}

func TestSidebarMenuItemHeight(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	item := graft.SidebarMenuItem("Test")
	size := uitest.LayoutWidget(item, 240, 400)

	m := metrics.Sidebar
	if size.Height != m.ItemHeight {
		t.Errorf("item height: got %v, want %v", size.Height, m.ItemHeight)
	}
}

func TestSidebarDrawTokens(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	th := forceLightMode(t)
	tok := th.Active()

	sb := buildSidebar()
	uitest.LayoutWidget(sb, 256, 400)
	cv := uitest.DrawWidget(sb)

	// The sidebar should draw a background rect with the Sidebar token color.
	foundBg := false
	for _, r := range cv.Rects {
		if r.Color == tok.Sidebar {
			foundBg = true
			break
		}
	}
	if !foundBg {
		t.Error("no rect with Sidebar background token found")
	}

	// The sidebar should draw a border rect with the SidebarBorder token color.
	foundBorder := false
	for _, r := range cv.Rects {
		if r.Color == tok.SidebarBorder {
			foundBorder = true
			break
		}
	}
	if !foundBorder {
		t.Error("no rect with SidebarBorder token found")
	}

	// The active item should draw a round rect with the SidebarAccent token.
	foundAccent := false
	for _, rr := range cv.RoundRects {
		if rr.Color == tok.SidebarAccent {
			foundAccent = true
			break
		}
	}
	if !foundAccent {
		t.Error("no round rect with SidebarAccent token found for active item")
	}
}

func TestGoldenSidebar(t *testing.T) {
	gtest.GoldenLightDark(t, "sidebar-expanded", func() widget.Widget {
		return primitives.Box(
			graft.Sidebar(
				graft.SidebarHeader(graft.H4("App")),
				graft.SidebarContent(
					graft.SidebarGroup("Navigation",
						graft.SidebarMenuItem("Dashboard").Icon(icons.PanelLeft).Active(true),
						graft.SidebarMenuItem("Search").Icon(icons.Search),
						graft.SidebarMenuItem("Alerts").Icon(icons.CircleAlert),
					),
					graft.SidebarGroup("More",
						graft.SidebarMenuItem("Info").Icon(icons.Info),
					),
				),
				graft.SidebarFooter(graft.MutedText("v1.0")),
			),
		).Width(256).Height(400)
	})

	gtest.GoldenLightDark(t, "sidebar-collapsed", func() widget.Widget {
		return primitives.Box(
			graft.Sidebar(
				graft.SidebarHeader(graft.H4("A")),
				graft.SidebarContent(
					graft.SidebarGroup("",
						graft.SidebarMenuItem("Dashboard").Icon(icons.PanelLeft).Active(true),
						graft.SidebarMenuItem("Search").Icon(icons.Search),
					),
				),
				graft.SidebarFooter(graft.MutedText("v1")),
			).Collapsed(true),
		).Width(48).Height(300)
	})

	gtest.GoldenLightDark(t, "sidebar-layout", func() widget.Widget {
		sb := graft.Sidebar(
			graft.SidebarHeader(graft.H4("App")),
			graft.SidebarContent(
				graft.SidebarGroup("Nav",
					graft.SidebarMenuItem("Dashboard").Icon(icons.PanelLeft).Active(true),
					graft.SidebarMenuItem("Search").Icon(icons.Search),
				),
			),
			graft.SidebarFooter(graft.MutedText("v1.0")),
		)
		main := primitives.Box(graft.Text("Main content area")).Padding(16)
		return graft.SidebarLayout(sb, main)
	})
}

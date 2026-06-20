// Package graftapp is the one-call desktop launcher for graft.
//
// The core graft module deliberately carries no GPU/windowing dependency so
// components stay importable for headless and offscreen use. graftapp is a
// separate module that adds gogpu/gogpu + gogpu/ui/desktop and folds the
// window-setup boilerplate into a single call:
//
//	package main
//
//	import (
//	    "log"
//
//	    "github.com/TimLai666/graft"
//	    "github.com/TimLai666/graft/graftapp"
//	)
//
//	func main() {
//	    th := graft.NewTheme(graft.BaseColor(graft.Zinc), graft.Radius(10))
//	    err := graftapp.New().
//	        Title("Login").
//	        Size(960, 640).
//	        Theme(th).
//	        Run(graft.Card(
//	            graft.CardHeader(graft.CardTitle("Login to your account")),
//	            graft.CardContent(graft.Input().Placeholder("m@example.com")),
//	            graft.CardFooter(graft.Button("Login").Full()),
//	        ))
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	}
package graftapp

import (
	"fmt"
	"runtime"

	_ "github.com/gogpu/gg/gpu" // GPU SDF acceleration (else every boundary falls back to CPU)

	"github.com/gogpu/gogpu"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/desktop"

	"github.com/TimLai666/graft"
)

// Default window configuration used when a field is left unset.
const (
	defaultTitle  = "graft"
	defaultWidth  = 960
	defaultHeight = 640
)

// App is a fluent builder for a graft desktop window. The zero value is not
// usable; construct it with [New].
type App struct {
	title      string
	width      int
	height     int
	theme      *graft.Theme
	scrollable bool
}

// New returns an App with graft's defaults: a 960x640 window titled "graft",
// the current theme, and a scrollable root.
func New() *App {
	return &App{
		title:      defaultTitle,
		width:      defaultWidth,
		height:     defaultHeight,
		scrollable: true,
	}
}

// Title sets the window title.
func (a *App) Title(title string) *App {
	a.title = title
	return a
}

// Size sets the initial window size in logical pixels.
func (a *App) Size(width, height int) *App {
	a.width = width
	a.height = height
	return a
}

// Theme pins the theme to install. When unset, graft.CurrentTheme() is used.
// The theme is made current before the root is built so component constructors
// snapshot it.
func (a *App) Theme(th *graft.Theme) *App {
	a.theme = th
	return a
}

// Scrollable controls whether the root is wrapped in a shadcn-styled scroll
// area (default true). Pass false when the root manages its own scrolling or
// must fill the window exactly (e.g. a SidebarLayout).
func (a *App) Scrollable(v bool) *App {
	a.scrollable = v
	return a
}

// Run opens the window and blocks until it closes. The root widget is built
// against the installed theme; pass a graft component tree (or any gogpu/ui
// widget).
//
// Run installs graft (fonts, icons, theme) and wires the gogpu window,
// platform, and event source into a gogpu/ui app, then drives the desktop
// render loop.
func (a *App) Run(root graft.Widget) error {
	if root == nil {
		return fmt.Errorf("graftapp: root widget must not be nil")
	}

	th := a.theme
	if th == nil {
		th = graft.CurrentTheme()
	}
	// Make the theme current before wiring so any components built after this
	// point (and the scroll painter below) resolve against it.
	graft.SetTheme(th)

	cfg := gogpu.DefaultConfig().
		WithTitle(a.title).
		WithSize(a.width, a.height)
	// On Windows, default to DX12 instead of Vulkan: AMD Radeon GPUs have a
	// Vulkan driver bug where the even-odd stencil-then-cover pass renders thin
	// strokes (1px borders, spinner arcs, focus rings) as solid filled boxes
	// (gogpu/gg#374). DX12 renders them correctly. Only override when the API is
	// still Auto, so an explicit GOGPU_GRAPHICS_API env choice still wins.
	if runtime.GOOS == "windows" && cfg.GraphicsAPI == gogpu.GraphicsAPIAuto {
		cfg = cfg.WithGraphicsAPI(gogpu.GraphicsAPIDX12)
	}
	gpuApp := gogpu.NewApp(cfg)

	// Wire the OS clipboard into graft-owned widgets (Textarea). Read errors
	// are swallowed to an empty string so a paste with no/unreadable clipboard
	// is a no-op rather than a crash.
	graft.SetClipboard(
		func() string {
			text, err := gpuApp.ClipboardRead()
			if err != nil {
				return ""
			}
			return text
		},
		func(text string) { _ = gpuApp.ClipboardWrite(text) },
	)

	uiApp := app.New(
		app.WithWindowProvider(gpuApp),
		app.WithPlatformProvider(gpuApp),
		app.WithEventSource(gpuApp.EventSource()),
		app.WithTheme(th.AsUITheme()),
	)
	if err := graft.Install(uiApp, th); err != nil {
		return err
	}

	if a.scrollable {
		uiApp.SetRoot(scrollview.New(root,
			scrollview.PainterOpt(graft.PaintersFor(th).Scrollbar)))
	} else {
		uiApp.SetRoot(root)
	}

	return desktop.Run(gpuApp, uiApp)
}

// Run is a convenience wrapper for the common case: open a default-sized
// window titled title showing root (scrollable, current theme).
//
//	graftapp.Run("My App", myRoot)
func Run(title string, root graft.Widget) error {
	return New().Title(title).Run(root)
}

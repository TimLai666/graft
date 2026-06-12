// Package graft is a Go-native, pixel-faithful port of shadcn/ui built on
// github.com/gogpu/ui.
//
// graft mirrors shadcn's usage model: components are constructed with
// content-first constructors and chainable variant/size methods, themed
// through the same CSS-variable token system (copied verbatim, OKLCH
// included), and customized the same way — pick a base color, set a
// radius, override individual tokens, or paste a whole theme-editor CSS
// block.
//
//	th := graft.NewTheme(graft.BaseColor(graft.Zinc), graft.Radius(10))
//	graft.Install(uiApp, th)
//
//	card := graft.Card(
//	    graft.CardHeader(
//	        graft.CardTitle("Login to your account"),
//	        graft.CardDescription("Enter your email below"),
//	    ),
//	    graft.CardContent(form),
//	)
package graft

import (
	"sync"

	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/painters"
	"github.com/TimLai666/graft/theme"
)

// Widget is gogpu/ui's widget interface; graft trees mix freely with raw
// primitives and any other gogpu/ui widget.
type Widget = widget.Widget

// Variant selects a shadcn component variant (button, badge, ...).
type Variant uint8

// Component variants, matching shadcn's cva variant names.
const (
	VariantDefault Variant = iota
	VariantSecondary
	VariantDestructive
	VariantOutline
	VariantGhost
	VariantLink
)

// Size selects a shadcn component size (button, ...).
type Size uint8

// Component sizes, matching shadcn's cva size names. The Icon sizes are
// square icon-button sizes.
const (
	SizeDefault Size = iota
	SizeXS
	SizeSM
	SizeLG
	SizeIcon
	SizeIconXS
	SizeIconSM
	SizeIconLG
)

// Style is the per-component escape hatch: targeted overrides applied on
// top of the shadcn look (the Go equivalent of adding utility classes).
// Nil fields keep the component default.
type Style struct {
	Background *widget.Color
	Foreground *widget.Color
	Radius     *float32
	PadX, PadY *float32
	MinWidth   *float32
	MaxWidth   *float32
}

// Re-exports so simple apps only import graft.
type (
	// Theme is the graft design system (see the theme package).
	Theme = theme.Theme

	// Mode selects light, dark, or system token resolution.
	Mode = theme.Mode

	// BaseColorPreset names one of the five shadcn base-color presets.
	BaseColorPreset = theme.Base
)

// Theme construction re-exports.
var (
	// NewTheme builds a Theme; see theme.New.
	NewTheme = theme.New

	// BaseColor selects a base-color preset; see theme.BaseColor.
	BaseColor = theme.BaseColor

	// Radius sets --radius in pixels; see theme.Radius.
	Radius = theme.Radius

	// Fonts overrides the font family names; see theme.Fonts.
	Fonts = theme.Fonts

	// ParseThemeCSS imports a shadcn/tweakcn theme-editor CSS block.
	ParseThemeCSS = theme.ParseThemeCSS

	// OKLCH converts a CSS oklch() color so shadcn token values can be
	// copied verbatim; see theme.OKLCH.
	OKLCH = theme.OKLCH

	// OKLCHA is OKLCH with explicit alpha.
	OKLCHA = theme.OKLCHA
)

// Mode and base-color constant re-exports.
const (
	ModeLight  = theme.ModeLight
	ModeDark   = theme.ModeDark
	ModeSystem = theme.ModeSystem

	Neutral = theme.Neutral
	Stone   = theme.Stone
	Zinc    = theme.Zinc
	Gray    = theme.Gray
	Slate   = theme.Slate
)

var (
	currentMu    sync.RWMutex
	currentTheme *theme.Theme

	// painterCache maps *theme.Theme to its shared *painters.Painters
	// bundle so every component built against a theme reuses one bundle.
	painterCache sync.Map
)

// CurrentTheme returns the process-wide theme that component constructors
// snapshot. Before Install/SetTheme it lazily defaults to the stock
// neutral theme, so components work out of the box.
func CurrentTheme() *theme.Theme {
	currentMu.RLock()
	t := currentTheme
	currentMu.RUnlock()
	if t != nil {
		return t
	}

	currentMu.Lock()
	defer currentMu.Unlock()
	if currentTheme == nil {
		currentTheme = theme.New()
	}
	return currentTheme
}

// SetTheme replaces the process-wide current theme. Components constructed
// afterwards use it; existing components keep the theme they snapshotted.
func SetTheme(t *theme.Theme) {
	if t == nil {
		return
	}
	currentMu.Lock()
	currentTheme = t
	currentMu.Unlock()
}

// PaintersFor returns the shared shadcn painter bundle for a theme,
// building it on first use. A nil theme resolves to CurrentTheme.
//
// Raw gogpu/ui users can wire these painters manually:
//
//	p := graft.PaintersFor(th)
//	btn := button.New(button.Text("Run"), button.PainterOpt(p.Button))
func PaintersFor(t *theme.Theme) *painters.Painters {
	if t == nil {
		t = CurrentTheme()
	}
	if p, ok := painterCache.Load(t); ok {
		return p.(*painters.Painters)
	}
	p, _ := painterCache.LoadOrStore(t, painters.New(t))
	return p.(*painters.Painters)
}

var installOnce sync.Once

// loadAssets registers fonts, icons, and text-measurement faces exactly
// once per process. Exposed through Install and the test harness.
func loadAssets() error {
	var err error
	installOnce.Do(func() {
		if e := fonts.Load(); e != nil {
			err = e
			return
		}
		for name, face := range fonts.Faces() {
			if e := textmetrics.Register(name, face.Data); e != nil {
				err = e
				return
			}
		}
		err = icons.Register()
	})
	return err
}

// LoadAssets makes graft's embedded fonts and icons available without an
// app (offscreen rendering, tests). Install calls it automatically.
func LoadAssets() error { return loadAssets() }

// Install wires graft into a gogpu/ui app: loads fonts and icons, makes th
// the current theme, applies it to the app (window clear color, primitive
// text defaults), and re-applies it whenever the theme mode changes so a
// SetMode call repaints the whole window in the other scheme.
//
// uiApp may be nil (headless use); Install then only loads assets and sets
// the current theme.
func Install(uiApp *app.App, th *theme.Theme) error {
	if th == nil {
		th = CurrentTheme()
	}
	if err := loadAssets(); err != nil {
		return err
	}
	SetTheme(th)

	if uiApp != nil {
		uiApp.SetTheme(th.AsUITheme())
		th.OnModeChange(func(theme.Mode) {
			uiApp.SetTheme(th.AsUITheme())
		})
	}
	return nil
}

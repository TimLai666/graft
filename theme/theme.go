package theme

import (
	"slices"
	"sync"

	uitheme "github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

// Mode selects which token set a Theme resolves: light, dark, or the
// operating system preference.
type Mode uint8

const (
	// ModeLight always resolves the light token set.
	ModeLight Mode = iota

	// ModeDark always resolves the dark token set.
	ModeDark

	// ModeSystem follows the operating system preference, queried through
	// SystemDarkHook. With the default hook it resolves to light.
	ModeSystem
)

// String returns a human-readable name for the mode.
func (m Mode) String() string {
	switch m {
	case ModeLight:
		return "light"
	case ModeDark:
		return "dark"
	case ModeSystem:
		return "system"
	default:
		return "unknown"
	}
}

// SystemDarkHook reports whether the operating system currently prefers a
// dark color scheme. It is consulted whenever the theme mode is ModeSystem.
//
// The default implementation is dependency-free and always reports false
// (light), so ModeSystem falls back to light until a real detector is
// installed. graft.Install (or any application) may override it, e.g. on
// Windows by reading HKCU\Software\Microsoft\Windows\CurrentVersion\
// Themes\Personalize\AppsUseLightTheme:
//
//	theme.SystemDarkHook = func() bool { return readAppsUseLightTheme() == 0 }
//
// The hook may be called from any goroutine; implementations must be safe
// for concurrent use.
var SystemDarkHook = func() bool { return false }

// Default font family names registered by the graft fonts package.
const (
	// DefaultFontSans is the default sans-serif family ("Geist").
	DefaultFontSans = "Geist"

	// DefaultFontMono is the default monospace family ("Geist Mono").
	DefaultFontMono = "Geist Mono"
)

// defaultRadius is shadcn's --radius default: 0.625rem = 10px.
const defaultRadius float32 = 10

// Theme is the graft design system: both shadcn token sets plus the radius
// and font knobs from components.json. The zero value is usable but empty;
// construct with New.
//
// Token fields may be overridden directly per mode, exactly like editing
// the :root/.dark blocks:
//
//	th := theme.New(theme.BaseColor(theme.Zinc), theme.Radius(10))
//	th.Light.Primary = theme.OKLCH(0.55, 0.20, 260)
//	th.Dark.Ring = theme.OKLCH(0.55, 0.15, 260)
//
// Mode handling (SetMode/Mode/IsDark/Active) is safe for concurrent use.
// Token and Radius fields follow gogpu/ui's theme convention: configure
// them up front, treat them as immutable afterwards.
type Theme struct {
	// Light and Dark are the two token sets from the :root and .dark
	// blocks. Active returns the one matching the current mode.
	Light, Dark Tokens

	// Radius is shadcn's --radius in pixels (default 10 = 0.625rem). The
	// derived scale (RadiusSM, RadiusMD, ...) is computed from it; see
	// radius.go.
	Radius float32

	// FontSans is the registered sans-serif font family name used for all
	// component text (default "Geist").
	FontSans string

	// FontMono is the registered monospace family name (default
	// "Geist Mono").
	FontMono string

	mu        sync.RWMutex
	mode      Mode
	listeners []func(Mode)
}

// Option configures New.
type Option func(*config)

type config struct {
	base     Base
	radius   *float32
	fontSans string
	fontMono string
}

// BaseColor selects one of the five shadcn base-color presets (the
// components.json "baseColor" knob). The default is Neutral.
func BaseColor(b Base) Option {
	return func(c *config) { c.base = b }
}

// Radius sets shadcn's --radius in pixels (the components.json radius
// knob). The default is 10 (0.625rem).
func Radius(px float32) Option {
	return func(c *config) { c.radius = &px }
}

// Fonts sets the sans-serif and monospace font family names. The defaults
// are DefaultFontSans and DefaultFontMono.
func Fonts(sans, mono string) Option {
	return func(c *config) {
		c.fontSans = sans
		c.fontMono = mono
	}
}

// New builds a Theme from a base-color preset. With no options it is the
// stock shadcn neutral theme: radius 10, Geist fonts, mode ModeSystem
// (which resolves to light under the default SystemDarkHook).
func New(opts ...Option) *Theme {
	cfg := config{
		base:     Neutral,
		fontSans: DefaultFontSans,
		fontMono: DefaultFontMono,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	p := preset(cfg.base)
	t := &Theme{
		Light:    p.light,
		Dark:     p.dark,
		Radius:   p.radius,
		FontSans: cfg.fontSans,
		FontMono: cfg.fontMono,
		mode:     ModeSystem,
	}
	if cfg.radius != nil {
		t.Radius = *cfg.radius
	}
	return t
}

// Mode returns the current mode.
func (t *Theme) Mode() Mode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.mode
}

// SetMode switches the theme mode and notifies OnModeChange subscribers.
// Setting the current mode again is a no-op. Callbacks run synchronously
// on the calling goroutine, outside the theme's internal lock.
func (t *Theme) SetMode(m Mode) {
	t.mu.Lock()
	if t.mode == m {
		t.mu.Unlock()
		return
	}
	t.mode = m
	listeners := slices.Clone(t.listeners)
	t.mu.Unlock()

	for _, fn := range listeners {
		fn(m)
	}
}

// OnModeChange registers a callback invoked from SetMode whenever the mode
// actually changes. Callbacks cannot be removed; keep them owned by
// long-lived installers (graft.Install uses this to re-apply the gogpu/ui
// app theme on mode switches).
func (t *Theme) OnModeChange(fn func(Mode)) {
	if fn == nil {
		return
	}
	t.mu.Lock()
	t.listeners = append(t.listeners, fn)
	t.mu.Unlock()
}

// IsDark reports whether the theme currently resolves to the dark token
// set (ModeSystem consults SystemDarkHook). It is half of
// widget.ThemeProvider.
func (t *Theme) IsDark() bool {
	switch t.Mode() {
	case ModeDark:
		return true
	case ModeSystem:
		if hook := SystemDarkHook; hook != nil {
			return hook()
		}
		return false
	default:
		return false
	}
}

// Active returns the token set for the current mode. Painters call this
// inside every draw, so mode switches repaint correctly without rebuilding
// the widget tree.
//
// The returned pointer aliases the Theme's Light or Dark field; treat it
// as read-only.
func (t *Theme) Active() *Tokens {
	if t.IsDark() {
		return &t.Dark
	}
	return &t.Light
}

// OnSurface returns the default text/icon color (the active Foreground
// token). It is the other half of widget.ThemeProvider.
func (t *Theme) OnSurface() widget.Color {
	return t.Active().Foreground
}

// Colors with no shadcn counterpart, needed to fill gogpu/ui's M3-shaped
// ColorPalette in AsUITheme. Values are Tailwind v4 amber-500, green-600,
// and blue-500; graft components never read them — they only matter for
// raw gogpu/ui widgets painted with the default painters.
var (
	paletteWarning = MustOKLCH("oklch(0.769 0.188 70.08)")
	paletteSuccess = MustOKLCH("oklch(0.627 0.194 149.214)")
	paletteInfo    = MustOKLCH("oklch(0.623 0.214 259.815)")
)

// AsUITheme adapts the active token set to a gogpu/ui *theme.Theme for
// app.WithTheme/SetTheme. This only drives the window clear color and
// primitive text defaults — core-widget visuals come from the graft
// painters, which read Tokens directly.
//
// The mapping (DESIGN.md §2.1): Background→Background, Surface→Card,
// SurfaceVariant→Muted, Primary→Primary, OnPrimary→PrimaryForeground,
// Secondary→Secondary, OnBackground/OnSurface→Foreground,
// Error→Destructive, OnError→DestructiveForeground, Outline/Divider→Border,
// Shadow→black at 10% alpha.
//
// The result is registered in the gogpu/ui theme registry as "graft-light"
// or "graft-dark" (matching the resolved mode), and the graft Theme itself
// is attached as the typed extension "graft" so third-party gogpu/ui code
// can recover the full token set:
//
//	if g, ok := theme.ExtensionAs[*graftheme.Theme](t, "graft"); ok { ... }
func (t *Theme) AsUITheme() *uitheme.Theme {
	dark := t.IsDark()
	tok := t.Active()

	name := "graft-light"
	mode := uitheme.ModeLight
	if dark {
		name = "graft-dark"
		mode = uitheme.ModeDark
	}

	ut := uitheme.New(name, mode)
	ut.Colors = uitheme.ColorPalette{
		Primary:      tok.Primary,
		PrimaryLight: tok.Primary,
		PrimaryDark:  tok.Primary,

		Secondary:      tok.Secondary,
		SecondaryLight: tok.Secondary,
		SecondaryDark:  tok.Secondary,

		Background:     tok.Background,
		Surface:        tok.Card,
		SurfaceVariant: tok.Muted,

		Error:   tok.Destructive,
		Warning: paletteWarning,
		Success: paletteSuccess,
		Info:    paletteInfo,

		OnPrimary:    tok.PrimaryForeground,
		OnSecondary:  tok.SecondaryForeground,
		OnBackground: tok.Foreground,
		OnSurface:    tok.Foreground,
		OnError:      tok.DestructiveForeground,

		Divider: tok.Border,
		Outline: tok.Border,
		Shadow:  widget.RGBA(0, 0, 0, 0.10),
	}
	if dark {
		ut.Shadows = uitheme.DefaultShadowsDark()
	}
	ut.Typography.FontFamily = t.FontSans
	ut.Radii = uitheme.RadiusScale{
		None: 0,
		XS:   t.RadiusXS(),
		S:    t.RadiusSM(),
		M:    t.RadiusMD(),
		L:    t.RadiusLG(),
		XL:   t.RadiusXL(),
		XXL:  t.Radius2XL(),
		Full: t.RadiusFull(),
	}

	// Interop courtesy (DESIGN.md §2.6): graft widgets hold *Theme
	// directly; the extension only lets third-party code recover tokens.
	ut.RegisterExtension(t)

	uitheme.Register(name, ut, uitheme.ThemeInfo{
		Name:        "graft",
		Description: "shadcn/ui design system for gogpu/ui",
		Author:      "TimLai666",
		Version:     "0.1.0",
		Variants:    []uitheme.ThemeVariant{uitheme.VariantLight, uitheme.VariantDark},
	})
	return ut
}

// Name returns the typed-extension key "graft"
// (gogpu/ui theme.ThemeExtension).
func (t *Theme) Name() string { return "graft" }

// Merge implements theme.ThemeExtension: the other graft theme (the
// override) wins wholesale; non-graft extensions are ignored.
func (t *Theme) Merge(other uitheme.ThemeExtension) uitheme.ThemeExtension {
	if o, ok := other.(*Theme); ok {
		return o
	}
	return t
}

// Lerp implements theme.ThemeExtension. Graft themes are not animated
// through the extension system, so this is a step function: t below the
// midpoint, other at or above it.
func (t *Theme) Lerp(other uitheme.ThemeExtension, amount float32) uitheme.ThemeExtension {
	if o, ok := other.(*Theme); ok && amount >= 0.5 {
		return o
	}
	return t
}

// CopyWith implements theme.ThemeExtension as a passthrough; graft themes
// are customized through their exported fields instead.
func (t *Theme) CopyWith(map[string]any) uitheme.ThemeExtension { return t }

// Compile-time interface checks.
var (
	_ widget.ThemeProvider   = (*Theme)(nil)
	_ uitheme.ThemeExtension = (*Theme)(nil)
)

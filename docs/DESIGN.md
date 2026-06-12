# graft — Design Document

**Module:** `github.com/TimLai666/graft` · **Target:** pixel-faithful Go-native port of shadcn/ui (new-york-v4 registry) on `github.com/gogpu/ui`.
**Status:** Final. Implementers follow this document without further research. Source-of-truth references: Report 1 (gogpu/ui theming internals), Report 2 (widget cookbook), Report 3 (shadcn pixel spec), Report 4 (inventory + API proposal). Where this doc and a report disagree, this doc wins.

**Ground rules**

- graft is two things at once: (a) a 5th gogpu/ui design system — a token model + a `Painters` bundle in the devtools/fluent pattern — usable with raw gogpu/ui widgets, and (b) a high-level shadcn-shaped component API on top.
- Every pixel value in painters and composites comes from the `metrics` package (which transcribes Report 3), never inline literals.
- Every painter file ends with a compile-time interface check (`var _ button.Painter = Button{}`).
- One source file per component in the root package, named after the shadcn registry entry (`button.go`, `dropdown-menu.go` → `dropdownmenu.go`). This is what makes parallel agent work safe (§7).
- graft components never cache token colors; all colors resolve from the theme at `Draw` time so `SetTheme`/mode-switch works without rebuilding the tree.

---

## 1. Package layout

```
github.com/TimLai666/graft
├── go.mod                      # deps: github.com/gogpu/ui, github.com/gogpu/gg, github.com/gogpu/gogpu, golang.org/x/image
├── graft.go                    # Widget alias, Variant/Size enums, Style escape hatch, Install(), current-theme global
├── button.go                   # one file per component, shadcn registry names (see §4); ~50 files total over all phases
├── card.go, input.go, checkbox.go, switch.go, badge.go, dialog.go, select.go, tabs.go,
│   dropdownmenu.go, label.go, field.go, form.go, typography.go, alert.go, separator.go,
│   tooltip.go, popover.go, scrollarea.go, slider.go, progress.go, spinner.go, radiogroup.go, ... 
├── theme/                      # the design-system package (public; usable without the root API)
│   ├── tokens.go               # Tokens struct (1:1 shadcn CSS variables), Mode type
│   ├── theme.go                # Theme{Light,Dark,Radius,...}, Active(), AsUITheme(), widget.ThemeProvider impl, registry registration
│   ├── oklch.go                # OKLCH()/OKLCHA() constructors + OKLab→sRGB math + gamut mapping (§2.2)
│   ├── presets.go              # base-color presets (Neutral/Stone/Zinc/Gray/Slate) as embedded CSS parsed at init (§2.4)
│   ├── presets/*.css           # verbatim :root/.dark blocks copied from ui.shadcn.com/docs/theming (go:embed)
│   ├── css.go                  # ParseThemeCSS — imports shadcn/tweakcn theme-editor CSS
│   └── radius.go               # radius scale derivation: XS..4XL/Full from --radius (§2.5)
├── metrics/                    # exact px constants from Report 3, one file per component (§5.1); pure data, no imports beyond stdlib
│   ├── button.go, input.go, checkbox.go, switch.go, card.go, dialog.go, select.go, tabs.go, menu.go,
│   ├── badge.go, alert.go, slider.go, table.go, ...
│   └── focus.go, shadow.go     # focus-ring + shadow-layer constants shared across components
├── painters/                   # gogpu/ui core-widget painters in the devtools bundle pattern (§3)
│   ├── painters.go             # type Painters struct{...}; func New(t *theme.Theme) *Painters
│   └── button.go, checkbox.go, radio.go, textfield.go, dropdown.go, slider.go, dialog.go,
│       scrollbar.go, tabview.go, popover.go, collapsible.go, progressbar.go, splitview.go,
│       datatable.go, listview.go, menu.go, linechart.go        # one painter per file
├── icons/                      # lucide subset: vendored SVG XML (go:embed) + icon.FromSVGXML registration; gen/ script to vendor more
├── fonts/                      # embedded Geist + Geist Mono TTFs (OFL) + Load() registering families (§5.5)
├── internal/
│   ├── draw/                   # shared paint helpers: FocusRing, Shadow, InsideBorder, Fade, Alpha/MulAlpha, styled text (§5)
│   ├── textmetrics/            # sfnt-based text measurement (golang.org/x/image) for layout-time widths (§5.6)
│   └── widgets/                # new interactive widgets (C-class): switchw/, toggle/, menu/ (shared menu engine),
│                               #   sheet/, toastregion/, calendar/, otp/, textarea/, command/, carousel/
├── examples/                   # compile-checked docs: login card, settings page, dashboard, kitchen-sink "reference sheet"
└── testdata/golden/            # golden PNGs per component/state (§6)
```

Import graph (no cycles): root → `theme`, `painters`, `metrics`, `icons`, `fonts`, `internal/*`; `painters` → `theme`, `metrics`, `internal/draw`; `internal/widgets/*` → `theme`, `metrics`, `internal/draw`, `internal/textmetrics`. `theme` and `metrics` import nothing from graft.

---

## 2. Theme architecture

### 2.1 Tokens — 1:1 mirror of shadcn CSS variables

```go
package theme

import "github.com/gogpu/ui/widget"

// Tokens mirrors every variable in shadcn's :root/.dark blocks (Report 3 §1).
type Tokens struct {
    Background, Foreground             widget.Color
    Card, CardForeground               widget.Color
    Popover, PopoverForeground         widget.Color
    Primary, PrimaryForeground         widget.Color
    Secondary, SecondaryForeground     widget.Color
    Muted, MutedForeground             widget.Color
    Accent, AccentForeground           widget.Color
    Destructive                        widget.Color
    DestructiveForeground              widget.Color // canonical theme omits it; default = white both modes (shadcn uses literal text-white)
    Border, Input, Ring                widget.Color
    Chart                              [5]widget.Color
    Sidebar, SidebarForeground         widget.Color
    SidebarPrimary, SidebarPrimaryForeground widget.Color
    SidebarAccent, SidebarAccentForeground   widget.Color
    SidebarBorder, SidebarRing         widget.Color
}

type Mode uint8
const (ModeLight Mode = iota; ModeDark; ModeSystem)

type Theme struct {
    Light, Dark Tokens
    Radius      float32 // px; default 10 (--radius: 0.625rem)
    FontSans    string  // registered family name; default "Geist"
    FontMono    string  // default "Geist Mono"

    mode     state.Signal[Mode] // installed by graft.Install
    painters *painters.Painters // lazy, shared; built on first use
}

func New(opts ...Option) *Theme            // Option: BaseColor(b), Radius(px), Fonts(sans, mono string)
func (t *Theme) Active() *Tokens           // resolves mode (System → platform hint, fallback Light); read by every painter on every Draw
func (t *Theme) IsDark() bool              // widget.ThemeProvider
func (t *Theme) OnSurface() widget.Color   // widget.ThemeProvider → Active().Foreground
func (t *Theme) AsUITheme() *uitheme.Theme // maps tokens → theme.ColorPalette (below) for app.WithTheme/SetTheme
func ParseThemeCSS(css string) (*Theme, error)
func (t *Theme) ApplyCSS(css string) error // merge a :root/.dark block onto an existing theme (partial overrides)
```

`AsUITheme()` mapping (so window clear color and `primitives.Text` default color are right): `Background→Background`, `Surface→Card`, `SurfaceVariant→Muted`, `Primary→Primary`, `OnPrimary→PrimaryForeground`, `Secondary→Secondary`, `OnBackground/OnSurface→Foreground`, `Error→Destructive`, `OnError→DestructiveForeground`, `Outline→Border`, `Divider→Border`, `Shadow→RGBA(0,0,0,0.10)`. Mode maps to `ModeLight/ModeDark`. It also calls `uitheme.Register("graft-light"/"graft-dark", ...)` once via `init()`-safe registration, with `ThemeInfo{Name:"graft", Author:"TimLai666"}`.

Remember Report 1 §7.1: `app.SetTheme` does **not** retheme core widgets — that is what `painters` + graft constructors are for. `AsUITheme` only handles window background + primitive text defaults.

### 2.2 OKLCH support — token values copied verbatim from shadcn CSS

`theme/oklch.go` implements CSS Color 4 `oklch()` → sRGB so every literal in shadcn CSS pastes directly:

```go
func OKLCH(l, c, h float64) widget.Color        // h in degrees, as in CSS
func OKLCHA(l, c, h, alpha float64) widget.Color
```

Algorithm (implement exactly; no library):
1. `a = c·cos(h·π/180)`, `b = c·sin(h·π/180)`.
2. OKLab → LMS′: `l′ = L + 0.3963377774a + 0.2158037573b`; `m′ = L − 0.1055613458a − 0.0638541728b`; `s′ = L − 0.0894841775a − 1.2914855480b`. Cube each: `l = l′³` etc.
3. LMS → linear sRGB: `R = 4.0767416621l − 3.3077115913m + 0.2309699292s`; `G = −1.2684380046l + 2.6097574011m − 0.3413193965s`; `B = −0.0041960863l − 0.7034186147m + 1.7076147010s`.
4. Gamma: `f(x) = x ≤ 0.0031308 ? 12.92x : 1.055·x^(1/2.4) − 0.055`.
5. **Gamut mapping (decision, REVISED after empirical validation — implemented & tested in `theme/oklch.go`):** out-of-gamut channels are **clipped per linear channel** (clamp linear RGB to [0,1] before gamma encoding). This was validated against Tailwind v4's published sRGB fallbacks: `oklch(0.577 0.245 27.325)` (shadcn `--destructive`) → `#e7000b` **exactly** under linear clipping, while chroma-reduction produces a visibly different `#e30017`-ish result. Chrome on sRGB displays clips the same way, so clipping — not the CSS4 chroma-reduction algorithm — is what "matches what Chrome renders". The original recommendation in this section was wrong and is superseded.

Unit-test vectors (exact, Tailwind **v4** palette — note v4 re-tuned chroma vs v3; do not use v3 hexes):
| oklch | hex | note |
|---|---|---|
| `oklch(1 0 0)` | `#ffffff` | |
| `oklch(0.985 0 0)` | `#fafafa` | |
| `oklch(0.97 0 0)` | `#f5f5f5` | |
| `oklch(0.922 0 0)` | `#e5e5e5` | |
| `oklch(0.708 0 0)` | `#a1a1a1` | v3 was `#a3a3a3` (L=0.715); v4 is `#a1a1a1` |
| `oklch(0.556 0 0)` | `#737373` | |
| `oklch(0.269 0 0)` | `#262626` | |
| `oklch(0.205 0 0)` | `#171717` | |
| `oklch(0.145 0 0)` | `#0a0a0a` | |
| `oklch(0.637 0.237 25.331)` | `#fb2c36` | in-gamut chromatic (red-500) |
| `oklch(0.546 0.245 262.881)` | `#155dfc` | in-gamut chromatic (blue-600) |
| `oklch(0.577 0.245 27.325)` | `#e7000b` | out-of-gamut → clipping path (red-600 / destructive) |

These are implemented and green in `theme/oklch_test.go`, plus an independent inverse-transform round-trip property test.

### 2.3 Light/dark mode

- `Theme` carries **both** token sets; `Active()` picks by the mode signal. Painters hold `*theme.Theme` (one shared pointer, devtools pattern) and call `Active()` inside every `Paint*` — therefore **mode switching requires no tree rebuild and no painter reinstall**.
- `graft.Install(uiApp, th, mode)` subscribes to the mode signal: on change it calls `uiApp.SetTheme(th.AsUITheme())` (updates window clear color, triggers full relayout/redraw per Report 1 §3.2) — that single call repaints everything because painters re-read `Active()`.
- `ModeSystem`: v1 resolves via a platform hook — Windows reads `HKCU\...\Themes\Personalize\AppsUseLightTheme`; other platforms fall back to Light. (Decision: don't block v1 on cross-platform detection; the hook is a single `func() bool` variable apps can override.)
- Dark-mode alpha tokens (`border = oklch(1 0 0 / 10%)`, `input = 15%`) are stored **with alpha intact** — the canvas alpha-blends, which exactly reproduces CSS compositing over varying surfaces. Never precomposite.

### 2.4 User customization API (the shadcn workflow, in Go)

```go
// 1. components.json step: base color + radius
th := theme.New(theme.BaseColor(theme.Zinc), theme.Radius(10))
// BaseColor presets: Neutral (default), Stone, Zinc, Gray, Slate

// 2. :root/.dark step: override token fields directly per mode
th.Light.Primary = theme.OKLCH(0.55, 0.20, 260)
th.Light.PrimaryForeground = theme.OKLCH(0.98, 0, 0)
th.Dark.Ring = theme.OKLCH(0.55, 0.15, 260)

// 3. Or paste a whole theme from the shadcn/tweakcn theme editor
th, err := theme.ParseThemeCSS(cssText)   // parses :root{} and .dark{} blocks; oklch()/hsl()/hex/rgb() values; --radius in rem or px

// 4. Install: dark mode is a signal
mode := state.NewSignal(theme.ModeSystem)
graft.Install(uiApp, th, mode)            // sets global current theme + app theme + mode subscription + loads fonts/icons
```

**Preset implementation (decision):** the five base-color presets are stored as embedded CSS files copied **verbatim** from ui.shadcn.com/docs/theming (one `:root` + `.dark` block each) and run through `ParseThemeCSS` at package init. Zero transcription risk, and the CSS parser is exercised on every startup. Tradeoff: a few ms of init parsing — negligible.

The root package re-exports for ergonomics: `type Theme = theme.Theme`, `func NewTheme(...) = theme.New` wrapper, `func OKLCH(...)` wrapper, `const Neutral/Stone/Zinc/Gray/Slate`, `type Mode = theme.Mode`.

### 2.5 Radius scale

`--radius` default 10px. Derived (multiplicative, current formula — Report 3 §2): `XS = 2` (fixed Tailwind), `SM = 0.6r = 6`, `MD = 0.8r = 8`, `LG = r = 10`, `XL = 1.4r = 14`, `2XL = 1.8r`, `3XL = 2.2r`, `4XL = 2.6r`, `Full = 9999`. `theme/radius.go` exposes `func (t *Theme) RadiusSM() float32` etc.; component code never hardcodes 6/8/10/14 — it asks the theme, so the user's `Radius` knob propagates everywhere, exactly like shadcn.

### 2.6 ThemeExtension

`AsUITheme()` additionally attaches the graft theme as a typed extension (`theme.RegisterExtension`, name `"graft"`) so any third-party gogpu/ui code can recover tokens via `ExtensionAs[*theme.Theme](t, "graft")`. graft's own widgets do **not** go through the extension (they hold the `*Theme` directly) — the extension is interop courtesy only.

---

## 3. Painter strategy

### 3.1 Painters for existing gogpu/ui core widgets (Class A)

`painters.New(t *theme.Theme) *Painters` returns the bundle (devtools pattern). Each struct is `{ Theme *theme.Theme }` and resolves colors via `Theme.Active()` at paint time, **honoring `PaintState` overrides** (`Background`, `Radius`, non-zero `ColorScheme`) per the M3 convention.

| Painter struct | Implements (Report 1 §2.2) | Methods | Used by graft component(s) |
|---|---|---|---|
| `painters.Button` | `core/button.Painter` | `PaintButton` | Button (all 6 variants × 8 sizes) |
| `painters.Checkbox` | `core/checkbox.Painter` | `PaintCheckbox` | Checkbox |
| `painters.Radio` | `core/radio.Painter` | `PaintRadio` | RadioGroup items |
| `painters.TextField` | `core/textfield.Painter` | `PaintTextField` | Input (theming flows only through painter struct — no ColorScheme in its PaintState) |
| `painters.Dropdown` | `core/dropdown.Painter` | `PaintTrigger`, `PaintMenu` | Select |
| `painters.Slider` | `core/slider.Painter` | `PaintSlider` | Slider |
| `painters.Dialog` | `core/dialog.Painter` | `PaintDialog` | Dialog, AlertDialog |
| `painters.Scrollbar` | `core/scrollview.Painter` | `PaintScrollbar` | ScrollArea (and every scrollable graft surface) |
| `painters.TabView` | `core/tabview.Painter` | `PaintTabBar` | Tabs |
| `painters.Popover` | `core/popover.Painter` | `PaintPopover`, `PaintTooltip` | Popover, Tooltip, HoverCard |
| `painters.Collapsible` | `core/collapsible.Painter` | `PaintHeader` | Collapsible, Accordion items |
| `painters.ProgressBar` | `core/progressbar.Painter` | `PaintProgressBar` | Progress |
| `painters.SplitView` | `core/splitview.Painter` | `PaintDivider` | Resizable |
| `painters.DataTable` | `core/datatable.Painter` | `PaintHeader/HeaderCell/Row/Cell/EmptyState` | DataTable |
| `painters.ListView` | `core/listview.Painter` | `PaintDivider/EmptyState/ItemBackground/Selection` | Command, Combobox lists |
| `painters.Menu` | `core/menu.Painter` | `PaintMenuBar`, `PaintMenu` | Menubar (bar + its drop-downs) |
| `painters.LineChart` (Phase 3) | `core/linechart.Painter` | `PaintChart` | Chart |

**Not implemented** (no shadcn counterpart / out of scope): treeview, gridview, toolbar, docking, titlebar, stripe, progress (circular — Spinner is a spinning lucide icon instead, matching shadcn exactly).

### 3.2 New widgets and composites (Class B/C)

- **Composites of primitives (B)** — built from `primitives.Box/HBox/Text/Image` + graft typography, in root component files: Card, Alert, Badge, Label, Separator, Field/FieldGroup/FieldSet, Form, Typography (H1–H4, P, Blockquote, Code, Lead, Large, Small, Muted), Avatar (+Group), Skeleton, Kbd, Breadcrumb, Pagination, Empty, Item, ButtonGroup, InputGroup, Table (static styled), Sidebar, DatePicker (Popover+Calendar+Button), AspectRatio, Spinner (animated icon).
- **New interactive widgets (C)** — in `internal/widgets/*`, each following the Report 2 §1.6 widget template (WidgetBase embed, painter-less — they draw directly via `internal/draw` + `metrics`):
  - `switchw` — **Switch is a new widget**, not a checkbox painter (deviation from Report 4): painters cannot affect layout (Report 1 §7.6) and the checkbox widget's box is a fixed square; the 32×18.4 track + animated 14px thumb travel (150ms, cubic-bezier(0.4,0,0.2,1) via `animation.CubicBezier`) need their own `Layout`.
  - `toggle` — Toggle/ToggleGroup: `button.PaintState` has no on-state, so Toggle is a small new widget reusing Button metrics; ToggleGroup adds the group controller and fused-corner compositing (§5.4).
  - `menu` — **the shared shadcn menu engine**: an overlay list widget whose rows are arbitrary graft-built row widgets (item/checkbox-item/radio-item/label/separator/sub-trigger with icons, shortcuts, inset, destructive). Used by DropdownMenu, ContextMenu, and Menubar sub-menus that exceed `core/menu.MenuItem`'s fixed model. Positioning via `core/popover.CalculatePosition` (12 placements) + `ctx.OverlayManager()` recipe (Report 2 §4.4). Decision: Select keeps `core/dropdown` (its item anatomy — check on the right, pl-8 — differs from menu anatomy and core/dropdown already handles keyboard/scroll); the menu engine serves the menu family only.
  - `sheet` — edge-anchored panel: `overlay.NewContainer(panel, ..., WithModal(true))` + `transition.Wrap(SlideIn(FromRight/...))`, 500ms in / 300ms out.
  - `toastregion` — Sonner: ONE persistent corner overlay managing its own toast children (LIFO stack removal gotcha, Report 2 §4.5); `graft.Toast(...)` is the imperative API; `graft.Toaster()` mounts the region.
  - `calendar` — month grid, 32px `--cell-size`, ghost nav buttons, selected-day primary pill, range states.
  - `textarea` — **Textarea is a new widget** (deviation from Report 4 Tier 1): `core/textfield` is single-line (single `CursorPos`, no line model). Scope v1: plain text, soft wrap, vertical scroll, line-based cursor/selection, `field-sizing: content` auto-grow with min-h 64px. Scheduled Phase 2.
  - `otp` — InputOTP segmented input (focus handoff, paste splitting).
  - `command` — cmdk: filter input + listview + keyboard nav + group headers; `CommandDialog` composes with Dialog; global shortcut via `Window.FocusManager().RegisterShortcut`.
  - `carousel` — scrollview + snap + prev/next (Phase 3).

**Omitted (D), final:** NativeSelect (collapses into Select), Drawer (Sheet side=bottom), legacy Toast (Sonner only), NavigationMenu (Menubar/Sidebar/Tabs are the desktop idioms), Direction/RTL (deferred to gogpu/ui text layer), react-hook-form Form (replaced by signal-backed `graft.Form`, §4).

---

## 4. Component API convention (final)

**Adopted:** Report 4's proposal — one flat package, shadcn names 1:1, content-first constructors, chainable prop methods, typed enums with sugar, signals for controlled state — with these refinements:

1. **Type naming:** constructor is the shadcn name (`Button`), the widget type is `*<Name>Widget` (`*ButtonWidget`), matching gogpu/ui's `BoxWidget/TextWidget` convention (replaces Report 4's `ButtonW`).
2. **`graft.Widget`** is `type Widget = widget.Widget` (alias) — graft trees mix freely with raw `primitives.*` and any gogpu/ui widget; simple apps import only graft.
3. **Controlled vs uncontrolled:** `Value(v)` = initial value (shadcn `defaultValue`); `Bind(sig state.Signal[T])` = controlled (read for render, written on interaction — replaces the React `value`+`onValueChange` pair). Both may coexist with `OnChange` (observer). Bindings are registered in `Mount` via `state.BindToScheduler` + `AddBinding` (Report 2 §6) — never raw subscriptions in constructors.
4. **Event names:** `OnClick(func())`, `OnChange(func(T))` (covers shadcn's `onCheckedChange`/`onValueChange`), `OnOpenChange(func(bool))`, `OnSelect(func())` (menu items), `OnSubmit` (form/input). Handlers run on the UI thread.
5. **Enums + sugar:**
```go
type Variant uint8
const (VariantDefault Variant = iota; VariantSecondary; VariantDestructive; VariantOutline; VariantGhost; VariantLink)
type Size uint8
const (SizeDefault Size = iota; SizeXS; SizeSM; SizeLG; SizeIcon; SizeIconXS; SizeIconSM; SizeIconLG)
```
6. **Escape hatch:** every builder has `.Style(func(*graft.Style))`; `Style` v1 fields: `Background, Foreground *widget.Color; Radius *float32; PadX, PadY *float32; MinWidth, MaxWidth *float32` — flows into per-widget PaintState overrides (`Background *widget.Color`, `Radius *float32`) where the core widget supports them.
7. **Layout sugar:** `.W(px)`, `.Full()` on anything sized.
8. **Theme resolution:** constructors snapshot `graft.CurrentTheme()` (process-global set by `Install`, mutex-guarded) and pass the matching painter from the theme's shared `Painters` bundle via `xxx.PainterOpt(...)` at `New` time (painter cannot be set later — Report 1 §7.1). Optional `.Theme(th)` per-widget override for multi-theme windows.
9. **Overlay open state is a signal**; `XxxTrigger(w, open)` sugar wires click→`open.Set(true)` for the literal shadcn shape.
10. Compile-time checks per component: `var (_ widget.Widget = (*ButtonWidget)(nil); _ widget.Focusable = ...; _ widget.Lifecycle = ...; _ a11y.Accessible = ...)` with the right `a11y.Role`.

### Final signatures

```go
// ── Button (button.go) ─────────────────────────────────────────────
func Button(label string, children ...Widget) *ButtonWidget
func (b *ButtonWidget) Variant(v Variant) *ButtonWidget
func (b *ButtonWidget) Secondary() *ButtonWidget   // sugar for each Variant
func (b *ButtonWidget) Destructive() *ButtonWidget
func (b *ButtonWidget) Outline() *ButtonWidget
func (b *ButtonWidget) Ghost() *ButtonWidget
func (b *ButtonWidget) Link() *ButtonWidget
func (b *ButtonWidget) Size(s Size) *ButtonWidget
func (b *ButtonWidget) XS() *ButtonWidget; func (b *ButtonWidget) Sm() *ButtonWidget; func (b *ButtonWidget) Lg() *ButtonWidget
func (b *ButtonWidget) Icon(ic icon.IconData) *ButtonWidget      // leading icon, 16px (12px at XS); switches px to has->svg value
func (b *ButtonWidget) IconOnly(ic icon.IconData) *ButtonWidget  // square icon sizes (36/24/32/40)
func (b *ButtonWidget) OnClick(fn func()) *ButtonWidget
func (b *ButtonWidget) Disabled(v bool) *ButtonWidget
func (b *ButtonWidget) BindDisabled(sig state.Signal[bool]) *ButtonWidget
func (b *ButtonWidget) BindLabel(sig state.Signal[string]) *ButtonWidget
func (b *ButtonWidget) Loading(sig state.Signal[bool]) *ButtonWidget // swaps in Spinner, disables
func (b *ButtonWidget) Full() *ButtonWidget
func (b *ButtonWidget) W(px float32) *ButtonWidget
func (b *ButtonWidget) Style(fn func(*Style)) *ButtonWidget

// ── Card (card.go) ─────────────────────────────────────────────────
func Card(children ...Widget) *CardWidget                 // radius 14, 1px border, py 24, gap 24, shadow-sm
func CardHeader(children ...Widget) *CardSectionWidget    // px 24, gap 8; lays CardAction top-right
func CardTitle(text string) *TypographyWidget             // 16px / weight 600 / leading-none
func CardDescription(text string) *TypographyWidget       // 14px muted-foreground
func CardAction(child Widget) *CardActionWidget
func CardContent(children ...Widget) *CardSectionWidget   // px 24
func CardFooter(children ...Widget) *CardSectionWidget    // px 24, horizontal row
func (c *CardWidget) W(px float32) *CardWidget
func (c *CardWidget) Style(fn func(*Style)) *CardWidget

// ── Input (input.go) ───────────────────────────────────────────────
func Input() *InputWidget
func (i *InputWidget) Bind(sig state.Signal[string]) *InputWidget
func (i *InputWidget) Value(s string) *InputWidget
func (i *InputWidget) Placeholder(s string) *InputWidget
func (i *InputWidget) Password() *InputWidget
func (i *InputWidget) Disabled(v bool) *InputWidget
func (i *InputWidget) Invalid(v bool) *InputWidget                       // aria-invalid ring/border
func (i *InputWidget) BindInvalid(sig state.ReadonlySignal[bool]) *InputWidget // Form wires this
func (i *InputWidget) OnChange(fn func(string)) *InputWidget
func (i *InputWidget) OnSubmit(fn func(string)) *InputWidget             // Enter
func (i *InputWidget) W(px float32) *InputWidget

// ── Checkbox (checkbox.go) ─────────────────────────────────────────
func Checkbox() *CheckboxWidget
func (c *CheckboxWidget) Label(text string) *CheckboxWidget
func (c *CheckboxWidget) Bind(sig state.Signal[bool]) *CheckboxWidget
func (c *CheckboxWidget) Checked(v bool) *CheckboxWidget
func (c *CheckboxWidget) Indeterminate(sig state.ReadonlySignal[bool]) *CheckboxWidget
func (c *CheckboxWidget) OnChange(fn func(bool)) *CheckboxWidget
func (c *CheckboxWidget) Disabled(v bool) *CheckboxWidget

// ── Switch (switch.go → internal/widgets/switchw) ──────────────────
func Switch() *SwitchWidget
func (s *SwitchWidget) Sm() *SwitchWidget                  // 24×14 track / 12px thumb
func (s *SwitchWidget) Bind(sig state.Signal[bool]) *SwitchWidget
func (s *SwitchWidget) Checked(v bool) *SwitchWidget
func (s *SwitchWidget) OnChange(fn func(bool)) *SwitchWidget
func (s *SwitchWidget) Disabled(v bool) *SwitchWidget

// ── Badge (badge.go) ───────────────────────────────────────────────
func Badge(text string, children ...Widget) *BadgeWidget   // pill, px 8 / py 2, 12px/500
func (b *BadgeWidget) Variant(v Variant) *BadgeWidget      // + Secondary/Destructive/Outline/Ghost/Link sugar
func (b *BadgeWidget) Icon(ic icon.IconData) *BadgeWidget  // 12px
func (b *BadgeWidget) OnClick(fn func()) *BadgeWidget      // enables [a&] hover states + pointer cursor

// ── Dialog (dialog.go) ─────────────────────────────────────────────
func Dialog(content *DialogContentWidget) *DialogWidget    // zero-size host; watches open signal, pushes overlay
func (d *DialogWidget) Bind(open state.Signal[bool]) *DialogWidget
func (d *DialogWidget) OnOpenChange(fn func(bool)) *DialogWidget
func DialogTrigger(trigger Widget, open state.Signal[bool]) Widget   // click → open.Set(true)
func DialogContent(children ...Widget) *DialogContentWidget          // max-w 512, p 24, gap 16, radius 10, shadow-lg, X button
func (c *DialogContentWidget) HideClose() *DialogContentWidget
func DialogHeader(children ...Widget) *DialogSectionWidget           // column, gap 8
func DialogTitle(text string) *TypographyWidget                      // 18px / 600 / leading-none
func DialogDescription(text string) *TypographyWidget                // 14px muted
func DialogFooter(children ...Widget) *DialogSectionWidget           // row, gap 8, right-aligned
func AlertDialog(content *DialogContentWidget) *DialogWidget         // modal: no X, no outside-dismiss; Esc closes

// ── Select (select.go) ─────────────────────────────────────────────
func Select(entries ...SelectEntry) *SelectWidget
func SelectItem(value, label string) *SelectItemEntry      // implements SelectEntry; .Disabled(bool)
func SelectGroup(label string, items ...SelectEntry) SelectEntry
func SelectSeparator() SelectEntry
func (s *SelectWidget) Bind(sig state.Signal[string]) *SelectWidget
func (s *SelectWidget) Value(v string) *SelectWidget
func (s *SelectWidget) Placeholder(text string) *SelectWidget
func (s *SelectWidget) OnChange(fn func(string)) *SelectWidget
func (s *SelectWidget) Sm() *SelectWidget                  // h 32
func (s *SelectWidget) Disabled(v bool) *SelectWidget
func (s *SelectWidget) Invalid(v bool) *SelectWidget
func (s *SelectWidget) W(px float32) *SelectWidget

// ── Tabs (tabs.go) ─────────────────────────────────────────────────
func Tabs(children ...Widget) *TabsWidget                  // expects TabsList + TabsContent(s)
func TabsList(triggers ...*TabsTriggerWidget) *TabsListWidget
func TabsTrigger(value, label string) *TabsTriggerWidget   // .Disabled(bool), .Icon(ic)
func TabsContent(value string, content Widget) *TabsContentWidget
func (t *TabsWidget) Bind(sig state.Signal[string]) *TabsWidget
func (t *TabsWidget) Value(v string) *TabsWidget           // uncontrolled initial
func (t *TabsWidget) Line() *TabsWidget                    // line variant (transparent list, 2px underline)

// ── DropdownMenu (dropdownmenu.go → internal/widgets/menu) ─────────
func DropdownMenu(children ...Widget) *DropdownMenuWidget  // one Trigger + one Content
func DropdownMenuTrigger(w Widget) *DropdownMenuTriggerWidget
func DropdownMenuContent(entries ...MenuEntry) *DropdownMenuContentWidget
func DropdownMenuItem(label string) *MenuItemEntry
func (e *MenuItemEntry) Icon(ic icon.IconData) *MenuItemEntry
func (e *MenuItemEntry) Shortcut(s string) *MenuItemEntry
func (e *MenuItemEntry) OnSelect(fn func()) *MenuItemEntry
func (e *MenuItemEntry) Destructive() *MenuItemEntry
func (e *MenuItemEntry) Disabled(v bool) *MenuItemEntry
func (e *MenuItemEntry) Inset() *MenuItemEntry             // pl 32
func DropdownMenuCheckboxItem(label string) *MenuCheckboxEntry  // .Bind(state.Signal[bool]), .OnChange
func DropdownMenuRadioGroup(sig state.Signal[string], items ...*MenuRadioEntry) MenuEntry
func DropdownMenuRadioItem(value, label string) *MenuRadioEntry
func DropdownMenuLabel(text string) MenuEntry
func DropdownMenuSeparator() MenuEntry
func DropdownMenuGroup(entries ...MenuEntry) MenuEntry
func DropdownMenuSub(label string, entries ...MenuEntry) MenuEntry
func (m *DropdownMenuWidget) Bind(open state.Signal[bool]) *DropdownMenuWidget
func (m *DropdownMenuWidget) OnOpenChange(fn func(bool)) *DropdownMenuWidget
```

**Form** (Go replacement for react-hook-form, per Report 4 §3.7): `graft.NewForm()`, `graft.FormValue(form, name, initial, validators...)` (validators `Required()`, `MinLen(n)`, `Email()`, `Custom(func(T) error)`), `graft.Form(form, children...).OnSubmit(fn)`, `graft.FieldError(errsSignal)`; `Button(...).Submit()` triggers `form.Validate()`; FormValue auto-wires `BindInvalid` on its control.

---

## 5. Pixel-fidelity strategy

### 5.1 Metrics package — Report 3 tables become Go constants

One file per component in `graft/metrics`, each constant annotated with its source class string. Example shape (every component follows it):

```go
package metrics

// Source: shadcn new-york-v4 buttonVariants cva (Report 3 §5).
type ButtonSize struct {
    Height, PadX, PadXWithIcon, PadY, Gap, FontSize, IconSize float32
    RadiusFn func(t *theme.Theme) float32 // → t.RadiusMD() (8px @ default) — radius routes through theme
}
var Button = struct {
    Default, XS, SM, LG, Icon, IconXS, IconSM, IconLG ButtonSize
    FontWeight int // 500 → family "Geist Medium" (§5.5)
}{
    Default: ButtonSize{Height: 36, PadX: 16, PadXWithIcon: 12, PadY: 8, Gap: 8, FontSize: 14, IconSize: 16, ...},
    XS:      ButtonSize{Height: 24, PadX: 8,  PadXWithIcon: 6,  Gap: 4, FontSize: 12, IconSize: 12, ...},
    SM:      ButtonSize{Height: 32, PadX: 12, PadXWithIcon: 10, Gap: 6, FontSize: 14, IconSize: 16, ...},
    LG:      ButtonSize{Height: 40, PadX: 24, PadXWithIcon: 16, Gap: 8, FontSize: 14, IconSize: 16, ...},
    // Icon: 36×36, IconXS: 24×24 (12px icon), IconSM: 32×32, IconLG: 40×40
}
```

Same pattern for: Input (h36/px12/py4/14px/radius MD), Checkbox (16×16, radius 4 literal, check 14px), RadioGroup (16 circle, 8 dot, group gap 12), Switch (32×18.4 / thumb 16 / travel 14; sm 24×14 / 12 / 10), Card (radius XL=14, py24/px24/gap24/8), Badge (px8/py2/12px/pill), Alert (radius LG, px16/py12, icon 16 +2px Y, col gap 12, row gap 2), Dialog (maxW 512, p24, gap16, radius LG, close 16px@(16,16) corner), Select (trigger h36/sm32, px12, content minW 128, viewport p4, item py6/pl8/pr32/radius SM, check 16 right 8), Menu (content minW128/p4/radius MD, item px8/py6/radius SM/gap8, inset pl32, shortcut 12px +0.1em, separator -mx4/my4), Tabs (list h36/radius LG/pad 3, trigger radius MD/px8/py4, line underline 2px at −5px), Tooltip (px12/py6/radius MD/12px, fg-on-bg inverted, arrow 10px@45° r2), Popover (w288/p16/radius MD), Slider (track 6, thumb 16 white + 1px primary border, hover/focus ring 4), Progress (h8 pill, track primary@20%), Table (head h40/px8/500, cell p8, 1px row borders), Avatar (24/32/40), Skeleton (accent fill, radius MD, pulse 2s @50% opacity), Kbd, Accordion (trigger py16, chevron 16 rotating 180° 200ms, content pb16), Sheet (maxW 384, p16/gap6), ScrollArea (gutter 10, thumb pill border-color), Calendar (cell 32), Command. Shared: `metrics/focus.go` (RingWidth 3, RingAlpha 0.5, InvalidAlphaLight 0.2 / Dark 0.4, LegacyCloseRing{W:2, Offset:2}, SliderRing 4), `metrics/shadow.go` (§5.3).

### 5.2 Focus ring

shadcn recipe: `focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50` — a 3px no-offset box-shadow ring **outside** the border box, plus the element border turning solid `--ring`.

`internal/draw`:
```go
// Ring band occupies [bounds, bounds+3]. Strokes are center-drawn (Report 1 §4),
// so stroke width 3 at Expand(1.5) with radius+1.5.
func FocusRing(c widget.Canvas, bounds geometry.Rect, radius float32, ring widget.Color) {
    c.StrokeRoundRect(bounds.Expand(1.5), ring, radius+1.5, 3)
}
```
Painters, when `state.Focused`: (1) draw `FocusRing(bounds, r, Alpha(t.Ring, 0.5))`; (2) draw the element's 1px border in **solid** `t.Ring` instead of `t.Border/t.Input`. Invalid state: same with `t.Destructive` at 0.2 (dark 0.4) + solid destructive border. Slider thumb: ring width 4, also on hover. Dialog/Sheet close button keeps the legacy `ring-2 ring-offset-2` recipe: stroke 2 at `Expand(3)` plus a 2px stroke in `t.Background` at `Expand(1)` to fill the offset gap (Report 1 §7.8). Focus only renders for keyboard focus: widgets track whether focus arrived via Tab (focus-visible semantics); mouse-press focus does not draw the ring.

Clipping caveat: the ring paints outside widget bounds; container composites built by graft (Card sections, Field rows) reserve no extra space but also do not clip (Box only clips when it has border/radius — Report 2 §2); inside ScrollArea the ring may clip at the viewport edge — accepted, identical to browser behavior with `overflow: hidden`.

### 5.3 Hover/press color math

`internal/draw` helpers:
```go
func Alpha(c widget.Color, a float32) widget.Color    // REPLACE alpha — for opaque tokens (primary/90)
func MulAlpha(c widget.Color, f float32) widget.Color // MULTIPLY — for tokens that already carry alpha (dark border/input)
func Fade(c widget.Color, disabled bool) widget.Color // disabled ⇒ MulAlpha(c, 0.5)
```
CSS `bg-primary/90` over the page background is plain alpha compositing — drawing `Alpha(Primary, 0.9)` over the already-painted background reproduces it exactly. Per-state table (from Report 3 §7) is encoded as constants next to each painter: primary hover .90, secondary hover .80, destructive hover .90 (dark base `Alpha(Destructive,.60)`), outline/ghost hover = solid Accent (dark ghost `Alpha(Accent,.50)`, dark outline `MulAlpha(Input,.50)`), item focus = solid Accent, destructive item focus `Alpha(Destructive,.10)` (dark .20), table row hover `Alpha(Muted,.50)`, toggle hover solid Muted / on solid Accent.

**Press state (decision):** the registry defines no `:active` styles — graft renders pressed identical to hovered. Tradeoff: less tactile than typical desktop UI; faithful to shadcn. An opt-in `graft.PressFeedback(true)` global (98%-scale-free 60% opacity dim, matching the docs-site mobile rule) may come later; not in v1.
**Disabled (decision):** no canvas-wide opacity layer exists on the base Canvas, so painters `Fade` every color (fills, borders, text, icons) by 0.5. Visually equivalent for shadcn's flat fills; the only divergence is overlapping-element compositing, which shadcn controls don't have. Transitions are instant (no 150ms color tween) — accepted for v1; hover tweens need widget-driven animation and are not worth the complexity (Report 1 §6).

### 5.4 Borders, radii, corners

- **Inside borders:** CSS borders sit inside the box. `draw.InsideBorder(c, bounds, radius, color, w)` = `StrokeRoundRect(bounds inset by w/2, color, radius−w/2, w)` (the M3 checkbox trick, Report 1 §7.3). All 1px component borders use it.
- **Per-corner radii do not exist** on Canvas. Affected: ToggleGroup/ButtonGroup fused segments (first/last rounded outer corners only). `draw.SquareCorners(canvas, bounds, radius, fill, corners)` draws the round rect then overpaints chosen corner quadrants with plain rects of the same fill before the border pass. Items share 1px borders by overlapping −1px and suppressing the left border on non-first items.
- `Full` radius = 9999 (pill) — clamps fine.

### 5.5 Shadows

No blur primitive (Report 1 §7.4) — each shadcn shadow level is approximated with stacked low-alpha round-rects drawn **before** the element fill. `metrics/shadow.go`:

```go
type ShadowLayer struct{ DY, Grow, Alpha float32 } // Grow = Expand() amount; color always black
var ShadowXS = []ShadowLayer{{1, 0, .04}, {1, 1, .03}}                      // 0 1px 2px 0 / 5%
var ShadowSM = []ShadowLayer{{1, 0, .06}, {1, 1, .05}, {2, 2, .03}}         // sm
var ShadowMD = []ShadowLayer{{2, 0, .06}, {4, 2, .05}, {6, 4, .03}}         // md
var ShadowLG = []ShadowLayer{{4, 2, .05}, {8, 5, .04}, {12, 9, .03}}        // lg
```
`draw.Shadow(canvas, bounds, radius, layers)` translates each layer to `DrawRoundRect(bounds.Expand(Grow).TranslateXY(0, DY), RGBA(0,0,0,Alpha), radius+Grow)`. These starting values are tuned once against the kitchen-sink example during Phase 0, then **frozen** — goldens lock them. Alphas stay ≤ 0.08 to avoid hard edges. Usage map (Report 3 §4): XS → outline button, input, textarea, checkbox, radio, switch, select trigger; SM → card, slider thumb, active tab; MD → popover, select/dropdown content; LG → dialog, sheet, sub-menus. Tradeoff: not Gaussian-identical to CSS; accepted until gogpu grows a blur primitive (tracked in §8).

### 5.6 Text: fonts, weights, measurement

- **Geist** (OFL) embedded in `graft/fonts`. The public canvas only resolves Regular/Bold per family (Report 1 §7.5), so weights register as **separate families** via `plugin` asset loader: `"Geist"` (400), `"Geist Medium"` (500), `"Geist SemiBold"` (600), `"Geist Bold"` (700, also usable as Bold of "Geist" where the bool path suffices), plus `"Geist Mono"`. `fonts.Family(weight int) string` maps 400/500/600/700 → family names. `graft.Install` calls `fonts.Load()`.
- All graft text draws through the `StyledTextDrawer` capability (`widget.TextStyle{FontFamily, FontSize, Color, Align}`); if the canvas lacks it (mock canvas), fall back to `DrawText` with `bold = weight ≥ 600`.
- **Layout-time measurement (decision):** gogpu's heuristic (`len(text)·fontSize·0.55`) is too crude for pixel-faithful widths. `internal/textmetrics` measures advances from the embedded Geist TTFs via `golang.org/x/image/font/sfnt` at layout time. For wrapped core widgets (button), graft computes the exact desired width (text advance + paddings + icon + gap) and pins it with `MinWidth(v)`+`MaxWidth(v)`. For graft-owned widgets, `Layout` uses textmetrics directly.
- Heights: shadcn's `h-9` button = 36px while `button.Size` offers 32/40/48 — graft uses `SizeOpt(button.Small)` + `PaddingXY(padX, 18)` so `max(32, 36) = 36` (Report 1 §7.6); each size has its PaddingXY pair derived from `metrics.Button`.
- Typography tokens on `uitheme.Typography` are advisory and unread by painters — graft ignores them and drives sizes from `metrics`.

### 5.7 Icons

Lucide via `icon.FromSVGXML` (Route B — honors stroke/linecap, pixel-perfect; Report 2 §9). `graft/icons` vendors the ~45 SVGs the components need (check, chevron-{up,down,left,right}, chevrons-up-down, x, circle, search, more-horizontal, loader-2, calendar, minus, panel-left, circle-check, info, triangle-alert, octagon-x, ...) with `//go:embed`, registers them in `icon.DefaultRegistry()` from `graft.Install`, and exposes `icons.Check`, `icons.ChevronDown`, ... as `icon.IconData`. `icons/gen` is a `go run` script that vendors additional lucide icons by name.

---

## 6. Verification strategy

### 6.1 Golden-image tests (primary)

Helper in `internal/gtest` (test-only):

```go
func Golden(t *testing.T, name string, build func() widget.Widget) {
    th := theme.New() // Neutral default; dark cases use a th with mode forced dark
    r := offscreen.NewRenderer(0, 0,
        offscreen.WithFitSize(), offscreen.WithMaxSize(800, 600),
        offscreen.WithTheme(th),                       // th implements widget.ThemeProvider
        offscreen.WithBackground(th.Active().Background),
        offscreen.WithScale(2),                        // 2x catches sub-pixel border/ring errors
    )
    r.Render(build())
    comparePNG(t, "testdata/golden/"+name+".png", r.Image()) // -update flag regenerates
}
```

- Cases per component: each variant × size × {default, hovered, focused, disabled, invalid, checked/open where applicable} × {light, dark}. Hover/press/focus states are produced by driving `uitest` events (`MouseEnter`, `SimulateClick`, `SetFocused(true)` / synthesized FocusGained) before rendering, or via test-only state setters on graft widgets.
- Comparison: **exact byte match** expected — rendering is pure-CPU gg with embedded fonts, hence deterministic. If cross-platform float drift appears, fall back to per-channel tolerance ≤1 with a max-diff-pixel budget of 0.1% (env `GRAFT_GOLDEN_TOLERANCE=1`). CI renders on linux; Windows devs run the same tests and report drift as a bug first, tolerance second.
- Overlay components (Dialog, Tooltip, menus): render the overlay content widget directly at its measured size (the overlay chassis is positioned logic, tested separately).

### 6.2 Paint-call spec tests (cheap, exact)

`uitest.DrawWidget` MockCanvas assertions verify the *numbers* independent of rasterization: e.g. Button focused ⇒ exactly one `StrokeRoundRectCall{StrokeWidth: 3, Radius: r+1.5, Bounds: bounds.Expand(1.5), Color: ring@0.5}`; Checkbox box is 16×16 at radius 4; Card border color equals `Tokens.Border`. Every painter gets one spec test per state; these run in milliseconds and pin the metrics tables to the painters.

### 6.3 Conversion + parser tests

`theme/oklch_test.go` against §2.2 vectors; `theme/css_test.go` round-trips the five embedded presets and a tweakcn export sample. `metrics` has no logic — its values are pinned by 6.2.

### 6.4 Eyeball harness

`examples/kitchensink` renders every component in every state in a scrollable window (and `go run ./examples/kitchensink -png` dumps the sheet via offscreen) for side-by-side comparison with ui.shadcn.com during development. This is the tool used to tune shadow layers once in Phase 0/1.

---

## 7. Implementation phases

Rules: a batch = a set of components whose files are disjoint; batches marked ∥ can be assigned to independent agents concurrently. Shared files (`graft.go`, `theme/*`, `metrics/focus.go`, `metrics/shadow.go`, `internal/draw/*`, `painters/painters.go` bundle struct) are **frozen after Phase 0** — painter/component files only add new files plus one line to the bundle struct (coordinate bundle-struct edits via tiny sequential PRs or pre-declare all fields in Phase 0; **decision: pre-declare every planned painter field in Phase 0** so no shared-file edits are ever needed).

### Phase 0 — Foundation (sequential, one agent; everything else depends on it)
1. `go.mod`, CI (golden test job, lint), repo scaffolding.
2. `theme/`: Tokens, Theme, Mode, oklch (+tests), radius scale, ParseThemeCSS (+tests), presets (embedded CSS), AsUITheme + registry registration.
3. `fonts/` (Geist embed + family registration), `icons/` (lucide subset + registry init + gen script).
4. `internal/draw/` (FocusRing, InsideBorder, Shadow, Alpha/MulAlpha/Fade, styled-text helper), `internal/textmetrics/`.
5. `metrics/` shared files (focus, shadow) + skeletons.
6. `painters/painters.go` with the **complete** `Painters` struct (all 17 fields declared, nil-tolerant `New`).
7. `graft.go`: Widget alias, Variant/Size, Style, Install/CurrentTheme.
8. `internal/gtest` golden harness; first golden: a token swatch sheet (validates oklch + theme end-to-end).

### Phase 1 — Tier 1 (4 parallel batches ∥)
Each item = component file + its metrics file (+ painter file if Class A) + goldens + spec tests.

| Batch | Components (dependency order within batch) | Notes |
|---|---|---|
| **1A ∥ Form controls** | Button(+painter) → Label → Input(+textfield painter) → Checkbox(+painter) → RadioGroup(+radio painter) → Switch (new widget) → Field/FieldGroup/FieldSet → Form | Form last (uses Field + Input invalid wiring) |
| **1B ∥ Static surfaces** | Typography set → Separator → Badge → Alert → Card family → Spinner → Progress(+progressbar painter) | No cross-batch deps |
| **1C ∥ Overlays** | ScrollArea(+scrollbar painter) → Popover(+popover painter) → Tooltip → Dialog(+dialog painter) → AlertDialog | DialogClose draws its own internal icon button (no dependency on 1A Button — intentional, keeps batches disjoint) |
| **1D ∥ Selection & menus** | Slider(+painter) → Tabs(+tabview painter) → Select(+dropdown painter) → internal/widgets/menu engine → DropdownMenu | Menu engine is the big rock; placed with its only Phase-1 consumer |

Phase 1 exit: login-card and settings examples compile and match references visually; all goldens green.

### Phase 2 — Tier 2 (4 parallel batches ∥; requires Phase 1 merged)

| Batch | Components | Cross-batch deps |
|---|---|---|
| **2A ∥ Disclosure & display** | Collapsible(+painter) → Accordion → Avatar(+Group) → Skeleton → Toggle → ToggleGroup (fused corners §5.4) | none |
| **2B ∥ Composition rows** | ButtonGroup → InputGroup → Item family → Table (static) → Textarea (new widget — biggest item here) | uses 1A Input conventions only |
| **2C ∥ App chrome** | Sheet → Resizable(+splitview painter) → Menubar(+menu painter) → ContextMenu (reuses 1D menu engine) → Sidebar | Sidebar last (uses Sheet for mobile-style collapse, Separator, Tooltip) |
| **2D ∥ Data & time** | Toast/Sonner (toastregion) → DataTable(+datatable painter) → Calendar (new widget) → DatePicker (Popover+Calendar+Button) → Combobox (Popover + listview painter + filter input) | DatePicker after Calendar; Combobox after listview painter (move listview painter here, owned by 2D) |

### Phase 3 — Tier 3 (parallel, any order ∥)
Command (+CommandDialog; uses listview painter, Dialog, FocusManager shortcut), Breadcrumb, Pagination (reuses Button), Empty, Kbd, HoverCard (popover hover trigger + delay), InputOTP, Carousel, Chart (+linechart painter), AspectRatio.

**Permanently omitted:** NativeSelect, Drawer, legacy Toast, NavigationMenu, Direction (§3.2 rationale).

---

## 8. Risks and decisions (all decided)

1. **No global painter installation** (Report 1 §7.1). *Decision:* graft constructors always inject painters from `CurrentTheme()`; raw gogpu/ui users get `painters.New(th)` and wire `PainterOpt` themselves (documented). Process-global current theme with `.Theme(th)` per-widget override. Tradeoff: global mutable state; accepted — it is the only way to keep constructor calls clean, and it is snapshot-at-construction so behavior is predictable.
2. **CSS box-shadow cannot be replicated without blur.** *Decision:* fixed layered approximation (§5.5), tuned once, frozen by goldens. Tradeoff: visible difference under close inspection vs browser; revisit when gogpu/gg exposes blur — layer tables are isolated in `metrics/shadow.go` so the swap is one file.
3. **font-medium 500 / semibold 600 not addressable by weight.** *Decision:* separate-family registration ("Geist Medium", "Geist SemiBold") via the public plugin loader. Tradeoff: family-name hack, mock-canvas fallback degrades 500→400 and 600→700 in spec tests only.
4. **Text width heuristics break pixel widths.** *Decision:* own sfnt measurement (`internal/textmetrics`) + `MinWidth/MaxWidth` pinning on wrapped core widgets. Tradeoff: graft's measurement must agree with gg's advances — same TTF bytes, asserted by a spec test comparing `MeasureStyledText` vs textmetrics within 0.5px.
5. **Switch/Toggle/Textarea misclassified as Class A in Report 4.** *Decision:* all three are new widgets (§3.2): painters can't change layout, button has no on-state, textfield is single-line. Textarea slips to Phase 2 with a minimal plain-text editor scope. Tradeoff: Tier 1 ships Input-only text entry; acceptable.
6. **DropdownMenu anatomy exceeds `core/menu`'s item model** (icons, checkbox/radio items, destructive, inset). *Decision:* custom menu engine (`internal/widgets/menu`) for the menu family; Select stays on `core/dropdown`; Menubar bar chrome on `core/menu` with graft's engine for its drop-downs if `MenuItem` proves insufficient (engine is reusable by design). Tradeoff: two menu paint paths to keep visually consistent — mitigated by both reading the same `metrics/menu.go`.
7. **Per-corner radii unavailable.** *Decision:* corner overpaint compositing for ToggleGroup/ButtonGroup only (§5.4); everything else in shadcn is uniform-radius. Tradeoff: compositing trick breaks if a translucent fill sits on a busy background — group fills are opaque token colors, so safe.
8. **Press/active feedback absent in spec.** *Decision:* pressed == hovered, no extra styling (faithful). Tradeoff noted in §5.3; revisit only on user demand.
9. **Disabled = 50% opacity without an opacity layer.** *Decision:* per-color `Fade` multiplication in painters. Equivalent for flat shadcn surfaces.
10. **OKLCH out-of-gamut chart/destructive colors.** *Decision (revised, see §2.2):* per-linear-channel clipping — empirically matches Tailwind v4 published fallbacks and Chrome sRGB rendering exactly; chroma-reduction was tested and rejected (produces wrong pixels for `--destructive`).
11. **Focus ring paints outside bounds and can clip.** *Decision:* accept (matches `overflow:hidden` browser behavior); dense graft layouts (menu items) use inset rings? No — menu items use background-fill focus (per spec, `focus:bg-accent`), so the issue only arises in user layouts; documented.
12. **Theme switching with baked painters.** *Decision:* painters hold `*Theme` and read `Active()` per draw; `Install`'s mode subscription calls `app.SetTheme(AsUITheme())` which forces full repaint — no tree rebuild ever needed (verified by a golden test that flips mode between two renders of the same tree).
13. **Zero-value ColorScheme sentinel** (Report 1 §7.7). *Decision:* graft never passes widget-level ColorSchemes; painters resolve from theme and honor only `Background`/`Radius` pointer overrides. Spec tests assert painters ignore zero schemes.
14. **Toast stacking vs LIFO overlay removal.** *Decision:* single persistent toast-region overlay managing its own children (Report 2 §4.5).
15. **Golden determinism across platforms.** *Decision:* exact-match goldens generated on CI linux; tolerance escape hatch (≤1/channel, 0.1% pixels) behind an env var; Windows drift treated as a bug before loosening.
16. **`destructive-foreground` removed from canonical theme.** *Decision:* keep the token (default white in both modes) because theme-editor CSS imports may define it and destructive surfaces need a text color; painters use it exactly where shadcn writes `text-white`.

---

### Appendix: app skeleton (documentation target)

```go
func main() {
    th := graft.NewTheme(graft.BaseColor(graft.Zinc), graft.Radius(10))
    mode := state.NewSignal(graft.ModeSystem)

    gpuApp := gogpu.NewApp(gogpu.DefaultConfig().WithTitle("App").WithSize(960, 640).WithContinuousRender(false))
    uiApp := app.New(
        app.WithWindowProvider(gpuApp), app.WithPlatformProvider(gpuApp),
        app.WithEventSource(gpuApp.EventSource()),
        app.WithTheme(th.AsUITheme()),
    )
    graft.Install(uiApp, th, mode) // fonts + icons + painters + mode subscription

    uiApp.SetRoot(buildUI())
    log.Fatal(desktop.Run(gpuApp, uiApp))
}
```

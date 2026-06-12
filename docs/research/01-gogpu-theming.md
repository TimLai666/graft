# gogpu/ui Theming & Painter System — Research Report for a "shadcn" Theme

Source: local clone at `C:\Users\tingzhen\AppData\Local\Temp\gogpu-ui` (module `github.com/gogpu/ui`). All signatures below are copied verbatim from source.

---

## 1. The `theme.Theme` struct and DefaultLight/DefaultDark

### 1.1 `theme.Theme` (theme/theme.go)

```go
type Theme struct {
    Name       string            // human-readable, e.g. "Light"
    Mode       ThemeMode         // ModeLight | ModeDark | ModeSystem (uint8)
    Colors     ColorPalette
    Typography Typography
    Spacing    SpacingScale
    Shadows    ShadowStyles
    Radii      RadiusScale
    Extensions map[string]any    // untyped extension bag, keyed by package path
    typedExts  *typedExtensions  // unexported; access via RegisterExtension/TypedExtension
}
```

Constructors/methods (all on `*Theme`):
```go
func New(name string, mode ThemeMode) *Theme  // fills defaults, Colors left empty for caller
func (t *Theme) Clone() *Theme                 // deep copy; Extensions shallow-copied
func (t *Theme) WithName(name string) *Theme
func (t *Theme) WithMode(mode ThemeMode) *Theme
func (t *Theme) WithColors(colors *ColorPalette) *Theme
func (t *Theme) WithTypography(typography *Typography) *Theme
func (t *Theme) WithSpacing(spacing SpacingScale) *Theme
func (t *Theme) WithShadows(shadows *ShadowStyles) *Theme
func (t *Theme) WithRadii(radii RadiusScale) *Theme
func (t *Theme) SetExtension(key string, value any)
func (t *Theme) GetExtension(key string) (any, bool)
func (t *Theme) RegisterExtension(ext ThemeExtension)
func (t *Theme) TypedExtension(name string) ThemeExtension
func (t *Theme) TypedExtensions() map[string]ThemeExtension
func (t *Theme) MergeExtensions(other *Theme)
func (t *Theme) LerpExtensions(other *Theme, amount float32)
func (t *Theme) IsLight() bool
func (t *Theme) IsDark() bool
func (t *Theme) OnSurface() widget.Color   // satisfies widget.ThemeProvider; returns t.Colors.OnSurface
func (t *Theme) ScaleTypography(factor float32) *Theme
func (t *Theme) ScaleSpacing(factor float32) *Theme
func (t *Theme) Compact() *Theme      // 75% spacing
func (t *Theme) Comfortable() *Theme  // 125% spacing
```

### 1.2 `theme.ColorPalette` (theme/colors.go) — 21 fields, all `widget.Color`

```go
type ColorPalette struct {
    Primary, PrimaryLight, PrimaryDark          widget.Color
    Secondary, SecondaryLight, SecondaryDark    widget.Color
    Background, Surface, SurfaceVariant         widget.Color
    Error, Warning, Success, Info               widget.Color
    OnPrimary, OnSecondary, OnBackground, OnSurface, OnError widget.Color
    Divider, Outline, Shadow                    widget.Color
}
func (p *ColorPalette) WithAlpha(alpha float32) ColorPalette
func (p *ColorPalette) Lerp(other *ColorPalette, t float32) ColorPalette
```
Package helpers (theme/colors.go):
```go
func ContrastColor(background, onLight, onDark widget.Color) widget.Color // simple 0.299/0.587/0.114 luminance, >=0.5 → onLight
func Lighten(c widget.Color, amount float32) widget.Color  // Lerp toward white
func Darken(c widget.Color, amount float32) widget.Color   // Lerp toward black
func WithOpacity(c widget.Color, opacity float32) widget.Color // multiplies existing alpha
```

### 1.3 Typography (theme/typography.go)

```go
type FontWeight uint16 // FontWeightThin=100 … FontWeightNormal=400, FontWeightMedium=500, FontWeightSemiBold=600, FontWeightBold=700 … FontWeightBlack=900
type FontStyle uint8   // FontStyleNormal, FontStyleItalic, FontStyleOblique

type TextStyle struct {
    Font          string      // family name, "System" default
    Size          float32     // logical px
    Weight        FontWeight
    Style         FontStyle
    LineHeight    float32     // logical px; 0 = derived from size
    LetterSpacing float32     // logical px
}
// builders: WithSize, WithWeight, WithStyle, WithFont, WithLineHeight, WithLetterSpacing, Bold(), Italic()

type Typography struct {
    FontFamily string
    DisplayLarge, DisplayMedium, DisplaySmall    TextStyle
    HeadlineLarge, HeadlineMedium, HeadlineSmall TextStyle
    TitleLarge, TitleMedium, TitleSmall          TextStyle
    BodyLarge, BodyMedium, BodySmall             TextStyle
    LabelLarge, LabelMedium, LabelSmall          TextStyle
}
func DefaultTypography() Typography  // M3 scale: Display 57/45/36, Headline 32/28/24, Title 22/16/14, Body 16/14/12, Label 14/12/11
func (t *Typography) WithFontFamily(fontFamily string) Typography // sets FontFamily + every style's Font
func (t *Typography) Scale(factor float32) Typography
```
**Important:** these typography tokens are *advisory* — built-in painters never read them (they use hardcoded `fontSize` constants). See gotchas (§7).

### 1.4 Spacing (theme/spacing.go)

```go
type SpacingScale struct { XXS, XS, S, M, L, XL, XXL, XXXL float32 }
func DefaultSpacing() SpacingScale // 2,4,8,16,24,32,48,64
func (s SpacingScale) Scale(factor float32) SpacingScale
func (s SpacingScale) Inset(value float32) (top, right, bottom, left float32)
func (s SpacingScale) InsetHorizontal(value float32) (top, right, bottom, left float32)
func (s SpacingScale) InsetVertical(value float32) (top, right, bottom, left float32)
func (s SpacingScale) Compact() SpacingScale   // 0.75x
func (s SpacingScale) Relaxed() SpacingScale   // 1.5x
func DenseSpacing() SpacingScale       // 1,2,4,8,12,16,24,32
func ComfortableSpacing() SpacingScale // 4,8,12,24,32,48,64,96
```

### 1.5 Shadows (theme/shadows.go)

```go
type Shadow struct { OffsetX, OffsetY, Blur, Spread float32; Color widget.Color } // CSS box-shadow model
func (s Shadow) WithAlpha(alpha float32) Shadow
func (s Shadow) WithBlur(blur float32) Shadow
func (s Shadow) WithOffset(x, y float32) Shadow
func (s Shadow) Scale(factor float32) Shadow
func (s Shadow) IsZero() bool

type ElevationShadow struct { Key Shadow; Ambient Shadow }
type ShadowStyles struct { Level0, Level1, Level2, Level3, Level4, Level5 ElevationShadow }
func (s *ShadowStyles) ForElevation(level int) ElevationShadow
func DefaultShadowsLight() ShadowStyles // key rgba(0,0,0,.14), ambient rgba(0,0,0,.08)
func DefaultShadowsDark() ShadowStyles  // key .30, ambient .20
```
**These are pure data tokens — no canvas API consumes them directly.** (See §7.)

### 1.6 Radii (theme/radii.go)

```go
type RadiusScale struct { None, XS, S, M, L, XL, XXL, Full float32 }
func DefaultRadii() RadiusScale // 0,2,4,8,12,16,24,9999
func SharpRadii() RadiusScale   // 0,1,2,3,4,6,8,9999
func SoftRadii() RadiusScale    // 0,4,8,16,24,32,48,9999
func (r RadiusScale) Scale(factor float32) RadiusScale
func (r RadiusScale) Clamp(value, minVal, maxVal float32) float32

type CornerRadius struct { TopLeft, TopRight, BottomRight, BottomLeft float32 }
func Uniform(radius float32) CornerRadius; func Top(...)/Bottom(...)/Left(...)/Right(...) CornerRadius
func (c CornerRadius) IsUniform() bool; Max() float32; Scale(factor float32) CornerRadius
```
**`CornerRadius` exists only as a token type — Canvas has no per-corner round-rect call** (see §4/§7).

### 1.7 DefaultLight / DefaultDark (theme/presets.go)

`DefaultLight()` returns a `*Theme` with Name "Light", `ModeLight`, Material-blue palette:
Primary `Hex(0x1976D2)`, PrimaryLight `0x63A4FF`, PrimaryDark `0x004BA0`; Secondary teal `0x009688`/`0x52C7B8`/`0x00675B`; Background `0xFFFFFF`, Surface `0xFAFAFA`, SurfaceVariant `0xF5F5F5`; Error `0xD32F2F`, Warning `0xED6C02`, Success `0x2E7D32`, Info `0x0288D1`; OnPrimary/OnSecondary white, OnBackground/OnSurface `0x212121`, OnError white; Divider `RGBA(0,0,0,0.12)`, Outline `RGBA(0,0,0,0.23)`, Shadow `RGBA(0,0,0,0.20)`; plus `DefaultTypography()`, `DefaultSpacing()`, `DefaultShadowsLight()`, `DefaultRadii()`.

`DefaultDark()`: Name "Dark", `ModeDark`; Primary `0x64B5F6`/`0x9BE7FF`/`0x2286C3`; Secondary `0x4DB6AC`/`0x82E9DE`/`0x00867D`; Background `0x121212`, Surface `0x1E1E1E`, SurfaceVariant `0x2C2C2C`; Error `0xEF5350`, Warning `0xFFA726`, Success `0x66BB6A`, Info `0x29B6F6`; OnPrimary/OnSecondary/OnError black, OnBackground/OnSurface `0xE0E0E0`; Divider `RGBA(1,1,1,0.12)`, Outline `RGBA(1,1,1,0.23)`, Shadow `RGBA(0,0,0,0.50)`; `DefaultShadowsDark()`.

Also: `DefaultHighContrast()`, `Purple()`, `Green()`, `Orange()`, `ForMode(mode ThemeMode) *Theme`.

### 1.8 ThemeMode (theme/mode.go)

```go
type ThemeMode uint8
const ( ModeLight ThemeMode = iota; ModeDark; ModeSystem )
func (m ThemeMode) IsLight()/IsDark()/IsSystem() bool
func (m ThemeMode) ResolvedMode(preferLight bool) ThemeMode
```

---

## 2. The Painter pattern — every widget hook, exact signatures

### 2.1 Core mechanic

Every widget in `core/*` owns a `painter Painter` field (its own per-package `Painter` interface). The widget's `Draw(ctx widget.Context, canvas widget.Canvas)` builds a `PaintState` snapshot and calls the painter. **There is NO global painter registry and NO dispatch through the Theme** — a painter is installed *per widget instance* via the construction option `PainterOpt(p Painter)` (radio uses `GroupPainter`, docking uses `PainterOpt(p Painter) HostOption`, menu uses `PainterOpt(p Painter) BarOption`). If unset, the package's `DefaultPainter{}` is used.

A design system is therefore a package of painter structs, each holding `Theme *Theme` (the design system's own theme type), plus a convenience bundle. The only existing bundle precedent is `theme/devtools/painters.go`:

```go
type Painters struct {
    Button      ButtonPainter
    Checkbox    CheckboxPainter
    Radio       RadioPainter
    TextField   TextFieldPainter
    Dropdown    DropdownPainter
    Slider      SliderPainter
    Dialog      DialogPainter
    Scrollbar   ScrollbarPainter
    TabView     TabViewPainter
    TreeView    TreeViewPainter
    DataTable   DataTablePainter
    Toolbar     ToolbarPainter
    Menu        MenuPainter
    Collapsible CollapsiblePainter
    Progress    ProgressPainter
    SplitView   SplitViewPainter
    Docking     DockingPainter
    Popover     PopoverPainter
    LineChart   LineChartPainter
    ListView    ListViewPainter
    TitleBar    TitleBarPainter
    Stripe      StripePainter
}
func NewPainters(t *Theme) Painters // all share the same *Theme pointer
```
Usage pattern (examples/ide/main.go, examples/gallery/main.go):
```go
p := devtools.NewPainters(dt)
btn := button.New(button.TextOpt("Run"), button.PainterOpt(p.Button))
tf  := textfield.New(textfield.PainterOpt(p.TextField))
```

The Material 3 painter pattern (theme/material3/button.go — replicate this for shadcn):
```go
type ButtonPainter struct { Theme *Theme } // nil → built-in fallback palette
func (p ButtonPainter) PaintButton(canvas widget.Canvas, state button.PaintState) { ... }
var _ button.Painter = ButtonPainter{}
```
M3 painters first honor `state.ColorScheme` if non-zero, else `p.resolveColors()` from `p.Theme`, else hardcoded defaults.

### 2.2 Complete catalog of painter interfaces (all 24 widget packages under `core/`)

| Package | Interface (always named `Painter`) | Methods |
|---|---|---|
| `core/button` | `Painter` | `PaintButton(canvas widget.Canvas, state PaintState)` |
| `core/checkbox` | `Painter` | `PaintCheckbox(canvas widget.Canvas, state PaintState)` |
| `core/radio` | `Painter` | `PaintRadio(canvas widget.Canvas, state PaintState)` |
| `core/textfield` | `Painter` | `PaintTextField(canvas widget.Canvas, state PaintState)` |
| `core/dropdown` | `Painter` | `PaintTrigger(canvas widget.Canvas, state *TriggerPaintState)`; `PaintMenu(canvas widget.Canvas, state *MenuPaintState)` |
| `core/slider` | `Painter` | `PaintSlider(canvas widget.Canvas, state PaintState)` |
| `core/dialog` | `Painter` | `PaintDialog(canvas widget.Canvas, state PaintState)` |
| `core/scrollview` | `Painter` | `PaintScrollbar(canvas widget.Canvas, state PaintState)` |
| `core/tabview` | `Painter` | `PaintTabBar(canvas widget.Canvas, state PaintState)` |
| `core/treeview` | `Painter` | `PaintRowBackground(canvas, state RowPaintState)`; `PaintSelection(canvas, state RowPaintState)`; `PaintExpandIcon(canvas, state ExpandIconState)`; `PaintConnectorLines(canvas, state ConnectorState)`; `PaintLabel(canvas, state LabelState)`; `PaintEmptyState(canvas, bounds geometry.Rect)` |
| `core/datatable` | `Painter` | `PaintHeader(canvas, bounds geometry.Rect, state HeaderPaintState)`; `PaintHeaderCell(canvas, bounds geometry.Rect, state HeaderCellPaintState)`; `PaintRow(canvas, state RowPaintState)`; `PaintCell(canvas, state CellPaintState)`; `PaintEmptyState(canvas, bounds geometry.Rect)` |
| `core/toolbar` | `Painter` | `PaintToolbar(canvas, state PaintToolbarState)`; `PaintButtonItem(canvas, state PaintButtonState)`; `PaintSeparator(canvas, bounds geometry.Rect)` |
| `core/menu` | `Painter` | `PaintMenuBar(canvas, state *MenuBarPaintState)`; `PaintMenu(canvas, state *MenuPaintState)` |
| `core/collapsible` | `Painter` | `PaintHeader(canvas, state HeaderState)` |
| `core/progress` | `Painter` | `PaintProgress(canvas, state PaintState)` (circular spinner) |
| `core/progressbar` | `Painter` | `PaintProgressBar(canvas, state PaintState)` |
| `core/splitview` | `Painter` | `PaintDivider(canvas, state PaintState)` |
| `core/docking` | `Painter` | `PaintZoneTabs(canvas, state ZoneTabsPaintState)`; `PaintZoneBorder(canvas, borderRect geometry.Rect, zone Zone)` |
| `core/popover` | `Painter` | `PaintPopover(canvas, state *PopoverPaintState)`; `PaintTooltip(canvas, state *TooltipPaintState)` |
| `core/linechart` | `Painter` | `PaintChart(canvas widget.Canvas, bounds geometry.Rect, state PaintState)` |
| `core/listview` | `Painter` | `PaintDivider(canvas, state DividerState)`; `PaintEmptyState(canvas, bounds geometry.Rect)`; `PaintItemBackground(canvas, state ItemPaintState)`; `PaintSelection(canvas, state ItemPaintState)` |
| `core/gridview` | `Painter` | `PaintCellBackground(canvas, state CellPaintState)`; `PaintSelection(canvas, state CellPaintState)`; `PaintEmptyState(canvas, bounds geometry.Rect)` |
| `core/titlebar` | `Painter` | `DrawBackground(canvas, bounds geometry.Rect, state BackgroundState)`; `DrawControlButton(canvas, bounds geometry.Rect, control ControlType, state ControlState)` |
| `core/stripe` | `Painter` | `PaintBackground(canvas, bounds geometry.Rect)`; `PaintButton(canvas, state ButtonPaintState)` |

(`core/scrollview` is the scrollbar painter for scrollable areas. `core/dialog`, `core/popover` render into overlays.)

### 2.3 Key PaintState structs (verbatim)

**button** (core/button/painter.go):
```go
type PaintState struct {
    Text     string
    Variant  Variant   // Filled | Outlined | TextOnly | Tonal (uint8)
    Size     Size      // Small(32px h, 12px font) | Medium(40, 14) | Large(48, 16)
    Hovered  bool
    Pressed  bool
    Focused  bool
    Disabled bool
    Bounds   geometry.Rect
    Background  *widget.Color      // per-widget override; nil = painter decides
    Radius      *float32           // per-widget override; nil = painter decides
    ColorScheme ButtonColorScheme  // zero value = painter uses its own theme
}
type ButtonColorScheme struct {
    FilledBg, FilledFg, OutlinedBorder, TextBgHover, TonalBg, TonalFg,
    Primary, DisabledBg, DisabledFg, FocusRing widget.Color
}
```

**textfield** (core/textfield/painter.go) — note: NO ColorScheme field in PaintState; theme colors come only from the painter struct itself:
```go
type PaintState struct {
    Text        string
    Placeholder string
    Focused     bool
    Hovered     bool
    Disabled    bool
    HasError    bool
    ErrorMsg    string
    CursorPos   int
    SelectStart int
    SelectEnd   int
    InputType   InputType   // TypeText / TypePassword etc.
    Bounds      geometry.Rect
}
type TextFieldColorScheme struct {
    Background, Border, FocusBorder, ErrorBorder, TextColor, Placeholder,
    CursorColor, DisabledBg, DisabledFg, SelectionBg, ErrorText widget.Color
}
```

**checkbox** (core/checkbox/painter.go):
```go
type PaintState struct {
    Label         string
    Checked       bool
    Indeterminate bool
    Hovered, Pressed, Focused, Disabled bool
    Bounds        geometry.Rect
    Background  *widget.Color
    ColorScheme CheckboxColorScheme
}
type CheckboxColorScheme struct { CheckedBg, CheckedFg, UncheckedBorder, LabelColor, DisabledBg, DisabledFg, FocusRing widget.Color }
```

**radio**: `PaintState{ Label string; Selected, Hovered, Pressed, Focused, Disabled bool; Bounds geometry.Rect; ColorScheme RadioColorScheme }`; `RadioColorScheme{ SelectedBg, SelectedFg, UnselectedBorder, LabelColor, DisabledBg, DisabledFg, FocusRing }`.

**slider**: `PaintState{ Value, Min, Max, Progress float32; Hovered, Dragging, Focused, Disabled bool; Bounds geometry.Rect; Orientation Orientation; Marks []Mark; ColorScheme SliderColorScheme }`; `SliderColorScheme{ ActiveTrack, InactiveTrack, Thumb, ThumbBorder, FocusRing, DisabledTrack, DisabledThumb, MarkColor }`.

**dropdown**: `TriggerPaintState{ Bounds geometry.Rect; SelectedText string; IsPlaceholder, Open, Focused, Hovered, Disabled bool; ColorScheme DropdownColorScheme }`; `MenuPaintState{ Bounds geometry.Rect; Items []ItemDef; HighlightedIndex, SelectedIndex, ScrollOffset, VisibleCount int; ItemHeight float32; ColorScheme DropdownColorScheme }`; `DropdownColorScheme{ Background, Border, FocusBorder, TextColor, PlaceholderText, DisabledBg, DisabledFg, MenuBg, MenuBorder, ItemHover, ItemSelected, ItemDisabled, ChevronColor, FocusRing }`.

**dialog**: `PaintState{ Title string; HasContent bool; Actions []Action; Focused bool; Bounds geometry.Rect; ColorScheme DialogColorScheme }`; `DialogColorScheme{ Backdrop, Surface, Title, Content, Border, Shadow, ActionFg, ActionBg }`.

**scrollview**: `PaintState{ Bounds geometry.Rect; Direction ScrollDirection; Focused, Hovered, Dragging bool; ColorScheme ScrollbarColorScheme; VScrollVisible bool; VThumbRect, VTrackRect geometry.Rect; HScrollVisible bool; HThumbRect, HTrackRect geometry.Rect }`; `ScrollbarColorScheme{ Track, Thumb, ThumbHover, ThumbDrag }`.

**tabview**: `PaintState{ Bounds geometry.Rect; Tabs []TabState; SelectedIdx int; Position TabPosition; Focused bool; ColorScheme TabColorScheme }`; `TabColorScheme{ Background, SelectedText, UnselectedText, Indicator, HoverBackground, CloseButton, FocusRing }`.

**treeview**: `RowPaintState{ Bounds geometry.Rect; Node *TreeNode; Depth int; Selected, Focused, Hovered, Disabled bool; ColorScheme TreeColorScheme }`; `ExpandIconState{ Bounds; Expanded, Hovered bool; ColorScheme }`; `ConnectorState{ RowBounds; Depth int; IndentWidth float32; IsLastChild, HasChildren bool; ParentHasMore []bool; ColorScheme }`; `LabelState{ Bounds; Text string; Selected, Disabled bool; ColorScheme }`; `TreeColorScheme{ SelectionColor, HoverColor, FocusColor, LabelColor, LineColor, IconColor, EmptyTextColor }`.

**datatable**: `HeaderPaintState{ Disabled bool; ColorScheme TableColorScheme }`; `HeaderCellPaintState{ Title string; Align widget.TextAlign; Sortable bool; SortDir SortDirection; Hovered, Disabled bool; ColorScheme }`; `RowPaintState{ Bounds; RowIndex int; Selected, Focused, Hovered, Disabled bool; ColorScheme }`; `CellPaintState{ Bounds; Value string; Align widget.TextAlign; RowIndex, ColIndex int; Selected, Disabled bool; ColorScheme }`; `TableColorScheme{ HeaderBackground, HeaderText, RowBackground, RowAlternate, SelectionColor, HoverColor, FocusColor, CellText, Divider, EmptyText }`.

**menu**: `MenuBarPaintState{ Bounds; Menus []TopMenu; MenuRects []geometry.Rect; OpenIndex, HoveredIndex int; Focused bool; ColorScheme MenuColorScheme }`; `MenuPaintState{ Bounds; Items []MenuItem; HighlightedIndex int; ItemHeight, SeparatorHeight float32; SubMenuOpenIndex int; ColorScheme }`; `MenuColorScheme{ BarBackground, BarText, BarHover, BarActiveText, MenuBackground, MenuBorder, ItemText, ItemHover, ItemDisabledText, ShortcutText, SeparatorColor, SubMenuArrow }`.

**popover**: `PopoverPaintState{ Bounds geometry.Rect; Placement Placement; ColorScheme PopoverColorScheme }`; `TooltipPaintState{ Bounds; Text string; Placement Placement; ColorScheme TooltipColorScheme }`; `PopoverColorScheme{ Background, Border, Shadow widget.Color; ShadowBlur float32 }`; `TooltipColorScheme{ Background, TextColor, Border }`.

**collapsible**: `HeaderState{ Title string; Expanded, Hovered, Pressed, Focused bool; Bounds geometry.Rect; ArrowProgress float32; HeaderColor, ArrowColor widget.Color }`.

**progress** (circular): `PaintState{ Value float64; Bounds; Diameter, StrokeWidth float32; ShowLabel bool; Label string; Indeterminate bool; Rotation float64; AnimationPhase float64; Disabled bool; ColorScheme ProgressColorScheme }`; `ProgressColorScheme{ Indicator, Track, Label, DisabledIndicator, DisabledTrack widget.Color; /* + unexported indicatorSet, trackSet bools */ }`.

**progressbar**: `PaintState{ Value float64; Bounds; BarHeight, Radius float32; ShowLabel bool; Label string; Disabled bool; ProgressBarColorScheme ProgressBarColorScheme }` (note field name!); `ProgressBarColorScheme{ Bar, Track, Label, DisabledBar, DisabledTrack }`.

**splitview**: `PaintState{ DividerRect geometry.Rect; Orientation Orientation; Hovered, Dragging, Collapsed bool; ColorScheme DividerColorScheme }`; `DividerColorScheme{ Divider, DividerHover, DividerDrag, Handle }`.

**docking**: `ZoneTabsPaintState{ Zone Zone; TabBarBounds geometry.Rect; Tabs []ZoneTabState; ActiveIdx int; ColorScheme ZoneColorScheme }`; `ZoneColorScheme{ TabBarBackground, ActiveTabText, InactiveTabText, ActiveTabBackground, HoverBackground, Border, CloseButton }`.

**listview**: `DividerState{ Bounds; ItemIndex int; ColorScheme ListColorScheme }`; `ItemPaintState{ Bounds; Index int; Selected, Focused, Hovered, Disabled bool; ColorScheme }`; `ListColorScheme{ DividerColor, SelectionColor, HoverColor, FocusColor, EmptyTextColor, ItemBackground, ItemBackgroundAlt }`.

**gridview**: `CellPaintState{ Bounds; Index, Row, Col int; Selected, Focused, Hovered, Disabled bool; ColorScheme GridColorScheme }`; `GridColorScheme{ SelectionColor, HoverColor, FocusColor, EmptyTextColor, CellBackground }`.

**toolbar**: `PaintToolbarState{ Bounds geometry.Rect }`; `PaintButtonState{ Label string; Icon icon.IconData; ShowLabel, Hovered, Pressed, Focused, Disabled bool; Bounds geometry.Rect }`.

**titlebar**: `BackgroundState{ Focused bool }`; `ControlState{ Hovered, Pressed bool }`.

**stripe**: `ButtonPaintState{ Bounds geometry.Rect; Icon icon.IconData; Label string; Active, Hovered, Pressed, ShowLabel bool }`.

**linechart**: `PaintState{ Series []Series; MaxPoints int; YMin, YMax float64; ShowGrid, ShowLabels bool; GridColor, Background widget.Color }`.

### 2.4 Who fills ColorScheme

`PaintState.ColorScheme` is filled from the *widget config*, not the app theme: several widgets expose `ColorSchemeOpt(cs XxxColorScheme) Option` (confirmed in splitview, progress, progressbar, docking; dropdown carries `m.colorScheme`). Button/checkbox PaintState ColorScheme is currently never set by the widget itself (always zero) — the design-system painter supplies colors. M3 painters check `if colors == (button.ButtonColorScheme{}) { colors = p.resolveColors() }`, so a non-zero PaintState scheme wins over the painter's theme.

---

## 3. Theme registration, selection, application; overrides; extensions

### 3.1 Theme registry (theme/registry.go)

```go
type ThemeVariant string // VariantLight "light", VariantDark "dark", VariantSystem "system"
type ThemeInfo struct { Name, Description, Author, Version string; Variants []ThemeVariant; Preview string }
type ThemeRegistry struct { /* mutex-guarded maps */ }
func NewThemeRegistry() *ThemeRegistry
func (r *ThemeRegistry) Register(name string, theme *Theme, info ...ThemeInfo)
func (r *ThemeRegistry) Unregister(name string) bool
func (r *ThemeRegistry) Get(name string) (*Theme, bool)
func (r *ThemeRegistry) MustGet(name string) *Theme
func (r *ThemeRegistry) List() []string
func (r *ThemeRegistry) Info(name string) (ThemeInfo, bool)
func (r *ThemeRegistry) Count() int; Has(name string) bool; Clear(); ListByVariant(variant ThemeVariant) []string
// package-level wrappers on a global registry:
func Register(name string, theme *Theme, info ...ThemeInfo)
func Get(name string) (*Theme, bool); MustGet(name string) *Theme; List() []string; Info(...)
func GlobalRegistry() *ThemeRegistry
```
Built-in `init()` registers: "light", "dark", "high-contrast", "purple", "green", "orange". A shadcn theme package would do `theme.Register("shadcn", shadcnTheme, theme.ThemeInfo{...})` in its `init()`.

### 3.2 App-level application (app/app.go, app/window.go)

```go
func WithTheme(t *theme.Theme) Option        // app.New(app.WithTheme(t)); nil → DefaultLight()
func (a *App) SetTheme(t *theme.Theme)       // updates App + Window, full relayout/redraw
func (a *App) Theme() *theme.Theme
func (w *Window) Theme() *theme.Theme
func (w *Window) ThemeBackground() widget.Color  // theme.Colors.Background, used as window clear color in RenderModeFrameworkManaged
```
Internally `Window.setTheme` calls `w.ctx.SetThemeProvider(t)` and marks `needsLayout/needsRedraw/needsFullRepaint` + `widget.MarkRedrawInTree(root)`.

Widgets access the theme via `widget.Context`:
```go
ThemeProvider() ThemeProvider // on widget.Context interface; nil in headless mode
// widget/theme.go:
type ThemeProvider interface {
    IsDark() bool
    OnSurface() Color
}
```
`*theme.Theme`, `*material3.Theme`, `*fluent.Theme`, `*cupertino.Theme`, `*devtools.Theme` all implement it. This is the ONLY theme information generic widgets see — e.g. `primitives.TextWidget` resolves text color as: explicit > `ThemeProvider().OnSurface()` > `ColorBlack`.

### 3.3 Subtree scoping (primitives/themescope.go)

```go
func ThemeScope(theme widget.ThemeProvider, children ...widget.Widget) *ThemeScopeWidget
func (ts *ThemeScopeWidget) Theme() widget.ThemeProvider
func (ts *ThemeScopeWidget) SetTheme(theme widget.ThemeProvider)
```
Wraps the context in a `themeScopeContext` overriding only `ThemeProvider()`. Priority chain (documented): `Widget override > Nearest ThemeScope > App theme > Default`. Note: ThemeScope changes what `ctx.ThemeProvider()` returns — it does **not** swap painters; painters bound at construction stay.

### 3.4 Per-widget style overrides (button as the model)

Construction options (core/button/options.go): `TextOpt(string)`, `TextFn(func() string)`, `TextSignal(state.Signal[string])`, `TextReadonlySignal(...)`, `OnClick(func())`, `Disabled(bool)`, `DisabledFn/DisabledSignal/DisabledReadonlySignal`, `VariantOpt(Variant)`, `SizeOpt(Size)`, `A11yHint(string)`, `BackgroundOpt(widget.Color)`, `RoundedOpt(radius float32)`, `PainterOpt(p Painter)`.

Fluent post-construction (core/button/styling.go): `Padding(v float32)`, `PaddingXY(x, y float32)`, `SetBackground(c widget.Color)`, `SetRounded(radius float32)`, `MinWidth(v float32)`, `MaxWidth(v float32)` — all return `*Widget`.

These overrides flow into PaintState as `Background *widget.Color` and `Radius *float32`; painters must honor them (`if state.Radius != nil { radius = *state.Radius }`).

### 3.5 ThemeExtension (theme/extension.go) — the official way to hang shadcn tokens on theme.Theme

```go
type ThemeExtension interface {
    Name() string
    Merge(other ThemeExtension) ThemeExtension
    Lerp(other ThemeExtension, t float32) ThemeExtension
    CopyWith(overrides map[string]any) ThemeExtension
}
func ExtensionAs[T ThemeExtension](t *Theme, name string) (T, bool) // generic typed getter
func LerpString(a, b string, t float32) string
func LerpFloat32(a, b, t float32) float32
func LerpInt(a, b int, t float32) int
```
Register with `theme.RegisterExtension(ext)`; retrieve with `ExtensionAs[*ShadcnExtension](t, "shadcn")`. Simple untyped storage alternative: `t.SetExtension("shadcn", v)` / `t.GetExtension("shadcn")`. `Theme.MergeExtensions`/`LerpExtensions` give inheritance and animated transitions for free.

---

## 4. Canvas API available inside painters (widget/canvas.go)

```go
type Canvas interface {
    Clear(color Color)
    DrawRect(r geometry.Rect, color Color)
    FillRectDirect(r geometry.Rect, color Color) // CPU-only fill (dirty-region clearing)
    StrokeRect(r geometry.Rect, color Color, strokeWidth float32)
    DrawRoundRect(r geometry.Rect, color Color, radius float32)               // UNIFORM radius only
    StrokeRoundRect(r geometry.Rect, color Color, radius float32, strokeWidth float32)
    DrawCircle(center geometry.Point, radius float32, color Color)
    StrokeCircle(center geometry.Point, radius float32, color Color, strokeWidth float32)
    StrokeArc(center geometry.Point, radius float32, startAngle, sweepAngle float64, color Color, strokeWidth float32) // radians, 0 = 3 o'clock
    DrawLine(from, to geometry.Point, color Color, strokeWidth float32)
    DrawText(text string, bounds geometry.Rect, fontSize float32, color Color, bold bool, align TextAlign)
    MeasureText(text string, fontSize float32, bold bool) float32
    DrawImage(img image.Image, at geometry.Point)
    PushClip(r geometry.Rect)
    PushClipRoundRect(r geometry.Rect, radius float32)
    PopClip()
    PushTransform(offset geometry.Point) // translation only
    PopTransform()
    TransformOffset() geometry.Point
    ScreenOriginBase() geometry.Point
    ClipBounds() geometry.Rect
    ReplayScene(s *scene.Scene)
}
```

Optional capability interfaces (type-assert the canvas):
```go
type ArcStroker interface { StrokeArcStyled(center geometry.Point, radius float32, startAngle, sweepAngle float64, color Color, strokeWidth float32, lineCap LineCap) }
type LineCap uint8 // LineCapButt, LineCapRound, LineCapSquare
type SVGFiller interface { FillSVGPath(svgData string, viewBox float32, bounds geometry.Rect, color Color) }
type SVGRenderer interface { RenderSVG(svgXML []byte, bounds geometry.Rect, color Color) }
type DeviceScaler interface { SetDeviceScale(scale float32) }
type TextModeController interface { SetTextMode(mode TextMode); TextMode() TextMode } // Auto/MSDF/Vector/Bitmap/GlyphMask
type StyledTextDrawer interface {
    DrawStyledText(text string, bounds geometry.Rect, style TextStyle)
    MeasureStyledText(text string, style TextStyle) float32
}
type widget.TextStyle struct {  // for StyledTextDrawer — NOT theme.TextStyle
    FontFamily string  // "" → embedded Inter
    FontSize   float32
    Bold       bool    // 700 vs 400 — no numeric weights
    Italic     bool
    Color      Color
    Align      TextAlign
}
type TextAlign uint8 // TextAlignLeft, TextAlignCenter, TextAlignRight; Float64() → 0/0.5/1
```

What does NOT exist on Canvas: per-corner round rects, gradients, blur/drop-shadow primitive, global opacity/layer push, arbitrary paths (beyond SVG-path fill via optional interface), rotation/scale transforms (translation only), dashed strokes, inner shadows.

Geometry helpers (geometry pkg): `Pt(x, y float32) Point`, `Sz(w, h float32) Size`, `NewRect(x, y, width, height float32) Rect`, `FromPointSize(p Point, s Size) Rect`; `Rect{Min, Max Point}` with `Width()`, `Height()`, `Center()`, `Contains(p)`, `ContainsRect`, `Intersects`, `Intersection`, `Union`, `Inset(insets Insets)`, `Expand(delta float32)`, `IsEmpty()`.

Rendering implementation notes (internal/render/canvas.go): all shapes route through gg's `DrawRectangle`/`DrawRoundedRectangle` + `Fill()`/`Stroke()`; **strokes are centered on the path** (clip-visibility check uses `r.Expand(strokeWidth/2)`), so a 1px border paints 0.5px outside and 0.5px inside the rect. `DrawText` vertically centers within bounds using font metrics.

Testing: `uitest` package has a recording canvas capturing `DrawRectCall{Bounds, Color}`, `DrawRoundRectCall{Bounds, Color, Radius}`, `StrokeRoundRectCall{Bounds, Color, Radius, StrokeWidth}`, `DrawCircleCall`, `StrokeArcCall`, `DrawLineCall`, etc. — ideal for unit-testing shadcn painters without a GPU.

---

## 5. Colors

`widget.Color` (widget/canvas.go):
```go
type Color struct { R, G, B, A float32 } // each 0..1
func RGBA(r, g, b, a float32) Color
func RGB(r, g, b float32) Color
func RGBA8(r, g, b, a uint8) Color
func RGB8(r, g, b uint8) Color
func Hex(hex uint32) Color   // 0xRRGGBB, opaque
func HexA(hex uint32) Color  // 0xRRGGBBAA
func (c Color) WithAlpha(a float32) Color   // REPLACES alpha (does not multiply)
func (c Color) Lerp(other Color, t float32) Color
func (c Color) IsOpaque() bool; IsTransparent() bool
func (c Color) RGBA8() (r, g, b, a uint8)
// constants: ColorTransparent, ColorBlack, ColorWhite, ColorRed, ColorGreen, ColorBlue,
//            ColorYellow, ColorCyan, ColorMagenta, ColorGray, ColorLightGray, ColorDarkGray
```
Doc comment claims "premultiplied alpha" but values are passed straight to gg `SetRGBA` — treat as straight alpha in practice.

Color-space tooling: only inside `theme/material3` and it is **unexported**: `hct{Hue, Chroma, Tone}` with `hctFromColor`/`hctToColor`/`withTone` — an HSL-based approximation of HCT, not CAM16. `material3.Light(seed widget.Color) ColorScheme` / `Dark(seed)` generate full M3 schemes via tonal palettes. There is **no OKLCH/HSL public helper** — for shadcn (which is specified in OKLCH/HSL), precompute all colors as hex literals; nothing in the library needs runtime color-space math. `theme.Lighten/Darken/WithOpacity` and `Color.Lerp` are the only public manipulation helpers.

The Material 3 `ColorScheme` struct (theme/material3/color.go) shows the full role set if you want to mirror it: Primary/OnPrimary/PrimaryContainer/OnPrimaryContainer, same for Secondary/Tertiary/Error, Surface + OnSurface + SurfaceVariant + OnSurfaceVariant + SurfaceContainer{Lowest,Low,"",High,Highest}, Background/OnBackground, Outline/OutlineVariant, InverseSurface/InverseOnSurface/InversePrimary.

---

## 6. Interactive state flow into painters

- Each widget keeps an internal `interactionState` (e.g. button: `stateNormal/stateHover/statePressed`) updated in its `Event(ctx, e)` handler from `event.MouseEvent` enter/leave/press/release. On `Draw`, it snapshots into PaintState: `Hovered: w.state == stateHover`, `Pressed: w.state == statePressed`.
- **Focus** comes from `widget.WidgetBase.IsFocused()`, managed by `Context.RequestFocus/ReleaseFocus` (the Window's focus manager handles Tab traversal). Painters receive plain `Focused bool` and are responsible for drawing the focus ring themselves. Convention (button/checkbox M3 + default): ring = `canvas.StrokeRoundRect(bounds.Expand(2), focusColor, radius+2, 2)` (offset 2, stroke 2).
- **Disabled** is resolved from config (`ResolvedDisabled()`: ReadonlySignal > Signal > Fn > static bool) and passed as `Disabled bool`. Disabled widgets are not focusable (`IsFocusable()` checks it).
- Slider/scrollbar/splitview get `Dragging bool` instead of/in addition to Pressed. Lists/tables/trees get `Selected`, `Hovered`, `Focused` per row/cell. Dropdown gets `Open`.
- There are no built-in transition/animation hooks for state changes — hover/press color changes are instant unless a painter implements its own animation via `ctx` timing (painters don't receive ctx, only canvas + state, so animations must be driven by the widget; progress widgets pass `Rotation`/`AnimationPhase`).
- Hover/press color modulation conventions: default + M3 painters use `base.Lerp(widget.ColorWhite, 0.1)` for hover and `base.Lerp(widget.ColorBlack, 0.15)` for press. For shadcn you'll instead want explicit tokens (e.g. `primary/90` on hover) — fully under painter control.

---

## 7. Gotchas / constraints for replicating shadcn visuals

1. **No global painter installation.** `app.SetTheme(*theme.Theme)` only changes `ThemeProvider` data (background clear color, `OnSurface` for primitives.Text). Core widgets' `DefaultPainter`s use hardcoded constants and never read `ctx.ThemeProvider()`. A shadcn theme MUST ship a `Painters` bundle (like `devtools.NewPainters`) and the app must pass `xxx.PainterOpt(p.Xxx)` to every widget it creates. There is no way to retro-theme an already-built tree, and no `SetPainter` post-construction setter (painter is set only via option at `New`).

2. **No per-corner radii.** `DrawRoundRect`/`StrokeRoundRect`/`PushClipRoundRect` take a single uniform `radius float32`. `theme.CornerRadius` is a token type only. shadcn visuals are almost all uniform-radius so this is mostly fine; things like joined button groups with selective rounding would need manual compositing (overlap a plain rect over the corner you want squared).

3. **Borders are center-stroked.** `StrokeRoundRect` strokes centered on the rect path (half in, half out). shadcn borders are CSS inside-borders. To emulate a 1px inside border: stroke a rect inset by `strokeWidth/2` (`bounds.Expand(-0.5)` equivalent via `Inset`), as material3 does for the checkbox box (`m3CheckboxBoxRect` insets by `borderWidth/2`, see issue #117 comment). Otherwise adjacent widgets will visually overlap borders and focus rings get clipped weirdly against backgrounds.

4. **No real shadows/blur.** Canvas has no blur primitive. `theme.ShadowStyles` are inert tokens. Existing approaches: (a) `primitives.BoxWidget.ShadowLevel(0-5)` fakes Gaussian blur with 2–4 stacked `DrawRoundRect` layers (see `shadowPresets` in primitives/style.go — offsets/spread/alpha per level); (b) material3 popover draws a single offset round-rect in a translucent color. shadcn's `shadow-xs/sm/md` (e.g. `0 1px 2px 0 rgb(0 0 0 / 0.05)`) must be approximated the same way: 1–3 concentric `DrawRoundRect` calls with low alpha, larger radius, offset Y. They will not look identical to CSS box-shadow — keep alphas very low (0.03–0.08) to avoid hard edges.

5. **Text rendering is the biggest constraint.**
   - `Canvas.DrawText(text, bounds, fontSize, color, bold, align)` supports ONLY the embedded **Inter** font, in exactly two weights (Regular 400 / Bold 700), single-line, auto vertical-centering. No letter-spacing, no line-height, no font family.
   - The optional `StyledTextDrawer` (implemented by the real render.Canvas and SceneCanvas) adds `FontFamily` + `Italic`, but `widget.TextStyle.Bold` is a bool — `resolveStyledFontSource` maps it to `font.Regular` or `font.Bold` only. **shadcn's `font-medium` (500) cannot be requested by weight.** Workaround: register the Medium-weight TTF under its own family name (e.g. family "GeistMedium", weight Regular) via the plugin asset loader (`plugin` package wires `LoadFont(name, data)` → `render.GlobalFontRegistry().Register(name, font.Regular, font.Normal, data)`; `internal/render` is not importable directly). The `theme/font` package (`font.Registry`, `Family`, `Face`, `Weight` 100–900, CSS weight matching) supports full weights, but the public canvas surface doesn't expose them.
   - `theme.Typography` tokens (Size/Weight/LineHeight/LetterSpacing) are not consumed by any painter — your painters must hardcode shadcn sizes (14px text-sm etc.) or read them from your own theme struct.
   - `MeasureText`/`MeasureStyledText` exist, so precise text layout inside painters is possible.
   - Button width estimation in `Layout` is crude: `len(text) * fontSize * 0.55` (charWidthRatio) — long/short labels will be sized approximately, painters can't change layout.

6. **Painter cannot affect layout.** Sizing lives in the widget (`Layout`): button heights are fixed by `Size` (Small=32, Medium=40, Large=48; fonts 12/14/16). shadcn's default button is h-9 (36px): the only lever is `PaddingXY` — `Layout` uses `totalHeight = max(sizeHeight(size), paddingY*2)`, so `PaddingXY(16, 18)` yields 36px height with `Small`... actually `max(32, 36) = 36` — use `SizeOpt(button.Small)` + `PaddingXY(16,18)`. Text field, checkbox box size (M3: 18px, default painter similar), etc., are likewise widget/painter constants.

7. **Zero-value ColorScheme sentinel.** Painters detect "no scheme" by comparing to the zero struct (`state.ColorScheme == (button.ButtonColorScheme{})`). If a custom scheme legitimately wants all-transparent colors it would be treated as unset; always set at least one non-zero field. Also note `textfield.PaintState` has NO ColorScheme field — text field theming flows exclusively through the painter struct.

8. **Focus ring drawing.** Painters draw it via `StrokeRoundRect(bounds.Expand(2), color, radius+2, 2)`. It paints OUTSIDE widget bounds; in dense layouts/clipped containers (`PushClip`) it can be cut off. shadcn's `ring-2 ring-offset-2` look maps well (Expand(2) ≈ offset, stroke 2 = ring width), but the "offset" gap will show whatever is behind the widget (no ring-offset background fill) — draw an extra stroke in the page background color between button edge and ring if exact replication matters.

9. **Hover/press states are binary and instant** — no opacity transitions. shadcn relies on subtle `hover:bg-accent` switches which work fine, but animated transitions would need widget-level work, not painter-level.

10. **Radius `Full` = 9999** is the established pill convention; `DrawRoundRect` clamps fine in gg.

11. **Theme struct vs design-system struct.** Follow the established dual pattern: a `shadcn.Theme` struct (`Colors ColorScheme`, maybe `Radius float32`, `dark bool`) implementing `widget.ThemeProvider` (`IsDark()`, `OnSurface()`), plus `AsTheme() *theme.Theme` mapping to `theme.ColorPalette` (Primary, Background, Surface, Outline=border, Divider, Shadow...) so `app.WithTheme(sh.AsTheme())` sets the window background and primitives text color. Fluent (`theme/fluent/fluent.go`) is the minimal viable surface: Theme struct + Options (`WithAccentColor`) + `NewTheme/NewDarkTheme` + `IsDark/OnSurface/AsTheme` + custom `RadiusScale` + per-widget painters (button, checkbox, color, dialog, dropdown, radio, scrollbar, slider, tabview, textfield). Cupertino is even smaller (no AsTheme). DevTools is the most complete (22 painters + `NewPainters` bundle) — the best template for a full shadcn theme.

12. **Compile-time checks**: end each painter file with `var _ button.Painter = ButtonPainter{}` etc., as all existing themes do.

13. **HiDPI**: all coordinates are logical px; `ctx.Scale()` exists but painters don't get ctx — the canvas handles scaling. 1px borders render as hairlines correctly through gg.

### Minimal shadcn theme blueprint (derived from all of the above)

```go
package shadcn
type Theme struct { Colors ColorScheme; Radius float32; dark bool } // implement widget.ThemeProvider + AsTheme()
type ColorScheme struct { Background, Foreground, Card, CardForeground, Popover, PopoverForeground,
    Primary, PrimaryForeground, Secondary, SecondaryForeground, Muted, MutedForeground,
    Accent, AccentForeground, Destructive, DestructiveForeground, Border, Input, Ring widget.Color }
// 22+ painter structs each { Theme *Theme }, satisfying:
// button.Painter, checkbox.Painter, radio.Painter, textfield.Painter, dropdown.Painter,
// slider.Painter, dialog.Painter, scrollview.Painter, tabview.Painter, treeview.Painter,
// datatable.Painter, toolbar.Painter, menu.Painter, collapsible.Painter, progress.Painter,
// progressbar.Painter, splitview.Painter, docking.Painter, popover.Painter, linechart.Painter,
// listview.Painter, gridview.Painter, titlebar.Painter, stripe.Painter
// + type Painters struct{...}; func NewPainters(t *Theme) Painters   (devtools pattern)
// + func init() { theme.Register("shadcn", New().AsTheme(), ...) }   (optional registry entry)
```

Key files for implementation reference:
- `C:\Users\tingzhen\AppData\Local\Temp\gogpu-ui\theme\theme.go`, `colors.go`, `presets.go`, `radii.go`, `shadows.go`, `spacing.go`, `typography.go`, `registry.go`, `extension.go`, `mode.go`
- `C:\Users\tingzhen\AppData\Local\Temp\gogpu-ui\theme\material3\{theme,button,textfield,checkbox,color,popover}.go` (painter pattern)
- `C:\Users\tingzhen\AppData\Local\Temp\gogpu-ui\theme\devtools\painters.go` (bundle pattern, full widget coverage list)
- `C:\Users\tingzhen\AppData\Local\Temp\gogpu-ui\core\<widget>\painter.go` (every Painter interface + PaintState + ColorScheme)
- `C:\Users\tingzhen\AppData\Local\Temp\gogpu-ui\widget\canvas.go` (Canvas, Color, TextStyle, optional interfaces)
- `C:\Users\tingzhen\AppData\Local\Temp\gogpu-ui\primitives\style.go` (multi-layer fake-shadow presets), `themescope.go`
- `C:\Users\tingzhen\AppData\Local\Temp\gogpu-ui\app\app.go`, `window.go` (WithTheme/SetTheme/ThemeBackground)
- `C:\Users\tingzhen\AppData\Local\Temp\gogpu-ui\internal\render\canvas.go`, `fontregistry.go`, `scene_canvas.go` (stroke centering, font resolution limits)

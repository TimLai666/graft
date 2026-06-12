# gogpu/ui Implementation Cookbook (for a shadcn/ui port)

Source: local clone at `C:\Users\tingzhen\AppData\Local\Temp\gogpu-ui`. All signatures below are copied verbatim from source.

Module path: `github.com/gogpu/ui`. Rendering backend: `github.com/gogpu/gg` (+ `github.com/gogpu/gogpu` window host). Signals come from `github.com/coregx/signals` (re-exported as type aliases in `state`).

---

## 1. Minimal complete custom widget

### 1.1 The Widget interface (`widget/widget.go`)

```go
type Widget interface {
    Layout(ctx Context, constraints geometry.Constraints) geometry.Size
    Draw(ctx Context, canvas Canvas)
    Event(ctx Context, e event.Event) bool   // return true = consumed
    Children() []Widget                       // nil for leaves, z-order bottom→top
}
```

Embed `widget.WidgetBase` (`widget/base.go`) to get everything else: bounds, visibility, enabled, focus flag, parent/children storage, dirty tracking, screen-origin tracking, binding cleanup. Key inherited methods (exact signatures):

```go
func NewWidgetBase() *WidgetBase                       // rarely needed; embedding zero value + SetVisible/SetEnabled is the norm
func (w *WidgetBase) Bounds() geometry.Rect
func (w *WidgetBase) SetBounds(bounds geometry.Rect)   // parent calls this after child.Layout
func (w *WidgetBase) Size() geometry.Size
func (w *WidgetBase) Position() geometry.Point
func (w *WidgetBase) IsFocused() bool
func (w *WidgetBase) SetFocused(focused bool)
func (w *WidgetBase) IsVisible() bool / SetVisible(bool)
func (w *WidgetBase) IsEnabled() bool / SetEnabled(bool)
func (w *WidgetBase) Parent() Widget / SetParent(parent Widget)
func (w *WidgetBase) Children() []Widget / AddChild / RemoveChild / ClearChildren / InsertChild
func (w *WidgetBase) ContainsPoint(p geometry.Point) bool
func (w *WidgetBase) ScreenOrigin() geometry.Point      // window-space origin, stamped during Draw
func (w *WidgetBase) ScreenBounds() geometry.Rect       // USE THIS to anchor overlays/popups
func (w *WidgetBase) LocalToGlobal(p geometry.Point) geometry.Point
func (w *WidgetBase) GlobalToLocal(p geometry.Point) geometry.Point
func (w *WidgetBase) NeedsRedraw() bool
func (w *WidgetBase) SetNeedsRedraw(v bool)             // v=true propagates dirty UP to nearest RepaintBoundary
func (w *WidgetBase) MarkRedrawLocal()                  // dirty self only (continuous animations)
func (w *WidgetBase) ClearRedraw()
func (w *WidgetBase) AddBinding(b Unbinder)             // signal binding auto-cleanup on unmount
func (w *WidgetBase) AddEffect(e Stopper)
func (w *WidgetBase) CleanupBindings()
func (w *WidgetBase) SetRepaintBoundary(enabled bool)   // widget/boundary.go — opt into scene caching
func (w *WidgetBase) IsRepaintBoundary() bool
func (w *WidgetBase) IsMounted() bool / SetMounted(bool)
```

### 1.2 Lifecycle (`widget/lifecycle.go`)

```go
type Lifecycle interface {
    Mount(ctx Context)   // create signal bindings here via AddBinding()
    Unmount()            // cleanup not covered by AddBinding; CleanupBindings() runs automatically first
}
func MountTree(w Widget, ctx Context)    // sets parent chain + calls Mount, done by Window.SetRoot
func UnmountTree(w Widget)               // bottom-up; calls CleanupBindings then Unmount, clears parent
```

`MountTree` also performs `SetParent` on every child (Flutter adoptChild). If you construct children yourself **before** mounting (constructor), set the parent chain manually like `primitives.Box` does:

```go
type parentSetter interface{ SetParent(widget.Widget) }
if ps, ok := child.(parentSetter); ok { ps.SetParent(b) }
```

### 1.3 Context (`widget/context.go`) — what you get in every phase

```go
type Context interface {
    RequestFocus(w Widget); ReleaseFocus(w Widget); IsFocused(w Widget) bool; FocusedWidget() Widget
    Now() time.Time; DeltaTime() time.Duration
    Invalidate(); InvalidateRect(r geometry.Rect)
    Cursor() CursorType; SetCursor(cursor CursorType)
    Scale() float32
    ThemeProvider() ThemeProvider            // nil-check in headless mode
    OverlayManager() OverlayManager          // nil-check in headless mode
    WindowSize() geometry.Size
    Scheduler() SchedulerRef                 // SchedulerRef{ MarkDirty(w Widget) }
}
```

Cursor types: `CursorDefault, CursorPointer, CursorText, CursorCrosshair, CursorMove, CursorResizeNS, CursorResizeEW, CursorResizeNESW, CursorResizeNWSE, CursorNotAllowed, CursorWait, CursorNone`. Cursor is reset to default each frame; set it on `MouseEnter`, restore on `MouseLeave`.

Optional capability interfaces obtained by type-asserting `ctx` (the framework's "interface extension" pattern):
- `widget.PointerCapturer` — `CapturePointer(w Widget)` / `ReleasePointer(w Widget)` for drags (slider/sheet drag).
- `widget.AnimationScheduler` — `ScheduleAnimationFrame()` for deferred 30fps animation ticks instead of `InvalidateRect` (fallback shown in context.go docs).
- `widget.DrawStatsProvider`, `DirtyTrackerProvider`, `ImageCacheProvider`, `DirtyBoundaryRegistrar` — internals, rarely needed.

### 1.4 Canvas (`widget/canvas.go`) — drawing vocabulary

```go
type Canvas interface {
    Clear(color Color)
    DrawRect(r geometry.Rect, color Color)
    FillRectDirect(r geometry.Rect, color Color)
    StrokeRect(r geometry.Rect, color Color, strokeWidth float32)
    DrawRoundRect(r geometry.Rect, color Color, radius float32)
    StrokeRoundRect(r geometry.Rect, color Color, radius float32, strokeWidth float32)
    DrawCircle(center geometry.Point, radius float32, color Color)
    StrokeCircle(center geometry.Point, radius float32, color Color, strokeWidth float32)
    StrokeArc(center geometry.Point, radius float32, startAngle, sweepAngle float64, color Color, strokeWidth float32)
    DrawLine(from, to geometry.Point, color Color, strokeWidth float32)
    DrawText(text string, bounds geometry.Rect, fontSize float32, color Color, bold bool, align TextAlign)
    MeasureText(text string, fontSize float32, bold bool) float32
    DrawImage(img image.Image, at geometry.Point)
    PushClip(r geometry.Rect); PushClipRoundRect(r geometry.Rect, radius float32); PopClip()
    PushTransform(offset geometry.Point); PopTransform(); TransformOffset() geometry.Point
    ScreenOriginBase() geometry.Point
    ClipBounds() geometry.Rect
    ReplayScene(s *scene.Scene)
}
```

Optional canvas capabilities (type-assert):
```go
type ArcStroker interface { StrokeArcStyled(center geometry.Point, radius float32, startAngle, sweepAngle float64, color Color, strokeWidth float32, lineCap LineCap) }
type SVGFiller interface { FillSVGPath(svgData string, viewBox float32, bounds geometry.Rect, color Color) }
type SVGRenderer interface { RenderSVG(svgXML []byte, bounds geometry.Rect, color Color) }
type StyledTextDrawer interface {
    DrawStyledText(text string, bounds geometry.Rect, style TextStyle)
    MeasureStyledText(text string, style TextStyle) float32
}
type TextModeController interface { SetTextMode(mode TextMode); TextMode() TextMode }
// transition.OpacityPusher: PushOpacity(opacity float64) / PopOpacity()
```

Colors: `widget.Color{R,G,B,A float32}` with constructors `RGBA(r,g,b,a float32)`, `RGB`, `RGBA8(r,g,b,a uint8)`, `RGB8`, `Hex(0xRRGGBB)`, `HexA(0xRRGGBBAA)`; methods `WithAlpha(a float32)`, `Lerp(other Color, t float32)`, `IsOpaque/IsTransparent`. Constants `ColorTransparent, ColorBlack, ColorWhite, ...`.

### 1.5 Coordinate model & child drawing — CRITICAL

- A widget's `Bounds()` is in **parent-local** coordinates (parents call `child.SetBounds` during their Layout).
- In `Draw`, a container does `canvas.PushTransform(bounds.Min)`, then for each child: `widget.StampScreenOrigin(child, canvas)` then `widget.DrawChild(child, ctx, canvas)`, then `canvas.PopTransform()`.
- In `Event`, mouse/wheel positions must be translated into child space the same way (`local.Position = e.Position.Sub(b.Bounds().Min)`), iterating children **top-most first** (reverse order). See `BoxWidget.dispatchMouseEvent`.

```go
// widget/stamp.go
func StampScreenOrigin(child Widget, canvas Canvas)
// widget/draw.go — use instead of child.Draw() so RepaintBoundary children get scene caching
func DrawChild(child Widget, ctx Context, canvas Canvas)
func DrawTree(w Widget, ctx Context, canvas Canvas) DrawStats   // root-level draw (offscreen renderer uses this)
```

### 1.6 Minimal widget template (synthesizing the Button pattern from `core/button/event.go`)

```go
type Badge struct {
    widget.WidgetBase
    label   string
    hovered bool
    pressed bool
}

func NewBadge(label string) *Badge {
    b := &Badge{label: label}
    b.SetVisible(true)
    b.SetEnabled(true)
    return b
}

func (b *Badge) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
    size := c.Constrain(geometry.Sz(measuredW, 22))
    b.SetBounds(geometry.FromPointSize(b.Position(), size))
    return size
}

func (b *Badge) Draw(ctx widget.Context, canvas widget.Canvas) {
    if !b.IsVisible() { return }
    bounds := b.Bounds()
    canvas.DrawRoundRect(bounds, bg, 11)
    canvas.DrawText(b.label, bounds, 12, fg, true, widget.TextAlignCenter)
}

func (b *Badge) Event(ctx widget.Context, e event.Event) bool {
    me, ok := e.(*event.MouseEvent)
    if !ok { return false }
    switch me.MouseType {
    case event.MouseEnter:
        b.hovered = true
        ctx.SetCursor(widget.CursorPointer)
        b.SetNeedsRedraw(true)
        ctx.InvalidateRect(b.Bounds())
        return true
    case event.MouseLeave:
        b.hovered = false
        ctx.SetCursor(widget.CursorDefault)
        b.SetNeedsRedraw(true)
        ctx.InvalidateRect(b.Bounds())
        return true
    case event.MousePress:
        if me.Button != event.ButtonLeft { return false }
        b.pressed = true
        ctx.RequestFocus(b)
        b.SetNeedsRedraw(true); ctx.InvalidateRect(b.Bounds())
        return true
    case event.MouseRelease:
        wasPressed := b.pressed
        b.pressed = false
        b.SetNeedsRedraw(true); ctx.InvalidateRect(b.Bounds())
        if wasPressed && b.Bounds().Contains(me.Position) { /* fire onClick */ }
        return true
    }
    return false
}

func (b *Badge) Children() []widget.Widget { return nil }
```

Notes:
- **Hover** is delivered as `MouseEnter`/`MouseLeave` by the Window's hover tracker (`Window.updateHover` does hit-testing and synthesizes Enter/Leave); you do NOT diff MouseMove yourself.
- **Redraw request** = `w.SetNeedsRedraw(true)` (marks dirty, propagates to nearest repaint boundary) + `ctx.InvalidateRect(w.Bounds())` (requests a frame). `ctx.Invalidate()` = full-window.
- **Focus participation** (`widget/focusable.go`): WidgetBase already has `SetFocused/IsFocused`; add only:

```go
type Focusable interface {
    IsFocusable() bool
    SetFocused(focused bool)
    IsFocused() bool
}
// concrete widgets typically:
func (b *Badge) IsFocusable() bool { return b.IsEnabled() && b.IsVisible() }
```
Tab/Shift+Tab traversal is handled by `focus.Manager` (owned by Window); keyboard events arrive at your `Event` when focused — check `w.IsFocused()` before handling `event.KeyEnter`/`event.KeySpace` (`KeyEvent{Key event.Key; KeyType event.KeyType; Rune rune}`, types `event.KeyPress`/`event.KeyRelease`).
- Window-level shortcuts: `Window.FocusManager().RegisterShortcut(focus.Shortcut{Key: event.KeyK, Ctrl: true}, func(){...})` — exactly what a Command palette needs. `focus.Shortcut{Key event.Key; Ctrl, Shift, Alt bool}`.
- Compile-time checks convention: `var ( _ widget.Widget = (*Badge)(nil); _ widget.Focusable = (*Badge)(nil); _ widget.Lifecycle = (*Badge)(nil) )`.

### 1.7 Events (`event/`)

```go
// MouseEvent: MouseType (MousePress/MouseRelease/MouseMove/MouseEnter/MouseLeave/MouseDrag/MouseDoubleClick),
//   Button (ButtonLeft/Right/Middle/X1/X2), Buttons ButtonState, Position geometry.Point (widget-local),
//   GlobalPosition geometry.Point, ClickCount int
// WheelEvent: Position, DeltaX/DeltaY
// KeyEvent: KeyType (KeyPress/KeyRelease), Key (KeyEnter=400.., KeyTab, KeyEscape, KeySpace, …), Rune, Modifiers()
// FocusEvent: gained/lost
```

---

## 2. Composition: Card = styled Box

`primitives.BoxWidget` (`primitives/box.go`) is the workhorse container (implements `widget.Widget`, `a11y.Accessible`, `widget.Lifecycle`). Constructors:

```go
func Box(children ...widget.Widget) *BoxWidget    // vertical by default
func HBox(children ...widget.Widget) *BoxWidget   // horizontal
func VBox(children ...widget.Widget) *BoxWidget   // alias of Box
```

Fluent builder methods (all return `*BoxWidget`):

```go
Padding(v float32) / PaddingXY(x, y float32) / PaddingTop/Right/Bottom/Left(v float32)
Background(c widget.Color)
Rounded(r float32)                              // corner radius
BorderStyle(width float32, color widget.Color)
ShadowLevel(level int)                          // 0–5 Material elevation (multi-layer fake gaussian)
Gap(v float32)                                  // spacing between children along main axis
CrossAlign(a CrossAxisAlignment)                // CrossAxisStart | CrossAxisCenter | CrossAxisEnd | CrossAxisStretch
SetDirection(d Direction)                       // DirectionVertical | DirectionHorizontal
DirectionSignal(sig state.ReadonlySignal[Direction])
Width(v float32) / Height(v float32)            // explicit size
MinWidthValue / MinHeightValue / MaxWidthValue / MaxHeightValue(v float32)
Label(label string)                             // a11y label
```

Behavior facts you'll rely on:
- `BoxStyle{Padding geometry.Insets; Background widget.Color; Radius float32; Border Border; Shadow Shadow; Gap float32; ExplicitWidth/Height, Min/Max...}`.
- Children are clipped (`PushClipRoundRect`) only when the box has a border or radius > 0. Border draws **after** children (on top).
- `primitives.Expanded(child widget.Widget) *ExpandedWidget` marks a child to fill remaining main-axis space; multiple Expanded children split it equally (two-pass layout). The marker is the private `layoutExpander` interface — only `primitives` can implement it, so wrap, don't duck-type.
- Default `crossAlign` is `CrossAxisStart` (note: zero value; constants are `CrossAxisStart, CrossAxisCenter, CrossAxisEnd, CrossAxisStretch`).
- Shadow presets (`primitives/style.go`): levels 1–5 map to 2–4 concentric rounded-rect layers; the doc comments say level 1 ≈ cards at rest, 2 ≈ raised cards, 3 ≈ menus, 4 ≈ dialogs, 5 ≈ tooltips/overlays.

So a shadcn Card is literally:

```go
card := primitives.Box(
    primitives.VBox(title, description).Gap(6),   // CardHeader
    content,                                       // CardContent
    primitives.HBox(actions...).Gap(8),            // CardFooter
).Padding(24).Gap(24).
  Background(theme.Colors.Surface).
  Rounded(12).
  BorderStyle(1, theme.Colors.Outline).
  ShadowLevel(1)
```

Other composition primitives:
- `primitives.ThemeScope(theme widget.ThemeProvider, children ...widget.Widget) *ThemeScopeWidget` — theme override for a subtree (dark dialog in a light app); nearest-wins like InheritedWidget.
- `primitives.Image(...)` (`primitives/image.go`) with `ImageFit` = `ImageFitContain/Cover/Fill/None` — Avatar building block.
- `layout` package has algorithm-style `FlexLayout`, `GridLayout` (`AutoTrack()/FixedTrack(size)/FractionTrack(fr)`, `SimpleGrid(numColumns int, gap float32) *GridLayout`), `StackLayout`/`ZStackLayout` — these operate on a `LayoutTree`/`NodeID` abstraction (a separate retained layout system), not on widgets directly. For widget composition, BoxWidget + Expanded covers nearly everything; for a Table you'll lay out cells manually or with fixed `Width()` columns.

### Geometry quick reference

```go
geometry.Pt(x, y float32) Point
geometry.Sz(width, height float32) Size
geometry.NewRect(x, y, width, height float32) Rect
geometry.FromPointSize(p Point, s Size) Rect
// Rect has Min/Max Points, Contains, Union, Expand, TranslateXY, Center, Width, Height, IsEmpty
geometry.Constraints{MinWidth, MaxWidth, MinHeight, MaxHeight float32}
geometry.Tight(size) / TightWidth / TightHeight / Loose(size) / Expand() / BoxConstraints(minW, maxW, minH, maxH)
(c Constraints) Constrain(size Size) Size / Loosen() / Tighten(size) / Deflate(insets Insets) / Enforce(other)
(c Constraints) HasBoundedWidth/HasBoundedHeight/IsTight/Biggest()/Smallest()
geometry.Infinity  // float32 max — unbounded max constraint
geometry.UniformInsets(all) / SymmetricInsets(horizontal, vertical) / InsetsLTRB / InsetsTRBL
(i Insets) Horizontal() / Vertical() / Size() / Add / Scale
```

---

## 3. Text

`primitives.TextWidget` (`primitives/text.go`):

```go
func Text(content string) *TextWidget                  // static; default 14px, ColorBlack, LineHeight 1.2
func TextFn(fn func() string) *TextWidget              // reactive via function
func (t *TextWidget) ContentSignal(sig state.ReadonlySignal[string]) *TextWidget  // highest priority

// fluent:
FontSize(size float32); Color(c widget.Color); Bold(); Italic()
FontFamily(name string)            // custom registered font; falls back to embedded Inter
Align(a TextAlign)                 // TextAlignStart/Center/End (= widget.TextAlignLeft/Center/Right)
MaxLines(n int)                    // 0 = unlimited
Ellipsis()                         // sets Overflow = TextOverflowEllipsis
LineHeight(v float32)
```

- Color resolution: explicit `Color()` > `ctx.ThemeProvider().OnSurface()` > black. So leave color unset for theme-aware text.
- Weights: only **Regular vs Bold** via the base `Canvas.DrawText(text, bounds, fontSize, color, bold, align)`. Full weight range (CSS 100–900) only via `widget.StyledTextDrawer.DrawStyledText` with `widget.TextStyle{FontFamily string; FontSize float32; Bold bool; Italic bool; Color Color; Align TextAlign}` — but the public `font.Weight` matching exists in the registry; the TextWidget surface only exposes Bold/Italic. For Geist weights 400/500/600 you can register separate families ("Geist", "Geist Medium", "Geist SemiBold") or register weights and call `DrawStyledText` from your own widget.
- Wrapping: `TextWidget.measureText` wraps with a heuristic (`charW = fontSize * 0.6`) when MaxWidth is bounded; the canvas does real measurement at draw. For precise measurement in your own widgets, use `canvas.MeasureText(text, fontSize, bold) float32` (only available during Draw) or `MeasureStyledText`.
- Layout returns `lines * fontSize * lineHeight` height; `MaxLines` caps lines, `Ellipsis()` truncates with "...".

### Font registration (custom TTF — yes, Geist works)

Public path = plugin asset loader wired to the global font registry (`plugin/context.go`, `plugin/assets.go`):

```go
//go:embed fonts/Geist-Regular.ttf
var geistRegular []byte

ctx := plugin.NewDefaultPluginContext()           // MemoryAssetLoader → render.GlobalFontRegistry()
err := ctx.Assets.LoadFont("Geist", geistRegular) // registers as weight Regular / style Normal

label := primitives.Text("Hello").FontFamily("Geist").FontSize(16)
```

`AssetLoader.LoadFont(name string, data []byte) error` registers TTF/OTF bytes. Note the default registerer registers only `font.Regular/font.Normal`; the underlying metadata registry (`theme/font`) supports full families:

```go
// theme/font/registry.go
func NewRegistry() *Registry
func (r *Registry) RegisterFamily(family Family)      // Family{Name string; Faces []Face}; Face{Weight Weight; Style Style; Data []byte}
func (r *Registry) Resolve(familyName string, weight Weight, style Style) ([]byte, bool)  // CSS font-matching algorithm
```

The render-side bridge `internal/render.GlobalFontRegistry().Register(family string, weight font.Weight, style font.Style, data []byte) error` is **internal**; the supported public hook for multiple weights/styles is `MemoryAssetLoader.SetFontRegisterer(FontRegisterer)` where `type FontRegisterer func(name string, data []byte) error` — or register Bold by loading a second family name. Default embedded font: **Inter** (Regular + Bold), family name `"Inter"`. Resolution chain: exact → CSS weight matching → Inter fallback.

---

## 4. Overlay / popover / floating system

### 4.1 Core mechanism

The Window owns an `overlay.Stack`. Widgets never import `overlay` for pushing — they use the `widget.OverlayManager` from Context:

```go
// widget/context.go
type OverlayManager interface {
    PushOverlay(w Widget, onDismiss func())
    PopOverlay()
    RemoveOverlay(w Widget)
}
```

Stack semantics (`overlay/overlay.go`): LIFO; events go to **top overlay first**; `Remove(o)` also removes everything above it; overlays draw after (on top of) the tree in bottom-to-top order; each overlay is laid out with `geometry.Tight(windowSize)`.

```go
type Overlay interface {
    widget.Widget
    Dismiss()
    Modal() bool    // modal overlays block ALL events from reaching the tree
}
```

Window's `windowOverlayManager.PushOverlay` (app/window.go:1065) wraps any plain `widget.Widget` so you don't have to implement `Overlay` yourself — but if your pushed widget DOES implement `Dismiss()/Modal()`, those are honored.

### 4.2 Backdrop + dismiss-outside + Escape: `overlay.Container`

```go
// overlay/container.go
func NewContainer(content widget.Widget, windowSize geometry.Size, opts ...ContainerOption) *Container
func WithOnDismiss(fn func()) ContainerOption
func WithModal(modal bool) ContainerOption   // modal => draws scrim RGBA(0,0,0,0.32) + consumes all events
```

Container behavior: content gets events first; `Escape` → `Dismiss()`; `MousePress` outside `content.Bounds()` → `Dismiss()`; `Modal()` returns the modal flag. This is your Dialog/Sheet chassis.

### 4.3 Anchored positioning

Two positioning APIs exist:

```go
// overlay/position.go — 4 placements + flip + clamp
type Placement uint8  // PlacementBelow, PlacementAbove, PlacementLeft, PlacementRight
func Position(placement Placement, anchorGlobal geometry.Rect, overlaySize geometry.Size,
    windowSize geometry.Size, gap float32) geometry.Point

// core/popover/placement.go — 12 placements (shadcn-style side+align) + auto-flip + clamp
type Placement uint8  // Bottom, BottomStart, BottomEnd, Top, TopStart, TopEnd,
                      // Left, LeftStart, LeftEnd, Right, RightStart, RightEnd
func CalculatePosition(placement Placement, triggerBounds geometry.Rect, overlaySize geometry.Size,
    windowSize geometry.Size, gap float32) geometry.Point
```

**Anchor rect must be screen-space**: use `w.ScreenBounds()` (stamped during Draw — correct even inside ScrollViews).

### 4.4 The canonical "show floating content" recipe (from `core/dropdown/widget.go` and `core/popover/popover.go`)

```go
func (w *MyTrigger) openMenu(ctx widget.Context) {
    om := ctx.OverlayManager()
    if om == nil { return }

    menu := buildMenuWidget()                                   // any widget.Widget
    size := menu.Layout(ctx, geometry.Loose(ctx.WindowSize()))  // measure natural size

    anchor := w.ScreenBounds()
    pos := overlay.Position(overlay.PlacementBelow, anchor, size, ctx.WindowSize(), 4)
    menu.SetBounds(geometry.FromPointSize(pos, size))           // pre-position BEFORE push

    om.PushOverlay(menu, func() { w.handleClosed(ctx) })        // onDismiss = click-outside/Escape
    w.SetNeedsRedraw(true)
}
// close: om.RemoveOverlay(menu); menu = nil; w.SetNeedsRedraw(true)
```

### 4.5 Ready-made building blocks

- **Tooltip** (`core/popover/tooltip.go`): `popover.NewTooltip(opts ...Option) *Tooltip` — hover-triggered with delay (default 500ms, checked in Layout via `ctx.Now()`), auto hide on leave/click. Options shared with Popover (`core/popover/options.go`): `TriggerWidget(w)`, `Content(w)`, `ContentSize(w,h)`, `TooltipText(s)`, `PlacementOpt(p)`, `Gap(g)`, `Delay(d)`, `DismissOnClickOutside(bool)`, `PainterOpt(p)`, `VisibleSignal(sig state.Signal[bool])`, `Disabled(bool)`, `DisabledFn(fn)`, `OnShow(fn)`, `OnHide(fn)`, `MaxWidth(w)`.
- **Popover** (`core/popover/popover.go`): `popover.NewPopover(opts ...Option) *Popover` — click-triggered, `Show(ctx)/Hide(ctx)/Toggle(ctx)/IsOpen()`; wraps content in an internal `overlayContent` that paints via a `Painter` interface (`PaintPopover(canvas, *PopoverPaintState)`) — supply your own painter for shadcn styling.
- **Dialog** (`core/dialog`): `dialog.New(opts ...Option) *Widget`, `Show(ctx)/Close(ctx)/IsOpen()`. Options: `Title(s)`, `TitleFn(fn)`, `TitleSignal(sig)`, `TitleReadonlySignal(sig)`, `Content(w widget.Widget)`, `Actions(actions ...Action)`, `DismissibleOpt(bool)`, `EscapeToCloseOpt(bool)`, `OnClose(fn)`, `MaxWidth(v)/MaxHeight(v)` (default max width 560), `PainterOpt(p)`. Internally pushes a `surfaceWidget` overlay.
- **Dropdown/Select** (`core/dropdown`): `dropdown.New(opts...)` with `Items(...)`, `Selected(idx)`, `OnChange(func(idx int, val string))`, `SignalOpt`, `PainterOpt`; opens a `menuWidget` overlay sized `triggerWidth × visibleCount*itemHeight`, positioned with `overlay.Position(overlay.PlacementBelow, ...)`, supports keyboard nav + wheel scrolling. Copy this for Select/Combobox.
- **Sheet**: build as `overlay.NewContainer(panel, ctx.WindowSize(), overlay.WithModal(true), overlay.WithOnDismiss(...))` where panel is a Box anchored to an edge, wrapped in `transition.Wrap(panel, transition.EnterEffect(transition.SlideIn(transition.FromRight)), ...)`.
- **Toast**: push a non-modal widget positioned at a corner via `om.PushOverlay`; note stack removal semantics (removing a middle overlay pops everything above), so for stacked toasts prefer ONE toast-region overlay that manages its own children.

Z-layering = stack order. Last pushed = topmost = first to receive events.

---

## 5. App skeleton

Exact minimal `main()` (verbatim pattern from `examples/hello/main.go`):

```go
package main

import (
    "log"
    _ "github.com/gogpu/gg/gpu" // enable GPU SDF acceleration

    "github.com/gogpu/gogpu"
    "github.com/gogpu/ui/app"
    "github.com/gogpu/ui/desktop"
    "github.com/gogpu/ui/primitives"
    "github.com/gogpu/ui/theme/material3"
    "github.com/gogpu/ui/widget"
)

func main() {
    m3 := material3.New(widget.Hex(0x6750A4))   // material3.NewDark(seed) for dark

    gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
        WithTitle("My App").
        WithSize(800, 600).
        WithContinuousRender(false))            // event-driven: 0% CPU idle

    uiApp := app.New(
        app.WithWindowProvider(gogpuApp),
        app.WithPlatformProvider(gogpuApp),
        app.WithEventSource(gogpuApp.EventSource()),
        app.WithTheme(m3.AsTheme()),
    )
    uiApp.SetRoot(primitives.Box(primitives.Text("Hello")).Padding(24))

    if err := desktop.Run(gogpuApp, uiApp); err != nil {   // blocks until window closes
        log.Fatal(err)
    }
}
```

Key App/Window API (`app/app.go`, `app/window.go`):

```go
func New(opts ...Option) *App
func WithWindowProvider(wp gpucontext.WindowProvider) Option
func WithPlatformProvider(pp gpucontext.PlatformProvider) Option
func WithEventSource(es gpucontext.EventSource) Option
func WithTheme(t *theme.Theme) Option
func WithRenderMode(mode RenderMode) Option
func (a *App) SetRoot(root widget.Widget)        // unmounts old tree, mounts new, makes root a RepaintBoundary
func (a *App) SetTheme(t *theme.Theme)           // triggers full relayout + redraw — THE dark-mode switch
func (a *App) Window() *Window; Theme() *theme.Theme; Scheduler() *state.Scheduler
func (a *App) Frame(); HandleEvent(e event.Event)
func (a *App) SetFrameCallback(cb FrameCallback) // FrameStats{LayoutDuration, DrawDuration, DrawSkipped, DrawStats…}
// Window:
func (w *Window) FocusManager() *ifocus.Manager  // RegisterShortcut etc.
func (w *Window) Overlays() *overlay.Stack
func (w *Window) Context() *widget.ContextImpl
func desktop.Run(gogpuApp *gogpu.App, uiApp *app.App) error
```

### Dark mode switching at runtime (exact pattern from `examples/gallery/main.go`)

```go
var onThemeChange func(int)
onThemeChange = func(idx int) {
    uiApp.SetTheme(pickTheme(idx))            // e.g. theme.DefaultDark() / material3.NewDark(seed).AsTheme()
    uiApp.SetRoot(buildUI(...))               // gallery rebuilds the tree so painters/colors refresh
    gogpuApp.RequestRedraw()
}
```

`App.SetTheme` alone marks the whole tree for redraw and updates `ctx.ThemeProvider()`; widgets that read theme colors at draw time (no cached colors) update without a rebuild. The gallery rebuilds because it bakes painter structs at construction. For a shadcn port: resolve colors from `ctx.ThemeProvider()` (or theme extensions) inside `Draw` and `SetTheme` is sufficient.

Theme model (`theme/`): `theme.Theme{Name string; Mode ThemeMode; Colors ColorPalette; Typography Typography; Spacing SpacingScale; Shadows ShadowStyles; Radii RadiusScale; Extensions map[string]any}`; presets `theme.DefaultLight() / DefaultDark() / DefaultHighContrast() / ForMode(mode)`; `ColorPalette` fields: `Primary, PrimaryLight, PrimaryDark, Secondary..., Background, Surface, SurfaceVariant, Error, Warning, Success, Info, OnPrimary, OnSecondary, OnBackground, OnSurface, OnError, Divider, Outline, Shadow`. `Theme` implements `widget.ThemeProvider{IsDark() bool; OnSurface() Color}`. Custom component tokens: `t.SetExtension(key string, value any)` / `GetExtension`, or typed `RegisterExtension(ext ThemeExtension)` + `TypedExtension(name)` — perfect place for shadcn CSS-variable-style tokens. `theme.Lighten/Darken/WithOpacity(c widget.Color, amount float32)` helpers.

---

## 6. Signals API (`state/`)

```go
// signal.go (aliases of github.com/coregx/signals)
type Signal[T any] = signals.Signal[T]                 // Get() T; Set(T); Update(...); Subscribe; AsReadonly()
type ReadonlySignal[T any] = signals.ReadonlySignal[T] // Get() T; Subscribe; SubscribeForever
func NewSignal[T any](initial T) Signal[T]
func NewSignalWithOptions[T any](initial T, opts Options[T]) Signal[T]   // Options{Equal EqualFunc[T], OnPanic}
func Subscribe[T any](sig ReadonlySignal[T], ctx context.Context, fn func(T)) Unsubscribe
func SubscribeForever[T any](sig ReadonlySignal[T], fn func(T)) Unsubscribe

// computed.go — deps must be passed explicitly
func NewComputed[T any](compute func() T, deps ...any) ReadonlySignal[T]
func NewComputedWithOptions[T any](compute func() T, opts Options[T], deps ...any) ReadonlySignal[T]

// effect.go
func NewEffect(fn func(), deps ...any) EffectRef                    // EffectRef has Stop()
func NewEffectWithCleanup(fn func() func(), deps ...any) EffectRef

// binding.go — connect signal → widget invalidation
func Bind[T any](sig ReadonlySignal[T], ctx widget.Context) *Binding              // DEPRECATED (full invalidate)
func BindToScheduler[T any](sig ReadonlySignal[T], w widget.Widget, sched widget.SchedulerRef) *Binding  // USE THIS
func BindToSchedulerFunc[T any](sig ReadonlySignal[T], shouldDirty func(T) bool, w widget.Widget, sched widget.SchedulerRef) *Binding
// Binding has Unbind() and IsActive()

// scheduler.go
func NewScheduler(flushFn func([]widget.Widget)) *Scheduler   // App creates this; you rarely do
func (s *Scheduler) MarkDirty(w widget.Widget); Flush(); Batch(fn func())
```

**The pattern every built-in widget uses** (copy it):

```go
// constructor: accept signal via option, store it
func (w *MyWidget) MySignal(sig state.ReadonlySignal[string]) *MyWidget { w.sig = sig; return w }

// Mount (widget.Lifecycle): bind → scheduler marks this widget dirty on change
func (w *MyWidget) Mount(ctx widget.Context) {
    sched := ctx.Scheduler()
    if sched == nil { return }            // headless
    if w.sig != nil {
        w.AddBinding(state.BindToScheduler(w.sig, w, sched))  // auto-Unbind on unmount
    }
}
func (w *MyWidget) Unmount() {}           // CleanupBindings() runs automatically before this

// Draw/Layout: read w.sig.Get() — value is always current
```

Flow: `sig.Set(v)` → subscription fires → `sched.MarkDirty(w)` → App's flush sets `w.SetNeedsRedraw(true)` + `RequestRedraw()` → next frame redraws just that boundary. Two-way binding = pass writable `state.Signal[T]` and call `sig.Set` from event handlers (see `checkbox.CheckedSignal`, `popover.VisibleSignal`).

---

## 7. Headless rendering → PNG (visual verification)

`offscreen.Renderer` (`offscreen/renderer.go`) — pure CPU, no window needed:

```go
func NewRenderer(width, height int, opts ...Option) *Renderer
func WithTheme(tp widget.ThemeProvider) Option     // default material3.New(0x6750A4) light
func WithScale(s float32) Option                   // 2.0 = retina
func WithBackground(c widget.Color) Option         // default transparent
func WithFitSize() Option                          // measure-then-render; pass 0,0 dims
func WithMaxSize(width, height int) Option
func (r *Renderer) Render(w widget.Widget)         // layout (Loose constraints) + DrawTree into buffer
func (r *Renderer) Image() *image.RGBA             // nil before Render
```

Exact PNG recipe:

```go
package main

import (
    "image/png"
    "os"

    "github.com/gogpu/ui/offscreen"
    "github.com/gogpu/ui/primitives"
    "github.com/gogpu/ui/theme/material3"
    "github.com/gogpu/ui/widget"
)

func main() {
    card := primitives.Box(
        primitives.Text("Card Title").FontSize(18).Bold(),
        primitives.Text("Card description goes here."),
    ).Padding(24).Gap(8).Background(widget.ColorWhite).Rounded(12).BorderStyle(1, widget.RGBA8(228, 228, 231, 255)).ShadowLevel(1)

    r := offscreen.NewRenderer(0, 0,
        offscreen.WithFitSize(),
        offscreen.WithMaxSize(800, 600),
        offscreen.WithTheme(material3.New(widget.Hex(0x6750A4))),
        offscreen.WithBackground(widget.ColorWhite),
    )
    r.Render(card)

    f, _ := os.Create("card.png")
    defer f.Close()
    _ = png.Encode(f, r.Image())
}
```

For fixed-size snapshots: `offscreen.NewRenderer(400, 300)` and `Loose` constraints apply. Interaction-state snapshots: drive events first with `uitest` then render.

`uitest` (logic-level testing, no pixels):

```go
func LayoutWidget(w widget.Widget, width, height float32) geometry.Size
func LayoutWidgetTight(w widget.Widget, width, height float32) geometry.Size
func DrawWidget(w widget.Widget) *MockCanvas                         // records DrawRectCalls, DrawTextCalls, …
func DrawWidgetWithContext(w widget.Widget, ctx widget.Context) *MockCanvas
func SimulateClick(w widget.Widget, x, y float32) bool
func SimulateClickWithContext(w widget.Widget, ctx widget.Context, x, y float32) bool
func SimulateKeyPress(w widget.Widget, code event.Key) bool
func NewMockContext() *MockContext        // full widget.Context impl; Reset(); invalidation/cursor assertions
// event builders: Click, Release, DoubleClick, RightClick, MouseMove, MouseEnter, MouseLeave, MouseDrag,
//   KeyPress, KeyRelease, KeyType, WheelScroll, WheelScrollH, FocusGained, FocusLost
// asserts: AssertDrawnText, AssertNoDrawnText, AssertRectDrawn, AssertInvalidated, AssertCursor,
//   AssertFocused, AssertNotFocused, AssertColorEqual
```

---

## 8. Animations

### 8.1 Tween engine (`animation/`)

```go
func To(signal signalFloat32, target float32) *AnimationBuilder   // signalFloat32 = interface{ Get() float32; Set(float32) }
func (b *AnimationBuilder) From(value float32) *AnimationBuilder  // default: signal.Get() at Start
func (b *AnimationBuilder) Duration(d time.Duration) *AnimationBuilder   // default 300ms (DurationMedium2)
func (b *AnimationBuilder) Delay(d time.Duration) *AnimationBuilder
func (b *AnimationBuilder) Ease(e Easing) *AnimationBuilder       // default M3Standard
func (b *AnimationBuilder) Repeat(count int) *AnimationBuilder    // -1 = infinite
func (b *AnimationBuilder) AutoReverse() *AnimationBuilder
func (b *AnimationBuilder) OnDone(fn func()) *AnimationBuilder
func (b *AnimationBuilder) Start(ctrl *Controller) *Animation     // auto-cancels prior anim on same signal

func NewController() *Controller
func (c *Controller) Tick(dt time.Duration) bool    // returns true while animations active
func (c *Controller) HasActive() bool; CancelAll(); Cancel(signal signalFloat32)

// springs
func SpringTo(signal signalFloat32, target float32) *SpringBuilder
//   .Stiffness(k).DampingRatio(ratio).Mass(m).InitialVelocity(v).OnDone(fn).Start(ctrl)
// composition
func NewSequence(items ...Startable) *SequenceBuilder   // .OnDone(fn).Start(ctrl)
func NewParallel(items ...Startable) *ParallelBuilder
// type-safe tweens for colors etc.
func NewColorTween(begin, end widget.Color) *Tween[widget.Color]; (tw *Tween[T]) At(t float32) T
func LerpColor(begin, end widget.Color, t float32) widget.Color
```

Easings: `Linear, EaseInQuad, EaseOutQuad, EaseInOutQuad, EaseInCubic, EaseOutCubic, EaseInOutCubic`, plus M3: `M3Standard, M3StandardAccelerate, M3StandardDecelerate, M3Emphasized, M3EmphasizedAccelerate, M3EmphasizedDecelerate`, `CubicBezier(x1,y1,x2,y2)`. Duration tokens `DurationShort1(50ms) … DurationExtraLong4(1000ms)`. Presets (`animation/presets.go`): `FadeIn/FadeOut(signal, duration)`, `SlideInFromBottom/Top/Left/Right(signal, distance, duration)`, `ScaleIn/ScaleOut`, `DialogEnter/DialogExit(opacity, scale)`, `MenuEnter/MenuExit(opacity, translateY, distance)`, `SnackbarEnter/SnackbarExit`.

### 8.2 Accordion height — the EXACT pattern (from `core/collapsible/collapsible.go`)

The widget owns a `*animation.Controller` and a `progress float32`, ticks the controller inside `Layout` using `ctx.DeltaTime()`, and scales content height by progress:

```go
// start (on toggle):
adapter := &progressAdapter{w: w}             // adapter implements Get()/Set(float32) over w.progress
animation.To(adapter, target).                // target 0 or 1
    From(w.progress).
    Duration(200 * time.Millisecond).
    Ease(animation.EaseInOutCubic).
    Start(w.animCtrl)

// in Layout(ctx, constraints):
w.tickAnimation(ctx)                          // animCtrl.Tick(clamped ctx.DeltaTime()); if HasActive: SetNeedsRedraw(true) + ctx.Invalidate()
totalH := headerH + contentH*w.progress       // animated height
return constraints.Constrain(geometry.Sz(constraints.MaxWidth, totalH))

// in Draw: clip content to animated height
clipRect := geometry.NewRect(bounds.Min.X, contentTop, bounds.Width(), w.contentSize.Height*w.progress)
canvas.PushClip(clipRect); content.Draw(ctx, canvas); canvas.PopClip()
```

The `ctx.Invalidate()` while `HasActive()` keeps frames pumping (Window enters continuous-render mode until animations finish).

### 8.3 Fade / slide / scale wrapper: `transition.Wrap` (`transition/`)

```go
func Wrap(child widget.Widget, opts ...Option) *Transition
func EnterEffect(e Effect) Option; func ExitEffect(e Effect) Option
func Duration(d time.Duration) Option              // default 250ms
func Easing(e animation.Easing) Option             // default EaseOutCubic
func (t *Transition) Show(); Hide(); IsShown() bool; IsAnimating() bool

// effects (transition/effects.go):
func FadeIn() Effect; FadeOut() Effect
func SlideIn(dir Direction) Effect; SlideOut(dir Direction) Effect  // FromTop/FromBottom/FromLeft/FromRight, To*
func ScaleIn() Effect   // 0.8→1.0 + fade
func ScaleOut() Effect  // 1.0→0.8 + fade
func None() Effect
```

- Opacity uses the optional canvas capability `transition.OpacityPusher{PushOpacity(opacity float64); PopOpacity()}` — silently skipped if unsupported.
- Slide uses `canvas.PushTransform` by a fraction of widget size (enter: offset→0; exit: 0→offset).
- Dialog fade-in: `transition.Wrap(dialogSurface, transition.EnterEffect(transition.ScaleIn()), transition.ExitEffect(transition.ScaleOut()), transition.Duration(200*time.Millisecond))` and call `.Show()` when pushing the overlay.
- Sheet slide-in: `transition.EnterEffect(transition.SlideIn(transition.FromRight))` + `ExitEffect(transition.SlideOut(transition.ToRight))`; on close, call `.Hide()` and remove the overlay in an `OnDone`/after `IsAnimating()` turns false (or after Duration via the exit effect completing — `Transition` sets `shown=false` when exit completes).

---

## 9. SVG icons (lucide is feasible)

`icon` package. An icon is data, not a widget:

```go
type IconData struct {
    Name        string
    ViewBox     float32     // square coordinate space (24 for lucide)
    Ops         []PathOp    // parsed path commands
    StrokeWidth float32     // viewbox units; 0 → default 1.5
    SVGData     string      // raw path data for fill rendering (SVGFiller)
    SVGXML      []byte      // full SVG XML for bitmap rasterization (SVGRenderer)
}

// constructors (icon/svg.go):
func FromSVG(name string, viewBox float32, svgPathData string) IconData         // parses path `d`; panics on bad data
func FromSVGStroke(name string, viewBox float32, svgPathData string) IconData   // stroked (outline) rendering
func FromSVGXML(name string, svgXML []byte) IconData                            // full SVG file → rasterized bitmap, pixel-perfect
func TryFromSVG(name string, viewBox float32, svgPathData string) (IconData, error)
// manual paths: Move(x,y), Line(x,y), Cubic(...), Quad(...), ClosePath()
```

Registries (`icon/registry.go`):

```go
func DefaultRegistry() *Registry           // global, pre-populated with built-ins (Close, Check, ChevronDown, Search, …)
func NewRegistry() *Registry
func (r *Registry) Register(icon IconData)            // keyed by Name, replaces existing
func (r *Registry) Get(name string) (IconData, bool)
// multi-color variants: MultiColorIcon{Groups []PathGroup}, Palette map[string]widget.Color, DefaultMultiColorRegistry()
```

Drawing (`icon/draw.go`, `icon/widget.go`):

```go
func Draw(canvas widget.Canvas, data IconData, bounds geometry.Rect, color widget.Color)
// rendering priority: SVGXML→SVGRenderer  >  SVGData→SVGFiller (filled path)  >  Ops stroked lines

func NewIcon(data IconData, opts ...IconOption) *IconWidget   // square widget, default 24px
func Size(s float32) IconOption
func Color(color widget.Color) IconOption
func ColorSignal(sig state.ReadonlySignal[widget.Color]) IconOption
func Label(text string) IconOption                            // a11y
```

Registering lucide icons — two viable routes:

```go
// Route A: path data (fill rendering — best for solid lucide-like icons)
var ChevronsUpDown = icon.FromSVG("chevrons-up-down", 24, "m7 15 5 5 5-5M7 9l5-5 5 5")

// Route B (recommended for true lucide stroke style): full SVG XML, pixel-perfect with stroke-linecap etc.
//go:embed lucide/check.svg
var checkSVG []byte
var LucideCheck = icon.FromSVGXML("lucide-check", checkSVG)

icon.DefaultRegistry().Register(LucideCheck)
w := icon.NewIcon(LucideCheck, icon.Size(16), icon.Color(widget.RGBA8(113,113,122,255)))
```

Caveat: lucide icons are stroke-based; `FromSVG` route fills the path (wrong for strokes), `FromSVGStroke` strokes the outline ops; `FromSVGXML` honors stroke/linecap/fill-rule via `gg/svg.RenderWithColor` and is the safest.

---

## 10. Accessibility basics (`a11y/`)

```go
type Accessible interface {
    AccessibilityRole() Role
    AccessibilityLabel() string
    AccessibilityHint() string
    AccessibilityValue() string
    AccessibilityState() State    // State{Disabled, Hidden, Selected, Checked, Expanded, ...}
    AccessibilityActions() []Action
}
```
Roles you'll need: `RoleButton, RoleCheckbox, RoleRadio, RoleTextField, RoleSlider, RoleSwitch, RoleComboBox, RoleLabel, RoleImage, RoleProgressBar, RoleTooltip, RoleAlert, RoleBadge, RoleHeading, RoleDialog, RoleAlertDialog, RoleMenu, RoleMenuItem, RoleGenericContainer, RoleGroup, RoleSeparator, RoleToolbar, ...`. `a11y.NewNodeFromAccessible(a Accessible) *Node` builds tree nodes. Every primitive implements Accessible; do the same in new components (Box delegates pattern is in `primitives/expanded.go`).

---

## 11. Gotchas / conventions checklist for new components

1. Constructor: `SetVisible(true); SetEnabled(true)` (zero value is invisible/disabled) and set parent chain on stored children.
2. `Layout` must call `SetBounds` on each child (parent-local coords) AND on itself, and return a size satisfying constraints (`constraints.Constrain(...)`).
3. `Draw`: check `IsVisible()`; `PushTransform(bounds.Min)` before children; `StampScreenOrigin(child, canvas)` before each child draw; use `widget.DrawChild` not `child.Draw` in containers.
4. `Event`: translate mouse/wheel positions for children; iterate reverse; consume by returning true; broadcast non-mouse events.
5. Visual state change = `SetNeedsRedraw(true)` + `ctx.InvalidateRect(w.Bounds())`. Layout-affecting change = also `ctx.Invalidate()`.
6. Signals: bind in `Mount` via `state.BindToScheduler` + `AddBinding`; never subscribe raw in constructors.
7. Overlays: anchor with `ScreenBounds()`, measure with `Layout(ctx, geometry.Loose(ctx.WindowSize()))`, `SetBounds` before `PushOverlay`.
8. Painter pattern: every `core/*` widget separates logic from painting via a `Painter` interface + `*PaintState` struct (e.g. `dropdown.TriggerPaintState{Bounds, SelectedText, Open, Focused, Hovered, Disabled}`). Follow it so shadcn styling is one painter per design system, themable later.
9. Heavy/animated subtrees: `w.SetRepaintBoundary(true)` for scene caching; the root gets it automatically in `SetRoot`.
10. Functional options (`func Option(*config)`) are the constructor convention across `core/*`; fluent builders are the convention in `primitives`.
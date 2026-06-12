# graft — Go API Design Proposal (shadcn/ui port on gogpu/ui)

Repo state: `C:\Users\tingzhen\Documents\GitHub\graft` is empty (README + LICENSE), so this is a greenfield API design.

## 1. Research summary

**shadcn/ui live inventory** (verified against https://ui.shadcn.com/docs/components, June 2026 — 59 entries): accordion, alert, alert-dialog, aspect-ratio, avatar, badge, breadcrumb, button, button-group, calendar, card, carousel, chart, checkbox, collapsible, combobox, command, context-menu, data-table, date-picker, dialog, direction, drawer, dropdown-menu, empty, field, hover-card, input, input-group, input-otp, item, kbd, label, menubar, native-select, navigation-menu, pagination, popover, progress, radio-group, resizable, scroll-area, select, separator, sheet, sidebar, skeleton, slider, sonner, spinner, switch, table, tabs, textarea, toast, toggle, toggle-group, tooltip, typography. Notes: `form` (react-hook-form) has been superseded by the **Field** family as the primary form-layout story; `toast` is legacy in favor of **sonner**; `direction` (RTL provider) and `typography` are utilities, not widgets; `combobox`, `data-table`, and `date-picker` are documented as composition *patterns*, not standalone registry primitives.

**Key usage patterns observed** (fetched: button, card, dialog, select, dropdown-menu, tabs, accordion, field, theming):
- Flat prop-driven leaves: `<Button variant="outline" size="sm">Cancel</Button>`; Button has 6 variants (default/secondary/destructive/outline/ghost/link) and sizes xs/sm/default/lg + icon sizes.
- Anatomy-driven containers: `Card > CardHeader (CardTitle, CardDescription, CardAction) > CardContent > CardFooter`; `Dialog > DialogTrigger + DialogContent > DialogHeader/DialogFooter`; `Tabs > TabsList > TabsTrigger + TabsContent (matched by value)`; `Accordion(type="single" collapsible) > AccordionItem(value) > AccordionTrigger + AccordionContent`.
- Controlled state: `value` + `onValueChange`, `open` + `onOpenChange`, or uncontrolled `defaultValue`.
- Theming: semantic background/foreground CSS-variable pairs (background, card, popover, primary, secondary, muted, accent + `-foreground`; destructive, border, input, ring; chart-1..5; sidebar family; `radius` scale), base-color presets (Neutral, Stone, Zinc, Mauve, Olive, Mist, Taupe), dark mode = token override set.

**gogpu/ui provides**: retained-mode GPU toolkit, per-widget packages with functional options (`button.New(button.TextSignal(label))`, `slider.New(slider.OnChange(...))`), `primitives.Box(children...).Padding(16).Background(...)` fluent layout, reactivity via `coregx/signals` (`TextFn(func() string)`), and pluggable design systems (Material 3, Fluent, Cupertino, JetBrains painter sets). **graft therefore has a natural architectural slot: graft = a 5th design system (the "shadcn painter set" with the token model below) + a higher-level shadcn-shaped composition API on top.**

## 2. Complete component table

Class: **A** = wrap/retheme existing gogpu/ui widget · **B** = composite of primitives (layout+style only) · **C** = new interactive widget needed · **D** = not applicable to native desktop (Go-idiomatic equivalent or omission noted).

| Component | Anatomy (sub-parts) | Class | gogpu/ui base | Tier |
|---|---|---|---|---|
| Accordion | Accordion / AccordionItem / AccordionTrigger / AccordionContent (single \| multiple, collapsible) | A | collapsible (group controller added) | 2 |
| Alert | Alert / AlertTitle / AlertDescription (+icon) | B | primitives Box/Text | 1 |
| Alert Dialog | Trigger / Content / Header / Title / Description / Footer / Action / Cancel | A | dialog (modal, no outside-dismiss) | 1 |
| Aspect Ratio | single wrapper | B | Box with ratio constraint | 3 |
| Avatar | Avatar / AvatarImage / AvatarFallback | B | image primitive + Box (circle clip) | 2 |
| Badge | single (variants: default/secondary/destructive/outline) | B | Box + Text | 1 |
| Breadcrumb | Breadcrumb / List / Item / Link / Page / Separator / Ellipsis | B | Box row + button(ghost) + dropdown | 3 |
| Button | single (6 variants, xs/sm/default/lg + icon sizes, loading w/ Spinner) | A | button | 1 |
| Button Group | ButtonGroup / ButtonGroupSeparator / ButtonGroupText | B | Box row (border-collapse styling) | 2 |
| Calendar | single (mode: single/multiple/range; month nav) | C | none — new date-grid widget | 2 |
| Card | Card / CardHeader / CardTitle / CardDescription / CardAction / CardContent / CardFooter | B | primitives Box/Text | 1 |
| Carousel | Carousel / Content / Item / Previous / Next | B | scrollview + snap logic + buttons | 3 |
| Chart | ChartContainer / ChartTooltip / ChartLegend (+config tokens chart-1..5) | A | linechart (bar/area painters later) | 3 |
| Checkbox | single (+ indeterminate) | A | checkbox | 1 |
| Collapsible | Collapsible / CollapsibleTrigger / CollapsibleContent | A | collapsible | 2 |
| Combobox | pattern: Popover + Command (filterable list) | C | popover + listview + textfield (new filter behavior) | 2 |
| Command | Command / CommandDialog / Input / List / Empty / Group / Item / Shortcut / Separator | C | dialog + listview + textfield (fuzzy filter, kb nav) | 3 |
| Context Menu | Trigger / Content / Item / CheckboxItem / RadioGroup+RadioItem / Label / Separator / Shortcut / Sub | A | contextmenu | 2 |
| Data Table | pattern: Table + column defs, sorting, selection, pagination | A | datatable | 2 |
| Date Picker | pattern: Popover + Calendar + Button | B | popover + graft Calendar | 2 |
| Dialog | Trigger / Content / Header / Title / Description / Footer / Close | A | dialog | 1 |
| Direction | RTL provider | D | defer to gogpu/ui text layer; expose theme token later, omit v1 | — |
| Drawer | Trigger / Content / Header / Title / Description / Footer / Close (vaul, drag-to-dismiss) | D | mobile-first; folded into Sheet (side="bottom") on desktop | — |
| Dropdown Menu | Trigger / Content / Group / Label / Item / CheckboxItem / RadioGroup+RadioItem / Separator / Shortcut / Sub | A | menu/dropdown | 1 |
| Empty | Empty / Header / Media / Title / Description / Content | B | Box/Text | 3 |
| Field | Field / FieldLabel / FieldDescription / FieldError / FieldGroup / FieldSet / FieldLegend / FieldSeparator | B | Box/Text (+ validation glue, see Form) | 1 |
| Form (react-hook-form) | Form / FormField / FormItem / FormLabel / FormControl / FormMessage | D | RHF is React-specific. Go equivalent: `graft.Form` — signal-backed values + `func(T) error` validators feeding `FieldError` (see §3.7) | 1 |
| Hover Card | HoverCard / Trigger / Content | A | popover (hover trigger + delay) | 3 |
| Input | single | A | textfield | 1 |
| Input Group | InputGroup / Addon / Button / Input / Text / Textarea | B | Box row wrapping textfield | 2 |
| Input OTP | InputOTP / Group / Slot / Separator | C | new segmented-input (focus handoff, paste) | 3 |
| Item | Item / Media / Content / Title / Description / Actions / Group / Header / Footer / Separator | B | Box/Text rows | 2 |
| Kbd | Kbd / KbdGroup | B | Box + Text | 3 |
| Label | single (click focuses control) | B | Text + focus link | 1 |
| Menubar | Menubar / Menu / Trigger / Content / Item / CheckboxItem / RadioItem / Sub / Separator / Shortcut | A | menubar | 2 |
| Native Select | styled `<select>` | D | no "native" control in a GPU toolkit — collapses into Select; omitted | — |
| Navigation Menu | NavigationMenu / List / Item / Trigger / Content / Link / Indicator / Viewport | D | website mega-nav; desktop apps use Menubar / Sidebar / Tabs instead; revisit only if web-style apps demand it | — |
| Pagination | Pagination / Content / Item / Link / Previous / Next / Ellipsis | B | button row | 3 |
| Popover | Popover / Trigger / Content / Anchor | A | popover | 1 |
| Progress | single (determinate) | A | progress/progressbar | 1 |
| Radio Group | RadioGroup / RadioGroupItem | A | radio | 1 |
| Resizable | ResizablePanelGroup / ResizablePanel / ResizableHandle | A | splitview | 2 |
| Scroll Area | ScrollArea / ScrollBar | A | scrollview (overlay scrollbar painter) | 1 |
| Select | Select / Trigger / Value / Content / Group / Label / Item / Separator / ScrollUp/Down | A | dropdown | 1 |
| Separator | single (h/v) | B | Box (1px line) | 1 |
| Sheet | Sheet / Trigger / Content(side) / Header / Title / Description / Footer / Close | C | dialog overlay layer + edge-anchored slide panel | 2 |
| Sidebar | SidebarProvider / Sidebar / Header / Content / Footer / Group(+Label/Action/Content) / Menu(+Item/Button/Action/Badge/Sub) / Trigger / Rail / Inset / Input / Separator | B | splitview/docking + collapsible + Box (large composite) | 2 |
| Skeleton | single (shimmer) | B | Box + animation (possibly `stripe`) | 2 |
| Slider | single (+range, multiple thumbs) | A | slider | 1 |
| Sonner (Toast) | Toaster (mount once) + `toast()` imperative API | C | new overlay stack manager (popover layer + timers) | 2 |
| Spinner | single | B | animated icon primitive | 1 |
| Switch | single | A | checkbox semantics + switch painter (thumb anim) | 1 |
| Table | Table / Header / Body / Footer / Row / Head / Cell / Caption | B | gridview / Box grid (static styled) | 2 |
| Tabs | Tabs / TabsList / TabsTrigger / TabsContent (value-matched) | A | tabview | 1 |
| Textarea | single | A | textfield (multiline mode; class C if gogpu textfield is single-line only) | 1 |
| Toast (legacy) | — | D | superseded by Sonner upstream; graft ships one Toast system | — |
| Toggle | single (pressed state; default/outline variants) | A | button (toggle semantics + painter) | 2 |
| Toggle Group | ToggleGroup / ToggleGroupItem (single \| multiple) | B | toggle + group controller | 2 |
| Tooltip | TooltipProvider / Tooltip / Trigger / Content | A | tooltip | 1 |
| Typography | h1–h4, p, blockquote, list, code, lead, large, small, muted | D→B | shipped as styled Text constructors (`graft.H1`, `graft.P`, `graft.Muted`, `graft.Code`...) | 1 |

Totals: A=22, B=21, C=7 (Calendar, Combobox, Command, InputOTP, Sheet, Sonner/Toast, Carousel-snap borderline), D=7 (with Go equivalents for Form and Typography).

## 3. The graft API

### 3.1 One convention: flat package, shadcn names, chainable props, signal binding

1. **One package, `graft`** (plus `graft/icons` and `graft/lucide` later). Per-component packages (`button.New`, `card.New`) were rejected: 15-line import blocks, stuttering names, and broken mental mapping from shadcn docs. With one package, every shadcn tag maps 1:1: `<CardHeader>` → `graft.CardHeader(...)`, `<DropdownMenuItem>` → `graft.DropdownMenuItem(...)`. Implementation lives in `internal/` packages, one source file per component, mirroring the shadcn registry layout.
2. **Constructors take content first, children variadic.** Leaves: `graft.Button("Cancel")`. Containers: `graft.Card(children ...ui.Widget)`. Sub-parts are plain constructors with the exact shadcn name.
3. **Props are chainable methods** returning the concrete pointer type — Go's equivalent of JSX attributes, with autocomplete acting as the prop schema. Every builder implements gogpu/ui's `ui.Widget`, so graft mixes freely with raw `primitives.Box` and any gogpu/ui widget.
4. **Variants/sizes are typed enums with sugar methods.** `type Variant uint8` (`VariantDefault, VariantSecondary, VariantDestructive, VariantOutline, VariantGhost, VariantLink`), `type Size uint8` (`SizeXS, SizeSM, SizeMD, SizeLG, SizeIcon...`). Sugar: `.Outline()`, `.Destructive()`, `.Ghost()`, `.Sm()`, `.Lg()` — so `<Button variant="outline" size="sm">` becomes `graft.Button("Cancel").Outline().Sm()`.
5. **Events**: `OnClick(func())`, `OnChange(func(T))`, `OnOpenChange(func(bool))`, `OnSubmit(func())` — same names as shadcn, Go-cased.
6. **Controlled state via coregx/signals, uncontrolled via plain values.** Rule: `Value(v)` sets an initial (uncontrolled, like `defaultValue`); `Bind(sig)` makes it controlled — the signal is read for render and written on interaction, replacing the React `value` + `onValueChange` pair with one argument. Components re-render automatically through gogpu/ui's signal invalidation. String values for Select/Tabs/Accordion/RadioGroup (mirroring shadcn); generics only where payloads demand it (`DataTable[T]`).
7. **Triggers**: native desktop doesn't need portal triggers, so overlays are signal-controlled (`.Bind(open)`), with `graft.DialogTrigger(widget)` sugar that wires the widget's click to `open.Set(true)` for users who want the literal shadcn shape.

Core sketch:

```go
package graft // import "github.com/TimLai666/graft"

type Variant uint8
const (
    VariantDefault Variant = iota
    VariantSecondary; VariantDestructive; VariantOutline; VariantGhost; VariantLink
)

type Size uint8
const ( SizeMD Size = iota; SizeXS; SizeSM; SizeLG; SizeIcon; SizeIconXS; SizeIconSM; SizeIconLG )

func Button(label string, children ...ui.Widget) *ButtonW
func (b *ButtonW) Variant(v Variant) *ButtonW   // or sugar: Outline(), Ghost(), ...
func (b *ButtonW) Size(s Size) *ButtonW         // or sugar: Sm(), Lg(), IconOnly(ic)
func (b *ButtonW) OnClick(fn func()) *ButtonW
func (b *ButtonW) Disabled(v bool) *ButtonW
func (b *ButtonW) BindDisabled(s *signals.Signal[bool]) *ButtonW
func (b *ButtonW) BindLabel(s *signals.Signal[string]) *ButtonW
func (b *ButtonW) Loading(s *signals.Signal[bool]) *ButtonW  // swaps in Spinner, like shadcn's <Spinner data-icon>
```

### 3.2 Example: login form card

```go
email := signals.New("")
password := signals.New("")

card := graft.Card(
    graft.CardHeader(
        graft.CardTitle("Login to your account"),
        graft.CardDescription("Enter your email below to login"),
        graft.CardAction(graft.Button("Sign Up").LinkStyle()),
    ),
    graft.CardContent(
        graft.FieldGroup(
            graft.Field(
                graft.FieldLabel("Email"),
                graft.Input().Bind(email).Placeholder("m@example.com"),
                graft.FieldDescription("We'll never share your email."),
            ),
            graft.Field(
                graft.FieldLabel("Password"),
                graft.Input().Bind(password).Password(),
            ),
        ),
    ),
    graft.CardFooter(
        graft.Button("Login").Full().OnClick(doLogin),
        graft.Button("Login with Google").Outline().Full(),
    ),
).W(380)
```

### 3.3 Example: dialog with footer buttons (controlled)

```go
open := signals.New(false)

graft.Dialog(
    graft.DialogContent(
        graft.DialogHeader(
            graft.DialogTitle("Delete project"),
            graft.DialogDescription("This action cannot be undone."),
        ),
        graft.DialogFooter(
            graft.Button("Cancel").Outline().OnClick(func() { open.Set(false) }),
            graft.Button("Delete").Destructive().OnClick(func() {
                deleteProject(); open.Set(false)
            }),
        ),
    ),
).Bind(open)

// shadcn-literal trigger sugar also works:
graft.DialogTrigger(graft.Button("Open").Outline(), open)
```

### 3.4 Example: settings page with switch rows

```go
notify := signals.New(true)
darkMode := signals.New(false)

graft.ItemGroup(
    graft.Item(
        graft.ItemContent(
            graft.ItemTitle("Notifications"),
            graft.ItemDescription("Receive alerts when builds finish."),
        ),
        graft.ItemActions(graft.Switch().Bind(notify)),
    ),
    graft.ItemSeparator(),
    graft.Item(
        graft.ItemContent(
            graft.ItemTitle("Dark mode"),
            graft.ItemDescription("Follow system or force dark."),
        ),
        graft.ItemActions(graft.Switch().Bind(darkMode).OnChange(applyTheme)),
    ),
)
```

### 3.5 Example: tabs

```go
tab := signals.New("account")

graft.Tabs(
    graft.TabsList(
        graft.TabsTrigger("account", "Account"),
        graft.TabsTrigger("password", "Password"),
    ),
    graft.TabsContent("account", accountForm),
    graft.TabsContent("password", passwordForm),
).Bind(tab) // or .Value("account") for uncontrolled
```

### 3.6 Example: dropdown menu with groups, checkbox item, submenu

```go
showLine := signals.New(true)

graft.DropdownMenu(
    graft.DropdownMenuTrigger(graft.Button("Open").Outline()),
    graft.DropdownMenuContent(
        graft.DropdownMenuLabel("My Account"),
        graft.DropdownMenuGroup(
            graft.DropdownMenuItem("Profile").Shortcut("Ctrl+P").OnSelect(openProfile),
            graft.DropdownMenuItem("Billing").OnSelect(openBilling),
        ),
        graft.DropdownMenuSeparator(),
        graft.DropdownMenuCheckboxItem("Show line numbers").Bind(showLine),
        graft.DropdownMenuSub("Share",
            graft.DropdownMenuItem("Email link"),
            graft.DropdownMenuItem("Copy link"),
        ),
        graft.DropdownMenuSeparator(),
        graft.DropdownMenuItem("Log out").Destructive(),
    ),
)
```

### 3.7 Example: form with validation (Go-idiomatic replacement for react-hook-form)

```go
form := graft.NewForm()
name := graft.FormValue(form, "name", "", graft.Required(), graft.MinLen(2))
mail := graft.FormValue(form, "email", "", graft.Required(), graft.Email())

graft.Form(form,
    graft.FieldSet(
        graft.FieldLegend("Profile"),
        graft.FieldGroup(
            graft.Field(
                graft.FieldLabel("Full name"),
                graft.Input().Bind(name.Signal()),
                graft.FieldError(name.Errors()), // signal of []error, renders only when invalid
            ),
            graft.Field(
                graft.FieldLabel("Email"),
                graft.Input().Bind(mail.Signal()),
                graft.FieldError(mail.Errors()),
            ),
        ),
    ),
    graft.Button("Save").Submit(), // triggers form.Validate(); OnSubmit fires only when clean
).OnSubmit(func() { save(name.Get(), mail.Get()) })
```

### 3.8 Example: data table + select + toast

```go
type User struct{ Name, Email, Role string }
rows := signals.New([]User{...})
selected := signals.New(map[int]bool{})

graft.DataTable[User](rows,
    graft.Column("Name", func(u User) ui.Widget { return graft.P(u.Name) }).Sortable(func(a, b User) bool { return a.Name < b.Name }),
    graft.Column("Email", func(u User) ui.Widget { return graft.Muted(u.Email) }),
    graft.Column("Role", func(u User) ui.Widget { return graft.Badge(u.Role).Secondary() }),
    graft.ColumnActions(func(u User) ui.Widget {
        return graft.DropdownMenu(
            graft.DropdownMenuTrigger(graft.Button("").Ghost().IconOnly(icons.MoreHorizontal)),
            graft.DropdownMenuContent(
                graft.DropdownMenuItem("Edit"),
                graft.DropdownMenuItem("Delete").Destructive().OnSelect(func() {
                    deleteUser(u)
                    graft.Toast("User deleted").Description(u.Email).Action("Undo", undo)
                }),
            ),
        )
    }),
).Selectable(selected).Paginate(10)

role := signals.New("viewer")
graft.Select(
    graft.SelectItem("viewer", "Viewer"),
    graft.SelectItem("editor", "Editor"),
    graft.SelectGroup("Admin roles",
        graft.SelectItem("admin", "Admin"),
        graft.SelectItem("owner", "Owner"),
    ),
).Bind(role).Placeholder("Select role").W(180)
```

(`graft.Toaster()` is mounted once at the app root, exactly like Sonner's `<Toaster />`; `graft.Toast(...)` is the imperative `toast(...)` equivalent.)

### 3.9 App shell with sidebar (composition with raw gogpu/ui)

```go
collapsed := signals.New(false)
shell := graft.SidebarProvider(collapsed,
    graft.Sidebar(
        graft.SidebarHeader(graft.H4("Acme")),
        graft.SidebarContent(
            graft.SidebarGroup("Platform",
                graft.SidebarMenuButton("Dashboard", icons.Home).Active(route.Is("/")),
                graft.SidebarMenuButton("Settings", icons.Gear).Badge("3"),
            ),
        ),
        graft.SidebarFooter(userMenu),
    ),
    graft.SidebarInset(mainContent), // any ui.Widget, including raw primitives.Box
)
```

## 4. Theming — mirroring the CSS-variable workflow in Go

graft registers itself as a gogpu/ui design system ("painter set"), with a token struct that is a 1:1 mirror of shadcn's CSS variables. The three shadcn customization knobs — base color, radius, token overrides per mode — map directly:

```go
// 1. Pick a base color + radius (the `components.json` step)
th := graft.NewTheme(
    graft.BaseColor(graft.Zinc),   // Neutral | Stone | Zinc | Mauve | Olive | Mist | Taupe
    graft.Radius(10),              // px; derives RadiusSM..Radius4XL scale
)

// 2. Override individual tokens per mode (the `:root` / `.dark` step)
th.Light.Primary = graft.OKLCH(0.55, 0.2, 260)
th.Light.PrimaryForeground = graft.OKLCH(0.98, 0, 0)
th.Dark.Ring = graft.OKLCH(0.55, 0.15, 260)

// 3. Or import a theme straight from the shadcn/tweakcn theme editor
th, err := graft.ParseThemeCSS(cssText) // parses the `:root { --primary: oklch(...) }` block

// 4. Install; mode switching is a signal, dark mode = swapping the token set
mode := signals.New(graft.ModeSystem) // ModeLight | ModeDark | ModeSystem
app.Use(graft.Install(th, mode))
```

```go
type Tokens struct {
    Background, Foreground             Color
    Card, CardForeground               Color
    Popover, PopoverForeground         Color
    Primary, PrimaryForeground         Color
    Secondary, SecondaryForeground     Color
    Muted, MutedForeground             Color
    Accent, AccentForeground           Color
    Destructive, DestructiveForeground Color
    Border, Input, Ring                Color
    Chart [5]Color
    Sidebar, SidebarForeground, SidebarPrimary, SidebarPrimaryForeground,
    SidebarAccent, SidebarAccentForeground, SidebarBorder, SidebarRing Color
    Radius float32
    FontSans, FontMono FontStack
}
```

`ParseThemeCSS` is the differentiator: every theme built with the shadcn theme editor or tweakcn pastes straight into a Go desktop app. Per-widget escape hatch mirrors `className` overrides: every builder gets `.Style(func(s *graft.Style))` for one-off tweaks.

## 5. Implementation priority

**Tier 1 — core primitives (every app, plus the theme engine):** theme engine + tokens + `ParseThemeCSS`; Typography constructors; Button, Label, Input, Textarea, Checkbox, Switch, RadioGroup, Select, Slider, Separator, Badge, Alert, Spinner, Progress, Card, ScrollArea, Tabs, Tooltip, Popover, Dialog, AlertDialog, DropdownMenu, Field + Form. Rationale: 19 of 24 are class A/B over existing gogpu/ui widgets — fast wins that prove the API; Field/Form unlocks real apps.

**Tier 2 — app chrome and data:** Accordion, Collapsible, Avatar, ButtonGroup, InputGroup, Item, Toggle, ToggleGroup, Skeleton, Table, DataTable, ContextMenu, Menubar, Resizable, Sheet, Sidebar, Sonner/Toast, Calendar, DatePicker, Combobox. Rationale: the desktop-app shell (Sidebar/Menubar/Resizable/Sheet) plus the two hard C-class widgets users ask for first (Calendar, Combobox) and the toast system.

**Tier 3 — long tail:** Command, Breadcrumb, Pagination, Empty, Kbd, HoverCard, InputOTP, Carousel, Chart, AspectRatio. Omitted with rationale: NativeSelect (no native controls in GPU toolkit — Select covers it), Drawer (mobile gesture pattern — Sheet side="bottom" covers it), legacy Toast (Sonner-style system only), NavigationMenu (web mega-nav — Menubar/Sidebar/Tabs are the desktop idioms), Direction (defer to gogpu/ui text layer), Form/react-hook-form (replaced by signal-backed `graft.Form`).

**Suggested repo layout:** `graft.go` (types, Variant/Size, Install), `theme.go`/`theme_css.go`, one file per component (`button.go`, `card.go`, ...) matching shadcn registry names, `internal/render` for painters, `examples/` reproducing shadcn demo blocks (login card, dashboard, settings) as compile-checked documentation.
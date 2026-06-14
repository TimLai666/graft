# graft

**shadcn/ui for Go.** A pixel-faithful, Go-native port of [shadcn/ui](https://ui.shadcn.com) built on [gogpu/ui](https://github.com/gogpu/ui) — GPU-accelerated, pure Go, zero CGO.

graft replicates shadcn/ui down to the token level: the same OKLCH CSS variables (copied verbatim), the same component anatomy and variants, the same Geist typography, the same theming workflow. If you know shadcn, you already know graft.

> Status: full shadcn component set — 58 components including Drawer, Navigation Menu, and a sortable/selectable/paginated Data Table, across every tier (form controls, overlays, navigation, data display, charts, command palette, sidebar, carousel, OTP). Each is golden-tested pixel-for-pixel against the shadcn spec. Not implemented (with reason): **Native Select** (no native `<select>` in a Go GPU GUI — `Select` is the equivalent), **Toast** (superseded by Sonner, which is included), **Direction/RTL** (needs bidi text shaping from gogpu's text layer).
>
> Pixel target: graft's metrics track the **current ui.shadcn.com style** (the "Radix UI" era — controls at `h-8`, `px-2.5`, `rounded-lg`; cards `rounded-xl`), re-transcribed from live computed CSS on **2026-06-14**. shadcn is a moving target; component geometry is verified against that snapshot (see `compare/shadcn-spec.md`). Colors use the canonical OKLCH token presets, validated against Tailwind v4.

## Why

Native Go apps deserve the design quality the web has had for years. gogpu/ui provides the engine — widgets, layout, GPU rendering; graft provides the design system on top: every component styled exactly like shadcn/ui, themeable exactly like shadcn/ui.

## Usage

```go
package main

import (
    "log"

    "github.com/TimLai666/graft"
    "github.com/TimLai666/graft/graftapp"
)

func main() {
    // Theme: the components.json workflow, in Go.
    th := graft.NewTheme(
        graft.BaseColor(graft.Zinc), // neutral | stone | zinc | gray | slate
        graft.Radius(10),            // --radius
    )

    err := graftapp.New().
        Title("App").
        Size(960, 640).
        Theme(th).
        Run(graft.Card(
            graft.CardHeader(
                graft.CardTitle("Login to your account"),
                graft.CardDescription("Enter your email below to login"),
            ),
            graft.CardContent(
                graft.Input().Placeholder("m@example.com"),
            ),
            graft.CardFooter(
                graft.Button("Login").Full(),
            ),
        ).W(384))
    if err != nil {
        log.Fatal(err)
    }
}
```

`graftapp` is a thin launcher module that folds the gogpu window/event/render
setup into one call. The core `graft` module stays free of GPU and windowing
dependencies, so you can also embed graft components in a hand-wired
`gogpu/ui` app or render them off-screen — see [`graft.Install`](graft.go) and
[`graft.PaintersFor`](graft.go) for the manual path.

## Theming, the shadcn way

```go
// 1. Pick a base color + radius (components.json)
th := graft.NewTheme(graft.BaseColor(graft.Neutral), graft.Radius(10))

// 2. Override tokens directly — values copied verbatim from any shadcn theme
th.Light.Primary = graft.OKLCH(0.55, 0.20, 260)
th.Dark.Ring = graft.OKLCH(0.55, 0.15, 260)

// 3. Or paste an entire theme from the shadcn / tweakcn theme editor
th, err := graft.ParseThemeCSS(`
:root { --primary: oklch(0.55 0.2 260); --radius: 0.5rem; }
.dark { --primary: oklch(0.65 0.18 260); }
`)

// 4. Dark mode is one call — every component repaints, no rebuild
th.SetMode(graft.ModeDark)
```

Colors are specified in OKLCH exactly as shadcn ships them. graft implements the CSS Color 4 conversion in Go, including the same sRGB gamut handling browsers use, validated against Tailwind v4's published fallbacks — out-of-gamut tokens like `--destructive` render the same pixels Chrome renders.

## What's inside

- `theme/` — the full shadcn token system: 5 base-color presets (verbatim from the shadcn registry), light/dark, radius scale, a CSS theme parser (`oklch()`/`hsl()`/`rgb()`/hex), OKLCH engine
- `fonts/` — embedded Geist + Geist Mono (OFL-1.1), registered per weight (400/500/600/700)
- `icons/` — the Lucide icons shadcn components use (ISC), with a vendoring tool for more
- `painters/` — shadcn-styled painters for raw gogpu/ui core widgets
- `metrics/` — every px constant from the shadcn spec, annotated with its source Tailwind classes
- `graftapp/` — one-call desktop launcher (a separate module, so the core library carries no GPU/windowing deps)
- Golden-image tests render every component headlessly at 2x and compare pixel-for-pixel (`GRAFT_UPDATE_GOLDEN=1 go test ./...` to re-record)

## Running the gallery

`examples/kitchensink` shows every component in one scrollable window. It is a
separate module (so the core library carries no GPU/windowing deps); a
`go.work` at the repo root wires it in, so from the root:

```sh
go run ./examples/kitchensink            # interactive window
go run ./examples/kitchensink -png sheet.png        # headless light sheet
go run ./examples/kitchensink -png sheet.png -dark  # headless dark sheet
```

(Without the workspace, run it from inside the directory: `cd examples/kitchensink && go run .`)

## Known limitations

- **HiDPI Windows displays (>100% scaling):** the gallery renders oversized and
  mouse hit-testing is offset on Windows displays scaled above 100%. This is an
  upstream gogpu bug — `gogpu.App.ScaleFactor()` returns `1.0` on a 200% display,
  so the render scale and the event/layout scale diverge. Tracked at
  [gogpu/gogpu#306](https://github.com/gogpu/gogpu/issues/306). It affects only
  the native-window path; graft's components are unaffected (golden tests render
  off-screen) and the gallery is correct at 100% scaling.

## License

MIT. Geist font © Vercel (OFL-1.1), Lucide icons (ISC), design system by [shadcn](https://ui.shadcn.com).

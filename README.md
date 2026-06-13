# graft

**shadcn/ui for Go.** A pixel-faithful, Go-native port of [shadcn/ui](https://ui.shadcn.com) built on [gogpu/ui](https://github.com/gogpu/ui) — GPU-accelerated, pure Go, zero CGO.

graft replicates shadcn/ui down to the token level: the same OKLCH CSS variables (copied verbatim), the same component anatomy and variants, the same Geist typography, the same theming workflow. If you know shadcn, you already know graft.

> Status: early development. Foundation (theme system, fonts, icons, golden-test pipeline) is complete; Tier 1 components are landing.

## Why

Native Go apps deserve the design quality the web has had for years. gogpu/ui provides the engine — widgets, layout, GPU rendering; graft provides the design system on top: every component styled exactly like shadcn/ui, themeable exactly like shadcn/ui.

## Usage

```go
package main

import (
    "log"

    "github.com/gogpu/gogpu"
    "github.com/gogpu/ui/app"
    "github.com/gogpu/ui/desktop"

    "github.com/TimLai666/graft"
)

func main() {
    // Theme: the components.json workflow, in Go.
    th := graft.NewTheme(
        graft.BaseColor(graft.Zinc), // neutral | stone | zinc | gray | slate
        graft.Radius(10),            // --radius
    )

    gpuApp := gogpu.NewApp(gogpu.DefaultConfig().WithTitle("App").WithSize(960, 640))
    uiApp := app.New(
        app.WithWindowProvider(gpuApp),
        app.WithPlatformProvider(gpuApp),
        app.WithEventSource(gpuApp.EventSource()),
        app.WithTheme(th.AsUITheme()),
    )
    if err := graft.Install(uiApp, th); err != nil {
        log.Fatal(err)
    }

    uiApp.SetRoot(graft.Card(
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

    log.Fatal(desktop.Run(gpuApp, uiApp))
}
```

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

## License

MIT. Geist font © Vercel (OFL-1.1), Lucide icons (ISC), design system by [shadcn](https://ui.shadcn.com).

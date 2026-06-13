# Phase 1 Batch 1 — completed components (reference for batch 2)

These shipped to main and are GREEN. Batch-2 agents must follow their
conventions and must NOT redefine the shared test helpers they introduced.

## Shipped components
- button.go + metrics/button.go (6 variants × 8 sizes, focus ring, dark mode)
- label.go + metrics/label.go
- separator.go + metrics/separator.go
- badge.go + metrics/badge.go
- card.go + metrics/card.go (Card/CardHeader/CardTitle/CardDescription/CardAction/CardContent/CardFooter)
- scrollarea.go + painters/scrollbar.go + metrics/scrollarea.go
- popover.go + metrics/popover.go (Popover/PopoverTrigger/PopoverContent)
- slider.go + painters/slider.go + metrics/slider.go
- tabs.go + metrics/tabs.go (Tabs/TabsList/TabsTrigger/TabsContent)

## SHARED TEST HELPERS already defined in package graft_test (DO NOT REDEFINE)
Defined across the merged test files — reuse them, never re-declare:
- `lightTokens(t *testing.T) *theme.Tokens` (button_test.go) — forces light mode, returns active tokens, restores on cleanup
- `darkTokens(t *testing.T) *theme.Tokens` (button_test.go)
- `forceLightMode(t *testing.T) *theme.Theme` (separator_test.go) — forces light, returns the theme
- `alpha(c widget.Color, a float32) widget.Color` (button_test.go) — sets alpha
- `mulAlpha(c widget.Color, f float32) widget.Color` (button_test.go) — multiplies alpha
- `looseConstraints()` and `layoutWidth(t, w)` (graft_test.go)

If you need a NEW shared helper, give it a component-prefixed unique name
(e.g. `dialogMockCtx`, not `mockCtx`) to avoid collisions at merge time.

## Conventions confirmed working
- Component = graft-owned widget in root package, one file per shadcn entry.
- Colors resolved from th.Active() inside Draw (never cached).
- Wrapped core widgets: scrollview (ScrollArea), slider (Slider) via painters/<name>.go methods on the pre-declared painters struct.
- Goldens via gtest.GoldenLightDark; eyeball every PNG before finishing.
- Metrics in metrics/<component>.go, annotated with source Tailwind classes.

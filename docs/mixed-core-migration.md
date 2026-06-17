# Mixed migration: graft API with gogpu/ui core mechanisms

Date: 2026-06-17

## Positioning

`github.com/gogpu/ui` is the underlying GUI toolkit: widget tree, layout, event dispatch, focus, state binding, offscreen rendering, core widgets, and Painter interfaces.

`github.com/TimLai666/graft` is the shadcn/ui port on top of that toolkit: shadcn-shaped Go APIs, theme tokens, metrics, fonts, icons, golden alignment, and component composition.

The migration target is therefore mixed, not global:

- Keep graft's public shadcn-like API stable.
- Move suitable interaction mechanisms to `gogpu/ui/core`.
- Keep shadcn anatomy in graft when core cannot represent it without visual or API regressions.

## Component status

| Component | Status | Mechanism ownership | Fidelity result |
| --- | --- | --- | --- |
| Button | Partial core wrapper | `core/button` handles hover, press, cursor, disabled event rejection, focus request, and mouse click state. graft keeps exact layout, icon/children slots, theme token painting, and key-press compatibility. | Existing Button goldens remain 0 pixel diff. |
| Checkbox | Core wrapper candidate | `core/checkbox` handles hover, press, cursor, disabled event rejection, focus request, mouse release toggle, and signal writeback. graft keeps 16px shadcn layout and theme token painting. | Existing Checkbox goldens remain 0 pixel diff. |
| Tabs | Keep graft-owned | `core/tabview` is not used. graft keeps list, trigger, content layout, trigger-level focus, hit testing, and keyboard navigation. | Existing Tabs goldens remain 0 pixel diff. |

## Evidence

Button can wrap `core/button`, but pure core layout is not shadcn-faithful today:

- core/button estimates text width as `len(text) * fontSize * 0.55`.
- core/button sizes are 32, 40, and 48px; graft needs shadcn's full size matrix, including 24px xs and 36px icon.
- core/button has no native icon/child slot model matching `Button(...children)`, `Icon`, and `IconOnly`.

Checkbox is the strongest migration candidate:

- core/checkbox already owns the useful toggle/focus/hover mechanism.
- graft still needs to override layout because core/checkbox paints an 18px box with a 24px minimum height, while shadcn Checkbox uses a 16px box.

Tabs should not wrap `core/tabview` yet:

- core/tabview uses a fixed 48px tab bar and equal-width tabs.
- shadcn Tabs needs a w-fit 32px list pill, variable-width triggers, line variant underline, vertical trigger columns, and per-trigger focus rings.
- graft exposes `TabsTriggerWidget` as the focusable unit; core/tabview focuses the tabview as a whole.

## Verification

Command:

```powershell
go test ./...
```

Result: pass.

Golden comparison used the existing graft golden system without `GRAFT_UPDATE_GOLDEN=1`. Since `go test ./...` passed, the relevant checked goldens had exact pixel diffs of `0`.

| Component | Goldens compared | Pixel diff |
| --- | --- | --- |
| Button | `button-variants`, `button-sizes`, `button-icon`, `button-disabled`, `button-focused`, `button-hovered`, light and dark | 0 pixels each |
| Checkbox | `checkbox-unchecked`, `checkbox-checked`, `checkbox-indeterminate`, `checkbox-with-label`, `checkbox-disabled`, `checkbox-focused`, light and dark | 0 pixels each |
| Tabs | `tabs-default`, `tabs-line`, `tabs-focus`, `tabs-vertical`, light and dark | 0 pixels each |

## Decision

Adopt selective core wrapping.

- Keep the Button and Checkbox wrapper direction, with explicit graft-side compatibility bridges where core cannot yet match shadcn behavior.
- Keep Tabs graft-owned until `core/tabview` supports variable tab bounds, per-trigger focus or external trigger widgets, custom tabbar sizing, and vertical orientation.
- Treat future migrations component-by-component. Prefer simple leaf controls and existing core Painter matches before Radix-like compound widgets.

Likely next candidates after this branch: Progress, Slider, ScrollArea, Popover, and Tooltip. Avoid broad migration of menu, tabs, calendar, command, and other compound components until their core models match shadcn anatomy.

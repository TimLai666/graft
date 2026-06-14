# Current shadcn live-style spec (extracted 2026-06-14 from ui.shadcn.com)

Re-transcription target for graft → match the current "radix"-era style (newer
than new-york-v4). Computed px at 1rem=16px, --radius=0.625rem(10px).
radius scale: sm=6 md=8 lg=10 xl=14, rounded-4xl≈26, rounded-full=pill.

## Form controls
| component | live key classes | live px | change vs graft |
|---|---|---|---|
| button default | h-8 px-2.5 py-2 gap-1.5 rounded-lg text-sm font-medium | h32 px10 gap6 r10 | h36→32, px16→10, gap8→6, md→lg |
| button sm | (verify; likely h-7 px-2 gap-1) | — | verify |
| button lg | (verify; likely h-9/10 px-4/6) | — | verify |
| button icon | size-8 (verify) | — | size36→32 |
| input | h-8 px-2.5 py-1 rounded-lg | h32 px10 r10 fs14 | h36→32, px12→10, md→lg |
| textarea | min-h-16 px-2.5 py-2 rounded-lg field-sizing-content | minH64 px10 py8 r10 | px? py? md→lg |
| select-trigger | h-8 gap-1.5 px-2.5(py-2) rounded-lg | h32 px10 r10 fs14 | h36→32, px→10, md→lg |
| checkbox | size-4 rounded-[4px] border | 16 r4 | radius →4px |
| radio-group-item | size-4 rounded-full border | 16 full | UNCHANGED ✓ |
| switch | track w-8 h-[1.15rem] rounded-full; thumb size-4 | 32x18.4 thumb16 | UNCHANGED ✓ |
| label | gap-2 text-sm leading-none font-medium | fs14 fw500 gap8 | UNCHANGED ✓ |

## Surfaces / display
| component | live key classes | live px | change vs graft |
|---|---|---|---|
| card | py-(spacing) gap-(spacing) rounded-xl; spacing=16 | py16 gap16 r14 | py/gap 24→16 |
| card-header | gap-1 px-16 | gap4 px16 | gap→4 |
| card-title | text-base leading-snug font-medium | fs16 fw500 | text-2xl/semibold → base/medium |
| card-description | text-sm | fs14 | — |
| card-content | px-16 | px16 | — |
| card-footer | border-t p-16 gap-2 | px16 py16 gap8 | — |
| alert | gap-0.5 px-2.5 py-2 rounded-lg border text-sm | px10 py8 r10 | px16→10, py→8, md→lg |
| alert-title | font-medium | fs14 fw500 | — |
| alert-description | text-sm | fs14 | — |
| slider track | rounded-full h-1(4) | h4 | track thinner (was 6/8 →4) |
| slider thumb | size-3 rounded-full border-ring | 12 | thumb 16→12 |
| progress | h-1 rounded-full | h4 | h 8→4 |
| tabs-list | rounded-lg p-[3px] | h32 r10 p3 | rounded-md→lg, p→3 |
| tabs-trigger | h-[calc(100%-1px)] gap-1.5 rounded-md px-1.5 py-0.5 text-sm font-medium | h25 px6 r8 | px3→1.5, rounded-sm→md |
| accordion-trigger | rounded-lg py-2.5 text-sm font-medium | py10 fs14 fw500 | py16→10 |
| badge | h-5 px-2 py-0.5 gap-1 rounded-4xl border text-xs font-medium | h20 px8 py2 r26 | radius →rounded-4xl(26) |

## Still to extract (overlays need open state): tooltip, dialog, popover,
## dropdown-menu item, select-content/item, context-menu, menubar, sheet,
## alert-dialog, hover-card, command. Plus: avatar, separator, kbd, skeleton,
## spinner, toggle, toggle-group, breadcrumb, pagination, table, scroll-area,
## calendar, datepicker, combobox, sonner, sidebar, carousel, inputotp, item,
## field, empty, resizable, aspect-ratio.

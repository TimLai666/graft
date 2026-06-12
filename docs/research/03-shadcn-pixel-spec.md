# shadcn/ui Exact Design Specification (for pixel-perfect Go port)

**Source of truth used:** `shadcn-ui/ui` repo, branch `main`, registry `apps/v4/registry/new-york-v4/ui/*.tsx` (the current canonical "new-york-v4" registry, fetched 2026-06-13). Theme variables cross-checked against `https://ui.shadcn.com/docs/theming` and `https://ui.shadcn.com/docs/installation/manual` (the CSS that `shadcn init` actually ships to user projects). Shadow values from Tailwind CSS v4 docs. All conversions assume `1rem = 16px`, Tailwind `--spacing = 0.25rem = 4px`.

> NOTE: `apps/v4/app/globals.css` is the **docs-site** stylesheet — it overrides a few tokens (foreground `oklch(0% 0 0)`, chart colors mapped to blue scale, extra `surface/code/selection` tokens, `--destructive-foreground`). The **canonical distributed theme** is the one below from docs/theming + installation/manual. Both are reproduced; use the canonical one for the port.

---

## 1. DESIGN TOKENS — default ("neutral") theme, Tailwind v4, OKLCH

### Canonical `:root` (verbatim, from ui.shadcn.com/docs/theming and /docs/installation/manual)

```css
:root {
  --radius: 0.625rem;
  --background: oklch(1 0 0);
  --foreground: oklch(0.145 0 0);
  --card: oklch(1 0 0);
  --card-foreground: oklch(0.145 0 0);
  --popover: oklch(1 0 0);
  --popover-foreground: oklch(0.145 0 0);
  --primary: oklch(0.205 0 0);
  --primary-foreground: oklch(0.985 0 0);
  --secondary: oklch(0.97 0 0);
  --secondary-foreground: oklch(0.205 0 0);
  --muted: oklch(0.97 0 0);
  --muted-foreground: oklch(0.556 0 0);
  --accent: oklch(0.97 0 0);
  --accent-foreground: oklch(0.205 0 0);
  --destructive: oklch(0.577 0.245 27.325);
  --border: oklch(0.922 0 0);
  --input: oklch(0.922 0 0);
  --ring: oklch(0.708 0 0);
  --chart-1: oklch(0.646 0.222 41.116);
  --chart-2: oklch(0.6 0.118 184.704);
  --chart-3: oklch(0.398 0.07 227.392);
  --chart-4: oklch(0.828 0.189 84.429);
  --chart-5: oklch(0.769 0.188 70.08);
  --sidebar: oklch(0.985 0 0);
  --sidebar-foreground: oklch(0.145 0 0);
  --sidebar-primary: oklch(0.205 0 0);
  --sidebar-primary-foreground: oklch(0.985 0 0);
  --sidebar-accent: oklch(0.97 0 0);
  --sidebar-accent-foreground: oklch(0.205 0 0);
  --sidebar-border: oklch(0.922 0 0);
  --sidebar-ring: oklch(0.708 0 0);
}
```

### Canonical `.dark` (verbatim)

```css
.dark {
  --background: oklch(0.145 0 0);
  --foreground: oklch(0.985 0 0);
  --card: oklch(0.205 0 0);
  --card-foreground: oklch(0.985 0 0);
  --popover: oklch(0.205 0 0);
  --popover-foreground: oklch(0.985 0 0);
  --primary: oklch(0.922 0 0);
  --primary-foreground: oklch(0.205 0 0);
  --secondary: oklch(0.269 0 0);
  --secondary-foreground: oklch(0.985 0 0);
  --muted: oklch(0.269 0 0);
  --muted-foreground: oklch(0.708 0 0);
  --accent: oklch(0.269 0 0);
  --accent-foreground: oklch(0.985 0 0);
  --destructive: oklch(0.704 0.191 22.216);
  --border: oklch(1 0 0 / 10%);
  --input: oklch(1 0 0 / 15%);
  --ring: oklch(0.556 0 0);
  --chart-1: oklch(0.488 0.243 264.376);
  --chart-2: oklch(0.696 0.17 162.48);
  --chart-3: oklch(0.769 0.188 70.08);
  --chart-4: oklch(0.627 0.265 303.9);
  --chart-5: oklch(0.645 0.246 16.439);
  --sidebar: oklch(0.205 0 0);
  --sidebar-foreground: oklch(0.985 0 0);
  --sidebar-primary: oklch(0.488 0.243 264.376);
  --sidebar-primary-foreground: oklch(0.985 0 0);
  --sidebar-accent: oklch(0.269 0 0);
  --sidebar-accent-foreground: oklch(0.985 0 0);
  --sidebar-border: oklch(1 0 0 / 10%);
  --sidebar-ring: oklch(0.556 0 0);
}
```

Notes:
- `--destructive-foreground` no longer appears in the canonical neutral theme; destructive button/badge use literal `text-white`. The docs-site CSS defines `--destructive-foreground: oklch(0.97 0.01 17)` (light) / `oklch(0.58 0.22 27)` (dark) — docs-site only.
- Docs-site (`apps/v4/app/globals.css`) deviations, for reference only: `--foreground/--card-foreground/--popover-foreground/--primary/--sidebar-foreground: oklch(0% 0 0)`, `--accent` dark = `oklch(0.371 0 0)`, `--chart-1..5: var(--color-blue-300/500/600/700/800)`, extra tokens `--surface: oklch(0.98 0 0)` (dark `oklch(0.2 0 0)`), `--selection: oklch(0% 0 0)`/`--selection-foreground: oklch(1 0 0)`, dark `--sidebar-ring: oklch(0.439 0 0)`.
- Convention: every surface token pairs with a `-foreground` token ("background suffix is omitted": `primary` + `primary-foreground`).

### `@theme inline` color mapping (verbatim, what makes `bg-primary` etc. work)

```css
@theme inline {
  --color-background: var(--background);
  --color-foreground: var(--foreground);
  --color-card: var(--card);
  --color-card-foreground: var(--card-foreground);
  --color-popover: var(--popover);
  --color-popover-foreground: var(--popover-foreground);
  --color-primary: var(--primary);
  --color-primary-foreground: var(--primary-foreground);
  --color-secondary: var(--secondary);
  --color-secondary-foreground: var(--secondary-foreground);
  --color-muted: var(--muted);
  --color-muted-foreground: var(--muted-foreground);
  --color-accent: var(--accent);
  --color-accent-foreground: var(--accent-foreground);
  --color-destructive: var(--destructive);
  --color-destructive-foreground: var(--destructive-foreground);
  --color-border: var(--border);
  --color-input: var(--input);
  --color-ring: var(--ring);
  --color-chart-1: var(--chart-1);
  /* ... chart-2..5, sidebar + 7 sidebar-* identically ... */
  --radius-sm: calc(var(--radius) * 0.6);
  --radius-md: calc(var(--radius) * 0.8);
  --radius-lg: var(--radius);
  --radius-xl: calc(var(--radius) * 1.4);
  --radius-2xl: calc(var(--radius) * 1.8);
  --radius-3xl: calc(var(--radius) * 2.2);
  --radius-4xl: calc(var(--radius) * 2.6);
}
@layer base {
  * { @apply border-border outline-ring/50; }
  body { @apply bg-background text-foreground; }
}
```

The `* { border-border }` base rule is why bare `border` classes everywhere render in `--border` color.

---

## 2. RADIUS SYSTEM

`--radius: 0.625rem` = **10px** default.

Current (2025+) **multiplicative** formulas (both docs-site and `shadcn init` output now use these):

| Token | Formula | px @ default |
|---|---|---|
| `--radius-xs` (Tailwind default, not overridden) | `0.125rem` | 2px |
| `--radius-sm` | `calc(var(--radius) * 0.6)` | 6px |
| `--radius-md` | `calc(var(--radius) * 0.8)` | 8px |
| `--radius-lg` | `var(--radius)` | 10px |
| `--radius-xl` | `calc(var(--radius) * 1.4)` | 14px |
| `--radius-2xl` | `calc(var(--radius) * 1.8)` | 18px |
| `--radius-3xl` | `calc(var(--radius) * 2.2)` | 22px |
| `--radius-4xl` | `calc(var(--radius) * 2.6)` | 26px |

(The earlier v4 release used additive formulas `sm = radius − 4px`, `md = radius − 2px`, `xl = radius + 4px` — identical px results at the 10px default: 6/8/14. Use multiplicative; it's current.)

So in components: `rounded-sm`=6px, `rounded-md`=8px, `rounded-lg`=10px, `rounded-xl`=14px, `rounded-xs`=2px, `rounded-full`=9999px, `rounded-[4px]`=4px, `rounded-[2px]`=2px.

---

## 3. TYPOGRAPHY

- **Font stack:** registry components don't set a family; they inherit `--font-sans`. Tailwind v4 default: `ui-sans-serif, system-ui, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"`. Mono: `ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace`. The ui.shadcn.com site itself loads **Geist** (`--font-sans`) and **Geist Mono** via next/font. For a faithful port, ship Geist; system stack is the spec'd fallback.
- **Sizes** (Tailwind v4 defaults):
  - `text-xs` = 12px / line-height 16px (`calc(1/0.75)`)
  - `text-sm` = 14px / 20px (`calc(1.25/0.875)`)
  - `text-base` = 16px / 24px
  - `text-lg` = 18px / 28px
  - `text-[0.8rem]` = 12.8px (calendar weekday)
- **Weights used:** `font-normal` 400, `font-medium` **500** (buttons, labels, menu labels, alert title, table head, tabs), `font-semibold` **600** (CardTitle, DialogTitle, SheetTitle).
- `leading-none` = line-height 1 (Label, CardTitle, DialogTitle).
- `tracking-tight` = −0.025em (AlertTitle); `tracking-widest` = +0.1em (menu/command shortcuts).
- Inputs: `text-base` (16px) on mobile, `md:text-sm` (14px) at ≥768px — desktop spec is **14px**.
- Body rendering: `font-synthesis-weight: none; text-rendering: optimizeLegibility;` (docs site).

---

## 4. SHADOWS (Tailwind v4 defaults, used verbatim by shadcn)

| Class | Exact box-shadow |
|---|---|
| `shadow-2xs` | `0 1px rgb(0 0 0 / 0.05)` |
| `shadow-xs` | `0 1px 2px 0 rgb(0 0 0 / 0.05)` |
| `shadow-sm` | `0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)` |
| `shadow-md` | `0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)` |
| `shadow-lg` | `0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)` |
| `shadow-xl` | `0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)` |
| `shadow-2xl` | `0 25px 50px -12px rgb(0 0 0 / 0.25)` |

Usage map: `shadow-xs` → outline button, input, textarea, checkbox, radio, switch, select trigger, toggle-outline. `shadow-sm` → card, slider thumb, active tab. `shadow-md` → popover, select content, dropdown content. `shadow-lg` → dialog, sheet, dropdown sub-content.

### Spacing quick table (Tailwind, 4px grid)
`0.5`=2px, `1`=4px, `1.5`=6px, `2`=8px, `2.5`=10px, `3`=12px, `3.5`=14px, `4`=16px, `5`=20px, `6`=24px, `7`=28px, `8`=32px, `9`=36px, `10`=40px, `12`=48px, `16`=64px, `44`=176px, `72`=288px; `px`=1px.

---

## 5. COMPONENT METRICS

### Button — full cva block VERBATIM

```ts
const buttonVariants = cva(
  "inline-flex shrink-0 items-center justify-center gap-2 rounded-md text-sm font-medium whitespace-nowrap transition-all outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        destructive:
          "bg-destructive text-white hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:bg-destructive/60 dark:focus-visible:ring-destructive/40",
        outline:
          "border bg-background shadow-xs hover:bg-accent hover:text-accent-foreground dark:border-input dark:bg-input/30 dark:hover:bg-input/50",
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost:
          "hover:bg-accent hover:text-accent-foreground dark:hover:bg-accent/50",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        default: "h-9 px-4 py-2 has-[>svg]:px-3",
        xs: "h-6 gap-1 rounded-md px-2 text-xs has-[>svg]:px-1.5 [&_svg:not([class*='size-'])]:size-3",
        sm: "h-8 gap-1.5 rounded-md px-3 has-[>svg]:px-2.5",
        lg: "h-10 rounded-md px-6 has-[>svg]:px-4",
        icon: "size-9",
        "icon-xs": "size-6 rounded-md [&_svg:not([class*='size-'])]:size-3",
        "icon-sm": "size-8",
        "icon-lg": "size-10",
      },
    },
    defaultVariants: { variant: "default", size: "default" },
  }
)
```

**px translation:** base = inline-flex centered, gap 8px, radius 8px, font 14px/20px weight 500, no outline; svg icons 16px. Sizes: default **h 36px, px 16px** (12px if it directly contains an svg), py 8px; xs **h 24px**, gap 4px, px 8px (6px w/ icon), font 12px, icons 12px; sm **h 32px**, gap 6px, px 12px (10px w/ icon); lg **h 40px**, px 24px (16px w/ icon); icon 36×36; icon-xs 24×24 (icon 12px); icon-sm 32×32; icon-lg 40×40. States: hover per variant above (90%/80% alpha or accent fill); disabled = opacity 0.5 + no pointer events; focus-visible = border→ring color + 3px ring of `ring` @50%; aria-invalid = border destructive + 3px ring destructive @20% (dark 40%). Outline variant border = 1px `--border` (dark: 1px `--input`, fill `--input`@30%, hover `--input`@50%). Link: underline offset 4px on hover only. Docs site adds global `a:active, button:active { opacity-60 md:opacity-100 }` (mobile-only press dim — site-level, not component spec).

### Badge (verbatim variants)

Base: `"inline-flex w-fit shrink-0 items-center justify-center gap-1 overflow-hidden rounded-full border border-transparent px-2 py-0.5 text-xs font-medium whitespace-nowrap transition-[color,box-shadow] focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 [&>svg]:pointer-events-none [&>svg]:size-3"`
- `default: "bg-primary text-primary-foreground [a&]:hover:bg-primary/90"`
- `secondary: "bg-secondary text-secondary-foreground [a&]:hover:bg-secondary/90"`
- `destructive: "bg-destructive text-white focus-visible:ring-destructive/20 dark:bg-destructive/60 dark:focus-visible:ring-destructive/40 [a&]:hover:bg-destructive/90"`
- `outline: "border-border text-foreground [a&]:hover:bg-accent [a&]:hover:text-accent-foreground"`
- `ghost: "[a&]:hover:bg-accent [a&]:hover:text-accent-foreground"`
- `link: "text-primary underline-offset-4 [a&]:hover:underline"`

px: **pill (rounded-full)**, px 8px, py 2px, font 12px/16px weight 500, gap 4px, 1px transparent border (outline variant: `--border`), svg 12px. Hover states apply only when rendered as `<a>` (`[a&]`). (Recent change: badge is now fully rounded, was `rounded-md` in older registry.)

### Input
Verbatim: `"h-9 w-full min-w-0 rounded-md border border-input bg-transparent px-3 py-1 text-base shadow-xs transition-[color,box-shadow] outline-none selection:bg-primary selection:text-primary-foreground file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm dark:bg-input/30"` + `"focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"` + `"aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40"`

px: h **36px**, radius 8px, border 1px `--input`, px 12px, py 4px, font 14px (desktop), shadow-xs, transparent bg (dark: `--input`@30%), placeholder `--muted-foreground`, file-button h 28px, selection bg `--primary`.

### Textarea
`"flex field-sizing-content min-h-16 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-xs ... md:text-sm dark:bg-input/30 ..."` (same focus/invalid recipe). px: min-h **64px**, px 12px, py 8px, radius 8px, auto-grows with content (`field-sizing: content`).

### Label
`"flex items-center gap-2 text-sm leading-none font-medium select-none group-data-[disabled=true]:pointer-events-none group-data-[disabled=true]:opacity-50 peer-disabled:cursor-not-allowed peer-disabled:opacity-50"` → 14px font, line-height 1, weight 500, gap 8px; 50% opacity when peer/group disabled.

### Checkbox
Root: `"peer size-4 shrink-0 rounded-[4px] border border-input shadow-xs transition-shadow outline-none ... data-[state=checked]:border-primary data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground dark:bg-input/30 ... dark:data-[state=checked]:bg-primary"`; indicator `"grid place-content-center text-current transition-none"` with `<CheckIcon className="size-3.5" />`.
px: box **16×16**, radius **4px** (literal), border 1px `--input` (checked: `--primary`), check icon **14px** stroke in `--primary-foreground`, shadow-xs, dark unchecked fill `--input`@30%.

### Radio Group
Group: `"grid gap-3"` (12px gap). Item: `"aspect-square size-4 shrink-0 rounded-full border border-input text-primary shadow-xs ... dark:bg-input/30"`; indicator: centered `CircleIcon` `"size-2 ... fill-primary"`. px: circle **16×16**, 1px `--input` border, inner dot **8px** filled `--primary`.

### Switch (now has 2 sizes)
Track: `"... rounded-full border border-transparent shadow-xs ... data-[size=default]:h-[1.15rem] data-[size=default]:w-8 data-[size=sm]:h-3.5 data-[size=sm]:w-6 data-[state=checked]:bg-primary data-[state=unchecked]:bg-input dark:data-[state=unchecked]:bg-input/80"`
Thumb: `"pointer-events-none block rounded-full bg-background ring-0 transition-transform group-data-[size=default]/switch:size-4 group-data-[size=sm]/switch:size-3 data-[state=checked]:translate-x-[calc(100%-2px)] data-[state=unchecked]:translate-x-0 dark:data-[state=checked]:bg-primary-foreground dark:data-[state=unchecked]:bg-foreground"`
px: default track **32×18.4** (h = 1.15rem), thumb **16px**, checked translateX = thumb width − 2px = **14px**; sm track **24×14**, thumb **12px** (translate 10px). Track border 1px transparent. Colors: on `--primary`, off `--input` (dark off `--input`@80%); thumb light `--background`; dark thumb on `--primary-foreground`, off `--foreground`.

### Card
`Card`: `"flex flex-col gap-6 rounded-xl border bg-card py-6 text-card-foreground shadow-sm"` → radius **14px**, 1px border, vertical padding 24px, internal gap 24px, shadow-sm.
`CardHeader`: `"@container/card-header grid auto-rows-min grid-rows-[auto_auto] items-start gap-2 px-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6"` → px 24px, title/desc gap 8px.
`CardTitle`: `"leading-none font-semibold"` (16px/16px, 600). `CardDescription`: `"text-sm text-muted-foreground"`. `CardAction`: top-right grid cell. `CardContent`: `"px-6"`. `CardFooter`: `"flex items-center px-6 [.border-t]:pt-6"`.

### Separator
`"shrink-0 bg-border data-[orientation=horizontal]:h-px data-[orientation=horizontal]:w-full data-[orientation=vertical]:h-full data-[orientation=vertical]:w-px"` → **1px** line in `--border`.

### Alert
Base: `"relative grid w-full grid-cols-[0_1fr] items-start gap-y-0.5 rounded-lg border px-4 py-3 text-sm has-[>svg]:grid-cols-[calc(var(--spacing)*4)_1fr] has-[>svg]:gap-x-3 [&>svg]:size-4 [&>svg]:translate-y-0.5 [&>svg]:text-current"`
Variants: `default: "bg-card text-card-foreground"`; `destructive: "bg-card text-destructive *:data-[slot=alert-description]:text-destructive/90 [&>svg]:text-current"`.
Title: `"col-start-2 line-clamp-1 min-h-4 font-medium tracking-tight"`. Description: `"col-start-2 grid justify-items-start gap-1 text-sm text-muted-foreground [&_p]:leading-relaxed"`.
px: radius 10px, border 1px, px 16px, py 12px, icon 16px nudged down 2px, icon column 16px + 12px gap, row gap 2px, title 14px/500/−0.025em, description 14px muted (destructive: destructive@90%).

### Avatar (sized)
Root: `"... size-8 ... rounded-full ... data-[size=lg]:size-10 data-[size=sm]:size-6"` → sm **24**, default **32**, lg **40**; image fills; fallback `"... bg-muted text-sm text-muted-foreground ..."` (12px text at sm). AvatarGroup overlaps with `-space-x-2` (−8px) and 2px background ring per avatar.

### Skeleton
`"animate-pulse rounded-md bg-accent"` → fill `--accent`, radius 8px, animation: `pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite` with keyframe `50% { opacity: 0.5 }`.

### Progress
Root: `"relative h-2 w-full overflow-hidden rounded-full bg-primary/20"`; indicator `"h-full w-full flex-1 bg-primary transition-all"` with `transform: translateX(-${100 - value}%)`. px: height **8px**, full pill radius, track = `--primary` @20%, bar = `--primary`.

### Slider
Track: `"... rounded-full bg-muted data-[orientation=horizontal]:h-1.5 ... data-[orientation=vertical]:w-1.5"` → **6px** thick, `--muted`. Range: `bg-primary`. Thumb: `"block size-4 shrink-0 rounded-full border border-primary bg-white shadow-sm ring-ring/50 transition-[color,box-shadow] hover:ring-4 focus-visible:ring-4 focus-visible:outline-hidden disabled:pointer-events-none disabled:opacity-50"` → **16px** circle, 1px `--primary` border, **always white** fill, shadow-sm, hover/focus 4px ring `--ring`@50%. Vertical: min-h 176px. Disabled root opacity 50%.

### Tabs
Root: `"group/tabs flex gap-2 data-[orientation=horizontal]:flex-col"` (8px gap to content).
List (cva): `"group/tabs-list inline-flex w-fit items-center justify-center rounded-lg p-[3px] text-muted-foreground group-data-[orientation=horizontal]/tabs:h-9 ..."`; variants `default: "bg-muted"`, `line: "gap-1 bg-transparent"`. → list h **36px**, radius 10px, padding **3px**, fill `--muted`.
Trigger (verbatim, key parts): `"relative inline-flex h-[calc(100%-1px)] flex-1 items-center justify-center gap-1.5 rounded-md border border-transparent px-2 py-1 text-sm font-medium whitespace-nowrap text-foreground/60 transition-all ... focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:outline-1 focus-visible:outline-ring ... group-data-[variant=default]/tabs-list:data-[state=active]:shadow-sm ... dark:text-muted-foreground dark:hover:text-foreground [&_svg:not([class*='size-'])]:size-4"` + `"data-[state=active]:bg-background data-[state=active]:text-foreground dark:data-[state=active]:border-input dark:data-[state=active]:bg-input/30 dark:data-[state=active]:text-foreground"` + line-variant underline: `"after:absolute after:bg-foreground ... after:bottom-[-5px] after:h-0.5 ... group-data-[variant=line]/tabs-list:data-[state=active]:after:opacity-100"`.
px: trigger height = list inner − 1px (≈ 29px), radius 8px, px 8px, py 4px, font 14px/500, idle text `foreground`@60% (dark: muted-foreground); active: bg `--background`, shadow-sm (dark: 1px `--input` border + `--input`@30% fill). Line variant: 2px underline 5px below, in `--foreground`.

### Tooltip
Content: `"z-50 w-fit origin-(--radix-tooltip-content-transform-origin) animate-in rounded-md bg-foreground px-3 py-1.5 text-xs text-balance text-background fade-in-0 zoom-in-95 data-[side=...]:slide-in-from-*-2 ..."`; Arrow: `"z-50 size-2.5 translate-y-[calc(-50%_-_2px)] rotate-45 rounded-[2px] bg-foreground fill-foreground"`.
px: bg **`--foreground`**, text **`--background`** (inverted — current version; the older one used primary), px 12px, py 6px, radius 8px, font 12px/16px; arrow = 10px square rotated 45°, 2px corner radius. `delayDuration = 0`.

### Popover
Content: `"z-50 w-72 origin-(--radix-popover-content-transform-origin) rounded-md border bg-popover p-4 text-popover-foreground shadow-md outline-hidden ... fade-in-0 zoom-in-95"` → width **288px**, padding **16px**, radius 8px, 1px border, shadow-md, side offset 4px.

### Dialog
Overlay: `"fixed inset-0 z-50 bg-black/50 ..."` → **pure black @ 50%** both modes.
Content: `"fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] gap-4 rounded-lg border bg-background p-6 shadow-lg duration-200 outline-none ... data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95 sm:max-w-lg"` → centered, max-w **512px** (mobile: viewport − 32px), padding **24px**, internal gap 16px, radius 10px, 1px border, bg `--background`, shadow-lg; open/close anim 200ms fade + scale 0.95↔1.
Close button: `"absolute top-4 right-4 rounded-xs opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:ring-2 focus:ring-ring focus:ring-offset-2 focus:outline-hidden ... [&_svg:not([class*='size-'])]:size-4"` → 16px from corner, 2px radius, 70%→100% opacity, X icon 16px, focus = 2px ring + 2px offset.
Header `"flex flex-col gap-2 text-center sm:text-left"`; Footer `"flex flex-col-reverse gap-2 sm:flex-row sm:justify-end"`; Title `"text-lg leading-none font-semibold"` (18px/18px, 600); Description `"text-sm text-muted-foreground"`.

### Select
Trigger: `"flex w-fit items-center justify-between gap-2 rounded-md border border-input bg-transparent px-3 py-2 text-sm whitespace-nowrap shadow-xs transition-[color,box-shadow] ... data-[placeholder]:text-muted-foreground data-[size=default]:h-9 data-[size=sm]:h-8 ... dark:bg-input/30 dark:hover:bg-input/50 ... [&_svg:not([class*='size-'])]:size-4 [&_svg:not([class*='text-'])]:text-muted-foreground"` + chevron `"size-4 opacity-50"`. px: h **36px** (sm **32px**), px 12px, gap 8px, radius 8px, 1px `--input` border, shadow-xs, font 14px, chevron 16px @50% opacity.
Content: `"... min-w-[8rem] origin-... overflow-x-hidden overflow-y-auto rounded-md border bg-popover text-popover-foreground shadow-md ..."` + popper position adds 4px translate toward side; viewport `"p-1"`. px: min-w 128px, radius 8px, 1px border, shadow-md, 4px inner padding.
Item: `"relative flex w-full cursor-default items-center gap-2 rounded-sm py-1.5 pr-8 pl-2 text-sm outline-hidden select-none focus:bg-accent focus:text-accent-foreground data-[disabled]:opacity-50 ... [&_svg:not([class*='size-'])]:size-4 ..."`; check indicator: `"absolute right-2 flex size-3.5 ..."` with `CheckIcon size-4`. px: py 6px, pl 8px, pr 32px, radius 6px, font 14px; check 16px at right 8px.
Label `"px-2 py-1.5 text-xs text-muted-foreground"`; Separator `"-mx-1 my-1 h-px bg-border"`; scroll buttons `py-1` with 16px chevrons.

### Dropdown Menu
Content: `"z-50 max-h-(--radix-dropdown-menu-content-available-height) min-w-[8rem] ... rounded-md border bg-popover p-1 text-popover-foreground shadow-md ..."` → min-w **128px**, padding 4px, radius 8px, shadow-md, side offset 4px. SubContent identical but `shadow-lg`.
Item: `"relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 data-[inset]:pl-8 data-[variant=destructive]:text-destructive data-[variant=destructive]:focus:bg-destructive/10 ... dark:data-[variant=destructive]:focus:bg-destructive/20 ... [&_svg:not([class*='size-'])]:size-4 [&_svg:not([class*='text-'])]:text-muted-foreground ..."` → px 8px, py 6px, radius 6px, font 14px, gap 8px, icons 16px muted; hover/focus = `--accent` fill; destructive item: text `--destructive`, focus fill destructive@10% (dark 20%); inset pl 32px.
CheckboxItem/RadioItem: `"... py-1.5 pr-2 pl-8 ..."` with indicator at `left-2` in a 14px box; check 16px, radio dot `CircleIcon size-2 fill-current` (8px).
Label: `"px-2 py-1.5 text-sm font-medium data-[inset]:pl-8"`. Separator: `"-mx-1 my-1 h-px bg-border"`. Shortcut: `"ml-auto text-xs tracking-widest text-muted-foreground"`.

### Table
Table `"w-full caption-bottom text-sm"`; Header `"[&_tr]:border-b"`; Body `"[&_tr:last-child]:border-0"`; Footer `"border-t bg-muted/50 font-medium [&>tr]:last:border-b-0"`; Row `"border-b transition-colors hover:bg-muted/50 has-aria-expanded:bg-muted/50 data-[state=selected]:bg-muted"`; Head `"h-10 px-2 text-left align-middle font-medium whitespace-nowrap text-foreground [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]"`; Cell `"p-2 align-middle whitespace-nowrap ..."`; Caption `"mt-4 text-sm text-muted-foreground"`.
px: font 14px; header cell h **40px**, px 8px, weight 500; body cell padding **8px**; 1px row borders; hover row `--muted`@50%; selected `--muted` full.

### Toggle + Toggle Group
Toggle base: `"inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium whitespace-nowrap transition-[color,box-shadow] outline-none hover:bg-muted hover:text-muted-foreground ... data-[state=on]:bg-accent data-[state=on]:text-accent-foreground ... [&_svg:not([class*='size-'])]:size-4"`; variants: `default: "bg-transparent"`, `outline: "border border-input bg-transparent shadow-xs hover:bg-accent hover:text-accent-foreground"`; sizes: `default: "h-9 min-w-9 px-2"` (36px, px 8px), `sm: "h-8 min-w-8 px-1.5"` (32px, px 6px), `lg: "h-10 min-w-10 px-2.5"` (40px, px 10px). On-state = `--accent` fill.
ToggleGroup root: `"group/toggle-group flex w-fit items-center gap-[--spacing(var(--gap))] rounded-md data-[spacing=default]:data-[variant=outline]:shadow-xs"`. Items add: `"w-auto min-w-0 shrink-0 px-3 focus:z-10 focus-visible:z-10"` and at spacing 0: `"data-[spacing=0]:rounded-none data-[spacing=0]:shadow-none data-[spacing=0]:first:rounded-l-md data-[spacing=0]:last:rounded-r-md data-[spacing=0]:data-[variant=outline]:border-l-0 data-[spacing=0]:data-[variant=outline]:first:border-l"` → fused segmented control, shared 1px borders, 8px outer corners, item px 12px.

### Kbd
`"pointer-events-none inline-flex h-5 w-fit min-w-5 items-center justify-center gap-1 rounded-sm bg-muted px-1 font-sans text-xs font-medium text-muted-foreground select-none"` (+ icons 12px; inside tooltips: `bg-background/20 text-background`, dark `bg-background/10`). px: h **20px**, min-w 20px, px 4px, radius 6px, font 12px/500.

### Spinner
`Loader2Icon` with `"size-4 animate-spin"` → 16px icon, `spin 1s linear infinite`.

### Breadcrumb
List: `"flex flex-wrap items-center gap-1.5 text-sm break-words text-muted-foreground sm:gap-2.5"` (gap 6px → 10px ≥640px, font 14px, muted). Item: `"inline-flex items-center gap-1.5"`. Link: `"transition-colors hover:text-foreground"`. Page: `"font-normal text-foreground"`. Separator: `"[&>svg]:size-3.5"` (ChevronRight 14px). Ellipsis: `"flex size-9 items-center justify-center"` (36px hit area, 16px icon).

### Pagination
Content: `"flex flex-row items-center gap-1"` (4px). Links reuse buttonVariants: `variant: isActive ? "outline" : "ghost"`, default `size="icon"` → 36×36; active page = outline button (1px border + shadow-xs), others ghost. Prev/Next: `size="default"` + `"gap-1 px-2.5 sm:pl-2.5"` with 16px chevrons. Ellipsis `"flex size-9 ..."` 36px box.

### Accordion
Item: `"border-b last:border-b-0"`. Trigger: `"flex flex-1 items-start justify-between gap-4 rounded-md py-4 text-left text-sm font-medium transition-all outline-none hover:underline focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 [&[data-state=open]>svg]:rotate-180"` + chevron `"pointer-events-none size-4 shrink-0 translate-y-0.5 text-muted-foreground transition-transform duration-200"`. Content: `"overflow-hidden text-sm data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down"`, inner `"pt-0 pb-4"`.
px: trigger py **16px**, font 14px/500, gap 16px, hover underlines; chevron 16px, rotates 180° in 200ms; content collapse animation `accordion-down/up` = height 0 ↔ `var(--radix-accordion-content-height)` over **0.2s ease-out**; content bottom padding 16px.

### Scroll Area
Viewport: focus ring `"focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:outline-1"`. Scrollbar: `"flex touch-none p-px transition-colors select-none"` + vertical `"h-full w-2.5 border-l border-l-transparent"` / horizontal `"h-2.5 flex-col border-t border-t-transparent"`; Thumb: `"relative flex-1 rounded-full bg-border"`. px: scrollbar gutter **10px** wide with 1px padding → thumb ~8px, pill-shaped, color `--border`.

### Sheet
Overlay identical to Dialog (`bg-black/50`). Content base: `"fixed z-50 flex flex-col gap-4 bg-background shadow-lg transition ease-in-out data-[state=closed]:animate-out data-[state=closed]:duration-300 data-[state=open]:animate-in data-[state=open]:duration-500"`. Sides: right/left = `"inset-y-0 ... h-full w-3/4 border-l|border-r ... sm:max-w-sm"` (width 75% of viewport, max **384px**, 1px border, slide-in-from-right/left); top/bottom = `"inset-x-0 ... h-auto border-b|border-t"` (auto height). Open anim **500ms**, close **300ms**, ease-in-out slide. Header `"flex flex-col gap-1.5 p-4"` (16px padding, 6px gap); Footer `"mt-auto flex flex-col gap-2 p-4"`; Title `"font-semibold text-foreground"`; Description `"text-sm text-muted-foreground"`; same close button as Dialog (top/right 16px).

### Command (cmdk)
Root: `"flex h-full w-full flex-col overflow-hidden rounded-md bg-popover text-popover-foreground"`. Input wrapper: `"flex h-9 items-center gap-2 border-b px-3"` (h 36px, px 12px, 16px search icon @50% opacity); input: `"flex h-10 w-full rounded-md bg-transparent py-3 text-sm outline-hidden placeholder:text-muted-foreground ..."`. List: `"max-h-[300px] scroll-py-1 overflow-x-hidden overflow-y-auto"`. Empty: `"py-6 text-center text-sm"`. Group: `"overflow-hidden p-1 text-foreground [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-xs [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:text-muted-foreground"`. Item: `"relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none data-[disabled=true]:opacity-50 data-[selected=true]:bg-accent data-[selected=true]:text-accent-foreground [&_svg:not([class*='size-'])]:size-4 [&_svg:not([class*='text-'])]:text-muted-foreground"`. Separator `"-mx-1 h-px bg-border"`. In CommandDialog: input wrapper h **48px**, items `px-2 py-3` (12px), icons 20px.

### Calendar (react-day-picker v9 wrapper) — basics
Cell size driven by `--cell-size` (default `--spacing(8)` = **32px**). Key classNames (verbatim): `nav: "absolute inset-x-0 top-0 flex w-full items-center justify-between gap-1"`, `button_previous/next: "size-(--cell-size) p-0 ..."` (ghost icon buttons, 32px), `month_caption: "flex h-(--cell-size) ... justify-center"`, `weekday: "flex-1 rounded-md text-[0.8rem] font-normal text-muted-foreground select-none"` (12.8px), `week: "mt-2 flex w-full"`, `day: "group/day relative aspect-square h-full w-full p-0 text-center select-none"`. DayButton: `"flex aspect-square size-auto w-full min-w-(--cell-size) flex-col gap-1 leading-none font-normal ... data-[selected-single=true]:bg-primary data-[selected-single=true]:text-primary-foreground data-[range-middle=true]:rounded-none data-[range-middle=true]:bg-accent data-[range-start=true]:rounded-md data-[range-start=true]:bg-primary ... group-data-[focused=true]/day:border-ring group-data-[focused=true]/day:ring-[3px] group-data-[focused=true]/day:ring-ring/50"` — selected day = `--primary` pill (8px radius), range middle = `--accent` square, focus = standard 3px ring.

### Sonner (Toaster)
Maps CSS vars: `--normal-bg: var(--popover)`, `--normal-text: var(--popover-foreground)`, `--normal-border: var(--border)`, `--border-radius: var(--radius)` (10px). Status icons all 16px lucide (`CircleCheckIcon`, `InfoIcon`, `TriangleAlertIcon`, `OctagonXIcon`, `Loader2Icon` spinning). Sonner defaults otherwise (width 356px, padding 16px, shadow ≈ lg).

---

## 6. FOCUS RING — exact treatment

Standard recipe on nearly every interactive control (button, badge-link, input, textarea, checkbox, radio, switch, select trigger, toggle, tabs trigger, accordion trigger, calendar day):

```
outline-none  focus-visible:border-ring  focus-visible:ring-[3px]  focus-visible:ring-ring/50
```

= remove native outline; on keyboard focus draw a **3px** ring (Tailwind ring = `box-shadow: 0 0 0 3px <color>`, **no offset**) in `--ring` at **50% alpha**, AND change the element's border color to solid `--ring`. Light `--ring` = `oklch(0.708 0 0)`; dark = `oklch(0.556 0 0)`.

Variations:
- **Invalid state**: `aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40` → 3px ring `--destructive` @20% (dark @40%), border solid destructive.
- **Destructive button**: focus ring uses `destructive/20` (dark `/40`) instead of `ring/50`.
- **Dialog/Sheet close buttons** (older recipe): `focus:ring-2 focus:ring-ring focus:ring-offset-2` → 2px solid ring with 2px offset (offset color = `--background`).
- **Tabs trigger** additionally: `focus-visible:outline-1 focus-visible:outline-ring` (1px outline on top of ring).
- **Slider thumb**: `hover:ring-4 focus-visible:ring-4` with `ring-ring/50` → 4px ring, also on hover.
- Global base: `* { outline-color: ring/50 }` (`outline-ring/50`).

---

## 7. STATE LAYERS — exact hover/active/disabled formulas

| Context | Class | Formula |
|---|---|---|
| Primary button hover | `hover:bg-primary/90` | `--primary` at 90% alpha (composited over page bg) |
| Secondary button hover | `hover:bg-secondary/80` | `--secondary` @80% |
| Destructive button/badge hover | `hover:bg-destructive/90` | `--destructive` @90% |
| Badge (link) hovers | `[a&]:hover:bg-primary/90`, `secondary/90` | @90% |
| Outline/ghost hover | `hover:bg-accent hover:text-accent-foreground` | solid `--accent` fill (no alpha) |
| Ghost dark hover | `dark:hover:bg-accent/50` | `--accent` @50% |
| Outline dark hover | `dark:hover:bg-input/50` | `--input` @50% |
| Menu/select/command item focus/selected | `focus:bg-accent` / `data-[selected=true]:bg-accent` | solid `--accent` |
| Destructive menu item focus | `focus:bg-destructive/10` (dark `/20`) | destructive @10/20% |
| Table row hover | `hover:bg-muted/50` | `--muted` @50% |
| Table row selected | `bg-muted` | solid |
| Toggle hover | `hover:bg-muted hover:text-muted-foreground` | solid `--muted` |
| Toggle on | `bg-accent` | solid |
| Tabs trigger hover | `hover:text-foreground` (text only) | — |
| Link button/badge hover | `hover:underline` (offset 4px) | — |
| Disabled (universal) | `disabled:opacity-50` + `disabled:pointer-events-none` (or `cursor-not-allowed` on form fields) | 50% element opacity |
| Dialog/Sheet close idle/hover | `opacity-70` → `hover:opacity-100` | |
| Press/active | no dedicated active styles in registry; docs site only: `button:active { opacity: .6 }` below `md` breakpoint |

Transitions: buttons `transition-all`; form fields `transition-[color,box-shadow]`; tables `transition-colors`; default Tailwind duration 150ms ease (`cubic-bezier(0.4,0,0.2,1)`); accordion chevron 200ms; dialog content `duration-200`; sheet 500ms in / 300ms out.

---

## 8. DARK MODE DIFFERENCES BEYOND VARIABLES

1. **Input-family fill**: light mode fields are `bg-transparent`; dark adds `dark:bg-input/30` (white@15% var at 30% alpha ≈ 4.5% white) — on input, textarea, checkbox, radio item, select trigger, outline button. Hover on outline button & select trigger: `dark:bg-input/50`.
2. **Destructive solid surfaces dimmed**: `dark:bg-destructive/60` on destructive button and badge.
3. **Invalid/destructive rings brighter**: `/20` → `dark:.../40` everywhere.
4. **Switch**: `dark:data-[state=unchecked]:bg-input/80`; thumb swaps to `dark:data-[state=checked]:bg-primary-foreground` / `dark:data-[state=unchecked]:bg-foreground`.
5. **Tabs active trigger**: gains visible border+fill — `dark:data-[state=active]:border-input dark:data-[state=active]:bg-input/30`; idle text uses `dark:text-muted-foreground`.
6. **Ghost button hover** is translucent in dark: `dark:hover:bg-accent/50`.
7. **Destructive menu item focus** `dark:...:bg-destructive/20`.
8. **Borders are alpha in dark**: `--border: oklch(1 0 0 / 10%)`, `--input: oklch(1 0 0 / 15%)` — i.e., 10%/15% white over surface (light mode uses opaque gray `oklch(0.922 0 0)`).
9. Dialog/sheet overlay stays `bg-black/50` in both modes; tooltip auto-inverts via `bg-foreground/text-background`.
10. Checkbox keeps `dark:data-[state=checked]:bg-primary` (explicit override so the /30 fill doesn't bleed through when checked).

---

## Master px cheat-sheet (most-used classes in this registry)

| Class | Value | Class | Value |
|---|---|---|---|
| `h-9` | 36px | `rounded-sm` | 6px |
| `h-8` | 32px | `rounded-md` | 8px |
| `h-10` | 40px | `rounded-lg` | 10px |
| `h-6` | 24px | `rounded-xl` | 14px |
| `size-4` | 16px | `rounded-xs` | 2px |
| `size-3.5` | 14px | `text-xs` | 12px/16px |
| `size-3` | 12px | `text-sm` | 14px/20px |
| `px-2 py-1.5` (menu item) | 8px / 6px | `text-base` | 16px/24px |
| `p-1` (menu padding) | 4px | `text-lg` | 18px/28px |
| `p-4` | 16px | `font-medium` | 500 |
| `p-6` | 24px | `font-semibold` | 600 |
| `gap-2` | 8px | `ring-[3px]` | 3px box-shadow spread |
| `border` | 1px solid `--border` | `underline-offset-4` | 4px |

**Fetch failures / fallbacks:** `packages/shadcn/tailwind.css` raw path 404'd (the `shadcn/tailwind.css` import is published only in the npm package); its radius/color mapping content was obtained instead from the `/docs/installation/manual` generated CSS, which is authoritative for what end-user apps get. All component files were fetched successfully from `apps/v4/registry/new-york-v4/ui/` on `main` — no fallback to the legacy `apps/www/registry/new-york/ui/` location was needed.
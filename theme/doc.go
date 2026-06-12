// Package theme implements the graft design system: a pixel-faithful Go
// port of the shadcn/ui (new-york-v4) token model on top of gogpu/ui.
//
// The package mirrors the shadcn theming workflow one-to-one:
//
//   - [Tokens] is a 1:1 mirror of the CSS variables in shadcn's :root and
//     .dark blocks (--background, --primary, --ring, ...).
//   - [New] builds a [Theme] from one of the five shadcn base-color presets
//     (Neutral, Stone, Zinc, Gray, Slate) with the --radius knob and font
//     families, exactly like configuring components.json.
//   - Token fields can be overridden directly per mode, or a whole theme can
//     be pasted from the shadcn/tweakcn theme editor via [ParseThemeCSS] /
//     [Theme.ApplyCSS].
//   - [OKLCH], [OKLCHA], [ParseOKLCH] and [MustOKLCH] convert CSS Color 4
//     oklch() literals to sRGB so shadcn token values copy verbatim.
//
// A Theme carries both light and dark token sets. [Theme.Active] resolves
// the current set from the mode, so painters that read tokens at draw time
// switch modes without any tree rebuild. [Theme.AsUITheme] adapts the theme
// to gogpu/ui's theme.Theme for window background and primitive defaults,
// and Theme itself satisfies widget.ThemeProvider.
package theme

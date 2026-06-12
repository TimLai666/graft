package theme

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gogpu/ui/widget"
)

// ParseThemeCSS builds a Theme from shadcn/tweakcn theme-editor CSS. It
// starts from the Neutral preset and merges every variable found in the
// :root and .dark blocks, so partial exports work and missing variables
// keep their stock values:
//
//	th, err := theme.ParseThemeCSS(cssText)
//
// Supported value syntax per variable is documented on Theme.ApplyCSS.
func ParseThemeCSS(css string) (*Theme, error) {
	t := New()
	if err := t.ApplyCSS(css); err != nil {
		return nil, err
	}
	return t, nil
}

// ApplyCSS merges shadcn-style CSS custom-property blocks onto the theme.
// Only the variables present in the CSS are touched (partial override).
//
// Selectors containing ":root" or ".light" (also bare "html"/"body")
// target the Light token set; selectors containing ".dark" target Dark.
// Comma groups like ":root, .light" are honored, comments and unknown
// selectors are tolerated, and a missing .dark block simply leaves Dark
// unchanged.
//
// Recognized variables are the kebab-case shadcn token names (background,
// card-foreground, chart-1..chart-5, sidebar-ring, ...) plus radius.
// Unknown variables are ignored for forward compatibility; a recognized
// variable with an unparseable value is an error.
//
// Color values may be oklch(), hsl()/hsla() (including the legacy bare
// "H S% L%" triplet form), rgb()/rgba(), or #hex (3/4/6/8 digits).
// --radius accepts rem (multiplied by 16), px, or a unitless pixel number.
func (t *Theme) ApplyCSS(css string) error {
	rest := stripCSSComments(css)
	for {
		open := strings.IndexByte(rest, '{')
		if open < 0 {
			if strings.ContainsRune(rest, '}') {
				return fmt.Errorf("theme: unbalanced braces in theme CSS")
			}
			return nil
		}
		selector := strings.TrimSpace(rest[:open])

		// Find the matching close brace (depth-aware so an at-rule body
		// is skipped as a unit rather than torn apart).
		depth := 1
		i := open + 1
		for i < len(rest) && depth > 0 {
			switch rest[i] {
			case '{':
				depth++
			case '}':
				depth--
			}
			i++
		}
		if depth != 0 {
			return fmt.Errorf("theme: unbalanced braces in theme CSS")
		}
		body := rest[open+1 : i-1]
		rest = rest[i:]

		light, dark := classifySelector(selector)
		if !light && !dark {
			continue
		}
		if err := t.applyDeclarations(body, light, dark); err != nil {
			return err
		}
	}
}

// classifySelector decides which token set(s) a CSS block targets.
func classifySelector(selector string) (light, dark bool) {
	for _, part := range strings.Split(selector, ",") {
		part = strings.TrimSpace(part)
		switch {
		case strings.Contains(part, ".dark"):
			dark = true
		case strings.Contains(part, ":root"), strings.Contains(part, ".light"),
			part == "html", part == "body":
			light = true
		}
	}
	return light, dark
}

// applyDeclarations parses "--name: value;" declarations from a block body
// and writes recognized tokens into the targeted set(s).
func (t *Theme) applyDeclarations(body string, light, dark bool) error {
	for _, decl := range strings.Split(body, ";") {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}
		name, value, ok := strings.Cut(decl, ":")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if !strings.HasPrefix(name, "--") || value == "" {
			continue
		}
		name = strings.TrimPrefix(name, "--")

		if name == "radius" {
			px, err := parseCSSLength(value)
			if err != nil {
				return fmt.Errorf("theme: bad --radius value %q: %w", value, err)
			}
			t.Radius = px
			continue
		}

		field, known := tokenField(name)
		if !known {
			continue // forward compatibility: ignore unknown variables
		}
		c, err := parseCSSColor(value)
		if err != nil {
			return fmt.Errorf("theme: bad value for --%s: %w", name, err)
		}
		if light {
			*field(&t.Light) = c
		}
		if dark {
			*field(&t.Dark) = c
		}
	}
	return nil
}

// tokenField maps a kebab-case shadcn variable name to the matching
// Tokens field.
func tokenField(name string) (func(*Tokens) *widget.Color, bool) {
	f, ok := tokenFields[name]
	return f, ok
}

var tokenFields = map[string]func(*Tokens) *widget.Color{
	"background":                 func(tk *Tokens) *widget.Color { return &tk.Background },
	"foreground":                 func(tk *Tokens) *widget.Color { return &tk.Foreground },
	"card":                       func(tk *Tokens) *widget.Color { return &tk.Card },
	"card-foreground":            func(tk *Tokens) *widget.Color { return &tk.CardForeground },
	"popover":                    func(tk *Tokens) *widget.Color { return &tk.Popover },
	"popover-foreground":         func(tk *Tokens) *widget.Color { return &tk.PopoverForeground },
	"primary":                    func(tk *Tokens) *widget.Color { return &tk.Primary },
	"primary-foreground":         func(tk *Tokens) *widget.Color { return &tk.PrimaryForeground },
	"secondary":                  func(tk *Tokens) *widget.Color { return &tk.Secondary },
	"secondary-foreground":       func(tk *Tokens) *widget.Color { return &tk.SecondaryForeground },
	"muted":                      func(tk *Tokens) *widget.Color { return &tk.Muted },
	"muted-foreground":           func(tk *Tokens) *widget.Color { return &tk.MutedForeground },
	"accent":                     func(tk *Tokens) *widget.Color { return &tk.Accent },
	"accent-foreground":          func(tk *Tokens) *widget.Color { return &tk.AccentForeground },
	"destructive":                func(tk *Tokens) *widget.Color { return &tk.Destructive },
	"destructive-foreground":     func(tk *Tokens) *widget.Color { return &tk.DestructiveForeground },
	"border":                     func(tk *Tokens) *widget.Color { return &tk.Border },
	"input":                      func(tk *Tokens) *widget.Color { return &tk.Input },
	"ring":                       func(tk *Tokens) *widget.Color { return &tk.Ring },
	"chart-1":                    func(tk *Tokens) *widget.Color { return &tk.Chart[0] },
	"chart-2":                    func(tk *Tokens) *widget.Color { return &tk.Chart[1] },
	"chart-3":                    func(tk *Tokens) *widget.Color { return &tk.Chart[2] },
	"chart-4":                    func(tk *Tokens) *widget.Color { return &tk.Chart[3] },
	"chart-5":                    func(tk *Tokens) *widget.Color { return &tk.Chart[4] },
	"sidebar":                    func(tk *Tokens) *widget.Color { return &tk.Sidebar },
	"sidebar-foreground":         func(tk *Tokens) *widget.Color { return &tk.SidebarForeground },
	"sidebar-primary":            func(tk *Tokens) *widget.Color { return &tk.SidebarPrimary },
	"sidebar-primary-foreground": func(tk *Tokens) *widget.Color { return &tk.SidebarPrimaryForeground },
	"sidebar-accent":             func(tk *Tokens) *widget.Color { return &tk.SidebarAccent },
	"sidebar-accent-foreground":  func(tk *Tokens) *widget.Color { return &tk.SidebarAccentForeground },
	"sidebar-border":             func(tk *Tokens) *widget.Color { return &tk.SidebarBorder },
	"sidebar-ring":               func(tk *Tokens) *widget.Color { return &tk.SidebarRing },
}

// stripCSSComments removes /* ... */ comments. An unterminated comment
// swallows the rest of the input, matching browser error recovery.
func stripCSSComments(s string) string {
	var b strings.Builder
	for {
		i := strings.Index(s, "/*")
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		j := strings.Index(s[i+2:], "*/")
		if j < 0 {
			return b.String()
		}
		s = s[i+2+j+2:]
	}
}

// parseCSSLength parses a CSS length as pixels: "0.625rem" (1rem = 16px),
// "10px", or a unitless number treated as pixels.
func parseCSSLength(s string) (float32, error) {
	s = strings.TrimSpace(s)
	scale := 1.0
	switch {
	case strings.HasSuffix(s, "rem"):
		s = strings.TrimSuffix(s, "rem")
		scale = 16
	case strings.HasSuffix(s, "px"):
		s = strings.TrimSuffix(s, "px")
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, err
	}
	return float32(v * scale), nil
}

// parseCSSColor parses the CSS color syntaxes that appear in shadcn and
// theme-editor exports: oklch(), hsl()/hsla(), rgb()/rgba(), #hex, and the
// legacy shadcn bare HSL triplet ("224 71.4% 4.1%").
func parseCSSColor(s string) (widget.Color, error) {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "oklch("):
		return ParseOKLCH(lower)
	case strings.HasPrefix(lower, "hsl("), strings.HasPrefix(lower, "hsla("):
		return parseHSLColor(s)
	case strings.HasPrefix(lower, "rgb("), strings.HasPrefix(lower, "rgba("):
		return parseRGBColor(s)
	case strings.HasPrefix(s, "#"):
		return parseHexColor(s)
	}
	// Legacy shadcn (Tailwind v3) HSL triplet without the hsl() wrapper.
	if f := strings.Fields(s); len(f) == 3 &&
		strings.HasSuffix(f[1], "%") && strings.HasSuffix(f[2], "%") {
		return parseHSLColor("hsl(" + s + ")")
	}
	return widget.Color{}, fmt.Errorf("unsupported CSS color %q", s)
}

// parenBody returns the text between the first '(' and the last ')'.
func parenBody(s string) (string, error) {
	open := strings.IndexByte(s, '(')
	closing := strings.LastIndexByte(s, ')')
	if open < 0 || closing < open {
		return "", fmt.Errorf("malformed functional color %q", s)
	}
	return s[open+1 : closing], nil
}

// splitColorComponents splits a functional color body into its three
// channel components plus alpha, handling both modern space-separated
// syntax with "/ alpha" and legacy comma syntax with a fourth component.
func splitColorComponents(body string) (comps []string, alpha float64, err error) {
	alpha = 1
	main := body
	if i := strings.IndexByte(body, '/'); i >= 0 {
		alpha, err = parseAlphaValue(strings.TrimSpace(body[i+1:]))
		if err != nil {
			return nil, 0, err
		}
		main = body[:i]
	}

	if strings.ContainsRune(main, ',') {
		for _, p := range strings.Split(main, ",") {
			comps = append(comps, strings.TrimSpace(p))
		}
	} else {
		comps = strings.Fields(main)
	}

	if len(comps) == 4 { // hsla(h, s, l, a) / rgba(r, g, b, a)
		alpha, err = parseAlphaValue(comps[3])
		if err != nil {
			return nil, 0, err
		}
		comps = comps[:3]
	}
	if len(comps) != 3 {
		return nil, 0, fmt.Errorf("expected 3 color components, got %d", len(comps))
	}
	return comps, alpha, nil
}

// parseAlphaValue parses an alpha component: "50%" or "0.5".
func parseAlphaValue(s string) (float64, error) {
	if p, ok := strings.CutSuffix(s, "%"); ok {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return 0, err
		}
		return clamp01(v / 100), nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return clamp01(v), nil
}

// parseHSLColor parses hsl()/hsla() in modern or legacy syntax.
func parseHSLColor(s string) (widget.Color, error) {
	body, err := parenBody(s)
	if err != nil {
		return widget.Color{}, err
	}
	comps, alpha, err := splitColorComponents(body)
	if err != nil {
		return widget.Color{}, fmt.Errorf("bad hsl() value %q: %w", s, err)
	}
	h, err := parseHue(comps[0])
	if err != nil {
		return widget.Color{}, fmt.Errorf("bad hsl() hue in %q: %w", s, err)
	}
	sat, err := parseFraction(comps[1])
	if err != nil {
		return widget.Color{}, fmt.Errorf("bad hsl() saturation in %q: %w", s, err)
	}
	lig, err := parseFraction(comps[2])
	if err != nil {
		return widget.Color{}, fmt.Errorf("bad hsl() lightness in %q: %w", s, err)
	}
	r, g, b := hslToRGB(h, clamp01(sat), clamp01(lig))
	return widget.Color{
		R: float32(r),
		G: float32(g),
		B: float32(b),
		A: float32(alpha),
	}, nil
}

// parseFraction parses "95.9%" as 0.959; a bare number is taken as an
// already-normalized fraction.
func parseFraction(s string) (float64, error) {
	if p, ok := strings.CutSuffix(s, "%"); ok {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return 0, err
		}
		return v / 100, nil
	}
	return strconv.ParseFloat(s, 64)
}

// hslToRGB converts HSL (h in degrees, s and l in [0,1]) to sRGB in [0,1]
// using the CSS Color specification algorithm.
func hslToRGB(h, s, l float64) (r, g, b float64) {
	h = math.Mod(math.Mod(h, 360)+360, 360) / 360
	if s == 0 {
		return l, l, l
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	return hueToRGB(p, q, h+1.0/3), hueToRGB(p, q, h), hueToRGB(p, q, h-1.0/3)
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	switch {
	case t < 1.0/6:
		return p + (q-p)*6*t
	case t < 1.0/2:
		return q
	case t < 2.0/3:
		return p + (q-p)*(2.0/3-t)*6
	default:
		return p
	}
}

// parseRGBColor parses rgb()/rgba() in modern or legacy syntax. Channels
// are 0-255 numbers or percentages.
func parseRGBColor(s string) (widget.Color, error) {
	body, err := parenBody(s)
	if err != nil {
		return widget.Color{}, err
	}
	comps, alpha, err := splitColorComponents(body)
	if err != nil {
		return widget.Color{}, fmt.Errorf("bad rgb() value %q: %w", s, err)
	}
	var ch [3]float64
	for i, comp := range comps {
		if p, ok := strings.CutSuffix(comp, "%"); ok {
			v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
			if err != nil {
				return widget.Color{}, fmt.Errorf("bad rgb() channel in %q: %w", s, err)
			}
			ch[i] = clamp01(v / 100)
			continue
		}
		v, err := strconv.ParseFloat(comp, 64)
		if err != nil {
			return widget.Color{}, fmt.Errorf("bad rgb() channel in %q: %w", s, err)
		}
		ch[i] = clamp01(v / 255)
	}
	return widget.Color{
		R: float32(ch[0]),
		G: float32(ch[1]),
		B: float32(ch[2]),
		A: float32(alpha),
	}, nil
}

// parseHexColor parses #rgb, #rgba, #rrggbb, and #rrggbbaa.
func parseHexColor(s string) (widget.Color, error) {
	hexDigits := strings.TrimPrefix(s, "#")
	var pairs [4]string
	switch len(hexDigits) {
	case 3, 4:
		for i := 0; i < len(hexDigits); i++ {
			pairs[i] = string(hexDigits[i]) + string(hexDigits[i])
		}
	case 6, 8:
		for i := 0; i*2 < len(hexDigits); i++ {
			pairs[i] = hexDigits[i*2 : i*2+2]
		}
	default:
		return widget.Color{}, fmt.Errorf("bad hex color %q", s)
	}
	if pairs[3] == "" {
		pairs[3] = "ff"
	}
	var ch [4]float32
	for i, p := range pairs {
		v, err := strconv.ParseUint(p, 16, 16)
		if err != nil {
			return widget.Color{}, fmt.Errorf("bad hex color %q: %w", s, err)
		}
		ch[i] = float32(v) / 255
	}
	return widget.Color{R: ch[0], G: ch[1], B: ch[2], A: ch[3]}, nil
}

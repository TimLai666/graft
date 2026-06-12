package theme

import (
	"embed"
	"fmt"
	"sync"

	"github.com/gogpu/ui/widget"
)

// Base identifies one of the five shadcn base-color presets — the
// components.json "baseColor" choice. Pass it to New via BaseColor.
type Base uint8

const (
	// Neutral is the default shadcn base color (pure gray).
	Neutral Base = iota

	// Stone is the warm gray base color.
	Stone

	// Zinc is the cool gray base color.
	Zinc

	// Gray is the blue-tinted gray base color.
	Gray

	// Slate is the strongly blue-tinted gray base color.
	Slate

	baseCount // number of presets; keep last
)

// String returns the shadcn registry name of the base color.
func (b Base) String() string {
	switch b {
	case Neutral:
		return "neutral"
	case Stone:
		return "stone"
	case Zinc:
		return "zinc"
	case Gray:
		return "gray"
	case Slate:
		return "slate"
	default:
		return "unknown"
	}
}

// presetsFS embeds the verbatim :root/.dark CSS blocks copied from the
// shadcn distribution (see the header comment in each file for the exact
// source URL). Storing the presets as CSS and parsing them through
// ApplyCSS removes transcription risk and exercises the parser on every
// startup (DESIGN.md §2.4).
//
//go:embed presets/*.css
var presetsFS embed.FS

// presetData is the parsed form of one embedded preset.
type presetData struct {
	light, dark Tokens
	radius      float32
}

var (
	presetOnce  [baseCount]sync.Once
	presetCache [baseCount]presetData
)

// preset returns the parsed token sets for a base color, parsing the
// embedded CSS lazily exactly once per base. It panics on an unknown Base
// or a malformed embedded file (both are programmer errors: the files are
// fixed at compile time).
func preset(b Base) presetData {
	if b >= baseCount {
		panic(fmt.Sprintf("theme: unknown base color %d", b))
	}
	presetOnce[b].Do(func() {
		name := "presets/" + b.String() + ".css"
		css, err := presetsFS.ReadFile(name)
		if err != nil {
			panic(fmt.Sprintf("theme: missing embedded preset %s: %v", name, err))
		}
		t := newBaseTheme()
		if err := t.ApplyCSS(string(css)); err != nil {
			panic(fmt.Sprintf("theme: malformed embedded preset %s: %v", name, err))
		}
		presetCache[b] = presetData{light: t.Light, dark: t.Dark, radius: t.Radius}
	})
	return presetCache[b]
}

// newBaseTheme returns the blank theme preset CSS is applied onto. It
// carries the defaults the canonical shadcn CSS does not spell out:
// DestructiveForeground is literal white in both modes (shadcn uses
// text-white on destructive surfaces; DESIGN.md §8.16) and Radius is 10px
// in case a preset ever omits --radius.
func newBaseTheme() *Theme {
	white := widget.RGB(1, 1, 1)
	return &Theme{
		Light:  Tokens{DestructiveForeground: white},
		Dark:   Tokens{DestructiveForeground: white},
		Radius: defaultRadius,
	}
}

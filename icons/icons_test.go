package icons

import (
	stdcolor "image/color"
	"strings"
	"testing"

	ggsvg "github.com/gogpu/gg/svg"
	"github.com/gogpu/ui/icon"
)

// iconCount is the number of vendored Lucide icons.
const iconCount = 21

func TestAllReturnsEveryIcon(t *testing.T) {
	icons := All()
	if len(icons) != iconCount {
		t.Fatalf("All() returned %d icons, want %d", len(icons), iconCount)
	}
	seen := make(map[string]bool, len(icons))
	for _, ic := range icons {
		if !strings.HasPrefix(ic.Name, RegistryPrefix) {
			t.Errorf("icon %q: name missing %q prefix", ic.Name, RegistryPrefix)
		}
		if seen[ic.Name] {
			t.Errorf("duplicate icon name %q", ic.Name)
		}
		seen[ic.Name] = true
		if len(ic.SVGXML) == 0 {
			t.Errorf("icon %q: empty SVGXML", ic.Name)
		}
		if ic.ViewBox != lucideViewBox {
			t.Errorf("icon %q: ViewBox = %v, want %v", ic.Name, ic.ViewBox, float32(lucideViewBox))
		}
	}
}

func TestEveryEmbeddedSVGParses(t *testing.T) {
	for _, ic := range All() {
		doc, err := ggsvg.Parse(ic.SVGXML)
		if err != nil {
			t.Errorf("icon %q: SVG parse failed: %v", ic.Name, err)
			continue
		}
		if doc.ViewBox.Width != lucideViewBox || doc.ViewBox.Height != lucideViewBox {
			t.Errorf("icon %q: parsed viewBox = %gx%g, want %dx%d",
				ic.Name, doc.ViewBox.Width, doc.ViewBox.Height, lucideViewBox, lucideViewBox)
		}
		if len(doc.Elements) == 0 {
			t.Errorf("icon %q: parsed document has no elements", ic.Name)
		}
	}
}

func TestNormalizePushesRootAttrsOntoShapes(t *testing.T) {
	raw, err := lucideFS.ReadFile("lucide/check.svg")
	if err != nil {
		t.Fatalf("reading embedded check.svg: %v", err)
	}
	normalized, err := normalizeLucideSVG(raw)
	if err != nil {
		t.Fatalf("normalizeLucideSVG: %v", err)
	}
	got := string(normalized)

	for _, want := range []string{
		`stroke="currentColor"`,
		`stroke-width="2"`,
		`stroke-linecap="round"`,
		`stroke-linejoin="round"`,
		`fill="none"`,
	} {
		// Root + the single path element of check.svg.
		if n := strings.Count(got, want); n != 2 {
			t.Errorf("normalized check.svg: %s occurs %d times, want 2\noutput: %s", want, n, got)
		}
	}
	if !strings.Contains(got, `d="M20 6 9 17l-5-5"`) {
		t.Errorf("normalized check.svg lost path data:\n%s", got)
	}
}

func TestNormalizeKeepsExplicitShapeAttrs(t *testing.T) {
	in := []byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">` +
		`<path stroke="red" d="M0 0L1 1"/></svg>`)
	out, err := normalizeLucideSVG(in)
	if err != nil {
		t.Fatalf("normalizeLucideSVG: %v", err)
	}
	got := string(out)
	if strings.Count(got, `stroke="red"`) != 1 {
		t.Errorf("explicit stroke lost or duplicated:\n%s", got)
	}
	// The path already declares stroke, so only the root keeps currentColor.
	if n := strings.Count(got, `stroke="currentColor"`); n != 1 {
		t.Errorf("inherited stroke pushed despite explicit attr (%d occurrences):\n%s", n, got)
	}
	// stroke-width is not declared on the path, so it is inherited.
	if n := strings.Count(got, `stroke-width="2"`); n != 2 {
		t.Errorf("stroke-width occurs %d times, want 2:\n%s", n, got)
	}
}

func TestNormalizeRejectsNonSVG(t *testing.T) {
	if _, err := normalizeLucideSVG([]byte(`<html></html>`)); err == nil {
		t.Error("normalizeLucideSVG accepted input without an <svg> root")
	}
	if _, err := normalizeLucideSVG([]byte(`<svg><path`)); err == nil {
		t.Error("normalizeLucideSVG accepted malformed XML")
	}
}

func TestRegisterIdempotent(t *testing.T) {
	if err := Register(); err != nil {
		t.Fatalf("Register() = %v, want nil", err)
	}
	reg := icon.DefaultRegistry()
	lenAfterFirst := reg.Len()

	if err := Register(); err != nil {
		t.Fatalf("second Register() = %v, want nil", err)
	}
	if got := reg.Len(); got != lenAfterFirst {
		t.Errorf("registry length changed on second Register(): %d -> %d", lenAfterFirst, got)
	}

	for _, ic := range All() {
		got, ok := reg.Get(ic.Name)
		if !ok {
			t.Errorf("registry missing %q after Register()", ic.Name)
			continue
		}
		if len(got.SVGXML) == 0 {
			t.Errorf("registered icon %q has empty SVGXML", ic.Name)
		}
	}
}

func TestRenderSpotCheck(t *testing.T) {
	// Offline CPU rasterization through the same gg/svg pipeline the
	// gogpu/ui canvases use for widget.SVGRenderer.
	const size = 24
	img, err := ggsvg.RenderWithColor(Check.SVGXML, size, size,
		stdcolor.NRGBA{R: 0, G: 0, B: 0, A: 255})
	if err != nil {
		t.Fatalf("RenderWithColor: %v", err)
	}

	drawn := 0
	for y := range size {
		for x := range size {
			if _, _, _, a := img.At(x, y).RGBA(); a > 0 {
				drawn++
			}
		}
	}
	if drawn == 0 {
		t.Fatal("rendered check icon has no visible pixels")
	}
	// A stroked checkmark must not cover most of the canvas. If the
	// normalizer regressed (shapes filled instead of stroked), coverage
	// would jump far above this bound.
	if drawn > size*size/2 {
		t.Errorf("rendered check icon covers %d of %d pixels; expected sparse stroke coverage",
			drawn, size*size)
	}
}

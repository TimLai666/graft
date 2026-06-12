// Package textmetrics provides sfnt-based text measurement for layout-time
// widths (DESIGN.md section 5.6).
//
// gogpu/ui's built-in heuristic (rune count x fontSize x 0.55) is too crude
// for pixel-faithful widths, so graft measures real glyph advances from the
// embedded Geist TTFs via golang.org/x/image/font/sfnt. Fonts register once
// per family ([Register]); [Width] and [LineHeight] then resolve metrics for
// any registered family. Measured widths are memoized in a bounded LRU cache
// because measurement runs on every layout pass.
//
// All functions are safe for concurrent use.
package textmetrics

import (
	"fmt"
	"math"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// fallbackAdvanceFactor is the per-rune advance estimate (in em) used when a
// family is not registered or a glyph cannot be measured. It matches the
// gogpu/ui heuristic (len(text) * fontSize * 0.55) so unmeasured text
// degrades to the same widths the core widgets assume.
const fallbackAdvanceFactor = 0.55

// widthCacheCapacity bounds the number of memoized (family, size, text)
// width entries.
const widthCacheCapacity = 4096

// fontEntry holds one parsed font and a pool of reusable sfnt buffers.
type fontEntry struct {
	font *sfnt.Font
	bufs sync.Pool // of *sfnt.Buffer
}

// getBuffer takes a buffer from the pool.
func (e *fontEntry) getBuffer() *sfnt.Buffer {
	return e.bufs.Get().(*sfnt.Buffer)
}

// putBuffer returns a buffer to the pool.
func (e *fontEntry) putBuffer(b *sfnt.Buffer) {
	e.bufs.Put(b)
}

var (
	// registryMu guards registry.
	registryMu sync.RWMutex

	// registry maps a family name to its parsed font.
	registry = map[string]*fontEntry{}

	// widths memoizes measured text widths.
	widths = newWidthCache(widthCacheCapacity)
)

// Register parses ttf and makes it available for measurement under the given
// family name. The font is parsed once; subsequent measurement calls reuse
// the parsed tables through a buffer pool.
//
// Registering an already-registered family replaces it and invalidates the
// width cache.
func Register(family string, ttf []byte) error {
	if family == "" {
		return fmt.Errorf("textmetrics: family name must not be empty")
	}
	f, err := sfnt.Parse(ttf)
	if err != nil {
		return fmt.Errorf("textmetrics: parse font %q: %w", family, err)
	}
	e := &fontEntry{font: f}
	e.bufs.New = func() any { return &sfnt.Buffer{} }

	registryMu.Lock()
	registry[family] = e
	registryMu.Unlock()

	// Replacing a family could change cached widths; drop everything.
	widths.clear()
	return nil
}

// Registered reports whether a family has been registered. Callers that need
// to detect the estimation fallback (e.g. to log it) can check this before
// calling Width.
func Registered(family string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	_, ok := registry[family]
	return ok
}

// Width returns the width in px of text rendered in the given family at
// sizePx: the sum of glyph advances plus kerning adjustments.
//
// Unknown runes measure with the font's .notdef advance. If the family is
// not registered, Width falls back to the gogpu/ui heuristic
// (rune count * sizePx * 0.55); use [Registered] to detect that case.
func Width(family string, sizePx float32, text string) float32 {
	if text == "" || sizePx <= 0 {
		return 0
	}
	registryMu.RLock()
	e := registry[family]
	registryMu.RUnlock()
	if e == nil {
		return estimateWidth(sizePx, text)
	}

	key := widthKey{family: family, size: sizePx, text: text}
	if w, ok := widths.get(key); ok {
		return w
	}
	w := measureWidth(e, sizePx, text)
	widths.put(key, w)
	return w
}

// LineHeight returns the vertical metrics in px for the given family at
// sizePx: ascent and descent are both positive distances from the baseline,
// lineGap is the extra leading between lines.
//
// If the family is not registered, it returns the common estimate
// (0.8 * sizePx ascent, 0.2 * sizePx descent, zero gap).
func LineHeight(family string, sizePx float32) (ascent, descent, lineGap float32) {
	if sizePx <= 0 {
		return 0, 0, 0
	}
	registryMu.RLock()
	e := registry[family]
	registryMu.RUnlock()
	if e == nil {
		return sizePx * 0.8, sizePx * 0.2, 0
	}

	b := e.getBuffer()
	defer e.putBuffer(b)

	m, err := e.font.Metrics(b, ppem(sizePx), font.HintingNone)
	if err != nil {
		return sizePx * 0.8, sizePx * 0.2, 0
	}
	ascent = fixedToPx(m.Ascent)
	descent = fixedToPx(m.Descent)
	lineGap = fixedToPx(m.Height) - ascent - descent
	if lineGap < 0 {
		lineGap = 0
	}
	return ascent, descent, lineGap
}

// measureWidth sums glyph advances and kerning for text at sizePx.
func measureWidth(e *fontEntry, sizePx float32, text string) float32 {
	b := e.getBuffer()
	defer e.putBuffer(b)

	p := ppem(sizePx)
	perRuneFallback := fixed.Int26_6(math.Round(float64(sizePx) * fallbackAdvanceFactor * 64))

	var total fixed.Int26_6
	var prev sfnt.GlyphIndex
	hasPrev := false
	for _, r := range text {
		gi, err := e.font.GlyphIndex(b, r)
		if err != nil {
			// Corrupt cmap lookup; estimate this rune and break kerning.
			total += perRuneFallback
			hasPrev = false
			continue
		}
		// A missing rune yields glyph index 0 (.notdef) with no error, so
		// its advance below is the .notdef advance, as intended.
		adv, err := e.font.GlyphAdvance(b, gi, p, font.HintingNone)
		if err != nil {
			total += perRuneFallback
			hasPrev = false
			continue
		}
		if hasPrev {
			// Kern returns ErrNotFound for unkerned pairs in some fonts;
			// treat any error as a zero adjustment.
			if k, err := e.font.Kern(b, prev, gi, p, font.HintingNone); err == nil {
				total += k
			}
		}
		total += adv
		prev = gi
		hasPrev = true
	}
	return fixedToPx(total)
}

// estimateWidth is the unregistered-family fallback, matching gogpu/ui's
// heuristic.
func estimateWidth(sizePx float32, text string) float32 {
	return float32(len([]rune(text))) * sizePx * fallbackAdvanceFactor
}

// ppem converts a pixel size to the 26.6 fixed-point ppem sfnt expects.
func ppem(sizePx float32) fixed.Int26_6 {
	return fixed.Int26_6(math.Round(float64(sizePx) * 64))
}

// fixedToPx converts a 26.6 fixed-point value to float32 px.
func fixedToPx(v fixed.Int26_6) float32 {
	return float32(v) / 64
}

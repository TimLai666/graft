package textmetrics

import (
	"os"
	"sync"
	"testing"
)

// geistOnce loads and registers the repo's Geist Regular TTF exactly once.
var geistOnce sync.Once

// registerGeist registers fonts/Geist-Regular.ttf under the family "Geist".
// The relative path is test-only: go test runs with the package directory
// as the working directory.
func registerGeist(t *testing.T) {
	t.Helper()
	geistOnce.Do(func() {
		ttf, err := os.ReadFile("../../fonts/Geist-Regular.ttf")
		if err != nil {
			t.Fatalf("read Geist TTF: %v", err)
		}
		if err := Register("Geist", ttf); err != nil {
			t.Fatalf("Register: %v", err)
		}
	})
	if !Registered("Geist") {
		t.Fatal("Geist not registered (earlier registration failed)")
	}
}

func TestRegisterRejectsBadInput(t *testing.T) {
	if err := Register("", []byte{0, 1, 2}); err == nil {
		t.Error("Register with empty family: got nil error, want error")
	}
	if err := Register("Broken", []byte("not a font")); err == nil {
		t.Error("Register with garbage bytes: got nil error, want error")
	}
	if Registered("Broken") {
		t.Error("failed Register must not register the family")
	}
}

func TestWidthHelloSaneBand(t *testing.T) {
	registerGeist(t)
	w := Width("Geist", 14, "Hello")
	if w <= 20 || w >= 60 {
		t.Errorf("Width(Geist, 14, %q) = %v, want in (20, 60)", "Hello", w)
	}
}

func TestWidthEmptyTextIsZero(t *testing.T) {
	registerGeist(t)
	if w := Width("Geist", 14, ""); w != 0 {
		t.Errorf("Width of empty text = %v, want 0", w)
	}
}

func TestWidthZeroSizeIsZero(t *testing.T) {
	registerGeist(t)
	if w := Width("Geist", 0, "Hello"); w != 0 {
		t.Errorf("Width at size 0 = %v, want 0", w)
	}
}

func TestWidthMonotonicInSize(t *testing.T) {
	registerGeist(t)
	sizes := []float32{10, 12, 14, 18, 24, 36}
	var prev float32
	for i, s := range sizes {
		w := Width("Geist", s, "The quick brown fox")
		if w <= 0 {
			t.Fatalf("Width at size %v = %v, want > 0", s, w)
		}
		if i > 0 && w <= prev {
			t.Errorf("Width at size %v = %v, not greater than width %v at size %v",
				s, w, prev, sizes[i-1])
		}
		prev = w
	}
}

func TestWidthMonotonicInText(t *testing.T) {
	registerGeist(t)
	short := Width("Geist", 14, "Hello")
	long := Width("Geist", 14, "Hello, world")
	if long <= short {
		t.Errorf("longer text width %v not greater than shorter %v", long, short)
	}
}

func TestWidthUnknownRuneUsesNotdef(t *testing.T) {
	registerGeist(t)
	// U+FFFF is unmapped in Geist; it must still contribute a (notdef)
	// advance rather than panicking or measuring zero-width text as zero.
	w := Width("Geist", 14, "a"+string(rune(0xFFFF))+"b")
	ab := Width("Geist", 14, "ab")
	if w < ab {
		t.Errorf("width with unknown rune = %v, want >= width of %q = %v", w, "ab", ab)
	}
}

func TestWidthUnregisteredFamilyFallsBack(t *testing.T) {
	const family = "NoSuchFamily"
	if Registered(family) {
		t.Fatalf("%q unexpectedly registered", family)
	}
	got := Width(family, 10, "ab")
	want := float32(2) * 10 * fallbackAdvanceFactor
	if got != want {
		t.Errorf("fallback width = %v, want %v", got, want)
	}
}

func TestWidthCacheHitReturnsSameValue(t *testing.T) {
	registerGeist(t)
	first := Width("Geist", 14, "cached text")
	second := Width("Geist", 14, "cached text")
	if first != second {
		t.Errorf("cache hit returned %v, first measurement was %v", second, first)
	}
}

func TestWidthConcurrentAccess(t *testing.T) {
	registerGeist(t)
	want := Width("Geist", 14, "concurrent")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				if w := Width("Geist", 14, "concurrent"); w != want {
					t.Errorf("concurrent Width = %v, want %v", w, want)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestLineHeight(t *testing.T) {
	registerGeist(t)
	ascent, descent, lineGap := LineHeight("Geist", 14)
	if ascent <= 0 {
		t.Errorf("ascent = %v, want > 0", ascent)
	}
	if descent < 0 {
		t.Errorf("descent = %v, want >= 0", descent)
	}
	if lineGap < 0 {
		t.Errorf("lineGap = %v, want >= 0", lineGap)
	}
	if ascent <= descent {
		t.Errorf("ascent %v not greater than descent %v", ascent, descent)
	}
	if total := ascent + descent + lineGap; total < 14 || total > 28 {
		t.Errorf("total line height = %v, want within [14, 28] for size 14", total)
	}
}

func TestLineHeightScalesWithSize(t *testing.T) {
	registerGeist(t)
	a14, d14, _ := LineHeight("Geist", 14)
	a28, d28, _ := LineHeight("Geist", 28)
	if a28 <= a14 {
		t.Errorf("ascent at 28px (%v) not greater than at 14px (%v)", a28, a14)
	}
	if d28 < d14 {
		t.Errorf("descent at 28px (%v) smaller than at 14px (%v)", d28, d14)
	}
}

func TestLineHeightUnregisteredFamilyFallsBack(t *testing.T) {
	ascent, descent, lineGap := LineHeight("NoSuchFamily", 10)
	if ascent != 8 || descent != 2 || lineGap != 0 {
		t.Errorf("fallback LineHeight = (%v, %v, %v), want (8, 2, 0)", ascent, descent, lineGap)
	}
}

func TestLineHeightZeroSizeIsZero(t *testing.T) {
	registerGeist(t)
	ascent, descent, lineGap := LineHeight("Geist", 0)
	if ascent != 0 || descent != 0 || lineGap != 0 {
		t.Errorf("LineHeight at size 0 = (%v, %v, %v), want zeros", ascent, descent, lineGap)
	}
}

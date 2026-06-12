// Package gtest is graft's golden-image test harness: it renders widget
// trees headlessly through gogpu/ui's offscreen renderer (pure CPU, no
// window) and compares the result pixel-for-pixel against checked-in
// golden PNGs.
//
// Regenerate goldens with:
//
//	$env:GRAFT_UPDATE_GOLDEN = "1"; go test ./...
//
// Goldens live in testdata/golden/<name>.png relative to the calling test
// package. On mismatch the rendered image is written next to the golden as
// <name>.got.png for visual diffing.
package gtest

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/gogpu/ui/offscreen"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/theme"
)

// updateEnv is the environment variable that switches golden comparison to
// golden (re)generation.
const updateEnv = "GRAFT_UPDATE_GOLDEN"

// Scale is the render scale for goldens. 2x catches sub-pixel border and
// ring placement errors that 1x rounds away.
const Scale = 2

// Golden renders the widget produced by build and compares it against
// testdata/golden/<name>.png. The build function runs after the theme mode
// is applied, so constructors snapshot the right mode.
func Golden(t *testing.T, name string, dark bool, build func() widget.Widget) {
	t.Helper()

	if err := graft.LoadAssets(); err != nil {
		t.Fatalf("gtest: loading fonts/icons: %v", err)
	}

	th := graft.CurrentTheme()
	prev := th.Mode()
	if dark {
		th.SetMode(theme.ModeDark)
	} else {
		th.SetMode(theme.ModeLight)
	}
	defer th.SetMode(prev)

	r := offscreen.NewRenderer(0, 0,
		offscreen.WithFitSize(),
		offscreen.WithMaxSize(900, 4000),
		offscreen.WithTheme(th),
		offscreen.WithBackground(th.Active().Background),
		offscreen.WithScale(Scale),
	)
	r.Render(build())
	got := r.Image()
	if got == nil {
		t.Fatal("gtest: offscreen renderer produced no image")
	}

	goldenPath := filepath.Join("testdata", "golden", name+".png")

	if os.Getenv(updateEnv) != "" {
		writePNG(t, goldenPath, got)
		t.Logf("gtest: wrote golden %s (%dx%d)", goldenPath, got.Bounds().Dx(), got.Bounds().Dy())
		return
	}

	want, err := readPNG(goldenPath)
	if err != nil {
		gotPath := goldenPath + ".got.png"
		writePNG(t, gotPath, got)
		t.Fatalf("gtest: missing golden %s (rendered output saved to %s; set %s=1 to record)",
			goldenPath, gotPath, updateEnv)
	}

	if diff := diffImages(want, got); diff != "" {
		gotPath := filepath.Join("testdata", "golden", name+".got.png")
		writePNG(t, gotPath, got)
		t.Fatalf("gtest: %s differs from golden: %s (got saved to %s; set %s=1 to re-record)",
			name, diff, gotPath, updateEnv)
	}
}

// GoldenLightDark runs Golden twice, as <name>-light and <name>-dark.
func GoldenLightDark(t *testing.T, name string, build func() widget.Widget) {
	t.Helper()
	Golden(t, name+"-light", false, build)
	Golden(t, name+"-dark", true, build)
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("gtest: %v", err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("gtest: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("gtest: encoding %s: %v", path, err)
	}
}

func readPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

// diffImages compares two images pixel-exactly and returns a human-readable
// summary of the first difference plus the total differing-pixel count, or
// "" when identical.
func diffImages(want, got image.Image) string {
	wb, gb := want.Bounds(), got.Bounds()
	if wb.Dx() != gb.Dx() || wb.Dy() != gb.Dy() {
		return sizeDiff(wb, gb)
	}

	var count int
	firstX, firstY := -1, -1
	for y := 0; y < wb.Dy(); y++ {
		for x := 0; x < wb.Dx(); x++ {
			wr, wg, wbl, wa := want.At(wb.Min.X+x, wb.Min.Y+y).RGBA()
			gr, gg, gbl, ga := got.At(gb.Min.X+x, gb.Min.Y+y).RGBA()
			if wr != gr || wg != gg || wbl != gbl || wa != ga {
				if count == 0 {
					firstX, firstY = x, y
				}
				count++
			}
		}
	}
	if count == 0 {
		return ""
	}
	total := wb.Dx() * wb.Dy()
	return diffSummary(count, total, firstX, firstY)
}

func sizeDiff(wb, gb image.Rectangle) string {
	return fmt.Sprintf("size mismatch: golden %dx%d vs got %dx%d", wb.Dx(), wb.Dy(), gb.Dx(), gb.Dy())
}

func diffSummary(count, total, x, y int) string {
	return fmt.Sprintf("%d/%d pixels differ (first at %d,%d)", count, total, x, y)
}

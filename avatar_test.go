package graft_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
)

// solidImage returns a deterministic px×px solid-color image for reproducible
// avatar-image goldens and spec tests.
func solidImage(px int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, px, px))
	for y := 0; y < px; y++ {
		for x := 0; x < px; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// TestAvatarSpecFallback verifies the fallback draws a Muted circle (full
// radius) with centered MutedForeground initials at the default 14px size.
func TestAvatarSpecFallback(t *testing.T) {
	th := alertForceLight(t)
	tok := th.Active()

	a := graft.Avatar(graft.AvatarFallback("CN"))
	size := a.Layout(nil, fixedWidthLoose(32))
	if size.Width != 32 || size.Height != 32 {
		t.Errorf("avatar size = %vx%v, want 32x32 (default)", size.Width, size.Height)
	}

	canvas := uitest.DrawWidget(a)
	if len(canvas.RoundRects) != 1 {
		t.Fatalf("fallback drew %d round-rects, want 1 (Muted circle)", len(canvas.RoundRects))
	}
	circle := canvas.RoundRects[0]
	if circle.Color != tok.Muted {
		t.Errorf("fallback circle color = %v, want Muted %v", circle.Color, tok.Muted)
	}
	if circle.Radius != th.RadiusFull() {
		t.Errorf("fallback circle radius = %v, want RadiusFull %v", circle.Radius, th.RadiusFull())
	}
	if len(canvas.StyledTexts) != 1 {
		t.Fatalf("fallback drew %d text runs, want 1", len(canvas.StyledTexts))
	}
	txt := canvas.StyledTexts[0]
	if txt.Text != "CN" {
		t.Errorf("fallback text = %q, want CN", txt.Text)
	}
	if txt.Style.Color != tok.MutedForeground {
		t.Errorf("fallback text color = %v, want MutedForeground %v", txt.Style.Color, tok.MutedForeground)
	}
	if txt.Style.FontSize != 14 {
		t.Errorf("fallback text size = %v, want 14 (default)", txt.Style.FontSize)
	}
	if txt.Style.Align != widget.TextAlignCenter {
		t.Errorf("fallback text align = %v, want center", txt.Style.Align)
	}
}

// TestAvatarSpecFallbackSmall verifies the sm avatar (24px) uses 12px text.
func TestAvatarSpecFallbackSmall(t *testing.T) {
	alertForceLight(t)
	a := graft.Avatar(graft.AvatarFallback("AB")).Sm()
	size := a.Layout(nil, fixedWidthLoose(24))
	if size.Width != 24 {
		t.Errorf("sm avatar width = %v, want 24", size.Width)
	}
	canvas := uitest.DrawWidget(a)
	if len(canvas.StyledTexts) != 1 || canvas.StyledTexts[0].Style.FontSize != 12 {
		t.Errorf("sm avatar fallback font size = %v, want 12", canvas.StyledTexts)
	}
}

// TestAvatarSpecImage verifies the image path: a circular clip plus a
// DrawImage call (no fallback round-rect when only an image is given).
func TestAvatarSpecImage(t *testing.T) {
	alertForceLight(t)
	img := solidImage(64, color.RGBA{R: 0x6d, G: 0x28, B: 0xd9, A: 0xff})

	a := graft.Avatar(graft.AvatarImage(img))
	a.Layout(nil, fixedWidthLoose(32))
	canvas := uitest.DrawWidget(a)

	if len(canvas.RoundRects) != 0 {
		t.Errorf("image-only avatar drew %d round-rects, want 0", len(canvas.RoundRects))
	}
	if len(canvas.ClipRoundRects) != 1 {
		t.Fatalf("image avatar drew %d round-rect clips, want 1 (circular clip)", len(canvas.ClipRoundRects))
	}
	if len(canvas.Images) != 1 {
		t.Fatalf("image avatar drew %d images, want 1", len(canvas.Images))
	}
	// Image is resampled to the 32px diameter.
	if b := canvas.Images[0].Image.Bounds(); b.Dx() != 32 || b.Dy() != 32 {
		t.Errorf("scaled image size = %dx%d, want 32x32", b.Dx(), b.Dy())
	}
}

// TestAvatarSpecImageOverFallback verifies that giving both shows the fallback
// circle behind and the image clipped on top.
func TestAvatarSpecImageOverFallback(t *testing.T) {
	alertForceLight(t)
	img := solidImage(64, color.RGBA{R: 0x10, G: 0xb9, B: 0x81, A: 0xff})

	a := graft.Avatar(graft.AvatarImage(img), graft.AvatarFallback("CN"))
	a.Layout(nil, fixedWidthLoose(32))
	canvas := uitest.DrawWidget(a)

	if len(canvas.RoundRects) != 1 {
		t.Errorf("expected 1 fallback circle behind the image, got %d", len(canvas.RoundRects))
	}
	if len(canvas.Images) != 1 {
		t.Errorf("expected the image drawn over the fallback, got %d images", len(canvas.Images))
	}
}

// TestGoldenAvatar renders fallback initials at the three sizes plus an image
// avatar, in light and dark modes, with a deterministic generated image.
func TestGoldenAvatar(t *testing.T) {
	gtest.GoldenLightDark(t, "avatar-gallery", func() widget.Widget {
		img := solidImage(80, color.RGBA{R: 0x6d, G: 0x28, B: 0xd9, A: 0xff})
		return primitives.HBox(
			graft.Avatar(graft.AvatarFallback("CN")).Sm(),
			graft.Avatar(graft.AvatarFallback("CN")),
			graft.Avatar(graft.AvatarFallback("CN")).Lg(),
			graft.Avatar(graft.AvatarImage(img)).Lg(),
		).Gap(16).Padding(24).CrossAlign(primitives.CrossAxisCenter)
	})
}

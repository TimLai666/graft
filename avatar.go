package graft

import (
	"image"

	xdraw "golang.org/x/image/draw"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// AvatarWidget is graft's circular avatar: a rounded-full box at one of three
// sizes (sm 24 / default 32 / lg 40) showing either an image clipped to the
// circle, fallback initials on a Muted circle, or the image over the fallback
// (so the initials show while the image is unset), matching the shadcn Avatar
// (docs/research/03-shadcn-pixel-spec.md §5).
//
// It is a graft-OWNED widget: the circular image clip (PushClipRoundRect with
// the full pill radius) and the live-resolved Muted/MutedForeground fallback
// colors are simplest to draw directly.
//
//	graft.Avatar(graft.AvatarImage(img))
//	graft.Avatar(graft.AvatarFallback("CN")).Sm()
type AvatarWidget struct {
	widget.WidgetBase

	size float32

	img      image.Image
	fallback string

	// scaled caches the source image resampled to the current diameter so
	// DrawImage (which draws at native pixel size) fills the circle exactly.
	scaled   image.Image
	scaledPx int

	theme *theme.Theme
}

// avatarImage and avatarFallback tag an avatar's content so the Avatar
// constructor can route children without exporting widget types.
type avatarImage struct{ img image.Image }

func (avatarImage) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(0, 0))
}
func (avatarImage) Draw(widget.Context, widget.Canvas)     {}
func (avatarImage) Event(widget.Context, event.Event) bool { return false }
func (avatarImage) Children() []widget.Widget              { return nil }

type avatarFallback struct{ text string }

func (avatarFallback) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(0, 0))
}
func (avatarFallback) Draw(widget.Context, widget.Canvas)     {}
func (avatarFallback) Event(widget.Context, event.Event) bool { return false }
func (avatarFallback) Children() []widget.Widget              { return nil }

// Avatar builds an avatar from AvatarImage and/or AvatarFallback children. The
// default diameter is 32px (use Sm/Lg to change). If both an image and a
// fallback are given, the image draws over the fallback.
func Avatar(children ...widget.Widget) *AvatarWidget {
	a := &AvatarWidget{size: metrics.Avatar.SizeDefault, theme: CurrentTheme()}
	a.SetVisible(true)
	a.SetEnabled(true)
	for _, c := range children {
		switch w := c.(type) {
		case avatarImage:
			a.img = w.img
		case avatarFallback:
			a.fallback = w.text
		}
	}
	return a
}

// AvatarImage supplies the avatar image, clipped to the circle.
func AvatarImage(img image.Image) avatarImage { return avatarImage{img: img} }

// AvatarFallback supplies the fallback initials shown on a Muted circle (when
// there is no image, or behind one).
func AvatarFallback(text string) avatarFallback { return avatarFallback{text: text} }

// Sm sets the small size (24px).
func (a *AvatarWidget) Sm() *AvatarWidget {
	a.size = metrics.Avatar.SizeSM
	return a
}

// Lg sets the large size (40px).
func (a *AvatarWidget) Lg() *AvatarWidget {
	a.size = metrics.Avatar.SizeLG
	return a
}

// Theme pins a specific theme instead of the process-wide current theme.
func (a *AvatarWidget) Theme(th *theme.Theme) *AvatarWidget {
	a.theme = th
	return a
}

func (a *AvatarWidget) resolvedTheme() *theme.Theme {
	if a.theme != nil {
		return a.theme
	}
	return CurrentTheme()
}

// fallbackFontSize returns the initials font size for the current diameter
// (12px at sm, 14px otherwise).
func (a *AvatarWidget) fallbackFontSize() float32 {
	if a.size <= metrics.Avatar.SizeSM {
		return metrics.Avatar.FallbackFontSizeSM
	}
	return metrics.Avatar.FallbackFontSize
}

// Layout sizes the avatar to a square of its diameter.
func (a *AvatarWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	size := c.Constrain(geometry.Sz(a.size, a.size))
	a.SetBounds(geometry.FromPointSize(a.Position(), size))
	return size
}

// Draw paints the fallback circle + initials (if any), then the image clipped
// to the circle on top (if any).
func (a *AvatarWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !a.IsVisible() {
		return
	}
	th := a.resolvedTheme()
	tok := th.Active()
	bounds := a.Bounds()
	radius := th.RadiusFull()

	// Fallback: Muted circle with centered initials.
	if a.fallback != "" {
		canvas.DrawRoundRect(bounds, tok.Muted, radius)
		fs := a.fallbackFontSize()
		family := avatarFontFamily(th, metrics.Avatar.FallbackWeight)
		if std, ok := canvas.(widget.StyledTextDrawer); ok {
			std.DrawStyledText(a.fallback, bounds, widget.TextStyle{
				FontFamily: family,
				FontSize:   fs,
				Color:      tok.MutedForeground,
				Align:      widget.TextAlignCenter,
			})
		} else {
			canvas.DrawText(a.fallback, bounds, fs, tok.MutedForeground, metrics.Avatar.FallbackWeight >= 600, widget.TextAlignCenter)
		}
	}

	// Image: scaled to the diameter, clipped to the circle.
	if a.img != nil {
		px := int(bounds.Width() + 0.5)
		if px > 0 {
			if a.scaled == nil || a.scaledPx != px {
				a.scaled = scaleToSquare(a.img, px)
				a.scaledPx = px
			}
			canvas.PushClipRoundRect(bounds, radius)
			canvas.DrawImage(a.scaled, bounds.Min)
			canvas.PopClip()
		}
	}
}

// Event ignores all input; the avatar is a display element.
func (a *AvatarWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; the avatar owns its rendering directly.
func (a *AvatarWidget) Children() []widget.Widget { return nil }

// avatarFontFamily resolves the registered Geist family for the weight,
// honoring a custom theme sans family (mirrors TypographyWidget.family).
func avatarFontFamily(th *theme.Theme, weight int) string {
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(weight)
}

// scaleToSquare resamples src to a px×px image with high-quality filtering so
// the avatar image fills its circular box (DrawImage draws at native size).
func scaleToSquare(src image.Image, px int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, px, px))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

var (
	_ widget.Widget = (*AvatarWidget)(nil)
	_ widget.Widget = avatarImage{}
	_ widget.Widget = avatarFallback{}
)

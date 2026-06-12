package draw

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
)

// Shadow approximates a CSS box-shadow by stacking low-alpha black
// round-rects under the element (DESIGN.md section 5.5). Call it before
// painting the element fill.
//
// Each layer draws bounds expanded by Grow, translated down by DY, with the
// corner radius enlarged by Grow so the layer stays concentric. The layer
// tables (metrics.ShadowXS/SM/MD/LG) live in the metrics package.
func Shadow(c widget.Canvas, bounds geometry.Rect, radius float32, layers []metrics.ShadowLayer) {
	for _, l := range layers {
		c.DrawRoundRect(
			bounds.Expand(l.Grow).TranslateXY(0, l.DY),
			widget.RGBA(0, 0, 0, l.Alpha),
			radius+l.Grow,
		)
	}
}

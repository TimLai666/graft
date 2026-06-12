// Package draw provides the shared low-level painting helpers used by every
// graft painter and widget (DESIGN.md section 5).
//
// The helpers fall into two groups:
//
//   - Pure color math: [Alpha], [MulAlpha], [Fade], [Over]. These implement
//     the CSS alpha-compositing conventions shadcn relies on (bg-primary/90,
//     disabled:opacity-50, ...) on top of widget.Color.
//
//   - Canvas helpers: [FocusRing], [FocusRingWidth], [OffsetRing],
//     [InsideBorder], [SquareCorners], [Shadow]. These translate shadcn CSS
//     box-model conventions (outside box-shadow rings, inside borders,
//     per-corner radii, blurred shadows) into the primitives the
//     widget.Canvas interface actually offers. Strokes on the canvas are
//     center-drawn, so each helper offsets bounds and radius accordingly.
//
// All pixel constants used here route through the metrics package; the
// helpers themselves are geometry-generic.
package draw

// Package metrics holds the exact pixel constants transcribed from the
// shadcn/ui new-york-v4 registry (docs/research/03-shadcn-pixel-spec.md).
//
// Every pixel value drawn by graft painters and composites comes from this
// package, never from inline literals, so the whole library can be audited
// against the shadcn CSS in one place.
//
// The package is pure data: it imports nothing (stdlib included) and
// contains no logic. Its values are pinned by trivial unit tests here and
// exercised for real by the painter spec tests (DESIGN.md section 6.2).
package metrics

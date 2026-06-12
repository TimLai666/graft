// Package icons embeds the subset of Lucide icons used by graft widgets
// and exposes them as gogpu/ui icon.IconData values.
//
// Each icon is embedded as the original 24x24 stroke-based Lucide SVG
// and normalized at package init: Lucide places its presentation
// attributes (fill="none", stroke="currentColor", stroke-width="2",
// stroke-linecap="round", stroke-linejoin="round") on the root <svg>
// element, but the gogpu/gg SVG renderer only inherits the root fill
// attribute, so the normalizer copies the root presentation attributes
// onto every shape element. The result is rendered through
// widget.SVGRenderer (implemented by both production canvases) with the
// stroke color overridden by the requested icon color, exactly matching
// Lucide's currentColor behavior.
//
// Icons are available as package-level variables (for direct use with
// icon.Draw or icon.IconWidget) and, after calling [Register], from
// icon.DefaultRegistry() under "lucide:<name>" keys, e.g.
// "lucide:chevron-down".
//
// # Vendoring more icons
//
// Additional Lucide icons are vendored with the gen tool, which
// downloads pristine SVGs from the Lucide repository into icons/lucide:
//
//	go run ./icons/gen check x chevron-down
//
// After downloading, add a matching exported variable to icons.go and
// list it in all().
//
// # License
//
// The Lucide icons are licensed under the ISC License, copyright (c)
// for portions of Lucide are held by Cole Bemis 2013-2022 as part of
// Feather (MIT), and copyright (c) 2022 Lucide Contributors for all
// other portions. The full license text is embedded in this directory
// as lucide/LICENSE.
package icons

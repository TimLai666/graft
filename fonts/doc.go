// Package fonts embeds the Geist and Geist Mono typefaces and registers
// them with the gogpu/ui text rendering pipeline.
//
// graft renders all text in Geist (sans) and Geist Mono, matching the
// shadcn/ui reference. Because the public gogpu/ui registration path
// (plugin.AssetLoader.LoadFont) registers every font as a single
// weight-400 face, and widget.TextStyle addresses fonts only by family
// name plus a Bold flag, each weight is registered under its own family
// name (see DESIGN.md section 5.6):
//
//	400 -> "Geist"
//	500 -> "Geist Medium"
//	600 -> "Geist SemiBold"
//	700 -> "Geist Bold"
//	400 -> "Geist Mono"
//	500 -> "Geist Mono Medium"
//
// Call [Load] once at startup (graft.Install does this) and resolve
// family names with [Family] and [MonoFamily].
//
// # License
//
// The Geist and Geist Mono font families are copyright (c) 2023 Vercel,
// in collaboration with basement.studio, and are licensed under the SIL
// Open Font License 1.1 (OFL-1.1). The full license text is embedded in
// this directory as OFL.txt and at
// https://github.com/vercel/geist-font/blob/main/LICENSE.TXT.
package fonts

package theme

// Radius scale derived from the --radius knob, mirroring shadcn's Tailwind
// v4 @theme block (docs/research/03-shadcn-pixel-spec.md §2 — authoritative
// per the design brief):
//
//	--radius-sm:  calc(var(--radius) * 0.6)
//	--radius-md:  calc(var(--radius) * 0.8)
//	--radius-lg:  var(--radius)
//	--radius-xl:  calc(var(--radius) * 1.4)
//	--radius-2xl: calc(var(--radius) * 1.8)
//	--radius-3xl: calc(var(--radius) * 2.2)
//	--radius-4xl: calc(var(--radius) * 2.6)
//
// These are the current (2025+) multiplicative formulas shipped by both
// the shadcn docs site and `shadcn init`. An earlier Tailwind v4 revision
// used additive offsets instead (sm = r-4px, md = r-2px, xl = r+4px);
// Report 3 §2 explicitly marks that variant as superseded, and at the
// default --radius of 10px both produce identical pixels (6/8/10/14), so
// the multiplicative form is implemented. RadiusXS is Tailwind's default
// --radius-xs (0.125rem = 2px), which shadcn does not override, and Full
// is the 9999px pill clamp.
//
// Component code never hardcodes 6/8/10/14 — it asks the theme, so the
// user's Radius knob propagates everywhere, exactly like shadcn.

// RadiusXS returns the rounded-xs radius: fixed 2px (Tailwind default,
// independent of --radius).
func (t *Theme) RadiusXS() float32 { return 2 }

// RadiusSM returns the rounded-sm radius: 0.6 × Radius (6px at the
// default 10).
func (t *Theme) RadiusSM() float32 { return max(0, t.Radius*0.6) }

// RadiusMD returns the rounded-md radius: 0.8 × Radius (8px at the
// default 10).
func (t *Theme) RadiusMD() float32 { return max(0, t.Radius*0.8) }

// RadiusLG returns the rounded-lg radius: Radius itself (10px default).
func (t *Theme) RadiusLG() float32 { return max(0, t.Radius) }

// RadiusXL returns the rounded-xl radius: 1.4 × Radius (14px at the
// default 10).
func (t *Theme) RadiusXL() float32 { return max(0, t.Radius*1.4) }

// Radius2XL returns the rounded-2xl radius: 1.8 × Radius (18px at the
// default 10).
func (t *Theme) Radius2XL() float32 { return max(0, t.Radius*1.8) }

// Radius3XL returns the rounded-3xl radius: 2.2 × Radius (22px at the
// default 10).
func (t *Theme) Radius3XL() float32 { return max(0, t.Radius*2.2) }

// Radius4XL returns the rounded-4xl radius: 2.6 × Radius (26px at the
// default 10).
func (t *Theme) Radius4XL() float32 { return max(0, t.Radius*2.6) }

// RadiusFull returns the rounded-full pill radius (9999px; round-rect
// drawing clamps it to the half-extent).
func (t *Theme) RadiusFull() float32 { return 9999 }

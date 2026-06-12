package metrics

import "testing"

// TestShadowTables pins the shadow layer tables to the values frozen in
// DESIGN.md section 5.5.
func TestShadowTables(t *testing.T) {
	tests := []struct {
		name   string
		layers []ShadowLayer
		want   []ShadowLayer
	}{
		{"ShadowXS", ShadowXS, []ShadowLayer{{1, 0, 0.04}, {1, 1, 0.03}}},
		{"ShadowSM", ShadowSM, []ShadowLayer{{1, 0, 0.06}, {1, 1, 0.05}, {2, 2, 0.03}}},
		{"ShadowMD", ShadowMD, []ShadowLayer{{2, 0, 0.06}, {4, 2, 0.05}, {6, 4, 0.03}}},
		{"ShadowLG", ShadowLG, []ShadowLayer{{4, 2, 0.05}, {8, 5, 0.04}, {12, 9, 0.03}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.layers) != len(tt.want) {
				t.Fatalf("%s has %d layers, want %d", tt.name, len(tt.layers), len(tt.want))
			}
			for i, l := range tt.layers {
				if l != tt.want[i] {
					t.Errorf("%s[%d] = %+v, want %+v", tt.name, i, l, tt.want[i])
				}
			}
		})
	}
}

// TestShadowAlphaCeiling enforces the DESIGN.md section 5.5 rule that layer
// alphas stay at or below 0.08 to avoid hard edges.
func TestShadowAlphaCeiling(t *testing.T) {
	tables := map[string][]ShadowLayer{
		"ShadowXS": ShadowXS,
		"ShadowSM": ShadowSM,
		"ShadowMD": ShadowMD,
		"ShadowLG": ShadowLG,
	}
	for name, layers := range tables {
		for i, l := range layers {
			if l.Alpha > 0.08 {
				t.Errorf("%s[%d].Alpha = %v, exceeds 0.08 ceiling", name, i, l.Alpha)
			}
		}
	}
}

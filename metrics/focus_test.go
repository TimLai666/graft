package metrics

import "testing"

// TestFocusConstants pins the shared focus-ring values to the shadcn spec
// (DESIGN.md section 5.1 "Shared"). Any change here is a spec change.
func TestFocusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  float32
		want float32
	}{
		{"RingWidth", RingWidth, 3},
		{"RingAlpha", RingAlpha, 0.5},
		{"InvalidRingAlphaLight", InvalidRingAlphaLight, 0.2},
		{"InvalidRingAlphaDark", InvalidRingAlphaDark, 0.4},
		{"SliderRingWidth", SliderRingWidth, 4},
		{"LegacyCloseRingWidth", LegacyCloseRingWidth, 2},
		{"LegacyCloseRingOffset", LegacyCloseRingOffset, 2},
		{"DisabledOpacity", DisabledOpacity, 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

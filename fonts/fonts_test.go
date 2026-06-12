package fonts

import (
	"sync"
	"testing"

	"golang.org/x/image/font/sfnt"
)

func TestLoad(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatalf("Load() = %v, want nil", err)
	}
}

func TestLoadIdempotent(t *testing.T) {
	first := Load()
	second := Load()
	if first != second {
		t.Fatalf("Load() not idempotent: first = %v, second = %v", first, second)
	}
}

func TestLoadConcurrent(t *testing.T) {
	const goroutines = 16
	var wg sync.WaitGroup
	errs := make([]error, goroutines)
	for i := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = Load()
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: Load() = %v, want nil", i, err)
		}
	}
}

func TestFamily(t *testing.T) {
	tests := []struct {
		weight int
		want   string
	}{
		{100, FamilySans},
		{300, FamilySans},
		{400, FamilySans},
		{450, FamilySans}, // tie resolves to lighter
		{500, FamilyMedium},
		{550, FamilyMedium}, // tie resolves to lighter
		{600, FamilySemiBold},
		{650, FamilySemiBold}, // tie resolves to lighter
		{700, FamilyBold},
		{800, FamilyBold},
		{900, FamilyBold},
	}
	for _, tt := range tests {
		if got := Family(tt.weight); got != tt.want {
			t.Errorf("Family(%d) = %q, want %q", tt.weight, got, tt.want)
		}
	}
}

func TestMonoFamily(t *testing.T) {
	tests := []struct {
		weight int
		want   string
	}{
		{100, FamilyMono},
		{400, FamilyMono},
		{450, FamilyMono}, // tie resolves to lighter
		{500, FamilyMonoMedium},
		{600, FamilyMonoMedium},
		{700, FamilyMonoMedium},
		{900, FamilyMonoMedium},
	}
	for _, tt := range tests {
		if got := MonoFamily(tt.weight); got != tt.want {
			t.Errorf("MonoFamily(%d) = %q, want %q", tt.weight, got, tt.want)
		}
	}
}

func TestFaces(t *testing.T) {
	got := Faces()
	want := map[string]struct {
		weight int
		mono   bool
	}{
		FamilySans:       {400, false},
		FamilyMedium:     {500, false},
		FamilySemiBold:   {600, false},
		FamilyBold:       {700, false},
		FamilyMono:       {400, true},
		FamilyMonoMedium: {500, true},
	}
	if len(got) != len(want) {
		t.Fatalf("Faces() returned %d entries, want %d", len(got), len(want))
	}
	for family, w := range want {
		f, ok := got[family]
		if !ok {
			t.Errorf("Faces() missing family %q", family)
			continue
		}
		if f.Family != family {
			t.Errorf("Faces()[%q].Family = %q, want %q", family, f.Family, family)
		}
		if f.Weight != w.weight {
			t.Errorf("Faces()[%q].Weight = %d, want %d", family, f.Weight, w.weight)
		}
		if f.Mono != w.mono {
			t.Errorf("Faces()[%q].Mono = %v, want %v", family, f.Mono, w.mono)
		}
		if len(f.Data) == 0 {
			t.Errorf("Faces()[%q].Data is empty", family)
		}
	}
}

func TestEmbeddedFacesParseAsSFNT(t *testing.T) {
	for family, face := range Faces() {
		f, err := sfnt.Parse(face.Data)
		if err != nil {
			t.Errorf("sfnt.Parse(%q) failed: %v", family, err)
			continue
		}
		if n := f.NumGlyphs(); n == 0 {
			t.Errorf("font %q has 0 glyphs", family)
		}
	}
}

func TestData(t *testing.T) {
	for _, family := range []string{
		FamilySans, FamilyMedium, FamilySemiBold, FamilyBold,
		FamilyMono, FamilyMonoMedium,
	} {
		if len(Data(family)) == 0 {
			t.Errorf("Data(%q) is empty", family)
		}
	}
	if Data("Comic Sans MS") != nil {
		t.Error("Data() for unknown family should return nil")
	}
}

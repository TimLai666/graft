package fonts

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/gogpu/ui/plugin"
)

// Embedded Geist TTF data (OFL-1.1, copyright Vercel; see doc.go and OFL.txt).
var (
	//go:embed Geist-Regular.ttf
	geistRegular []byte

	//go:embed Geist-Medium.ttf
	geistMedium []byte

	//go:embed Geist-SemiBold.ttf
	geistSemiBold []byte

	//go:embed Geist-Bold.ttf
	geistBold []byte

	//go:embed GeistMono-Regular.ttf
	geistMonoRegular []byte

	//go:embed GeistMono-Medium.ttf
	geistMonoMedium []byte
)

// Registered family names. One family per weight: the public gogpu/ui
// registration path (plugin.AssetLoader.LoadFont -> internal render
// FontRegistry.Register) stores every font as a weight-400 normal face,
// and widget.TextStyle selects fonts by family name plus a Bold flag
// only, so distinct weights must live under distinct family names.
const (
	// FamilySans is the Geist Regular (weight 400) family name.
	FamilySans = "Geist"

	// FamilyMedium is the Geist Medium (weight 500) family name.
	FamilyMedium = "Geist Medium"

	// FamilySemiBold is the Geist SemiBold (weight 600) family name.
	FamilySemiBold = "Geist SemiBold"

	// FamilyBold is the Geist Bold (weight 700) family name.
	//
	// Note: TextStyle.Bold=true with family "Geist" does NOT resolve to
	// this face. The render layer asks the CSS matcher for weight 700
	// within the "Geist" family, which only has a 400 face registered, so
	// it falls back to Regular. Callers that want bold text must request
	// family Family(700) explicitly.
	FamilyBold = "Geist Bold"

	// FamilyMono is the Geist Mono Regular (weight 400) family name.
	FamilyMono = "Geist Mono"

	// FamilyMonoMedium is the Geist Mono Medium (weight 500) family name.
	FamilyMonoMedium = "Geist Mono Medium"
)

// Face describes a single embedded font face.
//
// Each face is registered with the text pipeline under its own Family
// name. Weight records the true CSS weight of the underlying TTF so
// that consumers performing their own text measurement (for example
// graft's internal/textmetrics) can index faces by weight.
type Face struct {
	// Family is the family name the face is registered under,
	// e.g. "Geist Medium".
	Family string

	// Weight is the CSS weight of the underlying font file
	// (400, 500, 600, or 700).
	Weight int

	// Mono reports whether the face belongs to Geist Mono.
	Mono bool

	// Data is the raw TTF file contents. Callers must not mutate it.
	Data []byte
}

// faces lists all embedded faces in registration order.
// Sans faces are ordered by ascending weight, then mono faces.
var faces = []Face{
	{Family: FamilySans, Weight: 400, Mono: false, Data: geistRegular},
	{Family: FamilyMedium, Weight: 500, Mono: false, Data: geistMedium},
	{Family: FamilySemiBold, Weight: 600, Mono: false, Data: geistSemiBold},
	{Family: FamilyBold, Weight: 700, Mono: false, Data: geistBold},
	{Family: FamilyMono, Weight: 400, Mono: true, Data: geistMonoRegular},
	{Family: FamilyMonoMedium, Weight: 500, Mono: true, Data: geistMonoMedium},
}

// Faces returns all embedded font faces keyed by registered family name.
//
// The returned map is freshly allocated on every call, but the Data
// slices are shared with the package and must not be mutated. This is
// the wiring point for graft's internal/textmetrics: the root package
// feeds every (family, data) pair into the measurer at install time.
func Faces() map[string]Face {
	m := make(map[string]Face, len(faces))
	for _, f := range faces {
		m[f.Family] = f
	}
	return m
}

// Data returns the raw TTF bytes for the given registered family name,
// or nil if the family is not one of the embedded Geist families.
// The returned slice is shared and must not be mutated.
func Data(family string) []byte {
	for _, f := range faces {
		if f.Family == family {
			return f.Data
		}
	}
	return nil
}

var (
	loadOnce sync.Once
	loadErr  error
)

// Load registers all embedded Geist faces with the gogpu/ui text
// rendering pipeline. It is safe to call from multiple goroutines and
// runs at most once; subsequent calls return the first result.
//
// Registration goes through plugin.NewDefaultPluginContext, whose asset
// loader forwards each font to the process-global render font registry
// (the same registry widget.StyledTextDrawer resolves families from).
// That path parses each TTF eagerly, so a nil return also guarantees the
// embedded data is valid font data.
func Load() error {
	loadOnce.Do(func() {
		assets := plugin.NewDefaultPluginContext().Assets
		for _, f := range faces {
			if err := assets.LoadFont(f.Family, f.Data); err != nil {
				loadErr = fmt.Errorf("fonts: registering family %q: %w", f.Family, err)
				return
			}
		}
	})
	return loadErr
}

// weightEntry pairs a CSS weight with its registered family name.
type weightEntry struct {
	weight int
	family string
}

// sansWeights and monoWeights are ordered by ascending weight so that
// nearestFamily resolves ties toward the lighter face.
var (
	sansWeights = []weightEntry{
		{400, FamilySans},
		{500, FamilyMedium},
		{600, FamilySemiBold},
		{700, FamilyBold},
	}
	monoWeights = []weightEntry{
		{400, FamilyMono},
		{500, FamilyMonoMedium},
	}
)

// Family returns the registered Geist family name for the given CSS
// font weight: 400 -> "Geist", 500 -> "Geist Medium", 600 -> "Geist
// SemiBold", 700 -> "Geist Bold". Other weights resolve to the nearest
// available weight; exact ties resolve to the lighter face.
func Family(weight int) string {
	return nearestFamily(weight, sansWeights)
}

// MonoFamily returns the registered Geist Mono family name for the
// given CSS font weight: 400 -> "Geist Mono", 500 -> "Geist Mono
// Medium". Other weights resolve to the nearest available weight;
// exact ties resolve to the lighter face.
func MonoFamily(weight int) string {
	return nearestFamily(weight, monoWeights)
}

// nearestFamily returns the family whose weight is closest to the
// requested weight. Entries must be ordered by ascending weight; on an
// exact distance tie the lighter (earlier) entry wins.
func nearestFamily(weight int, entries []weightEntry) string {
	best := entries[0]
	bestDist := absInt(weight - best.weight)
	for _, e := range entries[1:] {
		if d := absInt(weight - e.weight); d < bestDist {
			best, bestDist = e, d
		}
	}
	return best.family
}

// absInt returns the absolute value of v.
func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

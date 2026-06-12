package icons

import (
	"bytes"
	"embed"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/gogpu/ui/icon"
)

//go:embed lucide/*.svg
var lucideFS embed.FS

// RegistryPrefix is prepended to the Lucide icon file name to form the
// name used in icon registries, e.g. "lucide:chevron-down".
const RegistryPrefix = "lucide:"

// lucideViewBox is the side length of the square Lucide icon viewBox.
const lucideViewBox = 24

// Lucide icons used by graft widgets. Each value carries the normalized
// SVG XML (see normalizeLucideSVG) and renders through widget.SVGRenderer
// with the stroke color overridden by the requested icon color.
var (
	// ArrowLeft is the lucide arrow-left icon.
	ArrowLeft = mustLucide("arrow-left")

	// ArrowRight is the lucide arrow-right icon.
	ArrowRight = mustLucide("arrow-right")

	// Calendar is the lucide calendar icon.
	Calendar = mustLucide("calendar")

	// Check is the lucide check icon.
	Check = mustLucide("check")

	// ChevronDown is the lucide chevron-down icon.
	ChevronDown = mustLucide("chevron-down")

	// ChevronLeft is the lucide chevron-left icon.
	ChevronLeft = mustLucide("chevron-left")

	// ChevronRight is the lucide chevron-right icon.
	ChevronRight = mustLucide("chevron-right")

	// ChevronUp is the lucide chevron-up icon.
	ChevronUp = mustLucide("chevron-up")

	// ChevronsUpDown is the lucide chevrons-up-down icon.
	ChevronsUpDown = mustLucide("chevrons-up-down")

	// Circle is the lucide circle icon.
	Circle = mustLucide("circle")

	// CircleAlert is the lucide circle-alert icon.
	CircleAlert = mustLucide("circle-alert")

	// CircleCheck is the lucide circle-check icon.
	CircleCheck = mustLucide("circle-check")

	// Ellipsis is the lucide ellipsis icon.
	Ellipsis = mustLucide("ellipsis")

	// GripVertical is the lucide grip-vertical icon.
	GripVertical = mustLucide("grip-vertical")

	// Info is the lucide info icon.
	Info = mustLucide("info")

	// LoaderCircle is the lucide loader-circle icon (spinner).
	LoaderCircle = mustLucide("loader-circle")

	// Minus is the lucide minus icon.
	Minus = mustLucide("minus")

	// PanelLeft is the lucide panel-left icon.
	PanelLeft = mustLucide("panel-left")

	// Search is the lucide search icon.
	Search = mustLucide("search")

	// TriangleAlert is the lucide triangle-alert icon.
	TriangleAlert = mustLucide("triangle-alert")

	// X is the lucide x icon.
	X = mustLucide("x")
)

// all returns every embedded Lucide icon in stable (alphabetical) order.
func all() []icon.IconData {
	return []icon.IconData{
		ArrowLeft, ArrowRight, Calendar, Check,
		ChevronDown, ChevronLeft, ChevronRight, ChevronUp, ChevronsUpDown,
		Circle, CircleAlert, CircleCheck, Ellipsis, GripVertical,
		Info, LoaderCircle, Minus, PanelLeft, Search, TriangleAlert, X,
	}
}

// All returns every embedded Lucide icon in stable (alphabetical) order.
// The returned slice is freshly allocated; the IconData values share
// their underlying SVG data, which must not be mutated.
func All() []icon.IconData {
	return all()
}

var registerOnce sync.Once

// Register adds all embedded Lucide icons to icon.DefaultRegistry()
// under "lucide:<name>" keys (e.g. "lucide:check"). It is safe to call
// from multiple goroutines and registers at most once; subsequent calls
// are no-ops returning nil.
//
// The error return is reserved for future registration paths that can
// fail; the current implementation always returns nil.
func Register() error {
	registerOnce.Do(func() {
		reg := icon.DefaultRegistry()
		for _, ic := range all() {
			reg.Register(ic)
		}
	})
	return nil
}

// mustLucide loads, normalizes, and converts an embedded Lucide SVG into
// an icon.IconData. It panics on failure: the inputs are embedded assets,
// so any error is a programming or vendoring mistake that should fail
// fast at package init.
func mustLucide(name string) icon.IconData {
	raw, err := lucideFS.ReadFile("lucide/" + name + ".svg")
	if err != nil {
		panic("icons: missing embedded lucide icon " + name + ": " + err.Error())
	}
	normalized, err := normalizeLucideSVG(raw)
	if err != nil {
		panic("icons: normalizing lucide icon " + name + ": " + err.Error())
	}
	data := icon.FromSVGXML(RegistryPrefix+name, normalized)
	// FromSVGXML defaults ViewBox to 16; record the true Lucide viewBox.
	// Rendering reads the viewBox from the SVG XML itself, but the field
	// should be accurate for any consumer inspecting the IconData.
	data.ViewBox = lucideViewBox
	return data
}

// presentationAttrs are the SVG presentation attributes Lucide sets on
// the root <svg> element, in canonical output order.
var presentationAttrs = []string{
	"fill", "stroke", "stroke-width", "stroke-linecap", "stroke-linejoin",
}

// shapeElements are the SVG shape elements that receive inherited
// presentation attributes during normalization. Matches the element set
// supported by the gogpu/gg SVG parser.
var shapeElements = map[string]bool{
	"path":     true,
	"circle":   true,
	"rect":     true,
	"ellipse":  true,
	"line":     true,
	"polygon":  true,
	"polyline": true,
}

// normalizeLucideSVG rewrites a Lucide SVG so that it renders correctly
// through the gogpu/gg SVG renderer.
//
// Lucide icons declare all presentation attributes (fill="none",
// stroke="currentColor", stroke-width="2", stroke-linecap="round",
// stroke-linejoin="round") on the root <svg> element and rely on CSS
// inheritance. The gogpu/gg parser propagates only the root fill
// attribute; root stroke, stroke-width, stroke-linecap, and
// stroke-linejoin are dropped, which would leave shapes filled with
// solid color instead of stroked. The normalizer copies each root
// presentation attribute onto every shape element that does not already
// declare it, making the inheritance explicit.
//
// Limitation: attributes declared on intermediate <g> elements are not
// considered when deciding whether a shape already has an attribute.
// Lucide icons do not use <g> elements, so this does not affect the
// vendored set.
func normalizeLucideSVG(data []byte) ([]byte, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var buf bytes.Buffer
	var rootAttrs []xml.Attr
	sawRoot := false

	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("icons: invalid SVG XML: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch {
			case !sawRoot && t.Name.Local == "svg":
				sawRoot = true
				rootAttrs = collectPresentationAttrs(t.Attr)
				writeStartElement(&buf, t, nil)
			case shapeElements[t.Name.Local]:
				writeStartElement(&buf, t, missingAttrs(t.Attr, rootAttrs))
			default:
				writeStartElement(&buf, t, nil)
			}
		case xml.EndElement:
			buf.WriteString("</")
			buf.WriteString(t.Name.Local)
			buf.WriteByte('>')
		default:
			// Drop character data (whitespace only in icon SVGs),
			// comments, directives, and processing instructions.
		}
	}

	if !sawRoot {
		return nil, errors.New("icons: no <svg> root element found")
	}
	return buf.Bytes(), nil
}

// collectPresentationAttrs filters attrs down to the recognized SVG
// presentation attributes, preserving canonical order.
func collectPresentationAttrs(attrs []xml.Attr) []xml.Attr {
	var out []xml.Attr
	for _, name := range presentationAttrs {
		for _, a := range attrs {
			if a.Name.Space == "" && a.Name.Local == name {
				out = append(out, a)
				break
			}
		}
	}
	return out
}

// missingAttrs returns the subset of inherited that is not already
// present in have.
func missingAttrs(have, inherited []xml.Attr) []xml.Attr {
	var out []xml.Attr
	for _, in := range inherited {
		found := false
		for _, a := range have {
			if a.Name.Space == "" && a.Name.Local == in.Name.Local {
				found = true
				break
			}
		}
		if !found {
			out = append(out, in)
		}
	}
	return out
}

// writeStartElement serializes a start element with its original
// attributes followed by any extra (inherited) attributes.
//
// Element and attribute names are written using their local names; the
// gogpu/gg SVG parser matches local names only, so namespace prefixes
// are irrelevant. The default namespace declaration (xmlns) is
// preserved as-is.
func writeStartElement(buf *bytes.Buffer, se xml.StartElement, extra []xml.Attr) {
	buf.WriteByte('<')
	buf.WriteString(se.Name.Local)
	for _, a := range se.Attr {
		writeAttr(buf, a)
	}
	for _, a := range extra {
		writeAttr(buf, a)
	}
	buf.WriteByte('>')
}

// writeAttr serializes a single XML attribute with proper escaping.
func writeAttr(buf *bytes.Buffer, a xml.Attr) {
	buf.WriteByte(' ')
	if a.Name.Space == "xmlns" {
		buf.WriteString("xmlns:")
	}
	buf.WriteString(a.Name.Local)
	buf.WriteString(`="`)
	_ = xml.EscapeText(buf, []byte(a.Value))
	buf.WriteByte('"')
}

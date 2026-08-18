// Package pill renders a labeled cycle selector - "type       ‹ Task ›" - a
// compact one-of chooser for a form field that needs no text area. It is the
// counterpart to titlebox: where a box frames somewhere to type, a pill frames
// a value stepped through a fixed set, so the two read as distinct affordances.
//
// Like titlebox it is style-agnostic: the label and chevron colors arrive
// through [Styles], so a caller signals focus by handing it a brighter pair.
package pill

import (
	"strings"

	lg "charm.land/lipgloss/v2"
)

// Chevrons is the glyph pair framing the value. The glyphs travel as a pair:
// the zero value renders the default "‹ ›", and an override only applies when
// both are set, so a half-set pair can never render lopsided.
type Chevrons struct {
	Left  string
	Right string
}

// set reports a complete override pair.
func (c Chevrons) set() bool { return c.Left != "" && c.Right != "" }

// defaultChevrons frames the value when the caller overrides nothing.
var defaultChevrons = Chevrons{Left: "‹", Right: "›"}

// Styles are the pill's injected render styles, so the widget stays
// theme-agnostic: a caller signals focus by handing it a brighter pair.
type Styles struct {
	// Label styles the field name to the left of the value.
	Label lg.Style
	// Chevron styles the markers that frame the value.
	Chevron lg.Style
	// Chevrons overrides the marker glyphs (both together, or neither).
	Chevrons Chevrons
}

// Render draws "label<pad>‹ value ›". label is padded on the right to labelW
// columns so a column of pills aligns their values; a label already at or past
// labelW is left as is. The value carries no styling of its own - the caller
// styles it before passing it in if it wants to.
func Render(label, value string, labelW int, s Styles) string {
	if pad := labelW - lg.Width(label); pad > 0 {
		label += strings.Repeat(" ", pad)
	}
	chevrons := s.Chevrons
	if !chevrons.set() {
		chevrons = defaultChevrons
	}
	return s.Label.Render(label) +
		s.Chevron.Render(chevrons.Left+" ") +
		value +
		s.Chevron.Render(" "+chevrons.Right)
}

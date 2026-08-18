package form

import (
	"strings"

	"github.com/gechr/primer/titlebox"
)

// FocusAt routes a mouse click at Body-space coordinates (x, y) - e.g. a
// dialog's translated ClickMsg - to the field rendered there. One click does
// both: it focuses the field and, on a text field, moves the cursor to the
// clicked spot. It reports whether a field was hit; the title and anything
// below the last block miss.
func (m *Model) FocusAt(x, y int) bool {
	if !m.Active() {
		return false
	}
	top := 0
	if m.title != "" {
		top += strings.Count(m.styles.Title.Render(m.title), "\n") + 1
	}
	for i := range m.fields {
		span := strings.Count(m.fieldBlock(i), "\n")
		if y < top || y >= top+span {
			top += span
			continue
		}
		m.setFocus(i)
		m.placeCursor(i, x, y-top)
		return true
	}
	return false
}

// placeCursor lands a click inside field i's box: rowInBlock is
// the clicked row within the field's block, whose body starts below the titled
// top border (and, for every block after the first, the blank separator line),
// and titlebox.Inset columns in.
func (m *Model) placeCursor(i, x, rowInBlock int) {
	f := &m.fields[i]
	if f.isCycle() {
		return
	}
	innerY := rowInBlock - 1
	if i > 0 {
		innerY-- // the block opens on its separator line
	}
	innerX := x - titlebox.Inset
	switch {
	case innerY < 0:
		// The separator or border row carries no text to land on.
	case f.spec.Multiline:
		f.area.SetCursorAt(innerY, innerX)
	case innerY == 0:
		f.line.SetCursorAt(innerX)
	}
}

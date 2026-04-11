package helpsheet

import (
	"strings"

	lg "charm.land/lipgloss/v2"
)

// Pair is a single key/description entry in a help sheet.
type Pair struct {
	Key  string
	Desc string
}

// Styles controls the appearance of the help sheet.
type Styles struct {
	Key     lg.Style
	Text    lg.Style
	Dismiss lg.Style
	Box     lg.Style
}

// Model describes a two-column keybinding help sheet with an optional
// centered dismiss footer.
type Model struct {
	Pairs   []Pair
	Dismiss string // e.g. "Press any key to dismiss"
	Styles  Styles
	Gutter  int // horizontal space between columns (default 4)
}

// Render produces the help sheet as a styled, box-wrapped string.
func (m Model) Render() string {
	if len(m.Pairs) == 0 {
		return ""
	}

	gutter := m.Gutter
	if gutter <= 0 {
		gutter = 4
	}

	rows := (len(m.Pairs) + 1) / 2 //nolint:mnd // ceil division for two columns

	// Measure key column width for right-alignment.
	keyWidth := 0
	for _, p := range m.Pairs {
		if w := lg.Width(p.Key); w > keyWidth {
			keyWidth = w
		}
	}

	renderPair := func(p Pair) string {
		pad := max(0, keyWidth-lg.Width(p.Key))
		key := strings.Repeat(" ", pad) + p.Key
		return m.Styles.Key.Render(key) + "  " + m.Styles.Text.Render(p.Desc)
	}

	// Measure rendered column widths for alignment.
	leftWidth := 0
	rightWidth := 0
	for i := range rows {
		if w := lg.Width(renderPair(m.Pairs[i])); w > leftWidth {
			leftWidth = w
		}
		if i+rows < len(m.Pairs) {
			if w := lg.Width(renderPair(m.Pairs[i+rows])); w > rightWidth {
				rightWidth = w
			}
		}
	}
	totalWidth := leftWidth + gutter + rightWidth

	var b strings.Builder
	for i := range rows {
		left := renderPair(m.Pairs[i])
		if i+rows < len(m.Pairs) {
			right := renderPair(m.Pairs[i+rows])
			pad := leftWidth - lg.Width(left) + gutter
			b.WriteString(left + strings.Repeat(" ", pad) + right)
		} else {
			b.WriteString(left)
		}
		b.WriteString("\n")
	}

	if m.Dismiss != "" {
		dismiss := m.Styles.Dismiss.Render(m.Dismiss)
		pad := (totalWidth - lg.Width(dismiss)) / 2 //nolint:mnd // centering
		if pad > 0 {
			b.WriteString("\n" + strings.Repeat(" ", pad) + dismiss)
		} else {
			b.WriteString("\n" + dismiss)
		}
	}

	return m.Styles.Box.Render(b.String())
}

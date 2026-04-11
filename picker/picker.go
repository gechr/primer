package picker

import (
	"strings"

	lg "charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/layout"
)

// Row defines a single option row with a label and selectable choices.
type Row struct {
	Label   string
	Choices []string
}

// Styles controls the visual appearance of the picker overlay.
type Styles struct {
	Box            lg.Style // outer box style
	Cursor         string   // prefix for the cursor row (e.g. "❯ ")
	CursorLineBG   string   // raw ANSI background escape for cursor row (optional)
	CursorPad      string   // prefix for non-cursor rows (e.g. "  ")
	Default        lg.Style // default (but not selected) choice
	HelpKey        lg.Style // key hint style in footer
	HelpText       lg.Style // text hint style in footer
	Inactive       lg.Style // unselected choices
	Label          lg.Style // regular label style
	LockedInactive lg.Style // unselected choices on locked rows
	LockedLabel    lg.Style // label style for locked rows
	LockedSuffix   string   // text appended to locked rows (e.g. "  (CLI)")
	Selected       lg.Style // currently selected choice
}

// HelpHint defines a key/description pair for the footer.
type HelpHint struct {
	Key  string
	Desc string
}

// Model holds the state for an options picker overlay.
type Model struct {
	Rows     []Row
	Cursor   int
	Values   []int
	Defaults []int
	Locked   []bool
	IsReset  []bool
	Styles   Styles
	Hints    []HelpHint
}

// New creates a picker Model with the given rows and initial values.
// defaults, locked, and isReset slices must have the same length as rows.
func New(rows []Row, values, defaults []int, locked, isReset []bool, styles Styles) Model {
	return Model{
		Rows:     rows,
		Values:   values,
		Defaults: defaults,
		Locked:   locked,
		IsReset:  isReset,
		Styles:   styles,
	}
}

// Up moves the cursor up one row.
func (m *Model) Up() {
	m.Cursor = max(m.Cursor-1, 0)
}

// Down moves the cursor down one row.
func (m *Model) Down() {
	m.Cursor = min(m.Cursor+1, len(m.Rows)-1)
}

// Right selects the next choice in the current row (clamped).
func (m *Model) Right() {
	if m.isLocked() {
		return
	}
	m.IsReset[m.Cursor] = false
	n := len(m.Rows[m.Cursor].Choices)
	m.Values[m.Cursor] = min(m.Values[m.Cursor]+1, n-1)
}

// Left selects the previous choice in the current row (clamped).
func (m *Model) Left() {
	if m.isLocked() {
		return
	}
	m.IsReset[m.Cursor] = false
	m.Values[m.Cursor] = max(m.Values[m.Cursor]-1, 0)
}

// Cycle advances to the next choice, wrapping around to the first.
func (m *Model) Cycle() {
	if m.isLocked() {
		return
	}
	m.IsReset[m.Cursor] = false
	n := len(m.Rows[m.Cursor].Choices)
	if n > 0 {
		m.Values[m.Cursor] = (m.Values[m.Cursor] + 1) % n
	}
}

// Reset restores the current row to its default value.
func (m *Model) Reset() {
	if m.isLocked() {
		return
	}
	m.IsReset[m.Cursor] = true
	m.Values[m.Cursor] = m.Defaults[m.Cursor]
}

func (m *Model) isLocked() bool {
	return m.Cursor < len(m.Locked) && m.Locked[m.Cursor]
}

// View renders the picker overlay as a styled string.
func (m *Model) View() string {
	var b strings.Builder
	s := m.Styles

	labelWidth := 0
	for _, row := range m.Rows {
		if w := len(row.Label); w > labelWidth {
			labelWidth = w
		}
	}

	lines := make([]string, 0, len(m.Rows))
	for i, row := range m.Rows {
		var line strings.Builder
		locked := i < len(m.Locked) && m.Locked[i]

		if i == m.Cursor {
			line.WriteString(s.Cursor)
		} else {
			line.WriteString(s.CursorPad)
		}

		pad := strings.Repeat(" ", labelWidth-len(row.Label))
		label := pad + row.Label + "  "
		if locked {
			line.WriteString(s.LockedLabel.Render(label))
		} else {
			line.WriteString(s.Label.Render(label))
		}

		for j, choice := range row.Choices {
			if j > 0 {
				line.WriteString("  ")
			}
			selected := i < len(m.Values) && m.Values[i] == j
			isDefault := i < len(m.Defaults) && m.Defaults[i] == j
			switch {
			case selected:
				line.WriteString(s.Selected.Render(choice))
			case isDefault:
				line.WriteString(s.Default.Render(choice))
			case locked:
				line.WriteString(s.LockedInactive.Render(choice))
			default:
				line.WriteString(s.Inactive.Render(choice))
			}
		}

		if locked && s.LockedSuffix != "" {
			line.WriteString(s.LockedLabel.Render(s.LockedSuffix))
		}
		lines = append(lines, line.String())
	}

	footer := m.renderFooter()
	contentWidth := lg.Width(footer)
	for _, line := range lines {
		contentWidth = max(contentWidth, lg.Width(line))
	}

	for i, line := range lines {
		if i == m.Cursor && s.CursorLineBG != "" {
			b.WriteString(layout.PreserveBackgroundWidth(line, s.CursorLineBG, contentWidth))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(footer)

	return s.Box.Render(b.String())
}

func (m *Model) renderFooter() string {
	if len(m.Hints) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m.Hints))
	for _, h := range m.Hints {
		parts = append(parts, m.Styles.HelpKey.Render(h.Key)+m.Styles.HelpText.Render(" "+h.Desc))
	}
	sep := xansi.Strip(m.Styles.HelpText.Render("  "))
	return strings.Join(parts, sep)
}

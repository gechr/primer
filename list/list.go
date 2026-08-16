// Package list is a scrollable, single-selection list of pre-rendered rows.
// Its defining property is that the selected item (cursor) is tracked
// independently of the scroll offset: moving the cursor adjusts the offset
// only enough to keep the selection visible, and a resize never loses the
// selection. Rows are opaque strings, so the component is decoupled from how
// a caller styles them - it owns only the selection/scroll math and an
// optional scrollbar.
//
// [WithSearch] adds a type-to-filter line above the rows: typing narrows the
// list fzf-style (every printable key feeds the filter, arrows navigate), and
// the caller reads [Model.Selected] on its own submit key - the list never
// consumes enter or esc. Matching is a case-insensitive substring test
// against each row's visible (ANSI-stripped) text, and selection reports the
// original row index, so the caller keeps its own parallel value slice.
package list

import (
	"regexp"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/filter"
	"github.com/gechr/primer/scrollbar"
)

// sgrPattern matches ANSI SGR (color/style) escape sequences. The cursor row is
// stripped of them before the highlight is applied, because an inner reset
// (\x1b[m emitted by a styled cell) would otherwise terminate the highlight
// partway across the row.
var sgrPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

// Styles are the list's injected render styles.
type Styles struct {
	// Cursor highlights the selected row. New defaults it to reverse video.
	Cursor lg.Style
	// Dim styles the "(no match)" placeholder shown when a search filter
	// matches nothing.
	Dim lg.Style
}

// Model is a list viewport over caller-rendered rows.
type Model struct {
	rows    []string
	matches []int // search mode: indexes into rows that pass the filter, in order
	cursor  int   // position within the visible (filtered) list
	offset  int   // visible-list index of the first row on screen
	width   int
	height  int // total view height, including the filter line when shown

	search bool
	input  textinput.Model

	// Styles style the cursor row and placeholders.
	Styles Styles
	// Scrollbar draws a proportional scrollbar column when the list overflows.
	Scrollbar bool
}

// Option configures a Model.
type Option func(*Model)

// WithSearch adds the type-to-filter line: every printable key routed through
// Update feeds the filter, up/down move the cursor, and the visible rows are
// the matching subset.
func WithSearch() Option {
	return func(m *Model) {
		ti := textinput.New()
		ti.Prompt = "/ "
		ti.Placeholder = "filter"
		ti.Focus()
		m.search = true
		m.input = ti
	}
}

// New returns a list with a sensible default cursor highlight. It returns a
// pointer because every mutating method has a pointer receiver.
func New(opts ...Option) *Model {
	m := &Model{
		Styles:    Styles{Cursor: lg.NewStyle().Reverse(true)},
		Scrollbar: true,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// SetRows replaces the rows, refilters when searching, and clamps the
// cursor/offset to the new length so the selection stays valid (and visible)
// when the data shrinks.
func (m *Model) SetRows(rows []string) {
	m.rows = rows
	if m.search {
		m.refilter()
		return
	}
	m.clamp()
}

// SetSize sets the viewport dimensions and reclamps so the cursor stays
// visible. height is the total view height: when the filter line is showing,
// one row of it goes to the filter.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.clamp()
}

// Len returns the total number of rows, ignoring any active filter.
func (m *Model) Len() int { return len(m.rows) }

// Cursor returns the original index of the selected row, or -1 when nothing
// is selectable (the list is empty, or a search filter matches nothing).
func (m *Model) Cursor() int {
	if m.visLen() == 0 {
		return -1
	}
	return m.origIndex(m.cursor)
}

// Selected returns the original index of the row under the cursor; ok is
// false when nothing is selectable.
func (m *Model) Selected() (int, bool) {
	if m.visLen() == 0 {
		return 0, false
	}
	return m.origIndex(m.cursor), true
}

// CursorLine returns the cursor row's line offset within View - the filter
// line, when showing, counts as a line above it - or -1 when nothing is
// selectable. It is the region a dialog scroll hint reports so an outer
// viewport keeps the selection visible when a taller-than-the-box list
// scrolls inside it.
func (m *Model) CursorLine() int {
	if m.visLen() == 0 {
		return -1
	}
	line := m.cursor - m.offset
	if m.filterShown() {
		line++
	}
	return line
}

// SetCursor selects visible position i (equal to the row index when no filter
// is active), clamping to range and scrolling it into view.
func (m *Model) SetCursor(i int) {
	m.cursor = i
	m.clamp()
}

// MoveDown moves the cursor down by n (n may be negative).
func (m *Model) MoveDown(n int) { m.SetCursor(m.cursor + n) }

// MoveUp moves the cursor up by n.
func (m *Model) MoveUp(n int) { m.SetCursor(m.cursor - n) }

// Top selects the first visible row.
func (m *Model) Top() { m.SetCursor(0) }

// Bottom selects the last visible row.
func (m *Model) Bottom() { m.SetCursor(m.visLen() - 1) }

// PageDown moves the cursor down by a viewport height.
func (m *Model) PageDown() { m.MoveDown(m.pageStride()) }

// PageUp moves the cursor up by a viewport height.
func (m *Model) PageUp() { m.MoveUp(m.pageStride()) }

// Update routes messages when searching: up/down move the cursor and
// everything else (typing, paste, cursor movement inside the filter) goes to
// the filter input. Arrow keys match on Code, not String(), so modified
// arrows keep working; every printable key belongs to the filter (fzf
// semantics), so letter aliases like j/k are deliberately not navigation
// here. Enter and esc are deliberately not handled - submit/cancel belong to
// the caller. Without search, Update is a no-op: the caller drives the cursor
// through the Move methods.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.search {
		return nil
	}
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.Code {
		case tea.KeyUp:
			m.MoveUp(1)
			return nil
		case tea.KeyDown:
			m.MoveDown(1)
			return nil
		}
	}
	before := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != before {
		m.refilter()
	}
	return cmd
}

// VisibleRange returns the [start, end) visible-list positions currently on
// screen.
func (m *Model) VisibleRange() (int, int) {
	h := m.rowsHeight()
	if h <= 0 || m.visLen() == 0 {
		return 0, 0
	}
	return m.offset, min(m.offset+h, m.visLen())
}

// View renders the filter line (once the user typed something) and the
// visible rows, highlighting the cursor row and appending a scrollbar column
// when the list overflows and Scrollbar is enabled. Rows are padded/truncated
// to the content width (viewport width minus the scrollbar column) so the
// cursor highlight spans the full row and the output never exceeds the
// declared width.
func (m *Model) View() string {
	if m.height <= 0 {
		return ""
	}
	var sections []string
	if m.filterShown() {
		sections = append(sections, m.input.View())
	}
	if m.visLen() == 0 {
		if m.search {
			sections = append(sections, m.Styles.Dim.Render("(no match)"))
		}
		return strings.Join(sections, "\n")
	}

	start, end := m.VisibleRange()
	bar := m.scrollbarColumn()

	contentWidth := m.width
	if bar != nil {
		contentWidth-- // reserve a column for the scrollbar
	}
	var rowStyle lg.Style
	if contentWidth > 0 {
		rowStyle = lg.NewStyle().Width(contentWidth).MaxWidth(contentWidth).Inline(true)
	}

	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		row := m.rowAt(i)
		if i == m.cursor {
			// Drop per-cell colors so the highlight spans the whole row
			// instead of being cut short by an inner reset.
			row = sgrPattern.ReplaceAllString(row, "")
		}
		if contentWidth > 0 {
			row = rowStyle.Render(row)
		}
		if i == m.cursor {
			row = m.Styles.Cursor.Render(row)
		}
		if bar != nil {
			row = lg.JoinHorizontal(lg.Top, row, bar[i-start])
		}
		lines = append(lines, row)
	}
	sections = append(sections, strings.Join(lines, "\n"))
	return strings.Join(sections, "\n")
}

// refilter recomputes matches for the current filter text and snaps the
// cursor to the first match - a narrowed list under a stale cursor would
// submit something the user isn't looking at.
//
// Matching is a case-insensitive substring test against each row's visible
// (ANSI-stripped) text, keeping a case-blind contract: an uppercased query
// still matches a lowercase row. Fuzzy subsequence matching is deliberately
// not used - it broadens matches unhelpfully over rows that embed structured
// text.
func (m *Model) refilter() {
	q := strings.TrimSpace(m.input.Value())
	// A fresh slice, not a truncation: two models sharing a backing array
	// would stomp each other's matches.
	m.matches = nil
	term := filter.Term{Text: q, Case: filter.CaseInsensitive}
	for i, row := range m.rows {
		if q == "" || term.Match(xansi.Strip(row)) {
			m.matches = append(m.matches, i)
		}
	}
	m.cursor = 0
	m.clamp()
}

// visLen returns the number of selectable rows: the match count when
// searching, the full row count otherwise.
func (m *Model) visLen() int {
	if m.search {
		return len(m.matches)
	}
	return len(m.rows)
}

// origIndex maps a visible-list position to its original row index.
func (m *Model) origIndex(pos int) int {
	if m.search {
		return m.matches[pos]
	}
	return pos
}

// rowAt returns the rendered row at a visible-list position.
func (m *Model) rowAt(pos int) string { return m.rows[m.origIndex(pos)] }

// filterShown reports whether the filter line occupies a row of the view.
func (m *Model) filterShown() bool { return m.search && m.input.Value() != "" }

// rowsHeight returns the rows' share of the view height: the filter line,
// when showing, takes one row off the top.
func (m *Model) rowsHeight() int {
	if m.filterShown() {
		return m.height - 1
	}
	return m.height
}

func (m *Model) pageStride() int {
	if h := m.rowsHeight(); h > 1 {
		return h - 1 // keep one row of context across pages
	}
	return 1
}

// clamp keeps cursor in range and adjusts offset so the cursor is visible and
// the window never scrolls past the end. This single method is the only place
// that mutates offset, so the visible-window invariant holds everywhere.
func (m *Model) clamp() {
	n := m.visLen()
	if n == 0 {
		m.cursor, m.offset = 0, 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= n {
		m.cursor = n - 1
	}
	h := m.rowsHeight()
	if h <= 0 {
		m.offset = 0
		return
	}
	// Pull the window to the cursor.
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+h {
		m.offset = m.cursor - h + 1
	}
	// Never leave a gap at the bottom when enough rows exist to fill the view.
	maxOffset := max(0, n-h)
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

// scrollbarColumn returns one glyph per visible row, or nil when no scrollbar
// is needed (disabled, or the list fits). The thumb position uses the
// normalized scroll fraction (0 at the top, 1 at the bottom) - not the
// less-style percentage, which is anchored to the bottom of the viewport and
// would place the thumb at the end while the list is still at the top.
func (m *Model) scrollbarColumn() []string {
	n, h := m.visLen(), m.rowsHeight()
	if !m.Scrollbar || n <= h || h <= 0 {
		return nil
	}
	maxOffset := n - h
	fraction := float64(m.offset) / float64(maxOffset)
	thumbPos, thumbSize := scrollbar.ThumbMetrics(h, n, fraction)
	thumbEnd := thumbPos + thumbSize - 1
	col := make([]string, h)
	for i := range col {
		if i >= thumbPos && i <= thumbEnd {
			col[i] = "█"
		} else {
			col[i] = "│"
		}
	}
	return col
}

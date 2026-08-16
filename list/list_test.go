package list_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/list"
	"github.com/stretchr/testify/require"
)

func rows(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("row-%d", i)
	}
	return out
}

func press(m *list.Model, text string) {
	m.Update(tea.KeyPressMsg{Text: text})
}

// invariant asserts the core property: the cursor is always within the visible
// window and the window never scrolls past the end.
func invariant(t *testing.T, m *list.Model, total, height int) {
	t.Helper()
	if total == 0 {
		return
	}
	start, end := m.VisibleRange()
	cursor, ok := m.Selected()
	require.True(t, ok)
	require.GreaterOrEqual(t, cursor, start, "cursor above window")
	require.Less(t, cursor, end, "cursor below window")
	require.GreaterOrEqual(t, start, 0, "negative offset")
	if height > 0 && total-height >= 0 {
		require.LessOrEqual(t, start, total-height, "offset past max")
	}
}

// visibleLabels walks the cursor through the visible (filtered) list and
// collects the original rows it lands on - a readable stand-in for the
// unexported matches slice.
func visibleLabels(m *list.Model, all []string) []string {
	var out []string
	seen := make(map[int]bool)
	m.Top()
	for {
		idx, ok := m.Selected()
		if !ok || seen[idx] {
			break
		}
		seen[idx] = true
		out = append(out, all[idx])
		m.MoveDown(1)
	}
	return out
}

func TestEmptyListCursorIsNegative(t *testing.T) {
	t.Parallel()

	m := list.New()
	m.SetSize(10, 5)
	require.Equal(t, -1, m.Cursor())
	require.Empty(t, m.View())
}

func TestCursorMovesWindowToStayVisible(t *testing.T) {
	t.Parallel()

	m := list.New()
	m.SetSize(20, 5)
	m.SetRows(rows(100))

	m.Bottom()
	require.Equal(t, 99, m.Cursor())
	start, end := m.VisibleRange()
	require.Equal(t, 95, start)
	require.Equal(t, 100, end)
	invariant(t, m, 100, 5)

	m.Top()
	start, end = m.VisibleRange()
	require.Equal(t, 0, start)
	require.Equal(t, 5, end)
	invariant(t, m, 100, 5)
}

func TestScrollingHappensOnlyAtWindowEdge(t *testing.T) {
	t.Parallel()

	m := list.New()
	m.SetSize(20, 5)
	m.SetRows(rows(100))

	// Moving within the window does not scroll.
	m.MoveDown(4) // cursor 4, still in [0,5)
	start, _ := m.VisibleRange()
	require.Equal(t, 0, start, "offset moved early")

	// One more step pushes the window by one.
	m.MoveDown(1) // cursor 5 → window [1,6)
	start, _ = m.VisibleRange()
	require.Equal(t, 1, start)
	invariant(t, m, 100, 5)
}

func TestSetRowsShrinkClampsCursor(t *testing.T) {
	t.Parallel()

	m := list.New()
	m.SetSize(20, 5)
	m.SetRows(rows(50))
	m.Bottom() // cursor 49
	m.SetRows(rows(3))
	require.Equal(t, 2, m.Cursor())
	start, _ := m.VisibleRange()
	require.Equal(t, 0, start)
	invariant(t, m, 3, 5)
}

func TestShortListNeverScrolls(t *testing.T) {
	t.Parallel()

	m := list.New()
	m.SetSize(20, 10)
	m.SetRows(rows(3))
	m.Bottom()
	start, end := m.VisibleRange()
	require.Equal(t, 0, start)
	require.Equal(t, 3, end)
}

func TestMoveClampsAtBounds(t *testing.T) {
	t.Parallel()

	m := list.New()
	m.SetSize(20, 5)
	m.SetRows(rows(10))
	m.MoveUp(100)
	require.Equal(t, 0, m.Cursor())
	m.MoveDown(100)
	require.Equal(t, 9, m.Cursor())
	invariant(t, m, 10, 5)
}

func TestPageMovesByHeightLessOne(t *testing.T) {
	t.Parallel()

	m := list.New()
	m.SetSize(20, 5)
	m.SetRows(rows(100))
	m.PageDown()
	require.Equal(t, 4, m.Cursor(), "page stride should be height-1")
	invariant(t, m, 100, 5)
}

func TestResizeKeepsSelection(t *testing.T) {
	t.Parallel()

	m := list.New()
	m.SetSize(20, 10)
	m.SetRows(rows(100))
	m.SetCursor(50)
	m.SetSize(20, 4) // shrink viewport
	require.Equal(t, 50, m.Cursor(), "resize lost selection")
	invariant(t, m, 100, 4)
}

func TestViewHighlightsCursorRow(t *testing.T) {
	t.Parallel()

	m := list.New()
	m.SetSize(0, 5) // zero width leaves rows unpadded for exact comparison
	m.SetRows([]string{"a", "b"})
	want := lg.NewStyle().Reverse(true).Render("a") + "\nb"
	require.Equal(t, want, m.View())
}

func TestViewStripsInnerSGRFromCursorRow(t *testing.T) {
	t.Parallel()

	m := list.New()
	m.SetSize(0, 5)
	// The inner \x1b[m reset would cut a naive highlight short partway
	// across the row; the row must be stripped to plain text first.
	m.SetRows([]string{"\x1b[31mred\x1b[m rest", "plain"})
	want := lg.NewStyle().Reverse(true).Render("red rest") + "\nplain"
	require.Equal(t, want, m.View())
}

func TestViewScrollbarThumbTracksPosition(t *testing.T) {
	t.Parallel()

	m := list.New()
	m.SetSize(10, 10)
	m.SetRows(rows(100))

	m.Top()
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	require.Len(t, lines, 10)
	require.True(
		t,
		strings.HasSuffix(lines[0], "█"),
		"thumb not at top when scrolled to top: %q",
		lines[0],
	)

	m.Bottom()
	lines = strings.Split(ansi.Strip(m.View()), "\n")
	require.True(
		t,
		strings.HasSuffix(lines[9], "█"),
		"thumb not at bottom when scrolled to bottom: %q",
		lines[9],
	)
}

func TestSearchSelectedDefaultsToFirstRow(t *testing.T) {
	t.Parallel()

	m := list.New(list.WithSearch())
	m.SetSize(0, 10)
	m.SetRows([]string{"To Do", "In Progress", "Done"})
	idx, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, 0, idx)
}

func TestSearchUpDownKeysNavigate(t *testing.T) {
	t.Parallel()

	m := list.New(list.WithSearch())
	m.SetSize(0, 10)
	m.SetRows([]string{"a", "b", "c"})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	idx, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, 2, idx)

	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	idx, ok = m.Selected()
	require.True(t, ok)
	require.Equal(t, 1, idx)
}

func TestSearchTypingFiltersRows(t *testing.T) {
	t.Parallel()

	all := []string{"To Do", "In Progress", "Done"}
	m := list.New(list.WithSearch())
	m.SetSize(0, 10)
	m.SetRows(all)
	press(m, "pro")

	idx, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, 1, idx, "filter 'pro' should select In Progress")
	require.Equal(t, []string{"In Progress"}, visibleLabels(m, all))
}

func TestSearchMatchesVisibleTextNotStyling(t *testing.T) {
	t.Parallel()

	// The row's escape codes must not defeat (or satisfy) the filter: the
	// substring test runs against the ANSI-stripped text.
	all := []string{"\x1b[31mDone\x1b[m", "\x1b[32mOpen\x1b[m"}
	m := list.New(list.WithSearch())
	m.SetSize(0, 10)
	m.SetRows(all)
	press(m, "done")

	idx, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, 0, idx)
	press(m, "31m") // an escape-sequence fragment must match nothing
	_, ok = m.Selected()
	require.False(t, ok)
}

func TestSearchFilterIsCaseInsensitive(t *testing.T) {
	t.Parallel()

	m := list.New(list.WithSearch())
	m.SetSize(0, 10)
	m.SetRows([]string{"Done"})
	press(m, "DONE")
	_, ok := m.Selected()
	require.True(t, ok, "case-insensitive match failed")
}

func TestSearchNoMatchMeansNoSelection(t *testing.T) {
	t.Parallel()

	m := list.New(list.WithSearch())
	m.SetSize(0, 10)
	m.SetRows([]string{"a", "b"})
	press(m, "zzz")
	_, ok := m.Selected()
	require.False(t, ok)
	require.Equal(t, -1, m.Cursor())

	lines := strings.Split(ansi.Strip(m.View()), "\n")
	require.Equal(t, "(no match)", lines[len(lines)-1])
}

func TestSearchFilterChangeResetsCursor(t *testing.T) {
	t.Parallel()

	all := []string{"alpha", "beta", "gamma"}
	m := list.New(list.WithSearch())
	m.SetSize(0, 10)
	m.SetRows(all)
	m.MoveDown(2) // cursor on gamma
	press(m, "a") // all three match; cursor must snap back to the first
	idx, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, 0, idx, "cursor should reset to first match")
}

func TestSearchBackspaceWidensFilter(t *testing.T) {
	t.Parallel()

	all := []string{"Done", "Do not"}
	m := list.New(list.WithSearch())
	m.SetSize(0, 10)
	m.SetRows(all)
	press(m, "done")
	m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	// "don" matches only "Done"; one more backspace → "do" matches both.
	m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	require.Equal(t, []string{"Done", "Do not"}, visibleLabels(m, all))
}

func TestSearchEmptyListNeverSelects(t *testing.T) {
	t.Parallel()

	m := list.New(list.WithSearch())
	m.SetSize(0, 10)
	_, ok := m.Selected()
	require.False(t, ok)
	m.MoveDown(1) // must not panic
	_ = m.View()
}

func TestSearchPasteFiltersToo(t *testing.T) {
	t.Parallel()

	m := list.New(list.WithSearch())
	m.SetSize(0, 10)
	m.SetRows([]string{"alpha", "beta"})
	m.Update(tea.PasteMsg{Content: "bet"})
	idx, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, 1, idx)
}

func TestSearchFilterLineTakesOneRowOfHeight(t *testing.T) {
	t.Parallel()

	m := list.New(list.WithSearch())
	m.SetSize(0, 3)
	m.SetRows(rows(10))

	// No filter typed: the rows own the full height.
	_, end := m.VisibleRange()
	require.Equal(t, 3, end)

	// A visible filter line leaves height-1 rows on screen.
	press(m, "row")
	start, end := m.VisibleRange()
	require.Equal(t, 0, start)
	require.Equal(t, 2, end)
}

func TestCursorLine(t *testing.T) {
	t.Parallel()

	m := list.New(list.WithSearch())
	m.SetSize(0, 10)
	require.Equal(t, -1, m.CursorLine(), "empty list has no cursor line")

	m.SetRows([]string{"alpha", "beta", "gamma"})
	require.Equal(t, 0, m.CursorLine())
	m.MoveDown(2)
	require.Equal(t, 2, m.CursorLine())

	// A visible filter line shifts every row down by one.
	press(m, "a")
	require.Equal(t, 1, m.CursorLine(), "filter line should count as a line above the rows")
}

// TestSearchSubstringMatching pins the matching contract: a contiguous,
// case-insensitive substring of the visible text - not a fuzzy subsequence,
// which broadens matches unhelpfully over rows that embed structured text.
func TestSearchSubstringMatching(t *testing.T) {
	t.Parallel()

	all := []string{"To Do", "In Progress", "In Review", "Done"}
	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{"contiguous substring matches", "prog", []string{"In Progress"}},
		{"substring spanning both rows", "in", []string{"In Progress", "In Review"}},
		{"uppercase query stays case-insensitive", "DONE", []string{"Done"}},
		{"non-contiguous subsequence does not match", "ipr", nil},
		{"absent substring matches nothing", "zzz", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := list.New(list.WithSearch())
			m.SetSize(0, 10)
			m.SetRows(all)
			press(m, tt.query)
			require.Equal(t, tt.want, visibleLabels(m, all))
		})
	}
}

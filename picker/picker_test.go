package picker_test

import (
	"strings"
	"testing"

	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/picker"
	"github.com/stretchr/testify/require"
)

func testRows() []picker.Row {
	return []picker.Row{
		{Label: "Color", Choices: []string{"red", "green", "blue"}},
		{Label: "Size", Choices: []string{"small", "medium", "large"}},
	}
}

func testStyles() picker.Styles {
	return picker.Styles{
		Cursor:         "> ",
		CursorPad:      "  ",
		Label:          lg.NewStyle(),
		LockedLabel:    lg.NewStyle(),
		Selected:       lg.NewStyle(),
		Default:        lg.NewStyle(),
		Inactive:       lg.NewStyle(),
		LockedInactive: lg.NewStyle(),
		HelpKey:        lg.NewStyle(),
		HelpText:       lg.NewStyle(),
		Box:            lg.NewStyle(),
	}
}

func TestUpDown(t *testing.T) {
	m := picker.New(
		testRows(),
		[]int{0, 0}, []int{0, 0},
		[]bool{false, false}, []bool{false, false},
		testStyles(),
	)

	require.Equal(t, 0, m.Cursor)

	m.Down()
	require.Equal(t, 1, m.Cursor)

	m.Down()
	require.Equal(t, 1, m.Cursor) // clamped

	m.Up()
	require.Equal(t, 0, m.Cursor)

	m.Up()
	require.Equal(t, 0, m.Cursor) // clamped
}

func TestLeftRight(t *testing.T) {
	m := picker.New(
		testRows(),
		[]int{0, 0}, []int{0, 0},
		[]bool{false, false}, []bool{false, false},
		testStyles(),
	)

	m.Right()
	require.Equal(t, 1, m.Values[0])
	require.False(t, m.IsReset[0])

	m.Right()
	require.Equal(t, 2, m.Values[0])

	m.Right()
	require.Equal(t, 2, m.Values[0]) // clamped

	m.Left()
	require.Equal(t, 1, m.Values[0])

	m.Left()
	m.Left()
	require.Equal(t, 0, m.Values[0]) // clamped
}

func TestCycleWraps(t *testing.T) {
	m := picker.New(
		testRows(),
		[]int{2, 0}, []int{0, 0},
		[]bool{false, false}, []bool{false, false},
		testStyles(),
	)

	m.Cycle()
	require.Equal(t, 0, m.Values[0]) // wrapped
}

func TestReset(t *testing.T) {
	m := picker.New(
		testRows(),
		[]int{2, 1}, []int{0, 0},
		[]bool{false, false}, []bool{false, false},
		testStyles(),
	)

	m.Reset()
	require.Equal(t, 0, m.Values[0])
	require.True(t, m.IsReset[0])
}

func TestLockedRowIgnoresInput(t *testing.T) {
	m := picker.New(
		testRows(),
		[]int{0, 0}, []int{0, 0},
		[]bool{true, false}, []bool{false, false},
		testStyles(),
	)

	m.Right()
	require.Equal(t, 0, m.Values[0]) // unchanged

	m.Cycle()
	require.Equal(t, 0, m.Values[0]) // unchanged

	m.Reset()
	require.False(t, m.IsReset[0]) // unchanged
}

func TestViewContainsChoices(t *testing.T) {
	m := picker.New(
		testRows(),
		[]int{0, 1}, []int{0, 0},
		[]bool{false, false}, []bool{false, false},
		testStyles(),
	)

	view := ansi.Strip(m.View())
	require.Equal(t, "> Color  red  green  blue\n   Size  small  medium  large\n\n", view)
}

func TestViewContainsCursor(t *testing.T) {
	m := picker.New(
		testRows(),
		[]int{0, 0}, []int{0, 0},
		[]bool{false, false}, []bool{false, false},
		testStyles(),
	)

	view := ansi.Strip(m.View())
	lines := strings.Split(view, "\n")
	require.True(t, strings.HasPrefix(lines[0], "> "))
	require.True(t, strings.HasPrefix(lines[1], "  "))
}

func TestViewContainsLockedSuffix(t *testing.T) {
	s := testStyles()
	s.LockedSuffix = "  (CLI)"
	m := picker.New(
		testRows(),
		[]int{0, 0}, []int{0, 0},
		[]bool{true, false}, []bool{false, false},
		s,
	)

	view := ansi.Strip(m.View())
	require.Equal(t, "> Color  red  green  blue  (CLI)\n   Size  small  medium  large\n\n", view)
}

func TestViewContainsHints(t *testing.T) {
	m := picker.New(
		testRows(),
		[]int{0, 0}, []int{0, 0},
		[]bool{false, false}, []bool{false, false},
		testStyles(),
	)
	m.Hints = []picker.HelpHint{
		{Key: "enter", Desc: "apply"},
		{Key: "esc", Desc: "cancel"},
	}

	view := ansi.Strip(m.View())
	require.Equal(
		t,
		"> Color  red  green  blue\n   Size  small  medium  large\n\nenter apply  esc cancel",
		view,
	)
}

func TestViewCursorLineBG(t *testing.T) {
	s := testStyles()
	s.CursorLineBG = "\x1b[48;2;40;10;30m"
	m := picker.New(
		testRows(),
		[]int{0, 0}, []int{0, 0},
		[]bool{false, false}, []bool{false, false},
		s,
	)

	view := m.View()
	require.Equal(
		t,
		"\x1b[48;2;40;10;30m> Color  red  green  blue    \x1b[0m\n   Size  small  medium  large\n\n",
		view,
	)
}

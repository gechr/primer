package helpsheet_test

import (
	"strings"
	"testing"

	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/helpsheet"
	"github.com/stretchr/testify/require"
)

func TestRenderEmptyPairs(t *testing.T) {
	got := helpsheet.Model{}.Render()
	require.Empty(t, got)
}

func TestRenderTwoColumnLayout(t *testing.T) {
	got := helpsheet.Model{
		Pairs: []helpsheet.Pair{
			{Key: "j/k", Desc: "navigate"},
			{Key: "enter", Desc: "select"},
			{Key: "q", Desc: "quit"},
		},
		Styles: helpsheet.Styles{
			Key:  lg.NewStyle(),
			Text: lg.NewStyle(),
			Box:  lg.NewStyle(),
		},
		Gutter: 4,
	}.Render()

	lines := strings.Split(ansi.Strip(got), "\n")
	// 3 pairs -> 2 rows (ceil(3/2)).
	require.GreaterOrEqual(t, len(lines), 2)

	// First row has both columns: "j/k  navigate" + gutter + "q  quit".
	require.Equal(t, "  j/k  navigate        q  quit", lines[0])

	// Second row has only left column.
	require.Equal(t, "enter  select", lines[1])
}

func TestRenderDismissFooter(t *testing.T) {
	got := helpsheet.Model{
		Pairs: []helpsheet.Pair{
			{Key: "a", Desc: "alpha"},
			{Key: "b", Desc: "beta"},
		},
		Dismiss: "Press any key",
		Styles: helpsheet.Styles{
			Key:     lg.NewStyle(),
			Text:    lg.NewStyle(),
			Dismiss: lg.NewStyle(),
			Box:     lg.NewStyle(),
		},
	}.Render()

	require.Equal(t, "a  alpha    b  beta\n\n   Press any key", ansi.Strip(got))
}

func TestRenderOddPairCount(t *testing.T) {
	got := helpsheet.Model{
		Pairs: []helpsheet.Pair{
			{Key: "a", Desc: "one"},
		},
		Styles: helpsheet.Styles{
			Key:  lg.NewStyle(),
			Text: lg.NewStyle(),
			Box:  lg.NewStyle(),
		},
	}.Render()

	lines := strings.Split(ansi.Strip(got), "\n")
	// Single pair -> 1 row.
	require.Equal(t, "a  one", lines[0])
}

func TestRenderKeyRightAlignment(t *testing.T) {
	got := helpsheet.Model{
		Pairs: []helpsheet.Pair{
			{Key: "j", Desc: "down"},
			{Key: "enter", Desc: "select"},
		},
		Styles: helpsheet.Styles{
			Key:  lg.NewStyle(),
			Text: lg.NewStyle(),
			Box:  lg.NewStyle(),
		},
	}.Render()

	stripped := ansi.Strip(got)
	lines := strings.Split(stripped, "\n")
	// "j" should be right-aligned to match "enter" width - padded with spaces.
	require.True(t, strings.HasPrefix(strings.TrimLeft(lines[0], " "), " ") ||
		strings.Contains(lines[0], "    j"),
		"expected right-aligned key, got: %q", lines[0])
}

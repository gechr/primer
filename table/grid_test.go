package table_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/table"
	"github.com/stretchr/testify/require"
)

func TestNewGridAppliesDefaultsAndOptions(t *testing.T) {
	t.Parallel()

	defaultGrid := table.NewGrid(nil)
	require.Equal(t, 2, defaultGrid.ColumnPadding)
	require.Equal(t, table.PaddingLeft, defaultGrid.Padding)
	require.Equal(t, -1, defaultGrid.FlexCol)

	configured := table.NewGrid(
		[][]string{{"a"}},
		table.WithColumnPadding(4),
		table.WithPadding(table.PaddingRight),
	)
	require.Equal(t, 4, configured.ColumnPadding)
	require.Equal(t, table.PaddingRight, configured.Padding)
	require.Equal(t, -1, configured.FlexCol)
	require.Equal(t, [][]string{{"a"}}, configured.Rows)
}

func TestVisibleWidthIgnoresANSI(t *testing.T) {
	t.Parallel()

	require.Equal(t, 2, table.VisibleWidth("\x1b[31mgo\x1b[0m"))
}

func TestAlignColumnsPaddingModes(t *testing.T) {
	t.Parallel()

	rows := [][]string{
		{"a", "bb"},
		{"ccc", "d"},
	}

	cases := []struct {
		name           string
		padding        table.Padding
		expectedFirst  string
		expectedSecond string
	}{
		{
			name:           "left",
			padding:        table.PaddingLeft,
			expectedFirst:  "a    bb",
			expectedSecond: "ccc  d",
		},
		{
			name:           "right",
			padding:        table.PaddingRight,
			expectedFirst:  "  a  bb",
			expectedSecond: "ccc   d",
		},
		{
			name:           "center",
			padding:        table.PaddingCenter,
			expectedFirst:  " a   bb",
			expectedSecond: "ccc  d",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			grid := table.NewGrid(rows, table.WithPadding(tc.padding))
			aligned, widths := grid.AlignColumns()

			require.Equal(t, []int{3, 2}, widths)
			require.Equal(t, tc.expectedFirst, aligned[0])
			require.Equal(t, tc.expectedSecond, aligned[1])
		})
	}
}

func TestAlignColumnsTruncatesFlexColumnAndWrapsTTYSpaces(t *testing.T) {
	t.Parallel()

	grid := table.NewGrid([][]string{
		{"1", "abcdef"},
		{"2", "xy"},
	})
	grid.FlexCol = 1
	grid.MaxWidth = 8
	grid.TTY = true

	aligned, widths := grid.AlignColumns()

	require.Equal(t, []int{1, 5}, widths)
	require.Contains(t, aligned[0], "\x1b[8m")
	require.Contains(t, aligned[0], "\x1b[28m")
	require.Equal(t, 7, table.VisibleWidth(aligned[0]))
	require.NotEqual(t, "abcdef", grid.Rows[0][1])
	require.Contains(t, grid.Rows[0][1], "…")
	require.True(t, strings.HasPrefix(ansi.Strip(aligned[1]), "2"))
}

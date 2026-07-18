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
	require.Nil(t, defaultGrid.FlexCols)

	configured := table.NewGrid(
		[][]string{{"a"}},
		table.WithColumnPadding(4),
		table.WithPadding(table.PaddingRight),
	)
	require.Equal(t, 4, configured.ColumnPadding)
	require.Equal(t, table.PaddingRight, configured.Padding)
	require.Nil(t, configured.FlexCols)
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

func TestAlignColumnsWidthMethod(t *testing.T) {
	t.Parallel()

	// ZWJ (👨‍💻) and VS16 (⚠️) emoji sequences: wcwidth and grapheme
	// clustering disagree on their width, so the two rows only align under
	// the method that matches the measurement.
	rows := func() [][]string {
		return [][]string{
			{"👨‍💻 x ⚠️", "end"},
			{"plain ascii", "end"},
		}
	}

	t.Run("grapheme", func(t *testing.T) {
		t.Parallel()

		grid := table.NewGrid(rows(), table.WithWidthMethod(ansi.GraphemeWidth))
		aligned, _ := grid.AlignColumns()

		require.Equal(t,
			ansi.GraphemeWidth.StringWidth(aligned[0]),
			ansi.GraphemeWidth.StringWidth(aligned[1]),
		)
	})

	t.Run("default wcwidth", func(t *testing.T) {
		t.Parallel()

		grid := table.NewGrid(rows())
		aligned, _ := grid.AlignColumns()

		require.Equal(t,
			ansi.WcWidth.StringWidth(aligned[0]),
			ansi.WcWidth.StringWidth(aligned[1]),
		)
	})
}

func TestAlignColumnsTruncatesFlexColumnsWithWidthMethod(t *testing.T) {
	t.Parallel()

	grid := table.NewGrid([][]string{
		{"1", "👨‍💻👨‍💻👨‍💻"},
		{"2", "xy"},
	}, table.WithWidthMethod(ansi.GraphemeWidth), table.WithFlexColumns(1))
	grid.MaxWidth = 8

	aligned, widths := grid.AlignColumns()

	require.Equal(t, []int{1, 5}, widths)
	require.Equal(t, "👨‍💻👨‍💻…", grid.Rows[0][1])
	require.LessOrEqual(t, ansi.GraphemeWidth.StringWidth(aligned[0]), 8)
}

func TestAlignColumnsTruncatesFlexColumnsAndWrapsTTYSpaces(t *testing.T) {
	t.Parallel()

	grid := table.NewGrid([][]string{
		{"1", "abcdef"},
		{"2", "xy"},
	})
	grid.FlexCols = []int{1}
	grid.MaxWidth = 8
	grid.TTY = true

	aligned, widths := grid.AlignColumns()

	require.Equal(t, []int{1, 5}, widths)
	require.Equal(t, "1\x1b[8m  \x1b[28mabcd…", aligned[0])
	require.Equal(t, 8, table.VisibleWidth(aligned[0]))
	require.Equal(t, "abcd…", grid.Rows[0][1])
	require.True(t, strings.HasPrefix(ansi.Strip(aligned[1]), "2"))
}

func TestAlignColumnsShrinksMultipleFlexColumns(t *testing.T) {
	t.Parallel()

	grid := table.NewGrid([][]string{
		{"ID", "Alpha", "Beta", "End"},
		{"1", "abcdefghijkl", "wxyz", "ok"},
	}, table.WithFlexColumns(1, 2))
	grid.MaxWidth = 18

	aligned, widths := grid.AlignColumns()

	require.Equal(t, []int{2, 3, 4, 3}, widths)
	require.LessOrEqual(t, table.VisibleWidth(aligned[1]), 18)
	require.Equal(t, "ab…", ansi.Strip(grid.Rows[1][1]))
	require.Equal(t, "wxyz", ansi.Strip(grid.Rows[1][2]))
}

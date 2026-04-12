package table_test

import (
	"testing"
	"time"

	"github.com/gechr/primer/table"
	"github.com/stretchr/testify/require"
)

func TestSortRowsStringAscendingStableAndNilLast(t *testing.T) {
	t.Parallel()

	columns := []table.Column[int]{
		{Name: "name"},
		{Name: "score"},
	}
	rows := []table.Row[int]{
		{
			Item: 1,
			Cells: []table.Cell{
				table.TextCell("bravo"),
				table.SortableCell("10", "10", 10),
			},
		},
		{
			Item: 2,
			Cells: []table.Cell{
				table.TextCell("alpha"),
				table.SortableCell("10", "10", 10),
			},
		},
		{
			Item: 3,
			Cells: []table.Cell{
				table.TextCell("charlie"),
				table.DisplayOnly("na", "na"),
			},
		},
	}

	got := table.SortRows(rows, columns, "score", true)

	require.Equal(t, []int{1, 2, 3}, []int{got[0].Item, got[1].Item, got[2].Item})
	require.Equal(t, rows[0].Item, got[0].Item)
	require.Equal(t, rows[1].Item, got[1].Item)
	require.Equal(t, rows[2].Item, got[2].Item)
}

func TestSortRowsStringDescending(t *testing.T) {
	t.Parallel()

	columns := []table.Column[int]{
		{Name: "name"},
	}
	rows := []table.Row[int]{
		{Item: 1, Cells: []table.Cell{table.TextCell("alpha")}},
		{Item: 2, Cells: []table.Cell{table.TextCell("charlie")}},
		{Item: 3, Cells: []table.Cell{table.TextCell("bravo")}},
	}

	got := table.SortRows(rows, columns, "name", false)

	require.Equal(t, []int{2, 3, 1}, []int{got[0].Item, got[1].Item, got[2].Item})
}

func TestSortRowsTimeAscending(t *testing.T) {
	t.Parallel()

	columns := []table.Column[int]{{Name: "when"}}
	earlier := time.Date(2026, 4, 12, 9, 0, 0, 0, time.UTC)
	later := earlier.Add(time.Hour)
	rows := []table.Row[int]{
		{Item: 1, Cells: []table.Cell{table.TimeCell("later", later)}},
		{Item: 2, Cells: []table.Cell{table.TimeCell("earlier", earlier)}},
	}

	got := table.SortRows(rows, columns, "when", true)

	require.Equal(t, []int{2, 1}, []int{got[0].Item, got[1].Item})
}

func TestSortRowsIntAscending(t *testing.T) {
	t.Parallel()

	columns := []table.Column[int]{{Name: "rank"}}
	rows := []table.Row[int]{
		{Item: 1, Cells: []table.Cell{table.SortableCell("2", "2", 2)}},
		{Item: 2, Cells: []table.Cell{table.SortableCell("1", "1", 1)}},
	}

	got := table.SortRows(rows, columns, "rank", true)

	require.Equal(t, []int{2, 1}, []int{got[0].Item, got[1].Item})
}

func TestSortRowsUnknownColumnReturnsInput(t *testing.T) {
	t.Parallel()

	rows := []table.Row[int]{{Item: 1}}
	columns := []table.Column[int]{{Name: "name"}}

	got := table.SortRows(rows, columns, "missing", true)

	require.Same(t, &rows[0], &got[0])
}

package table_test

import (
	"testing"

	"github.com/gechr/primer/table"
	"github.com/stretchr/testify/require"
)

func TestRenderedTableStringEmptyRows(t *testing.T) {
	t.Parallel()

	got := table.RenderedTable[int]{Header: "header"}.String()

	require.Empty(t, got)
}

func TestRenderedTableStringWithHeaderAndRows(t *testing.T) {
	t.Parallel()

	got := table.RenderedTable[int]{
		Header: "header",
		Rows: []table.Row[int]{
			{Display: "row 1"},
			{Display: "row 2"},
		},
	}.String()

	require.Equal(t, "header\nrow 1\nrow 2", got)
}

func TestRenderedTableStringWithoutHeader(t *testing.T) {
	t.Parallel()

	got := table.RenderedTable[int]{
		Rows: []table.Row[int]{
			{Display: "row 1"},
			{Display: "row 2"},
		},
	}.String()

	require.Equal(t, "row 1\nrow 2", got)
}

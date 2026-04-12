package table_test

import (
	"testing"
	"time"

	"github.com/gechr/primer/table"
	"github.com/stretchr/testify/require"
)

func TestTextCell(t *testing.T) {
	t.Parallel()

	got := table.TextCell("hello")

	require.Equal(t, "hello", got.Text)
	require.Equal(t, "hello", got.Plain)
	require.Equal(t, "hello", got.SortKey)
}

func TestStyledCell(t *testing.T) {
	t.Parallel()

	got := table.StyledCell("\x1b[31mhello\x1b[0m", "hello")

	require.Equal(t, "\x1b[31mhello\x1b[0m", got.Text)
	require.Equal(t, "hello", got.Plain)
	require.Equal(t, "hello", got.SortKey)
}

func TestTimeCell(t *testing.T) {
	t.Parallel()

	when := time.Date(2026, 4, 12, 14, 30, 0, 0, time.UTC)
	got := table.TimeCell("2h ago", when)

	require.Equal(t, "2h ago", got.Text)
	require.Equal(t, "2h ago", got.Plain)
	require.Equal(t, when, got.SortKey)
}

func TestSortableCell(t *testing.T) {
	t.Parallel()

	got := table.SortableCell("display", "plain", 42)

	require.Equal(t, "display", got.Text)
	require.Equal(t, "plain", got.Plain)
	require.Equal(t, 42, got.SortKey)
}

func TestDisplayOnly(t *testing.T) {
	t.Parallel()

	got := table.DisplayOnly("display", "plain")

	require.Equal(t, "display", got.Text)
	require.Equal(t, "plain", got.Plain)
	require.Nil(t, got.SortKey)
}

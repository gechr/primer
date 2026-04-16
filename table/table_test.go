package table_test

import (
	"image/color"
	"strconv"
	"strings"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/table"
	"github.com/gechr/x/ansi"
	"github.com/stretchr/testify/require"
)

type testTheme struct{}

func (testTheme) RenderBold(s string) string { return "\x1b[1m" + s + "\x1b[0m" }

func (testTheme) RenderDim(s string) string { return "\x1b[2m" + s + "\x1b[0m" }

func (testTheme) EntityColors() []color.Color { return nil }

type panicBoldTheme struct{}

func (panicBoldTheme) RenderBold(string) string { panic("RenderBold should not be called") }

func (panicBoldTheme) RenderDim(s string) string { return s }

func (panicBoldTheme) EntityColors() []color.Color { return nil }

type record struct {
	ID   int
	Name string
}

func newContext(theme table.Theme) *table.RenderContext {
	return table.NewRenderContext(
		theme,
		ansi.New(ansi.WithHyperlinkFallback(ansi.HyperlinkFallbackText)),
	)
}

func TestNewRendererAndRenderHeaderOnlyUsesCustomHeaderRenderer(t *testing.T) {
	t.Parallel()

	columns := []table.Column[record]{
		{
			Name:   "id",
			Header: "ID",
			Render: func(_ record, _ *table.RenderContext) table.Cell { return table.DisplayOnly("", "") },
		},
		{
			Name:   "description",
			Header: "Description",
			Render: func(_ record, _ *table.RenderContext) table.Cell { return table.DisplayOnly("", "") },
			Flex:   true,
		},
	}
	renderer := table.NewRenderer(
		columns,
		newContext(panicBoldTheme{}),
		table.WithHeaderRenderer(func(_ string, header string, _ *table.RenderContext) string {
			return strings.ToUpper(header)
		}),
	)

	require.Len(t, renderer.Columns(), 2)

	header, widths := renderer.RenderHeaderOnly([]int{4, 2})

	require.Equal(t, "ID    DESCRIPTION", xansi.Strip(header))
	require.Equal(t, []int{4, 11}, widths)
}

func TestRenderFormatsRowsWithReverseIndex(t *testing.T) {
	t.Parallel()

	columns := []table.Column[record]{
		{
			Name:   "id",
			Header: "ID",
			Render: func(r record, _ *table.RenderContext) table.Cell { return table.TextCell(strconv.Itoa(r.ID)) },
		},
		{
			Name:   "name",
			Header: "Name",
			Render: func(r record, _ *table.RenderContext) table.Cell { return table.TextCell(r.Name) },
		},
	}
	renderer := table.NewRenderer(
		columns,
		newContext(testTheme{}),
		table.WithShowIndex(true),
		table.WithReverse(true),
	)

	got := renderer.Render([]record{
		{ID: 1, Name: "alpha"},
		{ID: 2, Name: "beta"},
	})

	require.Len(t, got.Rows, 2)
	require.Equal(t, 2, got.Rows[0].Item.ID)
	require.Equal(t, 1, got.Rows[1].Item.ID)
	require.True(t, strings.HasPrefix(xansi.Strip(got.Rows[0].Display), "1"))
	require.True(t, strings.HasPrefix(xansi.Strip(got.Rows[1].Display), "2"))
	require.Equal(t, "   ID  Name", xansi.Strip(got.Header))
	require.Len(t, got.ColWidths, 3)
}

func TestRenderTruncatesFlexColumnToTermWidth(t *testing.T) {
	t.Parallel()

	columns := []table.Column[record]{
		{
			Name:   "id",
			Header: "ID",
			Render: func(r record, _ *table.RenderContext) table.Cell { return table.TextCell(strconv.Itoa(r.ID)) },
		},
		{
			Name:   "name",
			Header: "Name",
			Flex:   true,
			Render: func(r record, _ *table.RenderContext) table.Cell { return table.TextCell(r.Name) },
		},
	}
	renderer := table.NewRenderer(columns, newContext(testTheme{}), table.WithTermWidth(8))

	got := renderer.Render([]record{{ID: 1, Name: "abcdef"}})

	require.Equal(t, []int{2, 4}, got.ColWidths)
	require.Equal(t, "1   ab…", xansi.Strip(got.Rows[0].Display))
	require.Equal(t, 7, xansi.WcWidth.StringWidth(xansi.Strip(got.Rows[0].Display)))
}

func TestRenderReturnsEmptyTableForEmptyInputs(t *testing.T) {
	t.Parallel()

	emptyColumns := table.NewRenderer([]table.Column[record]{}, newContext(testTheme{}))
	require.Equal(t, table.RenderedTable[record]{}, emptyColumns.Render([]record{{ID: 1}}))

	columns := []table.Column[record]{
		{
			Name:   "id",
			Header: "ID",
			Render: func(r record, _ *table.RenderContext) table.Cell { return table.TextCell(strconv.Itoa(r.ID)) },
		},
	}
	renderer := table.NewRenderer(columns, newContext(testTheme{}))
	require.Equal(t, table.RenderedTable[record]{}, renderer.Render(nil))
}

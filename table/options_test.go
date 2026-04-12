package table_test

import (
	"image/color"
	"strings"
	"testing"

	"github.com/gechr/primer/table"
	"github.com/stretchr/testify/require"
)

type optionTheme struct{}

func (optionTheme) RenderBold(s string) string { return "<b>" + s + "</b>" }

func (optionTheme) RenderDim(s string) string { return "<d>" + s + "</d>" }

func (optionTheme) EntityColors() []color.Color { return nil }

func TestGridOptions(t *testing.T) {
	t.Parallel()

	g := table.NewGrid(nil)

	table.WithColumnPadding(4)(g)
	table.WithPadding(table.PaddingRight)(g)

	require.Equal(t, 4, g.ColumnPadding)
	require.Equal(t, table.PaddingRight, g.Padding)
}

func TestRendererOptions(t *testing.T) {
	t.Parallel()

	var gotName, gotHeader string
	var gotCtx *table.RenderContext
	fn := func(name, header string, ctx *table.RenderContext) string {
		gotName = name
		gotHeader = header
		gotCtx = ctx
		return header
	}

	renderer := table.NewRenderer(
		[]table.Column[int]{
			{
				Name:   "name",
				Header: "Name",
				Render: func(item int, _ *table.RenderContext) table.Cell {
					return table.TextCell(strings.Repeat("x", item))
				},
			},
		},
		table.NewRenderContext(optionTheme{}, nil),
		table.WithHeaderRenderer(fn),
		table.WithReverse(true),
		table.WithShowIndex(true),
		table.WithTTY(true),
		table.WithTermWidth(80),
	)

	require.Len(t, renderer.Columns(), 1)
	require.Equal(t, "name", renderer.Columns()[0].Name)
	require.Equal(t, "Name", renderer.Columns()[0].Header)
	got := renderer.Render([]int{1, 2})

	require.Contains(t, got.Header, "\x1b[8m")
	require.Len(t, got.Rows, 2)
	require.Equal(t, 2, got.Rows[0].Item)
	require.Equal(t, 1, got.Rows[1].Item)
	require.Equal(t, "name", gotName)
	require.Equal(t, "Name", gotHeader)
	require.NotNil(t, gotCtx)
}

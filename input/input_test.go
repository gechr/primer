package input_test

import (
	"testing"

	"github.com/gechr/primer/input"
	"github.com/stretchr/testify/require"
)

func TestNewTextAreaDefaultsAndOptions(t *testing.T) {
	t.Parallel()

	ta := input.NewTextArea()
	require.Empty(t, ta.Prompt)
	require.Equal(t, "Enter text...", ta.Placeholder)
	require.False(t, ta.ShowLineNumbers)
	require.True(t, ta.DynamicHeight)
	require.Equal(t, 3, ta.MinHeight)
	require.Equal(t, 10, ta.MaxHeight)
	require.Equal(t, 80, ta.Width())

	ta = input.NewTextArea(
		input.WithPlaceholder("Write a body"),
		input.WithWidth(42),
		input.WithMinHeight(5),
		input.WithMaxHeight(11),
	)

	require.Empty(t, ta.Prompt)
	require.Equal(t, "Write a body", ta.Placeholder)
	require.False(t, ta.ShowLineNumbers)
	require.True(t, ta.DynamicHeight)
	require.Equal(t, 5, ta.MinHeight)
	require.Equal(t, 11, ta.MaxHeight)
	require.Equal(t, 42, ta.Width())
}

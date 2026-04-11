package key_test

import (
	"strings"
	"testing"

	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/key"
	"github.com/stretchr/testify/require"
)

func TestInlineSingleLetter(t *testing.T) {
	rendered, ok := key.Inline("a", "approve", lg.NewStyle(), lg.NewStyle())

	require.True(t, ok)
	require.Equal(t, "approve", ansi.Strip(rendered))
}

func TestInlineModifiedKey(t *testing.T) {
	rendered, ok := key.Inline("alt+c", "copy", lg.NewStyle(), lg.NewStyle())

	require.True(t, ok)
	require.Equal(t, "alt+copy", ansi.Strip(rendered))
}

func TestRendererWrapsAtWidth(t *testing.T) {
	r := key.Renderer{
		Styles: key.Styles{
			Key:  lg.NewStyle(),
			Text: lg.NewStyle(),
		},
		Width:  12,
		Inline: true,
	}

	got := r.Render([]key.Hint{
		{Key: "a", Desc: "approve"},
		{Key: "c", Desc: "comment"},
	})

	lines := strings.Split(ansi.Strip(got), "\n")
	require.Len(t, lines, 2)
}

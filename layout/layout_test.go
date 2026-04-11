package layout_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/layout"
	"github.com/stretchr/testify/require"
)

func TestNormalizeLineExpandsTabs(t *testing.T) {
	got := layout.NormalizeLine("left\tright", 14)

	require.NotContains(t, got, "\t")
	require.Equal(t, 14, ansi.WcWidth.StringWidth(got))
	require.Contains(t, got, "left    right")
}

func TestNormalizeLineTruncatesToWidth(t *testing.T) {
	got := layout.NormalizeLine("abcdefgh", 4)

	require.Equal(t, "abcd", got)
}

func TestFillPadsToHeight(t *testing.T) {
	got := layout.Fill("left\tright", 12, 2)

	require.NotContains(t, got, "\t")
	lines := strings.Split(got, "\n")
	require.Len(t, lines, 2)
	require.Contains(t, lines[0], "left    right")
	require.Equal(t, 12, ansi.WcWidth.StringWidth(lines[1]))
}

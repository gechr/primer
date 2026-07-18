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

	require.Equal(t, "left    right ", got)
	require.Equal(t, 14, ansi.WcWidth.StringWidth(got))
}

func TestNormalizeLineTruncatesToWidth(t *testing.T) {
	got := layout.NormalizeLine("abcdefgh", 4)

	require.Equal(t, "abcd", got)
}

func TestNormalizeLines(t *testing.T) {
	got := layout.NormalizeLines([]string{"ab", "cdef"}, 4)

	require.Len(t, got, 2)
	require.Equal(t, 4, ansi.WcWidth.StringWidth(got[0]))
	require.Equal(t, 4, ansi.WcWidth.StringWidth(got[1]))
}

func TestNormalizeLinesNil(t *testing.T) {
	require.Nil(t, layout.NormalizeLines(nil, 10))
}

func TestWrapLines(t *testing.T) {
	got := layout.WrapLines("abcdef\nghij", 4)

	require.Len(t, got, 3)
	require.Equal(t, "abcd", got[0])
}

func TestWrapLinesExpandsTabs(t *testing.T) {
	got := layout.WrapLines("\thello", 0)

	require.Len(t, got, 1)
	require.Equal(t, "    hello", got[0])
}

func TestHardWrapNoOp(t *testing.T) {
	got := layout.HardWrap("short", 80)

	require.Len(t, got, 1)
	require.Equal(t, "short", got[0])
}

func TestHardWrapSplits(t *testing.T) {
	got := layout.HardWrap("abcdefgh", 4)

	require.Greater(t, len(got), 1)
}

func TestSeparatorPlain(t *testing.T) {
	got := layout.Separator(5, -1)

	require.Equal(t, "─────", got)
}

func TestSeparatorWithJunction(t *testing.T) {
	got := layout.Separator(5, 2)

	require.Equal(t, "──┬──", got)
}

func TestExpandTabs(t *testing.T) {
	require.Equal(t, "    x", layout.ExpandTabs("\tx"))
}

func TestPreserveBackgroundPreservesVisibleText(t *testing.T) {
	line := "\x1b[32mhello\x1b[0m"

	got := layout.PreserveBackground(line, "\x1b[41m")

	require.Equal(t, "hello", ansi.Strip(got))
	require.True(t, strings.HasPrefix(got, "\x1b[41m"))
	require.True(t, strings.HasSuffix(got, "\x1b[0m"))
}

func TestPreserveBackgroundReappliesAfterSGR(t *testing.T) {
	line := "\x1b[32mhello\x1b[0m"

	got := layout.PreserveBackground(line, "\x1b[41m")

	require.GreaterOrEqual(t, strings.Count(got, "\x1b[41m"), 2)
}

func TestPreserveBackgroundWidthPadsVisibleWidth(t *testing.T) {
	got := layout.PreserveBackgroundWidth("abc", "\x1b[41m", 5)

	require.Equal(t, 5, ansi.WcWidth.StringWidth(ansi.Strip(got)))
	require.True(t, strings.HasSuffix(got, "\x1b[0m"))
}

func TestFillPadsToHeight(t *testing.T) {
	got := layout.Fill("left\tright", 12, 2)

	lines := strings.Split(got, "\n")
	require.Len(t, lines, 2)
	require.Equal(t, "left    right", lines[0])
	require.Equal(t, 12, ansi.WcWidth.StringWidth(lines[1]))
}

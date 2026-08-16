package titlebox_test

import (
	"strings"
	"testing"

	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/titlebox"
	"github.com/stretchr/testify/require"
)

// lines strips styling and splits a rendered box into rows for structural
// assertions.
func lines(s string) []string { return strings.Split(ansi.Strip(s), "\n") }

func TestRenderEmbedsTitleInTopBorder(t *testing.T) {
	t.Parallel()

	got := lines(titlebox.Render("summary", "hi", 20, titlebox.Styles{}))
	require.Len(t, got, 3, "a single-line body should box to 3 rows")
	require.True(
		t,
		strings.HasPrefix(got[0], "╭ summary "),
		"top border missing embedded title: %q",
		got[0],
	)
	require.True(t, strings.HasSuffix(got[0], "╮"), "top border malformed: %q", got[0])
	require.Equal(t, "│ hi", got[1][:len("│ hi")])
	require.True(t, strings.HasSuffix(got[1], "│"), "body row not framed: %q", got[1])
	require.True(t, strings.HasPrefix(got[2], "╰"), "bottom border malformed: %q", got[2])
	require.True(t, strings.HasSuffix(got[2], "╯"), "bottom border malformed: %q", got[2])
}

func TestRenderPadsEveryRowToWidth(t *testing.T) {
	t.Parallel()

	const width = 24
	for _, row := range lines(titlebox.Render("t", "a\nbb\nccc", width, titlebox.Styles{})) {
		require.Equal(t, width, lg.Width(row), "row %q", row)
	}
}

func TestRenderBlankTitleDrawsPlainEdge(t *testing.T) {
	t.Parallel()

	top := lines(titlebox.Render("", "x", 12, titlebox.Styles{}))[0]
	require.Equal(t, "╭──────────╮", top)
}

func TestRenderOverwideTitleFallsBackToPlainEdge(t *testing.T) {
	t.Parallel()

	// A title wider than the interior cannot embed; the box keeps its frame
	// rather than overflowing the width.
	top := lines(titlebox.Render("a very long field label", "x", 10, titlebox.Styles{}))[0]
	require.Equal(t, 10, lg.Width(top))
	require.Equal(t, "╭────────╮", top)
}

func TestRenderMultilineBodyBoxesEachLine(t *testing.T) {
	t.Parallel()

	got := lines(titlebox.Render("desc", "one\ntwo\nthree", 20, titlebox.Styles{}))
	require.Len(t, got, 5) // top + 3 body + bottom
	require.Equal(t, "│ one              │", got[1])
	require.Equal(t, "│ two              │", got[2])
	require.Equal(t, "│ three            │", got[3])
}

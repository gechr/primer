package pill_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/pill"
	"github.com/stretchr/testify/require"
)

func TestRenderFramesValueWithChevrons(t *testing.T) {
	t.Parallel()

	got := ansi.Strip(pill.Render("type", "Task", 10, pill.Styles{}))
	require.Equal(t, "type      ‹ Task ›", got)
}

func TestRenderPadsLabelToColumn(t *testing.T) {
	t.Parallel()

	// A short label is padded so its value starts at the same column as a
	// longer label's would.
	short := ansi.Strip(pill.Render("t", "A", 10, pill.Styles{}))
	long := ansi.Strip(pill.Render("project", "A", 10, pill.Styles{}))
	require.Equal(
		t,
		strings.Index(long, "‹"),
		strings.Index(short, "‹"),
		"values not column-aligned",
	)
}

func TestRenderLeavesOverlongLabelUnpadded(t *testing.T) {
	t.Parallel()

	// A label already at the column width is not truncated or re-padded.
	got := ansi.Strip(pill.Render("a-very-long-label", "V", 4, pill.Styles{}))
	require.Equal(t, "a-very-long-label‹ V ›", got)
}

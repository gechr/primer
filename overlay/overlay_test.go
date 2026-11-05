package overlay_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/overlay"
	"github.com/stretchr/testify/require"
)

func TestPlaceCentersForeground(t *testing.T) {
	background := strings.Join([]string{
		"........",
		"........",
		"........",
		"........",
	}, "\n")

	got := overlay.Place(background, "XX", 8, 4, overlay.Center)

	lines := strings.Split(ansi.Strip(got), "\n")
	require.Len(t, lines, 4)
	require.Equal(t, "...XX...", lines[1])
}

func TestPlacePadsShortBackground(t *testing.T) {
	got := overlay.Place("", "X", 4, 3, overlay.Center)

	lines := strings.Split(ansi.Strip(got), "\n")
	require.Len(t, lines, 3)
	require.Equal(t, " X", strings.TrimRight(lines[1], " "))
}

func TestOriginMatchesPlace(t *testing.T) {
	t.Parallel()

	// "XX" on an 8x4 screen centers at column 3, row 1 - the same cell
	// TestPlaceCenters observes in the composited output.
	x, y := overlay.Origin("XX", 8, 4, overlay.Center)
	require.Equal(t, 3, x)
	require.Equal(t, 1, y)

	// An oversized foreground clamps to the top-left rather than going
	// negative.
	x, y = overlay.Origin("WIDE\nLONG", 2, 1, overlay.Center)
	require.Equal(t, 0, x)
	require.Equal(t, 0, y)
}

// Contains must claim exactly the cells Place draws the foreground over.
func TestContainsMatchesPlace(t *testing.T) {
	t.Parallel()

	const width, height = 9, 6
	fg := "abcd\nwxyz"
	bg := strings.TrimSuffix(strings.Repeat(strings.Repeat(".", width)+"\n", height), "\n")
	lines := strings.Split(ansi.Strip(overlay.Place(bg, fg, width, height, overlay.Center)), "\n")
	for y, line := range lines {
		for x := range width {
			drawn := ansi.Cut(line, x, x+1) != "."
			inside := overlay.Contains(fg, width, height, overlay.Center, x, y)
			require.Equal(t, drawn, inside, "x=%d y=%d", x, y)
		}
	}
	require.False(t, overlay.Contains(fg, width, height, overlay.Center, -1, 0))
}

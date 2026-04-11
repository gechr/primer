package bar_test

import (
	"strings"
	"testing"

	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/scroll/bar"
	"github.com/stretchr/testify/require"
)

func TestPercent(t *testing.T) {
	// At the top: offset 0, viewport 10 of 100 lines = 10%.
	require.Equal(t, 10, bar.Percent(0, 100, 10))

	// At the end: offset 90, viewport 10 of 100 = 100%.
	require.Equal(t, 100, bar.Percent(90, 100, 10))

	// Midway: offset 45, viewport 10 of 100 = 55%.
	require.Equal(t, 55, bar.Percent(45, 100, 10))

	// Empty content: always 100%.
	require.Equal(t, 100, bar.Percent(0, 0, 10))

	// Content fits in viewport: 100%.
	require.Equal(t, 100, bar.Percent(0, 5, 10))
}

func TestPosition(t *testing.T) {
	require.Equal(t, "1-10/42 (23%)", bar.Position(0, 10, 42))
	require.Equal(t, "33-42/42 (100%)", bar.Position(32, 42, 42))
	require.Equal(t, "11-20/100 (20%)", bar.Position(10, 20, 100))
}

func TestThumbMetrics(t *testing.T) {
	pos, size := bar.ThumbMetrics(10, 40, 0.5)

	require.Positive(t, size)
	require.GreaterOrEqual(t, pos, 0)
	require.Less(t, pos, 10)
}

func TestRender(t *testing.T) {
	m := bar.Model{
		Height:     4,
		TotalLines: 20,
		Percent:    0.5,
		Styles: bar.Styles{
			Thumb: lg.NewStyle(),
			Track: lg.NewStyle(),
		},
	}

	got := m.Render()
	lines := strings.Split(ansi.Strip(got), "\n")

	require.Len(t, lines, 4)
}

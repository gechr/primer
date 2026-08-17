package scrollbar_test

import (
	"strings"
	"testing"

	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/scrollbar"
	"github.com/stretchr/testify/require"
)

func TestPercent(t *testing.T) {
	// At the top: offset 0, viewport 10 of 100 lines = 10%.
	require.Equal(t, 10, scrollbar.Percent(0, 100, 10))

	// At the end: offset 90, viewport 10 of 100 = 100%.
	require.Equal(t, 100, scrollbar.Percent(90, 100, 10))

	// Midway: offset 45, viewport 10 of 100 = 55%.
	require.Equal(t, 55, scrollbar.Percent(45, 100, 10))

	// Empty content: always 100%.
	require.Equal(t, 100, scrollbar.Percent(0, 0, 10))

	// Content fits in viewport: 100%.
	require.Equal(t, 100, scrollbar.Percent(0, 5, 10))
}

func TestPosition(t *testing.T) {
	require.Equal(t, "1-10/42 (23%)", scrollbar.Position(0, 10, 42))
	require.Equal(t, "33-42/42 (100%)", scrollbar.Position(32, 42, 42))
	require.Equal(t, "11-20/100 (20%)", scrollbar.Position(10, 20, 100))
}

func TestThumbMetrics(t *testing.T) {
	pos, size := scrollbar.ThumbMetrics(10, 40, 0.5)

	require.Positive(t, size)
	require.GreaterOrEqual(t, pos, 0)
	require.Less(t, pos, 10)
}

func TestThumbMetricsKeepsDefaultMaxThumbCap(t *testing.T) {
	_, size := scrollbar.ThumbMetrics(10, 11, 0)

	require.Equal(t, 5, size)
}

func TestThumbMetricsWithConfigAllowsProportionalThumb(t *testing.T) {
	_, size := scrollbar.ThumbMetricsWithConfig(10, 11, 0, scrollbar.Config{
		MaxThumbDivisor: 1,
	})

	require.Equal(t, 9, size)
}

func TestThumbMetricsWithConfigUsesMinimumThumbSize(t *testing.T) {
	_, size := scrollbar.ThumbMetricsWithConfig(10, 1000, 0, scrollbar.Config{
		MinThumbSize: 3,
	})

	require.Equal(t, 3, size)
}

func TestThumbMetricsWithConfigClampsMinimumToMaximum(t *testing.T) {
	_, size := scrollbar.ThumbMetricsWithConfig(3, 1000, 0, scrollbar.Config{
		MinThumbSize: 3,
	})

	require.Equal(t, 1, size)
}

func TestThumbMetricsAlwaysFitsTrack(t *testing.T) {
	pos, size := scrollbar.ThumbMetrics(1, 100, 2)

	require.Zero(t, pos)
	require.Equal(t, 1, size)
}

func TestThumbMetricsClampsPercent(t *testing.T) {
	top, _ := scrollbar.ThumbMetrics(10, 40, -1)
	bottom, size := scrollbar.ThumbMetrics(10, 40, 2)

	require.Zero(t, top)
	require.Equal(t, 10-size, bottom)
}

func TestViewportThumbMetrics(t *testing.T) {
	top, size := scrollbar.ViewportThumbMetrics(20, 100, 10, 0)
	middle, middleSize := scrollbar.ViewportThumbMetrics(20, 100, 10, 45)
	bottom, bottomSize := scrollbar.ViewportThumbMetrics(20, 100, 10, 90)

	require.Equal(t, 2, size)
	require.Equal(t, size, middleSize)
	require.Equal(t, size, bottomSize)
	require.Zero(t, top)
	require.Equal(t, 9, middle)
	require.Equal(t, 18, bottom)
}

func TestViewportThumbMetricsUsesIndependentTrackHeight(t *testing.T) {
	pos, size := scrollbar.ViewportThumbMetricsWithConfig(5, 100, 20, 80, scrollbar.Config{
		MaxThumbDivisor: 1,
	})

	require.Equal(t, 4, pos)
	require.Equal(t, 1, size)
}

func TestViewportThumbMetricsClampsOffset(t *testing.T) {
	top, _ := scrollbar.ViewportThumbMetrics(10, 100, 10, -10)
	bottom, size := scrollbar.ViewportThumbMetrics(10, 100, 10, 1000)

	require.Zero(t, top)
	require.Equal(t, 10-size, bottom)
}

func TestViewportThumbMetricsEmptyContent(t *testing.T) {
	pos, size := scrollbar.ViewportThumbMetrics(10, 0, 10, 0)

	require.Zero(t, pos)
	require.Zero(t, size)
}

func TestRender(t *testing.T) {
	m := scrollbar.Model{
		Height:     4,
		TotalLines: 20,
		Percent:    0.5,
		Styles: scrollbar.Styles{
			Thumb: lg.NewStyle(),
			Track: lg.NewStyle(),
		},
	}

	got := m.Render()
	lines := strings.Split(ansi.Strip(got), "\n")

	require.Len(t, lines, 4)
}

func TestRenderUsesConfiguredSymbols(t *testing.T) {
	m := scrollbar.Model{
		Height:     3,
		TotalLines: 9,
		Config: scrollbar.Config{
			ThumbSymbol:     "#",
			TrackSymbol:     ".",
			MaxThumbDivisor: 1,
		},
		Styles: scrollbar.Styles{
			Thumb: lg.NewStyle(),
			Track: lg.NewStyle(),
		},
	}

	require.Equal(t, "#\n.\n.", ansi.Strip(m.Render()))
}

func TestRenderCanHideTrack(t *testing.T) {
	m := scrollbar.Model{
		Height:     3,
		TotalLines: 9,
		Config: scrollbar.Config{
			HideTrack:       true,
			ThumbSymbol:     "#",
			MaxThumbDivisor: 1,
		},
		Styles: scrollbar.Styles{
			Thumb: lg.NewStyle(),
			Track: lg.NewStyle(),
		},
	}

	require.Equal(t, "#\n \n ", ansi.Strip(m.Render()))
}

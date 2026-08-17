package scrollbar

import (
	"fmt"
	"math"
	"strings"

	lg "charm.land/lipgloss/v2"
	xmath "github.com/gechr/x/math"
)

const (
	nl              = "\n"
	defaultThumb    = "█"
	defaultTrack    = "┃"
	maxThumbDivisor = 2
)

// Config controls scrollbar rendering, geometry, and interaction.
//
// Zero-valued fields preserve the package defaults.
type Config struct {
	// ThumbSymbol is the cell rendered for the scrollbar thumb.
	ThumbSymbol string
	// TrackSymbol is the cell rendered for the scrollbar track.
	TrackSymbol string
	// HideTrack renders blank cells outside the thumb while preserving the
	// scrollbar's width.
	HideTrack bool
	// PageTrackClicks makes presses outside the thumb move one viewport instead
	// of jumping to the pressed position and beginning a drag.
	PageTrackClicks bool
	// MaxThumbDivisor caps the thumb height to Height/MaxThumbDivisor.
	// The default is 2. Set to 1 for a fully proportional thumb capped only
	// by the track height.
	MaxThumbDivisor int
	// MinThumbSize sets the minimum thumb height. The default is 1. Values
	// larger than the configured maximum are clamped to that maximum.
	MinThumbSize int
}

type Styles struct {
	Thumb lg.Style
	Track lg.Style
}

type Model struct {
	Config     Config
	Height     int
	TotalLines int
	Percent    float64
	Styles     Styles
}

// Chars returns the scrollbar as one rendered cell per line.
func (m Model) Chars() []string {
	if m.Height <= 0 {
		return nil
	}
	cfg := m.Config.withDefaults()
	thumbPos, thumbSize := ThumbMetricsWithConfig(m.Height, m.TotalLines, m.Percent, cfg)
	chars := make([]string, m.Height)
	for i := range m.Height {
		if i >= thumbPos && i < thumbPos+thumbSize {
			chars[i] = m.Styles.Thumb.Render(cfg.ThumbSymbol)
			continue
		}
		if cfg.HideTrack {
			chars[i] = " "
			continue
		}
		chars[i] = m.Styles.Track.Render(cfg.TrackSymbol)
	}
	return chars
}

// Render returns the scrollbar as a multi-line string.
func (m Model) Render() string {
	return strings.Join(m.Chars(), nl)
}

// Percent returns the scroll position as a percentage in the style of less(1):
// the percentage of content above and including the bottom of the viewport.
// This means the value never reaches 0% when there is content, and reaches
// 100% exactly at the end.
func Percent(offset, total, viewport int) int {
	const percentMax = 100
	if total <= 0 {
		return percentMax
	}
	return min(percentMax*(offset+viewport)/total, percentMax)
}

// Position returns a formatted scroll position string like "1-10/42 (24%)".
// The start parameter is 0-indexed; it is displayed as 1-indexed in the output.
func Position(start, end, total int) string {
	pct := Percent(start, total, end-start)
	return fmt.Sprintf("%d-%d/%d (%d%%)", start+1, end, total, pct)
}

// ThumbMetrics returns the thumb position and size for a proportional scrollbar.
func ThumbMetrics(height, totalLines int, percent float64) (int, int) {
	return ThumbMetricsWithConfig(height, totalLines, percent, Config{})
}

// ThumbMetricsWithConfig returns the thumb position and size for a proportional
// scrollbar using the supplied config.
func ThumbMetricsWithConfig(height, totalLines int, percent float64, cfg Config) (int, int) {
	if height <= 0 {
		return 0, 0
	}
	cfg = cfg.withDefaults()
	thumbSize := configuredThumbSize(height, height*height/max(1, totalLines), cfg)
	trackSpace := max(0, height-thumbSize)
	thumbPos := 0
	if trackSpace > 0 {
		percent = xmath.Clamp01(percent)
		thumbPos = int(math.Round(percent * float64(trackSpace)))
	}
	return thumbPos, thumbSize
}

// ViewportThumbMetrics returns the thumb position and size for an explicit
// viewport and scroll offset. The track height is independent of the number of
// content lines visible in the viewport.
func ViewportThumbMetrics(
	trackHeight, totalLines, viewportLines, offset int,
) (int, int) {
	return ViewportThumbMetricsWithConfig(
		trackHeight, totalLines, viewportLines, offset, Config{},
	)
}

// ViewportThumbMetricsWithConfig returns the thumb position and size for an
// explicit viewport and scroll offset using the supplied config.
func ViewportThumbMetricsWithConfig(
	trackHeight, totalLines, viewportLines, offset int,
	cfg Config,
) (int, int) {
	if trackHeight <= 0 || totalLines <= 0 {
		return 0, 0
	}
	cfg = cfg.withDefaults()
	viewportLines = max(0, viewportLines)
	visibleLines := min(viewportLines, totalLines)
	proportionalSize := roundedRatio(visibleLines*trackHeight, totalLines)
	thumbSize := configuredThumbSize(trackHeight, proportionalSize, cfg)

	maxOffset := max(0, totalLines-viewportLines)
	if maxOffset == 0 {
		return 0, thumbSize
	}
	offset = xmath.Clamp(offset, 0, maxOffset)
	thumbPos := roundedRatio(offset*(trackHeight-thumbSize), maxOffset)
	return thumbPos, thumbSize
}

// ThumbRange returns the start row, inclusive, and end row, exclusive, of the
// rendered thumb within a track of the given height.
func ThumbRange(height, totalLines int, percent float64, cfg Config) (int, int) {
	pos, size := ThumbMetricsWithConfig(height, totalLines, percent, cfg)
	return pos, pos + size
}

func (c Config) withDefaults() Config {
	if c.ThumbSymbol == "" {
		c.ThumbSymbol = defaultThumb
	}
	if c.TrackSymbol == "" {
		c.TrackSymbol = defaultTrack
	}
	if c.MaxThumbDivisor <= 0 {
		c.MaxThumbDivisor = maxThumbDivisor
	}
	if c.MinThumbSize <= 0 {
		c.MinThumbSize = 1
	}
	return c
}

func configuredThumbSize(height, proportionalSize int, cfg Config) int {
	maxThumb := max(1, height/cfg.MaxThumbDivisor)
	minThumb := min(cfg.MinThumbSize, maxThumb)
	return xmath.Clamp(proportionalSize, minThumb, maxThumb)
}

func roundedRatio(numerator, denominator int) int {
	if denominator <= 0 {
		return 0
	}
	return (numerator + denominator/2) / denominator
}

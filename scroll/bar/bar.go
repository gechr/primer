package bar

import (
	"fmt"
	"math"
	"strings"

	lg "charm.land/lipgloss/v2"
)

const (
	nl              = "\n"
	maxThumbDivisor = 2
)

type Styles struct {
	Thumb lg.Style
	Track lg.Style
}

type Model struct {
	Height     int
	TotalLines int
	Percent    float64
	Styles     Styles
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
	if height <= 0 {
		return 0, 0
	}
	maxThumb := height / maxThumbDivisor
	thumbSize := min(maxThumb, max(1, height*height/max(1, totalLines)))
	trackSpace := max(0, height-thumbSize)
	thumbPos := 0
	if trackSpace > 0 {
		thumbPos = int(math.Round(percent * float64(trackSpace)))
	}
	return thumbPos, thumbSize
}

// Chars returns the scrollbar as one rendered cell per line.
func (m Model) Chars() []string {
	if m.Height <= 0 {
		return nil
	}
	thumbPos, thumbSize := ThumbMetrics(m.Height, m.TotalLines, m.Percent)
	chars := make([]string, m.Height)
	for i := range m.Height {
		if i >= thumbPos && i < thumbPos+thumbSize {
			chars[i] = m.Styles.Thumb.Render("█")
			continue
		}
		chars[i] = m.Styles.Track.Render("┃")
	}
	return chars
}

// Render returns the scrollbar as a multi-line string.
func (m Model) Render() string {
	return strings.Join(m.Chars(), nl)
}

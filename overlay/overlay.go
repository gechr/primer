package overlay

import (
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
)

const (
	nl            = "\n"
	centerDivisor = 2
)

type Placement int

const (
	Center Placement = iota
)

// Place overlays foreground content on top of background content.
func Place(background, foreground string, width, height int, placement Placement) string {
	switch placement {
	case Center:
		return center(background, foreground, width, height)
	default:
		return center(background, foreground, width, height)
	}
}

// Origin returns the screen cell (column, row) where Place puts foreground's
// top-left corner on a width-by-height screen. It shares Place's placement
// math, so a caller translating mouse coordinates into overlay space can
// never drift from where the overlay was actually drawn.
func Origin(foreground string, width, height int, _ Placement) (int, int) {
	fgWidth, fgHeight := size(foreground)
	startRow := max(0, (height-fgHeight)/centerDivisor)
	startCol := max(0, (width-fgWidth)/centerDivisor)
	return startCol, startRow
}

// Contains reports whether screen cell (x, y) falls inside the rectangle Place
// covers with foreground, so a caller can stop clicks on the overlay from
// reaching whatever it hides.
func Contains(foreground string, width, height int, placement Placement, x, y int) bool {
	startCol, startRow := Origin(foreground, width, height, placement)
	fgWidth, fgHeight := size(foreground)
	return x >= startCol && x < startCol+fgWidth && y >= startRow && y < startRow+fgHeight
}

// size returns foreground's width (its widest line) and height in cells.
func size(foreground string) (int, int) {
	fgLines := strings.Split(foreground, nl)
	fgWidth := 0
	for _, line := range fgLines {
		if w := xansi.WcWidth.StringWidth(line); w > fgWidth {
			fgWidth = w
		}
	}
	return fgWidth, len(fgLines)
}

func center(background, foreground string, width, height int) string {
	bgLines := strings.Split(background, nl)
	fgLines := strings.Split(foreground, nl)

	strWidth := xansi.WcWidth.StringWidth

	fgWidth := 0
	for _, line := range fgLines {
		if w := strWidth(line); w > fgWidth {
			fgWidth = w
		}
	}

	startCol, startRow := Origin(foreground, width, height, Center)

	for len(bgLines) < height {
		bgLines = append(bgLines, "")
	}

	for i, fgLine := range fgLines {
		row := startRow + i
		if row >= len(bgLines) {
			break
		}
		bgLine := bgLines[row]
		bgVisible := strWidth(bgLine)
		if bgVisible < startCol {
			bgLine += strings.Repeat(" ", startCol-bgVisible)
		}
		left := xansi.TruncateWc(bgLine, startCol, "")
		if gap := startCol - strWidth(left); gap > 0 {
			left += strings.Repeat(" ", gap)
		}
		rightStart := startCol + fgWidth
		var right string
		if bgVisible > rightStart {
			right = xansi.TruncateLeftWc(bgLine, rightStart, "")
		}
		bgLines[row] = left + "\033[0m" + fgLine + right
	}

	return strings.Join(bgLines, nl)
}

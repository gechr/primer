package layout

import (
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
)

const nl = "\n"

// NormalizeLine expands tabs, clamps visual width, and pads the line to width.
func NormalizeLine(line string, width int) string {
	line = expandTabs(line)
	if width <= 0 {
		return line
	}

	lineWidth := xansi.WcWidth.StringWidth(line)
	if lineWidth > width {
		line = xansi.WcWidth.Truncate(line, width, "")
		lineWidth = width
	}
	if pad := width - lineWidth; pad > 0 {
		line += strings.Repeat(" ", pad)
	}
	return line
}

// Fill expands tabs and pads the output with blank rows to fill the terminal.
func Fill(output string, width, height int) string {
	output = expandTabs(output)
	if width <= 0 || height <= 0 {
		return output
	}

	lines := strings.Split(output, nl)
	blank := strings.Repeat(" ", width)
	for len(lines) < height {
		lines = append(lines, blank)
	}
	return strings.Join(lines, nl)
}

func expandTabs(s string) string {
	return strings.ReplaceAll(s, "\t", "    ")
}

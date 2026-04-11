package layout

import (
	"bytes"
	"strings"

	lg "charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

const nl = "\n"

// NormalizeLine expands tabs, clamps visual width, and pads the line to width.
func NormalizeLine(line string, width int) string {
	line = ExpandTabs(line)
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
	output = ExpandTabs(output)
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

// NormalizeLines applies [NormalizeLine] to each line.
func NormalizeLines(lines []string, width int) []string {
	if len(lines) == 0 {
		return nil
	}
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = NormalizeLine(line, width)
	}
	return out
}

// WrapLines splits text on newlines, expands tabs, and hard-wraps each line
// to width. Returns the resulting rows.
func WrapLines(text string, width int) []string {
	logicalLines := strings.Split(ExpandTabs(text), nl)
	if width <= 0 {
		return logicalLines
	}
	rows := make([]string, 0, len(logicalLines))
	for _, line := range logicalLines {
		rows = append(rows, HardWrap(line, width)...)
	}
	return rows
}

// HardWrap splits a single line into multiple rows at width, preserving ANSI
// escape sequences across the break. Returns the original line in a
// single-element slice when no wrapping is needed.
func HardWrap(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}
	wrapped := xansi.Hardwrap(line, width, true)
	if !strings.Contains(wrapped, nl) {
		return []string{line}
	}
	var buf bytes.Buffer
	writer := lg.NewWrapWriter(&buf)
	_, _ = writer.Write([]byte(wrapped))
	_ = writer.Close()
	return strings.Split(buf.String(), nl)
}

const (
	sepHorizontal = "─"
	sepJunction   = "┬"
)

// Separator returns a horizontal rule of width characters. If junctionCol is
// non-negative and within width, a ┬ junction is placed at that column.
func Separator(width, junctionCol int) string {
	if junctionCol >= 0 && junctionCol < width {
		return strings.Repeat(sepHorizontal, junctionCol) +
			sepJunction +
			strings.Repeat(sepHorizontal, width-junctionCol-1)
	}
	return strings.Repeat(sepHorizontal, width)
}

// ExpandTabs replaces tab characters with four spaces.
func ExpandTabs(s string) string {
	return strings.ReplaceAll(s, "\t", "    ")
}

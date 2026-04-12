package view

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	lg "charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/layout"
	"github.com/gechr/primer/scrollbar"
)

const nl = "\n"

const footerEllipsis = "…"

type FrameStyles struct {
	Separator lg.Style
	Scrollbar scrollbar.Styles
}

type FooterAlign string

const (
	FooterAlignLeft  FooterAlign = "left"
	FooterAlignRight FooterAlign = "right"
)

type FooterComponent struct {
	Align       FooterAlign
	BreakBefore bool
	Content     string
}

type FrameModel struct {
	Footer    []FooterComponent
	Header    string
	Height    int
	Lines     []string
	Separator string
	Styles    FrameStyles
	View      viewport.Model
	Width     int
}

func (m FrameModel) separator() string {
	if m.Separator != "" {
		return m.Separator
	}
	return "─"
}

// Sync normalizes lines to width, sets viewport dimensions, loads the
// content, and enables fill-height. Returns the normalized lines for rendering.
func Sync(vp *viewport.Model, lines []string, width, height int) []string {
	renderLines := layout.NormalizeLines(lines, width)
	vp.SetWidth(width)
	vp.SetHeight(height)
	vp.SetContentLines(renderLines)
	vp.FillHeight = true
	return renderLines
}

// RenderContent renders viewport lines with an optional single-column scrollbar.
func RenderContent(
	lines []string,
	vp viewport.Model,
	withScrollbar bool,
	scrollStyles scrollbar.Styles,
) string {
	height := vp.Height()
	width := max(0, vp.Width())
	if height <= 0 {
		return ""
	}

	start := min(vp.YOffset(), len(lines))
	end := min(start+height, len(lines))
	scrollChars := []string(nil)
	if withScrollbar {
		scrollChars = scrollbar.Model{
			Height:     height,
			TotalLines: vp.TotalLineCount(),
			Percent:    vp.ScrollPercent(),
			Styles:     scrollStyles,
		}.Chars()
	}
	blank := strings.Repeat(" ", width)

	var b strings.Builder
	for row := range height {
		if row > 0 {
			b.WriteByte('\n')
		}

		line := blank
		idx := start + row
		if idx < end {
			line = lines[idx]
		}
		b.WriteString(line)

		if row < len(scrollChars) {
			b.WriteString(scrollChars[row])
		}
	}
	return b.String()
}

func renderFooter(components []FooterComponent, width int) string {
	rows := make([][]FooterComponent, 0, 1)
	current := make([]FooterComponent, 0, len(components))

	for _, component := range components {
		if component.Content == "" {
			continue
		}
		if component.BreakBefore && len(current) > 0 {
			rows = append(rows, current)
			current = make([]FooterComponent, 0, len(components))
		}
		current = append(current, component)
	}
	if len(current) > 0 {
		rows = append(rows, current)
	}
	if len(rows) == 0 {
		return ""
	}

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, renderFooterRow(row, width))
	}
	return strings.Join(lines, nl)
}

func renderFooterRow(components []FooterComponent, width int) string {
	var left, right strings.Builder
	for _, component := range components {
		if component.Align == FooterAlignRight {
			right.WriteString(component.Content)
			continue
		}
		left.WriteString(component.Content)
	}

	leftLines := splitFooterLines(left.String())
	rightLines := splitFooterLines(right.String())
	height := max(len(leftLines), len(rightLines))
	if height == 0 {
		return ""
	}

	lines := make([]string, height)
	for i := range height {
		lines[i] = renderFooterLine(
			bottomAlignedLine(leftLines, i, height),
			bottomAlignedLine(rightLines, i, height),
			width,
		)
	}
	return strings.Join(lines, nl)
}

func splitFooterLines(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(text, nl)
}

func bottomAlignedLine(lines []string, row, height int) string {
	if len(lines) == 0 {
		return ""
	}
	offset := height - len(lines)
	idx := row - offset
	if idx < 0 || idx >= len(lines) {
		return ""
	}
	return lines[idx]
}

func renderFooterLine(left, right string, width int) string {
	if width <= 0 {
		switch {
		case left == "":
			return right
		case right == "":
			return left
		default:
			return left + right
		}
	}

	switch {
	case right == "":
		return layout.NormalizeLine(left, width)
	case left == "":
		return alignRight(right, width)
	}

	rightWidth := lg.Width(right)
	if rightWidth >= width {
		return xansi.Truncate(right, width, footerEllipsis)
	}

	leftWidth := lg.Width(left)
	if leftWidth+rightWidth > width {
		left = xansi.Truncate(left, width-rightWidth, "")
		leftWidth = lg.Width(left)
	}

	return left + strings.Repeat(" ", width-leftWidth-rightWidth) + right
}

func alignRight(text string, width int) string {
	textWidth := lg.Width(text)
	if textWidth >= width {
		return xansi.Truncate(text, width, footerEllipsis)
	}
	return strings.Repeat(" ", width-textWidth) + text
}

// RenderFrame renders a full-screen viewport with optional header and footer.
func RenderFrame(m FrameModel) string {
	var b strings.Builder

	if m.Header != "" {
		b.WriteString(m.Header)
		b.WriteString(nl)
		if m.Width > 0 {
			b.WriteString(m.Styles.Separator.Render(strings.Repeat(m.separator(), m.Width)))
		}
		b.WriteString(nl)
	}

	totalLines := m.View.TotalLineCount()
	vpHeight := m.View.Height()
	switch {
	case vpHeight <= 0:
		b.WriteString(nl)
	case totalLines > vpHeight:
		b.WriteString(RenderContent(m.Lines, m.View, true, m.Styles.Scrollbar))
	default:
		b.WriteString(RenderContent(m.Lines, m.View, false, m.Styles.Scrollbar))
	}
	b.WriteString(nl)

	if m.Width > 0 {
		b.WriteString(m.Styles.Separator.Render(strings.Repeat(m.separator(), m.Width)))
	}
	b.WriteString(nl)
	b.WriteString(renderFooter(m.Footer, m.Width))

	return layout.Fill(b.String(), m.Width, m.Height)
}

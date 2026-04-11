package view

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	lg "charm.land/lipgloss/v2"
	"github.com/gechr/primer/layout"
	"github.com/gechr/primer/scrollbar"
)

const nl = "\n"

type FrameStyles struct {
	Separator lg.Style
	Scrollbar scrollbar.Styles
}

type FrameModel struct {
	Footer    string
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
	b.WriteString(m.Footer)

	return layout.Fill(b.String(), m.Width, m.Height)
}

package helpbar

import (
	"strings"

	lg "charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/keyhint"
)

const (
	defaultGap = "   "
	nl         = "\n"
)

type Model struct {
	Hints    []keyhint.Hint
	Renderer keyhint.Renderer
	Status   string
	Width    int
	Ellipsis string
}

// Render returns the bottom help bar content, optionally with right-aligned status.
func (m Model) Render() string {
	help := m.Renderer.Render(m.Hints)
	return AppendRightStatus(help, m.Status, m.Width, m.ellipsis())
}

// Lines returns the number of rendered footer lines at the current width.
func (m Model) Lines() int {
	return strings.Count(m.Renderer.Render(m.Hints), nl) + 1
}

func (m Model) ellipsis() string {
	if m.Ellipsis != "" {
		return m.Ellipsis
	}
	return "…"
}

// AppendRightStatus appends a right-aligned status string to the last help line.
func AppendRightStatus(help, status string, width int, ellipsis string) string {
	if status == "" || width <= 0 {
		return help
	}
	usableWidth := max(1, width-1)
	lastNL := strings.LastIndex(help, nl)
	prefix := ""
	lastLine := help
	if lastNL >= 0 {
		prefix = help[:lastNL+1]
		lastLine = help[lastNL+1:]
	}
	sw := lg.Width(status)
	gap := defaultGap

	for {
		pad := usableWidth - lg.Width(lastLine) - sw
		if pad > 0 {
			return prefix + lastLine + strings.Repeat(" ", pad) + status
		}
		idx := strings.LastIndex(lastLine, gap)
		if idx < 0 {
			break
		}
		lastLine = lastLine[:idx]
	}

	if sw < usableWidth {
		return prefix + strings.Repeat(" ", usableWidth-sw) + status
	}
	return prefix + xansi.Truncate(status, usableWidth, ellipsis)
}

package helpbar

import (
	"strings"

	"github.com/gechr/primer/key"
)

const nl = "\n"

type Model struct {
	Hints    []key.Hint
	Renderer key.Renderer
}

// Render returns the bottom help bar content.
func (m Model) Render() string {
	return m.Renderer.Render(m.Hints)
}

// HintAt returns the hint drawn at column x of line y, relative to the top-left
// of Render's output. The whole "key desc" part is the target, so callers can
// replay the hint's key (see key.Replay) to make the help bar clickable.
func (m Model) HintAt(x, y int) (key.Hint, bool) {
	i, ok := m.Renderer.HintAt(m.Hints, x, y)
	if !ok {
		return key.Hint{}, false
	}
	return m.Hints[i], true
}

// Lines returns the number of rendered footer lines at the current width.
func (m Model) Lines() int {
	return strings.Count(m.Renderer.Render(m.Hints), nl) + 1
}

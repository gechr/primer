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

// Lines returns the number of rendered footer lines at the current width.
func (m Model) Lines() int {
	return strings.Count(m.Renderer.Render(m.Hints), nl) + 1
}

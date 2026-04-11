package helpbar_test

import (
	"strings"
	"testing"

	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/helpbar"
	"github.com/gechr/primer/key"
	"github.com/stretchr/testify/require"
)

func TestModelLines(t *testing.T) {
	m := helpbar.Model{
		Hints: []key.Hint{
			{Key: "a", Desc: "approve"},
			{Key: "c", Desc: "comment"},
		},
		Renderer: key.Renderer{
			Styles: key.Styles{Key: lg.NewStyle(), Text: lg.NewStyle()},
			Width:  12,
			Inline: true,
		},
		Width: 12,
	}

	require.Equal(t, 2, m.Lines())
}

func TestModelRenderRightAlignsStatusWithoutAddingLines(t *testing.T) {
	m := helpbar.Model{
		Hints: []key.Hint{
			{Key: "up/down", Desc: "scroll"},
			{Key: "c", Desc: "comment"},
		},
		Renderer: key.Renderer{
			Styles: key.Styles{Key: lg.NewStyle(), Text: lg.NewStyle()},
			Width:  24,
			Inline: true,
		},
		Status: "Diffing owner/repo#42…",
		Width:  24,
	}

	got := m.Render()
	lines := strings.Split(ansi.Strip(got), "\n")

	require.Len(t, lines, 2)
	require.Contains(t, lines[1], "Diffing")
}

func TestModelRenderPreservesANSIInDescription(t *testing.T) {
	on := lg.NewStyle().Foreground(lg.Color("2")).Render("on")
	m := helpbar.Model{
		Hints: []key.Hint{{Key: "r", Desc: "refresh " + on}},
		Renderer: key.Renderer{
			Styles: key.Styles{Key: lg.NewStyle(), Text: lg.NewStyle()},
			Width:  80,
		},
		Width: 80,
	}

	require.Contains(t, ansi.Strip(m.Render()), "refresh on")
}

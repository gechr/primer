package helpbar_test

import (
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
	}

	require.Equal(t, 2, m.Lines())
}

func TestModelRenderReturnsHelpOnly(t *testing.T) {
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
	}

	got := ansi.Strip(m.Render())
	require.Equal(t, " up/down scroll\ncomment", got)
}

func TestModelRenderPreservesANSIInDescription(t *testing.T) {
	on := lg.NewStyle().Foreground(lg.Color("2")).Render("on")
	m := helpbar.Model{
		Hints: []key.Hint{{Key: "r", Desc: "refresh " + on}},
		Renderer: key.Renderer{
			Styles: key.Styles{Key: lg.NewStyle(), Text: lg.NewStyle()},
			Width:  80,
		},
	}

	require.Equal(t, " r refresh on", ansi.Strip(m.Render()))
}

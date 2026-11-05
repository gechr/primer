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
			Width:  23,
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

func TestModelHintAtReturnsTheClickedHint(t *testing.T) {
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

	hint, ok := m.HintAt(0, 1)
	require.True(t, ok)
	require.Equal(t, key.Hint{Key: "c", Desc: "comment"}, hint)
	_, ok = m.HintAt(0, 0)
	require.False(t, ok, "the prefix is not part of a hint")
}

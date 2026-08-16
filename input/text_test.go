package input_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gechr/primer/input"
	"github.com/stretchr/testify/require"
)

func typeText(t *testing.T, update func(tea.Msg) tea.Cmd, text string) {
	t.Helper()
	update(tea.KeyPressMsg{Text: text})
}

func TestLineTypesAndBackspaces(t *testing.T) {
	t.Parallel()

	l := input.NewLine("> ", "hint")
	typeText(t, l.Update, "abcd")
	l.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	require.Equal(t, "abc", l.Value())
}

func TestLineSetValueAppendsAtEnd(t *testing.T) {
	t.Parallel()

	l := input.NewLine("", "")
	l.SetValue("status = done")
	typeText(t, l.Update, "!")
	require.Equal(t, "status = done!", l.Value())
}

func TestLineBeforeCursorFollowsPosition(t *testing.T) {
	t.Parallel()

	l := input.NewLine("", "")
	typeText(t, l.Update, "hello @al")
	require.Equal(t, "hello @al", l.BeforeCursor())
	l.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	require.Equal(t, "hello @a", l.BeforeCursor())
}

func TestLineReplaceBeforeCursorKeepsSuffix(t *testing.T) {
	t.Parallel()

	l := input.NewLine("", "")
	typeText(t, l.Update, "@al tail")
	for range len(" tail") {
		l.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	}
	l.ReplaceBeforeCursor(len("@al"), "@alice")
	require.Equal(t, "@alice tail", l.Value())
	// The cursor sits at the end of the replacement, not the end of the line.
	require.Equal(t, "@alice", l.BeforeCursor())
}

func TestLineSetWidthClampsNonPositive(t *testing.T) {
	t.Parallel()

	l := input.NewLine("/", "")
	l.SetWidth(-5) // a too-narrow pane must not panic the textinput
	typeText(t, l.Update, "ok")
	l.View()
	require.Equal(t, "ok", l.Value())
}

func TestLinePasteInserts(t *testing.T) {
	t.Parallel()

	l := input.NewLine("", "")
	l.Update(tea.PasteMsg{Content: "pasted text"})
	require.Equal(t, "pasted text", l.Value())
}

func TestLineSuggestionsRoundTrip(t *testing.T) {
	t.Parallel()

	l := input.NewLine("", "")
	l.SetSuggestions([]string{"alpha", "beta"})
	require.Equal(t, []string{"alpha", "beta"}, l.Suggestions())
}

func TestAreaCollectsMultilineText(t *testing.T) {
	t.Parallel()

	a := input.NewArea("hint", 40, 4)
	typeText(t, a.Update, "first")
	a.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	typeText(t, a.Update, "second")
	require.Equal(t, "first\nsecond", a.Value())
}

func TestAreaBeforeCursorIsCurrentLine(t *testing.T) {
	t.Parallel()

	a := input.NewArea("", 40, 4)
	typeText(t, a.Update, "first")
	a.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	typeText(t, a.Update, "sec")
	require.Equal(t, "sec", a.BeforeCursor())
}

func TestAreaReplaceBeforeCursorSwapsToken(t *testing.T) {
	t.Parallel()

	a := input.NewArea("", 40, 4)
	typeText(t, a.Update, "ping @al")
	a.ReplaceBeforeCursor(len("@al"), "@alice")
	require.Equal(t, "ping @alice", a.Value())
}

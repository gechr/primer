package dialog_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/gechr/primer/button"
	"github.com/gechr/primer/dialog"
	"github.com/stretchr/testify/require"
)

// cancelOK is a Cancel/OK pair - deliberately not Yes/No, and with the accept
// button second - with OK holding the default focus and each side a shortcut.
func cancelOK() []dialog.ConfirmButton {
	plain := button.Button{Focused: lg.NewStyle(), Blurred: lg.NewStyle()}
	cancel, ok := plain, plain
	cancel.Label = "Cancel"
	ok.Label = "OK"
	return []dialog.ConfirmButton{
		{Button: cancel, Keys: []string{"n", "N"}},
		{Button: ok, Accept: true, Keys: []string{"y", "Y"}, Default: true},
	}
}

func pressConfirm(c *dialog.Confirm, msg tea.Msg) dialog.Result {
	_, _, res := c.Update(msg)
	return res
}

func TestConfirmButtonsEnterPressesDefault(t *testing.T) {
	t.Parallel()

	c := dialog.NewConfirmButtons("Proceed?", cancelOK()...)
	require.Equal(t, dialog.ResultSubmit, pressConfirm(c, tea.KeyPressMsg{Code: tea.KeyEnter}))
	require.True(t, c.Confirmed())
}

func TestConfirmButtonsFocusStepsToDecline(t *testing.T) {
	t.Parallel()

	c := dialog.NewConfirmButtons("Proceed?", cancelOK()...)
	require.Equal(t, dialog.ResultNone, pressConfirm(c, tea.KeyPressMsg{Code: tea.KeyLeft}))
	require.Equal(t, dialog.ResultClose, pressConfirm(c, tea.KeyPressMsg{Code: tea.KeyEnter}))
	require.False(t, c.Confirmed())
}

func TestConfirmButtonsShortcutKeys(t *testing.T) {
	t.Parallel()

	// A shortcut presses its button regardless of where the focus sits.
	c := dialog.NewConfirmButtons("Proceed?", cancelOK()...)
	pressConfirm(c, tea.KeyPressMsg{Code: tea.KeyLeft}) // focus Cancel
	require.Equal(t, dialog.ResultSubmit, pressConfirm(c, tea.KeyPressMsg{Text: "y", Code: 'y'}))
	require.True(t, c.Confirmed())

	c = dialog.NewConfirmButtons("Proceed?", cancelOK()...)
	require.Equal(t, dialog.ResultClose, pressConfirm(c, tea.KeyPressMsg{Text: "n", Code: 'n'}))
	require.False(t, c.Confirmed())

	// Esc always declines, Keys or not.
	c = dialog.NewConfirmButtons("Proceed?", cancelOK()...)
	require.Equal(t, dialog.ResultClose, pressConfirm(c, tea.KeyPressMsg{Code: tea.KeyEscape}))
}

func TestConfirmButtonsShortcutsAreOptional(t *testing.T) {
	t.Parallel()

	buttons := cancelOK()
	buttons[0].Keys = nil
	buttons[1].Keys = nil
	c := dialog.NewConfirmButtons("Proceed?", buttons...)
	// Without shortcuts a letter is inert; focus+enter still works.
	require.Equal(t, dialog.ResultNone, pressConfirm(c, tea.KeyPressMsg{Text: "y", Code: 'y'}))
	require.Equal(t, dialog.ResultSubmit, pressConfirm(c, tea.KeyPressMsg{Code: tea.KeyEnter}))
}

func TestConfirmButtonsDefaultFocusFallsBackToFirst(t *testing.T) {
	t.Parallel()

	buttons := cancelOK()
	buttons[1].Default = false
	c := dialog.NewConfirmButtons("Proceed?", buttons...)
	// No Default marked: focus starts on the first button - here the decline.
	require.Equal(t, dialog.ResultClose, pressConfirm(c, tea.KeyPressMsg{Code: tea.KeyEnter}))
}

func TestDefaultDestructiveConfirmButtons(t *testing.T) {
	t.Parallel()

	buttons := dialog.DefaultDestructiveConfirmButtons()
	require.Len(t, buttons, 2)
	require.Equal(t, "No", buttons[0].Label)
	require.Equal(t, lg.Color("196"), buttons[0].Focused.GetBackground())
	require.Equal(t, lg.Color("#000000"), buttons[0].Focused.GetForeground())
	require.False(t, buttons[0].Accept)
	require.True(t, buttons[0].Blurred.GetBold())
	require.Equal(t, "Yes", buttons[1].Label)
	require.Equal(t, lg.Color("48"), buttons[1].Focused.GetBackground())
	require.Equal(t, lg.Color("#000000"), buttons[1].Focused.GetForeground())
	require.True(t, buttons[1].Accept)
	require.True(t, buttons[1].Default)
}

func TestConfirmButtonsHideHintRow(t *testing.T) {
	t.Parallel()

	require.Nil(t, dialog.NewConfirmButtons("Proceed?", cancelOK()...).Hints())
	require.Len(t, dialog.NewConfirm("Proceed?").Hints(), 2)
}

// The 12x5 box centers at (14, 3) on a 40x12 screen, so the content starts at
// (15, 4) and the button row - "Cancel  OK" under "Proceed?" - lands on
// screen row 6: Cancel spans columns 15-20, the gap 21-22, OK 23-24.
func TestConfirmButtonsClick(t *testing.T) {
	t.Parallel()

	open := func() (*dialog.Stack, *dialog.Confirm) {
		s := dialog.New(borderedFrame())
		c := dialog.NewConfirmButtons("Proceed?", cancelOK()...)
		s.Push(c)
		s.View(backdrop(12), 40, 12)
		return s, c
	}

	s, c := open()
	_, popped, res := s.Update(leftClick(23, 6))
	require.Equal(t, dialog.ResultSubmit, res)
	require.Same(t, c, popped)
	require.True(t, c.Confirmed())

	s, c = open()
	_, _, res = s.Update(leftClick(15, 6))
	require.Equal(t, dialog.ResultClose, res)
	require.False(t, c.Confirmed())

	// The gap between buttons and the prompt row press nothing.
	s, _ = open()
	_, _, res = s.Update(leftClick(21, 6))
	require.Equal(t, dialog.ResultNone, res)
	_, _, res = s.Update(leftClick(15, 4))
	require.Equal(t, dialog.ResultNone, res)
	require.True(t, s.Active())
}

func TestInfoClosesOnDecisionOrClick(t *testing.T) {
	t.Parallel()

	ok := button.Button{Label: "OK", Focused: lg.NewStyle(), Blurred: lg.NewStyle()}

	i := dialog.NewInfo("saved", ok)
	_, _, res := i.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Equal(t, dialog.ResultClose, res)

	// A stray letter must not dismiss the message.
	i = dialog.NewInfo("saved", ok)
	_, _, res = i.Update(tea.KeyPressMsg{Text: "x", Code: 'x'})
	require.Equal(t, dialog.ResultNone, res)

	// The 7x5 box centers at (16, 3): content (17, 4), OK on row 6 at 18-19.
	s := dialog.New(borderedFrame())
	s.Push(dialog.NewInfo("saved", ok))
	s.View(backdrop(12), 40, 12)
	_, _, res = s.Update(leftClick(17, 6)) // centering pad, not the button
	require.Equal(t, dialog.ResultNone, res)
	_, _, res = s.Update(leftClick(18, 6))
	require.Equal(t, dialog.ResultClose, res)
}

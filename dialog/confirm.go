package dialog

import (
	"slices"

	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/gechr/primer/button"
	"github.com/gechr/primer/key"
	"github.com/gechr/primer/prompt"
	xslices "github.com/gechr/x/slices"
)

// ConfirmButton is one button of a [NewConfirmButtons] dialog: its rendering
// plus its meaning. The caller chooses the labels, the order, and the styles -
// "No/Yes", "Cancel/OK", or anything else - and the semantics travel with each
// button rather than its position.
type ConfirmButton struct {
	button.Button

	// Accept makes pressing this button accept the confirmation; a button
	// without it declines.
	Accept bool
	// Keys, when set, are key names (as tea.KeyPressMsg.String() reports them)
	// that press this button directly from the keyboard, e.g. "y" and "Y" for
	// a Yes button. They are optional: a button without Keys is still pressed
	// through focus+enter or a click. Esc always declines regardless of Keys.
	Keys []string
	// Default marks the button the focus starts on. The first button marked
	// Default wins; with none marked, focus starts on the first button.
	Default bool
}

// Confirm is a generic confirmation dialog. It renders a prompt as its body
// and resolves to ResultSubmit when accepted or ResultClose when declined,
// exposing the decision through Confirmed. It defaults to not confirmed, so a
// dialog abandoned any other way - escape, or a lost overlay the Stack drops -
// is read as a decline, the safe default for a guard in front of a
// destructive action.
//
// Constructed with [NewConfirmButtons] it renders a focusable button row
// beneath the prompt: the arrow keys (or h/l, tab) move the focus, enter
// presses the focused button, each button's Keys press it directly, and a
// mouse click (delivered as [ClickMsg]) presses the button it lands on. The
// plain [NewConfirm] form keeps the key-only y/n contract and advertises it
// in the Frame's hint row instead.
//
// It carries pointer semantics: Update mutates and returns the same value, so
// the caller reads Confirmed off the very pointer the Stack pops.
type Confirm struct {
	prompt     string
	row        button.Row
	meta       []ConfirmButton // parallel to row.Buttons: Accept/Keys semantics
	buttons    bool
	rowX, rowY int // the button row's position within Content, for click hits
	confirmed  bool
}

// NewConfirm returns a Confirm that asks prompt, initially not confirmed.
func NewConfirm(prompt string) *Confirm {
	return &Confirm{prompt: prompt}
}

// NewConfirmButtons returns a Confirm that asks prompt above a button row
// built from buttons, in the order given. Focus starts on the first button
// marked Default (or the first button outright), so the caller decides
// whether a bare enter lands on the safe choice or the common one.
func NewConfirmButtons(prompt string, buttons ...ConfirmButton) *Confirm {
	row := button.Row{
		Buttons: xslices.Map(buttons, func(b ConfirmButton) button.Button { return b.Button }),
		Focus:   max(0, slices.IndexFunc(buttons, func(b ConfirmButton) bool { return b.Default })),
	}
	return &Confirm{prompt: prompt, row: row, meta: buttons, buttons: true}
}

// Confirmed reports whether the dialog was accepted. It is meaningful once the
// Stack has popped the dialog: a Confirm resolved with ResultSubmit reports
// true, and one that was closed - or is still open - reports false.
func (c *Confirm) Confirmed() bool { return c.confirmed }

// Title omits a heading; the prompt carries the whole message.
func (c *Confirm) Title() string { return "" }

// Content renders the prompt, with the button row centered beneath it in
// button mode. The Frame frames, centers, and wraps it, so Confirm is
// box-less and draws no border of its own.
func (c *Confirm) Content(int) string {
	if !c.buttons {
		return c.prompt
	}
	content, x, y := buttonContent(c.prompt, &c.row)
	c.rowX, c.rowY = x, y
	return content
}

// Hints advertises the accept and decline keys for the Frame's foot row. In
// button mode the buttons are the affordance and the foot row is omitted.
func (c *Confirm) Hints() []key.Hint {
	if c.buttons {
		return nil
	}
	return []key.Hint{
		{Key: "y", Desc: "yes"},
		{Key: "n", Desc: "no"},
	}
}

// Update resolves the dialog on a decision and ignores everything else. In
// key mode y/Y/enter accept and n/N/esc decline. In button mode enter presses
// the focused button, the arrows (h/l, tab) move the focus, each button's
// Keys press it directly, esc declines, and a translated click presses the
// button under it. The outcome is recorded before returning, so the popped
// value already carries it.
func (c *Confirm) Update(msg tea.Msg) (Dialog, tea.Cmd, Result) {
	switch m := msg.(type) {
	case ClickMsg:
		if c.buttons && m.Y == c.rowY {
			if i := c.row.Hit(m.X - c.rowX); i >= 0 {
				return c.press(i)
			}
		}
		return c, nil, ResultNone
	case tea.KeyPressMsg:
		return c.updateKey(m)
	}
	return c, nil, ResultNone
}

func (c *Confirm) updateKey(kp tea.KeyPressMsg) (Dialog, tea.Cmd, Result) {
	k := kp.String()
	if !c.buttons {
		switch k {
		case "y", "Y", key.Enter:
			return c.accept()
		case "n", "N", key.Esc:
			return c.decline()
		}
		return c, nil, ResultNone
	}
	pressed := slices.IndexFunc(c.meta, func(b ConfirmButton) bool {
		return slices.Contains(b.Keys, k)
	})
	if pressed >= 0 {
		return c.press(pressed)
	}
	switch k {
	case key.Esc:
		return c.decline()
	case key.Enter:
		return c.press(c.row.Focus)
	case key.Left, "h", key.ShiftTab:
		c.row.Step(-1)
	case key.Right, "l", key.Tab:
		c.row.Step(1)
	}
	return c, nil, ResultNone
}

// press resolves the dialog with button i's meaning.
func (c *Confirm) press(i int) (Dialog, tea.Cmd, Result) {
	if i < 0 || i >= len(c.meta) {
		return c, nil, ResultNone
	}
	if c.meta[i].Accept {
		return c.accept()
	}
	return c.decline()
}

func (c *Confirm) accept() (Dialog, tea.Cmd, Result) {
	c.confirmed = true
	return c, nil, ResultSubmit
}

func (c *Confirm) decline() (Dialog, tea.Cmd, Result) {
	c.confirmed = false
	return c, nil, ResultClose
}

// buttonContent lays a button row centered beneath a prompt, returning the
// composed content plus the row's column and row within it - what a ClickMsg
// is hit-tested against.
func buttonContent(promptText string, row *button.Row) (string, int, int) {
	view := row.View()
	centered := prompt.CenterRow(view, lg.Width(promptText))
	x := lg.Width(centered) - lg.Width(view)
	y := lg.Height(promptText) + 1
	return promptText + "\n\n" + centered, x, y
}

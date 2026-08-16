package dialog

import (
	tea "charm.land/bubbletea/v2"
	"github.com/gechr/primer/button"
	"github.com/gechr/primer/key"
)

// Info is a message-only dialog: a prompt above a single OK button. It always
// resolves to ResultClose - there is no decision to carry - on enter, esc,
// space, or a click on the button; other keys are ignored, so it cannot be
// dismissed by an accidental keystroke the way a momentary overlay can.
type Info struct {
	prompt     string
	row        button.Row
	rowX, rowY int // the button's position within Content, for click hits
}

// NewInfo returns an Info that shows prompt above the injected OK button.
func NewInfo(prompt string, ok button.Button) *Info {
	return &Info{prompt: prompt, row: button.Row{Buttons: []button.Button{ok}}}
}

// Title omits a heading; the prompt carries the whole message.
func (i *Info) Title() string { return "" }

// Content renders the prompt with the OK button centered beneath it.
func (i *Info) Content(int) string {
	content, x, y := buttonContent(i.prompt, &i.row)
	i.rowX, i.rowY = x, y
	return content
}

// Hints omits the foot row; the button is the affordance.
func (i *Info) Hints() []key.Hint { return nil }

// Update closes the dialog on a decision key or a click on the OK button.
func (i *Info) Update(msg tea.Msg) (Dialog, tea.Cmd, Result) {
	switch m := msg.(type) {
	case ClickMsg:
		if m.Y == i.rowY && i.row.Hit(m.X-i.rowX) == 0 {
			return i, nil, ResultClose
		}
	case tea.KeyPressMsg:
		switch m.String() {
		case key.Enter, key.Esc, key.Space:
			return i, nil, ResultClose
		}
	}
	return i, nil, ResultNone
}

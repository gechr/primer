package dialog

import (
	tea "charm.land/bubbletea/v2"
	"github.com/gechr/primer/key"
	"github.com/gechr/primer/list"
)

// Pick is a filterable single-choice dialog: it wraps a searching list.Model,
// letting the Frame lay out and scroll a type-to-filter select list. Typing
// narrows the list and the arrows move the cursor - routed straight into the
// list - while enter accepts the highlighted row (ResultSubmit) and esc
// cancels (ResultClose), the two keys the list itself deliberately leaves to
// its caller. Selection reports the original row index, so the caller keeps a
// parallel slice of values and maps the index back on submit.
//
// It carries pointer semantics: Update mutates and returns the same value, so
// the caller reads Selected off the very pointer the Stack pops.
type Pick struct {
	title string
	list  *list.Model
}

// NewPick returns a Pick over rows, titled title and with the cursor on the
// first entry. Rows may carry styling; the filter matches their visible text.
func NewPick(title string, rows []string) *Pick {
	l := list.New(list.WithSearch())
	// The Frame owns scrolling: give the list room for every row plus the
	// filter line and let the outer viewport do the windowing.
	l.Scrollbar = false
	l.SetSize(0, len(rows)+1)
	l.SetRows(rows)
	return &Pick{title: title, list: l}
}

// Selected returns the original index of the highlighted row; ok is false
// when the filter matches nothing (or the list is empty). It is meaningful
// once the Stack has popped the dialog with ResultSubmit.
func (p *Pick) Selected() (int, bool) { return p.list.Selected() }

// Title is the heading the Frame renders above the list.
func (p *Pick) Title() string { return p.title }

// Content renders the list - filter line and matching rows - as the body. The
// list does not wrap to a width, so the parameter is unused; the Frame lays out
// and scrolls whatever it returns.
func (p *Pick) Content(int) string { return p.list.View() }

// Hints advertises the accept and cancel keys for the Frame's foot row.
func (p *Pick) Hints() []key.Hint {
	return []key.Hint{
		{Key: "enter", Desc: "select"},
		{Key: "esc", Desc: "cancel"},
	}
}

// ScrollTo keeps the highlighted row visible as the cursor moves through a
// list taller than the box. Implements ScrollHint.
func (p *Pick) ScrollTo() (int, int, bool) {
	if line := p.list.CursorLine(); line >= 0 {
		return line, 1, true
	}
	return 0, 0, false
}

// Update resolves the dialog on enter (accept) or esc (cancel) and routes
// everything else - filter typing, paste, arrow navigation - into the list.
// Matching on Code, not String, keeps a modified enter or escape working and
// mirrors the list's own key handling.
func (p *Pick) Update(msg tea.Msg) (Dialog, tea.Cmd, Result) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.Code {
		case tea.KeyEnter:
			return p, nil, ResultSubmit
		case tea.KeyEscape:
			return p, nil, ResultClose
		}
	}
	return p, p.list.Update(msg), ResultNone
}

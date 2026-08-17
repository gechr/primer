package dialog

import (
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/gechr/primer/form"
	"github.com/gechr/primer/key"
)

// Form adapts a form.Model into a Dialog: the form's scrollable Body becomes
// Content, its pinned Foot the Footer, and its FocusRegion the scroll hint,
// so a tall form follows its focused field inside the Frame while the
// hint/confirm/submitting row stays visible. EventSubmit resolves to
// ResultSubmit and EventCancel to ResultClose - the owner reads the values
// off the popped *Form through Model. Events that keep the form open but need
// the owner to act - EventEditor, EventChanged, and EventSubmit under
// HoldOnSubmit - are delivered through the OnEvent callback.
type Form struct {
	model          form.Model
	onEvent        func(form.EventKind)
	discard        *Confirm
	discardButtons []ConfirmButton
	nerdFonts      bool

	// HoldOnSubmit keeps the form open when it submits: instead of resolving
	// ResultSubmit, the adapter reports EventSubmit through OnEvent and stays
	// put - for owners that freeze the form (SetSubmitting) around an async
	// write and close it themselves with [Stack.Pop] when the write resolves.
	HoldOnSubmit bool
}

// FormOption configures a form dialog created by [NewForm].
type FormOption func(*Form)

// WithDiscardButtons replaces the default red No / green Yes buttons used by
// the dirty-discard confirmation. An empty list leaves the defaults intact.
func WithDiscardButtons(buttons ...ConfirmButton) FormOption {
	return func(f *Form) {
		if len(buttons) > 0 {
			f.discardButtons = append([]ConfirmButton(nil), buttons...)
		}
	}
}

// WithNerdFonts renders form-owned buttons with Nerd Font half-circle caps.
// Callers should enable it only after detecting that the terminal font
// supports the glyphs.
func WithNerdFonts() FormOption {
	return func(f *Form) { f.nerdFonts = true }
}

// NewForm wraps m as a dialog. onEvent, when non-nil, observes the events
// that need the owner to act while the form stays open: EventEditor (take
// the draft to an external editor), EventChanged (a Notify cycle field
// stepped - refetch dependent options through Model), and EventSubmit when
// HoldOnSubmit is set.
func NewForm(m form.Model, onEvent func(form.EventKind), opts ...FormOption) *Form {
	f := &Form{
		model:          m,
		onEvent:        onEvent,
		discardButtons: DefaultDestructiveConfirmButtons(),
	}
	for _, opt := range opts {
		opt(f)
	}
	if f.nerdFonts {
		f.discardButtons = nerdFontButtons(f.discardButtons)
	}
	return f
}

func nerdFontButtons(buttons []ConfirmButton) []ConfirmButton {
	const (
		pillLeft  = "\ue0b6" // nf-ple-left_half_circle_thick
		pillRight = "\ue0b4" // nf-ple-right_half_circle_thick
	)
	out := append([]ConfirmButton(nil), buttons...)
	for i := range out {
		focused := out[i].Focused
		focusedCaps := lg.NewStyle().Foreground(focused.GetBackground())
		out[i].Focused = lg.NewStyle().Transform(func(label string) string {
			return focusedCaps.Render(
				pillLeft,
			) + focused.Render(
				label,
			) + focusedCaps.Render(
				pillRight,
			)
		})
		blurred := out[i].Blurred
		out[i].Blurred = lg.NewStyle().Transform(func(label string) string {
			return " " + blurred.Render(label) + " "
		})
	}
	return out
}

// Model returns the wrapped form, for reads and owner-driven mutations:
// Values after a submit, SetOptions after an EventChanged, SetSubmitting and
// SetError around an async write.
func (f *Form) Model() *form.Model { return &f.model }

// Title omits the Frame heading; the form renders its own title in its body.
func (f *Form) Title() string { return "" }

// Content is the form's scrollable body; the form is fixed-measure (its
// Config.Width), so the offered width is ignored.
func (f *Form) Content(int) string { return f.model.Body() }

// Footer pins the form's hint/confirm/submitting row below the viewport.
func (f *Form) Footer() string { return f.model.Foot() }

// Hints are omitted; the form's Foot carries its own bindings.
func (f *Form) Hints() []key.Hint { return nil }

// ScrollTo reports the focused field's region so the Frame follows it.
func (f *Form) ScrollTo() (int, int, bool) { return f.model.FocusRegion() }

// Update routes a message into the form and maps its event onto the dialog
// contract.
func (f *Form) Update(msg tea.Msg) (Dialog, tea.Cmd, Result) {
	// The form has no pointer affordances yet; swallowing clicks keeps them
	// from reaching the form's catch-all message path, which would clear
	// field errors and accepted-suggestion state as if the user had typed.
	if _, ok := msg.(ClickMsg); ok {
		return f, nil, ResultNone
	}
	cmd, ev, _ := f.model.Update(msg)
	switch ev {
	case form.EventSubmit:
		if f.HoldOnSubmit {
			f.notify(ev)
			return f, cmd, ResultNone
		}
		return f, cmd, ResultSubmit
	case form.EventCancel:
		return f, cmd, ResultClose
	case form.EventConfirmDiscard:
		f.discard = NewConfirmButtons(
			"Discard your changes?\n\nYour input will be lost.",
			f.discardButtons...,
		)
	case form.EventEditor, form.EventChanged:
		f.notify(ev)
	case form.EventNone:
	}
	return f, cmd, ResultNone
}

func (f *Form) takeChild() Dialog {
	if f.discard == nil {
		return nil
	}
	return f.discard
}

func (f *Form) resolveChild(child Dialog, result Result) (Dialog, Result, bool) {
	if f.discard == nil || child != f.discard {
		return f, ResultNone, false
	}
	f.discard = nil
	if f.model.ResolveDiscard(result == ResultSubmit) == form.EventCancel {
		return f, ResultClose, true
	}
	return f, ResultNone, true
}

func (f *Form) notify(ev form.EventKind) {
	if f.onEvent != nil {
		f.onEvent(ev)
	}
}

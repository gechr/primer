package dialog_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gechr/primer/dialog"
	"github.com/gechr/primer/form"
	"github.com/stretchr/testify/require"
)

// The adapter must plug into the Shell's pinned-footer and follow-focus
// scrolling seams, not just the base Dialog contract.
var (
	_ dialog.Footered   = (*dialog.Form)(nil)
	_ dialog.ScrollHint = (*dialog.Form)(nil)
)

func commentForm() form.Model {
	return form.New(form.Config{
		Title:  "add comment",
		Fields: []form.FieldSpec{{Label: "message"}},
		Width:  30,
	})
}

func TestFormDialogSubmitPopsWithValues(t *testing.T) {
	t.Parallel()

	s := dialog.New(borderedShell())
	s.Push(dialog.NewForm(commentForm(), nil))
	s.Update(tea.KeyPressMsg{Text: "hi"})
	_, popped, res := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Equal(t, dialog.ResultSubmit, res)
	f, ok := popped.(*dialog.Form)
	require.True(t, ok)
	require.Equal(t, []string{"hi"}, f.Model().Values())
	require.False(t, s.Active())
}

func TestFormDialogCancelCloses(t *testing.T) {
	t.Parallel()

	s := dialog.New(borderedShell())
	s.Push(dialog.NewForm(commentForm(), nil))
	_, _, res := s.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.Equal(t, dialog.ResultClose, res)
	require.False(t, s.Active())
}

func TestFormDialogEditorEventKeepsOpen(t *testing.T) {
	t.Parallel()

	var events []form.EventKind
	m := form.New(form.Config{
		Fields:      []form.FieldSpec{{Multiline: true}},
		EditorHatch: true,
		Width:       30,
	})
	s := dialog.New(borderedShell())
	s.Push(dialog.NewForm(m, func(ev form.EventKind) { events = append(events, ev) }))
	_, _, res := s.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	require.Equal(t, dialog.ResultNone, res)
	require.Equal(t, []form.EventKind{form.EventEditor}, events)
	require.True(t, s.Active())
}

func TestFormDialogHoldOnSubmit(t *testing.T) {
	t.Parallel()

	var events []form.EventKind
	d := dialog.NewForm(commentForm(), func(ev form.EventKind) { events = append(events, ev) })
	d.HoldOnSubmit = true
	s := dialog.New(borderedShell())
	s.Push(d)
	s.Update(tea.KeyPressMsg{Text: "hi"})
	_, _, res := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Equal(t, dialog.ResultNone, res)
	require.Equal(t, []form.EventKind{form.EventSubmit}, events)
	require.True(t, s.Active())

	// The async write resolved: the owner closes the held form itself.
	popped := s.Pop()
	require.Same(t, d, popped)
	require.False(t, s.Active())
	require.Nil(t, s.Pop())
}

func TestFormDialogRendersBodyAndPinnedFoot(t *testing.T) {
	t.Parallel()

	s := dialog.New(borderedShell())
	s.Push(dialog.NewForm(commentForm(), nil))
	got := s.View(backdrop(20), 60, 20)
	require.Equal(t, 1, shows(got, "add comment"))
	require.Equal(t, 1, shows(got, " message ")) // titlebox-embedded field label
	require.Equal(t, 1, shows(got, "submit"))    // the form's own foot row
}

func TestFormDialogSwallowsClicks(t *testing.T) {
	t.Parallel()

	s := dialog.New(borderedShell())
	s.Push(dialog.NewForm(commentForm(), nil))
	s.Update(tea.KeyPressMsg{Text: "draft"})
	s.View(backdrop(20), 60, 20)
	_, _, res := s.Update(leftClick(30, 10))
	require.Equal(t, dialog.ResultNone, res)
	require.True(t, s.Active())
	// The click reached the adapter, not the form: the draft is untouched.
	top, ok := s.Top().(*dialog.Form)
	require.True(t, ok)
	require.Equal(t, "draft", top.Model().Value(0))
}

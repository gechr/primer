package form_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/gechr/primer/form"
	"github.com/stretchr/testify/require"
)

func press(t *testing.T, m *form.Model, msg tea.KeyPressMsg) form.EventKind {
	t.Helper()
	_, ev, _ := m.Update(msg)
	return ev
}

func typeText(t *testing.T, m *form.Model, s string) {
	t.Helper()
	require.Equal(t, form.EventNone, press(t, m, tea.KeyPressMsg{Text: s}))
}

// shows counts how many times needle appears in the rendered view, so
// presence asserts through exact counting rather than a bare Contains.
func shows(view, needle string) int { return strings.Count(view, needle) }

var (
	esc   = tea.KeyPressMsg{Code: tea.KeyEscape}
	enter = tea.KeyPressMsg{Code: tea.KeyEnter}
	tab   = tea.KeyPressMsg{Code: tea.KeyTab}
	ctrlS = tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
	ctrlE = tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl}
	keyDn = tea.KeyPressMsg{Code: tea.KeyDown}
	shTab = tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
)

var (
	keyLeft  = tea.KeyPressMsg{Code: tea.KeyLeft}
	keyRight = tea.KeyPressMsg{Code: tea.KeyRight}
)

func TestFocusRegionTracksFocus(t *testing.T) {
	t.Parallel()

	// A titled three-field form: title (1 line) then one titled box per text
	// field (3 lines: top border, body, bottom border). The region top must land
	// on each box's top border as focus moves, matching View's own layout.
	m := form.New(form.Config{
		Title: "new item",
		Fields: []form.FieldSpec{
			{Label: "summary"},
			{Label: "assignee"},
			{Label: "labels"},
		},
		Width: 40,
	})

	top, height, ok := m.FocusRegion()
	require.True(t, ok)
	require.Equal(t, 1, top)
	require.Equal(t, 3, height)

	press(t, &m, tab) // focus the second field
	top, _, _ = m.FocusRegion()
	require.Equal(t, 4, top)

	press(t, &m, tab) // focus the third field
	top, _, _ = m.FocusRegion()
	require.Equal(t, 7, top)
	// The reported top must be a real line in View, and everything above it must
	// be the two earlier blocks plus the title.
	lines := strings.Split(m.View(), "\n")
	require.Equal(t, 1, shows(lines[top], "labels"))
}

func TestFocusRegionInertForm(t *testing.T) {
	t.Parallel()

	var m form.Model
	_, _, ok := m.FocusRegion()
	require.False(t, ok)
}

func TestCycleFieldStepsAndWraps(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{
			{Label: "type", Options: []string{"Task", "Bug", "Story"}, Initial: "Bug"},
		},
		Width: 40,
	})
	require.Equal(t, "Bug", m.Value(0))
	press(t, &m, keyRight)
	require.Equal(t, "Story", m.Value(0))
	press(t, &m, keyRight) // wraps past the end
	require.Equal(t, "Task", m.Value(0))
	press(t, &m, keyLeft) // wraps past the start
	require.Equal(t, "Story", m.Value(0))
}

func TestCycleFieldInitialUnknownStartsFirst(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{{Options: []string{"Task", "Bug"}, Initial: "Epic"}},
		Width:  40,
	})
	require.Equal(t, "Task", m.Value(0))
}

// A required cycle field is never blank, so a submit is never blocked on it.
func TestCycleFieldSubmitsWithoutBlocking(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{{Options: []string{"Task", "Bug"}}},
		Width:  40,
	})
	require.Equal(t, form.EventSubmit, press(t, &m, enter))
}

// A cycle field whose Initial matches no option (an unset default) still
// opens pristine: esc must cancel outright, not prompt to discard.
func TestCycleFieldUnsetInitialIsNotDirty(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{{Options: []string{"Task", "Epic"}, Initial: ""}},
		Width:  40,
	})
	require.Equal(t, form.EventCancel, press(t, &m, esc))
}

func TestCycleFieldViewShowsSelection(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{{Options: []string{"Task", "Bug"}, Initial: "Bug"}},
		Width:  40,
	})
	require.Equal(t, 1, shows(m.View(), "‹ Bug ›"))
}

// A field with an accepted suggestion exposes that suggestion's Detail (a
// user ID behind a display name), and a later edit invalidates it.
func TestAcceptedDetailTracksAcceptanceAndClearsOnEdit(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{{
			Autocomplete: &form.Autocomplete{
				MinQuery:   3,
				IsBoundary: func(rune) bool { return false }, // whole field is the query
				Fetch:      func(string) []form.Suggestion { return nil },
			},
		}},
		Width: 40,
	})
	typeText(t, &m, "ali")
	// The fetch runs async; feed the result the query is waiting on directly.
	m.Update(form.SuggestionsMsg{Field: 0, Query: "ali", Items: []form.Suggestion{
		{Value: "Alice Smith", Label: "Alice Smith", Detail: "acc-99"},
	}})
	press(t, &m, enter) // accept the highlighted suggestion
	require.Equal(t, "Alice Smith", m.Value(0))
	require.Equal(t, "acc-99", m.AcceptedDetail(0))
	typeText(t, &m, "x") // any edit invalidates the recorded detail
	require.Empty(t, m.AcceptedDetail(0))
}

func twoField() form.Model {
	return newTwoField(false)
}

func newTwoField(confirmDiscard bool) form.Model {
	return form.New(form.Config{
		Title:          "create item",
		ConfirmDiscard: confirmDiscard,
		Fields: []form.FieldSpec{
			{Label: "summary", Placeholder: "one line"},
			{Label: "description", Multiline: true, Optional: true},
		},
		Width: 40,
	})
}

func TestEnterOnLineAdvancesThenCtrlSSubmits(t *testing.T) {
	t.Parallel()

	m := twoField()
	typeText(t, &m, "fix the flux")
	require.Equal(t, form.EventNone, press(t, &m, enter))
	typeText(t, &m, "sparks when engaged")
	require.Equal(t, form.EventSubmit, press(t, &m, ctrlS))
	require.Equal(t, []string{"fix the flux", "sparks when engaged"}, m.Values())
}

func TestTabCyclesFocusBothWays(t *testing.T) {
	t.Parallel()

	m := twoField()
	press(t, &m, tab)
	typeText(t, &m, "body")
	press(t, &m, shTab)
	typeText(t, &m, "title")
	require.Equal(t, []string{"title", "body"}, m.Values())
}

func TestSubmitBlockedOnBlankRequiredField(t *testing.T) {
	t.Parallel()

	m := twoField()
	press(t, &m, tab) // land on the optional description
	typeText(t, &m, "only a description")
	require.Equal(t, form.EventNone, press(t, &m, ctrlS))
	// The failed submit re-focused the blank summary.
	typeText(t, &m, "now titled")
	require.Equal(t, form.EventSubmit, press(t, &m, ctrlS))
	require.Equal(t, "now titled", m.Value(0))
}

func TestEnterSubmitsSingleLineForm(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{Fields: []form.FieldSpec{{Placeholder: "who"}}, Width: 40})
	typeText(t, &m, "alice")
	require.Equal(t, form.EventSubmit, press(t, &m, enter))
}

func TestEscOnPristineFormCancels(t *testing.T) {
	t.Parallel()

	m := twoField()
	require.Equal(t, form.EventCancel, press(t, &m, esc))
}

func TestEscOnPrefilledUntouchedFormCancels(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{Fields: []form.FieldSpec{{Initial: "bug, ux"}}, Width: 40})
	require.Equal(t, form.EventCancel, press(t, &m, esc))
}

func TestDirtyEscCancelsByDefault(t *testing.T) {
	t.Parallel()

	m := twoField()
	typeText(t, &m, "half-written thought")
	require.Equal(t, form.EventCancel, press(t, &m, esc))
}

func TestDirtyEscAsksBeforeDiscarding(t *testing.T) {
	t.Parallel()

	m := newTwoField(true)
	typeText(t, &m, "half-written thought")
	require.Equal(t, form.EventConfirmDiscard, press(t, &m, esc))
	// Declining resumes editing with the draft intact.
	require.Equal(t, form.EventNone, m.ResolveDiscard(false))
	require.Equal(t, "half-written thought", m.Value(0))
	// A second esc re-asks; accepting abandons.
	require.Equal(t, form.EventConfirmDiscard, press(t, &m, esc))
	require.Equal(t, form.EventCancel, m.ResolveDiscard(true))
}

func TestConfirmSwallowsStrayKeys(t *testing.T) {
	t.Parallel()

	m := newTwoField(true)
	typeText(t, &m, "draft")
	press(t, &m, esc)
	require.Equal(t, form.EventNone, press(t, &m, tea.KeyPressMsg{Text: "q", Code: 'q'}))
	require.Equal(t, "draft", m.Value(0))
}

func TestEditorHatchOnlyOnMultiline(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields:      []form.FieldSpec{{Placeholder: "title"}, {Multiline: true, Optional: true}},
		EditorHatch: true,
		Width:       40,
	})
	require.Equal(t, form.EventNone, press(t, &m, ctrlE))
	press(t, &m, tab)
	require.Equal(t, form.EventEditor, press(t, &m, ctrlE))
}

func TestHintsFollowShape(t *testing.T) {
	t.Parallel()

	single := form.New(form.Config{Fields: []form.FieldSpec{{}}, Width: 40})
	v := single.View()
	require.Equal(t, 1, shows(v, "submit"))
	require.Equal(t, 0, shows(v, "next field"))

	multi := form.New(form.Config{
		Fields:      []form.FieldSpec{{}, {Multiline: true}},
		EditorHatch: true,
		Width:       40,
	})
	v = multi.View()
	for _, want := range []string{"submit", "next field", "editor", "cancel"} {
		require.Equal(t, 1, shows(v, want))
	}
}

func TestValidateBlocksSubmitAndShowsInline(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{{
			Placeholder: "duration",
			Validate: func(s string) error {
				if s != "2h" {
					return errors.New("not a duration")
				}
				return nil
			},
		}},
		Styles: form.Styles{Error: lg.NewStyle()},
		Width:  40,
	})
	typeText(t, &m, "banana")
	require.Equal(t, form.EventNone, press(t, &m, enter))
	require.Equal(t, 1, shows(m.View(), "not a duration"))
	// Editing clears the inline error; a valid value then submits.
	for range len("banana") {
		press(t, &m, tea.KeyPressMsg{Code: tea.KeyBackspace})
	}
	require.Equal(t, 0, shows(m.View(), "not a duration"))
	typeText(t, &m, "2h")
	require.Equal(t, form.EventSubmit, press(t, &m, enter))
}

func TestOptionalMarkerRendersOnLabel(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{
			{Label: "summary"},
			{Label: "description", Optional: true},
		},
		Width: 40,
	})
	v := m.View()
	require.Equal(t, 1, shows(v, "description (optional)"))
	require.Equal(t, 0, shows(v, "summary (optional)"))
}

func TestSubmittingAndErrorFootRow(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{{Placeholder: "text"}},
		Styles: form.Styles{Error: lg.NewStyle()},
		Width:  40,
	})
	typeText(t, &m, "hi")
	m.SetSubmitting("*")
	v := m.View()
	require.Equal(t, 1, shows(v, "submitting…"))
	require.Equal(t, 1, shows(v, "*"))
	// A frozen form swallows keys so a second submit can't fire mid-write.
	require.Equal(t, form.EventNone, press(t, &m, ctrlS))
	require.Equal(t, "hi", m.Value(0))
	m.SetError("write failed")
	v = m.View()
	require.Equal(t, 1, shows(v, "write failed"))
	require.Equal(t, 0, shows(v, "submitting…"))
	require.Equal(t, "hi", m.Value(0))
}

func TestZeroValueIsInert(t *testing.T) {
	t.Parallel()

	var m form.Model
	require.False(t, m.Active())
	_, ev, consumed := m.Update(enter)
	require.Equal(t, form.EventNone, ev)
	require.False(t, consumed)
	require.Empty(t, m.View())
}

func TestUpdateReportsConsumed(t *testing.T) {
	t.Parallel()

	m := twoField()
	_, _, consumed := m.Update(tea.KeyPressMsg{Text: "x"})
	require.True(t, consumed)
}

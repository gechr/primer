package prompt_test

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/prompt"
	"github.com/gechr/primer/scrollbar"
	"github.com/stretchr/testify/require"
)

func TestRenderScrollableUsesViewportWhenContentOverflows(t *testing.T) {
	vp := viewport.New()
	vp.SetYOffset(1)

	got := prompt.RenderScrollable(prompt.ScrollableModel{
		BoxStyle:       lg.NewStyle(),
		Content:        "line 1\nline 2\nline 3\nline 4",
		View:           vp,
		ViewportHeight: 3,
		ViewWidth:      8,
		Styles: prompt.Styles{
			Scrollbar: scrollbar.Styles{
				Thumb: lg.NewStyle(),
				Track: lg.NewStyle(),
			},
		},
	})

	lines := strings.Split(ansi.Strip(got), "\n")
	require.NotEmpty(t, lines)
	require.Contains(t, got, "┃")
}

func TestRenderScrollableRendersSimpleBoxWhenContentFits(t *testing.T) {
	got := prompt.RenderScrollable(prompt.ScrollableModel{
		BoxStyle:       lg.NewStyle(),
		Content:        "line 1\nline 2",
		ViewportHeight: 4,
		ViewWidth:      8,
	})

	require.Equal(t, "line 1\nline 2", ansi.Strip(got))
}

func TestRenderChoiceGroups(t *testing.T) {
	got := prompt.RenderChoiceGroups(
		[]prompt.ChoiceGroup{
			{
				Label: "Provider",
				Choices: []prompt.Choice{
					{Label: "claude"},
					{Label: "codex"},
				},
			},
		},
		[]int{0},
		0,
		true,
		prompt.ChoiceGroupStyles{
			Label:          lg.NewStyle(),
			SelectedActive: lg.NewStyle(),
			Selected:       lg.NewStyle(),
			Active:         lg.NewStyle(),
			Inactive:       lg.NewStyle(),
		},
	)

	require.Equal(t, "Provider\nclaude  codex\n\n", ansi.Strip(got))
}

func TestRenderHintLines(t *testing.T) {
	got := prompt.RenderHintLines([][]prompt.Hint{
		{
			{Key: "tab", Text: "next"},
			{Key: "space", Text: "cycle"},
		},
		{
			{Key: "enter", Text: "submit"},
		},
	}, "   ", lg.NewStyle(), lg.NewStyle())

	require.Equal(t, "tab next   space cycle   enter submit", ansi.Strip(got))
}

func TestComposeContentWithoutField(t *testing.T) {
	got := prompt.ComposeContent(prompt.ContentModel{Prompt: "Confirm", HasField: false})

	require.Equal(t, "Confirm\n\n", got)
}

func TestComposeContentWithField(t *testing.T) {
	got := prompt.ComposeContent(prompt.ContentModel{
		Prompt:       "Confirm",
		Options:      "Provider\nclaude\n\n",
		FieldLabel:   "Comment",
		FieldBody:    "body",
		Hints:        "enter submit",
		IncludeHints: true,
		HasField:     true,
	})

	require.Equal(t, "Confirm\n\nProvider\nclaude\n\nComment\nbody\n\nenter submit", got)
}

func TestCenterRow(t *testing.T) {
	require.Equal(t, "  ok", prompt.CenterRow("ok", 6))
}

func TestFocusOptionsCursor(t *testing.T) {
	require.Equal(t, 0, prompt.FocusOptionsCursor(3, false))
	require.Equal(t, 2, prompt.FocusOptionsCursor(3, true))
}

func TestNavigateOptions(t *testing.T) {
	require.Equal(
		t,
		prompt.NavResult{Cursor: 1},
		prompt.NavigateOptions(0, 3, true, prompt.NavDown),
	)
	require.Equal(
		t,
		prompt.NavResult{Cursor: 2, MoveToInput: true},
		prompt.NavigateOptions(2, 3, true, prompt.NavDown),
	)
	require.Equal(
		t,
		prompt.NavResult{Cursor: 0, MoveToInput: true},
		prompt.NavigateOptions(0, 3, true, prompt.NavShiftTab),
	)
	require.Equal(
		t,
		prompt.NavResult{Cursor: 0, MoveToInput: true},
		prompt.NavigateOptions(0, 3, true, prompt.NavUp),
	)
	require.Equal(
		t,
		prompt.NavResult{Cursor: 2},
		prompt.NavigateOptions(0, 3, false, prompt.NavShiftTab),
	)
}

func TestStateToggleYes(t *testing.T) {
	s := prompt.State{Yes: false}
	s.ToggleYes()
	require.True(t, s.Yes)
	s.ToggleYes()
	require.False(t, s.Yes)
}

func TestStateEnterOptions(t *testing.T) {
	s := prompt.State{}
	s.EnterOptions(3, false)
	require.True(t, s.OptFocus)
	require.Equal(t, 0, s.OptCursor)

	s.EnterOptions(3, true)
	require.Equal(t, 2, s.OptCursor)
}

func TestStateNavigate(t *testing.T) {
	s := prompt.State{OptFocus: true, OptCursor: 0}

	// Move down - cursor updated in-place.
	result := s.Navigate(prompt.NavDown, 3, true)
	require.False(t, result.MoveToInput)
	require.Equal(t, 1, s.OptCursor)

	// Move down to last, then down again - moves to input, cursor unchanged.
	s.OptCursor = 2
	result = s.Navigate(prompt.NavDown, 3, true)
	require.True(t, result.MoveToInput)
	require.Equal(t, 2, s.OptCursor) // not mutated on MoveToInput
}

func TestStateStep(t *testing.T) {
	s := prompt.State{OptCursor: 1, OptValues: []int{0, 1, 0}}
	s.Step(4, 1, false)
	require.Equal(t, []int{0, 2, 0}, s.OptValues)

	// Out-of-bounds cursor is a no-op.
	s.OptCursor = 5
	s.Step(4, 1, false)
	require.Equal(t, []int{0, 2, 0}, s.OptValues)
}

func TestStateFocusLine(t *testing.T) {
	s := prompt.State{OptFocus: true, OptCursor: 1}
	require.Equal(t, 5, s.FocusLine(3, true))

	s.OptFocus = false
	require.Equal(t, 12, s.FocusLine(3, true))
}

func TestFocusLineOffset(t *testing.T) {
	// Option focused - first option.
	require.Equal(t, 2, prompt.FocusLineOffset(true, 0, 3, true))

	// Option focused - second option (3 lines per option group).
	require.Equal(t, 5, prompt.FocusLineOffset(true, 1, 3, true))

	// Input focused with options (3 prompt lines + 4×3 option lines = 15).
	require.Equal(t, 15, prompt.FocusLineOffset(false, 0, 4, true))

	// Input focused without options.
	require.Equal(t, 3, prompt.FocusLineOffset(false, 0, 0, true))

	// No focus target.
	require.Equal(t, -1, prompt.FocusLineOffset(false, 0, 0, false))

	// Option focus but no options falls through to input.
	require.Equal(t, 3, prompt.FocusLineOffset(true, 0, 0, true))
}

func TestStepChoice(t *testing.T) {
	require.Equal(t, 2, prompt.StepChoice(1, 4, 1, false))
	require.Equal(t, 0, prompt.StepChoice(0, 4, -1, false))
	require.Equal(t, 0, prompt.StepChoice(3, 4, 1, true))
}

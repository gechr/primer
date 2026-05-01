package prompt

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	lg "charm.land/lipgloss/v2"
	"github.com/gechr/primer/scrollbar"
)

type Styles struct {
	Scrollbar scrollbar.Styles
}

type Hint struct {
	Key  string
	Text string
}

type Choice struct {
	Label string
}

type ChoiceGroup struct {
	Label   string
	Choices []Choice
}

type ChoiceGroupStyles struct {
	Label          lg.Style
	SelectedActive lg.Style
	Selected       lg.Style
	Active         lg.Style
	Inactive       lg.Style
	ChoiceGap      string
	GroupGap       string
}

type ScrollableModel struct {
	BoxStyle        lg.Style
	BoxWidth        int
	Content         string
	ScrollbarConfig scrollbar.Config
	View            viewport.Model
	ViewportHeight  int
	ViewWidth       int
	Styles          Styles
}

type NavDirection int

const (
	NavUp NavDirection = iota
	NavDown
	NavTab
	NavShiftTab
)

type NavResult struct {
	Cursor      int
	MoveToInput bool
}

type ContentModel struct {
	Prompt       string
	Options      string
	FieldLabel   string
	FieldBody    string
	Hints        string
	IncludeHints bool
	HasField     bool
}

// State holds generic interaction state for a prompt with optional choice
// groups and a yes/no selection. Apps embed this alongside their own
// app-specific fields (textarea, viewport, action semantics, etc.).
type State struct {
	OptFocus  bool
	OptCursor int
	OptValues []int
	Yes       bool
}

// ToggleYes flips the yes/no selection.
func (s *State) ToggleYes() {
	s.Yes = !s.Yes
}

// EnterOptions sets focus to the option list and positions the cursor at the
// first or last option depending on reverse.
func (s *State) EnterOptions(optionCount int, reverse bool) {
	s.OptFocus = true
	s.OptCursor = FocusOptionsCursor(optionCount, reverse)
}

// Navigate moves the option cursor in the given direction and returns whether
// focus should move to the input field. The cursor is updated in-place when
// the result does not move to input.
func (s *State) Navigate(dir NavDirection, optionCount int, hasInput bool) NavResult {
	result := NavigateOptions(s.OptCursor, optionCount, hasInput, dir)
	if !result.MoveToInput {
		s.OptCursor = result.Cursor
	}
	return result
}

// Step moves or cycles the selected value for the currently focused option
// group.
func (s *State) Step(choiceCount, delta int, wrap bool) {
	if s.OptCursor >= 0 && s.OptCursor < len(s.OptValues) {
		s.OptValues[s.OptCursor] = StepChoice(s.OptValues[s.OptCursor], choiceCount, delta, wrap)
	}
}

// FocusLine returns the content line offset of the currently focused element,
// for use with viewport scroll-to-visible. Returns -1 when there is nothing
// to scroll to.
func (s *State) FocusLine(optionCount int, hasInput bool) int {
	return FocusLineOffset(s.OptFocus && optionCount > 0, s.OptCursor, optionCount, hasInput)
}

// RenderScrollable renders content inside a styled box, using a viewport and
// scrollbar when the content exceeds the available height.
func RenderScrollable(m ScrollableModel) string {
	lines := strings.Split(m.Content, "\n")
	if len(lines) <= m.ViewportHeight {
		if m.BoxWidth > 0 {
			return m.BoxStyle.Width(m.BoxWidth).Render(m.Content)
		}
		return m.BoxStyle.Render(m.Content)
	}

	view := m.View
	view.SetWidth(max(1, m.ViewWidth))
	view.SetHeight(m.ViewportHeight)
	view.SetContent(m.Content)

	scroll := scrollbar.Model{
		Config:     m.ScrollbarConfig,
		Height:     m.ViewportHeight,
		TotalLines: view.TotalLineCount(),
		Percent:    view.ScrollPercent(),
		Styles:     m.Styles.Scrollbar,
	}.Render()
	inner := lg.JoinHorizontal(lg.Top, view.View(), scroll)

	boxStyle := m.BoxStyle.PaddingRight(1)
	if m.BoxWidth > 0 {
		return boxStyle.Width(m.BoxWidth).Render(inner)
	}
	return boxStyle.Render(inner)
}

// ComposeContent builds prompt body content from already-rendered parts.
func ComposeContent(m ContentModel) string {
	if !m.HasField {
		return m.Prompt + "\n\n"
	}

	var b strings.Builder
	b.WriteString(m.Prompt)
	b.WriteString("\n\n")
	if m.Options != "" {
		b.WriteString(m.Options)
	}
	if m.FieldLabel != "" {
		b.WriteString(m.FieldLabel)
		b.WriteString("\n")
	}
	b.WriteString(m.FieldBody)
	if m.IncludeHints && m.Hints != "" {
		b.WriteString("\n\n")
		b.WriteString(m.Hints)
	}
	return b.String()
}

// RenderChoiceGroups renders labeled groups of mutually exclusive choices.
func RenderChoiceGroups(
	groups []ChoiceGroup,
	selected []int,
	activeGroup int,
	focus bool,
	styles ChoiceGroupStyles,
) string {
	choiceGap := styles.ChoiceGap
	if choiceGap == "" {
		choiceGap = "  "
	}
	groupGap := styles.GroupGap
	if groupGap == "" {
		groupGap = "\n\n"
	}

	var b strings.Builder
	for i, group := range groups {
		if i > 0 {
			b.WriteString(groupGap)
		}
		b.WriteString(styles.Label.Render(group.Label))
		b.WriteString("\n")

		for j, choice := range group.Choices {
			if j > 0 {
				b.WriteString(choiceGap)
			}
			isSelected := i < len(selected) && selected[i] == j
			isActive := focus && i == activeGroup
			switch {
			case isSelected && isActive:
				b.WriteString(styles.SelectedActive.Render(choice.Label))
			case isSelected:
				b.WriteString(styles.Selected.Render(choice.Label))
			case isActive:
				b.WriteString(styles.Active.Render(choice.Label))
			default:
				b.WriteString(styles.Inactive.Render(choice.Label))
			}
		}
	}
	if len(groups) > 0 {
		b.WriteString("\n\n")
	}
	return b.String()
}

// RenderHintLines renders one or more lines of key/text hints.
func RenderHintLines(lines [][]Hint, gap string, keyStyle, textStyle lg.Style) string {
	if gap == "" {
		gap = "   "
	}

	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := make([]string, 0, len(line))
		for _, hint := range line {
			parts = append(parts, keyStyle.Render(hint.Key)+" "+textStyle.Render(hint.Text))
		}
		rendered = append(rendered, strings.Join(parts, gap))
	}
	return strings.Join(rendered, gap)
}

// CenterRow left-pads a row so it appears centered within the given width.
func CenterRow(row string, width int) string {
	if pad := (width - lg.Width(row)) / 2; pad > 0 { //nolint:mnd // centering
		return strings.Repeat(" ", pad) + row
	}
	return row
}

// FocusOptionsCursor returns the cursor position to use when entering options.
func FocusOptionsCursor(optionCount int, reverse bool) int {
	if optionCount <= 0 {
		return 0
	}
	if reverse {
		return optionCount - 1
	}
	return 0
}

// NavigateOptions computes the next options cursor or whether focus should
// move from the option list into the input field.
func NavigateOptions(cursor, optionCount int, hasInput bool, dir NavDirection) NavResult {
	if optionCount <= 0 {
		return NavResult{}
	}

	switch dir {
	case NavDown:
		if hasInput && cursor == optionCount-1 {
			return NavResult{Cursor: cursor, MoveToInput: true}
		}
		return NavResult{Cursor: min(cursor+1, optionCount-1)}
	case NavUp:
		if hasInput && cursor == 0 {
			return NavResult{Cursor: cursor, MoveToInput: true}
		}
		return NavResult{Cursor: max(cursor-1, 0)}
	case NavTab:
		if hasInput && cursor == optionCount-1 {
			return NavResult{Cursor: cursor, MoveToInput: true}
		}
		return NavResult{Cursor: (cursor + 1) % optionCount}
	case NavShiftTab:
		if hasInput && cursor == 0 {
			return NavResult{Cursor: cursor, MoveToInput: true}
		}
		return NavResult{Cursor: (cursor - 1 + optionCount) % optionCount}
	default:
		return NavResult{Cursor: cursor}
	}
}

// FocusLineOffset returns the content line offset of the focused element
// within prompt content laid out by ComposeContent and RenderChoiceGroups.
//
// When optionFocus is true and optionCount > 0, it returns the line of the
// option at optionCursor. Otherwise, if hasInput is true, it returns the
// line where the input field begins. Returns -1 when there is nothing to
// scroll to.
func FocusLineOffset(optionFocus bool, optionCursor, optionCount int, hasInput bool) int {
	// These constants mirror the layout produced by ComposeContent (prompt +
	// "\n\n" = 3 lines) and RenderChoiceGroups (label + choices + "\n\n" = 3
	// lines per group).
	const promptLines = 3
	const linesPerOption = 3

	if optionFocus && optionCount > 0 {
		return max(0, promptLines+optionCursor*linesPerOption-1)
	}
	if hasInput {
		return max(0, promptLines+optionCount*linesPerOption)
	}
	return -1
}

// StepChoice moves or cycles the selected value within a choice group.
func StepChoice(current, choiceCount, delta int, wrap bool) int {
	if choiceCount <= 0 {
		return current
	}
	switch {
	case wrap:
		return (current + delta + choiceCount) % choiceCount
	case delta > 0:
		return min(current+delta, choiceCount-1)
	case delta < 0:
		return max(current+delta, 0)
	default:
		return current
	}
}

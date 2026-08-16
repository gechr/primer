package button_test

import (
	"strings"
	"testing"

	lg "charm.land/lipgloss/v2"
	"github.com/gechr/primer/button"
	"github.com/stretchr/testify/require"
)

func yesNo() button.Row {
	return button.Row{
		Buttons: []button.Button{
			{Label: "No", Focused: lg.NewStyle(), Blurred: lg.NewStyle()},
			{Label: "Yes", Focused: lg.NewStyle(), Blurred: lg.NewStyle()},
		},
		Focus: 1,
	}
}

func TestViewJoinsWithGap(t *testing.T) {
	t.Parallel()

	r := yesNo()
	require.Equal(t, "No  Yes", r.View())
	r.Gap = " | "
	require.Equal(t, "No | Yes", r.View())
}

func TestViewStylesFocusedButton(t *testing.T) {
	t.Parallel()

	r := yesNo()
	r.Buttons[1].Focused = lg.NewStyle().Padding(0, 1)
	require.Equal(t, "No   Yes ", r.View())
	// Moving focus swaps which button renders its focused style.
	r.Step(1)
	require.Equal(t, 0, r.Focus)
	require.Equal(t, "No  Yes", r.View())
}

func TestStepWraps(t *testing.T) {
	t.Parallel()

	r := yesNo()
	r.Step(1)
	require.Equal(t, 0, r.Focus)
	r.Step(-1)
	require.Equal(t, 1, r.Focus)
	r.Step(-1)
	require.Equal(t, 0, r.Focus)
}

func TestStepOnEmptyRowIsInert(t *testing.T) {
	t.Parallel()

	var r button.Row
	r.Step(1)
	require.Equal(t, 0, r.Focus)
	require.Empty(t, r.View())
	require.Equal(t, -1, r.Hit(0))
}

func TestHitMapsColumnsToButtons(t *testing.T) {
	t.Parallel()

	// Rendered row: "No  Yes" - No spans columns 0-1, the gap 2-3, Yes 4-6.
	r := yesNo()
	require.Equal(t, 0, r.Hit(0))
	require.Equal(t, 0, r.Hit(1))
	require.Equal(t, -1, r.Hit(2))
	require.Equal(t, -1, r.Hit(3))
	require.Equal(t, 1, r.Hit(4))
	require.Equal(t, 1, r.Hit(6))
	require.Equal(t, -1, r.Hit(7))
	require.Equal(t, -1, r.Hit(-1))
}

func TestHitCountsStylePadding(t *testing.T) {
	t.Parallel()

	// Padding widens the pressable surface: " Yes " spans five columns.
	r := yesNo()
	r.Buttons[1].Focused = lg.NewStyle().Padding(0, 1)
	require.Equal(t, "No   Yes ", r.View())
	require.Equal(t, 1, r.Hit(4))
	require.Equal(t, 1, r.Hit(8))
	require.Equal(t, -1, r.Hit(9))
}

func TestHitIgnoresAnsiInStyledWidths(t *testing.T) {
	t.Parallel()

	r := yesNo()
	r.Buttons[0].Blurred = lg.NewStyle().Bold(true)
	// The bold escape codes must not shift the hit columns.
	require.Equal(t, 1, strings.Count(r.View(), "No"))
	require.Equal(t, 0, r.Hit(1))
	require.Equal(t, 1, r.Hit(4))
}

package dialog_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gechr/primer/dialog"
	"github.com/stretchr/testify/require"
)

func TestConfirm(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		key           tea.KeyPressMsg
		wantResult    dialog.Result
		wantConfirmed bool
	}{
		{
			name:          "lowercase y submits confirmed",
			key:           tea.KeyPressMsg{Text: "y", Code: 'y'},
			wantResult:    dialog.ResultSubmit,
			wantConfirmed: true,
		},
		{
			name:          "uppercase y submits confirmed",
			key:           tea.KeyPressMsg{Text: "Y", Code: 'y', Mod: tea.ModShift},
			wantResult:    dialog.ResultSubmit,
			wantConfirmed: true,
		},
		{
			name:          "enter submits confirmed",
			key:           tea.KeyPressMsg{Code: tea.KeyEnter},
			wantResult:    dialog.ResultSubmit,
			wantConfirmed: true,
		},
		{
			name:          "lowercase n closes not confirmed",
			key:           tea.KeyPressMsg{Text: "n", Code: 'n'},
			wantResult:    dialog.ResultClose,
			wantConfirmed: false,
		},
		{
			name:          "uppercase n closes not confirmed",
			key:           tea.KeyPressMsg{Text: "N", Code: 'n', Mod: tea.ModShift},
			wantResult:    dialog.ResultClose,
			wantConfirmed: false,
		},
		{
			name:          "esc closes not confirmed",
			key:           tea.KeyPressMsg{Code: tea.KeyEscape},
			wantResult:    dialog.ResultClose,
			wantConfirmed: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := dialog.NewConfirm("delete everything?")
			next, cmd, res := c.Update(tc.key)
			require.Equal(t, tc.wantResult, res)
			require.Nil(t, cmd)
			// Update returns the same pointer so the caller reads the outcome off
			// the value the Stack pops.
			require.Same(t, c, next)
			require.Equal(t, tc.wantConfirmed, c.Confirmed())
		})
	}
}

func TestConfirmDefaults(t *testing.T) {
	t.Parallel()

	t.Run("defaults to not confirmed", func(t *testing.T) {
		t.Parallel()

		require.False(t, dialog.NewConfirm("proceed?").Confirmed())
	})

	t.Run("an unrelated key keeps it open and undecided", func(t *testing.T) {
		t.Parallel()

		c := dialog.NewConfirm("proceed?")
		_, _, res := c.Update(tea.KeyPressMsg{Text: "x", Code: 'x'})
		require.Equal(t, dialog.ResultNone, res)
		require.False(t, c.Confirmed())
	})

	t.Run("a non-key message is ignored", func(t *testing.T) {
		t.Parallel()

		c := dialog.NewConfirm("proceed?")
		_, _, res := c.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		require.Equal(t, dialog.ResultNone, res)
	})

	t.Run("the prompt is rendered as the body", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "proceed?", dialog.NewConfirm("proceed?").Content(40))
	})
}

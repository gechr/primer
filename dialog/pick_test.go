package dialog_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/dialog"
	"github.com/stretchr/testify/require"
)

func pickRows() []string {
	return []string{"alpha", "beta", "gamma"}
}

func TestPick(t *testing.T) {
	t.Parallel()

	t.Run("open shows every row", func(t *testing.T) {
		t.Parallel()

		p := dialog.NewPick("choose one", pickRows())
		require.Equal(t, "alpha\nbeta\ngamma", ansi.Strip(p.Content(40)))
	})

	t.Run("the shell renders the title", func(t *testing.T) {
		t.Parallel()

		// The list draws no title of its own, so Pick hands it to the Shell's
		// heading instead.
		require.Equal(t, "choose one", dialog.NewPick("choose one", pickRows()).Title())
	})

	t.Run("filter narrows to the match", func(t *testing.T) {
		t.Parallel()

		p := dialog.NewPick("choose one", pickRows())
		_, _, res := p.Update(tea.KeyPressMsg{Text: "bet"})
		require.Equal(t, dialog.ResultNone, res, "typing resolved the dialog")

		view := ansi.Strip(p.Content(40))
		lines := strings.Split(view, "\n")
		require.Equal(
			t,
			"beta",
			lines[len(lines)-1],
			"filtered view should keep only the match:\n%s",
			view,
		)

		idx, ok := p.Selected()
		require.True(t, ok)
		require.Equal(t, 1, idx, "selection should map back to the original row index")
	})

	t.Run("enter submits the highlighted row", func(t *testing.T) {
		t.Parallel()

		p := dialog.NewPick("choose one", pickRows())
		// Move the cursor down one so the choice is not just the default first
		// row, proving the submission reflects navigation.
		p.Update(tea.KeyPressMsg{Code: tea.KeyDown})

		next, _, res := p.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		require.Equal(t, dialog.ResultSubmit, res)
		require.Same(t, p, next)

		idx, ok := p.Selected()
		require.True(t, ok)
		require.Equal(t, 1, idx)
	})

	t.Run("esc closes without a selection commitment", func(t *testing.T) {
		t.Parallel()

		p := dialog.NewPick("choose one", pickRows())
		_, _, res := p.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		require.Equal(t, dialog.ResultClose, res)
	})

	t.Run("the cursor row feeds the scroll hint", func(t *testing.T) {
		t.Parallel()

		p := dialog.NewPick("choose one", pickRows())
		p.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		top, height, ok := p.ScrollTo()
		require.True(t, ok)
		require.Equal(t, 1, top)
		require.Equal(t, 1, height)
	})
}

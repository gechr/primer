package input_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/input"
	"github.com/stretchr/testify/require"
)

func testEditorStyles() input.EditorStyles {
	return input.EditorStyles{
		BlurredText: lg.NewStyle(),
		Counter:     lg.NewStyle(),
		Dirty:       lg.NewStyle(),
		DimLabel:    lg.NewStyle(),
		FocusedText: lg.NewStyle(),
		Header:      lg.NewStyle(),
		HelpKey:     lg.NewStyle(),
		HelpText:    lg.NewStyle(),
		Label:       lg.NewStyle(),
	}
}

func keyPress(code rune, mod tea.KeyMod, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code, Mod: mod, Text: text})
}

func newTestEditor(opts ...input.EditorOption) input.Editor {
	return input.NewEditor([]input.EditorEntry{
		{Label: "first", Title: "one"},
		{Label: "second", Title: "two"},
	}, opts...)
}

func strippedView(m input.Editor) string {
	return ansi.Strip(m.View().Content)
}

func intField(v any, name string) int {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return int(rv.FieldByName(name).Int())
}

func funcFieldIsNil(v any, name string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return rv.FieldByName(name).IsNil()
}

func applyCmd(t *testing.T, m input.Editor, cmd tea.Cmd) input.Editor {
	t.Helper()
	require.NotNil(t, cmd)
	msg := cmd()
	updated, next := m.Update(msg)
	require.Nil(t, next)
	em, ok := updated.(input.Editor)
	require.True(t, ok)
	return em
}

func TestNewEditorAppliesOptionsAndInitialState(t *testing.T) {
	t.Parallel()

	styles := testEditorStyles()
	m := input.NewEditor(
		[]input.EditorEntry{
			{Label: "first", Title: "one"},
			{Label: "second", Title: "two"},
		},
		input.WithEditorWidth(42),
		input.WithEditorBodyMinHeight(5),
		input.WithEditorStyles(styles),
		input.WithBodyFetch(func(int) (string, error) { return "body", nil }),
	)

	require.Equal(t, 42, intField(m, "width"))
	require.Equal(t, 5, intField(m, "bodyMin"))
	require.False(t, funcFieldIsNil(m, "fetchBody"))
	require.Equal(t, 0, intField(m, "current"))
	require.NotNil(t, m.View().Cursor)

	results := m.Results()
	require.Equal(t, []input.EditorResult{
		{Label: "first", Title: "one", Body: "", Changed: false},
		{Label: "second", Title: "two", Body: "", Changed: false},
	}, results)
}

func TestEditorInitAndFetchBodyCmd(t *testing.T) {
	t.Parallel()

	require.Nil(t, input.NewEditor([]input.EditorEntry{{Label: "only", Title: "one"}}).Init())

	success := input.NewEditor(
		[]input.EditorEntry{{Label: "only", Title: "one"}},
		input.WithBodyFetch(
			func(int) (string, error) { return "loaded body", nil },
		),
	)
	updated, next := success.Update(success.Init()())
	require.Nil(t, next)
	em, ok := updated.(input.Editor)
	require.True(t, ok)
	require.Equal(t, "loaded body", em.Results()[0].Body)

	failing := input.NewEditor(
		[]input.EditorEntry{{Label: "only", Title: "one"}},
		input.WithBodyFetch(
			func(int) (string, error) { return "", errors.New("boom") },
		),
	)
	updated, next = failing.Update(failing.Init()())
	require.Nil(t, next)
	em, ok = updated.(input.Editor)
	require.True(t, ok)
	require.Empty(t, em.Results()[0].Body)
	require.NotEmpty(t, strippedView(em))
}

func TestEditorWindowResizeAdjustsBodyHeight(t *testing.T) {
	t.Parallel()

	m := input.NewEditor([]input.EditorEntry{{Label: "only", Title: "one"}}, input.WithBodyFetch(
		func(int) (string, error) { return "line 1\nline 2\nline 3\nline 4", nil },
	))
	m = applyCmd(t, m, m.Init())

	updated, next := m.Update(tea.WindowSizeMsg{Width: 80, Height: 0})
	require.Nil(t, next)
	em, ok := updated.(input.Editor)
	require.True(t, ok)

	updated, next = em.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	require.Nil(t, next)
	em, ok = updated.(input.Editor)
	require.True(t, ok)

	require.Equal(t, "line 1\nline 2\nline 3\nline 4", em.Results()[0].Body)
	require.NotEmpty(t, strippedView(em))
}

func TestEditorUpdateBranches(t *testing.T) {
	t.Parallel()

	t.Run("title input and submission", func(t *testing.T) {
		t.Parallel()

		m := newTestEditor()
		updated, next := m.Update(keyPress('a', 0, "a"))
		_ = next
		em, ok := updated.(input.Editor)
		require.True(t, ok)
		require.Equal(t, "onea", em.Results()[0].Title)

		updated, next = em.Update(keyPress('s', tea.ModCtrl, ""))
		_ = next
		_, ok = updated.(input.Editor)
		require.True(t, ok)
	})

	t.Run("body input", func(t *testing.T) {
		t.Parallel()

		m := newTestEditor()
		updated, next := m.Update(keyPress(tea.KeyTab, 0, ""))
		_ = next
		em, ok := updated.(input.Editor)
		require.True(t, ok)

		updated, next = em.Update(keyPress('b', 0, "b"))
		_ = next
		em, ok = updated.(input.Editor)
		require.True(t, ok)
		require.Equal(t, "b", em.Results()[0].Body)
		require.True(t, em.Results()[0].Changed)
	})

	t.Run("abort keys", func(t *testing.T) {
		t.Parallel()

		for name, msg := range map[string]tea.Msg{
			"esc":    keyPress(tea.KeyEsc, 0, ""),
			"ctrl+c": keyPress('c', tea.ModCtrl, ""),
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				m := newTestEditor()
				updated, next := m.Update(msg)
				_ = next
				_, ok := updated.(input.Editor)
				require.True(t, ok)
			})
		}
	})

	t.Run("navigation and field switching", func(t *testing.T) {
		t.Parallel()

		m := input.NewEditor([]input.EditorEntry{
			{Label: "first", Title: "one"},
			{Label: "second", Title: "two"},
		}, input.WithBodyFetch(func(index int) (string, error) {
			return fmt.Sprintf("body %d", index), nil
		}))

		titleCursor := m.View().Cursor
		require.NotNil(t, titleCursor)

		updated, next := m.Update(keyPress(tea.KeyTab, 0, ""))
		_ = next
		em, ok := updated.(input.Editor)
		require.True(t, ok)

		bodyCursor := em.View().Cursor
		require.NotNil(t, bodyCursor)
		require.Positive(t, bodyCursor.Y-titleCursor.Y)

		updated, next = em.Update(keyPress(tea.KeyTab, tea.ModShift, ""))
		_ = next
		em, ok = updated.(input.Editor)
		require.True(t, ok)
		require.NotNil(t, em.View().Cursor)
		require.Equal(t, titleCursor.Y, em.View().Cursor.Y)

		updated, next = em.Update(keyPress('n', tea.ModCtrl, ""))
		_ = next
		em, ok = updated.(input.Editor)
		require.True(t, ok)
		require.Equal(t, 1, intField(em, "current"))

		updated, next = em.Update(keyPress('p', tea.ModCtrl, ""))
		_ = next
		em, ok = updated.(input.Editor)
		require.True(t, ok)
		require.Equal(t, 0, intField(em, "current"))
	})
}

func TestEditorResults(t *testing.T) {
	t.Parallel()

	m := input.NewEditor([]input.EditorEntry{
		{Label: "first", Title: "one"},
		{Label: "second", Title: "two"},
	})

	updated, next := m.Update(keyPress('a', 0, "a"))
	_ = next
	em, ok := updated.(input.Editor)
	require.True(t, ok)

	updated, next = em.Update(keyPress(tea.KeyTab, 0, ""))
	_ = next
	em, ok = updated.(input.Editor)
	require.True(t, ok)

	updated, next = em.Update(keyPress('b', 0, "b"))
	_ = next
	em, ok = updated.(input.Editor)
	require.True(t, ok)

	got := em.Results()
	require.Equal(t, []input.EditorResult{
		{Label: "first", Title: "onea", Body: "b", Changed: true},
		{Label: "second", Title: "two", Body: "", Changed: false},
	}, got)
}

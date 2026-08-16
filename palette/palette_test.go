package palette_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/palette"
	"github.com/stretchr/testify/require"
)

// entries builds a palette entry per name; Desc and Key are derived so a row
// carries all three fields without the tests spelling them out.
func entries(names ...string) []palette.Entry {
	out := make([]palette.Entry, len(names))
	for i, n := range names {
		out[i] = palette.Entry{Name: n, Desc: n + " desc", Key: n[:1]}
	}
	return out
}

// press feeds text to the palette as a single key press, the same shape the
// input substrate delivers typed runes.
func press(m *palette.Model, text string) {
	m.Update(tea.KeyPressMsg{Text: text})
}

// visibleNames walks the cursor through the visible rows and collects their
// names in order - a readable stand-in for the unexported matches slice.
func visibleNames(m *palette.Model) []string {
	// Clamp to the top first; the cursor may sit anywhere.
	for range 100 {
		m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	}
	var out []string
	seen := make(map[string]bool)
	for {
		e, ok := m.Selected()
		if !ok || seen[e.Name] {
			break
		}
		seen[e.Name] = true
		out = append(out, e.Name)
		m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	return out
}

func TestCursorLineTracksSelection(t *testing.T) {
	t.Parallel()

	m := palette.New("Commands", entries("transition", "comment", "assign"), palette.Styles{})
	// Title and query occupy the first two lines, so the first row sits at 2.
	require.Equal(t, 2, m.CursorLine())

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, 4, m.CursorLine())

	// A query that matches nothing has no row to point at.
	press(&m, "zzz")
	require.Equal(t, 0, m.CursorLine())
}

func TestEmptyQueryShowsAll(t *testing.T) {
	t.Parallel()

	m := palette.New("Commands", entries("transition", "comment", "assign"), palette.Styles{})
	require.Equal(t, []string{"transition", "comment", "assign"}, visibleNames(&m))

	m = palette.New("Commands", entries("transition", "comment", "assign"), palette.Styles{})
	sel, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, "transition", sel.Name, "cursor should start on the first entry")
}

func TestFuzzyQueryNarrows(t *testing.T) {
	t.Parallel()

	all := []string{"transition", "comment", "assign", "attach"}
	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{"contiguous prefix", "tran", []string{"transition"}},
		{"non-contiguous subsequence", "tsn", []string{"transition"}},
		{"shared subsequence keeps input order", "a", []string{"transition", "assign", "attach"}},
		{"uppercase stays case-insensitive", "COMMENT", []string{"comment"}},
		{"no subsequence matches nothing", "zzz", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := palette.New("t", entries(all...), palette.Styles{})
			press(&m, tt.query)
			require.Equal(t, tt.want, visibleNames(&m))
		})
	}
}

func TestCursorUpDownClamps(t *testing.T) {
	t.Parallel()

	m := palette.New("t", entries("one", "two", "three"), palette.Styles{})

	// Down past the end clamps on the last entry.
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	sel, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, "three", sel.Name)

	// Up past the start clamps on the first entry.
	for range 5 {
		m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	}
	sel, ok = m.Selected()
	require.True(t, ok)
	require.Equal(t, "one", sel.Name)
}

func TestCtrlNCtrlPNavigate(t *testing.T) {
	t.Parallel()

	m := palette.New("t", entries("one", "two", "three"), palette.Styles{})
	m.Update(tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'n'})
	sel, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, "two", sel.Name)

	m.Update(tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'p'})
	sel, ok = m.Selected()
	require.True(t, ok)
	require.Equal(t, "one", sel.Name)
}

func TestBackspaceRestoresQuery(t *testing.T) {
	t.Parallel()

	m := palette.New("t", entries("comment", "commit"), palette.Styles{})
	press(&m, "comme") // only "comment" is a subsequence
	require.Equal(t, []string{"comment"}, visibleNames(&m))

	// Backspace twice widens "comme" -> "com", which both entries match.
	m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	require.Equal(t, "com", m.Query())
	require.Equal(t, []string{"comment", "commit"}, visibleNames(&m))
}

func TestSelectedTracksCursorAndReportsEmpty(t *testing.T) {
	t.Parallel()

	m := palette.New("t", entries("transition", "comment"), palette.Styles{})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	sel, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, "comment", sel.Name)
	require.Equal(t, "c", sel.Key)

	press(&m, "zzz")
	_, ok = m.Selected()
	require.False(t, ok)

	lines := strings.Split(ansi.Strip(m.View()), "\n")
	require.Equal(t, "no commands match", lines[len(lines)-1])
}

func TestQueryEchoesTypedText(t *testing.T) {
	t.Parallel()

	m := palette.New("t", entries("transition"), palette.Styles{})
	require.Empty(t, m.Query())

	press(&m, "tr")
	press(&m, "an")
	require.Equal(t, "tran", m.Query())
}

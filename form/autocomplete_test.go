package form_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gechr/primer/form"
	"github.com/stretchr/testify/require"
)

// sug builds a suggestion whose value and label are equal - the common case for
// plain value lists.
func sug(s string) form.Suggestion { return form.Suggestion{Value: s, Label: s} }

// mentionForm is a one-area form completing @-mentions from a fixed set.
func mentionForm(fetched *[]string) form.Model {
	return form.New(form.Config{
		ConfirmDiscard: true,
		Fields: []form.FieldSpec{{
			Multiline: true,
			Autocomplete: &form.Autocomplete{
				Trigger: '@',
				Fetch: func(query string) []form.Suggestion {
					if fetched != nil {
						*fetched = append(*fetched, query)
					}
					var out []form.Suggestion
					for _, name := range []string{"alice", "alan", "bob"} {
						if strings.HasPrefix(name, query) {
							out = append(out, sug(name))
						}
					}
					return out
				},
			},
		}},
		Width: 40,
	})
}

// typeAndFetch types s and runs any fetch command the keystroke produced,
// feeding the resulting SuggestionsMsg back in - the owner's Update loop in
// miniature.
func typeAndFetch(t *testing.T, m *form.Model, s string) {
	t.Helper()
	cmd, ev, _ := m.Update(tea.KeyPressMsg{Text: s})
	require.Equal(t, form.EventNone, ev)
	drain(m, cmd)
}

// drain executes a command tree (Update wraps the field cmd and the fetch cmd
// in a batch) and routes any SuggestionsMsg back into the form.
func drain(m *form.Model, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, c := range msg {
			drain(m, c)
		}
	case form.SuggestionsMsg:
		m.Update(msg)
	}
}

// TestBareTokenCompletion pins bare mode (no trigger rune) with a custom
// boundary: each comma-separated entry completes alone and acceptance swaps
// the token without a prefix.
func TestBareTokenCompletion(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{{
			Autocomplete: &form.Autocomplete{
				IsBoundary: func(r rune) bool { return r == ',' || r == ' ' },
				Fetch: func(query string) []form.Suggestion {
					var out []form.Suggestion
					for _, name := range []string{"issues", "epics", "search"} {
						if strings.HasPrefix(name, query) {
							out = append(out, sug(name))
						}
					}
					return out
				},
			},
		}},
		Width: 40,
	})
	typeAndFetch(t, &m, "issues, ep")
	// The bare token after the comma armed the list: "epics" renders in it.
	require.Equal(t, 1, shows(m.View(), "epics"))
	press(t, &m, enter)
	require.Equal(t, "issues, epics", m.Value(0))
}

func TestSuggestionsFetchRenderAccept(t *testing.T) {
	t.Parallel()

	m := mentionForm(nil)
	typeAndFetch(t, &m, "ping ")
	typeAndFetch(t, &m, "@a")
	view := m.View()
	require.Equal(t, 1, shows(view, "alice"))
	require.Equal(t, 1, shows(view, "alan"))
	// Move to the second suggestion and accept it.
	press(t, &m, keyDn)
	require.Equal(t, form.EventNone, press(t, &m, enter))
	require.Equal(t, "ping @alan", m.Value(0))
	// The list closed on acceptance: the unpicked candidate is gone.
	require.Equal(t, 0, shows(m.View(), "alice"))
}

// TestSuggestionValueDiffersFromLabel pins the {Value,Label} split: the list
// renders the human-readable Label while acceptance inserts the opaque Value.
func TestSuggestionValueDiffersFromLabel(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{{
			Autocomplete: &form.Autocomplete{
				Trigger: '@',
				Fetch: func(string) []form.Suggestion {
					return []form.Suggestion{{Value: "acc-123", Label: "Ann Example"}}
				},
			},
		}},
		Width: 40,
	})
	typeAndFetch(t, &m, "@a")
	v := m.View()
	require.Equal(t, 1, shows(v, "Ann Example"))
	require.Equal(t, 0, shows(v, "acc-123"))
	press(t, &m, enter)
	require.Equal(t, "@acc-123", m.Value(0))
}

func TestEnterAcceptsInsteadOfNewline(t *testing.T) {
	t.Parallel()

	m := mentionForm(nil)
	typeAndFetch(t, &m, "@b")
	press(t, &m, enter)
	require.Equal(t, 0, shows(m.Value(0), "\n"))
}

func TestEscDismissesListWithoutGuard(t *testing.T) {
	t.Parallel()

	m := mentionForm(nil)
	typeAndFetch(t, &m, "@a")
	require.Equal(t, 1, shows(m.View(), "alice"))
	require.Equal(t, form.EventNone, press(t, &m, esc))
	// The list is gone but the draft survives.
	require.Equal(t, 0, shows(m.View(), "alice"))
	// The next esc reaches the form proper and asks the owner to confirm.
	require.Equal(t, form.EventConfirmDiscard, press(t, &m, esc))
}

func TestStaleSuggestionsDrop(t *testing.T) {
	t.Parallel()

	m := mentionForm(nil)
	typeAndFetch(t, &m, "@a")
	// A slow response for an older query arrives after the token moved on.
	m.Update(form.SuggestionsMsg{Field: 0, Query: "zz", Items: []form.Suggestion{sug("zoe")}})
	require.Equal(t, 0, shows(m.View(), "zoe"))
}

func TestQueryChangeRefetches(t *testing.T) {
	t.Parallel()

	var queries []string
	m := mentionForm(&queries)
	typeAndFetch(t, &m, "@a")
	typeAndFetch(t, &m, "l")
	require.Equal(t, []string{"a", "al"}, queries)
	require.Equal(t, 0, shows(m.View(), "bob"))
}

func TestNoFetchBelowMinQuery(t *testing.T) {
	t.Parallel()

	var queries []string
	m := form.New(form.Config{
		Fields: []form.FieldSpec{{
			Autocomplete: &form.Autocomplete{
				Trigger:  '@',
				MinQuery: 2,
				Fetch: func(q string) []form.Suggestion {
					queries = append(queries, q)
					return nil
				},
			},
		}},
		Width: 40,
	})
	typeAndFetch(t, &m, "@a")
	require.Empty(t, queries)
	typeAndFetch(t, &m, "l")
	require.Equal(t, []string{"al"}, queries)
}

func TestSuggestionListCapped(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{{
			Autocomplete: &form.Autocomplete{
				Trigger: '@',
				Fetch: func(string) []form.Suggestion {
					return []form.Suggestion{
						sug("a1"), sug("a2"), sug("a3"), sug("a4"), sug("a5"), sug("a6"), sug("a7"),
					}
				},
			},
		}},
		Width: 40,
	})
	typeAndFetch(t, &m, "@x")
	// The rendered list caps at five items: the fifth shows, the sixth does not.
	v := m.View()
	require.Equal(t, 1, shows(v, "a5"))
	require.Equal(t, 0, shows(v, "a6"))
}

func TestLineFieldMidTextAcceptance(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{{
			Autocomplete: &form.Autocomplete{
				Trigger: '@',
				Fetch:   func(string) []form.Suggestion { return []form.Suggestion{sug("alice")} },
			},
		}},
		Width: 40,
	})
	typeAndFetch(t, &m, "@a tail")
	// Walk the cursor back to sit right after "@a" so the token re-arms
	// mid-text; the last arrow's sync must fire the fetch again.
	for range len(" tail") {
		cmd, _, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
		drain(&m, cmd)
	}
	require.Equal(t, 1, shows(m.View(), "alice"))
	press(t, &m, enter)
	require.Equal(t, "@alice tail", m.Value(0))
	// The cursor sits at the end of the completion, not the end of the line:
	// the next keystroke lands right after the accepted token.
	typeText(t, &m, "X")
	require.Equal(t, "@aliceX tail", m.Value(0))
}

func TestPasteDuringConfirmIsSwallowed(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		ConfirmDiscard: true,
		Fields:         []form.FieldSpec{{}},
		Width:          40,
	})
	typeAndFetch(t, &m, "draft")
	press(t, &m, esc) // dirty → confirming
	_, ev, consumed := m.Update(tea.PasteMsg{Content: "clipboard noise"})
	require.Equal(t, form.EventNone, ev)
	require.True(t, consumed)
	require.Equal(t, "draft", m.Value(0))
}

func TestFocusChangeClearsSuggestions(t *testing.T) {
	t.Parallel()

	m := form.New(form.Config{
		Fields: []form.FieldSpec{
			{Autocomplete: &form.Autocomplete{
				Trigger: '@',
				Fetch:   func(string) []form.Suggestion { return []form.Suggestion{sug("alice")} },
			}},
			{Optional: true},
		},
		Width: 40,
	})
	typeAndFetch(t, &m, "@a")
	require.Equal(t, 1, shows(m.View(), "alice"))
	// Tab would accept the suggestion (the list owns it); shift+tab moves focus
	// with the list up, and the suggestions must not follow it.
	press(t, &m, shTab)
	require.Equal(t, 0, shows(m.View(), "alice"))
}

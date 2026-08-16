package key_test

import (
	"testing"

	bkey "charm.land/bubbles/v2/key"
	"github.com/gechr/primer/key"
	"github.com/stretchr/testify/require"
)

// ownerMap is a minimal owner key map: the registry never owns bindings, it
// only indexes the owner's fields - Rebind must mutate them in place.
type ownerMap struct {
	Open bkey.Binding
	Quit bkey.Binding
}

func newOwner() (*ownerMap, key.Registry) {
	o := &ownerMap{
		Open: bkey.NewBinding(bkey.WithKeys("enter"), bkey.WithHelp("enter", "open")),
		Quit: bkey.NewBinding(bkey.WithKeys("q", "ctrl+c"), bkey.WithHelp("q", "quit")),
	}
	return o, key.Registry{"open": &o.Open, "quit": &o.Quit}
}

func TestRebindOverridesKeysAndKeepsDescription(t *testing.T) {
	t.Parallel()

	o, r := newOwner()
	require.NoError(t, r.Rebind(map[string][]string{"open": {"o", "O"}}))
	require.Equal(t, []string{"o", "O"}, o.Open.Keys())
	require.Equal(t, "o", o.Open.Help().Key)
	require.Equal(t, "open", o.Open.Help().Desc)
}

func TestRebindEmptyOverrideIgnored(t *testing.T) {
	t.Parallel()

	o, r := newOwner()
	require.NoError(t, r.Rebind(map[string][]string{"quit": {}}))
	require.Equal(t, []string{"q", "ctrl+c"}, o.Quit.Keys())
}

func TestRebindUnknownNameIsError(t *testing.T) {
	t.Parallel()

	_, r := newOwner()
	require.Error(t, r.Rebind(map[string][]string{"nope": {"z"}}))
}

func TestRebindUnknownNameMutatesNothing(t *testing.T) {
	t.Parallel()

	o, r := newOwner()
	err := r.Rebind(map[string][]string{"open": {"o"}, "nope": {"z"}})
	require.Error(t, err)
	// The valid override in the same call must not have landed: a failed
	// Rebind is atomic, whatever the map iteration order.
	require.Equal(t, []string{"enter"}, o.Open.Keys())
}

func TestNamesSorted(t *testing.T) {
	t.Parallel()

	_, r := newOwner()
	require.Equal(t, []string{"open", "quit"}, r.Names())
}

package key_test

import (
	"strings"
	"testing"

	bkey "charm.land/bubbles/v2/key"
	"github.com/gechr/primer/key"
	"github.com/stretchr/testify/require"
)

// TestReplayRoundTrips is the load-bearing guarantee behind key replay: every
// key string a binding can carry must rebuild into a KeyPressMsg whose
// String() is byte-for-byte that string, or key.Matches would never fire and a
// replayed selection would silently do nothing.
func TestReplayRoundTrips(t *testing.T) {
	t.Parallel()

	for _, ks := range []string{
		// Plain runes and named keys.
		"t", "q", "/", "?",
		"tab", "enter", "esc", "space", "backspace", "delete", "insert",
		"up", "down", "left", "right",
		"home", "end", "pgup", "pgdown",
		// Function keys generate in a loop; pin the edges and a middle value.
		"f1", "f5", "f12", "f63",
		// Rebinding is free-form, so modifier + named/function keys must
		// replay too. A naive first-rune fallback silently turned "ctrl+f5"
		// into "ctrl+f".
		"ctrl+f5", "shift+f1", "alt+f12",
		"ctrl+home", "alt+pgdown", "shift+tab",
		"ctrl+insert", "alt+delete", "ctrl+end",
		"ctrl+u", "alt+c",
	} {
		require.Equal(t, ks, key.Replay(ks).String())
		// Consumers match through key.Matches, so pin that path too.
		require.True(t, bkey.Matches(key.Replay(ks), bkey.NewBinding(bkey.WithKeys(ks))),
			"Replay(%q) does not match a binding on %q", ks, ks)
	}
}

// TestReplayUnknownModifiedKeyIsInert pins the fallback: an unknown modified
// key may fail to match its own (unsupported) key, but it must never drop the
// modifier - a bare "f99" or a truncated "ctrl+f" could fire a command bound
// to that literal.
func TestReplayUnknownModifiedKeyIsInert(t *testing.T) {
	t.Parallel()

	got := key.Replay("ctrl+f99").String()
	require.NotEqual(t, "f99", got)
	require.NotEqual(t, "ctrl+f", got)
	require.NotEqual(t, "f", got)
	require.True(t, strings.HasPrefix(got, "ctrl+"),
		"Replay(%q).String() = %q, want a ctrl+ prefix (modifier dropped)", "ctrl+f99", got)
}

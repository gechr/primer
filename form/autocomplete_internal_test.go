package form

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTriggerToken(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		before string
		query  string
		width  int
		ok     bool
	}{
		{name: "bare text", before: "hello", query: "", width: 0, ok: false},
		{name: "trigger at start", before: "@al", query: "al", width: 3, ok: true},
		{name: "trigger mid-sentence", before: "ping @bo", query: "bo", width: 3, ok: true},
		{name: "trigger inside word", before: "mail@ex", query: "", width: 0, ok: false},
		{name: "bare trigger", before: "@", query: "", width: 1, ok: true},
		{name: "space ends token", before: "@al done", query: "", width: 0, ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			query, width, ok := triggerToken(tc.before, '@', nil)
			require.Equal(t, tc.query, query)
			require.Equal(t, tc.width, width)
			require.Equal(t, tc.ok, ok)
		})
	}
}

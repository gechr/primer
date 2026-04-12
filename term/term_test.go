package term_test

import (
	"os"
	"testing"

	"github.com/gechr/primer/term"
	"github.com/stretchr/testify/require"
)

func TestIsNilFile(t *testing.T) {
	require.False(t, term.Is(nil))
}

func TestIsNonTerminalFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "non-terminal")
	require.NoError(t, err)
	defer f.Close()

	require.False(t, term.Is(f))
}

func TestWidthNilFile(t *testing.T) {
	require.Zero(t, term.Width(nil))
}

func TestWidthNonTerminalFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "non-terminal")
	require.NoError(t, err)
	defer f.Close()

	require.Zero(t, term.Width(f))
}

func TestSizeNilFile(t *testing.T) {
	width, height := term.Size(nil)

	require.Zero(t, width)
	require.Zero(t, height)
}

func TestSizeNonTerminalFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "non-terminal")
	require.NoError(t, err)
	defer f.Close()

	width, height := term.Size(f)
	require.Zero(t, width)
	require.Zero(t, height)
}

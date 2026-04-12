package human_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gechr/primer/human"
	"github.com/stretchr/testify/require"
)

func TestContractHomeSubpath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	require.Equal(
		t,
		"~/Documents/file.txt",
		human.ContractHome(filepath.Join(home, "Documents", "file.txt")),
	)
}

func TestContractHomeExact(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	require.Equal(t, "~", human.ContractHome(home))
}

func TestContractHomeUnrelated(t *testing.T) {
	require.Equal(t, "/tmp/other", human.ContractHome("/tmp/other"))
}

func TestExpandPathTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	require.Equal(t, home, human.ExpandPath("~"))
}

func TestExpandPathTildeSubpath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	require.Equal(t, filepath.Join(home, "Documents"), human.ExpandPath("~/Documents"))
}

func TestExpandPathEmpty(t *testing.T) {
	require.Empty(t, human.ExpandPath(""))
}

func TestExpandPathAbsolute(t *testing.T) {
	require.Equal(t, "/tmp/file", human.ExpandPath("/tmp/file"))
}

func TestExpandPathEnvVar(t *testing.T) {
	t.Setenv("TEST_PRIMER_DIR", "/opt/data")
	require.Equal(t, "/opt/data/file", human.ExpandPath("$TEST_PRIMER_DIR/file"))
}

func TestContractExpandRoundTrip(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	original := filepath.Join(home, "projects", "primer")
	contracted := human.ContractHome(original)
	expanded := human.ExpandPath(contracted)
	require.Equal(t, original, expanded)
}

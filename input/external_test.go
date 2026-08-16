package input_test

import (
	"testing"

	"github.com/gechr/primer/input"
	"github.com/stretchr/testify/require"
)

// No t.Parallel: t.Setenv is incompatible with parallel tests.
func TestExternalEditorPrecedence(t *testing.T) {
	t.Setenv("EDITOR", "vi")
	require.Equal(t, "nvim -f", input.ExternalEditor("nvim -f"))
	require.Equal(t, "vi", input.ExternalEditor(""))
	t.Setenv("EDITOR", "")
	require.Empty(t, input.ExternalEditor(""))
}

// No t.Parallel: t.Setenv is incompatible with parallel tests.
func TestExternalEditorCmdWithoutEditorReturnsError(t *testing.T) {
	t.Setenv("EDITOR", "")
	cmd := input.ExternalEditorCmd("", "comment:42", "")
	require.NotNil(t, cmd)
	msg, ok := cmd().(input.ExternalEditorFinishedMsg)
	require.True(t, ok)
	require.Equal(t, "comment:42", msg.ID)
	require.Error(t, msg.Err)
}

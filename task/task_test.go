package task_test

import (
	"testing"

	"github.com/gechr/primer/task"
	"github.com/stretchr/testify/require"
)

func TestManagerAcceptsLatestGenerationOnly(t *testing.T) {
	t.Parallel()

	m := task.New()
	const scope task.Scope = "issues"

	m.Start(task.Spec{Scope: scope})
	gen1 := m.Generation(scope)
	m.Start(task.Spec{Scope: scope})
	gen2 := m.Generation(scope)

	require.Greater(t, gen2, gen1)
	require.False(t, m.Accept(scope, gen1), "a superseded fetch must be dropped")
	require.True(t, m.Accept(scope, gen2))
}

func TestManagerScopesAreIndependent(t *testing.T) {
	t.Parallel()

	m := task.New()
	m.Start(task.Spec{Scope: "issues"})
	m.Start(task.Spec{Scope: "search"})

	require.Equal(t, uint64(1), m.Generation("issues"))
	require.Equal(t, uint64(1), m.Generation("search"))
	require.True(t, m.Accept("issues", 1))
	require.True(t, m.Accept("search", 1))
}

func TestManagerStartRunsAndReportsResult(t *testing.T) {
	t.Parallel()

	m := task.New()
	cmd := m.Start(task.Spec{
		Scope: "issues",
		Run:   func() (any, error) { return 42, nil },
	})
	require.NotNil(t, cmd)
	// Start records the generation immediately; executing the command produces
	// the FinishedMsg carrying that generation.
	require.Equal(t, uint64(1), m.Generation("issues"))

	fin, ok := cmd().(task.FinishedMsg)
	require.True(t, ok, "command must produce a FinishedMsg")
	require.Equal(t, task.FinishedMsg{Scope: "issues", Gen: 1, Result: 42, Err: nil}, fin)
}

func TestZeroValueManagerIsUsable(t *testing.T) {
	t.Parallel()

	var m task.Manager
	cmd := m.Start(task.Spec{Scope: "issues"})
	require.NotNil(t, cmd)
	require.Equal(t, uint64(1), m.Generation("issues"))
	require.True(t, m.Accept("issues", 1))
}

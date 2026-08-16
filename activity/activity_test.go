package activity_test

import (
	"errors"
	"testing"

	"github.com/gechr/primer/activity"
	"github.com/stretchr/testify/require"
)

func TestStartReturnsMonotonicIDsFromOne(t *testing.T) {
	t.Parallel()

	r := activity.New()
	for _, want := range []uint64{1, 2, 3} {
		require.Equal(t, want, r.Start("op"))
	}
}

func TestStartZeroValueRegistry(t *testing.T) {
	t.Parallel()

	// A zero-value Registry (New not called) still starts ids at 1.
	var r activity.Registry
	require.Equal(t, uint64(1), r.Start("op"))
}

func TestFinish(t *testing.T) {
	t.Parallel()

	r := activity.New()
	id := r.Start("creating issue in PROJ…")
	r.Finish(id, "PROJ-88 created", "PROJ-88")

	e, ok := r.Recent()
	require.True(t, ok)
	require.Equal(t, activity.Entry{
		ID:      id,
		Pending: "creating issue in PROJ…",
		Done:    "PROJ-88 created",
		Ref:     "PROJ-88",
		Status:  activity.StatusDone,
		Err:     nil,
	}, e)
}

func TestFinishResolvesInPlaceSameID(t *testing.T) {
	t.Parallel()

	r := activity.New()
	id1 := r.Start("first")
	id2 := r.Start("second")
	r.Finish(id1, "first done", "")

	log := r.Log()
	require.Len(t, log, 2)
	// Newest-first: id2 then id1; id1 resolved in place, ids unchanged.
	require.Equal(t, id2, log[0].ID)
	require.Equal(t, id1, log[1].ID)
	require.Equal(t, activity.StatusDone, log[1].Status)
	require.Equal(t, activity.StatusPending, log[0].Status)
}

func TestFail(t *testing.T) {
	t.Parallel()

	r := activity.New()
	id := r.Start("deleting PROJ-9")
	wantErr := errors.New("boom")
	r.Fail(id, wantErr)

	e, ok := r.Recent()
	require.True(t, ok)
	require.Equal(t, id, e.ID)
	require.Equal(t, activity.StatusFailed, e.Status)
	require.ErrorIs(t, e.Err, wantErr)
}

func TestUnknownIDIsNoOp(t *testing.T) {
	t.Parallel()

	r := activity.New()
	id := r.Start("op")
	// Neither of these should panic or alter the recorded entry.
	r.Finish(999, "nope", "NOPE")
	r.Fail(777, errors.New("nope"))

	e, ok := r.Recent()
	require.True(t, ok)
	require.Equal(t, activity.Entry{
		ID:      id,
		Pending: "op",
		Status:  activity.StatusPending,
	}, e)
}

func TestInFlightCountsOnlyPending(t *testing.T) {
	t.Parallel()

	r := activity.New()
	a := r.Start("a")
	r.Start("b")
	c := r.Start("c")
	require.Equal(t, 3, r.InFlight())

	r.Finish(a, "a done", "")
	r.Fail(c, errors.New("c failed"))
	require.Equal(t, 1, r.InFlight())
}

func TestRecent(t *testing.T) {
	t.Parallel()

	r := activity.New()
	_, ok := r.Recent()
	require.False(t, ok)

	r.Start("first")
	r.Start("second")
	e, ok := r.Recent()
	require.True(t, ok)
	require.Equal(t, "second", e.Pending, "recent should be newest (highest id)")
}

func TestLogNewestFirst(t *testing.T) {
	t.Parallel()

	r := activity.New()
	r.Start("first")
	r.Start("second")
	r.Start("third")

	pending := make([]string, 0, 3)
	for _, e := range r.Log() {
		pending = append(pending, e.Pending)
	}
	require.Equal(t, []string{"third", "second", "first"}, pending)
}

func TestLogCappedWithCap(t *testing.T) {
	t.Parallel()

	const limit = 10
	const extra = 5

	r := activity.New(activity.WithCap(limit))
	var lastID uint64
	for range limit + extra {
		lastID = r.Start("op")
	}

	log := r.Log()
	require.Len(t, log, limit)
	// The newest survive: log[0] is the very last Start, and the oldest
	// retained id is total-limit+1 (the first `extra` were dropped).
	require.Equal(t, lastID, log[0].ID)
	require.Equal(t, uint64(extra+1), log[len(log)-1].ID)
}

func TestReturnedEntriesAreCopies(t *testing.T) {
	t.Parallel()

	r := activity.New()
	id := r.Start("op")

	// Mutating an Entry returned by Recent must not touch registry state.
	e, ok := r.Recent()
	require.True(t, ok)
	e.Pending = "tampered"
	e.Status = activity.StatusFailed
	require.Equal(t, "tampered", e.Pending, "a returned Entry should be a freely-mutable copy")
	require.Equal(t, activity.StatusFailed, e.Status)

	// Mutating an Entry returned by Log must not either.
	log := r.Log()
	log[0].Done = "tampered"
	require.Equal(t, "tampered", log[0].Done)

	fresh, ok := r.Recent()
	require.True(t, ok)
	require.Equal(t, activity.Entry{
		ID:      id,
		Pending: "op",
		Status:  activity.StatusPending,
	}, fresh)
}

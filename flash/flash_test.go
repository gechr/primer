package flash_test

import (
	"testing"

	"github.com/gechr/primer/flash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetStoresMessage(t *testing.T) {
	t.Parallel()
	var s flash.State
	s.Set("saved", false)
	assert.Equal(t, "saved", s.Msg)
	assert.False(t, s.Err)
}

func TestSetStoresError(t *testing.T) {
	t.Parallel()
	var s flash.State
	s.Set("failed", true)
	assert.Equal(t, "failed", s.Msg)
	assert.True(t, s.Err)
}

func TestClearMatchingID(t *testing.T) {
	t.Parallel()
	var s flash.State
	clearMsg := s.Set("hello", false)
	require.True(t, s.Active())

	s.Clear(clearMsg)
	assert.False(t, s.Active())
	assert.Empty(t, s.Msg)
}

func TestClearStaleIDIsNoOp(t *testing.T) {
	t.Parallel()
	var s flash.State
	stale := s.Set("old", false)
	s.Set("new", false)

	s.Clear(stale)
	assert.True(t, s.Active())
	assert.Equal(t, "new", s.Msg)
}

func TestActiveReflectsState(t *testing.T) {
	t.Parallel()
	var s flash.State
	assert.False(t, s.Active())

	s.Set("hi", false)
	assert.True(t, s.Active())
}

func TestRapidSetsOnlyLatestSurvives(t *testing.T) {
	t.Parallel()
	var s flash.State
	s.Set("first", false)
	s.Set("second", false)
	clearMsg := s.Set("third", true)

	assert.Equal(t, "third", s.Msg)
	assert.True(t, s.Err)

	s.Clear(clearMsg)
	assert.False(t, s.Active())
}

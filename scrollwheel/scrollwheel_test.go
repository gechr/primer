package scrollwheel_test

import (
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/gechr/primer/scrollwheel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type target uint8

const (
	targetList target = iota + 1
	targetDiff
)

type stubModel struct{ target target }

func (stubModel) Init() tea.Cmd                         { return nil }
func (m stubModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (stubModel) View() tea.View                        { return tea.View{} }

func resolver(m tea.Model) (target, bool) {
	s, ok := m.(stubModel)
	if !ok || s.target == 0 {
		return 0, false
	}
	return s.target, true
}

func collect(ch chan tea.Msg, timeout time.Duration) []scrollwheel.Msg[target] {
	var msgs []scrollwheel.Msg[target]
	deadline := time.After(timeout)
	for {
		select {
		case raw := <-ch:
			if wm, ok := raw.(scrollwheel.Msg[target]); ok {
				msgs = append(msgs, wm)
			}
		case <-deadline:
			return msgs
		}
	}
}

func TestSameTargetAccumulation(t *testing.T) {
	t.Parallel()
	ch := make(chan tea.Msg, 10)
	c := scrollwheel.New(resolver, func(msg tea.Msg) { ch <- msg },
		scrollwheel.WithDelay(20*time.Millisecond))
	defer c.Stop()

	model := stubModel{target: targetList}
	down := tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelDown})

	require.Nil(t, c.Filter(model, down))
	require.Nil(t, c.Filter(model, down))
	require.Nil(t, c.Filter(model, down))

	msgs := collect(ch, 100*time.Millisecond)
	require.Len(t, msgs, 1)
	assert.Equal(t, targetList, msgs[0].Target)
	assert.Equal(t, 3, msgs[0].Delta)
}

func TestTargetChangeFlushesPrevious(t *testing.T) {
	t.Parallel()
	ch := make(chan tea.Msg, 10)
	c := scrollwheel.New(resolver, func(msg tea.Msg) { ch <- msg },
		scrollwheel.WithDelay(20*time.Millisecond))
	defer c.Stop()

	listModel := stubModel{target: targetList}
	diffModel := stubModel{target: targetDiff}
	down := tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelDown})

	require.Nil(t, c.Filter(listModel, down))
	require.Nil(t, c.Filter(listModel, down))
	// Switch target - should flush list immediately.
	require.Nil(t, c.Filter(diffModel, down))

	msgs := collect(ch, 100*time.Millisecond)
	require.Len(t, msgs, 2)

	assert.Equal(t, targetList, msgs[0].Target)
	assert.Equal(t, 2, msgs[0].Delta)

	assert.Equal(t, targetDiff, msgs[1].Target)
	assert.Equal(t, 1, msgs[1].Delta)
}

func TestStopCancelsPendingFlush(t *testing.T) {
	t.Parallel()
	ch := make(chan tea.Msg, 10)
	c := scrollwheel.New(resolver, func(msg tea.Msg) { ch <- msg },
		scrollwheel.WithDelay(50*time.Millisecond))

	model := stubModel{target: targetList}
	down := tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelDown})

	require.Nil(t, c.Filter(model, down))
	c.Stop()

	msgs := collect(ch, 100*time.Millisecond)
	assert.Empty(t, msgs)
}

func TestNonWheelMessagesPassThrough(t *testing.T) {
	t.Parallel()
	c := scrollwheel.New(resolver, func(tea.Msg) {},
		scrollwheel.WithDelay(20*time.Millisecond))
	defer c.Stop()

	model := stubModel{target: targetList}
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}

	result := c.Filter(model, keyMsg)
	assert.Equal(t, keyMsg, result)
}

func TestUnresolvableModelPassesThrough(t *testing.T) {
	t.Parallel()
	c := scrollwheel.New(resolver, func(tea.Msg) {},
		scrollwheel.WithDelay(20*time.Millisecond))
	defer c.Stop()

	model := stubModel{target: 0} // resolver returns false
	down := tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelDown})

	result := c.Filter(model, down)
	assert.NotNil(t, result)
}

func TestUpDeltaIsNegative(t *testing.T) {
	t.Parallel()
	ch := make(chan tea.Msg, 10)
	c := scrollwheel.New(resolver, func(msg tea.Msg) { ch <- msg },
		scrollwheel.WithDelay(20*time.Millisecond))
	defer c.Stop()

	model := stubModel{target: targetList}
	up := tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelUp})

	require.Nil(t, c.Filter(model, up))
	require.Nil(t, c.Filter(model, up))

	msgs := collect(ch, 100*time.Millisecond)
	require.Len(t, msgs, 1)
	assert.Equal(t, -2, msgs[0].Delta)
}

func TestConcurrentFilterCalls(t *testing.T) {
	t.Parallel()
	ch := make(chan tea.Msg, 100)
	c := scrollwheel.New(resolver, func(msg tea.Msg) { ch <- msg },
		scrollwheel.WithDelay(20*time.Millisecond))
	defer c.Stop()

	model := stubModel{target: targetList}
	down := tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelDown})

	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			c.Filter(model, down)
		})
	}
	wg.Wait()

	msgs := collect(ch, 200*time.Millisecond)
	total := 0
	for _, m := range msgs {
		total += m.Delta
	}
	assert.Equal(t, 20, total)
}

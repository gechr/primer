// Package flash provides state management for transient status messages
// in Bubble Tea applications.
//
// A typical TUI shows a brief "Approved #123" or "Error: ..." message
// that auto-clears after a few seconds. Both the message and a monotonic
// ID are tracked so that stale [ClearMsg] ticks never wipe a newer message.
//
// The package is pure state - the application owns the [tea.Tick] scheduling
// and the rendering. flash only manages the set/clear lifecycle.
package flash

// ClearMsg is sent after the flash duration expires. Pass it to
// [State.Clear] in the application's Update loop.
type ClearMsg struct{ id int }

// State tracks a transient status message with expiry.
type State struct {
	id  int
	Msg string
	Err bool
}

// Set updates the flash message, increments the internal ID and returns
// a [ClearMsg] for use with [tea.Tick]. The caller schedules the tick:
//
//	clear := m.flash.Set("Done", false)
//	return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg { return clear })
func (s *State) Set(msg string, isErr bool) ClearMsg {
	s.id++
	s.Msg = msg
	s.Err = isErr
	return ClearMsg{id: s.id}
}

// Clear resets the flash only when the [ClearMsg] ID matches the current
// flash. A stale clear (from an earlier Set) is silently ignored.
func (s *State) Clear(msg ClearMsg) {
	if msg.id != s.id {
		return
	}
	s.Msg = ""
	s.Err = false
}

// Active reports whether a flash message is currently showing.
func (s *State) Active() bool { return s.Msg != "" }

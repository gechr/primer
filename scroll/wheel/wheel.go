// Package wheel coalesces rapid mouse wheel events into batched scroll
// messages, preventing per-event Update + Render cycles from pegging CPU.
//
// The coalescer intercepts wheel events via a Bubble Tea filter, accumulates
// deltas over a short debounce window and flushes them as a single [Msg].
// A user-supplied [Resolver] maps the current model to an application-defined
// scroll target so the coalescer can route events correctly and flush
// immediately when the target changes mid-batch.
package wheel

import (
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
)

// DefaultDelay is the debounce window for coalescing wheel events.
const DefaultDelay = 12 * time.Millisecond

// Msg carries an accumulated scroll delta for a given target.
// Positive delta means scroll down, negative means scroll up.
type Msg[T comparable] struct {
	Target T
	Delta  int
}

// Resolver inspects the current model and returns the scroll target
// for wheel events. Return false when the model is not in a scrollable
// state; the wheel event will pass through to the normal update loop.
type Resolver[T comparable] func(tea.Model) (T, bool)

// Option configures a [Coalescer].
type Option func(*config)

type config struct {
	delay time.Duration
}

// WithDelay sets the debounce window. Default is [DefaultDelay].
func WithDelay(d time.Duration) Option {
	return func(c *config) { c.delay = d }
}

// Coalescer batches rapid mouse wheel events and flushes them as a
// single [Msg] after a short debounce window.
type Coalescer[T comparable] struct {
	resolve Resolver[T]
	send    func(tea.Msg)
	delay   time.Duration

	mu     sync.Mutex
	active bool
	target T
	delta  int
	timer  *time.Timer
	zero   T // cached zero value for comparisons
}

// New creates a coalescer. The send function is typically
// [tea.Program.Send]; it must be safe to call from any goroutine.
func New[T comparable](resolve Resolver[T], send func(tea.Msg), opts ...Option) *Coalescer[T] {
	cfg := config{delay: DefaultDelay}
	for _, o := range opts {
		o(&cfg)
	}
	return &Coalescer[T]{
		resolve: resolve,
		send:    send,
		delay:   cfg.delay,
	}
}

// Filter is a [tea.WithFilter]-compatible function. It swallows mouse
// wheel events directed at a scrollable view, accumulates their delta
// and returns nil so Bubble Tea skips the normal Update + Render cycle.
// Non-wheel messages pass through unchanged.
func (c *Coalescer[T]) Filter(model tea.Model, msg tea.Msg) tea.Msg {
	wm, ok := msg.(tea.MouseWheelMsg)
	if !ok {
		return msg
	}

	var delta int
	switch wm.Button {
	case tea.MouseWheelDown:
		delta = 1
	case tea.MouseWheelUp:
		delta = -1
	default:
		return msg
	}

	target, ok := c.resolve(model)
	if !ok {
		return msg
	}

	c.enqueue(target, delta)
	return nil
}

func (c *Coalescer[T]) enqueue(target T, delta int) {
	var immediate *Msg[T]
	startTimer := false

	c.mu.Lock()
	switch {
	case !c.active:
		c.active = true
		c.target = target
		c.delta = delta
		startTimer = true
	case c.target == target:
		c.delta += delta
	default:
		if c.delta != 0 {
			immediate = &Msg[T]{Target: c.target, Delta: c.delta}
		}
		if c.timer != nil {
			c.timer.Stop()
			c.timer = nil
		}
		c.active = true
		c.target = target
		c.delta = delta
		startTimer = true
	}
	c.mu.Unlock()

	if immediate != nil {
		c.dispatch(*immediate)
	}
	if startTimer {
		c.scheduleFlush()
	}
}

func (c *Coalescer[T]) scheduleFlush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.timer != nil {
		c.timer.Stop()
	}
	c.timer = time.AfterFunc(c.delay, c.flush)
}

func (c *Coalescer[T]) flush() {
	c.mu.Lock()
	target := c.target
	delta := c.delta
	c.active = false
	c.target = c.zero
	c.delta = 0
	c.timer = nil
	c.mu.Unlock()

	if delta == 0 {
		return
	}
	c.dispatch(Msg[T]{Target: target, Delta: delta})
}

func (c *Coalescer[T]) dispatch(msg tea.Msg) {
	if c.send == nil {
		return
	}
	go c.send(msg)
}

// Stop cancels any pending flush and resets accumulated state.
func (c *Coalescer[T]) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	c.active = false
	c.target = c.zero
	c.delta = 0
}

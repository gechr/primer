package scrollwheel

import "time"

// Option configures a [Coalescer].
type Option func(*config)

type config struct {
	delay time.Duration
}

// WithDelay sets the debounce window. Default is [DefaultDelay].
func WithDelay(d time.Duration) Option {
	return func(c *config) { c.delay = d }
}

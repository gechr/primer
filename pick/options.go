package pick

// Option configures a selection presented by [MultiSelect].
type Option func(*config)

type config struct {
	filterable bool
}

// WithFilter enables incremental filtering of the list: huh binds "/" to start
// filtering, typing narrows the visible items, and "esc" clears the filter.
// Off by default so short lists stay key-for-key simple.
func WithFilter() Option { return func(c *config) { c.filterable = true } }

func newConfig(opts []Option) config {
	var c config
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

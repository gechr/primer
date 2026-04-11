package input

// Option configures a textarea created by [NewTextArea].
type Option func(*config)

type config struct {
	maxHeight   int
	minHeight   int
	placeholder string
	width       int
}

// WithMaxHeight sets the maximum height of the textarea.
func WithMaxHeight(h int) Option { return func(c *config) { c.maxHeight = h } }

// WithMinHeight sets the minimum height of the textarea.
func WithMinHeight(h int) Option { return func(c *config) { c.minHeight = h } }

// WithPlaceholder sets the placeholder text shown when the textarea is empty.
func WithPlaceholder(s string) Option { return func(c *config) { c.placeholder = s } }

// WithWidth sets the width of the textarea.
func WithWidth(w int) Option { return func(c *config) { c.width = w } }

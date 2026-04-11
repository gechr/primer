package input

import "charm.land/bubbles/v2/textarea"

const (
	defaultMaxHeight   = 10
	defaultMinHeight   = 3
	defaultPlaceholder = "Enter text..."
	defaultWidth       = 80
)

// NewTextArea creates a textarea.Model with sensible defaults for use inside
// TUI overlays: no prompt prefix, no line numbers, and dynamic height.
// The caller should configure styles after creation.
func NewTextArea(opts ...Option) textarea.Model {
	cfg := config{
		maxHeight:   defaultMaxHeight,
		minHeight:   defaultMinHeight,
		placeholder: defaultPlaceholder,
		width:       defaultWidth,
	}
	for _, o := range opts {
		o(&cfg)
	}

	ta := textarea.New()
	ta.Prompt = ""
	ta.Placeholder = cfg.placeholder
	ta.ShowLineNumbers = false
	ta.SetWidth(cfg.width)
	ta.DynamicHeight = true
	ta.MinHeight = cfg.minHeight
	ta.MaxHeight = cfg.maxHeight
	return ta
}

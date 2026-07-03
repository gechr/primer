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

// EditorOption configures an Editor created by [NewEditor].
type EditorOption func(*editorConfig)

type editorConfig struct {
	bodyMinHeight int
	fetchBody     BodyFetchFunc
	styles        EditorStyles
	width         int
}

// WithBodyFetch sets the function used to lazily fetch entry bodies.
func WithBodyFetch(fn BodyFetchFunc) EditorOption {
	return func(c *editorConfig) { c.fetchBody = fn }
}

// WithEditorBodyMinHeight sets the minimum body textarea height.
func WithEditorBodyMinHeight(h int) EditorOption {
	return func(c *editorConfig) { c.bodyMinHeight = h }
}

// WithEditorStyles sets the editor styles.
func WithEditorStyles(s EditorStyles) EditorOption { return func(c *editorConfig) { c.styles = s } }

// WithEditorWidth sets the editor width.
func WithEditorWidth(w int) EditorOption { return func(c *editorConfig) { c.width = w } }

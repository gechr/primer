package hyperlink

import (
	xansi "github.com/charmbracelet/x/ansi"
)

// Fallback controls how hyperlinks render when the output is not a terminal.
type Fallback int

const (
	// FallbackExpanded renders "text (url)".
	FallbackExpanded Fallback = iota
	// FallbackMarkdown renders "[text](url)".
	FallbackMarkdown
	// FallbackText renders only the display text, discarding the URL.
	FallbackText
	// FallbackURL renders only the URL, discarding the display text.
	FallbackURL
)

// Writer produces OSC 8 terminal hyperlinks, falling back to plain text
// when the output is not a terminal.
type Writer struct {
	terminal bool
	fallback Fallback
}

// Option configures a Writer.
type Option func(*Writer)

// WithFallback sets how hyperlinks render when the output is not a terminal.
func WithFallback(fallback Fallback) Option {
	return func(w *Writer) {
		w.fallback = fallback
	}
}

// WithTerminal sets whether the output target is a terminal.
func WithTerminal(v bool) Option {
	return func(w *Writer) {
		w.terminal = v
	}
}

// New creates a Writer with the given options.
func New(opts ...Option) *Writer {
	w := &Writer{}
	for _, o := range opts {
		o(w)
	}
	return w
}

// Terminal reports whether the output target is a terminal.
func (w *Writer) Terminal() bool { return w.terminal }

// Render creates an OSC 8 terminal hyperlink.
// When the output is not a terminal, the Fallback mode controls
// how the link is rendered in plain text.
func (w *Writer) Render(url, text string) string {
	if !w.terminal {
		switch w.fallback {
		case FallbackExpanded:
			return text + " (" + url + ")"
		case FallbackMarkdown:
			return "[" + text + "](" + url + ")"
		case FallbackText:
			return text
		case FallbackURL:
			return url
		}
	}
	return xansi.SetHyperlink(url) + text + xansi.ResetHyperlink()
}

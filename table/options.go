package table

import (
	xansi "github.com/charmbracelet/x/ansi"
)

// Padding controls the text alignment within a column.
type Padding int

const (
	PaddingLeft   Padding = iota // Default: left-aligned (pad right).
	PaddingCenter                // Center-aligned (pad both sides).
	PaddingRight                 // Right-aligned (pad left).
)

// GridOption configures a Grid.
type GridOption func(*Grid)

// WithColumnPadding sets the number of spaces between columns.
func WithColumnPadding(n int) GridOption {
	return func(g *Grid) {
		g.ColumnPadding = n
	}
}

// WithFlexColumns sets the columns that shrink to fit MaxWidth.
func WithFlexColumns(cols ...int) GridOption {
	return func(g *Grid) {
		g.FlexCols = cols
	}
}

// WithPadding sets the text alignment within columns.
func WithPadding(p Padding) GridOption {
	return func(g *Grid) {
		g.Padding = p
	}
}

// WithWidthMethod sets the method used to measure and truncate cell text.
// The default (xansi.WcWidth) counts per-rune wcwidth; use
// xansi.GraphemeWidth on terminals that perform grapheme clustering
// (mode 2027), where ZWJ and VS16 emoji sequences occupy a single glyph.
func WithWidthMethod(m xansi.Method) GridOption {
	return func(g *Grid) {
		g.WidthMethod = m
	}
}

// Option configures a Renderer.
type Option func(*config)

// HeaderRenderer customizes how column headers are rendered before alignment.
// The returned string may include ANSI styling.
type HeaderRenderer func(name, header string, ctx *RenderContext) string

type config struct {
	reverse        bool
	showIndex      bool
	tty            bool // true when outputting to a terminal
	termWidth      int  // terminal width for flex columns (0 = disabled)
	gridOpts       []GridOption
	headerRenderer HeaderRenderer
}

// WithGridOptions sets grid options applied to every grid the renderer
// builds, e.g. WithWidthMethod(xansi.GraphemeWidth) for grapheme-clustering
// terminals. Options that the renderer manages itself (flex columns, max
// width, TTY) are overridden by the renderer's own configuration.
func WithGridOptions(opts ...GridOption) Option {
	return func(c *config) { c.gridOpts = opts }
}

// WithReverse sets whether to reverse row order (newest first at top).
func WithReverse(v bool) Option { return func(c *config) { c.reverse = v } }

// WithShowIndex sets whether to show row indices.
func WithShowIndex(v bool) Option { return func(c *config) { c.showIndex = v } }

// WithTTY sets whether output is going to a terminal.
func WithTTY(v bool) Option { return func(c *config) { c.tty = v } }

// WithTermWidth sets the terminal width for flex columns.
// When set, columns marked Flex=true are truncated so rows fit within this width.
func WithTermWidth(w int) Option { return func(c *config) { c.termWidth = w } }

// WithHeaderRenderer sets a custom header renderer used before column alignment.
func WithHeaderRenderer(fn HeaderRenderer) Option {
	return func(c *config) { c.headerRenderer = fn }
}

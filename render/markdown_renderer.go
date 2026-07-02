package render

import (
	"hash/fnv"
	"strings"
	"sync"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
)

const defaultMarkdownCacheEntries = 256

type markdownTermRenderer interface {
	Render(string) (string, error)
}

var newMarkdownTermRenderer = func(style ansi.StyleConfig, width int) (markdownTermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(width),
	)
}

// MarkdownOption configures a [MarkdownRenderer].
type MarkdownOption func(*MarkdownRenderer)

// MarkdownRenderer renders markdown with identity-keyed, content-invalidating
// caches for repaint-heavy terminal UIs.
type MarkdownRenderer struct {
	mu              sync.Mutex
	style           ansi.StyleConfig
	maxCacheEntries int
	trimPadding     bool
	outputs         map[markdownCacheKey]markdownCacheEntry
	renderers       map[int]markdownTermRenderer
}

type markdownCacheKey struct {
	id    string
	width int
}

type markdownCacheEntry struct {
	hash uint64
	out  string
}

// NewMarkdownRenderer creates a renderer for repeated markdown rendering.
//
// The renderer is safe for concurrent use. It serializes access because
// glamour renderers keep internal buffers; the lock cost is negligible beside
// markdown parsing and ANSI rendering.
func NewMarkdownRenderer(style ansi.StyleConfig, opts ...MarkdownOption) *MarkdownRenderer {
	r := &MarkdownRenderer{
		style:           style,
		maxCacheEntries: defaultMarkdownCacheEntries,
		trimPadding:     true,
		outputs:         make(map[markdownCacheKey]markdownCacheEntry),
		renderers:       make(map[int]markdownTermRenderer),
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.maxCacheEntries < 1 {
		r.maxCacheEntries = 1
	}
	return r
}

// WithMaxCacheEntries caps both rendered-output and per-width glamour renderer
// caches. On overflow the affected cache is reset wholesale because resize
// storms produce many cheap-to-rebuild widths.
func WithMaxCacheEntries(n int) MarkdownOption {
	return func(r *MarkdownRenderer) {
		r.maxCacheEntries = n
	}
}

// WithTrimPadding controls whether Render trims glamour's outer newlines.
// Panes usually own their spacing, so trimming is enabled by default.
func WithTrimPadding(trim bool) MarkdownOption {
	return func(r *MarkdownRenderer) {
		r.trimPadding = trim
	}
}

// Render returns md rendered at width, cached under (id, width), and
// invalidated when md's content hash changes.
//
// Widths smaller than one are clamped so Bubble Tea resize transients still
// render something. Glamour construction or render errors return the raw
// markdown because stale text is better than a blank pane.
func (r *MarkdownRenderer) Render(id string, width int, md string) string {
	if md == "" {
		return ""
	}
	if width < 1 {
		width = 1
	}

	hash := markdownHash(md)
	key := markdownCacheKey{id: id, width: width}

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.outputs[key]
	if ok && entry.hash == hash {
		return entry.out
	}

	out, rendered := r.renderLocked(width, md)
	if rendered && r.trimPadding {
		out = strings.Trim(out, "\n")
	}
	r.storeOutput(key, markdownCacheEntry{hash: hash, out: out})
	return out
}

func (r *MarkdownRenderer) renderLocked(width int, md string) (string, bool) {
	renderer, ok := r.renderers[width]
	if !ok {
		var err error
		renderer, err = newMarkdownTermRenderer(r.style, width)
		if err != nil {
			return md, false
		}
		r.storeRenderer(width, renderer)
	}

	chromaStyleMu.Lock()
	out, err := renderer.Render(md)
	chromaStyleMu.Unlock()
	if err != nil {
		return md, false
	}
	return out, true
}

func (r *MarkdownRenderer) storeOutput(key markdownCacheKey, entry markdownCacheEntry) {
	if len(r.outputs) >= r.maxCacheEntries {
		r.outputs = make(map[markdownCacheKey]markdownCacheEntry)
	}
	r.outputs[key] = entry
}

func (r *MarkdownRenderer) storeRenderer(width int, renderer markdownTermRenderer) {
	if len(r.renderers) >= r.maxCacheEntries {
		r.renderers = make(map[int]markdownTermRenderer)
	}
	r.renderers[width] = renderer
}

func markdownHash(md string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(md))
	return h.Sum64()
}

// Package render provides terminal-friendly markdown and diff rendering.
//
// Markdown is rendered via glamour with a cached renderer keyed on
// (width, style). Diffs are highlighted with chroma using a unified
// diff lexer. Both functions are safe for concurrent use.
package render

import (
	"bytes"
	"strings"
	"sync"

	"charm.land/glamour/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// glamourCache holds a cached renderer to avoid re-creating it on every call.
var glamourCache struct {
	mu       sync.Mutex
	style    string
	width    int
	renderer *glamour.TermRenderer
}

// Markdown renders markdown text for terminal display using glamour.
// style is a glamour style name ("dracula", "dark", "light", "notty").
// Falls back to indented plain text on error or when width is zero.
// Safe for concurrent use.
func Markdown(text string, width int, style string) string {
	if text == "" {
		return ""
	}
	if width <= 0 {
		width = 80
	}

	// Hold the lock across both cache lookup and Render() because
	// glamour's TermRenderer is not safe for concurrent use.
	glamourCache.mu.Lock()
	r := resolveRenderer(width, style)
	if r == nil {
		glamourCache.mu.Unlock()
		return plainFallback(text, width)
	}
	rendered, err := r.Render(text)
	glamourCache.mu.Unlock()

	if err != nil || strings.TrimSpace(rendered) == "" {
		return plainFallback(text, width)
	}
	return strings.TrimRight(rendered, "\n")
}

// Diff highlights a unified diff string using chroma syntax highlighting.
// Returns the input unchanged if highlighting fails.
func Diff(text string) string {
	if text == "" {
		return ""
	}

	lexer := lexers.Get("diff")
	if lexer == nil {
		return text
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	formatter := formatters.TTY256

	iterator, err := lexer.Tokenise(nil, text)
	if err != nil {
		return text
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return text
	}
	return buf.String()
}

// resolveRenderer returns a cached or freshly created renderer.
// Caller must hold glamourCache.mu.
func resolveRenderer(width int, style string) *glamour.TermRenderer {
	if width == glamourCache.width && style == glamourCache.style && glamourCache.renderer != nil {
		return glamourCache.renderer
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}

	glamourCache.style = style
	glamourCache.width = width
	glamourCache.renderer = r
	return r
}

func plainFallback(text string, width int) string {
	indent := "    "
	maxLen := width - len(indent)
	if maxLen <= 0 {
		maxLen = width
		indent = ""
	}

	var sb strings.Builder
	for line := range strings.SplitSeq(text, "\n") {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(indent)
		if len(line) > maxLen {
			sb.WriteString(line[:maxLen])
		} else {
			sb.WriteString(line)
		}
	}
	return sb.String()
}

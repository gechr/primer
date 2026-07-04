// Package render provides terminal-friendly markdown and diff rendering.
//
// Markdown is rendered via glamour with a cached renderer keyed on
// (width, style) for one-shot CLI output. MarkdownRenderer provides a
// stateful cache keyed by caller identity and width for TUIs that repaint
// stable documents while background refreshes may change their content.
// Diffs are highlighted with chroma using a unified diff lexer. All renderers
// in this package are safe for concurrent use.
package render

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"sync"

	"charm.land/glamour/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	xstrings "github.com/gechr/x/strings"
)

// glamourCache holds a cached renderer to avoid re-creating it on every call.
var glamourCache struct {
	mu       sync.Mutex
	style    string
	width    int
	renderer *glamour.TermRenderer
}

var chromaStyleMu sync.Mutex

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
	chromaStyleMu.Lock()
	rendered, err := r.Render(text)
	chromaStyleMu.Unlock()
	glamourCache.mu.Unlock()

	if err != nil || xstrings.IsBlank(rendered) {
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

	chromaStyleMu.Lock()
	style := styles.Get("monokai")
	chromaStyleMu.Unlock()
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

// DiffStyled renders a diff using delta when available, falling back to Diff.
// When RepoURL and CommitSHA are set, delta file links point at that GitHub blob.
func DiffStyled(text string, opts DiffOptions) string {
	if text == "" {
		return ""
	}
	if out, err := DiffWithDelta(text, opts); err == nil {
		return out
	}
	return Diff(text)
}

// DiffWithDelta renders a diff through delta. It returns an error when delta
// is unavailable or execution fails.
func DiffWithDelta(text string, opts DiffOptions) (string, error) {
	if text == "" {
		return "", nil
	}

	deltaBin := opts.DeltaBin
	if deltaBin == "" {
		path, err := exec.LookPath("delta")
		if err != nil {
			return "", err
		}
		deltaBin = path
	}

	args := []string{"--true-color=always"}
	if opts.RepoURL != "" && opts.CommitSHA != "" {
		args = append(args,
			"--hyperlinks",
			"--hyperlinks-file-link-format",
			opts.RepoURL+"/blob/"+opts.CommitSHA+"{path}?plain=1#L{line}",
		)
	}

	cmd := exec.CommandContext(context.Background(), deltaBin, args...)
	if opts.RepoURL != "" && opts.CommitSHA != "" {
		// Delta resolves {path} against CWD; "/" yields "/{relative_path}".
		cmd.Dir = "/"
	}
	cmd.Stdin = strings.NewReader(text)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
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

package render_test

import (
	"testing"

	"github.com/gechr/primer/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownRendersNonEmpty(t *testing.T) {
	t.Parallel()
	out := render.Markdown("# Hello\n\nWorld", 80, "dracula")
	require.NotEmpty(t, out)
	assert.Contains(t, out, "World")
}

func TestMarkdownEmptyInput(t *testing.T) {
	t.Parallel()
	assert.Empty(t, render.Markdown("", 80, "dracula"))
}

func TestMarkdownFallbackOnZeroWidth(t *testing.T) {
	t.Parallel()
	out := render.Markdown("hello", 0, "dracula")
	assert.Contains(t, out, "hello")
}

func TestMarkdownDifferentStyles(t *testing.T) {
	t.Parallel()
	for _, style := range []string{"dracula", "dark", "light", "notty"} {
		out := render.Markdown("**bold**", 80, style)
		assert.NotEmpty(t, out, "style=%s", style)
	}
}

func TestDiffHighlights(t *testing.T) {
	t.Parallel()
	diff := `--- a/file.go
+++ b/file.go
@@ -1,3 +1,3 @@
 package main
-var old = 1
+var new = 2
`
	out := render.Diff(diff)
	require.NotEmpty(t, out)
	// Chroma adds ANSI escape codes, so the output should differ from raw.
	assert.NotEqual(t, diff, out)
}

func TestDiffEmptyInput(t *testing.T) {
	t.Parallel()
	assert.Empty(t, render.Diff(""))
}

func TestDiffPlainTextPassthrough(t *testing.T) {
	t.Parallel()
	// Non-diff text should still be processed by the diff lexer
	// but the result should be non-empty.
	out := render.Diff("just some text")
	assert.NotEmpty(t, out)
}

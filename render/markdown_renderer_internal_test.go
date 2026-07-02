package render

import (
	"errors"
	"testing"

	"charm.land/glamour/v2/ansi"
	"github.com/stretchr/testify/require"
)

type failingMarkdownRenderer struct{}

func (failingMarkdownRenderer) Render(string) (string, error) {
	return "", errors.New("render failed")
}

func TestMarkdownRendererFallsBackToRawMarkdownOnError(t *testing.T) {
	original := newMarkdownTermRenderer
	newMarkdownTermRenderer = func(ansi.StyleConfig, int) (markdownTermRenderer, error) {
		return failingMarkdownRenderer{}, nil
	}
	t.Cleanup(func() {
		newMarkdownTermRenderer = original
	})

	r := NewMarkdownRenderer(ansi.StyleConfig{})
	md := "\n# raw\n"

	got := r.Render("doc", 80, md)

	require.Equal(t, md, got)
}

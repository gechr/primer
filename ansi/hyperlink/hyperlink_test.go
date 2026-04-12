package hyperlink_test

import (
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/ansi/hyperlink"
	"github.com/stretchr/testify/require"
)

func TestNewNoOptions(t *testing.T) {
	w := hyperlink.New()

	require.False(t, w.Terminal())
}

func TestNewWithTerminalTrue(t *testing.T) {
	w := hyperlink.New(hyperlink.WithTerminal(true))

	require.True(t, w.Terminal())
}

func TestNewWithTerminalFalse(t *testing.T) {
	w := hyperlink.New(hyperlink.WithTerminal(false))

	require.False(t, w.Terminal())
}

func TestRenderFallbackExpanded(t *testing.T) {
	w := hyperlink.New()

	got := w.Render("https://example.com", "link")

	require.Equal(t, "link (https://example.com)", got)
}

func TestRenderFallbackMarkdown(t *testing.T) {
	w := hyperlink.New(hyperlink.WithFallback(hyperlink.FallbackMarkdown))

	got := w.Render("https://example.com", "link")

	require.Equal(t, "[link](https://example.com)", got)
}

func TestRenderFallbackText(t *testing.T) {
	w := hyperlink.New(hyperlink.WithFallback(hyperlink.FallbackText))

	got := w.Render("https://example.com", "link")

	require.Equal(t, "link", got)
}

func TestRenderFallbackURL(t *testing.T) {
	w := hyperlink.New(hyperlink.WithFallback(hyperlink.FallbackURL))

	got := w.Render("https://example.com", "link")

	require.Equal(t, "https://example.com", got)
}

func TestRenderUnknownFallbackDefaultsToTerminalSequenceWhenEnabled(t *testing.T) {
	w := hyperlink.New(
		hyperlink.WithFallback(hyperlink.Fallback(99)),
		hyperlink.WithTerminal(true),
	)

	got := w.Render("https://example.com", "link")

	expected := xansi.SetHyperlink("https://example.com") + "link" + xansi.ResetHyperlink()
	require.Equal(t, expected, got)
	require.Equal(t, "link", xansi.Strip(got))
}

func TestRenderForceTerminalUsesOSC8(t *testing.T) {
	w := hyperlink.New(hyperlink.WithTerminal(true))

	got := w.Render("https://example.com", "link")

	expected := xansi.SetHyperlink("https://example.com") + "link" + xansi.ResetHyperlink()
	require.Equal(t, expected, got)
	require.Equal(t, "link", xansi.Strip(got))
}

package overlay_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/overlay"
	"github.com/stretchr/testify/require"
)

func TestPlaceCentersForeground(t *testing.T) {
	background := strings.Join([]string{
		"........",
		"........",
		"........",
		"........",
	}, "\n")

	got := overlay.Place(background, "XX", 8, 4, overlay.Center)

	lines := strings.Split(ansi.Strip(got), "\n")
	require.Len(t, lines, 4)
	require.Equal(t, "...XX...", lines[1])
}

func TestPlacePadsShortBackground(t *testing.T) {
	got := overlay.Place("", "X", 4, 3, overlay.Center)

	lines := strings.Split(ansi.Strip(got), "\n")
	require.Len(t, lines, 3)
	require.Equal(t, " X", strings.TrimRight(lines[1], " "))
}

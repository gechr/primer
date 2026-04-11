package view_test

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/layout"
	"github.com/gechr/primer/scrollbar"
	"github.com/gechr/primer/view"
	"github.com/stretchr/testify/require"
)

func TestRenderContentUsesCachedLinesWithScrollbar(t *testing.T) {
	lines := []string{
		layout.NormalizeLine("line 1", 10),
		layout.NormalizeLine("line 2", 10),
		layout.NormalizeLine("line 3", 10),
		layout.NormalizeLine("line 4", 10),
	}
	vp := viewport.New()
	vp.SetWidth(10)
	vp.SetHeight(3)
	vp.SetContentLines(lines)
	vp.SetYOffset(1)

	got := ansi.Strip(view.RenderContent(lines, vp, true, scrollbar.Styles{
		Thumb: lg.NewStyle(),
		Track: lg.NewStyle(),
	}))
	rows := strings.Split(got, "\n")

	require.Len(t, rows, 3)
	require.Contains(t, rows[0], "line 2")
	require.Contains(t, rows[1], "line 3")
	require.Contains(t, rows[2], "line 4")
	for _, row := range rows {
		require.Equal(t, 11, ansi.WcWidth.StringWidth(row))
	}
}

func TestRenderFrameFillsTerminalRectangle(t *testing.T) {
	lines := []string{layout.NormalizeLine("body", 10)}
	vp := viewport.New()
	vp.SetWidth(10)
	vp.SetHeight(3)
	vp.SetContentLines(lines)

	got := view.RenderFrame(view.FrameModel{
		Header: "Header",
		Footer: "Footer",
		Lines:  lines,
		VP:     vp,
		Width:  12,
		Height: 7,
		Styles: view.FrameStyles{
			Separator: lg.NewStyle(),
			Scrollbar: scrollbar.Styles{
				Thumb: lg.NewStyle(),
				Track: lg.NewStyle(),
			},
		},
	})

	rows := strings.Split(ansi.Strip(got), "\n")
	require.Len(t, rows, 7)
}

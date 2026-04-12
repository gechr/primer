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
	require.Equal(t, "line 2    ┃", rows[0])
	require.Equal(t, "line 3    ┃", rows[1])
	require.Equal(t, "line 4    █", rows[2])
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
		View:   vp,
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

func TestSyncNormalizesViewportState(t *testing.T) {
	vp := viewport.New()

	lines := view.Sync(&vp, []string{"hi"}, 4, 2)

	require.Equal(t, []string{"hi  "}, lines)
	require.Equal(t, 4, vp.Width())
	require.Equal(t, 2, vp.Height())
	require.True(t, vp.FillHeight)
}

func TestRenderContentWithoutScrollbarPadsBlankRows(t *testing.T) {
	vp := viewport.New()
	vp.SetWidth(4)
	vp.SetHeight(2)
	vp.SetContentLines([]string{"line"})

	got := view.RenderContent([]string{"line"}, vp, false, scrollbar.Styles{})

	require.Equal(t, "line\n    ", got)
}

func TestRenderContentReturnsEmptyForZeroHeight(t *testing.T) {
	vp := viewport.New()
	vp.SetWidth(4)
	vp.SetHeight(0)

	require.Empty(t, view.RenderContent([]string{"line"}, vp, false, scrollbar.Styles{}))
}

func TestRenderFrameUsesScrollbarWhenContentExceedsViewport(t *testing.T) {
	lines := []string{
		layout.NormalizeLine("one", 4),
		layout.NormalizeLine("two", 4),
		layout.NormalizeLine("three", 4),
	}
	vp := viewport.New()
	vp.SetWidth(4)
	vp.SetHeight(2)
	vp.SetContentLines(lines)
	vp.SetYOffset(1)

	got := ansi.Strip(view.RenderFrame(view.FrameModel{
		Lines:  lines,
		View:   vp,
		Width:  4,
		Height: 4,
		Styles: view.FrameStyles{
			Separator: lg.NewStyle(),
			Scrollbar: scrollbar.Styles{
				Thumb: lg.NewStyle(),
				Track: lg.NewStyle(),
			},
		},
	}))

	require.Equal(t, "two ┃\nthre█\n────\n", got)
}

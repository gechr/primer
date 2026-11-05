package key_test

import (
	"strings"
	"testing"

	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/key"
	"github.com/gechr/primer/layout"
	"github.com/stretchr/testify/require"
)

func TestInlineSingleLetter(t *testing.T) {
	rendered, ok := key.Inline("a", "approve", lg.NewStyle(), lg.NewStyle())

	require.True(t, ok)
	require.Equal(t, "approve", ansi.Strip(rendered))
}

func TestInlineModifiedKey(t *testing.T) {
	rendered, ok := key.Inline("alt+c", "copy", lg.NewStyle(), lg.NewStyle())

	require.True(t, ok)
	require.Equal(t, "alt+copy", ansi.Strip(rendered))
}

func TestRendererPrefixDefault(t *testing.T) {
	r := key.Renderer{
		Styles: key.Styles{Key: lg.NewStyle(), Text: lg.NewStyle()},
	}

	got := r.Render([]key.Hint{{Key: "a", Desc: "approve"}})
	require.True(t, strings.HasPrefix(ansi.Strip(got), " "), "default prefix should be a space")
}

func TestRendererPrefixCustom(t *testing.T) {
	r := key.Renderer{
		Styles: key.Styles{Key: lg.NewStyle(), Text: lg.NewStyle()},
		Prefix: new(">> "),
	}

	got := r.Render([]key.Hint{{Key: "a", Desc: "approve"}})
	require.True(t, strings.HasPrefix(ansi.Strip(got), ">> "), "custom prefix should be used")
}

func TestRendererPrefixEmpty(t *testing.T) {
	r := key.Renderer{
		Styles: key.Styles{Key: lg.NewStyle(), Text: lg.NewStyle()},
		Prefix: new(""),
	}

	got := r.Render([]key.Hint{{Key: "a", Desc: "approve"}})
	require.Equal(t, "a approve", ansi.Strip(got), "empty prefix should produce no leading space")
}

func TestRendererWrapsAtWidth(t *testing.T) {
	r := key.Renderer{
		Styles: key.Styles{
			Key:  lg.NewStyle(),
			Text: lg.NewStyle(),
		},
		Width:  12,
		Inline: true,
	}

	got := r.Render([]key.Hint{
		{Key: "a", Desc: "approve"},
		{Key: "c", Desc: "comment"},
	})

	lines := strings.Split(ansi.Strip(got), "\n")
	require.Len(t, lines, 2)
}

// Every cell a hint claims must spell out exactly that hint's rendered text,
// and the cells between hints must belong to none.
func TestRendererHintAtMatchesRender(t *testing.T) {
	hints := []key.Hint{
		{Key: "enter", Desc: "show"},
		{Key: "space", Desc: "select"},
		{Key: "/", Desc: "filter"},
		{Key: "c", Desc: "comment"},
		{Key: "alt+c", Desc: "copy"},
		{Key: "r", Desc: "refresh " + lg.NewStyle().Bold(true).Render("off")},
		{Key: "?", Desc: "help"},
	}
	want := []string{
		"enter show", "space select", "/ filter", "comment", "alt+copy", "refresh off", "? help",
	}
	empty := ""
	for _, prefix := range []*string{nil, &empty} {
		for _, width := range []int{0, 12, 20, 33, 80} {
			r := key.Renderer{
				Styles: key.Styles{Key: lg.NewStyle().Bold(true), Text: lg.NewStyle()},
				Prefix: prefix,
				Width:  width,
				Inline: true,
			}
			lines := strings.Split(ansi.Strip(r.Render(hints)), "\n")
			got := make([]string, len(hints))
			for y, line := range lines {
				for x := range ansi.StringWidth(line) + 2 {
					i, ok := r.HintAt(hints, x, y)
					if !ok {
						continue
					}
					got[i] += ansi.Cut(line, x, x+1)
				}
			}
			require.Equal(t, want, got, "width=%d prefix=%v", width, prefix != nil)
		}
	}
}

func TestRendererHintAtMissesOutsideHints(t *testing.T) {
	r := key.Renderer{Styles: key.Styles{Key: lg.NewStyle(), Text: lg.NewStyle()}, Width: 80}
	hints := []key.Hint{{Key: "a", Desc: "approve"}, {Key: "b", Desc: "back"}}
	for _, at := range [][2]int{{0, 0}, {10, 0}, {11, 0}, {18, 0}, {5, 1}, {-1, 0}} {
		_, ok := r.HintAt(hints, at[0], at[1])
		require.False(t, ok, "x=%d y=%d", at[0], at[1])
	}
	i, ok := r.HintAt(hints, 12, 0)
	require.True(t, ok)
	require.Equal(t, 1, i)
}

func TestRendererHoverPaintsOnlyThatHint(t *testing.T) {
	const bg = "\x1b[48;2;1;2;3m"
	hints := []key.Hint{{Key: "a", Desc: "approve"}, {Key: "b", Desc: "back"}}
	r := key.Renderer{Styles: key.Styles{Key: lg.NewStyle(), Text: lg.NewStyle()}, HoverBG: bg}
	plain := r.Render(hints)

	hover := 1
	r.Hover = &hover
	require.Equal(t, " a approve  "+layout.PreserveBackground("b back", bg), r.Render(hints))
	i, ok := r.HintAt(hints, 12, 0)
	require.True(t, ok)
	require.Equal(t, 1, i, "hover leaves hit areas unchanged")

	for _, out := range []int{-1, 2} {
		hover = out
		require.Equal(t, plain, r.Render(hints), "out-of-range hover paints nothing")
	}
	hover = 0
	r.HoverBG = ""
	require.Equal(t, plain, r.Render(hints), "no background, no paint")
}

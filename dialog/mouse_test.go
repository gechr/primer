package dialog_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/gechr/primer/dialog"
	"github.com/stretchr/testify/require"
)

func leftClick(x, y int) tea.MouseClickMsg {
	return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}
}

// The borderedFrame boxes "PAYLOAD" (7 columns, 1 row) into a 9x3 box, which
// centers on a 40x12 screen at column 15, row 4; the content cell inside the
// border is (16, 5).
func TestClickTranslatesIntoContentSpace(t *testing.T) {
	t.Parallel()

	var seen []tea.Msg
	s := dialog.New(borderedFrame())
	s.Push(stubDialog{body: "PAYLOAD", seen: &seen})

	// Before any render there is no geometry: the raw click passes through.
	s.Update(leftClick(16, 5))
	require.Equal(t, []tea.Msg{leftClick(16, 5)}, seen)

	seen = nil
	s.View(backdrop(12), 40, 12)
	s.Update(leftClick(16, 5))
	s.Update(leftClick(22, 5))
	require.Equal(t, []tea.Msg{dialog.ClickMsg{X: 0, Y: 0}, dialog.ClickMsg{X: 6, Y: 0}}, seen)

	// The border, the space outside the box, and non-left buttons all arrive
	// untranslated.
	seen = nil
	s.Update(leftClick(15, 5))
	s.Update(leftClick(0, 0))
	right := tea.MouseClickMsg{X: 16, Y: 5, Button: tea.MouseRight}
	s.Update(right)
	require.Equal(t, []tea.Msg{leftClick(15, 5), leftClick(0, 0), right}, seen)
}

func TestClickAccountsForFrameTitle(t *testing.T) {
	t.Parallel()

	var seen []tea.Msg
	s := dialog.New(borderedFrame())
	s.Push(stubDialog{title: "T", body: "PAYLOAD", seen: &seen})
	s.View(backdrop(12), 40, 12)

	// The inner area is title (row 0) over body (row 1): a body click maps to
	// body row 0, a title click to the negative row above it.
	// Box is 9x4 centered at (15, 4); content starts at (16, 5).
	s.Update(leftClick(16, 6))
	s.Update(leftClick(16, 5))
	require.Equal(t, []tea.Msg{dialog.ClickMsg{X: 0, Y: 0}, dialog.ClickMsg{X: 0, Y: -1}}, seen)
}

func TestClickTranslatesForSelfFramedDialog(t *testing.T) {
	t.Parallel()

	var seen []tea.Msg
	s := dialog.New(borderedFrame())
	s.Push(selfFramedStub{body: "BOX", seen: &seen, self: true})
	s.View(backdrop(12), 40, 12)

	// "BOX" (3x1) placed verbatim centers at (18, 5); no frame to skip.
	s.Update(leftClick(19, 5))
	require.Equal(t, []tea.Msg{dialog.ClickMsg{X: 1, Y: 0}}, seen)
}

func TestScrollbarHitbox(t *testing.T) {
	t.Parallel()

	s := dialog.New(dialog.NewFrame(dialog.FrameConfig{
		Styles:    dialog.Styles{Box: lg.NewStyle().Border(lg.NormalBorder())},
		MaxHeight: 6,
	}))

	// No dialog, no hitbox.
	_, ok := s.ScrollbarHitbox()
	require.False(t, ok)

	// A 10-row body in a 6-row box (4 viewport rows) scrolls; the bar sits one
	// column right of the 1-column body. The 5x6 box centers at (17, 3), so
	// the content starts at (18, 4) and the bar occupies column 19.
	s.Push(stubDialog{body: strings.TrimSuffix(strings.Repeat("L\n", 10), "\n")})
	_, ok = s.ScrollbarHitbox()
	require.False(t, ok) // pushed but not yet rendered
	require.False(t, s.SetScrollOffset(3))
	require.False(t, s.ScrollBy(1))

	s.View(backdrop(12), 40, 12)
	hb, ok := s.ScrollbarHitbox()
	require.True(t, ok)
	require.Equal(t, 19, hb.X)
	require.Equal(t, 4, hb.Y)
	require.Equal(t, 4, hb.Height)
	require.Equal(t, 10, hb.TotalLines)
	require.Zero(t, s.ScrollPercent())
	require.True(t, s.SetScrollOffset(3))
	s.View(backdrop(12), 40, 12)
	require.Positive(t, s.ScrollPercent())
	require.True(t, s.ScrollBy(20))
	s.View(backdrop(12), 40, 12)
	require.GreaterOrEqual(t, s.ScrollPercent(), 0.9)

	// A short body needs no scrollbar.
	s.Push(stubDialog{body: "SHORT"})
	s.View(backdrop(12), 40, 12)
	_, ok = s.ScrollbarHitbox()
	require.False(t, ok)
}

func TestDismissResultClosesOnTranslatedClick(t *testing.T) {
	t.Parallel()

	require.Equal(t, dialog.ResultClose, dialog.DismissResult(dialog.ClickMsg{}))
}

package dialog_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/gechr/primer/dialog"
	"github.com/gechr/primer/key"
	"github.com/stretchr/testify/require"
)

// The grace windows stack.go declares; the tests pin them by value so a
// change to the tuning is a conscious, test-visible decision.
const (
	graceQuiet        = 200 * time.Millisecond
	graceReopenExempt = 500 * time.Millisecond
)

// stubDialog is a minimal Dialog for driving the Stack and Frame without any
// real dialog. seen is a pointer so the messages it records survive the value
// copies the Stack makes when it stores each Update's returned dialog.
type stubDialog struct {
	id     string
	title  string
	body   string
	hints  []key.Hint
	result dialog.Result
	seen   *[]tea.Msg
}

func (d stubDialog) Title() string      { return d.title }
func (d stubDialog) Content(int) string { return d.body }
func (d stubDialog) Hints() []key.Hint  { return d.hints }

func (d stubDialog) Update(msg tea.Msg) (dialog.Dialog, tea.Cmd, dialog.Result) {
	if d.seen != nil {
		*d.seen = append(*d.seen, msg)
	}
	return d, nil, d.result
}

// keyMsg is a stand-in for any non-resize message.
type keyMsg struct{}

// selfFramedStub is a stubDialog that reports itself self-framed.
type selfFramedStub struct {
	stubDialog

	self bool
}

func (d selfFramedStub) SelfFramed() bool { return d.self }

func borderedFrame() dialog.Frame {
	return dialog.NewFrame(dialog.FrameConfig{
		Styles: dialog.Styles{Box: lg.NewStyle().Border(lg.NormalBorder())},
	})
}

func backdrop(h int) string {
	return strings.TrimSuffix(strings.Repeat(strings.Repeat(" ", 40)+"\n", h), "\n")
}

// shows counts how many times needle appears in the rendered screen, so
// presence asserts through exact counting rather than a bare Contains.
func shows(screen, needle string) int { return strings.Count(screen, needle) }

func TestStackSelfFramed(t *testing.T) {
	t.Parallel()

	t.Run("plain dialog is framed by the frame", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(borderedFrame())
		s.Push(stubDialog{body: "PAYLOAD"})
		got := s.View(backdrop(12), 40, 12)
		require.True(
			t,
			strings.ContainsAny(got, "│─"),
			"plain dialog should gain the frame's border:\n%s",
			got,
		)
	})

	t.Run("self-framed dialog is placed verbatim", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(borderedFrame())
		s.Push(selfFramedStub{stubDialog: stubDialog{body: "PAYLOAD"}, self: true})
		got := s.View(backdrop(12), 40, 12)
		require.Equal(t, 1, shows(got, "PAYLOAD"), "self-framed content missing:\n%s", got)
		require.False(
			t,
			strings.ContainsAny(got, "│─"),
			"self-framed dialog must not gain the frame's border:\n%s",
			got,
		)
	})
}

// wrappingDialog is a Dialog whose Content honors the width it is given, hard-
// wrapping its body to that column count exactly as a real body-rendering
// dialog would. It is what makes an off-by-one in the width the Frame allots
// observable: a body given one column too few wraps a character early.
type wrappingDialog struct {
	stubDialog
}

func (d wrappingDialog) Content(width int) string {
	return lg.NewStyle().Width(width).Render(d.body)
}

func TestFrameFrameWidth(t *testing.T) {
	t.Parallel()

	t.Run("a body that fits the inner width is not wrapped a column early", func(t *testing.T) {
		t.Parallel()

		// MaxWidth 9 minus the border's two frame columns leaves an inner width
		// of exactly 7 - the width of the payload - so the body fits only when
		// the Frame does not steal a column for a scrollbar it will not draw.
		s := dialog.New(dialog.NewFrame(dialog.FrameConfig{
			MaxWidth: 9,
			Styles:   dialog.Styles{Box: lg.NewStyle().Border(lg.NormalBorder())},
		}))
		s.Push(wrappingDialog{stubDialog{body: "PAYLOAD"}})

		got := s.View(backdrop(12), 40, 12)
		require.Equal(
			t,
			1,
			shows(got, "PAYLOAD"),
			"7-column body wrapped inside a 7-column inner width:\n%s",
			got,
		)
	})

	t.Run("a scrolling body keeps full text width beside the scrollbar", func(t *testing.T) {
		t.Parallel()

		// MaxWidth 12 minus the border leaves a 10-column inner region; the
		// scrollbar claims one, leaving exactly 7 for the payload. MaxHeight 5
		// forces the five-line body to scroll so the scrollbar actually draws.
		s := dialog.New(dialog.NewFrame(dialog.FrameConfig{
			MaxWidth:  12,
			MaxHeight: 5,
			Styles:    dialog.Styles{Box: lg.NewStyle().Border(lg.NormalBorder())},
		}))
		s.Push(
			wrappingDialog{
				stubDialog{body: strings.TrimSuffix(strings.Repeat("PAYLOAD\n", 5), "\n")},
			},
		)

		got := s.View(backdrop(12), 40, 12)
		require.Positive(
			t,
			shows(got, "PAYLOAD"),
			"scrolling body lost a text column to the scrollbar:\n%s",
			got,
		)
		require.True(
			t,
			strings.ContainsAny(got, "│─"),
			"scrolling dialog dropped its border:\n%s",
			got,
		)
	})
}

func TestFrameTooNarrowShowsNoticeNotGarbage(t *testing.T) {
	t.Parallel()

	frame := dialog.NewFrame(dialog.FrameConfig{
		MaxWidth: 20,
		Styles:   dialog.Styles{Box: lg.NewStyle().Border(lg.NormalBorder())},
	})
	// A width-blind body wider than the 18-column inner width: boxing it would
	// interleave wrapped borders, so the Frame must swap in the notice.
	wide := strings.Repeat("X", 60)

	t.Run("plain frame", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(frame)
		s.Push(stubDialog{body: wide})
		got := s.View(backdrop(12), 40, 12)
		require.Equal(
			t,
			1,
			shows(got, "too narrow"),
			"overflowing body did not show the too-narrow notice:\n%s",
			got,
		)
		require.Zero(t, shows(got, "XXX"), "overflowing body leaked into the frame:\n%s", got)
	})

	t.Run("footered frame", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(frame)
		s.Push(footeredStub{
			scrollHintStub: scrollHintStub{stubDialog: stubDialog{body: wide}},
			footer:         "hints",
		})
		got := s.View(backdrop(12), 40, 12)
		require.Equal(
			t,
			1,
			shows(got, "too narrow"),
			"overflowing footered body did not show the too-narrow notice:\n%s",
			got,
		)
		require.Zero(
			t,
			shows(got, "XXX"),
			"overflowing footered body leaked into the frame:\n%s",
			got,
		)
	})

	t.Run("fitting body still renders", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(frame)
		s.Push(stubDialog{body: "PAYLOAD"})
		got := s.View(backdrop(12), 40, 12)
		require.Equal(t, 1, shows(got, "PAYLOAD"), "fitting body was refused:\n%s", got)
		require.Zero(t, shows(got, "too narrow"), "fitting body was refused:\n%s", got)
	})
}

func TestStack(t *testing.T) {
	t.Parallel()

	t.Run("empty stack is inert", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(dialog.NewFrame(dialog.FrameConfig{}))
		require.False(t, s.Active())
		require.Nil(t, s.Top())
		cmd, popped, res := s.Update(keyMsg{})
		require.Nil(t, cmd)
		require.Nil(t, popped)
		require.Equal(t, dialog.ResultNone, res)
	})

	t.Run("push and top", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(dialog.NewFrame(dialog.FrameConfig{}))
		s.Push(stubDialog{id: "a"})
		s.Push(stubDialog{id: "b"})
		require.True(t, s.Active())
		top, ok := s.Top().(stubDialog)
		require.True(t, ok)
		require.Equal(t, "b", top.id, "top should be the last pushed")
	})

	t.Run("routes only to the top dialog", func(t *testing.T) {
		t.Parallel()

		var bottomSeen, topSeen []tea.Msg
		s := dialog.New(dialog.NewFrame(dialog.FrameConfig{}))
		s.Push(stubDialog{id: "bottom", seen: &bottomSeen})
		s.Push(stubDialog{id: "top", seen: &topSeen})

		s.Update(keyMsg{})

		require.Len(t, topSeen, 1)
		require.Empty(t, bottomSeen)
	})

	popCases := []struct {
		name   string
		result dialog.Result
	}{
		{name: "submit pops and returns the dialog", result: dialog.ResultSubmit},
		{name: "close pops and returns the dialog", result: dialog.ResultClose},
	}
	for _, tc := range popCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := dialog.New(dialog.NewFrame(dialog.FrameConfig{}))
			s.Push(stubDialog{id: "under"})
			s.Push(stubDialog{id: "done", result: tc.result})

			_, popped, res := s.Update(keyMsg{})

			require.Equal(t, tc.result, res)
			require.NotNil(t, popped, "finished dialog was not returned")
			done, ok := popped.(stubDialog)
			require.True(t, ok)
			require.Equal(t, "done", done.id)
			top, ok := s.Top().(stubDialog)
			require.True(t, ok)
			require.Equal(t, "under", top.id)
		})
	}

	t.Run("ResultNone keeps the dialog and returns nil", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(dialog.NewFrame(dialog.FrameConfig{}))
		s.Push(stubDialog{id: "stay", result: dialog.ResultNone})

		_, popped, res := s.Update(keyMsg{})

		require.Nil(t, popped)
		require.Equal(t, dialog.ResultNone, res)
		require.True(t, s.Active(), "ResultNone popped the dialog")
	})
}

// fakeClock drives a Stack's grace windows without sleeping.
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newGraceStack() (*dialog.Stack, *fakeClock) {
	s := dialog.New(dialog.NewFrame(dialog.FrameConfig{}))
	c := &fakeClock{t: time.Unix(1000, 0)}
	s.SetClock(c.now)
	return s, c
}

// closeGraced pops a grace-pushed dialog cleanly: it advances past the quiet
// window so the closing key is delivered rather than absorbed.
func closeGraced(s *dialog.Stack, c *fakeClock, d dialog.Dialog) {
	s.PushWithGrace(d)
	c.advance(300 * time.Millisecond)
	s.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
}

func TestStackInputGrace(t *testing.T) {
	t.Parallel()

	t.Run("quiet window gates delivery", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name  string
			delay time.Duration
			want  bool // key reaches the dialog
		}{
			{name: "within the window is swallowed", delay: 100 * time.Millisecond, want: false},
			{name: "exactly at the window delivers", delay: graceQuiet, want: true},
			{name: "past the window delivers", delay: 250 * time.Millisecond, want: true},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				var seen []tea.Msg
				s, c := newGraceStack()
				s.PushWithGrace(stubDialog{seen: &seen})

				c.advance(tc.delay)
				s.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
				require.Equal(
					t,
					tc.want,
					len(seen) == 1,
					"after %v the key reached the dialog",
					tc.delay,
				)
				require.True(t, s.Active(), "the key popped the dialog")
			})
		}
	})

	t.Run("continuous input expires at the ceiling", func(t *testing.T) {
		t.Parallel()

		var seen []tea.Msg
		s, c := newGraceStack()
		s.PushWithGrace(stubDialog{seen: &seen})

		// A key every 100ms keeps the keyboard louder than the quiet window,
		// so only the ceiling can end the grace. The first delivery must land
		// on the keypress at exactly 1.5s (iteration 15): earlier would mean
		// a shrunken ceiling or a quiet window not refreshed by swallowed
		// keys, later a wedged grace.
		firstDelivered := 0
		for i := 1; i <= 20; i++ {
			c.advance(100 * time.Millisecond)
			before := len(seen)
			s.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
			if len(seen) > before && firstDelivered == 0 {
				firstDelivered = i
			}
		}
		require.Equal(t, 15, firstDelivered, "first delivery should land on the 1.5s ceiling")
		require.Len(t, seen, 6, "iterations 15-20 should deliver")
	})

	t.Run("same-kind reopen within the exemption takes keys immediately", func(t *testing.T) {
		t.Parallel()

		var seen []tea.Msg
		s, c := newGraceStack()
		closeGraced(s, c, stubDialog{result: dialog.ResultClose})

		c.advance(100 * time.Millisecond)
		s.PushWithGrace(stubDialog{seen: &seen})
		c.advance(10 * time.Millisecond)
		s.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		require.Len(t, seen, 1, "rapid same-kind reopen still ate the key")
	})

	t.Run("reopen of a different kind keeps the grace", func(t *testing.T) {
		t.Parallel()

		var seen []tea.Msg
		s, c := newGraceStack()
		closeGraced(s, c, stubDialog{result: dialog.ResultClose})

		c.advance(100 * time.Millisecond)
		s.PushWithGrace(selfFramedStub{stubDialog: stubDialog{seen: &seen}})
		c.advance(10 * time.Millisecond)
		s.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		require.Empty(t, seen, "a different dialog kind rode the reopen exemption")
	})

	t.Run("reopen at exactly the exemption window keeps the grace", func(t *testing.T) {
		t.Parallel()

		var seen []tea.Msg
		s, c := newGraceStack()
		closeGraced(s, c, stubDialog{result: dialog.ResultClose})

		c.advance(graceReopenExempt)
		s.PushWithGrace(stubDialog{seen: &seen})
		c.advance(10 * time.Millisecond)
		s.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		require.Empty(t, seen, "a reopen at the exemption boundary skipped the grace")
	})

	t.Run("a plain-pushed close never feeds the exemption", func(t *testing.T) {
		t.Parallel()

		var seen []tea.Msg
		s, c := newGraceStack()
		s.Push(stubDialog{result: dialog.ResultClose})
		s.Update(tea.KeyPressMsg{Code: tea.KeyEscape}) // user-driven close

		c.advance(100 * time.Millisecond)
		s.PushWithGrace(stubDialog{seen: &seen})
		c.advance(10 * time.Millisecond)
		s.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		require.Empty(t, seen, "a user-closed dialog disarmed an async pick's grace")
	})

	t.Run("plain push has no grace", func(t *testing.T) {
		t.Parallel()

		var seen []tea.Msg
		s, c := newGraceStack()
		s.Push(stubDialog{seen: &seen})
		c.advance(time.Millisecond)
		s.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		require.Len(t, seen, 1, "plain push swallowed a key")
	})

	t.Run("non-key messages route during the grace", func(t *testing.T) {
		t.Parallel()

		var seen []tea.Msg
		s, c := newGraceStack()
		s.PushWithGrace(stubDialog{seen: &seen})
		c.advance(10 * time.Millisecond)
		s.Update(keyMsg{}) // not a tea.KeyPressMsg - must pass
		require.Len(t, seen, 1, "non-key message was absorbed by the grace")
	})

	t.Run("a later push ends the grace", func(t *testing.T) {
		t.Parallel()

		var underSeen, topSeen []tea.Msg
		s, c := newGraceStack()
		s.PushWithGrace(stubDialog{seen: &underSeen})
		s.Push(stubDialog{seen: &topSeen})
		c.advance(time.Millisecond)
		s.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		require.Len(t, topSeen, 1, "stale grace swallowed a key meant for a user-opened dialog")
	})
}

// scrollHintStub is a stubDialog that reports a scroll region, so the Frame's
// focus-follows scrolling is observable: its ScrollTo names the body line to
// keep visible.
type scrollHintStub struct {
	stubDialog

	top, height int
}

func (d scrollHintStub) ScrollTo() (int, int, bool) { return d.top, d.height, true }

func TestFrameFollowsFocus(t *testing.T) {
	t.Parallel()

	// A 30-line body in a viewport far shorter than it: without a scroll hint the
	// Frame shows the top lines only. The hint points near the bottom, so the
	// Frame must scroll that line into view.
	lines := make([]string, 30)
	for i := range lines {
		lines[i] = fmt.Sprintf("L%02d", i)
	}
	body := strings.Join(lines, "\n")
	frame := dialog.NewFrame(dialog.FrameConfig{
		MaxHeight: 12,
		Styles:    dialog.Styles{Box: lg.NewStyle().Border(lg.NormalBorder())},
	})

	t.Run("no hint anchors at the top", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(frame)
		s.Push(stubDialog{body: body})
		got := s.View(backdrop(40), 40, 40)
		require.Equal(t, 1, shows(got, "L00"), "top of an unscrolled body should show:\n%s", got)
		require.Zero(
			t,
			shows(got, "L27"),
			"an unscrolled body should not reach its bottom:\n%s",
			got,
		)
	})

	t.Run("a hint scrolls the region into view", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(frame)
		s.Push(scrollHintStub{stubDialog: stubDialog{body: body}, top: 27, height: 1})
		got := s.View(backdrop(40), 40, 40)
		require.Equal(t, 1, shows(got, "L27"), "hinted line should scroll into view:\n%s", got)
		require.Zero(
			t,
			shows(got, "L00"),
			"scrolling to the bottom should leave the top off-screen:\n%s",
			got,
		)
	})
}

// footeredStub scrolls its body (via the embedded scroll hint) and pins a
// footer, so the Frame's footer-pinning is observable: the footer must show
// even when the body has scrolled its own top off.
type footeredStub struct {
	scrollHintStub

	footer string
}

func (d footeredStub) Footer() string { return d.footer }

func TestFramePinsFooterWhileBodyScrolls(t *testing.T) {
	t.Parallel()

	lines := make([]string, 30)
	for i := range lines {
		lines[i] = fmt.Sprintf("L%02d", i)
	}
	frame := dialog.NewFrame(dialog.FrameConfig{
		MaxHeight: 12,
		Styles:    dialog.Styles{Box: lg.NewStyle().Border(lg.NormalBorder())},
	})

	s := dialog.New(frame)
	// Focus is on the last body line; the footer must still be pinned on screen.
	s.Push(footeredStub{
		scrollHintStub: scrollHintStub{
			stubDialog: stubDialog{body: strings.Join(lines, "\n")},
			top:        29,
			height:     1,
		},
		footer: "PINNED-FOOT",
	})
	got := s.View(backdrop(40), 40, 40)
	require.Equal(
		t,
		1,
		shows(got, "PINNED-FOOT"),
		"footer must stay pinned while the body scrolls:\n%s",
		got,
	)
	require.Equal(
		t,
		1,
		shows(got, "L29"),
		"body should have scrolled to the focused last line:\n%s",
		got,
	)
	require.Zero(
		t,
		shows(got, "L00"),
		"scrolling to the bottom should leave the body top off-screen:\n%s",
		got,
	)
}

func TestStackView(t *testing.T) {
	t.Parallel()

	// A screenH-row backdrop with no trailing newline, so its measured height
	// is exactly the screen height the view is placed over.
	dotted := strings.TrimSuffix(strings.Repeat(strings.Repeat("·", 40)+"\n", 12), "\n")

	t.Run("inactive stack returns the backdrop unchanged", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(dialog.NewFrame(dialog.FrameConfig{}))
		require.Equal(t, dotted, s.View(dotted, 40, 12), "inactive View mutated the backdrop")
	})

	t.Run("active stack overlays the dialog body", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(dialog.NewFrame(dialog.FrameConfig{MaxWidth: 30, MaxHeight: 8}))
		s.Push(stubDialog{id: "a", title: "hello", body: "the dialog body"})

		got := s.View(dotted, 40, 12)
		require.Equal(t, 1, shows(got, "the dialog body"), "framed view is missing the dialog body")
		// overlay.Place composites onto the screenH-row backdrop, so a clamped
		// dialog can never make the view taller than the screen.
		require.Equal(t, 12, lg.Height(got), "framed view must match the screen height")
	})

	t.Run("stacked dialogs remain visible beneath the foreground", func(t *testing.T) {
		t.Parallel()

		s := dialog.New(borderedFrame())
		s.Push(stubDialog{body: "UNDERLAY LEFT                 UNDERLAY RIGHT"})
		s.Push(stubDialog{body: "FOREGROUND"})
		got := s.View(backdrop(12), 60, 12)
		require.Equal(t, 1, shows(got, "UNDERLAY LEFT"))
		require.Equal(t, 1, shows(got, "UNDERLAY RIGHT"))
		require.Equal(t, 1, shows(got, "FOREGROUND"))
	})

	t.Run("a body taller than the cap stays within the screen", func(t *testing.T) {
		t.Parallel()

		var tall strings.Builder
		for i := range 40 {
			tall.WriteString("line ")
			tall.WriteByte(byte('a' + i%26))
			tall.WriteByte('\n')
		}
		s := dialog.New(dialog.NewFrame(dialog.FrameConfig{MaxWidth: 30, MaxHeight: 8}))
		s.Push(stubDialog{id: "tall", body: tall.String()})

		require.Equal(
			t,
			20,
			lg.Height(s.View(backdrop(20), 40, 20)),
			"framed view must match the screen height",
		)
	})
}

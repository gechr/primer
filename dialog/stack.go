package dialog

import (
	"reflect"
	"time"

	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/gechr/primer/overlay"
	"github.com/gechr/primer/scrollbar"
)

// The input grace a dialog pushed from an async result opens behind. A dialog
// that appears under the user's fingers (a picker landing mid-keymash) must
// not let the in-flight keystroke filter or action it, so keys are absorbed
// until the keyboard has been quiet for graceQuiet. graceCeiling bounds the
// absorption so key auto-repeat can never hold the dialog shut, and
// graceReopenExempt skips the grace when the same kind of dialog just closed -
// in a rapid prompt loop each deliberate key must land, not feed a grace.
const (
	graceQuiet        = 200 * time.Millisecond
	graceCeiling      = 1500 * time.Millisecond
	graceReopenExempt = 500 * time.Millisecond
)

// Stack owns an ordered set of modal dialogs - the top is the last element,
// the one the user interacts with - plus the Frame that renders them. Only the
// top dialog is live; the rest wait beneath it and are redrawn only when they
// return to the top. The zero value is unusable; construct with New.
type Stack struct {
	frame   Frame
	dialogs []Dialog
	// gracedType records, per stacked dialog, the concrete type it was
	// PushWithGrace'd as (nil for a plain push). Only a popped graced dialog
	// feeds the same-kind reopen exemption - a user-closed picker must not
	// disarm the grace of an async pick that happens to land right after it -
	// and the type is captured at push time because Update's contract lets a
	// dialog return a different Dialog mid-life.
	gracedType []reflect.Type

	// now is the clock the grace windows are measured on - injected so tests
	// drive it instead of sleeping.
	now func() time.Time

	// Grace state for the top dialog (see PushWithGrace). graceActive gates
	// the whole mechanism; graceOpenedAt anchors the ceiling and
	// graceLastKeyAt the quiet window.
	graceActive    bool
	graceOpenedAt  time.Time
	graceLastKeyAt time.Time

	// The most recently popped graced dialog's concrete type and close time,
	// for the same-kind reopen exemption.
	lastClosedType reflect.Type
	lastClosedAt   time.Time

	// geo is the top dialog's placement from the last View, for translating
	// mouse coordinates into content space; every render refreshes it and any
	// push or pop invalidates it, so a click can never be translated against a
	// dialog that is no longer on top.
	geo frameGeometry
}

// New returns an empty Stack that frames its dialogs with frame.
func New(frame Frame) *Stack {
	return &Stack{frame: frame, now: time.Now}
}

// Push opens d as the new top dialog. A push is user-driven, so it also ends
// any input grace a previous top was holding.
func (s *Stack) Push(d Dialog) {
	s.dialogs = append(s.dialogs, d)
	s.gracedType = append(s.gracedType, nil)
	s.graceActive = false
	s.geo = frameGeometry{}
	s.frame.resetScroll()
}

// PushWithGrace opens d as the new top dialog behind a brief input grace, for
// dialogs that open on an async result (a picker pushed when a fetch lands)
// rather than on a keypress. Keys are absorbed until the keyboard has been
// quiet for graceQuiet, bounded by graceCeiling from the open; every other
// message routes normally. When a graced dialog of the same concrete type
// closed within graceReopenExempt, the grace is skipped - a rapid prompt loop
// is user-driven by then, and each deliberate key must land.
func (s *Stack) PushWithGrace(d Dialog) {
	s.Push(d)
	s.gracedType[len(s.gracedType)-1] = reflect.TypeOf(d)
	n := s.now()
	if reflect.TypeOf(d) == s.lastClosedType && n.Sub(s.lastClosedAt) < graceReopenExempt {
		return
	}
	s.graceActive = true
	s.graceOpenedAt = n
	s.graceLastKeyAt = n
}

// SetFrame swaps the frame the Stack draws its dialogs with, letting the owner
// refresh chrome styles (e.g. after a theme change) without dropping any open
// dialog.
func (s *Stack) SetFrame(frame Frame) {
	s.frame = frame
}

// SetClock replaces the clock the input-grace windows are measured on - the
// injected-now seam that lets tests drive the grace deterministically instead
// of sleeping through it.
func (s *Stack) SetClock(now func() time.Time) {
	s.now = now
}

// Active reports whether any dialog is open.
func (s *Stack) Active() bool {
	return len(s.dialogs) > 0
}

// Top returns the dialog currently receiving input, or nil when none is open.
func (s *Stack) Top() Dialog {
	if !s.Active() {
		return nil
	}
	return s.dialogs[len(s.dialogs)-1]
}

// Update routes a message to the top dialog. When it reports ResultSubmit or
// ResultClose it is popped and returned, so the caller can type-assert it and
// read a typed payload; otherwise the returned Dialog is nil. While an input
// grace is active (see PushWithGrace) key presses are absorbed instead of
// routed; the first key that arrives after the quiet window - or after the
// ceiling - ends the grace and drives the dialog.
func (s *Stack) Update(msg tea.Msg) (tea.Cmd, Dialog, Result) {
	if !s.Active() {
		return nil, nil, ResultNone
	}
	if _, ok := msg.(tea.KeyPressMsg); ok && s.graceActive {
		n := s.now()
		if n.Sub(s.graceOpenedAt) >= graceCeiling || n.Sub(s.graceLastKeyAt) >= graceQuiet {
			s.graceActive = false
		} else {
			s.graceLastKeyAt = n
			return nil, nil, ResultNone
		}
	}
	if _, ok := msg.(tea.KeyPressMsg); ok {
		s.frame.manualScroll = false
	}
	// A left click inside the framed content is translated into the dialog's
	// own coordinate space and delivered as a ClickMsg; any other click (the
	// chrome, outside the box, another button) routes through untranslated.
	if mc, ok := msg.(tea.MouseClickMsg); ok && mc.Button == tea.MouseLeft {
		if x, y, ok := s.geo.translate(mc.X, mc.Y); ok {
			msg = ClickMsg{X: x, Y: y}
		}
	}
	top := len(s.dialogs) - 1
	next, cmd, res := s.dialogs[top].Update(msg)
	s.dialogs[top] = next
	if res == ResultSubmit || res == ResultClose {
		s.dialogs = s.dialogs[:top]
		if t := s.gracedType[top]; t != nil {
			s.lastClosedType = t
			s.lastClosedAt = s.now()
		}
		s.gracedType = s.gracedType[:top]
		s.graceActive = false
		s.geo = frameGeometry{}
		s.frame.resetScroll()
		return cmd, next, res
	}
	return cmd, nil, ResultNone
}

// Pop removes and returns the top dialog without a result, or nil when none
// is open - the seam an owner uses to close a dialog it resolved itself, such
// as a held-open form (see [Form.HoldOnSubmit]) whose async write completed.
// For the grace bookkeeping it counts as a close, so a graced dialog of the
// same kind reopening right after skips its grace exactly as a keyed close
// would.
func (s *Stack) Pop() Dialog {
	if !s.Active() {
		return nil
	}
	top := len(s.dialogs) - 1
	d := s.dialogs[top]
	s.dialogs = s.dialogs[:top]
	if t := s.gracedType[top]; t != nil {
		s.lastClosedType = t
		s.lastClosedAt = s.now()
	}
	s.gracedType = s.gracedType[:top]
	s.graceActive = false
	s.geo = frameGeometry{}
	s.frame.resetScroll()
	return d
}

// ScrollbarHitbox returns the screen-space hitbox of the top dialog's internal
// scrollbar as of the last View, for owners that hit-test mouse presses or
// drags against it. It reports false when no dialog is open, the body fits
// without scrolling, or nothing has rendered since the stack last changed.
func (s *Stack) ScrollbarHitbox() (scrollbar.Hitbox, bool) {
	return s.geo.bar, s.geo.valid && s.geo.hasBar
}

// ScrollBy moves the top dialog's internal viewport by delta lines. It returns
// false when the dialog has not rendered yet or does not overflow its frame.
func (s *Stack) ScrollBy(delta int) bool {
	if !s.geo.valid || !s.geo.hasBar || delta == 0 {
		return false
	}
	if delta > 0 {
		s.frame.viewport.ScrollDown(delta)
	} else {
		s.frame.viewport.ScrollUp(-delta)
	}
	s.frame.manualScroll = true
	return true
}

// SetScrollOffset moves the top dialog's internal viewport to offset. It
// returns false when the dialog has not rendered yet or does not overflow.
func (s *Stack) SetScrollOffset(offset int) bool {
	if !s.geo.valid || !s.geo.hasBar {
		return false
	}
	s.frame.viewport.SetYOffset(offset)
	s.frame.manualScroll = true
	return true
}

// ScrollPercent reports the top dialog's internal scroll position. A dialog
// without an active scrollbar reports zero.
func (s *Stack) ScrollPercent() float64 {
	if !s.geo.valid || !s.geo.hasBar {
		return 0
	}
	return s.frame.viewport.ScrollPercent()
}

// View composites the top dialog over backdrop, which must already be screenW
// columns by screenH rows. With no dialog open it returns backdrop unchanged.
func (s *Stack) View(backdrop string, screenW, screenH int) string {
	if !s.Active() {
		s.geo = frameGeometry{}
		return backdrop
	}
	top := s.Top()
	// A self-framed dialog draws its own border and sizing; place it verbatim
	// rather than boxing and scrolling it through the Frame, which would clip
	// and interleave the dialog's pre-drawn box. Its whole rendering is its
	// content, so clicks translate against the placement origin alone.
	if sf, ok := top.(SelfFramed); ok && sf.SelfFramed() {
		content := top.Content(screenW)
		x, y := overlay.Origin(content, screenW, screenH, overlay.Center)
		s.geo = frameGeometry{
			contentX: x,
			contentY: y,
			contentW: lg.Width(content),
			contentH: lg.Height(content),
			valid:    true,
		}
		return overlay.Place(backdrop, content, screenW, screenH, overlay.Center)
	}
	var out string
	out, s.geo = s.frame.frame(backdrop, top, screenW, screenH)
	return out
}

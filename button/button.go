// Package button renders a horizontal row of focusable buttons - "No  Yes",
// or a lone "OK" - the affordance a confirm dialog puts under its prompt.
// One button holds focus; the arrow keys (via [Row.Step]) move it with wrap,
// and [Row.Hit] maps a clicked column back to a button so a mouse can press
// one directly.
//
// Like pill and titlebox it is style-agnostic: each button carries its own
// focused and blurred styles, so a caller renders a green Yes and a red No -
// or any other pair - without the package knowing.
package button

import (
	"strings"

	lg "charm.land/lipgloss/v2"
)

// defaultGap separates adjacent buttons when Row.Gap is empty.
const defaultGap = "  "

// Button is one pressable label with its two looks. The focused style renders
// when the button holds the row's focus, the blurred style otherwise; padding
// and background belong to the styles, so a pill-shaped button is just a
// style with a background and Padding(0, 1).
type Button struct {
	Label   string
	Focused lg.Style
	Blurred lg.Style
}

// Row is an ordered set of buttons with one focus. The zero value renders
// nothing and hits nothing.
type Row struct {
	Buttons []Button
	// Focus is the focused button's index; out-of-range renders every button
	// blurred.
	Focus int
	// Gap separates adjacent buttons; empty means two spaces.
	Gap string
}

// Step moves the focus by delta, wrapping at both ends.
func (r *Row) Step(delta int) {
	n := len(r.Buttons)
	if n == 0 {
		return
	}
	r.Focus = (r.Focus + delta + n) % n
}

// View renders the buttons side by side, the focused one in its focused style.
func (r *Row) View() string {
	var b strings.Builder
	for i, btn := range r.Buttons {
		if i > 0 {
			b.WriteString(r.gap())
		}
		b.WriteString(r.render(i, btn))
	}
	return b.String()
}

// Hit returns the index of the button occupying column x of the rendered row
// (0-based from the row's first column), or -1 when x falls in a gap or
// outside the row. Widths are measured on the styled rendering, so a style's
// padding counts as pressable surface.
func (r *Row) Hit(x int) int {
	if x < 0 {
		return -1
	}
	gapW := lg.Width(r.gap())
	col := 0
	for i, btn := range r.Buttons {
		if i > 0 {
			col += gapW
		}
		w := lg.Width(r.render(i, btn))
		if x >= col && x < col+w {
			return i
		}
		col += w
	}
	return -1
}

// render draws button i in its focus-appropriate style.
func (r *Row) render(i int, btn Button) string {
	if i == r.Focus {
		return btn.Focused.Render(btn.Label)
	}
	return btn.Blurred.Render(btn.Label)
}

// gap is the separator between adjacent buttons.
func (r *Row) gap() string {
	if r.Gap == "" {
		return defaultGap
	}
	return r.Gap
}

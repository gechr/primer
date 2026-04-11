package scrollbar

import "math"

// Hitbox describes the screen region occupied by a vertical scrollbar.
type Hitbox struct {
	X          int // column position
	Y          int // top row
	Height     int // visible track height
	TotalLines int // total content lines
}

// Contains reports whether the screen coordinate (x, y) falls within the hitbox.
func (h Hitbox) Contains(x, y int) bool {
	return x == h.X && y >= h.Y && y < h.Y+h.Height
}

// Drag tracks the state of a scrollbar drag interaction.
//
// Typical usage in a Bubble Tea Update loop:
//
//	case tea.MouseClickMsg:
//	    offset := drag.Press(hitbox, msg.Y, vp.ScrollPercent())
//	    vp.SetYOffset(offset)
//
//	case tea.MouseMotionMsg:
//	    if offset, ok := drag.Motion(hitbox, msg.Y); ok {
//	        vp.SetYOffset(offset)
//	    }
//
//	case tea.MouseReleaseMsg:
//	    drag.Release()
type Drag struct {
	Active bool
	grab   int
}

// Press begins a drag at the given mouse position. It computes the grab
// offset (preserving relative position if clicking on the thumb, centering
// otherwise) and returns the viewport offset to scroll to.
//
// The scrollPercent parameter is the viewport's current scroll position
// (0.0 to 1.0), used to determine the thumb's current position.
func (d *Drag) Press(h Hitbox, mouseY int, scrollPercent float64) int {
	const thumbCenterDivisor = 2

	thumbPos, thumbSize := ThumbMetrics(h.Height, h.TotalLines, scrollPercent)
	row := min(max(mouseY-h.Y, 0), h.Height-1)

	grab := thumbSize / thumbCenterDivisor
	if row >= thumbPos && row < thumbPos+thumbSize {
		grab = row - thumbPos
	}

	d.Active = true
	d.grab = grab

	return scrollToRow(h, row, grab)
}

// Motion updates the drag position and returns the new viewport offset.
// Returns ok=false if no drag is active.
func (d *Drag) Motion(h Hitbox, mouseY int) (int, bool) {
	if !d.Active {
		return 0, false
	}
	return scrollToRow(h, mouseY-h.Y, d.grab), true
}

// Release ends the drag interaction.
func (d *Drag) Release() {
	*d = Drag{}
}

// scrollToRow computes the viewport offset for a drag at the given track row.
func scrollToRow(h Hitbox, row, grab int) int {
	if h.Height <= 0 {
		return 0
	}

	maxOffset := max(0, h.TotalLines-h.Height)
	if maxOffset == 0 {
		return 0
	}

	// Thumb size is independent of scroll percent — pass 0.
	_, thumbSize := ThumbMetrics(h.Height, h.TotalLines, 0)
	trackSpace := max(0, h.Height-thumbSize)
	topRow := min(max(row-grab, 0), trackSpace)

	if trackSpace == 0 {
		return maxOffset
	}

	return int(math.Round(float64(topRow) / float64(trackSpace) * float64(maxOffset)))
}

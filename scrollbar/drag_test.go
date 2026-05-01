package scrollbar_test

import (
	"testing"

	"github.com/gechr/primer/scrollbar"
	"github.com/stretchr/testify/require"
)

func TestHitboxContains(t *testing.T) {
	h := scrollbar.Hitbox{X: 79, Y: 2, Height: 10, TotalLines: 40}

	require.True(t, h.Contains(79, 2))
	require.True(t, h.Contains(79, 11))
	require.False(t, h.Contains(79, 1))
	require.False(t, h.Contains(79, 12))
	require.False(t, h.Contains(78, 5))
}

func TestHitboxContainsThumb(t *testing.T) {
	h := scrollbar.Hitbox{X: 79, Y: 2, Height: 10, TotalLines: 40}

	require.True(t, h.ContainsThumb(79, 6, 0.5))
	require.False(t, h.ContainsThumb(79, 2, 0.5))
	require.False(t, h.ContainsThumb(78, 6, 0.5))
}

func TestHitboxContainsThumbUsesConfig(t *testing.T) {
	h := scrollbar.Hitbox{
		X:          79,
		Y:          0,
		Height:     10,
		TotalLines: 11,
		Config:     scrollbar.Config{MaxThumbDivisor: 1},
	}

	require.True(t, h.ContainsThumb(79, 8, 0))
}

func TestDragPressReturnsOffset(t *testing.T) {
	h := scrollbar.Hitbox{X: 79, Y: 0, Height: 10, TotalLines: 40}
	var d scrollbar.Drag

	// Click at the bottom of the track.
	offset := d.Press(h, 9, 0.0)

	require.True(t, d.Active)
	require.Positive(t, offset)
}

func TestDragPressAtTopReturnsZero(t *testing.T) {
	h := scrollbar.Hitbox{X: 79, Y: 0, Height: 10, TotalLines: 40}
	var d scrollbar.Drag

	offset := d.Press(h, 0, 0.0)

	require.True(t, d.Active)
	require.Zero(t, offset)
}

func TestDragPressAtBottomReturnsMaxOffset(t *testing.T) {
	h := scrollbar.Hitbox{X: 79, Y: 0, Height: 10, TotalLines: 40}
	var d scrollbar.Drag

	offset := d.Press(h, 9, 0.0)

	require.True(t, d.Active)
	require.LessOrEqual(t, offset, 30) // maxOffset = 40 - 10
}

func TestDragMotionWhenInactive(t *testing.T) {
	h := scrollbar.Hitbox{X: 79, Y: 0, Height: 10, TotalLines: 40}
	var d scrollbar.Drag

	_, ok := d.Motion(h, 5)
	require.False(t, ok)
}

func TestDragMotionAfterPress(t *testing.T) {
	h := scrollbar.Hitbox{X: 79, Y: 0, Height: 10, TotalLines: 40}
	var d scrollbar.Drag

	d.Press(h, 0, 0.0)

	offset, ok := d.Motion(h, 9)
	require.True(t, ok)
	require.Positive(t, offset)
}

func TestDragMotionMovesViewport(t *testing.T) {
	h := scrollbar.Hitbox{X: 79, Y: 0, Height: 10, TotalLines: 60}
	var d scrollbar.Drag

	// Press at top.
	offset1 := d.Press(h, 0, 0.0)

	// Drag to bottom.
	offset2, ok := d.Motion(h, 9)
	require.True(t, ok)
	require.Greater(t, offset2, offset1)
}

func TestDragRelease(t *testing.T) {
	h := scrollbar.Hitbox{X: 79, Y: 0, Height: 10, TotalLines: 40}
	var d scrollbar.Drag

	d.Press(h, 5, 0.0)
	require.True(t, d.Active)

	d.Release()
	require.False(t, d.Active)

	_, ok := d.Motion(h, 5)
	require.False(t, ok)
}

func TestDragPressOnThumbPreservesGrabOffset(t *testing.T) {
	h := scrollbar.Hitbox{X: 79, Y: 0, Height: 10, TotalLines: 40}
	var d scrollbar.Drag

	// Press in the middle of the track (thumb should be near top at 50% scroll).
	offset := d.Press(h, 5, 0.5)
	require.True(t, d.Active)

	// The offset should be near the midpoint of maxOffset (30).
	require.InDelta(t, 15, offset, 8)
}

func TestDragContentFitsViewport(t *testing.T) {
	h := scrollbar.Hitbox{X: 79, Y: 0, Height: 10, TotalLines: 5}
	var d scrollbar.Drag

	offset := d.Press(h, 5, 0.0)
	require.Zero(t, offset)
}

func TestDragWithVerticalOffset(t *testing.T) {
	// Scrollbar starts at row 3 (e.g. below a header + separator).
	h := scrollbar.Hitbox{X: 79, Y: 3, Height: 10, TotalLines: 40}
	var d scrollbar.Drag

	// Click at absolute Y=3 should be row 0 in the track → top.
	offset := d.Press(h, 3, 0.0)
	require.Zero(t, offset)

	// Click at absolute Y=12 should be row 9 → near bottom.
	d.Release()
	offset = d.Press(h, 12, 0.0)
	require.Positive(t, offset)
}

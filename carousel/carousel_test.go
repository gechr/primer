package carousel_test

import (
	"testing"

	"github.com/gechr/primer/carousel"
	"github.com/stretchr/testify/require"
)

func TestNextWrapsAroundEnd(t *testing.T) {
	t.Parallel()

	m := carousel.New("a", "b", "c")
	m.Next()
	require.Equal(t, 1, m.Active())

	m.Next()
	m.Next() // from index 2 wraps to 0
	require.Equal(t, 0, m.Active())
}

func TestPrevWrapsAroundStart(t *testing.T) {
	t.Parallel()

	m := carousel.New("a", "b", "c")
	m.Prev() // from 0 wraps to last
	require.Equal(t, 2, m.Active())
}

func TestEmptyCarousel(t *testing.T) {
	t.Parallel()

	m := carousel.New()
	require.Equal(t, -1, m.Active())
	require.Empty(t, m.ActiveItem())

	m.Next() // must not panic
	m.Prev()
	require.Empty(t, m.View())
}

func TestSetItemsClampsActive(t *testing.T) {
	t.Parallel()

	m := carousel.New("a", "b", "c", "d")
	m.SetActive(3)
	m.SetItems([]string{"x", "y"})
	require.Equal(t, 1, m.Active())
	require.Equal(t, "y", m.ActiveItem())
}

func TestItemsReturnsCopy(t *testing.T) {
	t.Parallel()

	m := carousel.New("a", "b")
	items := m.Items()
	items[0] = "tampered"
	require.Equal(t, []string{"a", "b"}, m.Items())
}

package dialog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrameClampWidth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		cfg    FrameConfig
		screen int
		want   int
	}{
		{name: "no caps returns the screen", cfg: FrameConfig{}, screen: 100, want: 100},
		{name: "margin is kept clear", cfg: FrameConfig{Margin: 6}, screen: 100, want: 94},
		{name: "absolute max binds", cfg: FrameConfig{MaxWidth: 40}, screen: 100, want: 40},
		{name: "fraction binds", cfg: FrameConfig{WidthFraction: 0.5}, screen: 100, want: 50},
		{
			name:   "smallest of all caps wins",
			cfg:    FrameConfig{MaxWidth: 60, WidthFraction: 0.5, Margin: 6},
			screen: 100,
			want:   50,
		},
		{name: "floors at one column", cfg: FrameConfig{Margin: 200}, screen: 10, want: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			frame := NewFrame(tc.cfg)
			require.Equal(t, tc.want, frame.clampWidth(tc.screen))
		})
	}
}

func TestFrameClampHeight(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		cfg    FrameConfig
		screen int
		want   int
	}{
		{name: "no caps returns the screen", cfg: FrameConfig{}, screen: 50, want: 50},
		{name: "margin is kept clear", cfg: FrameConfig{Margin: 4}, screen: 50, want: 46},
		{name: "absolute max binds", cfg: FrameConfig{MaxHeight: 30}, screen: 50, want: 30},
		{name: "fraction binds", cfg: FrameConfig{HeightFraction: 0.8}, screen: 50, want: 40},
		{
			name:   "smallest of all caps wins",
			cfg:    FrameConfig{MaxHeight: 30, HeightFraction: 0.8, Margin: 4},
			screen: 50,
			want:   30,
		},
		{name: "floors at one row", cfg: FrameConfig{Margin: 100}, screen: 10, want: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			frame := NewFrame(tc.cfg)
			require.Equal(t, tc.want, frame.clampHeight(tc.screen))
		})
	}
}

func TestScrollOffset(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		top, height, vph int
		want             int
	}{
		{name: "region above the fold stays at top", top: 0, height: 3, vph: 10, want: 0},
		{
			name:   "region straddling the fold scrolls just enough",
			top:    8,
			height: 3,
			vph:    10,
			want:   1,
		},
		{
			name:   "region far below scrolls to reveal its bottom",
			top:    20,
			height: 2,
			vph:    10,
			want:   12,
		},
		{name: "region taller than the window keeps its top", top: 5, height: 20, vph: 10, want: 5},
		{name: "exact bottom fit needs no scroll", top: 8, height: 2, vph: 10, want: 0},
		{name: "zero height counts as one line", top: 10, height: 0, vph: 10, want: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, scrollOffset(tc.top, tc.height, tc.vph))
		})
	}
}

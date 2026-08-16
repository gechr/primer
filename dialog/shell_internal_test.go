package dialog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShellClampWidth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		cfg    ShellConfig
		screen int
		want   int
	}{
		{name: "no caps returns the screen", cfg: ShellConfig{}, screen: 100, want: 100},
		{name: "margin is kept clear", cfg: ShellConfig{Margin: 6}, screen: 100, want: 94},
		{name: "absolute max binds", cfg: ShellConfig{MaxWidth: 40}, screen: 100, want: 40},
		{name: "fraction binds", cfg: ShellConfig{WidthFraction: 0.5}, screen: 100, want: 50},
		{
			name:   "smallest of all caps wins",
			cfg:    ShellConfig{MaxWidth: 60, WidthFraction: 0.5, Margin: 6},
			screen: 100,
			want:   50,
		},
		{name: "floors at one column", cfg: ShellConfig{Margin: 200}, screen: 10, want: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, NewShell(tc.cfg).clampWidth(tc.screen))
		})
	}
}

func TestShellClampHeight(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		cfg    ShellConfig
		screen int
		want   int
	}{
		{name: "no caps returns the screen", cfg: ShellConfig{}, screen: 50, want: 50},
		{name: "margin is kept clear", cfg: ShellConfig{Margin: 4}, screen: 50, want: 46},
		{name: "absolute max binds", cfg: ShellConfig{MaxHeight: 30}, screen: 50, want: 30},
		{name: "fraction binds", cfg: ShellConfig{HeightFraction: 0.8}, screen: 50, want: 40},
		{
			name:   "smallest of all caps wins",
			cfg:    ShellConfig{MaxHeight: 30, HeightFraction: 0.8, Margin: 4},
			screen: 50,
			want:   30,
		},
		{name: "floors at one row", cfg: ShellConfig{Margin: 100}, screen: 10, want: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, NewShell(tc.cfg).clampHeight(tc.screen))
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

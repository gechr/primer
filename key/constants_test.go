package key_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gechr/primer/key"
	"github.com/stretchr/testify/require"
)

// Keep the string constants compatible with real events, including keypad
// keys whose canonical strings differ from their friendly constant names.
func TestNamedConstantsMatchBubbleTea(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		code  rune
	}{
		{"Enter", key.Enter, tea.KeyEnter},
		{"Tab", key.Tab, tea.KeyTab},
		{"Backspace", key.Backspace, tea.KeyBackspace},
		{"Escape", key.Escape, tea.KeyEscape},
		{"Space", key.Space, tea.KeySpace},
		{"Up", key.Up, tea.KeyUp},
		{"Down", key.Down, tea.KeyDown},
		{"Left", key.Left, tea.KeyLeft},
		{"Right", key.Right, tea.KeyRight},
		{"Insert", key.Insert, tea.KeyInsert},
		{"Delete", key.Delete, tea.KeyDelete},
		{"PageUp", key.PageUp, tea.KeyPgUp},
		{"PageDown", key.PageDown, tea.KeyPgDown},
		{"Home", key.Home, tea.KeyHome},
		{"End", key.End, tea.KeyEnd},
		{"KeypadEnter", key.KeypadEnter, tea.KeyKpEnter},
		{"KeypadMultiply", key.KeypadMultiply, tea.KeyKpMultiply},
		{"KeypadPlus", key.KeypadPlus, tea.KeyKpPlus},
		{"KeypadMinus", key.KeypadMinus, tea.KeyKpMinus},
		{"KeypadDecimal", key.KeypadDecimal, tea.KeyKpDecimal},
		{"KeypadDivide", key.KeypadDivide, tea.KeyKpDivide},
		{"KeypadZero", key.KeypadZero, tea.KeyKp0},
		{"KeypadOne", key.KeypadOne, tea.KeyKp1},
		{"KeypadTwo", key.KeypadTwo, tea.KeyKp2},
		{"KeypadThree", key.KeypadThree, tea.KeyKp3},
		{"KeypadFour", key.KeypadFour, tea.KeyKp4},
		{"KeypadFive", key.KeypadFive, tea.KeyKp5},
		{"KeypadSix", key.KeypadSix, tea.KeyKp6},
		{"KeypadSeven", key.KeypadSeven, tea.KeyKp7},
		{"KeypadEight", key.KeypadEight, tea.KeyKp8},
		{"KeypadNine", key.KeypadNine, tea.KeyKp9},
		{"KeypadUp", key.KeypadUp, tea.KeyKpUp},
		{"KeypadDown", key.KeypadDown, tea.KeyKpDown},
		{"KeypadLeft", key.KeypadLeft, tea.KeyKpLeft},
		{"KeypadRight", key.KeypadRight, tea.KeyKpRight},
		{"KeypadPageUp", key.KeypadPageUp, tea.KeyKpPgUp},
		{"KeypadPageDown", key.KeypadPageDown, tea.KeyKpPgDown},
		{"KeypadHome", key.KeypadHome, tea.KeyKpHome},
		{"KeypadEnd", key.KeypadEnd, tea.KeyKpEnd},
		{"KeypadInsert", key.KeypadInsert, tea.KeyKpInsert},
		{"KeypadDelete", key.KeypadDelete, tea.KeyKpDelete},
		{"F1", key.F1, tea.KeyF1},
		{"F2", key.F2, tea.KeyF2},
		{"F3", key.F3, tea.KeyF3},
		{"F4", key.F4, tea.KeyF4},
		{"F5", key.F5, tea.KeyF5},
		{"F6", key.F6, tea.KeyF6},
		{"F7", key.F7, tea.KeyF7},
		{"F8", key.F8, tea.KeyF8},
		{"F9", key.F9, tea.KeyF9},
		{"F10", key.F10, tea.KeyF10},
		{"F11", key.F11, tea.KeyF11},
		{"F12", key.F12, tea.KeyF12},
		{"F13", key.F13, tea.KeyF13},
		{"F14", key.F14, tea.KeyF14},
		{"F15", key.F15, tea.KeyF15},
		{"F16", key.F16, tea.KeyF16},
		{"F17", key.F17, tea.KeyF17},
		{"F18", key.F18, tea.KeyF18},
		{"F19", key.F19, tea.KeyF19},
		{"F20", key.F20, tea.KeyF20},
		{"F21", key.F21, tea.KeyF21},
		{"F22", key.F22, tea.KeyF22},
		{"F23", key.F23, tea.KeyF23},
		{"F24", key.F24, tea.KeyF24},
		{"CapsLock", key.CapsLock, tea.KeyCapsLock},
		{"ScrollLock", key.ScrollLock, tea.KeyScrollLock},
		{"NumLock", key.NumLock, tea.KeyNumLock},
		{"PrintScreen", key.PrintScreen, tea.KeyPrintScreen},
		{"Pause", key.Pause, tea.KeyPause},
		{"Menu", key.Menu, tea.KeyMenu},
		{"MediaPlay", key.MediaPlay, tea.KeyMediaPlay},
		{"MediaPause", key.MediaPause, tea.KeyMediaPause},
		{"MediaPlayPause", key.MediaPlayPause, tea.KeyMediaPlayPause},
		{"MediaReverse", key.MediaReverse, tea.KeyMediaReverse},
		{"MediaStop", key.MediaStop, tea.KeyMediaStop},
		{"MediaFastForward", key.MediaFastForward, tea.KeyMediaFastForward},
		{"MediaRewind", key.MediaRewind, tea.KeyMediaRewind},
		{"MediaNext", key.MediaNext, tea.KeyMediaNext},
		{"MediaPrevious", key.MediaPrevious, tea.KeyMediaPrev},
		{"MediaRecord", key.MediaRecord, tea.KeyMediaRecord},
		{"VolumeDown", key.VolumeDown, tea.KeyLowerVol},
		{"VolumeUp", key.VolumeUp, tea.KeyRaiseVol},
		{"Mute", key.Mute, tea.KeyMute},
		{"LeftShift", key.LeftShift, tea.KeyLeftShift},
		{"LeftAlt", key.LeftAlt, tea.KeyLeftAlt},
		{"LeftCtrl", key.LeftCtrl, tea.KeyLeftCtrl},
		{"LeftSuper", key.LeftSuper, tea.KeyLeftSuper},
		{"LeftMeta", key.LeftMeta, tea.KeyLeftMeta},
		{"RightShift", key.RightShift, tea.KeyRightShift},
		{"RightAlt", key.RightAlt, tea.KeyRightAlt},
		{"RightCtrl", key.RightCtrl, tea.KeyRightCtrl},
		{"RightSuper", key.RightSuper, tea.KeyRightSuper},
		{"RightMeta", key.RightMeta, tea.KeyRightMeta},
		{"Esc", key.Esc, tea.KeyEsc},
		{"Return", key.Return, tea.KeyReturn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tea.KeyPressMsg{Code: tt.code}.String(), tt.value)
		})
	}
}

func TestAlphanumericConstantsMatchBubbleTea(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mod    tea.KeyMod
		values []string
	}{
		{
			"plain",
			0,
			[]string{
				key.A,
				key.B,
				key.C,
				key.D,
				key.E,
				key.F,
				key.G,
				key.H,
				key.I,
				key.J,
				key.K,
				key.L,
				key.M,
				key.N,
				key.O,
				key.P,
				key.Q,
				key.R,
				key.S,
				key.T,
				key.U,
				key.V,
				key.W,
				key.X,
				key.Y,
				key.Z,
				key.Zero,
				key.One,
				key.Two,
				key.Three,
				key.Four,
				key.Five,
				key.Six,
				key.Seven,
				key.Eight,
				key.Nine,
			},
		},
		{
			"ctrl",
			tea.ModCtrl,
			[]string{
				key.CtrlA,
				key.CtrlB,
				key.CtrlC,
				key.CtrlD,
				key.CtrlE,
				key.CtrlF,
				key.CtrlG,
				key.CtrlH,
				key.CtrlI,
				key.CtrlJ,
				key.CtrlK,
				key.CtrlL,
				key.CtrlM,
				key.CtrlN,
				key.CtrlO,
				key.CtrlP,
				key.CtrlQ,
				key.CtrlR,
				key.CtrlS,
				key.CtrlT,
				key.CtrlU,
				key.CtrlV,
				key.CtrlW,
				key.CtrlX,
				key.CtrlY,
				key.CtrlZ,
				key.CtrlZero,
				key.CtrlOne,
				key.CtrlTwo,
				key.CtrlThree,
				key.CtrlFour,
				key.CtrlFive,
				key.CtrlSix,
				key.CtrlSeven,
				key.CtrlEight,
				key.CtrlNine,
			},
		},
		{
			"alt",
			tea.ModAlt,
			[]string{
				key.AltA,
				key.AltB,
				key.AltC,
				key.AltD,
				key.AltE,
				key.AltF,
				key.AltG,
				key.AltH,
				key.AltI,
				key.AltJ,
				key.AltK,
				key.AltL,
				key.AltM,
				key.AltN,
				key.AltO,
				key.AltP,
				key.AltQ,
				key.AltR,
				key.AltS,
				key.AltT,
				key.AltU,
				key.AltV,
				key.AltW,
				key.AltX,
				key.AltY,
				key.AltZ,
				key.AltZero,
				key.AltOne,
				key.AltTwo,
				key.AltThree,
				key.AltFour,
				key.AltFive,
				key.AltSix,
				key.AltSeven,
				key.AltEight,
				key.AltNine,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			const characters = "abcdefghijklmnopqrstuvwxyz0123456789"
			require.Len(t, tt.values, len(characters))
			for i, code := range characters {
				require.Equal(t, tea.KeyPressMsg{Code: code, Mod: tt.mod}.String(), tt.values[i])
			}
		})
	}
}

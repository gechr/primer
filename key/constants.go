package key

const (
	ModAlt   = "alt+"
	ModCtrl  = "ctrl+"
	ModShift = "shift+"
)

const (
	A = "a"
	B = "b"
	C = "c"
	D = "d"
	E = "e"
	F = "f"
	G = "g"
	H = "h"
	I = "i"
	J = "j"
	K = "k"
	L = "l"
	M = "m"
	N = "n"
	O = "o"
	P = "p"
	Q = "q"
	R = "r"
	S = "s"
	T = "t"
	U = "u"
	V = "v"
	W = "w"
	X = "x"
	Y = "y"
	Z = "z"
)

const (
	Zero  = "0"
	One   = "1"
	Two   = "2"
	Three = "3"
	Four  = "4"
	Five  = "5"
	Six   = "6"
	Seven = "7"
	Eight = "8"
	Nine  = "9"
)

const (
	CtrlA = ModCtrl + A
	CtrlB = ModCtrl + B
	CtrlC = ModCtrl + C
	CtrlD = ModCtrl + D
	CtrlE = ModCtrl + E
	CtrlF = ModCtrl + F
	CtrlG = ModCtrl + G
	CtrlH = ModCtrl + H
	CtrlI = ModCtrl + I
	CtrlJ = ModCtrl + J
	CtrlK = ModCtrl + K
	CtrlL = ModCtrl + L
	CtrlM = ModCtrl + M
	CtrlN = ModCtrl + N
	CtrlO = ModCtrl + O
	CtrlP = ModCtrl + P
	CtrlQ = ModCtrl + Q
	CtrlR = ModCtrl + R
	CtrlS = ModCtrl + S
	CtrlT = ModCtrl + T
	CtrlU = ModCtrl + U
	CtrlV = ModCtrl + V
	CtrlW = ModCtrl + W
	CtrlX = ModCtrl + X
	CtrlY = ModCtrl + Y
	CtrlZ = ModCtrl + Z
)

const (
	CtrlZero  = ModCtrl + Zero
	CtrlOne   = ModCtrl + One
	CtrlTwo   = ModCtrl + Two
	CtrlThree = ModCtrl + Three
	CtrlFour  = ModCtrl + Four
	CtrlFive  = ModCtrl + Five
	CtrlSix   = ModCtrl + Six
	CtrlSeven = ModCtrl + Seven
	CtrlEight = ModCtrl + Eight
	CtrlNine  = ModCtrl + Nine
)

const (
	AltA = ModAlt + A
	AltB = ModAlt + B
	AltC = ModAlt + C
	AltD = ModAlt + D
	AltE = ModAlt + E
	AltF = ModAlt + F
	AltG = ModAlt + G
	AltH = ModAlt + H
	AltI = ModAlt + I
	AltJ = ModAlt + J
	AltK = ModAlt + K
	AltL = ModAlt + L
	AltM = ModAlt + M
	AltN = ModAlt + N
	AltO = ModAlt + O
	AltP = ModAlt + P
	AltQ = ModAlt + Q
	AltR = ModAlt + R
	AltS = ModAlt + S
	AltT = ModAlt + T
	AltU = ModAlt + U
	AltV = ModAlt + V
	AltW = ModAlt + W
	AltX = ModAlt + X
	AltY = ModAlt + Y
	AltZ = ModAlt + Z
)

const (
	AltZero  = ModAlt + Zero
	AltOne   = ModAlt + One
	AltTwo   = ModAlt + Two
	AltThree = ModAlt + Three
	AltFour  = ModAlt + Four
	AltFive  = ModAlt + Five
	AltSix   = ModAlt + Six
	AltSeven = ModAlt + Seven
	AltEight = ModAlt + Eight
	AltNine  = ModAlt + Nine
)

const (
	Enter     = "enter"
	Tab       = "tab"
	Backspace = "backspace"
	Escape    = "esc"
	Space     = "space"
	Up        = "up"
	Down      = "down"
	Left      = "left"
	Right     = "right"
	Insert    = "insert"
	Delete    = "delete"
	PageUp    = "pgup"
	PageDown  = "pgdown"
	Home      = "home"
	End       = "end"
	Esc       = Escape
	Return    = Enter
)

// Keypad strings follow ultraviolet's keystroke names, which overlap with
// ordinary digits and navigation keys.
const (
	KeypadEnter    = Enter
	KeypadMultiply = "mul"
	KeypadPlus     = "plus"
	KeypadMinus    = "minus"
	KeypadDecimal  = "period"
	KeypadDivide   = "div"
	KeypadZero     = Zero
	KeypadOne      = One
	KeypadTwo      = Two
	KeypadThree    = Three
	KeypadFour     = Four
	KeypadFive     = Five
	KeypadSix      = Six
	KeypadSeven    = Seven
	KeypadEight    = Eight
	KeypadNine     = Nine
	KeypadUp       = Up
	KeypadDown     = Down
	KeypadLeft     = Left
	KeypadRight    = Right
	KeypadPageUp   = PageUp
	KeypadPageDown = PageDown
	KeypadHome     = Home
	KeypadEnd      = End
	KeypadInsert   = Insert
	KeypadDelete   = Delete
)

const (
	F1  = "f1"
	F2  = "f2"
	F3  = "f3"
	F4  = "f4"
	F5  = "f5"
	F6  = "f6"
	F7  = "f7"
	F8  = "f8"
	F9  = "f9"
	F10 = "f10"
	F11 = "f11"
	F12 = "f12"
	F13 = "f13"
	F14 = "f14"
	F15 = "f15"
	F16 = "f16"
	F17 = "f17"
	F18 = "f18"
	F19 = "f19"
	F20 = "f20"
	F21 = "f21"
	F22 = "f22"
	F23 = "f23"
	F24 = "f24"
)

const (
	CapsLock    = "capslock"
	ScrollLock  = "scrolllock"
	NumLock     = "numlock"
	PrintScreen = "printscreen"
	Pause       = "pause"
	Menu        = "menu"
)

const (
	MediaPlay        = "mediaplay"
	MediaPause       = "mediapause"
	MediaPlayPause   = "mediaplaypause"
	MediaReverse     = "mediareverse"
	MediaStop        = "mediastop"
	MediaFastForward = "mediafastforward"
	MediaRewind      = "mediarewind"
	MediaNext        = "medianext"
	MediaPrevious    = "mediaprev"
	MediaRecord      = "mediarecord"
)

const (
	VolumeDown = "lowervol"
	VolumeUp   = "raisevol"
	Mute       = "mute"
)

const (
	LeftShift  = "leftshift"
	LeftAlt    = "leftalt"
	LeftCtrl   = "leftctrl"
	LeftSuper  = "leftsuper"
	LeftMeta   = "leftmeta"
	RightShift = "rightshift"
	RightAlt   = "rightalt"
	RightCtrl  = "rightctrl"
	RightSuper = "rightsuper"
	RightMeta  = "rightmeta"
)

const (
	AltEnter  = ModAlt + Enter
	ShiftDown = ModShift + Down
	ShiftUp   = ModShift + Up
	ShiftTab  = ModShift + Tab
)

const (
	ArrowsLeftRight     = "←→"
	ArrowsVertical      = "↕"
	ShiftArrowsVertical = ModShift + ArrowsVertical
)

package key

const (
	ModAlt   = "alt+"
	ModCtrl  = "ctrl+"
	ModShift = "shift+"
)

const (
	AltEnter = ModAlt + Enter

	CtrlA = ModCtrl + "a"
	CtrlB = ModCtrl + "b"
	CtrlC = ModCtrl + "c"
	CtrlD = ModCtrl + "d"
	CtrlE = ModCtrl + "e"
	CtrlF = ModCtrl + "f"
	CtrlN = ModCtrl + "n"
	CtrlP = ModCtrl + "p"
	CtrlS = ModCtrl + "s"
)

const (
	N     = "n"
	Down  = "down"
	Enter = "enter"
	Esc   = "esc"
	Left  = "left"
	Right = "right"
	Space = "space"
	Tab   = "tab"
	Up    = "up"
	Y     = "y"
)

const (
	ShiftDown = ModShift + Down
	ShiftUp   = ModShift + Up
	ShiftTab  = ModShift + Tab
)

const (
	ArrowsLeftRight   = "←→"
	ArrowsUpDown      = "↑↓"
	ShiftArrowsUpDown = ModShift + ArrowsUpDown
)

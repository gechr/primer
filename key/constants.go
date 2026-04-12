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
	CtrlF = ModCtrl + "f"
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

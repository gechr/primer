package key

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// namedKeys maps each non-text key's canonical name back to its code, so
// Replay can rebuild a KeyPressMsg whose String round-trips to the binding it
// came from. The names are sourced from bubbletea itself - every code is run
// through KeyPressMsg.String() - so this table can never disagree with the
// library's own spelling, and a version bump that renames a key updates it for
// free. This matters because rebinding is free-form: a user can bind an action
// to "ctrl+f5", and a replay must reproduce that literally, not a lossy
// approximation.
var namedKeys = buildNamedKeys()

// buildNamedKeys derives the name→code table by round-tripping every named key
// bubbletea exports (control, editing, navigation, and the function keys) back
// through String(). The function keys are consecutive codes, so they generate
// in a loop. Keypad, media, lock, and modifier keys are omitted deliberately:
// nobody binds an action to "kpenter" or "mute", and Replay's fallback turns
// any such rebind into a safe no-op rather than a wrong match.
func buildNamedKeys() map[string]rune {
	codes := []rune{
		tea.KeyTab, tea.KeyEnter, tea.KeyEscape, tea.KeySpace, tea.KeyBackspace,
		tea.KeyDelete, tea.KeyInsert,
		tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight,
		tea.KeyHome, tea.KeyEnd, tea.KeyPgUp, tea.KeyPgDown,
		tea.KeyBegin, tea.KeyFind, tea.KeySelect,
	}
	// KeyF1..KeyF63 are consecutive, so add them without listing all 63.
	for f := rune(0); tea.KeyF1+f <= tea.KeyF63; f++ {
		codes = append(codes, tea.KeyF1+f)
	}
	m := make(map[string]rune, len(codes))
	for _, c := range codes {
		m[tea.KeyPressMsg{Code: c}.String()] = c
	}
	return m
}

// Replay rebuilds the KeyPressMsg for a binding's key string ("t", "tab",
// "shift+tab", "ctrl+u", "ctrl+f5"), so replaying a command-palette selection
// is byte-for-byte the same event pressing the key would raise - key.Matches
// compares String(), and this reconstructs exactly that. Modifier prefixes
// strip left to right; the remainder is a named key, a lone rune, or - for a
// multi-rune name we don't know - an extended key.
//
// Text is set only when unmodified. A KeyPressMsg's String() returns its Text
// verbatim when present, ignoring the modifier, so setting Text on a modified
// key would collapse "ctrl+f99" to a bare "f99" and could fire a command bound
// to that literal. Leaving Text empty lets String() fall through to the
// modifier-prefixed keystroke form, so an unsupported modified rebind renders
// as a stub that matches nothing rather than misfiring.
func Replay(s string) tea.KeyPressMsg {
	var mod tea.KeyMod
	for {
		switch {
		case strings.HasPrefix(s, ModCtrl):
			mod |= tea.ModCtrl
			s = s[len(ModCtrl):]
		case strings.HasPrefix(s, ModAlt):
			mod |= tea.ModAlt
			s = s[len(ModAlt):]
		case strings.HasPrefix(s, ModShift):
			mod |= tea.ModShift
			s = s[len(ModShift):]
		default:
			if code, ok := namedKeys[s]; ok {
				return tea.KeyPressMsg{Mod: mod, Code: code}
			}
			code := rune(tea.KeyExtended)
			if r := []rune(s); len(r) == 1 {
				code = r[0]
			}
			km := tea.KeyPressMsg{Mod: mod, Code: code}
			if mod == 0 {
				km.Text = s
			}
			return km
		}
	}
}

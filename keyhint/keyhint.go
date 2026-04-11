package keyhint

import (
	"strings"

	lg "charm.land/lipgloss/v2"
)

const (
	nl             = "\n"
	asciiLowerMask = 0x20
)

type Hint struct {
	Key  string
	Desc string
}

type Styles struct {
	Key  lg.Style
	Text lg.Style
}

type Renderer struct {
	Styles Styles
	Gap    string
	Width  int
	Inline bool
}

// Inline highlights the key inside the description when it is a single-letter suffix.
func Inline(key, desc string, keyStyle, textStyle lg.Style) (string, bool) {
	keyPrefix, keyLetter, ok := splitInlineKey(key)
	if !ok {
		return "", false
	}
	idx := strings.Index(strings.ToLower(desc), strings.ToLower(keyLetter))
	if idx < 0 {
		return "", false
	}
	before := desc[:idx]
	after := desc[idx+1:]
	var part string
	if keyPrefix != "" {
		part = keyStyle.Render(keyPrefix)
	}
	if before != "" {
		part += textStyle.Render(before)
	}
	part += keyStyle.Render(keyLetter)
	if after != "" {
		part += textStyle.Render(after)
	}
	return part, true
}

func (r Renderer) Render(hints []Hint) string {
	parts := make([]string, 0, len(hints))
	gap := r.Gap
	if gap == "" {
		gap = "   "
	}

	for _, hint := range hints {
		if r.Inline {
			if inlined, ok := Inline(hint.Key, hint.Desc, r.Styles.Key, r.Styles.Text); ok {
				parts = append(parts, inlined)
				continue
			}
		}
		parts = append(
			parts,
			r.Styles.Key.Render(hint.Key)+" "+renderDesc(hint.Desc, r.Styles.Text),
		)
	}

	if r.Width <= 0 {
		return " " + strings.Join(parts, gap)
	}

	const indent = " "
	var lines []string
	var line string
	lineWidth := len(indent)
	gapWidth := lg.Width(gap)
	for i, part := range parts {
		partWidth := lg.Width(part)
		switch {
		case i == 0:
			line = indent + part
			lineWidth = len(indent) + partWidth
		case lineWidth+gapWidth+partWidth > r.Width:
			lines = append(lines, line)
			line = part
			lineWidth = partWidth
		default:
			line += gap + part
			lineWidth += gapWidth + partWidth
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, nl)
}

func renderDesc(desc string, textStyle lg.Style) string {
	if strings.Contains(desc, "\x1b[") {
		if idx := strings.Index(desc, "\x1b["); idx > 0 {
			return textStyle.Render(desc[:idx]) + desc[idx:]
		}
		return desc
	}
	return textStyle.Render(desc)
}

func splitInlineKey(key string) (string, string, bool) {
	if len(key) == 1 {
		ch := key[0] | asciiLowerMask
		if ch < 'a' || ch > 'z' {
			return "", "", false
		}
		return "", key, true
	}

	idx := strings.LastIndex(key, "+")
	if idx <= 0 || idx == len(key)-1 {
		return "", "", false
	}
	letter := key[idx+1:]
	if len(letter) != 1 {
		return "", "", false
	}
	ch := letter[0] | asciiLowerMask
	if ch < 'a' || ch > 'z' {
		return "", "", false
	}
	return key[:idx+1], letter, true
}

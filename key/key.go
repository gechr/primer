package key

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
	Prefix *string
	Width  int
	Inline bool
}

// hintSpan is where one rendered hint sits: columns [start, end) of line.
type hintSpan struct {
	hint  int
	line  int
	start int
	end   int
}

func (r Renderer) Render(hints []Hint) string {
	lines, _ := r.layout(hints)
	return strings.Join(lines, nl)
}

// HintAt returns the index of the hint drawn at column x of line y in Render's
// output. The whole "key desc" part is the target, so a click anywhere on it
// counts; the gap between parts belongs to no hint.
func (r Renderer) HintAt(hints []Hint, x, y int) (int, bool) {
	_, spans := r.layout(hints)
	for _, s := range spans {
		if s.line == y && x >= s.start && x < s.end {
			return s.hint, true
		}
	}
	return 0, false
}

// layout renders the hint parts into lines and records where each part lands,
// so Render and HintAt always agree.
func (r Renderer) layout(hints []Hint) ([]string, []hintSpan) {
	parts := make([]string, 0, len(hints))
	gap := r.Gap
	if gap == "" {
		gap = "  "
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

	prefix := " "
	if r.Prefix != nil {
		prefix = *r.Prefix
	}

	gapWidth := lg.Width(gap)
	spans := make([]hintSpan, 0, len(parts))

	if r.Width <= 0 {
		x := lg.Width(prefix)
		for i, part := range parts {
			if i > 0 {
				x += gapWidth
			}
			spans = append(spans, hintSpan{hint: i, start: x, end: x + lg.Width(part)})
			x = spans[i].end
		}
		return []string{prefix + strings.Join(parts, gap)}, spans
	}

	indent := prefix
	var lines []string
	var line string
	lineWidth := len(indent)
	x := 0 // display column where the next part's gap begins
	for i, part := range parts {
		partWidth := lg.Width(part)
		var start int
		switch {
		case i == 0:
			line = indent + part
			lineWidth = len(indent) + partWidth
			start = lg.Width(indent)
		case lineWidth+gapWidth+partWidth > r.Width:
			lines = append(lines, line)
			line = part
			lineWidth = partWidth
		default:
			line += gap + part
			lineWidth += gapWidth + partWidth
			start = x + gapWidth
		}
		x = start + partWidth
		spans = append(spans, hintSpan{hint: i, line: len(lines), start: start, end: x})
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines, spans
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

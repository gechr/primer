package filter

import "strings"

// Case controls case sensitivity for matching.
type Case uint8

const (
	// SmartCase is case-insensitive unless the term contains uppercase.
	SmartCase Case = iota
	// CaseSensitive always matches case exactly.
	CaseSensitive
	// CaseInsensitive always ignores case.
	CaseInsensitive
)

// Term represents a parsed search term with optional modifiers.
type Term struct {
	Text   string
	Prefix bool // ^ anchor
	Suffix bool // $ anchor
	Negate bool // ! invert
	Case   Case
}

// Parse parses a filter string into a Term.
//
// Supported modifiers:
//   - ! at the start negates the match
//   - ^ anchors to the start of the text
//   - $ anchors to the end of the text
//
// Case sensitivity uses smart case: case-insensitive unless the term
// contains an uppercase letter.
func Parse(f string) Term {
	var t Term

	if rest, ok := strings.CutPrefix(f, "!"); ok {
		t.Negate = true
		f = rest
	}
	if rest, ok := strings.CutPrefix(f, "^"); ok {
		t.Prefix = true
		f = rest
	}
	if rest, ok := strings.CutSuffix(f, "$"); ok {
		t.Suffix = true
		f = rest
	}

	if f != strings.ToLower(f) {
		t.Case = CaseSensitive
	}
	t.Text = f
	return t
}

// Match reports whether text matches the term.
func (t Term) Match(text string) bool {
	if t.Text == "" {
		return true
	}

	needle := t.Text
	foldCase := t.Case == CaseInsensitive ||
		(t.Case == SmartCase && needle == strings.ToLower(needle))
	if foldCase {
		text = strings.ToLower(text)
		needle = strings.ToLower(needle)
	}

	var matched bool
	switch {
	case t.Prefix && t.Suffix:
		matched = text == needle
	case t.Prefix:
		matched = strings.HasPrefix(text, needle)
	case t.Suffix:
		matched = strings.HasSuffix(text, needle)
	default:
		matched = strings.Contains(text, needle)
	}

	if t.Negate {
		return !matched
	}
	return matched
}

package filter

import (
	"strings"
	"unicode/utf8"
)

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

// Fuzzy reports whether every rune in query appears in text in order
// (but not necessarily contiguously). It returns the matched rune indexes,
// or nil if the query does not match. Case sensitivity is controlled by c
// using the same semantics as [Term.Match]: [SmartCase] is case-insensitive
// unless query contains uppercase.
//
// When multiple match positions exist, Fuzzy prefers the tightest
// (shortest-span) match using a three-pass algorithm:
//  1. Forward pass to confirm a match exists.
//  2. Backward pass to find the tightest window.
//  3. Forward pass over that window to collect exact indexes.
func Fuzzy(text, query string, c Case) []int {
	foldCase := c == CaseInsensitive ||
		(c == SmartCase && query == strings.ToLower(query))

	normalizedText := text
	normalizedQuery := query
	if foldCase {
		normalizedText = strings.ToLower(text)
		normalizedQuery = strings.ToLower(query)
	}

	textRunes := []rune(normalizedText)
	queryRunes := []rune(normalizedQuery)
	n := len(textRunes)
	qn := len(queryRunes)

	if qn == 0 {
		return []int{}
	}

	// Forward pass: confirm all query runes exist in order.
	qi := 0
	for i := 0; i < n && qi < qn; i++ {
		if textRunes[i] == queryRunes[qi] {
			qi++
		}
	}
	if qi < qn {
		return nil
	}

	// Backward pass: find the tightest window by matching in reverse.
	qi = qn - 1
	endIdx := 0
	startIdx := 0
	for i := n - 1; i >= 0 && qi >= 0; i-- {
		if textRunes[i] == queryRunes[qi] {
			if qi == qn-1 {
				endIdx = i
			}
			if qi == 0 {
				startIdx = i
			}
			qi--
		}
	}

	// Forward pass over the window to collect matched indexes.
	matched := make([]int, 0, qn)
	qi = 0
	for i := startIdx; i <= endIdx && qi < qn; i++ {
		if textRunes[i] == queryRunes[qi] {
			matched = append(matched, i)
			qi++
		}
	}
	return matched
}

// FuzzyBytes converts matched rune indexes (from [Fuzzy]) into
// byte offset pairs within text. Each pair [start, end) covers one matched
// rune. Returns nil when indexes is empty.
func FuzzyBytes(text string, indexes []int) [][2]int {
	if len(indexes) == 0 {
		return nil
	}

	needed := make(map[int]struct{}, len(indexes))
	for _, ri := range indexes {
		needed[ri] = struct{}{}
	}

	type pos struct{ off, size int }
	found := make(map[int]pos, len(needed))

	runeIdx := 0
	byteIdx := 0
	for byteIdx < len(text) && len(found) < len(needed) {
		_, size := utf8.DecodeRuneInString(text[byteIdx:])
		if _, ok := needed[runeIdx]; ok {
			found[runeIdx] = pos{byteIdx, size}
		}
		byteIdx += size
		runeIdx++
	}

	ranges := make([][2]int, len(indexes))
	for i, ri := range indexes {
		p := found[ri]
		ranges[i] = [2]int{p.off, p.off + p.size}
	}
	return ranges
}

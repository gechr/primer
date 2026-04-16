package filter_test

import (
	"testing"

	"github.com/gechr/primer/filter"
	"github.com/stretchr/testify/require"
)

func TestParseSimple(t *testing.T) {
	term := filter.Parse("hello")
	require.Equal(t, "hello", term.Text)
	require.False(t, term.Prefix)
	require.False(t, term.Suffix)
	require.False(t, term.Negate)
	require.Equal(t, filter.SmartCase, term.Case)
}

func TestParseNegate(t *testing.T) {
	term := filter.Parse("!draft")
	require.Equal(t, "draft", term.Text)
	require.True(t, term.Negate)
}

func TestParsePrefix(t *testing.T) {
	term := filter.Parse("^fix")
	require.Equal(t, "fix", term.Text)
	require.True(t, term.Prefix)
}

func TestParseSuffix(t *testing.T) {
	term := filter.Parse("bug$")
	require.Equal(t, "bug", term.Text)
	require.True(t, term.Suffix)
}

func TestParseCombined(t *testing.T) {
	term := filter.Parse("!^fix$")
	require.Equal(t, "fix", term.Text)
	require.True(t, term.Negate)
	require.True(t, term.Prefix)
	require.True(t, term.Suffix)
}

func TestParseSmartCaseLowercase(t *testing.T) {
	term := filter.Parse("hello")
	require.Equal(t, filter.SmartCase, term.Case)
}

func TestParseSmartCaseUppercase(t *testing.T) {
	term := filter.Parse("Hello")
	require.Equal(t, filter.CaseSensitive, term.Case)
}

func TestMatchExplicitCaseInsensitive(t *testing.T) {
	term := filter.Term{Text: "Hello", Case: filter.CaseInsensitive}
	require.True(t, term.Match("hello world"))
	require.True(t, term.Match("HELLO WORLD"))
}

func TestMatchContains(t *testing.T) {
	term := filter.Parse("fix")
	require.True(t, term.Match("bugfix: resolve crash"))
	require.False(t, term.Match("add new feature"))
}

func TestMatchCaseInsensitive(t *testing.T) {
	term := filter.Parse("fix")
	require.True(t, term.Match("Fix typo"))
	require.True(t, term.Match("BUGFIX"))
}

func TestMatchCaseSensitive(t *testing.T) {
	term := filter.Parse("Fix")
	require.True(t, term.Match("Fix typo"))
	require.False(t, term.Match("fix typo"))
}

func TestMatchPrefix(t *testing.T) {
	term := filter.Parse("^fix")
	require.True(t, term.Match("fix: resolve crash"))
	require.False(t, term.Match("bugfix: resolve crash"))
}

func TestMatchSuffix(t *testing.T) {
	term := filter.Parse("crash$")
	require.True(t, term.Match("fix: resolve crash"))
	require.False(t, term.Match("crash: resolved"))
}

func TestMatchExact(t *testing.T) {
	term := filter.Parse("^fix$")
	require.True(t, term.Match("fix"))
	require.False(t, term.Match("fix: typo"))
	require.False(t, term.Match("bugfix"))
}

func TestMatchNegate(t *testing.T) {
	term := filter.Parse("!draft")
	require.True(t, term.Match("fix: resolve crash"))
	require.False(t, term.Match("draft: new feature"))
}

func TestMatchEmpty(t *testing.T) {
	term := filter.Parse("")
	require.True(t, term.Match("anything"))
	require.True(t, term.Match(""))
}

func TestMatchNegatePrefix(t *testing.T) {
	term := filter.Parse("!^fix")
	require.True(t, term.Match("bugfix: resolve crash"))
	require.False(t, term.Match("fix: resolve crash"))
}

func TestFuzzyBasic(t *testing.T) {
	t.Parallel()
	require.Equal(t, []int{0}, filter.Fuzzy("abc", "a", filter.SmartCase))
	require.Equal(t, []int{0, 1}, filter.Fuzzy("abc", "ab", filter.SmartCase))
	require.Equal(t, []int{0, 2}, filter.Fuzzy("abc", "ac", filter.SmartCase))
	require.Equal(t, []int{0, 1, 2}, filter.Fuzzy("abc", "abc", filter.SmartCase))
	require.Equal(t, []int{1}, filter.Fuzzy("abc", "b", filter.SmartCase))
	require.Equal(t, []int{1, 2}, filter.Fuzzy("abc", "bc", filter.SmartCase))
	require.Equal(t, []int{2}, filter.Fuzzy("abc", "c", filter.SmartCase))
}

func TestFuzzyNoMatch(t *testing.T) {
	t.Parallel()
	require.Nil(t, filter.Fuzzy("abc", "cba", filter.SmartCase))
	require.Nil(t, filter.Fuzzy("abc", "d", filter.SmartCase))
	require.Nil(t, filter.Fuzzy("abc", "abcd", filter.SmartCase))
}

func TestFuzzyWithGaps(t *testing.T) {
	t.Parallel()
	require.Equal(t, []int{1, 3, 5}, filter.Fuzzy("xaxbxc", "abc", filter.SmartCase))
	require.Equal(t, []int{1, 5}, filter.Fuzzy("xaxbxc", "ac", filter.SmartCase))
}

func TestFuzzyEmptyQuery(t *testing.T) {
	t.Parallel()
	require.Equal(t, []int{}, filter.Fuzzy("anything", "", filter.SmartCase))
}

func TestFuzzySmartCaseInsensitive(t *testing.T) {
	t.Parallel()
	// Lowercase query → case-insensitive.
	require.Equal(t, []int{0, 1, 2, 3, 4}, filter.Fuzzy("Hello", "hello", filter.SmartCase))
}

func TestFuzzySmartCaseSensitive(t *testing.T) {
	t.Parallel()
	// Uppercase in query → case-sensitive.
	require.Equal(t, []int{0, 1, 2}, filter.Fuzzy("Abc", "Abc", filter.SmartCase))
	require.Nil(t, filter.Fuzzy("abc", "Abc", filter.SmartCase))
}

func TestFuzzyCaseSensitive(t *testing.T) {
	t.Parallel()
	require.Equal(t, []int{0, 1, 2}, filter.Fuzzy("abc", "abc", filter.CaseSensitive))
	require.Nil(t, filter.Fuzzy("abc", "ABC", filter.CaseSensitive))
	require.Nil(t, filter.Fuzzy("ABC", "abc", filter.CaseSensitive))
}

func TestFuzzyCaseInsensitive(t *testing.T) {
	t.Parallel()
	require.Equal(t, []int{0, 1, 2}, filter.Fuzzy("ABC", "abc", filter.CaseInsensitive))
	require.Equal(t, []int{0, 1, 2}, filter.Fuzzy("abc", "ABC", filter.CaseInsensitive))
}

func TestFuzzyTightestSpan(t *testing.T) {
	t.Parallel()
	// Prefers the later contiguous "abc" over the early spread "a...b...c".
	require.Equal(t, []int{7, 8, 9}, filter.Fuzzy("xaxbxcxabc", "abc", filter.SmartCase))
	// Prefers "a_b" at the end over "a____b" at the start.
	require.Equal(t, []int{10, 12}, filter.Fuzzy("a____b____a_b", "ab", filter.SmartCase))
	// Full substring tightest match.
	require.Equal(
		t,
		[]int{7, 8, 9, 10, 11, 12, 13},
		filter.Fuzzy("foobar-bar-baz", "bar-baz", filter.SmartCase),
	)
}

func TestFuzzyUnicode(t *testing.T) {
	t.Parallel()
	require.Equal(t, []int{0}, filter.Fuzzy("こんにちは", "こ", filter.SmartCase))
	require.Equal(t, []int{0, 2, 4}, filter.Fuzzy("こんにちは", "こには", filter.SmartCase))
}

func TestFuzzyBytesASCII(t *testing.T) {
	t.Parallel()
	ranges := filter.FuzzyBytes("hello", []int{0, 2, 4})
	require.Equal(t, [][2]int{{0, 1}, {2, 3}, {4, 5}}, ranges)
}

func TestFuzzyBytesUnicode(t *testing.T) {
	t.Parallel()
	// "über" - ü is 2 bytes in UTF-8.
	ranges := filter.FuzzyBytes("über", []int{0, 2})
	require.Equal(t, [][2]int{{0, 2}, {3, 4}}, ranges)
}

func TestFuzzyBytesEmpty(t *testing.T) {
	t.Parallel()
	require.Nil(t, filter.FuzzyBytes("hello", nil))
	require.Nil(t, filter.FuzzyBytes("hello", []int{}))
}

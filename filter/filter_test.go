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

package render_test

import (
	"image/color"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	gansi "charm.land/glamour/v2/ansi"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/gechr/primer/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownRendererInvalidatesChangedContent(t *testing.T) {
	t.Parallel()

	r := render.NewMarkdownRenderer(testMarkdownStyle())

	first := r.Render("issue-1", 80, "# First")
	second := r.Render("issue-1", 80, "# Second")
	third := r.Render("issue-1", 80, "# First")

	require.NotEmpty(t, first)
	require.NotEmpty(t, second)
	assert.NotEqual(t, first, second)
	assert.Equal(t, first, third)
}

func TestMarkdownRendererClampsWidth(t *testing.T) {
	t.Parallel()

	r := render.NewMarkdownRenderer(testMarkdownStyle())

	got := r.Render("zero", 0, "hello")
	want := r.Render("one", 1, "hello")

	require.Equal(t, want, got)
}

func TestMarkdownRendererResetsOutputCacheOnOverflow(t *testing.T) {
	t.Parallel()

	r := render.NewMarkdownRenderer(testMarkdownStyle(), render.WithMaxCacheEntries(1))

	first := r.Render("a", 80, "# One")
	_ = r.Render("b", 80, "# Two")
	again := r.Render("a", 80, "# One")

	require.Equal(t, first, again)
	assert.LessOrEqual(t, cacheLen(t, r, "outputs"), 1)
}

func TestMarkdownRendererResetsRendererCacheOnOverflow(t *testing.T) {
	t.Parallel()

	r := render.NewMarkdownRenderer(testMarkdownStyle(), render.WithMaxCacheEntries(1))

	first := r.Render("a", 20, "a paragraph that wraps")
	_ = r.Render("b", 21, "a paragraph that wraps")
	again := r.Render("c", 20, "a paragraph that wraps")

	require.Equal(t, first, again)
	assert.LessOrEqual(t, cacheLen(t, r, "renderers"), 1)
}

func TestMarkdownRendererTrimPadding(t *testing.T) {
	t.Parallel()

	style := testMarkdownStyle()

	trimmed := render.NewMarkdownRenderer(style).Render("trim", 80, "# Title")
	padded := render.NewMarkdownRenderer(style, render.WithTrimPadding(false)).
		Render("pad", 80, "# Title")

	require.False(t, strings.HasPrefix(trimmed, "\n"))
	require.False(t, strings.HasSuffix(trimmed, "\n"))
	require.True(t, strings.HasPrefix(padded, "\n"))
	require.True(t, strings.HasSuffix(padded, "\n"))
}

func TestMarkdownRendererConcurrentUse(t *testing.T) {
	t.Parallel()

	r := render.NewMarkdownRenderer(testMarkdownStyle(), render.WithMaxCacheEntries(8))

	var empty atomic.Bool
	var wg sync.WaitGroup
	for worker := range 16 {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := range 100 {
				id := string(rune('a' + worker%4))
				md := "# Title\n\nworker text"
				if r.Render(id, 20+i%6, md) == "" {
					empty.Store(true)
				}
			}
		}(worker)
	}
	wg.Wait()
	require.False(t, empty.Load())
}

func TestColorToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   color.Color
		want string
	}{
		{name: "basic ansi", in: xansi.BasicColor(4), want: "4"},
		{name: "indexed ansi", in: xansi.IndexedColor(212), want: "212"},
		{name: "true color", in: color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}, want: "#123456"},
		{name: "nil", in: nil, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, render.ColorToken(tt.in))
		})
	}
}

func TestStyleFromPaletteMapping(t *testing.T) {
	t.Parallel()

	style := render.StyleFromPalette(render.MarkdownPalette{
		Base:    render.BackgroundLight,
		Text:    xansi.BasicColor(7),
		Heading: xansi.IndexedColor(39),
		H1:      color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
		Link:    xansi.IndexedColor(212),
		Code:    xansi.BasicColor(3),
		Dim:     color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff},
	})

	assertStringPtr(t, style.Document.Color, "7")
	assertUintPtr(t, style.Document.Margin, 0)
	assertStringPtr(t, style.H1.Color, "#aabbcc")
	assert.Nil(t, style.H1.BackgroundColor)

	for name, got := range map[string]*string{
		"heading": style.Heading.Color,
		"h2":      style.H2.Color,
		"h3":      style.H3.Color,
		"h4":      style.H4.Color,
		"h5":      style.H5.Color,
		"h6":      style.H6.Color,
	} {
		assertStringPtr(t, got, "39", name)
	}

	for name, got := range map[string]*string{
		"strong":      style.Strong.Color,
		"emph":        style.Emph.Color,
		"item":        style.Item.Color,
		"enumeration": style.Enumeration.Color,
	} {
		assertStringPtr(t, got, "7", name)
	}

	assertStringPtr(t, style.Link.Color, "212")
	assertStringPtr(t, style.LinkText.Color, "212")
	assertStringPtr(t, style.Code.Color, "3")
	assert.Nil(t, style.Code.BackgroundColor)
	assertStringPtr(t, style.BlockQuote.Color, "#112233")
	assertStringPtr(t, style.HorizontalRule.Color, "#112233")
	assertStringPtr(t, style.CodeBlock.Chroma.Text.Color, "#2A2A2A")
}

func cacheLen(t *testing.T, r *render.MarkdownRenderer, name string) int {
	t.Helper()

	field := reflect.ValueOf(r).Elem().FieldByName(name)
	require.True(t, field.IsValid())

	return field.Len()
}

func assertStringPtr(t *testing.T, got *string, want string, msgAndArgs ...any) {
	t.Helper()

	require.NotNil(t, got, msgAndArgs...)
	assert.Equal(t, want, *got, msgAndArgs...)
}

func assertUintPtr(t *testing.T, got *uint, want uint) {
	t.Helper()

	require.NotNil(t, got)
	assert.Equal(t, want, *got)
}

func testMarkdownStyle() gansi.StyleConfig {
	return render.StyleFromPalette(render.MarkdownPalette{
		Base:    render.BackgroundDark,
		Text:    xansi.BasicColor(7),
		Heading: xansi.BasicColor(6),
		H1:      xansi.BasicColor(5),
		Link:    xansi.BasicColor(4),
		Code:    xansi.BasicColor(3),
		Dim:     xansi.BasicColor(8),
	})
}

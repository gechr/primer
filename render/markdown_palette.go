package render

import (
	"fmt"
	"image/color"

	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	xansi "github.com/charmbracelet/x/ansi"
)

// Background selects the light or dark glamour base used by [StyleFromPalette].
type Background string

const (
	// BackgroundDark uses glamour's dark defaults as the markdown base.
	BackgroundDark Background = "dark"
	// BackgroundLight uses glamour's light defaults as the markdown base.
	BackgroundLight Background = "light"

	rgbaTokenShift = 8
)

// MarkdownPalette carries the small set of app-theme colors needed to derive a
// glamour style while leaving code-block chroma and unlisted elements on the
// selected light or dark base.
type MarkdownPalette struct {
	Base    Background
	Text    color.Color
	Heading color.Color
	H1      color.Color
	Link    color.Color
	Code    color.Color
	Dim     color.Color
}

// StyleFromPalette derives a glamour style from p.
//
// The derived style clears document margins because panes own their padding,
// removes filled backgrounds from H1 and inline code so theme color does not
// create UI blocks, and pins list/bold/emphasis colors to Text so glamour
// defaults do not leak into themed prose.
func StyleFromPalette(p MarkdownPalette) ansi.StyleConfig {
	style := styles.DarkStyleConfig
	if p.Base == BackgroundLight {
		style = styles.LightStyleConfig
	}

	style.Document.Margin = new(uint(0))
	style.Document.Color = colorTokenPtr(p.Text)
	style.H1.Color = colorTokenPtr(p.H1)
	style.H1.BackgroundColor = nil

	heading := colorTokenPtr(p.Heading)
	style.Heading.Color = heading
	style.H2.Color = heading
	style.H3.Color = heading
	style.H4.Color = heading
	style.H5.Color = heading
	style.H6.Color = heading

	text := colorTokenPtr(p.Text)
	style.Strong.Color = text
	style.Emph.Color = text
	style.Item.Color = text
	style.Enumeration.Color = text

	link := colorTokenPtr(p.Link)
	style.Link.Color = link
	style.LinkText.Color = link

	style.Code.Color = colorTokenPtr(p.Code)
	style.Code.BackgroundColor = nil

	dim := colorTokenPtr(p.Dim)
	style.BlockQuote.Color = dim
	style.HorizontalRule.Color = dim

	return style
}

// ColorToken converts a color into the token format glamour style configs use.
//
// ANSI palette indexes pass through as their index string, such as "4" or
// "212". Round-tripping an indexed color through RGBA bakes in the standard VGA
// value and visibly drifts from themed UI rendered around it.
func ColorToken(c color.Color) string {
	switch v := c.(type) {
	case xansi.BasicColor:
		return fmt.Sprintf("%d", int(v))
	case xansi.IndexedColor:
		return fmt.Sprintf("%d", int(v))
	}
	if c == nil {
		return ""
	}

	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>rgbaTokenShift, g>>rgbaTokenShift, b>>rgbaTokenShift)
}

func colorTokenPtr(c color.Color) *string {
	token := ColorToken(c)
	if token == "" {
		return nil
	}
	return new(token)
}

package render_test

import (
	"fmt"
	"image/color"

	"github.com/gechr/primer/render"
)

func ExampleMarkdownRenderer() {
	style := render.StyleFromPalette(render.MarkdownPalette{
		Base:    render.BackgroundDark,
		Text:    color.RGBA{R: 0xee, G: 0xee, B: 0xee, A: 0xff},
		Heading: color.RGBA{R: 0x7a, G: 0xdf, B: 0xd6, A: 0xff},
		H1:      color.RGBA{R: 0xff, G: 0xb8, B: 0x6c, A: 0xff},
		Link:    color.RGBA{R: 0x8a, G: 0xb4, B: 0xf8, A: 0xff},
		Code:    color.RGBA{R: 0xff, G: 0x79, B: 0xc6, A: 0xff},
		Dim:     color.RGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xff},
	})
	renderer := render.NewMarkdownRenderer(style)

	out := renderer.Render("issue-123", 72, "# Summary\n\nRepeated dashboard content.")
	fmt.Println(out != "")
	// Output: true
}

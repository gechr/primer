package dialog

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	lg "charm.land/lipgloss/v2"
	"github.com/gechr/primer/key"
	"github.com/gechr/primer/overlay"
	"github.com/gechr/primer/prompt"
	"github.com/gechr/primer/scrollbar"
)

// Styles are the Shell's render styles, injected by the owner so the package
// stays theme-agnostic.
type Styles struct {
	// Box draws the border and padding around the whole dialog; inject the
	// app's overlay style.
	Box lg.Style
	// Title styles the heading line above the body.
	Title lg.Style
	// HintKey and HintText style the key and description halves of the foot
	// row, matching the rest of the TUI's hint rendering.
	HintKey  lg.Style
	HintText lg.Style
	// Scrollbar styles the internal scrollbar shown when the body overflows
	// the height cap.
	Scrollbar scrollbar.Styles
}

// ShellConfig declares a Shell's chrome and sizing bounds.
type ShellConfig struct {
	Styles Styles
	// MaxWidth and MaxHeight cap the box's total size in columns and rows; 0
	// disables the cap. They keep a long prefill or a tall body from bleeding
	// past a large screen.
	MaxWidth  int
	MaxHeight int
	// WidthFraction and HeightFraction cap the box to a share of the screen in
	// (0,1]; 0 disables. They bind on small screens where the absolute caps do
	// not.
	WidthFraction  float64
	HeightFraction float64
	// Margin keeps this many columns and rows clear at the screen edges when
	// the screen, not a cap, is the binding constraint.
	Margin int
	// Scrollbar configures the internal scrollbar's glyphs and geometry.
	Scrollbar scrollbar.Config
}

// scrollChrome is the columns a scrolling body spends beside itself: the
// scrollbar plus the padding column RenderScrollable adds between the bar and
// the border.
const scrollChrome = 2

// Shell frames a dialog: it caps the box to a fraction of the screen and to an
// absolute maximum, scrolls an overflowing body internally, and centers the
// result over a backdrop. It carries value semantics and is safe to copy - the
// Stack holds one by value.
type Shell struct {
	cfg      ShellConfig
	viewport viewport.Model
}

// NewShell builds a Shell from cfg. It constructs the viewport once so
// RenderScrollable reuses its default key mappings and configuration.
func NewShell(cfg ShellConfig) Shell {
	return Shell{cfg: cfg, viewport: viewport.New()}
}

// clampWidth is the maximum total box width for a screen.
func (s Shell) clampWidth(screenW int) int {
	return clampAxis(screenW, s.cfg.Margin, s.cfg.WidthFraction, s.cfg.MaxWidth)
}

// clampHeight is the maximum total box height for a screen; it is what lets
// an oversized body scroll inside the box instead of overrunning the screen.
func (s Shell) clampHeight(screenH int) int {
	return clampAxis(screenH, s.cfg.Margin, s.cfg.HeightFraction, s.cfg.MaxHeight)
}

// clampAxis is the maximum total box size along one axis: the smaller of the
// absolute cap, the fractional cap, and the screen minus the edge margin,
// floored at one so a tiny terminal can never push a negative size into
// lipgloss.
func clampAxis(screen, margin int, fraction float64, limit int) int {
	v := screen - margin
	if fraction > 0 {
		if f := int(float64(screen) * fraction); f < v {
			v = f
		}
	}
	if limit > 0 && limit < v {
		v = limit
	}
	return max(1, v)
}

// frameGeometry records where the Shell placed a dialog's content during the
// last render, so the Stack can translate screen-space mouse coordinates into
// the content space the dialog laid out. contentX/contentY are the screen
// cell of the content's top-left; contentW/contentH its visible size; titleH
// the Shell-drawn title rows above the dialog's body; scroll the viewport row
// offset applied. bar is the internal scrollbar's screen hitbox when one was
// drawn.
type frameGeometry struct {
	contentX, contentY int
	contentW, contentH int
	titleH             int
	scroll             int
	bar                scrollbar.Hitbox
	hasBar             bool
	valid              bool
}

// translate maps the screen cell (x, y) into dialog content coordinates,
// reporting false when it falls outside the visible content area.
func (g frameGeometry) translate(x, y int) (int, int, bool) {
	if !g.valid || x < g.contentX || x >= g.contentX+g.contentW ||
		y < g.contentY || y >= g.contentY+g.contentH {
		return 0, 0, false
	}
	return x - g.contentX, y - g.contentY + g.scroll - g.titleH, true
}

// geometry locates the content area of a framed box once placed on screen:
// the overlay origin plus the Box style's left and top frame.
func (s Shell) geometry(
	box string,
	screenW, screenH, contentW, contentH, titleH, scroll int,
) frameGeometry {
	originX, originY := overlay.Origin(box, screenW, screenH, overlay.Center)
	b := s.cfg.Styles.Box
	return frameGeometry{
		contentX: originX + b.GetMarginLeft() + b.GetBorderLeftSize() + b.GetPaddingLeft(),
		contentY: originY + b.GetMarginTop() + b.GetBorderTopSize() + b.GetPaddingTop(),
		contentW: contentW,
		contentH: contentH,
		titleH:   titleH,
		scroll:   scroll,
		valid:    true,
	}
}

// Frame renders d's title, body, and hints into a box, scrolls the composed
// content when it exceeds the height cap, and composites it centered over
// backdrop, which must already be screenW columns by screenH rows.
func (s Shell) Frame(backdrop string, d Dialog, screenW, screenH int) string {
	out, _ := s.frame(backdrop, d, screenW, screenH)
	return out
}

// frame is Frame plus the placement geometry of this render, which the Stack
// records to translate mouse coordinates and expose the scrollbar hitbox.
func (s Shell) frame(backdrop string, d Dialog, screenW, screenH int) (string, frameGeometry) {
	frameW := s.cfg.Styles.Box.GetHorizontalFrameSize()
	frameH := s.cfg.Styles.Box.GetVerticalFrameSize()

	maxInnerW := max(1, s.clampWidth(screenW)-frameW)
	viewportH := max(1, s.clampHeight(screenH)-frameH)

	// A footered dialog (a scrolling form) pins its foot row below the viewport,
	// so the body scrolls but the hint/confirm/submitting row is always visible.
	if f, ok := d.(Footered); ok {
		return s.frameFootered(backdrop, d, f, screenW, screenH, maxInnerW, viewportH, frameW)
	}

	title := s.renderTitle(d.Title())
	hints := s.renderHints(d.Hints())

	// Wrap the body to the full inner width first. At its widest wrap the body
	// spans the fewest rows, so a body that fits the height cap here needs no
	// scrollbar and keeps every column - content that fits is never wrapped a
	// column early. Only a body that still overflows surrenders the last inner
	// column to the scrollbar, and is re-wrapped one column narrower so the
	// viewport never clips its final column once the scrollbar appears.
	body := d.Content(maxInnerW)
	// A body wider than the box means the dialog ignored the offered width (a
	// fixed-measure form on a tiny terminal); boxing it anyway would wrap its
	// borders into unreadable interleaved fragments, so show the notice instead.
	if lg.Width(body) > maxInnerW {
		return s.frameTooNarrow(backdrop, screenW, screenH, frameW), frameGeometry{}
	}
	inner := joinNonEmpty(title, body, hints)

	// RenderScrollable renders the body through Box.Width(BoxWidth), and lipgloss
	// counts the border and padding inside that width; BoxWidth is therefore the
	// box's total width, not its text width, so the frame has to be added back or
	// the text area comes up frameW columns short and wraps.
	viewW := min(lg.Width(inner), maxInnerW)
	boxW := viewW + frameW
	// lipgloss.Height counts \n-separated lines, matching RenderScrollable's own
	// overflow test, so this predicts exactly when it will show a scrollbar.
	scrolls := lg.Height(inner) > viewportH
	if scrolls {
		// The scrollbar costs scrollChrome columns - the bar itself plus the
		// padding column RenderScrollable adds beside it - so the body
		// surrenders both, or the bar wraps into a second row inside the box.
		// Re-wrap only when the body actually uses the surrendered columns; a
		// narrower body renders byte-identical, so the second Content call
		// would be pure waste.
		viewportW := max(1, maxInnerW-scrollChrome)
		if lg.Width(body) > viewportW {
			body = d.Content(viewportW)
			inner = joinNonEmpty(title, body, hints)
		}
		viewW = min(lg.Width(inner), viewportW)
		boxW = viewW + frameW + scrollChrome
	}

	titleH := 0
	if title != "" {
		titleH = lg.Height(title)
	}

	// A scroll-hinting dialog (a tall form) asks the viewport to follow its
	// focus. The hint's top is relative to the body, so offset it past the
	// Shell-drawn title, then prime the viewport's bounds before setting the
	// offset - SetYOffset clamps against the viewport's current height and
	// content, so priming first is what keeps a valid offset from collapsing to
	// zero (RenderScrollable re-sets the same bounds, leaving the offset intact).
	scroll := 0
	if sh, ok := d.(ScrollHint); ok {
		if top, height, ok := sh.ScrollTo(); ok {
			s.viewport.SetWidth(max(1, viewW))
			s.viewport.SetHeight(viewportH)
			s.viewport.SetContent(inner)
			s.viewport.SetYOffset(scrollOffset(top+titleH, height, viewportH))
			scroll = s.viewport.YOffset()
		}
	}

	scrollable := prompt.RenderScrollable(prompt.ScrollableModel{
		BoxStyle:        s.cfg.Styles.Box,
		BoxWidth:        boxW,
		Content:         inner,
		ViewportHeight:  viewportH,
		ViewWidth:       viewW,
		View:            s.viewport,
		ScrollbarConfig: s.cfg.Scrollbar,
		Styles:          prompt.Styles{Scrollbar: s.cfg.Styles.Scrollbar},
	})
	geo := s.geometry(
		scrollable, screenW, screenH, viewW, min(lg.Height(inner), viewportH), titleH, scroll,
	)
	if scrolls {
		geo.bar = scrollbar.Hitbox{
			X:          geo.contentX + viewW,
			Y:          geo.contentY,
			Height:     viewportH,
			TotalLines: lg.Height(inner),
			Config:     s.cfg.Scrollbar,
		}
		geo.hasBar = true
	}
	return overlay.Place(backdrop, scrollable, screenW, screenH, overlay.Center), geo
}

// frameFootered frames a Footered dialog: the foot row is pinned at the bottom
// of the box and the body scrolls in the space above it, so the body follows
// focus (via ScrollHint) while the hint/confirm row never scrolls off. It
// composes the viewport and scrollbar directly rather than through
// RenderScrollable, which owns its own box and could not host a pinned row
// inside it.
func (s Shell) frameFootered(
	backdrop string,
	d Dialog,
	f Footered,
	screenW, screenH, maxInnerW, viewportH, frameW int,
) (string, frameGeometry) {
	footer := f.Footer()
	// The body gets whatever height the foot row leaves; at least one row, so a
	// tiny terminal still shows a sliver of body rather than none.
	bodyH := max(1, viewportH-lg.Height(footer))

	// The scrolling part carries the same title/body/hints assembly as Frame -
	// implementing Footered moves only where the foot row is pinned, it never
	// voids the base Dialog contract's other parts.
	title := s.renderTitle(d.Title())
	hints := s.renderHints(d.Hints())

	// Wrap the body to the full inner width; only when the assembly still
	// overflows bodyH does it surrender the last column to the scrollbar and
	// re-wrap - the same fits-first rule Frame uses, so a body that fits keeps
	// every column. The trailing newline is trimmed so joining the footer adds
	// exactly one row between them, not a blank line.
	body := strings.TrimSuffix(d.Content(maxInnerW), "\n")
	// Same guard as Frame: a body or foot row wider than the box cannot be
	// boxed legibly, so the notice stands in until the terminal widens.
	if lg.Width(body) > maxInnerW || lg.Width(footer) > maxInnerW {
		return s.frameTooNarrow(backdrop, screenW, screenH, frameW), frameGeometry{}
	}
	inner := joinNonEmpty(title, body, hints)
	viewW := min(lg.Width(inner), maxInnerW)
	scrolls := lg.Height(inner) > bodyH
	if scrolls {
		vw := max(1, maxInnerW-1)
		// Same guard as Frame: a body narrower than the surrendered column
		// re-renders byte-identical, so skip the second Content call.
		if lg.Width(body) > vw {
			body = strings.TrimSuffix(d.Content(vw), "\n")
			inner = joinNonEmpty(title, body, hints)
		}
		viewW = min(lg.Width(inner), vw)
	}

	titleH := 0
	if title != "" {
		titleH = lg.Height(title)
	}

	scroll := 0
	bodyView := inner
	if scrolls {
		vp := s.viewport
		vp.SetWidth(max(1, viewW))
		vp.SetHeight(bodyH)
		vp.SetContent(inner)
		vp.SetYOffset(hintOffset(d, title, bodyH))
		scroll = vp.YOffset()
		bar := scrollbar.Model{
			Config:     s.cfg.Scrollbar,
			Height:     bodyH,
			TotalLines: vp.TotalLineCount(),
			Percent:    vp.ScrollPercent(),
			Styles:     s.cfg.Styles.Scrollbar,
		}.Render()
		bodyView = lg.JoinHorizontal(lg.Top, vp.View(), bar)
	}

	outer := lg.JoinVertical(lg.Left, bodyView, footer)
	boxW := lg.Width(outer) + frameW
	boxed := s.cfg.Styles.Box.Width(boxW).Render(outer)
	geo := s.geometry(boxed, screenW, screenH, viewW, min(lg.Height(inner), bodyH), titleH, scroll)
	if scrolls {
		geo.bar = scrollbar.Hitbox{
			X:          geo.contentX + viewW,
			Y:          geo.contentY,
			Height:     bodyH,
			TotalLines: lg.Height(inner),
			Config:     s.cfg.Scrollbar,
		}
		geo.hasBar = true
	}
	return overlay.Place(backdrop, boxed, screenW, screenH, overlay.Center), geo
}

// hintOffset is the viewport offset a scroll-hinting dialog asks for: its
// hinted region scrolled into a window of viewH lines, or zero when the
// dialog gives no hint. The hint is body-relative, so it is offset past the
// Shell-drawn title first.
func hintOffset(d Dialog, title string, viewH int) int {
	sh, ok := d.(ScrollHint)
	if !ok {
		return 0
	}
	top, height, ok := sh.ScrollTo()
	if !ok {
		return 0
	}
	if title != "" {
		top += lg.Height(title)
	}
	return scrollOffset(top, height, viewH)
}

// frameTooNarrow is the fallback frame when a dialog's rendering is wider than
// the clamped box can hold. Only the rendering is replaced: the dialog stays
// open and keyed, so esc still cancels it and widening the terminal restores
// the real content on the next frame.
func (s Shell) frameTooNarrow(backdrop string, screenW, screenH, frameW int) string {
	notice := s.cfg.Styles.HintText.Render("terminal too narrow")
	boxW := min(lg.Width(notice)+frameW, screenW)
	boxed := s.cfg.Styles.Box.Width(boxW).Render(notice)
	return overlay.Place(backdrop, boxed, screenW, screenH, overlay.Center)
}

// scrollOffset is the viewport's top line for keeping [top, top+height)
// visible in a window of viewportH lines: zero while the region already fits
// above the fold, otherwise just far enough to reveal its bottom - but never
// past its top, so a region taller than the window still shows its start rather
// than scrolling clean through it. The viewport clamps the result to its own
// content bounds.
func scrollOffset(top, height, viewportH int) int {
	if height < 1 {
		height = 1
	}
	off := 0
	if top+height > viewportH {
		off = top + height - viewportH
	}
	if off > top {
		off = top
	}
	return off
}

// renderTitle styles the heading, or returns "" when the dialog has none.
func (s Shell) renderTitle(title string) string {
	if title == "" {
		return ""
	}
	return s.cfg.Styles.Title.Render(title)
}

// renderHints renders the foot row with the Shell's injected styles.
func (s Shell) renderHints(hints []key.Hint) string {
	return RenderHints(s.cfg.Styles.HintKey, s.cfg.Styles.HintText, hints)
}

// RenderHints renders a hint row through primer's inline key renderer, the
// same "(y)es" styling the rest of a TUI uses; no hints means no row. Shared
// so a Footered dialog pinning its own hint row renders it identically to the
// Shell's.
func RenderHints(keyStyle, textStyle lg.Style, hints []key.Hint) string {
	if len(hints) == 0 {
		return ""
	}
	prefix := " "
	return key.Renderer{
		Styles: key.Styles{Key: keyStyle, Text: textStyle},
		Prefix: &prefix,
		Inline: true,
	}.Render(hints)
}

// joinNonEmpty stacks the parts that carry content, so an absent title or hint
// row leaves no blank line behind.
func joinNonEmpty(parts ...string) string {
	kept := parts[:0]
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return lg.JoinVertical(lg.Left, kept...)
}

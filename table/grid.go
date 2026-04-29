package table

import (
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
)

const defaultColumnPadding = 2

// Grid is a table of cell values (rows x columns) with alignment options.
type Grid struct {
	ColumnPadding int
	FlexCols      []int // indexes of columns that shrink to fit
	MaxWidth      int   // terminal width; flex columns shrink to fit (0 = disabled)
	Padding       Padding
	Rows          [][]string
	TTY           bool // when true, wrap spaces in SGR 8 to prevent tab optimization
}

// NewGrid creates a Grid with the given rows and applies any options.
// Default column padding is 2 spaces, left-aligned.
func NewGrid(rows [][]string, opts ...GridOption) *Grid {
	g := &Grid{
		Rows:          rows,
		ColumnPadding: defaultColumnPadding,
		Padding:       PaddingLeft,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// AlignColumns aligns the grid into padded strings with gaps between columns.
// It returns the aligned strings and the computed visible width of each column.
func (g *Grid) AlignColumns() ([]string, []int) {
	if len(g.Rows) == 0 {
		return nil, nil
	}

	// Compute max visible width per column.
	maxCols := 0
	for _, row := range g.Rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	colWidths := make([]int, maxCols)
	for _, row := range g.Rows {
		for c, field := range row {
			w := VisibleWidth(field)
			if w > colWidths[c] {
				colWidths[c] = w
			}
		}
	}

	flexCols := validFlexCols(g.FlexCols, maxCols)
	if len(flexCols) > 0 && g.MaxWidth > 0 {
		shrinkFlexWidths(colWidths, flexCols, g.MaxWidth, g.ColumnPadding)
		for _, flexCol := range flexCols {
			for i, row := range g.Rows {
				if flexCol < len(row) {
					g.Rows[i][flexCol] = truncateVisible(row[flexCol], colWidths[flexCol])
				}
			}
		}
	}

	// Format output with padding.
	gap := spaces(g.ColumnPadding, g.TTY)
	result := make([]string, len(g.Rows))
	for i, row := range g.Rows {
		var sb strings.Builder
		for c, field := range row {
			if c > 0 {
				sb.WriteString(gap)
			}
			pad := colWidths[c] - VisibleWidth(field)
			lastCol := c == len(row)-1
			switch g.Padding {
			case PaddingLeft:
				sb.WriteString(field)
				if !lastCol {
					sb.WriteString(spaces(pad, g.TTY))
				}
			case PaddingRight:
				sb.WriteString(spaces(pad, g.TTY))
				sb.WriteString(field)
			case PaddingCenter:
				left := pad / 2 //nolint:mnd // halve for centering
				right := pad - left
				sb.WriteString(spaces(left, g.TTY))
				sb.WriteString(field)
				if !lastCol {
					sb.WriteString(spaces(right, g.TTY))
				}
			}
		}
		result[i] = sb.String()
	}
	return result, colWidths
}

func validFlexCols(cols []int, maxCols int) []int {
	flexCols := make([]int, 0, len(cols))
	seen := make(map[int]bool, len(cols))
	for _, col := range cols {
		if col < 0 || col >= maxCols || seen[col] {
			continue
		}
		seen[col] = true
		flexCols = append(flexCols, col)
	}
	return flexCols
}

func shrinkFlexWidths(colWidths []int, flexCols []int, maxWidth, columnPadding int) {
	overflow := totalGridWidth(colWidths, columnPadding) - maxWidth
	for overflow > 0 {
		flexCol := widestShrinkableFlexCol(colWidths, flexCols)
		if flexCol < 0 {
			return
		}
		colWidths[flexCol]--
		overflow--
	}
}

func widestShrinkableFlexCol(colWidths []int, flexCols []int) int {
	const minFlexWidth = 1

	widestCol := -1
	widestWidth := minFlexWidth
	for _, col := range flexCols {
		if colWidths[col] > widestWidth {
			widestCol = col
			widestWidth = colWidths[col]
		}
	}
	return widestCol
}

func totalGridWidth(colWidths []int, columnPadding int) int {
	total := 0
	for _, width := range colWidths {
		total += width
	}
	if len(colWidths) > 1 {
		total += (len(colWidths) - 1) * columnPadding
	}
	return total
}

// VisibleWidth computes the visible width of a string, ignoring ANSI escapes.
func VisibleWidth(s string) int {
	return xansi.WcWidth.StringWidth(s)
}

// spaces returns n space characters. When tty is true, the spaces are wrapped
// in SGR 8 (conceal/hidden) to prevent bubbletea v2's hard-tab cursor
// optimization from collapsing runs of plain spaces into tab characters.
func spaces(n int, tty bool) string {
	if n <= 0 {
		return ""
	}
	s := strings.Repeat(" ", n)
	if tty {
		return "\x1b[8m" + s + "\x1b[28m"
	}
	return s
}

// truncateVisible truncates s to maxWidth visible characters, appending "…" if
// truncated. ANSI escape sequences are preserved but the visible text is cut.
func truncateVisible(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	w := VisibleWidth(s)
	if w <= maxWidth {
		return s
	}
	return xansi.WcWidth.Truncate(s, maxWidth-1, "…")
}

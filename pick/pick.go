package pick

import (
	"errors"
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"github.com/gechr/x/terminal"
)

func terminalWidthForPrompt(file *os.File) int {
	width, _ := terminal.Size(file)
	if width <= 0 {
		return 0
	}
	return width
}

// ErrCanceled is returned when the user cancels an interactive selection.
var ErrCanceled = errors.New("canceled")

// Item pairs a display string with a value of type T.
type Item[T any] struct {
	Display  string
	Value    T
	Selected bool
}

// selectPadding is the extra rows the huh multi-select adds for chrome (title row).
// Help text is rendered by the form outside the field, so it is not included here.
const selectPadding = 1

// selectHeight returns the clamped height for the multi-select widget.
func selectHeight(maxHeight, itemCount int) int {
	return min(maxHeight, itemCount+selectPadding)
}

// buildOptions converts a slice of Item into huh.Option values
// whose underlying value is the item's index.
func buildOptions[T any](items []Item[T]) []huh.Option[int] {
	opts := make([]huh.Option[int], len(items))
	for i, item := range items {
		opt := huh.NewOption(item.Display, i)
		if item.Selected {
			opt = opt.Selected(true)
		}
		opts[i] = opt
	}
	return opts
}

// collectValues maps selected indices back to item values.
func collectValues[T any](indices []int, items []Item[T]) []T {
	result := make([]T, len(indices))
	for i, idx := range indices {
		result[i] = items[idx].Value
	}
	return result
}

// MultiSelect presents a multi-select UI and returns the selected values.
// Returns ErrCanceled if the user cancels. Pass [WithFilter] to enable "/"
// filtering of the list.
func MultiSelect[T any](
	title string,
	items []Item[T],
	theme huh.Theme,
	maxHeight int,
	showHelp bool,
	opts ...Option,
) ([]T, error) {
	if len(items) == 0 {
		return nil, nil
	}

	cfg := newConfig(opts)
	options := buildOptions(items)

	var selected []int

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title(title).
				Options(options...).
				Value(&selected).
				Filterable(cfg.filterable).
				Height(selectHeight(maxHeight, len(items))),
		),
	)

	// Seed the form with the live terminal width so the first frame doesn't
	// render with huh's default width and then "snap" into place on resize.
	// Let huh negotiate height from the actual startup WindowSize event.
	if width := terminalWidthForPrompt(os.Stderr); width > 0 {
		form.WithWidth(width)
	}

	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit"))
	km.MultiSelect.Toggle = key.NewBinding(
		key.WithKeys("space", "x"),
		key.WithHelp("space", "toggle"),
	)

	form = form.WithTheme(theme).WithShowHelp(showHelp).WithKeyMap(km)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, fmt.Errorf("%w: %w", ErrCanceled, err)
		}
		return nil, err
	}

	return collectValues(selected, items), nil
}

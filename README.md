# primer

A [Charm](https://github.com/charmbracelet)-native library of reusable terminal UI primitives.

## Packages

| Package       | Description                                                                |
| ------------- | -------------------------------------------------------------------------- |
| `activity`    | Observational mutation log with monotonic IDs and bounded retention        |
| `cache`       | Generic TTL memoizer with newest-fetch-wins commits                        |
| `carousel`    | Wrapping tab strip with one active label                                   |
| `dialog`      | Modal overlay stack, framing shell, and async input grace                  |
| `filter`      | Smart-case text matching with `!` negate, `^` prefix, `$` suffix modifiers |
| `flash`       | Transient status message state with monotonic-ID expiry                    |
| `form`        | Multi-field modal form with focus ring, dirty guard, and autocomplete      |
| `helpbar`     | Wrapped footer hints                                                       |
| `helpsheet`   | Two-column keybinding overlay sheet with dismiss footer                    |
| `input`       | Text-entry wrappers, textarea factory, and external editor hop             |
| `key`         | Key-name constants, inline key highlighting, rebinding, and key replay     |
| `layout`      | Line normalization, ANSI-aware hard wrapping, and separator rendering      |
| `list`        | Scrollable cursor list over rendered rows with optional type-to-filter     |
| `overlay`     | Centered foreground placement over background content                      |
| `palette`     | Fuzzy type-to-filter command list with replayable keybindings              |
| `pick`        | Generic multi-select interactive prompt built on huh                       |
| `picker`      | Cursor-navigable options overlay with row/choice selection                 |
| `pill`        | Labeled cycle-selector rendering (`label ‹ value ›`)                       |
| `prompt`      | Scrollable modal prompts with choice groups, hints, and interaction state  |
| `render`      | Terminal markdown, theme-derived glamour styles, and diff rendering        |
| `scrollbar`   | Proportional scrollbar rendering and scroll position math                  |
| `scrollwheel` | Mouse wheel event coalescing for Bubble Tea filters                        |
| `table`       | ANSI-aware column alignment, typed sorting, and generic table rendering    |
| `task`        | Generation-tracked async work manager with per-scope staleness             |
| `titlebox`    | Rounded box with the title embedded in the top border                      |
| `view`        | Viewport body rendering and composable fullscreen frame footers            |

## Install

```text
go get github.com/gechr/primer@latest
```

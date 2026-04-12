# primer

A [Charm](https://github.com/charmbracelet)-native library of reusable terminal UI primitives.

## Packages

| Package          | Description                                                                |
| ---------------- | -------------------------------------------------------------------------- |
| `ansi/hyperlink` | OSC 8 terminal hyperlinks with TTY-aware fallback modes                    |
| `filter`         | Smart-case text matching with `!` negate, `^` prefix, `$` suffix modifiers |
| `flash`          | Transient status message state with monotonic-ID expiry                    |
| `human`          | Human-readable time formatting, path contraction, and path expansion       |
| `helpbar`        | Wrapped footer hints with right-aligned status text                        |
| `input`          | Textarea factory with sensible TUI defaults and functional options         |
| `helpsheet`      | Two-column keybinding overlay sheet with dismiss footer                    |
| `key`            | Shared key-name constants plus inline key-highlight rendering              |
| `layout`         | Line normalization, ANSI-aware hard wrapping, and separator rendering      |
| `overlay`        | Centered foreground placement over background content                      |
| `pick`           | Generic multi-select interactive prompt built on huh                       |
| `picker`         | Cursor-navigable options overlay with row/choice selection                 |
| `prompt`         | Scrollable modal prompts with choice groups, hints, and interaction state  |
| `render`         | Terminal markdown (glamour) and diff (chroma) rendering                    |
| `scrollbar`      | Proportional scrollbar rendering and scroll position math                  |
| `scrollwheel`    | Mouse wheel event coalescing for Bubble Tea filters                        |
| `table`          | ANSI-aware column alignment, typed sorting, and generic table rendering    |
| `term`           | Terminal detection and size queries                                        |
| `view`           | Viewport body rendering and fullscreen frame composition                   |

## Install

```text
go get github.com/gechr/primer@latest
```

# primer

A [Charm](https://github.com/charmbracelet)-native library of reusable terminal UI primitives.

## Packages

| Package          | Description                                                                |
| ---------------- | -------------------------------------------------------------------------- |
| `ansi/hyperlink` | OSC 8 terminal hyperlinks with TTY-aware fallback modes                    |
| `filter`         | Smart-case text matching with `!` negate, `^` prefix, `$` suffix modifiers |
| `flash`          | Transient status message state with monotonic-ID expiry                    |
| `help/bar`       | Wrapped footer hints with right-aligned status text                        |
| `help/sheet`     | Two-column keybinding overlay sheet with dismiss footer                    |
| `keyhint`        | Inline key-highlight rendering for help bars                               |
| `layout`         | Line normalization and terminal-width padding                              |
| `overlay`        | Centered foreground placement over background content                      |
| `prompt`         | Scrollable modal prompts with choice groups, hints, and interaction state  |
| `render`         | Terminal markdown (glamour) and diff (chroma) rendering                    |
| `scroll/bar`     | Proportional scrollbar rendering and scroll position math                  |
| `scroll/wheel`   | Mouse wheel event coalescing for Bubble Tea filters                        |
| `table`          | ANSI-aware column alignment, typed sorting, and generic table rendering    |
| `term`           | Terminal detection and size queries                                        |
| `view`           | Viewport body rendering and fullscreen frame composition                   |

## Install

```text
go get github.com/gechr/primer@latest
```

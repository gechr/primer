package input

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	xshell "github.com/gechr/x/shell"
)

// ExternalEditor resolves the external editor command: override wins (the seam
// an app hangs its own setting on), $EDITOR is the fallback, "" means none -
// multiline fields then use the in-TUI textarea instead. One rule, no
// per-field knobs.
func ExternalEditor(override string) string {
	if override != "" {
		return override
	}
	return os.Getenv("EDITOR")
}

// ExternalEditorFinishedMsg delivers the text written in the external editor.
// ID is the opener's routing tag (e.g. "comment:42") so the consumer knows
// which flow to resume. Err is set when the editor could not run or the buffer
// could not be read back.
type ExternalEditorFinishedMsg struct {
	ID   string
	Text string
	Err  error
}

// ExternalEditorCmd suspends the TUI and opens the external editor on a temp
// file seeded with initial; the result returns as an ExternalEditorFinishedMsg.
// editor is resolved through [ExternalEditor] (pass "" to use $EDITOR) and
// runs through the shell, so commands with flags ("nvim -f") work.
func ExternalEditorCmd(editor, id, initial string) tea.Cmd {
	ed := ExternalEditor(editor)
	if ed == "" {
		return func() tea.Msg {
			return ExternalEditorFinishedMsg{
				ID:  id,
				Err: errors.New("no editor configured (set $EDITOR)"),
			}
		}
	}
	f, err := os.CreateTemp("", "primer-edit-*.md")
	if err != nil {
		return func() tea.Msg { return ExternalEditorFinishedMsg{ID: id, Err: err} }
	}
	path := f.Name()
	if _, err := f.WriteString(initial); err != nil {
		cleanupErr := errors.Join(f.Close(), os.Remove(path))
		return func() tea.Msg { return ExternalEditorFinishedMsg{ID: id, Err: errors.Join(err, cleanupErr)} }
	}
	if err := f.Close(); err != nil {
		removeErr := os.Remove(path)
		return func() tea.Msg { return ExternalEditorFinishedMsg{ID: id, Err: errors.Join(err, removeErr)} }
	}

	// ed is deliberately unquoted so values with flags ("nvim -f", "code -w")
	// work; the trade-off is that an editor binary living at a path with
	// spaces must be wrapped in quotes inside the variable itself, same as git.
	sh := ed + " " + xshell.Quote(path)
	cmd := exec.Command("sh", "-c", sh) //nolint:noctx // tea.ExecProcess owns the lifecycle
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		defer os.Remove(path) // completed editor cleanup is best-effort
		if err != nil {
			return ExternalEditorFinishedMsg{ID: id, Err: err}
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return ExternalEditorFinishedMsg{ID: id, Err: rerr}
		}
		return ExternalEditorFinishedMsg{ID: id, Text: strings.TrimRight(string(b), "\n")}
	})
}

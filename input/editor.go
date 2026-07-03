package input

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
)

const (
	editorDefaultWidth     = 120
	editorDefaultBodyMin   = 3
	editorDefaultChrome    = 8 // header + blank + title label + title + blank + body label + blank + help
	editorDefaultTitleYOff = 3 // header + blank + title label
	editorDefaultBodyYOff  = 5 // header + blank + title label + title-end + blank + body label
	editorFieldCount       = 2
)

const keyTab = "tab"

// EditorEntry defines an item to edit.
type EditorEntry struct {
	Label string // display label (e.g. "owner/repo#123")
	Title string // initial title value
}

// EditorResult holds the outcome for a single edited entry.
type EditorResult struct {
	Label   string
	Title   string
	Body    string
	Changed bool
}

// EditorStyles controls the visual appearance of the editor.
type EditorStyles struct {
	BlurredText lg.Style
	Counter     lg.Style
	Dirty       lg.Style
	DimLabel    lg.Style
	FocusedText lg.Style
	Header      lg.Style
	HelpKey     lg.Style
	HelpText    lg.Style
	Label       lg.Style
}

// BodyFetchFunc fetches the body for an entry at the given index.
type BodyFetchFunc func(index int) (string, error)

type editorField int

const (
	editorFieldTitle editorField = iota
	editorFieldBody
)

type editorEntry struct {
	label        string
	origTitle    string
	origBody     string
	title        textinput.Model
	body         textarea.Model
	bodyFetched  bool
	bodyFetchErr error
	bodyFetching bool
	focus        editorField
}

// bodyFetchedMsg is sent when a body has been fetched.
type bodyFetchedMsg struct {
	index int
	body  string
	err   error
}

// Editor is a Bubble Tea model for editing multiple title+body entries.
type Editor struct {
	entries    []editorEntry
	current    int
	styles     EditorStyles
	fetchBody  BodyFetchFunc
	termHeight int
	bodyMin    int
	width      int

	Submitted bool
	Aborted   bool
}

// NewEditor creates an Editor for the given entries.
func NewEditor(entries []EditorEntry, opts ...EditorOption) Editor {
	cfg := editorConfig{
		bodyMinHeight: editorDefaultBodyMin,
		width:         editorDefaultWidth,
	}
	for _, o := range opts {
		o(&cfg)
	}

	ee := make([]editorEntry, len(entries))
	for i, e := range entries {
		ti := newEditorTextInput(cfg.styles, cfg.width, e.Title)
		ta := newEditorTextArea(cfg.styles, cfg.width, cfg.bodyMinHeight)

		if i == 0 {
			ti.Focus()
		} else {
			ti.Blur()
		}
		ta.Blur()

		ee[i] = editorEntry{
			label:     e.Label,
			origTitle: e.Title,
			title:     ti,
			body:      ta,
			focus:     editorFieldTitle,
		}
	}

	return Editor{
		entries:   ee,
		styles:    cfg.styles,
		fetchBody: cfg.fetchBody,
		bodyMin:   cfg.bodyMinHeight,
		width:     cfg.width,
	}
}

func newEditorTextInput(styles EditorStyles, width int, value string) textinput.Model {
	tiStyles := textinput.DefaultDarkStyles()
	tiStyles.Focused.Text = styles.FocusedText
	tiStyles.Blurred.Text = styles.BlurredText
	tiStyles.Cursor.Shape = tea.CursorBlock
	tiStyles.Cursor.Blink = true

	ti := textinput.New()
	ti.Prompt = ""
	ti.SetWidth(width)
	ti.SetValue(value)
	ti.SetStyles(tiStyles)
	ti.SetVirtualCursor(false)
	return ti
}

func newEditorTextArea(styles EditorStyles, width, minHeight int) textarea.Model {
	taStyles := textarea.DefaultDarkStyles()
	taStyles.Focused.Text = styles.FocusedText
	taStyles.Focused.CursorLine = styles.FocusedText
	taStyles.Blurred.CursorLine = styles.BlurredText
	taStyles.Blurred.Text = styles.BlurredText
	taStyles.Cursor.Shape = tea.CursorBlock
	taStyles.Cursor.Blink = true

	ta := textarea.New()
	ta.Prompt = ""
	ta.SetWidth(width)
	ta.ShowLineNumbers = false
	ta.SetHeight(minHeight)
	ta.SetStyles(taStyles)
	ta.SetVirtualCursor(false)
	return ta
}

// Init implements tea.Model.
func (m Editor) Init() tea.Cmd {
	if len(m.entries) > 0 && m.fetchBody != nil {
		return m.fetchBodyCmd(0)
	}
	return nil
}

func (m Editor) fetchBodyCmd(index int) tea.Cmd {
	fetch := m.fetchBody
	return func() tea.Msg {
		body, err := fetch(index)
		return bodyFetchedMsg{index: index, body: body, err: err}
	}
}

// Update implements tea.Model.
func (m Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case bodyFetchedMsg:
		e := &m.entries[msg.index]
		e.bodyFetching = false
		if msg.err != nil {
			e.bodyFetchErr = msg.err
		} else {
			e.bodyFetched = true
			e.origBody = msg.body
			e.body.SetValue(msg.body)
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.termHeight = msg.Height
		for i := range m.entries {
			m.entries[i].title.SetWidth(msg.Width)
			m.entries[i].body.SetWidth(msg.Width)
			m.entries[i].body.SetHeight(m.bodyHeight())
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			m.Aborted = true
			return m, tea.Quit
		case keyTab, "shift+tab":
			return m.cycleField(msg.String())
		case "ctrl+s":
			m.Submitted = true
			return m, tea.Quit
		case "ctrl+n":
			return m.navigate(1)
		case "ctrl+p":
			return m.navigate(-1)
		}
	}

	return m.updateFocused(msg)
}

func (m Editor) bodyHeight() int {
	if m.termHeight == 0 {
		return m.bodyMin
	}
	if h := m.termHeight - editorDefaultChrome; h > m.bodyMin {
		return h
	}
	return m.bodyMin
}

func (m Editor) navigate(delta int) (tea.Model, tea.Cmd) {
	next := m.current + delta
	if next < 0 || next >= len(m.entries) {
		return m, nil
	}

	cur := &m.entries[m.current]
	cur.title.Blur()
	cur.body.Blur()

	m.current = next

	e := &m.entries[m.current]
	e.focus = editorFieldTitle
	cmd := e.title.Focus()
	e.body.Blur()

	cmds := []tea.Cmd{cmd, tea.ClearScreen}

	if m.fetchBody != nil && !e.bodyFetched && !e.bodyFetching && e.bodyFetchErr == nil {
		e.bodyFetching = true
		cmds = append(cmds, m.fetchBodyCmd(m.current))
	}

	return m, tea.Batch(cmds...)
}

func (m Editor) cycleField(k string) (tea.Model, tea.Cmd) {
	e := &m.entries[m.current]

	if k == keyTab {
		e.focus = (e.focus + 1) % editorFieldCount
	} else {
		e.focus = (e.focus - 1 + editorFieldCount) % editorFieldCount
	}

	var cmd tea.Cmd
	switch e.focus {
	case editorFieldTitle:
		cmd = e.title.Focus()
		e.body.Blur()
	case editorFieldBody:
		e.title.Blur()
		cmd = e.body.Focus()
	}
	return m, cmd
}

func (m Editor) updateFocused(msg tea.Msg) (tea.Model, tea.Cmd) {
	e := &m.entries[m.current]
	var cmd tea.Cmd
	switch e.focus {
	case editorFieldTitle:
		e.title, cmd = e.title.Update(msg)
	case editorFieldBody:
		e.body, cmd = e.body.Update(msg)
	}
	return m, cmd
}

const nl = "\n"

// View implements tea.Model.
func (m Editor) View() tea.View {
	e := &m.entries[m.current]
	var b strings.Builder

	b.WriteString(m.styles.Header.Render(e.label))
	if e.title.Value() != e.origTitle || (e.bodyFetched && e.body.Value() != e.origBody) {
		b.WriteString(m.styles.Dirty.Render(" (edited)"))
	}
	if len(m.entries) > 1 {
		b.WriteString(" ")
		b.WriteString(m.styles.Counter.Render(
			fmt.Sprintf("(%d/%d)", m.current+1, len(m.entries)),
		))
	}
	b.WriteString(nl + nl)

	titleLabel := m.styles.Label
	if e.focus != editorFieldTitle {
		titleLabel = m.styles.DimLabel
	}
	b.WriteString(titleLabel.Render("Title"))
	b.WriteString(nl)

	titleView := e.title.View()
	titleLines := strings.Count(titleView, nl) + 1
	b.WriteString(titleView)
	b.WriteString(nl + nl)

	bodyLabel := m.styles.Label
	if e.focus != editorFieldBody {
		bodyLabel = m.styles.DimLabel
	}
	b.WriteString(bodyLabel.Render("Body"))
	b.WriteString(nl)

	switch {
	case e.bodyFetching:
		b.WriteString(m.styles.BlurredText.Render("Loading…"))
	case e.bodyFetchErr != nil:
		b.WriteString(m.styles.BlurredText.Render(fmt.Sprintf("Error: %v", e.bodyFetchErr)))
	default:
		b.WriteString(e.body.View())
	}
	b.WriteString(nl + nl)

	b.WriteString(m.renderHelp())

	v := tea.NewView(b.String())

	if !e.bodyFetching && e.bodyFetchErr == nil {
		switch e.focus {
		case editorFieldTitle:
			if cur := e.title.Cursor(); cur != nil {
				cur.Y += editorDefaultTitleYOff
				v.Cursor = cur
			}
		case editorFieldBody:
			if cur := e.body.Cursor(); cur != nil {
				cur.Y += editorDefaultBodyYOff + titleLines
				v.Cursor = cur
			}
		}
	}

	return v
}

func (m Editor) renderHelp() string {
	type hint struct {
		key  string
		desc string
	}
	pairs := []hint{
		{keyTab, "switch field"},
		{"ctrl+s", "save all"},
		{"esc", "cancel"},
	}
	if len(m.entries) > 1 {
		pairs = append([]hint{
			{"ctrl+n", "next"},
			{"ctrl+p", "prev"},
		}, pairs...)
	}
	var parts []string
	for _, p := range pairs {
		parts = append(parts,
			m.styles.HelpKey.Render(p.key)+" "+m.styles.HelpText.Render(p.desc),
		)
	}
	return strings.Join(parts, m.styles.HelpText.Render(" · "))
}

// Results returns the outcome for all entries.
func (m Editor) Results() []EditorResult {
	results := make([]EditorResult, len(m.entries))
	for i, e := range m.entries {
		title := e.title.Value()
		body := e.body.Value()
		results[i] = EditorResult{
			Label:   e.label,
			Title:   title,
			Body:    body,
			Changed: title != e.origTitle || body != e.origBody,
		}
	}
	return results
}

// Run launches the editor as a standalone Bubble Tea program and returns results.
func Run(entries []EditorEntry, opts ...EditorOption) ([]EditorResult, bool, error) {
	m := NewEditor(entries, opts...)
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, false, fmt.Errorf("editor: %w", err)
	}
	em, ok := final.(Editor)
	if !ok {
		return nil, false, fmt.Errorf("editor: unexpected model type")
	}
	if em.Aborted || !em.Submitted {
		return nil, false, nil
	}
	return em.Results(), true, nil
}

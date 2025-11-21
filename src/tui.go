package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Enums & Types ---

type sessionState int

const (
	stateDefault sessionState = iota
	stateFilePicker
	stateSaveFilepath
	stateSelectModel
	stateSelectQType
	stateEnterSentences
)

type (
	fileReadMsg         struct{ content []byte; path string }
	fileWriteMsg        struct{ path string; err error }
	generationResultMsg struct{ text string; err error }
	resetStatusMsg      struct{}
	errMsg              struct{ err error }
)

func (e errMsg) Error() string { return e.err.Error() }

// --- Commands ---

func readFileCmd(path string) tea.Cmd {
	return func() tea.Msg {
		c, err := ioutil.ReadFile(path)
		if err != nil {
			return errMsg{err}
		}
		return fileReadMsg{content: c, path: path}
	}
}

func writeFileCmd(path, content string) tea.Cmd {
	return func() tea.Msg {
		if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
			return errMsg{err}
		}
		return fileWriteMsg{path: path}
	}
}

func generateCmd(apiKey, modelID, systemPrompt, userPrompt string) tea.Cmd {
	return func() tea.Msg {
		output, err := callChatGPT(apiKey, modelID, systemPrompt, userPrompt)
		if err != nil {
			return generationResultMsg{err: err}
		}
		return generationResultMsg{text: output}
	}
}

func resetStatusCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return resetStatusMsg{}
	})
}


// --- Styles ---
var (
	docStyle     = lipgloss.NewStyle().Margin(1, 2)
	focusedStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
	blurredStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)


// --- Model ---

type model struct {
	state         sessionState
	err           error
	status        string
	defaultStatus string

	// UI Components
	inputs     []textarea.Model
	focused    int
	pathInput  textinput.Model
	numInput   textinput.Model
	list       list.Model
	filepicker filepicker.Model

	// Content
	inputFilePath string
	apiKey        string

	// Generation Parameters
	selectedModel string
	selectedQType string
	numSentences  string
}

const (
	inputIdx  = 0
	outputIdx = 1
)

func initialModel() model {
	apiKey, err := loadAPIKey()
	if err != nil {
		log.Printf("Failed to load API key: %v", err)
	}

	defaultStatus := "Ctrl+O: Load | Ctrl+S: Save | Ctrl+G: Generate | Tab: Switch Panes"
	m := model{
		state:         stateDefault,
		status:        defaultStatus,
		defaultStatus: defaultStatus,
		inputs:        make([]textarea.Model, 2),
		focused:       0,
		apiKey:        apiKey,
		numSentences:  "2",
	}

	// Textareas
	for i := range m.inputs {
		t := textarea.New()
		t.ShowLineNumbers = true
		t.FocusedStyle.Base = focusedStyle
		t.BlurredStyle.Base = blurredStyle
		if i == inputIdx {
			t.Placeholder = "Load a vocabulary file or type 'word = meaning' here."
			t.Focus()
		} else {
			t.Placeholder = "Generated questions will appear here."
		}
		m.inputs[i] = t
	}

	// Inputs
	m.pathInput = textinput.New()
	m.pathInput.Placeholder = "Save file as..."
	m.pathInput.CharLimit = 256
	m.pathInput.Width = 80
	
m.numInput = textinput.New()
	m.numInput.Placeholder = "2"
	m.numInput.CharLimit = 2
	m.numInput.Width = 5

	// List
	m.list = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.list.SetShowHelp(false)

	// Filepicker
	fp := filepicker.New()
	fp.AllowedTypes = []string{ ".txt"}
	fp.CurrentDirectory, _ = filepath.Abs("./Tester") // Start in the Tester directory
	m.filepicker = fp

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.filepicker.Init())
}

// --- Update ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		panelWidth := (msg.Width - h) / 2
		listWidth := msg.Width - h
		panelHeight := msg.Height - v - 3

		for i := range m.inputs {
			m.inputs[i].SetWidth(panelWidth)
			m.inputs[i].SetHeight(panelHeight)
		}
		m.list.SetSize(listWidth, panelHeight)
		m.filepicker.Height = panelHeight
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// State-specific updates
		switch m.state {
		case stateFilePicker:
			// If the user presses escape, cancel the file picker
			if msg.String() == "esc" {
				m.state = stateDefault
				m.status = "File selection cancelled."
				return m, resetStatusCmd()
			}
			var cmd tea.Cmd
			m.filepicker, cmd = m.filepicker.Update(msg)
			if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
				m.state = stateDefault
				return m, readFileCmd(path)
			}
			return m, cmd
		case stateSaveFilepath:
			return updatePathInput(msg, m)
		case stateSelectModel, stateSelectQType:
			return updateListSelection(msg, m)
		case stateEnterSentences:
			return updateNumInput(msg, m)
		default:
			return updateDefault(msg, m)
		}

	case resetStatusMsg:
		m.status = m.defaultStatus
		return m, nil
	
	case fileReadMsg:
		m.inputs[inputIdx].SetValue(string(msg.content))
		m.inputFilePath = msg.path
		m.status = fmt.Sprintf("Loaded '%s'", filepath.Base(msg.path))
		m.state = stateDefault
		return m, resetStatusCmd()

	case fileWriteMsg:
		m.status = fmt.Sprintf("Saved to '%s'", filepath.Base(msg.path))
		m.state = stateDefault
		return m, resetStatusCmd()

	case generationResultMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Generation Error: %v", msg.err)
		} else {
			m.status = "Generation complete!"
			m.inputs[outputIdx].SetValue(msg.text)
		}
		m.state = stateDefault
		return m, resetStatusCmd()

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	// Update focused textarea in default state
	if m.state == stateDefault {
		var taCmd tea.Cmd
		m.inputs[m.focused], taCmd = m.inputs[m.focused].Update(msg)
		cmds = append(cmds, taCmd)
	}

	return m, tea.Batch(cmds...)
}

func updateDefault(msg tea.KeyMsg, m model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+o":
		m.state = stateFilePicker
		m.status = "Select a file to load."
		return m, m.filepicker.Init()

	case "ctrl+s":
		m.state = stateSaveFilepath
		originalName := "result"
		if m.inputFilePath != "" {
			originalName = strings.TrimSuffix(filepath.Base(m.inputFilePath), ".txt")
		}
		m.pathInput.SetValue(fmt.Sprintf("%s_problem.txt", originalName))
		m.pathInput.Focus()
		m.status = "Enter file path to save."
		return m, nil

	case "ctrl+g":
		if m.inputs[inputIdx].Value() == "" {
			m.status = "Cannot generate: Input vocabulary is empty."
			return m, resetStatusCmd()
		}
		if m.apiKey == "" {
			m.status = "Cannot generate: API Key is not configured in api.json."
			return m, resetStatusCmd()
		}
		m.state = stateSelectModel
		m.list.Title = "Select a Model"
		m.list.SetItems(getGenerationModels())
		return m, nil

	case "tab":
		m.inputs[m.focused].Blur()
		m.focused = (m.focused + 1) % len(m.inputs)
		m.inputs[m.focused].Focus()
		return m, textarea.Blink
	}
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func updatePathInput(msg tea.KeyMsg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		path := m.pathInput.Value()
		if path == "" { return m, nil }
		m.state = stateDefault
		m.status = "Saving..."
		return m, writeFileCmd(path, m.inputs[outputIdx].Value())
	case "esc":
		m.state = stateDefault
		m.status = "Cancelled save."
		return m, resetStatusCmd()
	}
	m.pathInput, cmd = m.pathInput.Update(msg)
	return m, cmd
}

func updateListSelection(msg tea.KeyMsg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		item := m.list.SelectedItem().(item)
		if m.state == stateSelectModel {
			m.selectedModel = item.id
			m.state = stateSelectQType
			m.list.Title = "Select Question Type"
			m.list.SetItems(getQTypes())
			if item.desc == "Warning: High Cost" {
				m.status = "Warning: High cost model selected!"
			}
		} else if m.state == stateSelectQType {
			m.selectedQType = item.id
			if m.selectedQType == "빈칸 추론" {
				m.state = stateEnterSentences
				m.numInput.SetValue(m.numSentences)
				m.numInput.Focus()
				m.status = "Enter number of sentences."
			} else {
				m.state = stateDefault
				m.status = "Generating..."
				parsed := parseVocabBlock(m.inputs[inputIdx].Value())
				system, user := buildPrompts(parsed, m.selectedQType, 1)
				return m, generateCmd(m.apiKey, m.selectedModel, system, user)
			}
		}
		return m, nil
	case "esc":
		m.state = stateDefault
		m.status = "Cancelled generation."
		return m, resetStatusCmd()
	}
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func updateNumInput(msg tea.KeyMsg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		m.numSentences = m.numInput.Value()
		m.state = stateDefault
		m.status = "Generating..."
		num, _ := strconv.Atoi(m.numSentences)
		parsed := parseVocabBlock(m.inputs[inputIdx].Value())
		system, user := buildPrompts(parsed, m.selectedQType, num)
		return m, generateCmd(m.apiKey, m.selectedModel, system, user)
	case "esc":
		m.state = stateDefault
		m.status = "Cancelled generation."
		return m, resetStatusCmd()
	}
	m.numInput, cmd = m.numInput.Update(msg)
	return m, cmd
}

// --- View ---

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\nPress ctrl+c to exit.", m.err)
	}

	switch m.state {
	case stateFilePicker:
		return docStyle.Render(m.filepicker.View())
	case stateSaveFilepath:
		return docStyle.Render(fmt.Sprintf("Save file as:\n\n%s", m.pathInput.View()) + "\n\nEnter: confirm | Esc: cancel")
	case stateSelectModel, stateSelectQType:
		return docStyle.Render(m.list.View())
	case stateEnterSentences:
		return docStyle.Render(fmt.Sprintf("Enter number of sentences:\n\n%s", m.numInput.View()) + "\n\nEnter: confirm | Esc: cancel")
	default:
		panels := lipgloss.JoinHorizontal(lipgloss.Top, m.inputs[inputIdx].View(), m.inputs[outputIdx].View())
		return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left, panels, helpStyle.Render(m.status)))
	}
}

// --- List Items ---

type item struct {
	title, desc, id string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func getGenerationModels() []list.Item {
	return []list.Item{
		item{title: "GPT-5 pro", id: "gpt-5-pro", desc: "Warning: High Cost"},
		item{title: "GPT-5", id: "gpt-5"},
		item{title: "GPT-5 mini", id: "gpt-5-mini"},
		item{title: "GPT-5 nano", id: "gpt-5-nano"},
		item{title: "GPT-4.1", id: "gpt-4.1", desc: "Legacy"},
	}
}

func getQTypes() []list.Item {
	return []list.Item{
		item{title: "빈칸 추론", id: "빈칸 추론"},
		item{title: "영영풀이", id: "영영풀이"},
		item{title: "뜻풀이 판단", id: "뜻풀이 판단"},
	}
}

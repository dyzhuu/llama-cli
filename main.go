package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg      error
	responseMsg string
	completeMsg struct{}
)

func listenForMessage(channel chan string, message string) tea.Cmd {
	return func() tea.Msg {
		StreamRequest(message, channel)
		return completeMsg{}
	}
}

func waitForMessage(channel chan string) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-channel)
	}
}

type model struct {
	channel       chan string
	viewport      viewport.Model
	messages      []string
	textarea      textarea.Model
	width         int
	height        int
	senderStyle   lipgloss.Style
	receiverStyle lipgloss.Style
	err           error
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 2000

	ta.SetHeight(2)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(0, 0)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		channel:       make(chan string),
		textarea:      ta,
		messages:      []string{},
		viewport:      vp,
		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		receiverStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("229")),
		err:           nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)

	textAreaHeight := m.textarea.Length()/m.textarea.Width() + 2

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - 1
		m.viewport.Width = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			prompt := m.textarea.Value()

			m.messages = append(m.messages, m.senderStyle.Render("You: ")+prompt)
			m.textarea.Reset()
			m.textarea.SetHeight(2)
			m.viewport.Height = m.height - 3
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			m.textarea.Blur()

			m.channel = make(chan string)
			m.messages = append(m.messages, "")

			return m, tea.Batch(
				tiCmd,
				listenForMessage(m.channel, prompt),
				waitForMessage(m.channel),
			)
		}
	case responseMsg:
		lastIndex := len(m.messages) - 1
		if m.messages[lastIndex] == "" {
			m.messages[lastIndex] = m.receiverStyle.Render("Llama: ")
		}
		m.messages[lastIndex] = m.messages[lastIndex] + string(msg)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, waitForMessage(m.channel)
	case completeMsg:
		lastIndex := len(m.messages) - 1
		m.messages[lastIndex] = m.messages[lastIndex] + "\n"
		m.textarea.Focus()
		return m, nil
	case errMsg:
		m.err = msg
		return m, nil
	}
	m.viewport.Height = m.height - textAreaHeight - 1
	m.textarea.SetHeight(textAreaHeight)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	)
}

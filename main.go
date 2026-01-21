package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateList state = iota
	stateInput
	stateGenerating
	stateTemplateInput
	stateManualDescription
)

type model struct {
	storage          *Storage
	currentState     state
	input            string
	cursor           int
	selected         int
	err              error
	generating       bool
	copyMessage      string
	templateVars     []string
	templateValues   map[string]string
	currentVarIndex  int
	templateCommand  string
	pendingCommand   string
}

type generatedMsg struct {
	command     string
	description string
	err         error
}

type copiedMsg struct {
	success bool
	err     error
}

func generateDescription(command string) tea.Cmd {
	return func() tea.Msg {
		desc, err := GenerateDescription(command)
		return generatedMsg{
			command:     command,
			description: desc,
			err:         err,
		}
	}
}

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("pbcopy")
		pipe, err := cmd.StdinPipe()
		if err != nil {
			return copiedMsg{success: false, err: err}
		}

		if err := cmd.Start(); err != nil {
			return copiedMsg{success: false, err: err}
		}

		_, err = pipe.Write([]byte(text))
		if err != nil {
			return copiedMsg{success: false, err: err}
		}

		pipe.Close()
		if err := cmd.Wait(); err != nil {
			return copiedMsg{success: false, err: err}
		}

		return copiedMsg{success: true}
	}
}

func initialModel() model {
	storage, err := LoadCommands()
	if err != nil {
		storage = &Storage{Commands: []Command{}}
	}

	return model{
		storage:      storage,
		currentState: stateList,
		selected:     0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.currentState {
		case stateList:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.selected > 0 {
					m.selected--
				}
				m.copyMessage = ""
			case "down", "j":
				if m.selected < len(m.storage.Commands)-1 {
					m.selected++
				}
				m.copyMessage = ""
			case "n":
				m.currentState = stateInput
				m.input = ""
				m.copyMessage = ""
			case "c", "enter":
				if len(m.storage.Commands) > 0 {
					command := m.storage.Commands[m.selected].Command
					vars := ExtractTemplateVars(command)

					if len(vars) > 0 {
						// Command has template variables, prompt for values
						m.currentState = stateTemplateInput
						m.templateVars = vars
						m.templateValues = make(map[string]string)
						m.currentVarIndex = 0
						m.templateCommand = command
						m.input = ""
					} else {
						// No template variables, copy directly
						return m, copyToClipboard(command)
					}
				}
			case "d":
				if len(m.storage.Commands) > 0 {
					m.storage.Commands = append(
						m.storage.Commands[:m.selected],
						m.storage.Commands[m.selected+1:]...,
					)
					if m.selected >= len(m.storage.Commands) && m.selected > 0 {
						m.selected--
					}
					SaveCommands(m.storage)
					m.copyMessage = ""
				}
			}

		case stateInput:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.currentState = stateList
				m.input = ""
			case tea.KeyEnter:
				if m.input != "" {
					m.currentState = stateGenerating
					return m, generateDescription(m.input)
				}
			case tea.KeyBackspace:
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case tea.KeySpace:
				m.input += " "
			case tea.KeyRunes:
				// Handle both typing and paste (paste comes through as multiple runes)
				m.input += string(msg.Runes)
			}

		case stateTemplateInput:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.currentState = stateList
				m.input = ""
				m.templateVars = nil
				m.templateValues = nil
				m.templateCommand = ""
			case tea.KeyEnter:
				// Save the current variable's value
				currentVar := m.templateVars[m.currentVarIndex]
				m.templateValues[currentVar] = m.input
				m.input = ""

				// Move to next variable or finish
				m.currentVarIndex++
				if m.currentVarIndex >= len(m.templateVars) {
					// All variables collected, substitute and copy
					finalCommand := SubstituteTemplateVars(m.templateCommand, m.templateValues)
					m.currentState = stateList
					m.templateVars = nil
					m.templateValues = nil
					m.templateCommand = ""
					m.currentVarIndex = 0
					return m, copyToClipboard(finalCommand)
				}
			case tea.KeyBackspace:
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case tea.KeySpace:
				m.input += " "
			case tea.KeyRunes:
				m.input += string(msg.Runes)
			}

		case stateManualDescription:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.currentState = stateList
				m.input = ""
				m.pendingCommand = ""
				m.err = nil
			case tea.KeyEnter:
				if m.input != "" {
					// Save command with manual description
					m.storage.Commands = append(m.storage.Commands, Command{
						Command:     m.pendingCommand,
						Description: m.input,
					})
					SaveCommands(m.storage)
					m.currentState = stateList
					m.selected = len(m.storage.Commands) - 1
					m.input = ""
					m.pendingCommand = ""
					m.err = nil
				}
			case tea.KeyBackspace:
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case tea.KeySpace:
				m.input += " "
			case tea.KeyRunes:
				m.input += string(msg.Runes)
			}
		}

	case generatedMsg:
		if msg.err != nil {
			// AI generation failed, prompt for manual description
			m.pendingCommand = msg.command
			m.currentState = stateManualDescription
			m.input = ""
			m.err = msg.err
		} else {
			m.storage.Commands = append(m.storage.Commands, Command{
				Command:     msg.command,
				Description: msg.description,
			})
			SaveCommands(m.storage)
			m.currentState = stateList
			m.input = ""
			m.selected = len(m.storage.Commands) - 1
			m.err = nil
		}

	case copiedMsg:
		if msg.success {
			m.copyMessage = "Copied to clipboard!"
		} else {
			m.copyMessage = fmt.Sprintf("Failed to copy: %v", msg.err)
		}
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("memorycat 🐱"))
	b.WriteString("\n")

	switch m.currentState {
	case stateList:
		if len(m.storage.Commands) == 0 {
			b.WriteString("No commands saved yet. Press 'n' to add a new command.\n")
		} else {
			for i, cmd := range m.storage.Commands {
				cursor := " "
				if i == m.selected {
					cursor = ">"
				}

				commandStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
				descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

				b.WriteString(fmt.Sprintf("%s %s\n", cursor, commandStyle.Render(cmd.Command)))
				b.WriteString(fmt.Sprintf("  %s\n\n", descStyle.Render(cmd.Description)))
			}
		}

		b.WriteString("\n")

		if m.copyMessage != "" {
			copyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
			b.WriteString(copyStyle.Render(m.copyMessage))
			b.WriteString("\n\n")
		}

		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		b.WriteString(helpStyle.Render("n: new command | enter/c: copy | d: delete | ↑/k ↓/j: navigate | q: quit"))

	case stateInput:
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

		b.WriteString("Enter command:\n\n")
		b.WriteString(inputStyle.Render(m.input + "█"))
		b.WriteString("\n\n")

		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		b.WriteString(helpStyle.Render("enter: save | esc: cancel"))

	case stateGenerating:
		b.WriteString("Generating description with Claude AI...\n")

	case stateTemplateInput:
		currentVar := m.templateVars[m.currentVarIndex]

		promptStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

		b.WriteString(promptStyle.Render(fmt.Sprintf("Enter value for: %s", currentVar)))
		b.WriteString(fmt.Sprintf(" (%d/%d)\n\n", m.currentVarIndex+1, len(m.templateVars)))
		b.WriteString(inputStyle.Render(m.input + "█"))
		b.WriteString("\n\n")

		// Show the command template with already filled values
		previewStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		preview := SubstituteTemplateVars(m.templateCommand, m.templateValues)
		b.WriteString(previewStyle.Render(fmt.Sprintf("Preview: %s", preview)))
		b.WriteString("\n\n")

		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		b.WriteString(helpStyle.Render("enter: next | esc: cancel"))

	case stateManualDescription:
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true)

		commandStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

		b.WriteString(errorStyle.Render("AI generation failed."))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("Command: %s\n\n", commandStyle.Render(m.pendingCommand)))
		b.WriteString("Please enter a description manually:\n\n")
		b.WriteString(inputStyle.Render(m.input + "█"))
		b.WriteString("\n\n")

		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		b.WriteString(helpStyle.Render("enter: save | esc: cancel"))
	}

	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		b.WriteString("\n\n")
		b.WriteString(errStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return b.String()
}

func handleStdinCommand() error {
	reader := bufio.NewReader(os.Stdin)
	command, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("no command provided")
	}

	fmt.Printf("Generating description for: %s\n", command)
	description, err := GenerateDescription(command)
	if err != nil {
		return err
	}

	storage, err := LoadCommands()
	if err != nil {
		return err
	}

	storage.Commands = append(storage.Commands, Command{
		Command:     command,
		Description: description,
	})

	if err := SaveCommands(storage); err != nil {
		return err
	}

	fmt.Printf("Saved: %s\n", description)
	return nil
}

func main() {
	// Check if stdin is being piped
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped to stdin
		if err := handleStdinCommand(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run interactive TUI
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

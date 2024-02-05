// spin.go

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// stopSpinner is a global channel to stop the spinner.
var stopSpinner chan struct{}

// StopSpinnerNow is a global channel to force stopping the spinner immediately
var StopSpinnerNow chan struct{}

// SpinCompleteFlag indicates whether the spinner should be considered complete
var SpinCompleteFlag bool

type errMsg error

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error
	stopChan <-chan struct{} // New field for stop channel
}

func initialModel(stopChan <-chan struct{}) model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	return model{spinner: s, stopChan: stopChan}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case errMsg:
		m.err = msg
		return m, nil

	default:
		select {
		case <-m.stopChan: // Check if the stop channel is closed
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	str := fmt.Sprintf("%s Working...", m.spinner.View())
	return str
}

func Spin() {
	stopSpinner = make(chan struct{})
	StopSpinnerNow = make(chan struct{}) // Initialize the channel
	SpinCompleteFlag = false             // Initialize the flag
	p := tea.NewProgram(initialModel(stopSpinner))
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// StopSpinner stops the spinner either after a timeout or when StopSpinnerNow is closed
func StopSpinner() {
	close(StopSpinnerNow)
	select {
	case <-stopSpinner: // spinner stopped by normal means
	case <-time.After(2 * time.Second): // timeout
		close(stopSpinner)
	}
}

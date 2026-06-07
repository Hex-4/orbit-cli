package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type model struct {
	spinner  spinner.Model
	progress progress.Model
	loading  bool
	percent  float64
	quitting bool
	sysInfo  string
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#209FB5"))

	return model{
		spinner:  s,
		progress: progress.New(progress.WithDefaultGradient()),
		loading:  true,
		sysInfo:  getSysInfo(),
	}
}

func getSysInfo() string {
	return fmt.Sprintf("OS: %s | Arch: %s | CPUs: %d", runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		
		if m.loading {
			m.percent += 0.05
			if m.percent >= 1.0 {
				m.loading = false
				m.percent = 1.0
			}
			return m, tea.Batch(cmd, tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
				return spinner.TickMsg{}
			}))
		}
		return m, cmd

	default:
		return m, nil
	}
}

func (m model) View() string {
	if m.quitting {
		return "\n  Orbit CLI: Powering down...\n\n"
	}

	str := fmt.Sprintf("\n  %s Orbit System Status\n\n", titleStyle.Render("●"))
	
	if m.loading {
		str += fmt.Sprintf("  %s Initializing modules... %s\n\n", m.spinner.View(), m.progress.ViewAs(m.percent))
	} else {
		str += fmt.Sprintf("  %s Modules Online\n\n", infoStyle.Render("✓"))
	}

	str += fmt.Sprintf("  %s %s\n", subtleStyle.Render("System:"), m.sysInfo)
	str += fmt.Sprintf("  %s %s\n", subtleStyle.Render("Bramble Hook:"), "Active")
	
	str += "\n  " + subtleStyle.Render("Press 'q' to exit") + "\n"

	return str
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "run" {
		fmt.Println("Orbit Engine starting...")
		time.Sleep(1 * time.Second)
		fmt.Println("Executing task sequence...")
		// Placeholder for automation logic
		return
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running Orbit:", err)
		os.Exit(1)
	}
}

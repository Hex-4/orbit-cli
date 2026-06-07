package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Constants for the scene
const (
	fps          = 30
	starCount    = 50
	tickInterval = time.Second / fps
)

// FrameMsg is sent on every animation frame
type FrameMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return FrameMsg(t)
	})
}

type star struct {
	x, y float64
	z    float64 // depth for parallax
	char string
}

type model struct {
	width, height int
	stars         []star
	startTime     time.Time
	frame         int
}

func initialModel() model {
	stars := make([]star, starCount)
	starChars := []string{".", "*", "⊹", "✧"}
	for i := range stars {
		stars[i] = star{
			x:    math.Mod(float64(i)*7.13, 1.0), // pseudo-random
			y:    math.Mod(float64(i)*3.57, 1.0),
			z:    0.5 + math.Mod(float64(i)*0.91, 1.5),
			char: starChars[i%len(starChars)],
		}
	}

	return model{
		stars:     stars,
		startTime: time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case FrameMsg:
		m.frame++
		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Create a canvas
	canvas := make([][]rune, m.height)
	for i := range canvas {
		canvas[i] = []rune(strings.Repeat(" ", m.width))
	}

	elapsed := time.Since(m.startTime).Seconds()

	// 1. Draw Stars with parallax
	for _, s := range m.stars {
		// Parallax scroll: shift X based on time and depth Z
		xShift := elapsed * 2.0 / s.z
		x := int(s.x*float64(m.width) - xShift)
		x = ((x % m.width) + m.width) % m.width // wrap around
		y := int(s.y * float64(m.height-5))     // stars in upper part
		
		if y >= 0 && y < m.height && x >= 0 && x < m.width {
			canvas[y][x] = []rune(s.char)[0]
		}
	}

	// 2. Draw the Ground (Beach)
	beachColor := lipgloss.Color("#F2E2BA")
	beachStyle := lipgloss.NewStyle().Foreground(beachColor)
	for x := 0; x < m.width; x++ {
		// Slight wave on the beach
		yOff := int(2 * math.Sin(float64(x)*0.1))
		yStart := m.height - 4 + yOff
		for y := yStart; y < m.height; y++ {
			if y >= 0 && y < m.height {
				canvas[y][x] = '▒'
			}
		}
	}

	// 3. Draw the Palm Tree
	// Base position
	treeX := m.width / 2
	treeBaseY := m.height - 4

	// Tree trunk sway math
	swayAmplitude := 3.0
	swayFrequency := 1.5
	sway := swayAmplitude * math.Sin(elapsed*swayFrequency)

	trunkColor := lipgloss.Color("#8B4513")
	trunkStyle := lipgloss.NewStyle().Foreground(trunkColor)
	
	// Draw trunk
	trunkHeight := int(float64(m.height) * 0.4)
	if trunkHeight < 5 { trunkHeight = 5 }
	
	topX, topY := 0, 0
	for i := 0; i < trunkHeight; i++ {
		// Quadratic trunk bend
		progress := float64(i) / float64(trunkHeight)
		currSway := sway * math.Pow(progress, 1.5)
		
		x := treeX + int(currSway)
		y := treeBaseY - i
		
		if y >= 0 && y < m.height && x >= 0 && x < m.width {
			canvas[y][x] = '█'
			topX, topY = x, y
		}
	}

	// Draw Fronds (leaves)
	frondColor := lipgloss.Color("#2E8B57")
	frondStyle := lipgloss.NewStyle().Foreground(frondColor)
	
	numFronds := 8
	frondLength := 12
	for i := 0; i < numFronds; i++ {
		angleBase := (float64(i) / float64(numFronds)) * 2 * math.Pi
		// Leaves sway too, slightly out of phase
		leafSway := 0.2 * math.Cos(elapsed*2.0 + float64(i))
		
		for j := 1; j <= frondLength; j++ {
			dist := float64(j)
			angle := angleBase + leafSway*(dist/float64(frondLength))
			
			lx := topX + int(dist*math.Cos(angle)*2.0) // multiplier for aspect ratio
			ly := topY + int(dist*math.Sin(angle))
			
			if ly >= 0 && ly < m.height && lx >= 0 && lx < m.width {
				canvas[ly][lx] = '🌿'
			}
		}
	}

	// Render canvas to string with colors
	var out strings.Builder
	for y, row := range canvas {
		for x, r := range row {
			char := string(r)
			style := lipgloss.NewStyle()
			
			// Simple color mapping based on character
			if r == '█' {
				style = trunkStyle
			} else if r == '▒' {
				style = beachStyle
			} else if r == '🌿' {
				style = frondStyle
			} else if r == '.' || r == '*' || r == '⊹' || r == '✧' {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFACD"))
			}
			
			out.WriteString(style.Render(char))
		}
		if y < m.height-1 {
			out.WriteRune('\n')
		}
	}

	return out.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

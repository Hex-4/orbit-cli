package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Constants for the scene
const (
	fps          = 30
	starCount    = 60
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

type shootingStar struct {
	active bool
	x, y   float64
	vx, vy float64
	life   float64
}

type model struct {
	width, height int
	stars         []star
	startTime     time.Time
	frame         int

	// Interactive state
	windForce    float64
	scrollSpeed  float64
	showShooter  bool
	shooter      shootingStar
	colorCycle   int
	starTypeIdx  int
}

func initialModel() model {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
	stars := make([]star, starCount)
	starChars := [][]string{
		{".", "*", "⊹", "✧"},
		{"o", "O", "°", "·"},
		{"+", "x", "*", " "},
	}
	
	typeIdx := 0
	for i := range stars {
		stars[i] = star{
			x:    rand.Float64(),
			y:    rand.Float64(),
			z:    0.5 + rand.Float64()*1.5,
			char: starChars[typeIdx][rand.Intn(len(starChars[typeIdx]))],
		}
	}

	return model{
		stars:       stars,
		startTime:   time.Now(),
		windForce:   1.0,
		scrollSpeed: 1.0,
		showShooter: false,
		colorCycle:  0,
		starTypeIdx: 0,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "left":
			m.windForce -= 0.2
		case "right":
			m.windForce += 0.2
		case "up":
			m.scrollSpeed += 0.2
		case "down":
			m.scrollSpeed -= 0.2
			if m.scrollSpeed < 0 { m.scrollSpeed = 0 }
		case "s":
			m.showShooter = !m.showShooter
		case "c":
			m.colorCycle = (m.colorCycle + 1) % 4
			m.starTypeIdx = (m.starTypeIdx + 1) % 3
			// Update star characters
			starChars := [][]string{
				{".", "*", "⊹", "✧"},
				{"o", "O", "°", "·"},
				{"+", "x", "*", " "},
			}
			for i := range m.stars {
				m.stars[i].char = starChars[m.starTypeIdx][rand.Intn(len(starChars[m.starTypeIdx]))]
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case FrameMsg:
		m.frame++
		
		// Update shooting star
		if m.showShooter {
			if !m.shooter.active && rand.Float64() < 0.02 {
				m.shooter.active = true
				m.shooter.x = rand.Float64() * float64(m.width)
				m.shooter.y = rand.Float64() * float64(m.height/2)
				m.shooter.vx = 2.0 + rand.Float64()*3.0
				m.shooter.vy = 0.5 + rand.Float64()
				m.shooter.life = 1.0
			}
			if m.shooter.active {
				m.shooter.x += m.shooter.vx
				m.shooter.y += m.shooter.vy
				m.shooter.life -= 0.05
				if m.shooter.life <= 0 || m.shooter.x > float64(m.width) {
					m.shooter.active = false
				}
			}
		}

		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing the beach..."
	}

	canvas := make([][]rune, m.height)
	for i := range canvas {
		canvas[i] = []rune(strings.Repeat(" ", m.width))
	}

	elapsed := time.Since(m.startTime).Seconds()

	// Colors
	beachColors := []string{"#F2E2BA", "#EEDC82", "#F5DEB3", "#D2B48C"}
	trunkColors := []string{"#8B4513", "#A0522D", "#6B4226", "#5C4033"}
	frondColors := []string{"#2E8B57", "#228B22", "#32CD32", "#006400"}
	
	beachStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(beachColors[m.colorCycle]))
	trunkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(trunkColors[m.colorCycle]))
	frondStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(frondColors[m.colorCycle]))
	starStyle  := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFACD"))

	// 1. Stars
	for _, s := range m.stars {
		xShift := elapsed * 2.0 * m.scrollSpeed / s.z
		x := int(s.x*float64(m.width) - xShift)
		x = ((x % m.width) + m.width) % m.width
		y := int(s.y * float64(m.height-6))
		
		if y >= 0 && y < m.height && x >= 0 && x < m.width {
			canvas[y][x] = []rune(s.char)[0]
		}
	}

	// 2. Shooting Star
	if m.shooter.active {
		sx, sy := int(m.shooter.x), int(m.shooter.y)
		if sy >= 0 && sy < m.height && sx >= 0 && sx < m.width {
			canvas[sy][sx] = '☄'
		}
	}

	// 3. Beach
	for x := 0; x < m.width; x++ {
		yOff := int(1.5 * math.Sin(float64(x)*0.15 + elapsed))
		yStart := m.height - 4 + yOff
		for y := yStart; y < m.height; y++ {
			if y >= 0 && y < m.height {
				canvas[y][x] = '▒'
			}
		}
	}

	// 4. Palm Tree
	treeX := m.width / 2
	treeBaseY := m.height - 4
	
	// Wind/Sway logic
	baseSway := 4.0 * math.Sin(elapsed * 1.2)
	totalSway := baseSway + (m.windForce * 5.0)

	trunkHeight := int(float64(m.height) * 0.45)
	if trunkHeight < 8 { trunkHeight = 8 }
	
	topX, topY := 0, 0
	for i := 0; i < trunkHeight; i++ {
		prog := float64(i) / float64(trunkHeight)
		currSway := totalSway * math.Pow(prog, 1.6)
		
		x := treeX + int(currSway)
		y := treeBaseY - i
		
		if y >= 0 && y < m.height && x >= 0 && x < m.width {
			canvas[y][x] = '█'
			topX, topY = x, y
		}
	}

	// Fronds
	numFronds := 10
	frondLength := 14
	for i := 0; i < numFronds; i++ {
		angleBase := (float64(i) / float64(numFronds)) * 2 * math.Pi
		leafSway := (0.3 * math.Cos(elapsed*2.0 + float64(i))) + (m.windForce * 0.1)
		
		for j := 1; j <= frondLength; j++ {
			dist := float64(j)
			angle := angleBase + leafSway*(dist/float64(frondLength))
			
			lx := topX + int(dist*math.Cos(angle)*2.2)
			ly := topY + int(dist*math.Sin(angle))
			
			if ly >= 0 && ly < m.height && lx >= 0 && lx < m.width {
				if canvas[ly][lx] != '█' {
					canvas[ly][lx] = '🍃'
				}
			}
		}
	}

	// Overlay Instructions
	instructions := fmt.Sprintf(" Wind: %.1f | Scroll: %.1f | Shooter: %v | Press 'c' to cycle colors ", m.windForce, m.scrollSpeed, m.showShooter)
	instrStyle := lipgloss.NewStyle().Background(lipgloss.Color("#3C3C3C")).Foreground(lipgloss.Color("#FFFFFF"))
	
	var out strings.Builder
	for y, row := range canvas {
		for _, r := range row {
			char := string(r)
			style := lipgloss.NewStyle()
			
			switch r {
			case '█': style = trunkStyle
			case '▒': style = beachStyle
			case '🍃': style = frondStyle
			case '☄': style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
			case ' ': style = lipgloss.NewStyle()
			default: style = starStyle
			}
			out.WriteString(style.Render(char))
		}
		if y < m.height-1 {
			out.WriteRune('\n')
		}
	}

	return out.String() + "\n" + instrStyle.Render(instructions)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

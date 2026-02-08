package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/febritecno/stockmap/internal/styles"
)

// ASCII Logo frames for animation
var logoFrames = []string{
	`
   _____ _             _    __  __
  / ____| |           | |  |  \/  |
 | (___ | |_ ___   ___| | _| \  / | __ _ _ __
  \___ \| __/ _ \ / __| |/ / |\/| |/ _' | '_ \
  ____) | || (_) | (__|   <| |  | | (_| | |_) |
 |_____/ \__\___/ \___|_|\_\_|  |_|\__,_| .__/
                                        | |
                                        |_|
`,
	`
   _____ _             _    __  __
  / ____| |           | |  |  \/  |
 | (___ | |_ ___   ___| | _| \  / | __ _ _ __
  \___ \| __/ _ \ / __| |/ / |\/| |/ _' | '_ \
  ____) | || (_) | (__|   <| |  | | (_| | |_) |
 |_____/ \__\___/ \___|_|\_\_|  |_|\__,_| .__/
    ◉                                   | |
                                        |_|
`,
	`
   _____ _             _    __  __
  / ____| |           | |  |  \/  |
 | (___ | |_ ___   ___| | _| \  / | __ _ _ __
  \___ \| __/ _ \ / __| |/ / |\/| |/ _' | '_ \
  ____) | || (_) | (__|   <| |  | | (_| | |_) |
 |_____/ \__\___/ \___|_|\_\_|  |_|\__,_| .__/
    ◉ ◉                                 | |
                                        |_|
`,
	`
   _____ _             _    __  __
  / ____| |           | |  |  \/  |
 | (___ | |_ ___   ___| | _| \  / | __ _ _ __
  \___ \| __/ _ \ / __| |/ / |\/| |/ _' | '_ \
  ____) | || (_) | (__|   <| |  | | (_| | |_) |
 |_____/ \__\___/ \___|_|\_\_|  |_|\__,_| .__/
    ◉ ◉ ◉                               | |
                                        |_|
`,
	`
   _____ _             _    __  __
  / ____| |           | |  |  \/  |
 | (___ | |_ ___   ___| | _| \  / | __ _ _ __
  \___ \| __/ _ \ / __| |/ / |\/| |/ _' | '_ \
  ____) | || (_) | (__|   <| |  | | (_| | |_) |
 |_____/ \__\___/ \___|_|\_\_|  |_|\__,_| .__/
    ◉ ◉ ◉ ◉                             | |
                                        |_|
`,
}

// Splash shows the animated intro screen
type Splash struct {
	width       int
	height      int
	frame       int
	totalFrames int
	done        bool
}

// NewSplash creates a new splash view
func NewSplash() *Splash {
	return &Splash{
		frame:       0,
		totalFrames: 20, // Total animation frames before auto-dismiss
	}
}

// SetSize sets the view dimensions
func (s *Splash) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// NextFrame advances the animation
func (s *Splash) NextFrame() {
	s.frame++
	if s.frame >= s.totalFrames {
		s.done = true
	}
}

// IsDone returns true when animation is complete
func (s *Splash) IsDone() bool {
	return s.done
}

// Skip marks the splash as done
func (s *Splash) Skip() {
	s.done = true
}

// View renders the splash screen
func (s *Splash) View() string {
	var b strings.Builder

	// Calculate vertical centering
	logoHeight := 10
	topPadding := (s.height - logoHeight - 6) / 2
	if topPadding < 0 {
		topPadding = 1
	}

	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}

	// Get current logo frame
	logoIndex := s.frame % len(logoFrames)
	logo := logoFrames[logoIndex]

	// Style the logo with gradient effect based on frame
	var logoStyle lipgloss.Style
	switch s.frame % 4 {
	case 0:
		logoStyle = lipgloss.NewStyle().Foreground(styles.ColorPrimary)
	case 1:
		logoStyle = lipgloss.NewStyle().Foreground(styles.ColorCyan)
	case 2:
		logoStyle = lipgloss.NewStyle().Foreground(styles.ColorHighlight)
	case 3:
		logoStyle = lipgloss.NewStyle().Foreground(styles.ColorPrimary)
	}

	// Center and render logo
	styledLogo := logoStyle.Render(logo)
	for _, line := range strings.Split(styledLogo, "\n") {
		b.WriteString(centerText(line, s.width))
		b.WriteString("\n")
	}

	// Tagline
	tagline := styles.MutedStyle().Render("Stock Screener")
	b.WriteString(centerText(tagline, s.width))
	b.WriteString("\n\n")

	// Version
	version := styles.InfoStyle.Render("v1.0.0")
	b.WriteString(centerText(version, s.width))
	b.WriteString("\n\n")

	// Press any key hint (show after a few frames)
	if s.frame > 5 {
		hint := styles.HelpStyle.Render("Press any key to continue...")
		b.WriteString(centerText(hint, s.width))
	}

	return b.String()
}

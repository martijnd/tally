package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"time"
)

type preferences struct {
	Speed, Theme int
	Bell         bool
}

var speeds = []string{"Standard", "Fast", "Instant"}
var themes = []string{"Terminal", "Paper", "Midnight"}
var nativeDark = lipgloss.HasDarkBackground()

func (m *model) tick() tea.Cmd {
	delay := 170 * time.Millisecond
	if m.prefs.Speed == 1 {
		delay = 30 * time.Millisecond
	}
	return tea.Tick(delay, func(time.Time) tea.Msg { return tickMsg{} })
}
func (m *model) settingsKey(key string) {
	switch key {
	case "s":
		m.prefs.Speed = (m.prefs.Speed + 1) % len(speeds)
	case "t":
		m.prefs.Theme = (m.prefs.Theme + 1) % len(themes)
	case "b":
		m.prefs.Bell = !m.prefs.Bell
	case ",", "esc":
		m.settings = false
	}
}
func (m *model) settingsView() string {
	return panel.Padding(1, 2).Render(fmt.Sprintf("Settings\n\ns  Scoring speed   %s\nt  Colour theme    %s\nb  Big-score bell  %t\n\nesc close", speeds[m.prefs.Speed], themes[m.prefs.Theme], m.prefs.Bell))
}
func (m *model) scoreDone() tea.Cmd {
	m.shown = len(m.events)
	m.finishPlay()
	if m.prefs.Bell && m.gain >= m.target() {
		return func() tea.Msg { fmt.Fprint(os.Stderr, "\a"); return nil }
	}
	return nil
}

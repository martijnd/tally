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
	m.finishPlay()
	if m.prefs.Bell && m.gain >= m.target() {
		return func() tea.Msg { fmt.Fprint(os.Stderr, "\a"); return nil }
	}
	return nil
}
func (m *model) mouse(e tea.MouseMsg) tea.Cmd {
	if e.Action != tea.MouseActionPress {
		return nil
	}
	key := ""
	switch e.Button {
	case tea.MouseButtonWheelUp:
		key = "up"
	case tea.MouseButtonWheelDown:
		key = "down"
	case tea.MouseButtonRight:
		key = "esc"
	case tea.MouseButtonLeft:
		switch {
		case m.settings:
			switch e.Y {
			case 4:
				key = "s"
			case 5:
				key = "t"
			case 6:
				key = "b"
			default:
				key = "esc"
			}
		case m.collection:
			key = "tab"
		case m.viewingDeck:
			key = "tab"
		case m.inventory:
			key = " "
		case m.overview:
			key = "down"
		case len(m.packChoices) > 0:
			if i := e.Y - 3; i >= 0 && i < len(m.packChoices) {
				key = fmt.Sprint(i + 1)
			}
		case m.phase == phaseStart:
			if i := e.Y - 3; i >= 0 && i < len(startingDecks) {
				m.deckID = i
				m.difficulty = min(m.difficulty, m.bestDifficulty[i])
			} else {
				key = "enter"
			}
		case m.phase == phasePlay:
			if i := e.Y - 4 - m.bossHeaderRows(); e.X < leftW+4 && i >= 0 && i < len(m.hand) {
				m.cursor = i
				key = " "
			} else if e.X >= leftW+4 {
				key = "enter"
			}
		case m.phase == phaseShop:
			if i := (e.Y - 5) / 2; e.Y >= 5 && i < len(m.offers) {
				m.shopCursor = i
				key = "enter"
			} else {
				key = "down"
			}
		default:
			key = "enter"
		}
	}
	if key == "" {
		return nil
	}
	var k tea.KeyMsg
	switch key {
	case "up":
		k.Type = tea.KeyUp
	case "down":
		k.Type = tea.KeyDown
	case "enter":
		k.Type = tea.KeyEnter
	case "esc":
		k.Type = tea.KeyEsc
	case "tab":
		k.Type = tea.KeyTab
	default:
		k.Type = tea.KeyRunes
		k.Runes = []rune(key)
	}
	_, cmd := m.Update(k)
	return cmd
}

func (m *model) bossHeaderRows() int {
	if m.boss().Rule != "" {
		return 1
	}
	return 0
}

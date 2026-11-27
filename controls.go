package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type tickMsg struct{}

type packTick struct {
	first *packChoice
}

func (m *model) packTick() tea.Cmd {
	msg := packTick{&m.packChoices[0]}
	delay := 160 * time.Millisecond
	if m.prefs.Speed == 1 {
		delay = 40 * time.Millisecond
	}
	return tea.Tick(delay, func(time.Time) tea.Msg { return msg })
}

func (m *model) advanceScoring() tea.Cmd {
	if m.shown < len(m.events) {
		m.shown++
		return m.tick()
	}
	if m.step < countSteps {
		m.step++
		m.shownScore = addScore(m.baseScore, m.gain/countSteps*m.step+(m.gain%countSteps)*m.step/countSteps)
		return m.tick()
	}
	return m.scoreDone()
}

func (m *model) Init() tea.Cmd { return nil }

// Normalize keyboard shortcuts before recording and dispatch.
func (m *model) resolveKey(msg tea.Msg) tea.Msg {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return msg
	}
	runeKey := func(r rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }
	switch k.String() {
	case "w":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "s":
		if !m.settings && (m.inventory || m.overview || m.phase == phaseStart || m.phase == phaseShop || m.phase == phasePlay) {
			return tea.KeyMsg{Type: tea.KeyDown}
		}
	case "t":
		if !m.settings && m.phase == phasePlay {
			return runeKey('s')
		}
		if !m.settings && m.phase == phaseShop {
			return runeKey('u')
		}
	case "e":
		if m.phase != phaseWon {
			return tea.KeyMsg{Type: tea.KeyEnter}
		}
	case "f":
		return runeKey('i')
	case "g":
		return runeKey('h')
	case "z":
		return runeKey('l')
	case "1", "2":
		if !m.inventory && len(m.packChoices) == 0 {
			left := k.String() == "1"
			if m.collection || m.phase == phaseStart {
				if left {
					return tea.KeyMsg{Type: tea.KeyLeft}
				}
				return tea.KeyMsg{Type: tea.KeyRight}
			}
			if left {
				return runeKey('[')
			}
			return runeKey(']')
		}

	}
	return k
}
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.updateResolved(m.resolveKey(msg))
}
func (m *model) updateResolved(msg tea.Msg) (tea.Model, tea.Cmd) {
	defer func() {
		if m.restriction("cash cap") {
			m.money = min(8, m.money)
		}
	}()
	switch msg := msg.(type) {
	case packTick:
		if m.packFrame > 0 && len(m.packChoices) > 0 && msg.first == &m.packChoices[0] {
			m.packFrame++
			if m.packFrame > len(celestialFrames) {
				m.packFrame = 0
			} else {
				return m, m.packTick()
			}
		}
	case tickMsg:
		if m.phase == phaseScoring {
			return m, m.advanceScoring()
		}
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if m.settings {
			m.settingsKey(key)
			return m, nil
		}
		if key == "," && m.phase != phaseScoring {
			m.settings = true
			return m, nil
		}
		if m.phase == phaseScoring && (key == "enter" || key == " ") {
			return m, m.scoreDone()
		}
		if m.collection {
			m.collectionKey(key)
			return m, nil
		}
		if key == "l" && m.phase != phaseScoring {
			m.collection = true
			return m, nil
		}
		if len(m.packChoices) > 0 {
			m.packFrame = 0
			m.pickPack(key)
			return m, nil
		}
		if m.viewingDeck {
			switch key {
			case "v", "esc":
				m.viewingDeck = false
			case "tab":
				m.remainingDeck = !m.remainingDeck
			}
			return m, nil
		}
		if key == "v" && !m.inventory && !m.overview && m.phase != phaseScoring {
			m.viewingDeck = true
			return m, nil
		}
		if m.inventory {
			m.inventoryKey(key)
			return m, nil
		}
		if key == "i" && !m.overview && (m.phase == phasePlay || m.phase == phaseShop) {
			m.inventory = true
			return m, nil
		}
		if m.phase == phaseStart {
			m.startKey(key)
			return m, nil
		}
		if key == "h" && m.phase != phaseScoring {
			m.overview = !m.overview
			return m, nil
		}
		if m.overview {
			return m, nil
		}
		switch m.phase {
		case phasePreview:
			if key == "enter" {
				m.startRound()
			} else if key == "s" && m.round%3 != 2 && !m.restriction("no skips") {
				m.earnTag()
				m.skipped++
				m.round++
			}
		case phasePlay:
			return m, m.updatePlay(key)
		case phaseRoundWon:
			if key == "enter" || key == " " {
				if m.round == numRounds-1 {
					m.phase = phaseWon
					m.bestDifficulty[m.deckID] = max(m.bestDifficulty[m.deckID], m.difficulty+1)
				} else {
					m.openShop()
				}
			}
		case phaseShop:
			m.updateShop(key)
			if len(m.packChoices) > 0 && (m.packKind == 0 || m.packKind == 2) && m.prefs.Speed != 2 {
				m.packFrame = 1
				return m, m.packTick()
			}
		case phaseLost, phaseWon:
			if key == "e" && m.phase == phaseWon {
				m.openShop()
			}
			if key == "n" {
				record := m.bestDifficulty
				*m = *newModel(m.seed)
				m.bestDifficulty = record
			}
		}
	}
	return m, nil
}

func (m *model) updatePlay(key string) tea.Cmd {
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.hand)-1 {
			m.cursor++
		}
	case " ", "x":
		if len(m.hand) == 0 {
			return nil
		}
		if !m.selected[m.cursor] && len(m.selectedIdx()) >= maxSelect {
			m.msg = fmt.Sprintf("you can select up to %d", maxSelect)
			return nil
		}
		if m.boss().Rule == "bell" && m.hand[m.cursor].ID == m.forcedID {
			return nil
		}
		m.selected[m.cursor] = !m.selected[m.cursor]
		m.msg = ""
	case "a":
		if m.play() {
			if m.prefs.Speed == 2 {
				return m.scoreDone()
			}
			return m.tick()
		}
		return nil
	case "d":
		m.discard()
	case "[", "]":
		dest := m.cursor - 1
		if key == "]" {
			dest = m.cursor + 1
		}
		if dest >= 0 && dest < len(m.hand) {
			m.hand[m.cursor], m.hand[dest] = m.hand[dest], m.hand[m.cursor]
			m.selected[m.cursor], m.selected[dest] = m.selected[dest], m.selected[m.cursor]
			m.cursor = dest
		}
	case "s":
		m.bySuit = !m.bySuit
		m.sortHand()
	case "c":
		for i := range m.selected {
			m.selected[i] = m.boss().Rule == "bell" && m.hand[i].ID == m.forcedID
		}
	}
	return nil
}

func (m *model) updateShop(key string) {
	switch key {
	case "[", "]":
		i := m.shopCursor - len(m.offers)
		j := i - 1
		if key == "]" {
			j = i + 1
		}
		if i >= 0 && i < len(m.mods) && j >= 0 && j < len(m.mods) {
			m.mods[i], m.mods[j] = m.mods[j], m.mods[i]
			m.shopCursor += j - i
		}
	case "u":
		m.buyUpgrade()
	case "b":
		m.openPack()
	case "up", "k":
		if m.shopCursor > 0 {
			m.shopCursor--
		}
	case "down", "j":
		if m.shopCursor < m.shopRowCount()-1 {
			m.shopCursor++
		}
	case "enter", " ":
		if m.shopRowCount() > 0 {
			m.shopSelect()
		}
	case "r":
		if m.restriction("no rerolls") {
			m.msg = "Challenge forbids rerolls"
			return
		}
		if m.money < m.rerollCharge() {
			m.msg = "not enough money to reroll"
			return
		}
		m.applyTags("reroll")
		m.money -= m.rerollCharge()
		if m.freeRerolls > 0 {
			m.freeRerolls--
		}
		m.emit("reroll")
		m.rerollPrice++
		m.rollOffers()
		m.clampShopCursor()
		m.msg = "rerolled"
	case "c", "n":
		m.round++
		m.bossDisabled = false
		m.phase = phasePreview
	}
}

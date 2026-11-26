package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"testing"
)

func TestTerminalControls(t *testing.T) {
	m := newModel("terminal")
	key := func(s string) { m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}) }
	key(",")
	key("s")
	key("s")
	key("b")
	key("t")
	if !strings.Contains(m.View(), "Instant") || !m.prefs.Bell || m.prefs.Theme != 1 {
		t.Fatal(m.View())
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m.chooseStart()
	m.startRound()
	m.Update(tea.MouseMsg{X: 3, Y: 4, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if !m.selected[0] {
		t.Fatal("mouse did not select first card")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.phase == phaseScoring || m.score == 0 {
		t.Fatal("instant score")
	}
	m.prefs.Speed = 0
	m.selected[0] = true
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.phase == phaseScoring {
		t.Fatal("manual skip")
	}
	if commas(1200000000000000) != "1.2e15" {
		t.Fatal("endless format")
	}
}
func TestContentEntryPoints(t *testing.T) {
	for i := range startingDecks {
		m := newModel("starts")
		m.unlockRound = 24
		m.discoveries = 100
		m.bestDifficulty[0] = 8
		for j := 0; j < 5; j++ {
			m.bestDifficulty[j] = 8
		}
		m.deckID = i
		m.startKey("enter")
		m.startRound()
		if len(m.stock) == 0 || len(m.hand) == 0 {
			t.Fatal(i)
		}
	}
	for i, c := range challenges {
		m := newModel("challenge")
		m.challengeID = i + 1
		m.chooseStart()
		if len(m.mods) != 1 || m.mods[0].Name != c.Mod || !m.restriction(c.Rule) {
			t.Fatal(c)
		}
	}
	for _, b := range append(append([]boss{}, regularBosses...), finalBosses...) {
		m := newModel("encounter")
		m.round = 2
		m.bosses[0] = b
		m.startRound()
		if !strings.Contains(m.View(), b.Effect) {
			t.Fatal(b.Name)
		}
		m.selected[0] = true
		m.play()
		m.finishPlay()
	}
	for _, b := range bargains {
		m := newModel("trade")
		m.startRound()
		m.selected[0] = true
		m.startModifier("Ember")
		if !m.strikeBargain(b) {
			t.Fatal(b.Name)
		}
	}
	for _, it := range supplies() {

		m := newModel("supply")
		m.round = 2
		m.startRound()
		m.selected[0] = true
		m.lastSupply = supplies()[1]
		m.startModifier("Ember")
		if it.Action == "death" {
			m.selected[1] = true
		}
		m.bag = []supply{it}
		m.useSupply(0)
		if len(m.bag) == 1 && m.bag[0] == it {
			t.Fatal(it.Name)
		}
	}
}
func TestSecretHands(t *testing.T) {
	for _, tt := range []struct {
		r, s []int
		want HandType
	}{{[]int{2, 2, 2, 2, 2}, []int{0, 1, 2, 3, 0}, FiveKind}, {[]int{2, 2, 2, 3, 3}, []int{0, 0, 0, 0, 0}, FlushHouse}, {[]int{2, 2, 2, 2, 2}, []int{1, 1, 1, 1, 1}, FlushFive}} {
		var cards []Card
		for i, r := range tt.r {
			cards = append(cards, Card{Rank: r, Suit: tt.s[i]})
		}
		h, scored := evaluate(cards)
		if h != tt.want || len(scored) != 5 {
			t.Fatal(h, scored)
		}
	}
}

func TestFirstRunScoreExplanation(t *testing.T) {
	m := newModel("first-run")
	m.chooseStart()
	m.startRound()
	m.simulateSelection()
	if preview := m.viewScore(); !strings.Contains(preview, "20 base + 28 card chips") || !strings.Contains(preview, "48 × 2 = 96 before effects") {
		t.Fatal(preview)
	}
	m.play()
	m.scoreDone()
	if result := m.viewScore(); !strings.Contains(result, "= 96") {
		t.Fatal(result)
	}
}

func TestPlayShortcutCase(t *testing.T) {
	m := newModel("shortcut")
	m.chooseStart()
	m.startRound()
	m.selected[0] = true
	key := func(r rune) { m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}) }
	key('l')
	if !m.collection || m.phase != phasePlay {
		t.Fatal("l should open the collection")
	}
	key('l')
	key('p')
	if m.phase != phasePlay || m.plays != playsPerRound || !m.selected[0] {
		t.Fatal("p should leave the selected hand untouched")
	}
	key('a')
	if m.phase != phaseScoring || m.plays != playsPerRound-1 {
		t.Fatal("lowercase a should play the selection")
	}
}

func TestLeftHandControls(t *testing.T) {
	m := newModel("left-hand")
	key := func(r rune) { m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}) }
	key('e')
	key('e')
	m.cursor = 0
	key('s')
	key('x')
	if m.cursor != 1 || !m.selected[1] {
		t.Fatal("move or select failed")
	}
	key('1')
	if m.cursor != 0 || !m.selected[0] {
		t.Fatal("shift lost selection")
	}
	key('t')
	if !m.bySuit {
		t.Fatal("sort failed")
	}
	key('z')
	key('z')
	key('f')
	key('f')
	key('g')
	key('g')
	key('a')
	if m.phase != phaseScoring {
		t.Fatal("play shortcut failed")
	}
}

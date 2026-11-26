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
		if it.Action == "bargain" {
			continue
		}
		m := newModel("supply")
		m.round = 2
		m.startRound()
		m.selected[0] = true
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

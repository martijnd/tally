package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"reflect"
	"testing"
)

func TestSeededRun(t *testing.T) {
	a, b := newModel("orchard"), newModel("orchard")
	a.rollOffers()
	a.startRound()
	b.startRound()
	if !reflect.DeepEqual(a.hand, b.hand) || !reflect.DeepEqual(a.bosses, b.bosses) {
		t.Fatal("independent seeded streams diverged")
	}
	seen := map[string]bool{}
	for stage, effect := range a.bosses {
		if seen[effect.Name] {
			t.Fatalf("repeated boss at stage %d", stage+1)
		}
		seen[effect.Name] = true
		final := false
		for _, f := range finalBosses {
			final = final || f == effect
		}
		if final != (stage == 7) {
			t.Fatal("final boss at wrong stage")
		}
	}
}
func TestEndlessTransition(t *testing.T) {
	m := newModel("endless")
	m.round = numRounds - 1
	m.winRound()
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.phase != phaseWon {
		t.Fatal("missing victory")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if m.phase != phaseShop {
		t.Fatal("missing post-victory shop")
	}
	m.updateShop("c")
	if m.phase != phasePreview || m.round != 24 || m.target() != 24000 {
		t.Fatal("incorrect endless opening")
	}
	for i := 24; i < 300; i++ {
		m.round = i
		if m.target() <= 0 {
			t.Fatal("target overflow")
		}
	}
}

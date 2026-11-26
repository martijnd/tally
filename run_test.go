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
	if m.phase != phasePreview || m.round != 24 || m.target() != 110000 {
		t.Fatal("incorrect endless opening")
	}
	for i := 24; i < 300; i++ {
		m.round = i
		if m.target() <= 0 {
			t.Fatal("target overflow")
		}
	}
}

func TestStakeCurvesAndDeckRules(t *testing.T) {
	for _, tc := range []struct{ stake, stage, base int }{{0, 1, 300}, {0, 8, 50000}, {2, 8, 100000}, {5, 8, 200000}, {0, 9, 110000}} {
		m := newModel("curve")
		m.round = 3 * (tc.stage - 1)
		m.difficulty = tc.stake
		if m.target() != tc.base {
			t.Fatal(tc, m.target())
		}
	}
	m := newModel("plasma")
	m.deckID = 13
	m.chooseStart()
	if m.target() != 600 {
		t.Fatal(m.target())
	}
	_, score := scoreHand([]Card{{Rank: 3}}, []int{0}, HighCard, nil, ScoreCtx{Plasma: true})
	if score != 16 {
		t.Fatal(score)
	}
	for _, i := range []int{5, 6, 7, 8, 9, 11, 14} {
		m := newModel("decks")
		m.deckID = i
		m.chooseStart()
		switch i {
		case 5:
			if m.bagSlots() != 3 || len(m.bag) != 2 || m.bag[1].Name != "The Fool" {
				t.Fatal("Magic")
			}
		case 6:
			m.timesPlayed[Pair] = 4
			m.openShop()
			m.packKind = 2
			m.money = 20
			m.openPack()
			if m.packChoices[0].item.Name != "Mercury" {
				t.Fatal("Telescope")
			}
		case 7:
			if m.bag[0].Name != "Hex" {
				t.Fatal("Ghost")
			}
		case 8:
			if len(m.stock) != 40 {
				t.Fatal("Abandoned")
			}
		case 9:
			for _, c := range m.stock {
				if c.Suit > 1 {
					t.Fatal("Checkered")
				}
			}
		case 11:
			if m.drawLimit() != 10 || m.modSlots() != 4 {
				t.Fatal("Painted")
			}
		case 14:
			if reflect.DeepEqual(m.stock, newDeck()) {
				t.Fatal("Erratic")
			}
		}
	}
}

func TestShowdownBossesAndEligibility(t *testing.T) {
	m := newModel("boss rotation")
	for stage := 0; stage < 32; stage++ {
		b := m.stageBoss(stage)
		final := false
		for _, f := range finalBosses {
			if f == b {
				final = true
			}
		}
		if final != (stage%8 == 7) {
			t.Fatal(stage, b)
		}
	}
	for _, tt := range []struct {
		name                    string
		target, plays, discards int
	}{{"The Wall", 1200, 4, 3}, {"The Needle", 300, 1, 3}, {"The Water", 600, 4, 0}, {"Violet Vessel", 1800, 4, 3}} {
		m := newModel(tt.name)
		m.round = 2
		for _, b := range append(append([]boss{}, regularBosses...), finalBosses...) {
			if b.Name == tt.name {
				m.bosses[0] = b
			}
		}
		m.startRound()
		if m.target() != tt.target || m.plays != tt.plays || m.discards != tt.discards {
			t.Fatal(tt.name)
		}
	}
}

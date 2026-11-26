package main

import "testing"

func namedSupply(t *testing.T, name string) supply {
	t.Helper()
	for _, s := range supplies() {
		if s.Name == name {
			return s
		}
	}
	t.Fatal(name)
	return supply{}
}
func TestDeathCryptidAndFool(t *testing.T) {
	m := newModel("arcana")
	m.startRound()
	m.selected[0] = true
	m.bag = []supply{namedSupply(t, "Death")}
	m.useSupply(0)
	if len(m.bag) != 1 {
		t.Fatal("Death consumed without two targets")
	}
	m.selected[1] = true
	m.hand[1].Seal = 3
	id := m.hand[0].ID
	want := m.hand[1]
	want.ID = id
	m.useSupply(0)
	if m.hand[0] != want {
		t.Fatal("Death must copy right to left")
	}
	m.bag = []supply{namedSupply(t, "The Fool")}
	m.useSupply(0)
	if len(m.bag) != 1 || m.bag[0].Name != "Death" {
		t.Fatal("Fool history")
	}
	m.selected[1] = false
	m.bag = []supply{namedSupply(t, "Cryptid")}
	size := len(m.hand)
	m.useSupply(0)
	if len(m.stock) != 54 || len(m.hand) != size+2 || m.hand[size].ID == id || m.hand[size].Seal != 3 {
		t.Fatal("Cryptid copies")
	}
}
func TestFractionalScoringAndPlanet(t *testing.T) {
	hand := []Card{{Rank: 2, Edition: 3}}
	_, n := scoreHand(hand, []int{0}, HighCard, nil, ScoreCtx{Training: 1, HeldPatient: 1})
	if n != 76 {
		t.Fatalf("Pluto, Polychrome and Steel: got %d, want floor(17 × 2 × 1.5 × 1.5)", n)
	}
	m := newModel("spectral")
	m.startRound()
	m.startModifier("Ember")
	m.startModifier("Momentum")
	if !m.applySupply(namedSupply(t, "Ectoplasm")) || !m.applySupply(namedSupply(t, "Ectoplasm")) || m.handBonus != -3 {
		t.Fatal("Ectoplasm stacking")
	}
}

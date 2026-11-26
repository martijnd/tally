package main

import "testing"

func TestDeckEditsSurviveRound(t *testing.T) {
	m := newModel("archive")
	m.startRound()
	m.selected[0] = true
	id := m.hand[0].ID
	m.editCards(supply{Action: "enhance", Value: 4})
	m.editCards(supply{Action: "copy"})
	if len(m.stock) != 53 || m.stock[52].ID == id || m.stock[52].Enhancement != 4 {
		t.Fatal("invalid duplicate")
	}
	m.editCards(supply{Action: "destroy"})
	m.startRound()
	count := 0
	for _, c := range append(append([]Card{}, m.hand...), m.deck...) {
		if c.ID == id {
			t.Fatal("destroyed card returned")
		}
		if c.Enhancement == 4 {
			count++
		}
	}
	if count != 1 || len(m.deck)+len(m.hand) != 52 {
		t.Fatal("edited deck failed round trip")
	}
}
func TestBlankAndEcho(t *testing.T) {
	cards := []Card{{Rank: 3}, {Rank: 3}, {Rank: 3, Enhancement: 7, Seal: 1}}
	kind, scored := evaluate(cards)
	if kind != Pair || len(scored) != 3 {
		t.Fatal(kind, scored)
	}
	_, score := scoreHand(cards, scored, kind, nil, ScoreCtx{})
	if score != (10+6+100)*2 {
		t.Fatal(score)
	}
}

func TestBargainRequiresPaymentAndTarget(t *testing.T) {
	m := newModel("press")
	m.startRound()
	item := bargain{Action: "edition", Value: 4, Cost: "modifier"}
	m.selected[0] = true
	if m.strikeBargain(item) || m.hand[0].Edition != 0 {
		t.Fatal("unpaid edition")
	}
	m.mods = []*Modifier{{Name: "test"}}
	m.selected[0] = false
	if m.strikeBargain(item) || len(m.mods) != 1 {
		t.Fatal("paid without a target")
	}
	m.selected[0] = true
	if !m.strikeBargain(item) || len(m.mods) != 0 || len(m.hand) != 9 {
		t.Fatal("spacious edition did not grant space")
	}
}

func TestConsumableNames(t *testing.T) {
	names := map[string]bool{}
	families := map[string]int{}
	for _, it := range supplies() {
		if it.Name == "" || names[it.Name] {
			t.Fatalf("invalid consumable name %q", it.Name)
		}
		names[it.Name] = true
		families[it.cardFamily()]++
		if it.Action == "level" && HandType(it.Value) == Pair && it.Name != "Mercury" {
			t.Fatal(it)
		}
	}
	if families["Planet"] != 12 || families["Spectral"] != 18 || families["Tarot"] != 22 {
		t.Fatal(families)
	}
}

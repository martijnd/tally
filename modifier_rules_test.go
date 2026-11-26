package main

import "testing"

func TestCatalogCopyRetriggerAndOrder(t *testing.T) {
	all := catalog()
	names := map[string]bool{}
	for _, mod := range all {
		if names[mod.Name] || mod.Name == "" {
			t.Fatal("duplicate", mod.Name)
		}
		names[mod.Name] = true
	}
	if len(all) != 150 {
		t.Fatal(len(all))
	}
	m := newModel("effects")
	for _, name := range []string{"Counterproof", "Fork Resonator", "Ember"} {
		m.startModifier(name)
	}
	hits := 0
	_, gain := scoreHand([]Card{{Rank: 2}}, []int{0}, HighCard, m.mods, ScoreCtx{OnScore: func(Card) { hits++ }})
	if hits != 3 || gain != 55 {
		t.Fatal(hits, gain)
	}
	m.shopCursor = 0
	m.updateShop("]")
	hits = 0
	_, gain = scoreHand([]Card{{Rank: 2}}, []int{0}, HighCard, m.mods, ScoreCtx{OnScore: func(Card) { hits++ }})
	if hits != 2 || gain != 81 {
		t.Fatal(hits, gain)
	}
}
func TestLifecycleDispatch(t *testing.T) {
	m := newModel("events")
	for _, name := range []string{"Opening Bond", "Appendix", "Accession Gauge", "Offcut Till", "Paper Mill", "Erasure Gauge", "Loaded Die"} {
		m.startModifier(name)
	}
	m.startRound()
	if m.mods[0].ResaleGrowth != 1 || m.odds() != 2 {
		t.Fatal("start and odds")
	}
	m.selected[0] = true
	m.discard()
	if len(m.stock) != 51 || m.money != 5 || m.mods[5].Counter != 1 {
		t.Fatal("discard event chain")
	}
	m.winRound()
	if len(m.stock) != 52 || m.mods[2].Counter != 1 {
		t.Fatal("end event chain")
	}
}

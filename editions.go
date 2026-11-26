package main

var editions = []string{"standard", "Foil", "Holographic", "Polychrome", "Negative"}

func editionScore(edition, chips int, mult float64) (int, float64) {
	switch edition {
	case 1:
		chips += 50
	case 2:
		mult += 10
	case 3:

		mult *= 1.5
	}
	return chips, mult
}
func (m *model) modSlots() int {
	if m.restriction("single slot") {
		return 1
	}
	n := maxMods + m.slotBonus + m.upgrades[cabinet]
	for _, mod := range m.mods {
		if mod.Edition == 4 {
			n++
		}
	}
	return n
}
func (m *model) drawLimit() int {
	penalty := 0
	if m.restriction("small hand") {
		penalty = 3
	}
	limit := handSize - penalty + m.handBonus + m.upgrades[folio]
	for _, c := range m.hand {
		if c.Edition == 4 {
			limit++
		}
	}
	if b := m.boss(); b.Rule == "size" {
		limit -= b.Amount
	}
	return limit
}

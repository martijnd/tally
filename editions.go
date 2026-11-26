package main

var editions = []string{"standard", "Etched", "Flared", "Mirrored", "Spacious"}

func editionScore(edition, chips, mult int) (int, int) {
	switch edition {
	case 1:
		chips += 40
	case 2:
		mult += 6
	case 3:
		mult = multiplyScore(mult, 2)
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

type bargain struct {
	Name, Desc string
	Action     string
	Value      int
	Cost       string
}

var bargains = []bargain{
	{"Deep Pocket", "Gain one hand space", "space", 1, "wallet"},
	{"Broad Shelf", "Gain one Joker slot", "slot", 1, "wallet"},
	{"Lost Receipt", "Gain $18", "cash", 18, "card"},
	{"Missing Page", "Train all hands once", "train", 1, "card"},
	{"Empty Purse", "Train all hands twice", "train", 2, "wallet"},
	{"Bent Frame", "Gain a Joker slot", "slot", 1, "card"},
	{"Loose Leaf", "Gain one hand space", "space", 1, "card"},
	{"Silent Auction", "Gain $24", "cash", 24, "modifier"},
	{"Last Lecture", "Train all hands twice", "train", 2, "modifier"},
	{"Open Vault", "Gain two hand spaces", "space", 2, "modifier"},
	{"Wide Margin", "Gain two Joker slots", "slot", 2, "modifier"},
	{"Copper Bath", "Etch selected cards", "edition", 1, "wallet"},
	{"Hot Press", "Flare selected cards", "edition", 2, "wallet"},
	{"Mirror Bath", "Mirror selected cards", "edition", 3, "wallet"},
	{"Wide Press", "Make selected cards spacious", "edition", 4, "wallet"},
	{"Deep Cut", "Mirror selected cards", "edition", 3, "modifier"},
	{"Lost Mould", "Make selected cards spacious", "edition", 4, "modifier"},
	{"Last Casting", "Flare selected cards", "edition", 2, "modifier"},
}

func (m *model) strikeBargain(b bargain) bool {
	switch b.Cost {
	case "wallet":
		if m.money == 0 {
			m.msg = "Requires money to sacrifice"
			return false
		}
	case "card":
		if len(m.stock) < 2 {
			m.msg = "Requires a spare card"
			return false
		}
	case "modifier":
		if len(m.mods) == 0 {
			m.msg = "Requires an owned Joker"
			return false
		}
	}
	switch b.Action {
	case "edition":
		if !m.editCards(supply{Action: "edition", Value: b.Value}) {
			return false
		}
	case "slot":
		m.slotBonus += b.Value
	case "space":
		m.handBonus += b.Value
	case "train":
		for i := range m.levels {
			m.levels[i] += b.Value
		}
	case "cash":
		m.money += b.Value
	}
	switch b.Cost {
	case "wallet":
		m.money = 0
	case "modifier":
		i := m.itemRNG.IntN(len(m.mods))
		m.mods = append(m.mods[:i], m.mods[i+1:]...)
	case "card":
		i := m.itemRNG.IntN(len(m.stock))
		id := m.stock[i].ID
		m.stock = append(m.stock[:i], m.stock[i+1:]...)
		m.emit("destroyed")
		for i, c := range m.hand {
			if c.ID == id {
				m.hand = append(m.hand[:i], m.hand[i+1:]...)
				m.selected = append(m.selected[:i], m.selected[i+1:]...)
				m.clampCursor()
				break
			}
		}
		for i, c := range m.deck {
			if c.ID == id {
				m.deck = append(m.deck[:i], m.deck[i+1:]...)
				break
			}
		}
	}
	m.draw()
	return true
}

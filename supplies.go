package main

import (
	"fmt"
	"slices"
	"strings"
)

type supply struct {
	Name, Description, Action string
	Value                     int
}

func oldSupplies() []supply {
	out := []supply{
		{"Tracing Paper", "Copy selected cards into the deck", "copy", 0},
		{"Margin Knife", "Remove selected cards from the deck", "destroy", 0},
		{"Ascending Ink", "Raise selected ranks, wrapping king to ace", "rank", 1},
		{"Descending Ink", "Lower selected ranks", "rank", -1},
		{"Bridge Stamp", "Advance selected ranks by three", "rank", 3},
		{"Overprint", "Gain $5", "money", 5},
		{"Proof Slip", "Create a random Planet card", "lesson", 0},
		{"Quiet Writ", "Disable the active boss for this round", "silence", 0},
	}
	for i, name := range []string{"Spade Plate", "Heart Plate", "Diamond Plate", "Club Plate"} {
		out = append(out, supply{name, "Set selected cards to " + name[:len(name)-6], "suit", i})
	}
	for i, name := range []string{"Weighted", "Loud", "Patient", "Prismatic", "Brittle", "Dividend", "Blank"} {
		out = append(out, supply{name + " Ink", "Apply " + name + " enhancement to selected cards", "enhance", i + 1})
	}
	for i, name := range []string{"Echo", "Mint", "Archive", "Scrap"} {
		out = append(out, supply{name + " Press", "Apply " + name + " seal to selected cards", "seal", i + 1})
	}
	for i, b := range bargains {
		out = append(out, supply{b.Name, b.Desc + " · sacrifice " + map[string]string{"wallet": "all money", "card": "a random playing card", "modifier": "a random Joker"}[b.Cost], "bargain", i})
	}
	for i, b := range handBases {
		out = append(out, supply{"Study of " + b.Name, "Raise " + b.Name + " one level", "level", i})
	}
	for i := range out {
		out[i].Name = consumableNames[out[i].Name]
	}
	return out
}

var enhancements = []string{"plain", "Bonus", "Mult", "Steel", "Wild", "Glass", "Gold", "Stone", "Lucky"}
var seals = []string{"none", "Red", "Gold", "Blue", "Purple"}

func (m *model) editCards(item supply) bool {
	indices := m.selectedIdx()
	if len(indices) == 0 {
		m.msg = "Select target cards first"
		return false
	}
	if item.Action == "destroy" && len(m.stock) <= len(indices) {
		m.msg = "Keep at least one card"
		return false
	}
	for _, i := range indices {
		c := m.hand[i]
		pos := slices.IndexFunc(m.stock, func(orig Card) bool { return orig.ID == c.ID })
		if pos < 0 {
			continue
		}
		switch item.Action {
		case "copy":
			m.nextID++
			c.ID = m.nextID
			m.stock = append(m.stock, c)
			m.emit("added")
			m.deck = append(m.deck, c)
			continue
		case "destroy":
			m.stock = slices.Delete(m.stock, pos, pos+1)
			m.emit("destroyed")
			continue
		case "rank":
			c.Rank = (c.Rank-1+item.Value+13)%13 + 1
		case "suit":
			c.Suit = item.Value
		case "edition":
			c.Edition = item.Value
		case "enhance":
			c.Enhancement = item.Value
		case "seal":
			c.Seal = item.Value
		}
		m.stock[pos] = c
		m.hand[i] = c
	}
	if item.Action == "destroy" {
		m.removeSelected()
		m.draw()
	}
	return true
}
func (m *model) useSupply(i int) {
	if i >= len(m.bag) {
		return
	}
	item := m.bag[i]
	if !m.legacyEffects {
		if m.applySupply(item) {
			m.bag = slices.Delete(m.bag, i, i+1)
			m.msg = "Used " + item.Name
			if len(m.hand) == 0 && m.phase == phasePlay {
				m.draw()
				if len(m.hand) == 0 {
					m.phase = phaseLost
					m.inventory = false
				}
			}
		}
		return
	}
	switch item.Action {
	case "bargain":
		if !m.strikeBargain(bargains[item.Value]) {
			return
		}
	case "money":
		m.money += item.Value
	case "level":
		m.levels[item.Value]++
	case "silence":
		if m.boss().Rule == "" || m.phase != phasePlay {
			m.msg = "Use during an active boss round"
			return
		}
		b := m.boss()
		m.bossDisabled = true
		switch b.Rule {
		case "size":
			m.draw()
		case "plays", "finalmute":
			m.plays++
		case "discards":
			m.discards += b.Amount
		}
	case "lesson":
		m.bag[i] = m.lesson()
		m.msg = "Created " + m.bag[i].Name
		return
	default:
		if !m.editCards(item) {
			return
		}
	}
	m.bag = slices.Delete(m.bag, i, i+1)
	m.msg = "Used " + item.Name
}
func (m *model) lesson() supply {
	if !m.legacyEffects {
		return m.randomSupply("Planet", "")
	}
	items := oldSupplies()
	return items[len(items)-len(handBases)+m.itemRNG.IntN(len(handBases))]
}
func (m *model) inventoryKey(key string) {
	switch key {
	case "i", "esc":
		m.inventory = false
	case "1", "2", "3", "4", "5":
		m.useSupply(int(key[0] - '1'))
	case "up", "down", " ":
		m.updatePlay(key)
	case "b":
		if m.phase == phaseShop && m.stockItem != nil && len(m.bag) < m.bagSlots() && m.money >= 3 {
			m.money -= 3
			m.bag = append(m.bag, *m.stockItem)
			m.stockItem = nil
		}
	}
}
func (m *model) inventoryView() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Consumables %d/%d · $%d\n", len(m.bag), m.bagSlots(), m.money)
	for i, it := range m.bag {
		fmt.Fprintf(&b, "%d  %s · %s card · %s\n", i+1, it.Name, it.cardFamily(), it.Description)
	}
	if m.phase == phaseShop && m.stockItem != nil {
		fmt.Fprintf(&b, "b buy %s ($3) · %s card · %s\n", m.stockItem.Name, m.stockItem.cardFamily(), m.stockItem.Description)
	}
	for i, c := range m.hand {
		marker := " "
		if i == m.cursor {
			marker = ">"
		}
		label := "face down"
		if m.boss().Rule != "blind" {
			label = fmt.Sprintf("%s %s · %s · %s edition · %s seal", c.Label(), suit(c), enhancements[c.Enhancement], editions[c.Edition], seals[c.Seal])
		}
		fmt.Fprintf(&b, "%s [%s] %s\n", marker, map[bool]string{true: "x", false: " "}[m.selected[i]], label)
	}
	fmt.Fprintf(&b, "w/s target · space select · number use · f close\n%s", m.msg)
	return panel.Padding(1, 2).Render(b.String())
}

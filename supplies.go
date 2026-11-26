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

var enhancements = []string{"plain", "Bonus", "Mult", "Steel", "Wild", "Glass", "Gold", "Stone", "Lucky"}
var seals = []string{"none", "Red", "Gold", "Blue", "Purple"}

func (m *model) useSupply(i int) {
	if i >= len(m.bag) {
		return
	}
	item := m.bag[i]

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
}
func (m *model) inventoryKey(key string) {
	switch key {
	case "i", "esc":
		m.inventory = false
		m.targetingSlot = 0
	case "1", "2", "3", "4", "5":
		slot := int(key[0] - '0')
		if slot > len(m.bag) {
			return
		}
		m.targetingSlot = 0
		lo, hi := m.bag[slot-1].targetCount()
		n := len(m.selectedIdx())
		if len(m.hand) > 0 && lo > 0 && (n < lo || n > hi) {
			m.targetingSlot = slot
			m.msg = ""
			return
		}
		m.useSupply(slot - 1)
	case "enter":
		if m.targetingSlot > 0 {
			slot := m.targetingSlot
			before := len(m.bag)
			m.useSupply(slot - 1)
			if len(m.bag) < before {
				m.targetingSlot = 0
			}
		}
	case "up", "down", " ", "x":
		m.updatePlay(key)
	case "b":
		switch {
		case m.phase != phaseShop:
			return
		case m.stockItem == nil:
			m.msg = "Consumable sold out"
		case len(m.bag) >= m.bagSlots():
			m.msg = "Consumable slots full"
		case m.money < 3:
			m.msg = "Need $3 to buy this consumable"
		default:
			m.msg = "Bought " + m.stockItem.Name
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
		if !m.faceDown(c) {
			label = fmt.Sprintf("%s %s · %s · %s edition · %s seal", c.Label(), suit(c), enhancements[c.Enhancement], editions[c.Edition], seals[c.Seal])
		}
		fmt.Fprintf(&b, "%s [%s] %s\n", marker, map[bool]string{true: "x", false: " "}[m.selected[i]], label)
	}
	if m.targetingSlot > 0 {
		it := m.bag[m.targetingSlot-1]
		lo, hi := it.targetCount()
		fmt.Fprintf(&b, "%s · select %d to %d cards · w/s move · x select · e use · f close\n%s", it.Name, lo, hi, m.msg)
		return panel.Padding(1, 2).Render(b.String())
	}
	if len(m.hand) > 0 {
		b.WriteString("w/s target · space/x select · ")
	}
	fmt.Fprintf(&b, "number use · f close\n%s", m.msg)
	return panel.Padding(1, 2).Render(b.String())
}

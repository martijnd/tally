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
		if !m.faceDown(c) {
			label = fmt.Sprintf("%s %s · %s · %s edition · %s seal", c.Label(), suit(c), enhancements[c.Enhancement], editions[c.Edition], seals[c.Seal])
		}
		fmt.Fprintf(&b, "%s [%s] %s\n", marker, map[bool]string{true: "x", false: " "}[m.selected[i]], label)
	}
	fmt.Fprintf(&b, "w/s target · space select · number use · f close\n%s", m.msg)
	return panel.Padding(1, 2).Render(b.String())
}

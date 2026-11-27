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
	const w = 64
	b.WriteString(spread(sAccent.Render("Consumables"), sGood.Render(fmt.Sprintf("$%d", m.money)), w))
	fmt.Fprintf(&b, "\n\n%s  %d/%d\n", sBold.Render("Inventory"), len(m.bag), m.bagSlots())
	if len(m.bag) == 0 {
		b.WriteString(sDim.Render("  Empty · Find consumables in shops and packs.") + "\n")
	}
	for i, it := range m.bag {
		key := fmt.Sprintf("[%d]", i+1)
		if i+1 == m.targetingSlot {
			key = "›" + key
		}
		b.WriteString(spread(sAccent.Render(key)+" "+sBold.Render(it.Name), sDim.Render(it.cardFamily()), w))
		b.WriteString("\n    " + sText.Width(w-4).Render(it.Description) + "\n")
	}
	if m.phase == phaseShop {
		b.WriteString("\n" + sBold.Render("For sale") + "\n")
		if it := m.stockItem; it != nil {
			b.WriteString(spread(it.Name+" · "+it.cardFamily(), "$3", w))
			b.WriteString("\n    " + sText.Width(w-4).Render(it.Description) + "\n")
			switch {
			case len(m.bag) >= m.bagSlots():
				b.WriteString(sWarn.Render("Inventory full") + "\n")
			case m.money < 3:
				fmt.Fprintf(&b, "Need $%d more\n", 3-m.money)
			default:
				b.WriteString(sAccent.Render("b buy") + "\n")
			}
		} else {
			b.WriteString(sDim.Render("  Sold out") + "\n")
		}
	}
	if len(m.hand) > 0 {
		fmt.Fprintf(&b, "\n%s  %d selected\n", sBold.Render("Hand"), len(m.selectedIdx()))
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
	if m.msg != "" {
		b.WriteString("\n" + sWarn.Render(m.msg) + "\n")
	}
	b.WriteString("\n")
	if m.targetingSlot > 0 {
		it := m.bag[m.targetingSlot-1]
		lo, hi := it.targetCount()
		fmt.Fprintf(&b, "%s · select %d to %d cards\nw/s move · x select · e use · f close", it.Name, lo, hi)
		return panel.Padding(0, 2).Width(w + 4).Render(b.String())
	}
	if len(m.hand) > 0 {
		b.WriteString("w/s target · space/x select · ")
	}
	if len(m.bag) > 0 {
		b.WriteString("number use · ")
	}
	b.WriteString("f close")
	return panel.Padding(0, 2).Width(w + 4).Render(b.String())
}

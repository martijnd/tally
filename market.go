package main

import (
	"fmt"
	"strings"
)

const (
	shelf = iota
	coupons
	savings
	rehearsal
	wastebasket
	folio
	gilding
	satchel
	cabinet
	dividends
	resale
	wholesale
	binding
	stipend
	patience
)

var upgradeLines = [...]struct{ Name, Desc string }{
	{"Shelving", "+1 Joker offer"}, {"Coupon roll", "Rerolls start $1 cheaper"}, {"Savings book", "Interest cap +$5"}, {"Rehearsal", "+1 play"}, {"Wastebasket", "+1 discard"}, {"Folio", "+1 hand size"}, {"Gilding press", "Higher edition chance"}, {"Satchel", "+1 consumable slot"}, {"Cabinet", "+1 Joker slot"}, {"Dividend bond", "Round reward +$2"}, {"Resale license", "Sell Jokers for $1 more"}, {"Wholesale license", "Joker purchases cost $1 less"}, {"Bookbinding", "Packs cost $1 less"}, {"Stipend", "Earn $1 entering each shop"}, {"Patience fund", "Unused plays earn $1 more"},
}

func (m *model) bagSlots() int {
	if m.restriction("no consumables") {
		return 0
	}
	return 2 + m.upgrades[satchel] + m.bagBonus
}
func (m *model) upgradePrice() int { return 8 + 4*m.upgrades[m.upgradeOffer] }
func (m *model) prepareMarket() {
	var eligible []int
	for i := range upgradeLines {
		if m.upgrades[i] < 2 && (i != cabinet || max(m.unlockRound, m.round+1) >= 3) {
			eligible = append(eligible, i)
		}
	}
	m.upgradeOffer = -1
	if len(eligible) > 0 {
		m.upgradeOffer = eligible[m.shopRNG.IntN(len(eligible))]
	}
	m.packKind = m.packRNG.IntN(5)
	m.packSize = m.packRNG.IntN(3)
	m.packSold = false
	m.money += m.upgrades[stipend]
}
func (m *model) buyUpgrade() {
	i := m.upgradeOffer
	if i < 0 {
		return
	}
	cost := m.upgradePrice()
	if m.money < cost {
		m.msg = "Not enough money for Voucher"
		return
	}
	m.money -= cost
	m.upgrades[i]++
	m.upgradeOffer = -1
	m.msg = "Bought " + upgradeLines[i].Name
}

var packKinds = []string{"Buffoon", "Standard", "Celestial", "Arcana", "Spectral"}
var packSizes = []string{"Normal", "Jumbo", "Mega"}

type packChoice struct {
	mod  *Modifier
	card *Card
	item *supply
}

func (p packChoice) label() string {
	if p.mod != nil {
		return p.mod.Name + " · " + p.mod.Desc
	}
	if p.card != nil {
		return cardTag(*p.card) + " · " + enhancements[p.card.Enhancement]
	}
	return p.item.Name + " · " + p.item.cardFamily() + " card · " + p.item.Description
}
func (m *model) packPrice() int {
	if m.freePacks > 0 {
		return 0
	}
	return 4 + m.packSize*2 - m.upgrades[binding]
}
func (m *model) openPack() {
	if m.restriction("no packs") {
		m.msg = "Challenge forbids packs"
		return
	}
	if m.packSold {
		return
	}
	if m.money < m.packPrice() {
		m.msg = "Not enough money for pack"
		return
	}
	m.money -= m.packPrice()
	if m.freePacks > 0 {
		m.freePacks--
	}
	m.packSold = true
	m.packChoices = nil
	m.packPicks = 1
	count := 3
	if m.packSize > 0 {
		count = 5
	}
	if m.packSize == 2 {
		m.packPicks = 2
	}
	switch m.packKind {
	case 0:
		var mods []Modifier
		for _, mod := range catalog() {
			if m.modifierUnlocked(mod.Name) {
				mods = append(mods, mod)
			}
		}
		for _, i := range m.packRNG.Perm(len(mods))[:count] {
			m.packChoices = append(m.packChoices, packChoice{mod: &mods[i]})
		}
	case 1:
		for range count {
			c := Card{Rank: 1 + m.packRNG.IntN(13), Suit: m.packRNG.IntN(4), Enhancement: m.packRNG.IntN(len(enhancements))}
			m.packChoices = append(m.packChoices, packChoice{card: &c})
		}
	default:
		var items []supply
		for _, it := range supplies() {
			if it.cardFamily() == [...]string{"Planet", "Tarot", "Spectral"}[m.packKind-2] {
				items = append(items, it)
			}
		}
		if m.packKind == 2 && m.deckIs(6) {
			best := 0
			for i, n := range m.timesPlayed {
				if n > m.timesPlayed[best] {
					best = i
				}
			}
			for i, it := range items {
				if it.Value == best {
					m.packChoices = append(m.packChoices, packChoice{item: &it})
					items = append(items[:i], items[i+1:]...)
					count--
					break
				}
			}
		}
		for _, i := range m.packRNG.Perm(len(items))[:count] {
			m.packChoices = append(m.packChoices, packChoice{item: &items[i]})
		}
	}
	m.msg = ""
}
func (m *model) pickPack(key string) {
	if key == "esc" {
		m.packChoices = nil
		return
	}
	if len(key) != 1 || key[0] < '1' || key[0] > '5' {
		return
	}
	i := int(key[0] - '1')
	if i >= len(m.packChoices) {
		return
	}
	p := m.packChoices[i]
	switch {
	case p.mod != nil:
		if len(m.mods) >= m.modSlots() {
			m.msg = "Joker slots full"
			return
		}
		m.mods = append(m.mods, p.mod)
	case p.card != nil:
		m.nextID++
		p.card.ID = m.nextID
		m.stock = append(m.stock, *p.card)
		m.emit("added")
		m.deck = append(m.deck, *p.card)
	default:
		if len(m.bag) >= m.bagSlots() {
			m.msg = "Consumable slots full"
			return
		}
		m.bag = append(m.bag, *p.item)
	}
	m.packChoices = append(m.packChoices[:i], m.packChoices[i+1:]...)
	m.packPicks--
	m.msg = ""
	if m.packPicks == 0 {
		m.packChoices = nil
	}
}
func (m *model) packView() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s · choose %d · esc leave\n", packKinds[m.packKind], m.packPicks)
	for i, p := range m.packChoices {
		fmt.Fprintf(&b, "%d  %s\n", i+1, p.label())
	}
	return panel.Padding(1, 2).Render(b.String() + m.msg)
}

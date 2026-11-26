package main

import (
	"fmt"
	"slices"
)

var planetGrowth = [...]struct{ chips, mult int }{{10, 1}, {15, 1}, {20, 1}, {20, 2}, {30, 3}, {15, 2}, {25, 2}, {30, 3}, {40, 4}, {35, 3}, {40, 4}, {50, 3}}

func supplies() []supply {
	out := []supply{
		{"The Fool", "Recreate the last Planet or Tarot used (except The Fool).", "recall", 0},
		{"The Magician", "Make one or two cards Lucky.", "enhance", 8},
		{"The High Priestess", "Generate up to two Planets in free slots.", "generate", 1},
		{"The Empress", "Give one or two cards the Mult enhancement.", "enhance", 2},
		{"The Emperor", "Generate up to two Tarots in free slots.", "generate", 0},
		{"The Hierophant", "Give one or two cards the Bonus enhancement.", "enhance", 1},
		{"The Lovers", "Make one card Wild.", "enhance", 4},
		{"The Chariot", "Make one card Steel.", "enhance", 3},
		{"Justice", "Make one card Glass.", "enhance", 5},
		{"The Hermit", "Match your cash, gaining at most $20.", "hermit", 0},
		{"The Wheel of Fortune", "1 in 4 chance to give an uneditioned Joker Foil, Holographic or Polychrome.", "wheel", 0},
		{"Strength", "Advance one or two ranks by one; Kings become Aces.", "rank", 1},
		{"The Hanged Man", "Destroy one or two selected cards.", "destroy", 0},
		{"Death", "Choose two cards. Transform the left one into the right one.", "death", 0},
		{"Temperance", "Gain the total sale price of your Jokers, up to $50.", "temperance", 0},
		{"The Devil", "Make one card Gold.", "enhance", 6},
		{"The Tower", "Make one card Stone.", "enhance", 7},
		{"The Star", "Convert up to three cards to Diamonds.", "suit", 2},
		{"The Moon", "Convert up to three cards to Clubs.", "suit", 3},
		{"The Sun", "Convert up to three cards to Hearts.", "suit", 1},
		{"Judgement", "Create a random Joker in an open slot.", "joker", -1},
		{"The World", "Convert up to three cards to Spades.", "suit", 0},
		{"Familiar", "Sacrifice a random held card for three enhanced face cards.", "summon", 3},
		{"Grim", "Sacrifice a random held card for two enhanced Aces.", "summon", 2},
		{"Incantation", "Sacrifice a random held card for four enhanced numbered cards.", "summon", 4},
		{"Talisman", "Give one card a Gold Seal.", "seal", 2},
		{"Aura", "Give one uneditioned playing card Foil, Holographic or Polychrome.", "aura", 0},
		{"Wraith", "Create a Rare Joker and set cash to $0.", "joker", 2},
		{"Sigil", "Change every held card to one random suit.", "sigil", 0},
		{"Ouija", "Change every held card to one random rank. Lose one hand size.", "ouija", 0},
		{"Ectoplasm", "Make an uneditioned Joker Negative. Lose hand size equal to uses this run, including this one.", "ectoplasm", 0},
		{"Immolate", "Destroy up to five random held cards and gain $20.", "immolate", 0},
		{"Ankh", "Duplicate one random Joker and destroy the others. The copy loses Negative.", "ankh", 0},
		{"Deja Vu", "Give one card a Red Seal.", "seal", 1},
		{"Hex", "Make a random uneditioned Joker Polychrome and destroy the others.", "hex", 0},
		{"Trance", "Give one card a Blue Seal.", "seal", 3},
		{"Medium", "Give one card a Purple Seal.", "seal", 4},
		{"Cryptid", "Add two copies of one selected playing card to your hand.", "duplicate", 0},
		{"The Soul", "Create a Legendary Joker in an open slot.", "joker", 3},
		{"Black Hole", "Raise every poker hand one level.", "allLevels", 0},
	}
	for i, b := range handBases {
		out = append(out, supply{planetNames[i], fmt.Sprintf("Level up %s (+%d Chips, +%d Mult).", b.Name, planetGrowth[i].chips, planetGrowth[i].mult), "level", i})
	}
	return out
}
func (m *model) addHeld(c Card) {
	m.nextID++
	c.ID = m.nextID
	m.stock = append(m.stock, c)
	m.hand = append(m.hand, c)
	m.selected = append(m.selected, false)
	m.emit("added")
}
func (m *model) removeHeld(i int) {
	id := m.hand[i].ID
	m.hand = slices.Delete(m.hand, i, i+1)
	m.selected = slices.Delete(m.selected, i, i+1)
	m.stock = slices.DeleteFunc(m.stock, func(c Card) bool { return c.ID == id })
	m.emit("destroyed")
	m.clampCursor()
}
func (m *model) replaceHeld(i int, c Card) {
	c.ID = m.hand[i].ID
	m.hand[i] = c
	for j := range m.stock {
		if m.stock[j].ID == c.ID {
			m.stock[j] = c
			break
		}
	}
}
func (m *model) failSupply(msg string) bool { m.msg = msg; return false }
func (m *model) randomSupply(family, exclude string) supply {
	var pool []supply
	for _, s := range supplies() {
		if s.cardFamily() == family && s.Name != exclude && !(s.Action == "level" && s.Value >= int(FiveKind) && m.timesPlayed[s.Value] == 0) {
			pool = append(pool, s)
		}
	}
	return pool[m.itemRNG.IntN(len(pool))]
}
func (m *model) randomEdition() int {
	roll := m.itemRNG.IntN(10)
	if roll < 5 {
		return 1
	}
	if roll < 8 {
		return 2
	}
	return 3
}
func (s supply) targetCount() (int, int) {
	lo, hi := 0, 0
	switch s.Action {
	case "enhance", "seal", "aura", "duplicate":
		lo, hi = 1, 1
		if s.Action == "enhance" && (s.Value == 1 || s.Value == 2 || s.Value == 8) {
			hi = 2
		}
	case "destroy", "rank":
		lo, hi = 1, 2
	case "suit":
		lo, hi = 1, 3
	case "death":
		lo, hi = 2, 2
	}
	return lo, hi
}
func (m *model) applySupply(s supply) bool {
	sel := m.selectedIdx()
	lo, hi := s.targetCount()
	if lo > 0 && len(m.hand) == 0 {
		return m.failSupply("Use " + s.Name + " during a Blind, when you have cards in hand")
	}
	if lo > 0 && (len(sel) < lo || len(sel) > hi) {
		return m.failSupply(fmt.Sprintf("Select %d–%d cards", lo, hi))
	}
	switch s.Action {
	case "enhance", "rank", "seal", "suit", "aura":
		for _, i := range sel {
			c := m.hand[i]
			switch s.Action {
			case "enhance":
				c.Enhancement = s.Value
			case "rank":
				c.Rank = c.Rank%13 + 1
			case "seal":
				c.Seal = s.Value
			case "suit":
				c.Suit = s.Value
			case "aura":
				if c.Edition != 0 {
					return m.failSupply("Choose a card without an edition")
				}
				c.Edition = m.randomEdition()
			}
			m.replaceHeld(i, c)
		}
	case "death":
		m.replaceHeld(sel[0], m.hand[sel[1]])
	case "duplicate":
		c := m.hand[sel[0]]
		m.addHeld(c)
		m.addHeld(c)
	case "destroy":
		for j := len(sel) - 1; j >= 0; j-- {
			m.removeHeld(sel[j])
		}
	case "level":
		m.levels[s.Value]++
	case "allLevels":
		for i := range m.levels {
			m.levels[i]++
		}
	case "recall":
		if m.lastSupply.Name == "" || m.lastSupply.Name == "The Fool" {
			return m.failSupply("Use another Tarot or Planet first")
		}
		m.bag = append(m.bag, m.lastSupply)
	case "generate":
		family := "Tarot"
		if s.Value == 1 {
			family = "Planet"
		}
		for range min(2, m.bagSlots()-len(m.bag)+1) {
			m.bag = append(m.bag, m.randomSupply(family, s.Name))
		}
	case "hermit":
		m.money += min(20, max(0, m.money))
	case "temperance":
		total := 0
		for _, mod := range m.mods {
			total += sellPrice(mod) + m.upgrades[resale]
		}
		m.money += min(50, total)
	case "joker":
		if len(m.mods) >= m.modSlots() {
			return m.failSupply("Requires a free Joker slot")
		}
		rarity := s.Value
		if rarity < 0 {
			n := m.itemRNG.IntN(100)
			rarity = 0
			if n >= 70 {
				rarity = 1
			}
			if n >= 95 {
				rarity = 2
			}
		}
		var pool []Modifier
		for _, mod := range catalog() {
			if mod.Rarity == rarity && m.modifierUnlocked(mod.Name) && !slices.ContainsFunc(m.mods, func(x *Modifier) bool { return x.Name == mod.Name }) {
				pool = append(pool, mod)
			}
		}
		if len(pool) == 0 {
			return m.failSupply("No eligible Joker")
		}
		mod := pool[m.itemRNG.IntN(len(pool))]
		m.mods = append(m.mods, &mod)
		if s.Value == 2 {
			m.money = 0
		}
	case "summon", "sigil", "ouija", "immolate":
		if len(m.hand) == 0 {
			return m.failSupply("Requires cards in hand")
		}
		switch s.Action {
		case "summon":
			m.removeHeld(m.itemRNG.IntN(len(m.hand)))
			for range s.Value {
				rank := 1
				if s.Value == 3 {
					rank = 11 + m.itemRNG.IntN(3)
				} else if s.Value == 4 {
					rank = 2 + m.itemRNG.IntN(9)
				}
				enh := 1 + m.itemRNG.IntN(7)
				if enh == 7 {
					enh = 8
				}
				m.addHeld(Card{Rank: rank, Suit: m.itemRNG.IntN(4), Enhancement: enh})
			}
		case "immolate":
			for range min(5, len(m.hand)) {
				m.removeHeld(m.itemRNG.IntN(len(m.hand)))
			}
			m.money += 20
		case "sigil":
			suit := m.itemRNG.IntN(4)
			for i, c := range m.hand {
				c.Suit = suit
				m.replaceHeld(i, c)
			}
		case "ouija":
			if m.drawLimit() <= 1 {
				return m.failSupply("Hand size too small")
			}
			rank := 1 + m.itemRNG.IntN(13)
			for i, c := range m.hand {
				c.Rank = rank
				m.replaceHeld(i, c)
			}
			m.handBonus--
		}
	case "ankh", "hex", "wheel", "ectoplasm":
		var eligible []*Modifier
		for _, mod := range m.mods {
			if s.Action == "ankh" || mod.Edition == 0 {
				eligible = append(eligible, mod)
			}
		}
		if len(eligible) == 0 {
			return m.failSupply("Requires an eligible Joker")
		}
		if s.Action == "ectoplasm" && m.drawLimit() <= m.ectoUses+1 {
			return m.failSupply("Hand size too small")
		}
		keep := eligible[m.itemRNG.IntN(len(eligible))]
		switch s.Action {
		case "wheel":
			if m.itemRNG.IntN(4) < m.odds() {
				keep.Edition = m.randomEdition()
			}
		case "ectoplasm":
			keep.Edition = 4
			m.ectoUses++
			m.handBonus -= m.ectoUses
		default:
			if s.Action == "ankh" {
				survivors := 0
				for _, mod := range m.mods {
					if mod == keep || mod.Sticker == "Bound" {
						survivors++
					}
				}
				if survivors >= m.modSlots() {
					return m.failSupply("No space for a copy")
				}
			}
			m.mods = slices.DeleteFunc(m.mods, func(mod *Modifier) bool { return mod != keep && mod.Sticker != "Bound" })
			if s.Action == "hex" {
				keep.Edition = 3
			} else {
				copy := *keep
				if copy.Edition == 4 {
					copy.Edition = 0
				}
				m.mods = append(m.mods, &copy)
			}
		}
	default:
		return false
	}
	if s.cardFamily() != "Spectral" {
		m.lastSupply = s
	}
	return true
}

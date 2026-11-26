package main

import (
	"fmt"
	"slices"
	"strings"
)

type startingDeck struct{ Name, Rule string }

var legacyDecks = []startingDeck{
	{"Ledger", "Standard 52-card deck"}, {"Draft", "One extra discard"}, {"Treasury", "Start with $10 extra"}, {"Courier", "One extra consumable slot"}, {"Republic", "No court cards"}, {"Duplex", "Two suits, one per colour"}, {"Collage", "Every card starts enhanced"}, {"Gallery", "One extra Joker slot"}, {"Panorama", "One extra card in hand"}, {"Encore", "One extra hand"}, {"Primer", "Pairs start at level three"}, {"Proof", "Begin with a copying tool"}, {"Mint", "Aces have Mint seals"}, {"Workshop", "Start with Ember"}, {"Thrift", "Rerolls start $1 cheaper"},
}

var startingDecks = []startingDeck{
	{"Red Deck", "+1 discard each round"}, {"Blue Deck", "+1 hand each round"}, {"Yellow Deck", "Begin with $10 extra"}, {"Green Deck", "Cash out unused hands for $2 and discards for $1. Earn no interest."}, {"Black Deck", "One extra Joker slot and one fewer hand"},
	{"Magic Deck", "Crystal Ball and two copies of The Fool"}, {"Nebula Deck", "Telescope and one fewer consumable slot"}, {"Ghost Deck", "Start with Hex. Spectral cards can appear in the shop."}, {"Abandoned Deck", "No face cards"}, {"Checkered Deck", "26 Hearts and 26 Spades"},
	{"Zodiac Deck", "Tarot Merchant, Planet Merchant and Overstock"}, {"Painted Deck", "Two extra cards in hand and one fewer Joker slot"}, {"Anaglyph Deck", "Earn a Double Tag after each Boss Blind"}, {"Plasma Deck", "Balance Chips and Mult before scoring. Double Blind sizes."}, {"Erratic Deck", "Begin with 52 randomly ranked and suited cards"},
}

func (m *model) decks() []startingDeck {
	if m.legacyEffects {
		return legacyDecks
	}
	return startingDecks
}
func (m *model) deckIs(id int) bool { return !m.legacyEffects && m.deckID == id }
func (m *model) chooseDeck() {
	switch m.deckID {
	case 0:
		m.discardBonus++
	case 1:
		m.playBonus++
	case 2:
		m.money += 10
	case 4:
		m.slotBonus++
		m.playBonus--
	case 5:
		m.upgrades[satchel] = 1
		m.bag = []supply{supplies()[0], supplies()[0]}
	case 6:
		m.bagBonus--
	case 7:
		for _, it := range supplies() {
			if it.Name == "Hex" {
				m.bag = append(m.bag, it)
				break
			}
		}
	case 8:
		m.stock = slices.DeleteFunc(m.stock, func(c Card) bool { return c.Rank >= 11 })
	case 9:
		for i := range m.stock {
			switch m.stock[i].Suit {
			case 2:
				m.stock[i].Suit = 1
			case 3:
				m.stock[i].Suit = 0
			}
		}
	case 10:
		m.upgrades[shelf] = 1
	case 11:
		m.handBonus += 2
		m.slotBonus--
	case 14:
		for i := range m.stock {
			m.stock[i].Rank = 1 + m.deckRNG.IntN(13)
			m.stock[i].Suit = m.deckRNG.IntN(4)
		}
	}
}

type challenge struct{ Name, Mod, Rule string }

var challenges = []challenge{
	{"One Shot", "Last Call", "single play"}, {"Small Purse", "Stockpile", "cash cap"}, {"Close Quarters", "Ember", "small hand"}, {"One Drawer", "Twins", "single slot"}, {"Final Draft", "Momentum", "no discards"}, {"Long Week", "Wide Net", "high targets"}, {"Cash Only", "Stockpile", "no interest"}, {"Unopened Mail", "Odd Ones Out", "no consumables"}, {"Uncut Sheet", "Even Keel", "no skips"}, {"Private Press", "Small Change", "no packs"},
	{"Last Dance", "Last Call", "no rerolls"}, {"Narrow Gauge", "Twins", "small hand"}, {"Thin Margins", "Momentum", "cash cap"}, {"Sealed Room", "Wide Net", "single slot"}, {"First Edition", "Odd Ones Out", "high targets"}, {"One Take", "Ember", "single play"}, {"Loose Change", "Small Change", "no interest"}, {"Fixed Stock", "Even Keel", "no rerolls"}, {"Full Commitment", "Last Call", "no skips"}, {"House Rules", "Momentum", "no packs"},
}
var difficultyRules = []string{"Standard", "No Small Blind reward", "Steeper targets", "Eternal Jokers appear", "One fewer discard", "Perishable Jokers appear", "Reduced interest", "Rental stickers and steeper targets"}

func (m *model) restriction(rule string) bool {
	return m.challengeID > 0 && challenges[m.challengeID-1].Rule == rule
}
func (m *model) startModifier(name string) {
	for _, mod := range catalog() {
		if mod.Name == name {
			m.mods = append(m.mods, &mod)
			return
		}
	}
}
func (m *model) legacyStart() {
	switch m.deckID {
	case 1:
		m.discardBonus++
	case 2:
		m.money += 10
	case 3:
		m.bagBonus++
	case 4:
		m.stock = slices.DeleteFunc(m.stock, func(c Card) bool { return c.Rank >= 11 })
	case 5:
		for i := range m.stock {
			if m.stock[i].Suit == 2 {
				m.stock[i].Suit = 1
			}
			if m.stock[i].Suit == 3 {
				m.stock[i].Suit = 0
			}
		}
	case 6:
		for i := range m.stock {
			m.stock[i].Enhancement = 1 + m.deckRNG.IntN(len(enhancements)-1)
		}
	case 7:
		m.slotBonus++
	case 8:
		m.handBonus++
	case 9:
		m.playBonus++
	case 10:
		m.levels[Pair] = 2
	case 11:
		m.bag = append(m.bag, m.supplyPool()[0])
	case 12:
		for i := range m.stock {
			if m.stock[i].Rank == 1 {
				m.stock[i].Seal = 2
			}
		}
	case 13:
		m.startModifier("Ember")
	case 14:
		m.upgrades[coupons] = 1
	}
}
func (m *model) chooseStart() {
	if m.legacyEffects {
		m.legacyStart()
	} else {
		m.chooseDeck()
	}
	if m.challengeID > 0 {
		m.startModifier(challenges[m.challengeID-1].Mod)
	}
	m.phase = phasePreview
}
func (m *model) startKey(key string) {
	switch key {
	case "up":
		m.deckID = (m.deckID + len(startingDecks) - 1) % len(startingDecks)
	case "down":
		m.deckID = (m.deckID + 1) % len(startingDecks)
	case "c":
		m.challengeID = (m.challengeID + 1) % (len(challenges) + 1)
	case "left":
		m.difficulty = max(0, m.difficulty-1)
	case "right":
		m.difficulty = min(7, m.bestDifficulty[m.deckID], m.difficulty+1)
	case "enter":
		if m.deckAvailable(m.deckID) {
			m.chooseStart()
		}
	}
	m.difficulty = min(m.difficulty, m.bestDifficulty[m.deckID])
}
func (m *model) startView() string {
	var b strings.Builder
	b.WriteString("Choose your deck\n")
	for i, d := range m.decks() {
		mark := " "
		if i == m.deckID {
			mark = ">"
		}
		rule := d.Rule
		if !m.deckAvailable(i) {
			rule = m.unlockHint(i)
		}
		fmt.Fprintf(&b, "%s %-12s %s\n", mark, d.Name, rule)
	}
	fmt.Fprintf(&b, "\n%s Stake · %s\n", stakeNames[m.difficulty], difficultyRules[m.difficulty])
	if m.challengeID == 0 {
		b.WriteString("Standard run\n")
	} else {
		c := challenges[m.challengeID-1]
		fmt.Fprintf(&b, "%s · start with %s · %s\n", c.Name, c.Mod, c.Rule)
	}
	b.WriteString("w/s deck · 1/2 stake · c challenge · e start · z collection · , settings")
	return panel.Padding(1, 2).Render(b.String())
}
func (m *model) roundReward() int {
	if m.difficulty >= 1 && m.round%3 == 0 {
		return 0
	}
	if !m.legacyEffects && m.round%24 == 23 {
		return 8
	}
	return 3 + m.round%3 + 2*m.upgrades[dividends]
}

func (m *model) stickerOffer(mod *Modifier) {
	if m.difficulty < 3 {
		return
	}
	roll := m.shopRNG.IntN(8)
	switch {
	case roll == 0:
		mod.Sticker = "Bound"
	case roll == 1 && m.difficulty >= 5:
		mod.Sticker = "Fading"
		mod.Life = 4
	case roll == 2 && m.difficulty >= 7:
		mod.Sticker = "Rent"
	}
	if mod.Sticker != "" {
		mod.Desc = map[string]string{"Bound": "Eternal", "Fading": "Perishable", "Rent": "Rental"}[mod.Sticker] + " · " + mod.Desc
	}
}
func (m *model) settleStickers() {
	kept := m.mods[:0]
	for _, mod := range m.mods {
		switch mod.Sticker {
		case "Fading":
			mod.Life--
			if mod.Life <= 0 {
				continue
			}
		case "Rent":
			m.money = max(0, m.money-1)
		}
		kept = append(kept, mod)
	}
	m.mods = kept
}

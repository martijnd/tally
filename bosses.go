package main

import "fmt"

type boss struct {
	Name, Effect, Rule string
	Amount             int
}

var regularBosses = []boss{
	{"The Hook", "Discard two random held cards after playing", "shed", 2}, {"The Club", "Clubs are debuffed", "suit", 3}, {"The Psychic", "Play exactly five cards", "minimum", 5}, {"The Goad", "Spades are debuffed", "suit", 0}, {"The Window", "Diamonds are debuffed", "suit", 2}, {"The Manacle", "Hand size reduced by one", "size", 1}, {"The Pillar", "Cards played earlier this Ante are debuffed", "pillar", 0}, {"The Head", "Hearts are debuffed", "suit", 1},
	{"The House", "First draw is face down", "house", 0}, {"The Wall", "Four times the base target", "target", 4}, {"The Wheel", "One in seven drawn cards is face down", "wheel", 0}, {"The Arm", "Played hands lose a level", "arm", 0}, {"The Fish", "Draw face down after playing", "fish", 0}, {"The Water", "Start with zero discards", "water", 0}, {"The Mouth", "Repeat the first poker hand type", "lock", 0}, {"The Needle", "Only one hand", "needle", 0}, {"The Flint", "Halve base Chips and Mult", "flint", 0}, {"The Mark", "Face cards are drawn face down", "mark", 0},
	{"The Eye", "Each hand type can score once", "eye", 0}, {"The Tooth", "Lose $1 per played card", "tooth", 0}, {"The Plant", "Face cards are debuffed", "faces", 0}, {"The Serpent", "Draw three after a hand or discard", "serpent", 0}, {"The Ox", "Lose all cash for playing the most-played hand", "ox", 0},
}
var finalBosses = []boss{
	{"Amber Acorn", "Shuffle Jokers and turn them face down", "acorn", 0}, {"Verdant Leaf", "All cards debuffed until a Joker is sold", "leaf", 0}, {"Violet Vessel", "Six times the base target", "target", 6}, {"Crimson Heart", "One random Joker is disabled per hand", "heart", 0}, {"Cerulean Bell", "One held card must remain selected", "bell", 0},
}

func (m *model) nextBoss(stage int) boss {
	pool := regularBosses
	switch ante := stage + 1; {
	case ante%8 == 0:
		pool = finalBosses
	case ante == 1:
		pool = pool[:8]
	case ante == 2:
		pool = pool[:18]
	case ante == 3:
		pool = pool[:20]
	case ante == 4:
		pool = pool[:21]
	case ante == 5:
		pool = pool[:22]
	}
	counts := map[string]int{}
	for _, b := range m.bosses {
		counts[b.Name]++
	}
	least := stage + 1
	var candidates []boss
	for _, b := range pool {
		n := counts[b.Name]
		if n < least {
			least = n
			candidates = nil
		}
		if n == least {
			candidates = append(candidates, b)
		}
	}
	return candidates[runStream(m.seed, fmt.Sprintf("boss-%d", stage)).IntN(len(candidates))]
}
func (m *model) cardDebuffed(c Card) bool {
	b := m.boss()
	return b.Rule == "leaf" || (b.Rule == "pillar" && m.playedAnte[c.ID]) || b.debuffs(c)
}
func (m *model) faceDown(c Card) bool { return m.hidden[c.ID] }
func (m *model) forceCard() {
	if m.boss().Rule == "bell" && len(m.hand) > 0 {
		for i, c := range m.hand {
			if c.ID == m.forcedID {
				m.selected[i] = true
				return
			}
		}
		i := m.deckRNG.IntN(len(m.hand))
		m.forcedID = m.hand[i].ID
		m.selected[i] = true
	}
}
func (m *model) boss() boss {
	if m.round%3 != 2 || m.bossDisabled {
		return boss{}
	}
	return m.stageBoss(m.round / 3)
}
func (m *model) stageBoss(stage int) boss {
	for len(m.bosses) <= stage {
		m.bosses = append(m.bosses, m.nextBoss(len(m.bosses)))
	}
	return m.bosses[stage]
}
func (b boss) debuffs(c Card) bool {
	if c.Enhancement == 7 {
		return false
	}
	switch b.Rule {
	case "suit":
		return c.Suit == b.Amount || c.Enhancement == 4
	case "faces":
		return c.Rank >= 11
	}
	return false
}
func (m *model) preview() string {
	description := fmt.Sprintf("Ante %d · %s\nTarget %s · reward $%d\n", m.round/3+1, roundNames[m.round%3], commas(m.target()), m.roundReward())
	if m.round%3 == 2 {
		b := m.boss()
		description += b.Name + " · " + b.Effect
	} else if !m.restriction("no skips") {
		t := tags[m.skipTag()]
		description += "s skip · " + t.Name + " · " + t.Desc
	}
	return panel.Padding(1, 2).Render(description + m.tagSummary() + "\n\ne begin · g hands · q quit")
}

var roundNames = []string{"Small Blind", "Big Blind", "Boss Blind"}

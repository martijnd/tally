package main

import "fmt"

type boss struct {
	Name, Effect, Rule string
	Amount             int
}

var legacyRegularBosses = []boss{
	{"Ash Orchard", "First suit loses card scoring", "suit", 0}, {"Glass Canal", "Second suit loses card scoring", "suit", 1}, {"Slate Choir", "Third suit loses card scoring", "suit", 2}, {"Copper Marsh", "Fourth suit loses card scoring", "suit", 3},
	{"Empty Gallery", "Court cards lose card scoring", "faces", 0}, {"Folded Room", "Draw one fewer card", "size", 1}, {"Single Thread", "Repeat the first hand type played", "lock", 0}, {"Fog Ledger", "Cards are drawn face down", "blind", 0}, {"Loose Binding", "Lose one random held card per play", "shed", 1}, {"Late Train", "First play earns zero", "late", 0}, {"Toll House", "Pay $1 per play", "toll", 1}, {"Closed Circuit", "Joker scoring disabled", "mute", 0},
	{"Even Tide", "Even ranks lose card scoring", "parity", 0}, {"Odd Tide", "Odd ranks lose card scoring", "parity", 1}, {"Low Ceiling", "Ranks 2 through 5 lose card scoring", "low", 0}, {"Broken Crown", "Aces lose card scoring", "ace", 0}, {"Short Shift", "One fewer hand", "plays", 1}, {"Dry Well", "Two fewer discards", "discards", 2}, {"Narrow Gate", "Score with at most three played cards", "limit", 3}, {"Full Bench", "Score only five-card plays", "minimum", 5}, {"Silent Partner", "Pairs score zero", "ban", int(Pair)}, {"Bare Branch", "High cards score zero", "ban", int(HighCard)},
}
var legacyFinalBosses = []boss{{"Iron Folio", "All Joker scoring disabled and one fewer hand", "finalmute", 0}, {"Night Ferry", "Pay $4 per play", "toll", 4}, {"Last Archive", "Draw two fewer cards", "size", 2}}

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
func (m *model) faceDown(c Card) bool { return m.boss().Rule == "blind" || m.hidden[c.ID] }
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
	if !m.legacyEffects {
		for len(m.bosses) <= stage {
			m.bosses = append(m.bosses, m.nextBoss(len(m.bosses)))
		}
	}
	if stage < len(m.bosses) {
		return m.bosses[stage]
	}
	amount := 5 + stage - len(m.bosses)
	return boss{fmt.Sprintf("Deep customs %d", amount), fmt.Sprintf("Pay $%d per play", amount), "toll", amount}
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
	case "parity":
		return c.Rank%2 == b.Amount
	case "low":
		return c.Rank >= 2 && c.Rank <= 5
	case "ace":
		return c.Rank == 1
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

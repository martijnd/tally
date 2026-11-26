package main

import "fmt"

type boss struct {
	Name, Effect, Rule string
	Amount             int
}

var regularBosses = []boss{
	{"Ash Orchard", "First suit loses card scoring", "suit", 0}, {"Glass Canal", "Second suit loses card scoring", "suit", 1}, {"Slate Choir", "Third suit loses card scoring", "suit", 2}, {"Copper Marsh", "Fourth suit loses card scoring", "suit", 3},
	{"Empty Gallery", "Court cards lose card scoring", "faces", 0}, {"Folded Room", "Draw one fewer card", "size", 1}, {"Single Thread", "Repeat the first hand type played", "lock", 0}, {"Fog Ledger", "Cards are drawn face down", "blind", 0}, {"Loose Binding", "Lose one random held card per play", "shed", 1}, {"Late Train", "First play earns zero", "late", 0}, {"Toll House", "Pay $1 per play", "toll", 1}, {"Closed Circuit", "Joker scoring disabled", "mute", 0},
	{"Even Tide", "Even ranks lose card scoring", "parity", 0}, {"Odd Tide", "Odd ranks lose card scoring", "parity", 1}, {"Low Ceiling", "Ranks 2 through 5 lose card scoring", "low", 0}, {"Broken Crown", "Aces lose card scoring", "ace", 0}, {"Short Shift", "One fewer hand", "plays", 1}, {"Dry Well", "Two fewer discards", "discards", 2}, {"Narrow Gate", "Score with at most three played cards", "limit", 3}, {"Full Bench", "Score only five-card plays", "minimum", 5}, {"Silent Partner", "Pairs score zero", "ban", int(Pair)}, {"Bare Branch", "High cards score zero", "ban", int(HighCard)},
}
var finalBosses = []boss{{"Iron Folio", "All Joker scoring disabled and one fewer hand", "finalmute", 0}, {"Night Ferry", "Pay $4 per play", "toll", 4}, {"Last Archive", "Draw two fewer cards", "size", 2}}

func (m *model) boss() boss {
	if m.round%3 != 2 || m.bossDisabled {
		return boss{}
	}
	return m.stageBoss(m.round / 3)
}
func (m *model) stageBoss(stage int) boss {
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
		return c.Suit == b.Amount
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
	return panel.Padding(1, 2).Render(description + m.tagSummary() + "\n\nenter begin · h hands · q quit")
}

var roundNames = []string{"Small Blind", "Big Blind", "Boss Blind"}

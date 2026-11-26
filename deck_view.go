package main

import (
	"fmt"
	"sort"
	"strings"
)

func (m *model) deckView() string {
	source, title := m.stock, "Full deck"
	if m.remainingDeck {
		source, title = m.deck, "Draw pile"
	}
	cards := append([]Card(nil), source...)
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].Suit != cards[j].Suit {
			return cards[i].Suit < cards[j].Suit
		}
		return cards[i].Order() > cards[j].Order()
	})
	var b strings.Builder
	fmt.Fprintf(&b, "%s · %d cards · tab change deck · v close\n", title, len(cards))
	for s, glyph := range suitGlyphs {
		fmt.Fprintf(&b, "\n%s ", glyph)
		for _, c := range cards {
			if c.Suit != s {
				continue
			}
			b.WriteString(c.Label())
			for _, detail := range []string{enhancements[c.Enhancement], editions[c.Edition], seals[c.Seal]} {
				if detail != "plain" && detail != "standard" && detail != "none" {
					fmt.Fprintf(&b, "[%s]", detail)
				}
			}
			b.WriteByte(' ')
		}
	}
	fmt.Fprintf(&b, "\n\nJokers %d/%d", len(m.mods), m.modSlots())
	if len(m.mods) == 0 {
		b.WriteString("\nNo Jokers")
	}
	for i, mod := range m.mods {
		rule := m.boss().Rule
		if rule == "acorn" {
			b.WriteString("\n◆ ??")
			continue
		}
		fmt.Fprintf(&b, "\n◆ %s", mod.Name)
		if rule == "heart" && i == m.disabledJoker {
			b.WriteString(" (disabled)")
		}
		if mod.Counter > 0 {
			fmt.Fprintf(&b, " (+%d)", mod.Counter)
		}
		if mod.Edition > 0 {
			fmt.Fprintf(&b, " [%s]", editions[mod.Edition])
		}
		if mod.Sticker != "" {
			fmt.Fprintf(&b, " [%s]", mod.Sticker)
		}
		fmt.Fprintf(&b, " · %s", mod.Desc)
	}
	return panel.Padding(1, 2).Render(b.String())
}

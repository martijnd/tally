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
	return panel.Padding(1, 2).Render(b.String())
}

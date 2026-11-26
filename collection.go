package main

import (
	"fmt"
	"strings"
)

func (m *model) deckUnlock(i int) int {
	if i < 12 {
		return 0
	}
	return (i - 11) * 3
}
func (m *model) modifierUnlocked(name string) bool {
	return name != "Last Call" || m.unlockRound >= 3 || m.round >= 2
}
func (m *model) collectionKey(k string) {
	switch k {
	case "a", "esc":
		m.collection = false
	case "tab":
		m.collectionGroup = (m.collectionGroup + 1) % 4
		m.collectionPage = 0
	case "right":
		m.collectionPage++
	case "left":
		m.collectionPage = max(0, m.collectionPage-1)
	}
}
func (m *model) collectionView() string {
	p := m.profile
	if p == nil {
		p = &profile{}
	}
	groups := []string{"Decks", "Modifiers", "Supplies", "Upgrades"}
	var names []string
	switch m.collectionGroup {
	case 0:
		for _, d := range startingDecks {
			names = append(names, d.Name)
		}
	case 1:
		for _, v := range catalog() {
			names = append(names, v.Name)
		}
	case 2:
		for _, v := range supplies() {
			names = append(names, v.Name)
		}
	case 3:
		for _, v := range upgradeLines {
			names = append(names, v.Name)
		}
	}
	page := min(m.collectionPage, (len(names)-1)/12)
	var b strings.Builder
	most, count := 0, 0
	for i, n := range p.Hands {
		if n > count {
			most, count = i, n
		}
	}
	fmt.Fprintf(&b, "Career · best play %s · highest round %d\nMost played · %s (%d)\n", commas(p.Best), p.Round, handBases[most].Name, count)
	wins := 0
	for i, deck := range p.Wins {
		for d, n := range deck {
			if n > 0 {
				wins += n
				fmt.Fprintf(&b, "%s · difficulty %d · %d wins\n", startingDecks[i].Name, d+1, n)
			}
		}
	}
	for _, a := range []struct {
		name   string
		earned bool
	}{{"First clearance", p.Round >= 2}, {"Four figures", p.Best >= 1000}, {"Master printer", wins > 0}} {
		mark := "○"
		if a.earned {
			mark = "●"
		}
		fmt.Fprintf(&b, "%s %s  ", mark, a.name)
	}
	fmt.Fprintf(&b, "\n\n%s · page %d/%d\n", groups[m.collectionGroup], page+1, (len(names)+11)/12)
	for _, name := range names[page*12 : min(len(names), (page+1)*12)] {
		if !p.Seen[name] {
			name = "?"
		}
		fmt.Fprintln(&b, name)
	}
	b.WriteString("tab category · ←→ page · a close")
	return panel.Padding(1, 2).Render(b.String())
}

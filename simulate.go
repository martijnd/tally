package main

import (
	"fmt"
	"io"
	"sort"
)

// The simulated player tries rank groups, suits and consecutive ranks, then
// plays the strongest candidate by base value. Modifier gains are measured
// during real scoring, so evaluating candidates never consumes random rolls.
func (m *model) simulateSelection() {
	best := 0
	check := func(indices []int) {
		if len(indices) == 0 {
			return
		}
		indices = indices[:min(len(indices), maxSelect)]
		var cards []Card
		for _, i := range indices {
			cards = append(cards, m.hand[i])
		}
		kind, scored := evaluate(cards)
		base := trainedBase(kind, m.levels[kind])
		chips := base.Chips
		for _, i := range scored {
			chips += cards[i].Chips()
		}
		value := chips * base.Mult
		if value > best {
			best = value
			clear(m.selected)
			for _, i := range indices {
				m.selected[i] = true
			}
		}
	}
	var ranks [15][]int
	var suits [4][]int
	for i, c := range m.hand {
		ranks[c.Order()] = append(ranks[c.Order()], i)
		suits[c.Suit] = append(suits[c.Suit], i)
	}
	for _, group := range suits {
		check(group)
	}
	var groups [][]int
	for _, r := range ranks {
		if len(r) > 0 {
			groups = append(groups, r)
		}
	}
	sort.SliceStable(groups, func(i, j int) bool { return len(groups[i]) > len(groups[j]) })
	for _, r := range groups {
		check(r)
	}
	if len(groups) > 1 {
		indices := append([]int(nil), groups[0]...)
		check(append(indices, groups[1]...))
	}
	ranks[1] = ranks[14]
	for lo := 1; lo <= 10; lo++ {
		var seq []int
		for r := lo; r < lo+5; r++ {
			if len(ranks[r]) > 0 {
				seq = append(seq, ranks[r][0])
			}
		}
		if len(seq) == 5 {
			check(seq)
		}
	}
}
func simulate(w io.Writer, n int, seed string) {
	reached := [24]int{}
	wins := 0
	type result struct{ runs, rounds int }
	byMod := map[string]result{}
	mods := catalog()
	for i := range n {
		m := newModel(fmt.Sprintf("%s/%d", seed, i))
		m.chooseStart()
		// Cycle starting modifiers to compare how far each build advances.
		starter := mods[i%len(mods)]
		m.mods = append(m.mods, &starter)
		for m.round < numRounds {
			reached[m.round]++
			m.startRound()
			for m.phase == phasePlay {
				m.simulateSelection()
				if !m.play() {
					break
				}
				m.finishPlay()
			}
			if m.phase != phaseRoundWon {
				break
			}
			m.round++
			if m.round == numRounds {
				wins++
				break
			}
			m.openShop()
			for len(m.mods) < m.modSlots() && len(m.offers) > 0 {
				j := -1
				for k, offer := range m.offers {
					if offer.Price <= m.money {
						j = k
						break
					}
				}
				if j < 0 {
					break
				}
				m.shopCursor = j
				m.shopSelect()
			}
		}
		r := byMod[starter.Name]
		r.runs++
		r.rounds += m.round
		byMod[starter.Name] = r
	}
	fmt.Fprintf(w, "%d runs · %d wins\n", n, wins)
	for i, v := range reached {
		fmt.Fprintf(w, "Round %2d reached %d\n", i+1, v)
	}
	for _, mod := range mods {
		r := byMod[mod.Name]
		if r.runs > 0 {
			fmt.Fprintf(w, "%-24s %5d runs  %.2f rounds cleared on average\n", mod.Name, r.runs, float64(r.rounds)/float64(r.runs))
		}
	}
}

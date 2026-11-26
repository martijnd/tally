package main

import (
	"sort"
	"strconv"
)

type Card struct {
	Rank int // 1 = A, 11 = J, 12 = Q, 13 = K
	Suit int // 0..3
}

// Label is the rank as shown on screen.
func (c Card) Label() string {
	switch c.Rank {
	case 1:
		return "A"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	}
	return strconv.Itoa(c.Rank)
}

// Order ranks A highest, for sorting and high card.
func (c Card) Order() int {
	if c.Rank == 1 {
		return 14
	}
	return c.Rank
}

// Chips is what the card adds when it scores: A = 11, J/Q/K = 10.
func (c Card) Chips() int {
	switch {
	case c.Rank == 1:
		return 11
	case c.Rank > 10:
		return 10
	}
	return c.Rank
}

func newDeck() []Card {
	d := make([]Card, 0, 52)
	for s := 0; s < 4; s++ {
		for r := 1; r <= 13; r++ {
			d = append(d, Card{Rank: r, Suit: s})
		}
	}
	return d
}

type HandType int

const (
	HighCard HandType = iota
	Pair
	TwoPair
	ThreeKind
	Straight
	Flush
	FullHouse
	FourKind
	StraightFlush
	FiveKind
	FlushHouse
	FlushFive
)

type handBase struct {
	Name        string
	Chips, Mult int
}

var handBases = [...]handBase{
	FiveKind: {"five of a kind", 115, 10}, FlushHouse: {"flush house", 135, 12}, FlushFive: {"flush five", 160, 14},
	HighCard:      {"high card", 5, 1},
	Pair:          {"pair", 10, 2},
	TwoPair:       {"two pair", 20, 2},
	ThreeKind:     {"three of a kind", 30, 3},
	Straight:      {"straight", 30, 4},
	Flush:         {"flush", 35, 4},
	FullHouse:     {"full house", 40, 4},
	FourKind:      {"four of a kind", 60, 7},
	StraightFlush: {"straight flush", 100, 8},
}

// evaluate returns the hand type for 1-5 cards and the indices of the cards
// that count toward the score, in ascending order.
func evaluate(cards []Card) (HandType, []int) {
	n := len(cards)
	all := make([]int, n)
	for i := range all {
		all[i] = i
	}

	type group struct {
		rank int
		idx  []int
	}
	byRank := map[int]*group{}
	var groups []*group
	for i, c := range cards {
		g, ok := byRank[c.Rank]
		if !ok {
			g = &group{rank: c.Order()}
			byRank[c.Rank] = g
			groups = append(groups, g)
		}
		g.idx = append(g.idx, i)
	}
	sort.SliceStable(groups, func(a, b int) bool {
		if len(groups[a].idx) != len(groups[b].idx) {
			return len(groups[a].idx) > len(groups[b].idx)
		}
		return groups[a].rank > groups[b].rank
	})

	flush := n == 5
	for _, c := range cards {
		if c.Suit != cards[0].Suit {
			flush = false
		}
	}
	straight := n == 5 && len(groups) == 5 && isRun(cards)

	top := len(groups[0].idx)
	second := 0
	if len(groups) > 1 {
		second = len(groups[1].idx)
	}
	join := func(a, b []int) []int {
		out := append(append([]int{}, a...), b...)
		sort.Ints(out)
		return out
	}

	switch {
	case top == 5 && flush:
		return FlushFive, all
	case top == 5:
		return FiveKind, all
	case top == 3 && second == 2 && flush:
		return FlushHouse, all
	case straight && flush:
		return StraightFlush, all
	case top == 4:
		return FourKind, join(groups[0].idx, nil)
	case top == 3 && second == 2:
		return FullHouse, join(groups[0].idx, groups[1].idx)
	case flush:
		return Flush, all
	case straight:
		return Straight, all
	case top == 3:
		return ThreeKind, join(groups[0].idx, nil)
	case top == 2 && second == 2:
		return TwoPair, join(groups[0].idx, groups[1].idx)
	case top == 2:
		return Pair, join(groups[0].idx, nil)
	}
	return HighCard, groups[0].idx[:1]
}

// isRun reports whether five distinct ranks are consecutive. 1 may sit
// below 2 or above 13.
func isRun(cards []Card) bool {
	ranks := make([]int, len(cards))
	for i, c := range cards {
		ranks[i] = c.Rank
	}
	sort.Ints(ranks)
	if ranks[4]-ranks[0] == 4 {
		return true
	}
	return ranks[0] == 1 && ranks[1] == 10 && ranks[4] == 13
}

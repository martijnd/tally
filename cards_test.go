package main

import (
	"slices"
	"testing"
)

func hand(cs ...[2]int) []Card {
	out := make([]Card, len(cs))
	for i, c := range cs {
		out[i] = Card{Rank: c[0], Suit: c[1]}
	}
	return out
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name    string
		cards   []Card
		want    HandType
		scoring []int
	}{
		{"high card", hand([2]int{3, 0}, [2]int{9, 1}), HighCard, []int{1}},
		{"pair", hand([2]int{7, 0}, [2]int{2, 1}, [2]int{7, 2}), Pair, []int{0, 2}},
		{"two pair", hand([2]int{7, 0}, [2]int{2, 1}, [2]int{7, 2}, [2]int{2, 3}, [2]int{9, 0}), TwoPair, []int{0, 1, 2, 3}},
		{"three", hand([2]int{5, 0}, [2]int{5, 1}, [2]int{5, 2}), ThreeKind, []int{0, 1, 2}},
		{"straight", hand([2]int{4, 0}, [2]int{5, 1}, [2]int{6, 2}, [2]int{7, 3}, [2]int{8, 0}), Straight, []int{0, 1, 2, 3, 4}},
		{"high straight", hand([2]int{1, 0}, [2]int{10, 1}, [2]int{11, 2}, [2]int{12, 3}, [2]int{13, 0}), Straight, []int{0, 1, 2, 3, 4}},
		{"flush", hand([2]int{2, 1}, [2]int{5, 1}, [2]int{9, 1}, [2]int{11, 1}, [2]int{13, 1}), Flush, []int{0, 1, 2, 3, 4}},
		{"full house", hand([2]int{3, 0}, [2]int{9, 1}, [2]int{3, 2}, [2]int{9, 3}, [2]int{3, 1}), FullHouse, []int{0, 1, 2, 3, 4}},
		{"four", hand([2]int{7, 0}, [2]int{7, 1}, [2]int{7, 2}, [2]int{7, 3}, [2]int{2, 0}), FourKind, []int{0, 1, 2, 3}},
		{"straight flush", hand([2]int{9, 2}, [2]int{10, 2}, [2]int{11, 2}, [2]int{12, 2}, [2]int{13, 2}), StraightFlush, []int{0, 1, 2, 3, 4}},
		{"ace is high card", hand([2]int{13, 0}, [2]int{1, 1}, [2]int{4, 2}), HighCard, []int{1}},
		{"low straight", hand([2]int{1, 0}, [2]int{2, 1}, [2]int{3, 2}, [2]int{4, 3}, [2]int{5, 0}), Straight, []int{0, 1, 2, 3, 4}},
		{"four-card run is not a straight", hand([2]int{4, 0}, [2]int{5, 1}, [2]int{6, 2}, [2]int{7, 3}), HighCard, []int{3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, scoring := evaluate(tt.cards)
			if got != tt.want {
				t.Fatalf("type = %s, want %s", handBases[got].Name, handBases[tt.want].Name)
			}
			if !slices.Equal(scoring, tt.scoring) {
				t.Fatalf("scoring = %v, want %v", scoring, tt.scoring)
			}
		})
	}
}

func TestScoreHandWithModifiers(t *testing.T) {
	cards := hand([2]int{7, 0}, [2]int{7, 1}, [2]int{2, 2})
	var ember, twins Modifier
	for _, m := range catalog() {
		switch m.Name {
		case "Ember":
			ember = m
		case "Twins":
			twins = m
		}
	}
	// pair: 10 chips + 7 + 7 = 24; mult 2 + 4 (Ember) = 6, ×2 (Twins) = 12
	_, total := scoreHand(cards, []int{0, 1}, Pair, []*Modifier{&ember, &twins}, ScoreCtx{Type: Pair})
	if total != 24*12 {
		t.Fatalf("total = %d, want %d", total, 24*12)
	}
}

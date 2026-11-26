package main

import "fmt"

type ScoreCtx struct {
	Plasma      bool
	Legacy      bool
	Lucky       func() float64
	Previous    int
	OnScore     func(Card)
	HeldPatient int
	Training    int
	Type        HandType
	Scoring     []Card
	Played      int
	PlaysLeft   int
	Money       int
}

// Modifier changes scoring. card runs once per scoring card, hand runs once
// after all cards, after runs when the play is finished.
type Modifier struct {
	Rarity, RepeatRank, ResaleGrowth int
	Copy                             bool
	Trigger, Effect                  string
	Sticker                          string
	Life                             int
	Edition                          int
	Name                             string
	Desc                             string
	Price                            int
	Counter                          int

	card  func(m *Modifier, c Card) (chips, mult int)
	hand  func(m *Modifier, ctx ScoreCtx, chips int, mult float64) (label string, newChips int, newMult float64)
	after func(m *Modifier, ctx ScoreCtx)
}

func containsPair(t HandType) bool {
	switch t {
	case Pair, TwoPair, ThreeKind, FullHouse, FourKind:
		return true
	}
	return false
}

func catalog() []Modifier {
	mods := []Modifier{
		{
			Name: "Ember", Desc: "+4 mult", Price: 4,
			hand: func(_ *Modifier, _ ScoreCtx, c int, mu float64) (string, int, float64) {
				return "+4 mult", c, mu + 4
			},
		},
		{
			Name: "Odd Ones Out", Desc: "+2 mult for each scoring odd card", Price: 5,
			card: func(_ *Modifier, c Card) (int, int) {
				if c.Rank%2 == 1 {
					return 0, 2
				}
				return 0, 0
			},
		},
		{
			Name: "Even Keel", Desc: "+12 chips for each scoring even card", Price: 5,
			card: func(_ *Modifier, c Card) (int, int) {
				if c.Rank%2 == 0 {
					return 12, 0
				}
				return 0, 0
			},
		},
		{
			Name: "Momentum", Desc: "+1 mult for every hand played while owned", Price: 6,
			hand: func(m *Modifier, _ ScoreCtx, c int, mu float64) (string, int, float64) {
				if m.Counter == 0 {
					return "", c, mu
				}
				return fmt.Sprintf("+%d mult", m.Counter), c, mu + float64(m.Counter)
			},
			after: func(m *Modifier, _ ScoreCtx) { m.Counter++ },
		},
		{
			Name: "Twins", Desc: "×2 mult if the hand contains a pair", Price: 7,
			hand: func(_ *Modifier, ctx ScoreCtx, c int, mu float64) (string, int, float64) {
				if !containsPair(ctx.Type) {
					return "", c, mu
				}
				return "×2 mult", c, mu * 2
			},
		},
		{
			Name: "Last Call", Desc: "×3 mult on the last play of a round", Price: 6,
			hand: func(_ *Modifier, ctx ScoreCtx, c int, mu float64) (string, int, float64) {
				if ctx.PlaysLeft != 0 {
					return "", c, mu
				}
				return "×3 mult", c, mu * 3
			},
		},
		{
			Name: "Small Change", Desc: "+50 chips if every scoring card is 5 or lower", Price: 5,
			hand: func(_ *Modifier, ctx ScoreCtx, c int, mu float64) (string, int, float64) {
				for _, card := range ctx.Scoring {
					if card.Order() > 5 {
						return "", c, mu
					}
				}
				return "+50 chips", c + 50, mu
			},
		},
		{
			Name: "Stockpile", Desc: "+1 mult for every $3 you hold", Price: 5,
			hand: func(_ *Modifier, ctx ScoreCtx, c int, mu float64) (string, int, float64) {
				n := ctx.Money / 3
				if n == 0 {
					return "", c, mu
				}
				return fmt.Sprintf("+%d mult", n), c, mu + float64(n)
			},
		},
		{
			Name: "Wide Net", Desc: "+30 chips when you play 5 cards", Price: 4,
			hand: func(_ *Modifier, ctx ScoreCtx, c int, mu float64) (string, int, float64) {
				if ctx.Played != 5 {
					return "", c, mu
				}
				return "+30 chips", c + 30, mu
			},
		},
	}

	for s := 0; s < 4; s++ {
		mods = append(mods, Modifier{
			Name:  suitNames[s] + " Bias",
			Desc:  fmt.Sprintf("+3 mult for each scoring %s card", suitGlyphs[s]),
			Price: 5,
			card: func(_ *Modifier, c Card) (int, int) {
				if c.Suit == s {
					return 0, 3
				}
				return 0, 0
			},
		})
	}
	mods = append(mods, expandedCatalog()...)
	for i := range mods {
		mods[i].Price += 2 * mods[i].Rarity
	}
	return mods
}

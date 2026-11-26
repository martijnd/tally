package main

import "fmt"

var rarities = []string{"Common", "Uncommon", "Rare", "Legendary"}
var rarityWeight = []int{16, 8, 3, 1}

func expandedCatalog() []Modifier {
	var out []Modifier
	for rank := 1; rank <= 13; rank++ {
		name := []string{"", "Needle", "Fork", "Tripod", "Window", "Palm", "Hive", "Rainbow", "Octave", "Orbit", "Decimal", "Herald", "Regent", "Monarch"}[rank]
		for tier, kind := range []string{"Weight", "Dial", "Resonator"} {
			mod := Modifier{Name: name + " " + kind, Rarity: tier, Price: 4}
			label := Card{Rank: rank}.Label()
			switch tier {
			case 0:
				amount := 14 + rank
				mod.Desc = fmt.Sprintf("Scoring %s grants +%d chips", label, amount)
				mod.card = func(_ *Modifier, c Card) (int, int) {
					if c.Rank == rank {
						return amount, 0
					}
					return 0, 0
				}
			case 1:
				mod.Desc = "Scoring " + label + " grants +4 mult"
				mod.card = func(_ *Modifier, c Card) (int, int) {
					if c.Rank == rank {
						return 0, 4
					}
					return 0, 0
				}
			case 2:
				mod.Desc = "Retrigger scoring " + label
				mod.RepeatRank = rank
			}
			out = append(out, mod)
		}
	}
	for i, base := range handBases {
		name := []string{"Solitary", "Couplet", "Doublet", "Triangle", "Parade", "Palette", "Homestead", "Quartet", "Spectrum", "Quintet", "Mosaic", "Prism"}[i]
		for tier, kind := range []string{"Reserve", "Beacon", "Lens", "Chronicle"} {
			mod := Modifier{Name: name + " " + kind, Rarity: tier, Price: 4}
			mod.Desc = []string{"+60 chips", "+8 mult", "×2 mult", "+2 mult per previous play this run"}[tier] + " on " + base.Name
			mod.hand = func(_ *Modifier, ctx ScoreCtx, c, mu int) (string, int, int) {
				if int(ctx.Type) != i {
					return "", c, mu
				}
				switch tier {
				case 0:
					return "+60 chips", c + 60, mu
				case 1:
					return "+8 mult", c, mu + 8
				case 2:
					return "×2 mult", c, mu * 2
				}
				n := 2 * ctx.Previous
				return fmt.Sprintf("+%d mult", n), c, mu + n
			}
			out = append(out, mod)
		}
	}
	for _, trigger := range []struct{ Event, Name string }{{"discard", "Offcut"}, {"start", "Opening"}, {"end", "Closing"}, {"shop", "Market"}, {"reroll", "Turntable"}, {"buy", "Purchase"}, {"sell", "Sale"}, {"added", "Accession"}, {"destroyed", "Erasure"}, {"held", "Reserve"}} {
		for tier, kind := range []string{"Till", "Gauge", "Bond"} {
			mod := Modifier{Name: trigger.Name + " " + kind, Price: 4, Rarity: tier, Trigger: trigger.Event, Effect: []string{"cash", "grow", "value"}[tier]}
			mod.Desc = []string{"Gain $1", "Permanently gain +1 scoring mult", "Gain $1 sell value"}[tier] + " on " + trigger.Event
			if tier == 1 {
				mod.hand = func(m *Modifier, _ ScoreCtx, c, mu int) (string, int, int) {
					return fmt.Sprintf("+%d mult", m.Counter), c, mu + m.Counter
				}
			}
			out = append(out, mod)
		}
	}
	specials := []Modifier{
		{Name: "Counterproof", Desc: "Copy the scoring effect of the next modifier", Copy: true, Rarity: 3},
		{Name: "Loaded Die", Desc: "Double effect odds", Effect: "odds", Rarity: 2},
		{Name: "Night School", Desc: "Create a lesson at round start", Trigger: "start", Effect: "lesson", Rarity: 1},
		{Name: "School Trip", Desc: "Create a lesson on shop entry", Trigger: "shop", Effect: "lesson", Rarity: 1},
		{Name: "Appendix", Desc: "Add a random card at round end", Trigger: "end", Effect: "add", Rarity: 2},
		{Name: "Loose Insert", Desc: "Add a random card on reroll", Trigger: "reroll", Effect: "add", Rarity: 2},
		{Name: "Paper Mill", Desc: "Destroy a random undrawn card on discard", Trigger: "discard", Effect: "destroy", Rarity: 2},
		{Name: "Sieve", Desc: "Destroy a random undrawn card at shop entry", Trigger: "shop", Effect: "destroy", Rarity: 2},
	}
	for _, trigger := range []struct{ Name, Event string }{{"Pocket", "score"}, {"Rest", "held"}, {"Exchange", "sell"}, {"Foundry", "added"}, {"Cinder", "destroyed"}, {"Interval", "discard"}} {
		for i, kind := range []string{"Lottery", "Scholarship"} {
			effect := "lottery"
			desc := "1 in 4 chance for $4"
			if i == 1 {
				effect = "scholarship"
				desc = "1 in 4 chance to create a lesson"
			}
			specials = append(specials, Modifier{Name: trigger.Name + " " + kind, Desc: desc + " on " + trigger.Event, Trigger: trigger.Event, Effect: effect, Rarity: 1 + i})
		}
	}
	for _, mod := range specials {
		mod.Price = 5
		out = append(out, mod)
	}
	return out
}
func (m *model) odds() int {
	n := 1
	for _, mod := range m.mods {
		if mod.Effect == "odds" {
			n = min(1024, 2*n)
		}
	}
	return n
}
func (m *model) emit(event string) {
	if m.phase == phasePlay && (m.boss().Rule == "mute" || m.boss().Rule == "finalmute") {
		return
	}
	for _, mod := range m.mods {
		if mod.Trigger != event {
			continue
		}
		switch mod.Effect {
		case "cash":
			m.money++
		case "grow":
			mod.Counter++
		case "value":
			mod.ResaleGrowth++
		case "lottery":
			if m.itemRNG.IntN(4) < m.odds() {
				m.money += 4
			}
		case "lesson", "scholarship":
			if len(m.bag) < m.bagSlots() && (mod.Effect == "lesson" || m.itemRNG.IntN(4) < m.odds()) {
				m.bag = append(m.bag, m.lesson())
			}
		case "add":
			m.nextID++
			c := Card{ID: m.nextID, Rank: 1 + m.itemRNG.IntN(13), Suit: m.itemRNG.IntN(4)}
			m.stock = append(m.stock, c)
			m.deck = append(m.deck, c)
			m.emit("added")
		case "destroy":
			if len(m.deck) > 0 && len(m.stock) > 1 {
				i := m.itemRNG.IntN(len(m.deck))
				id := m.deck[i].ID
				m.deck = append(m.deck[:i], m.deck[i+1:]...)
				for j, c := range m.stock {
					if c.ID == id {
						m.stock = append(m.stock[:j], m.stock[j+1:]...)
						break
					}
				}
				m.emit("destroyed")
			}
		}
	}
}
func scoringSource(mods []*Modifier, i int) *Modifier {
	for i < len(mods)-1 && mods[i].Copy {
		i++
	}
	return mods[i]
}

package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"sort"
)

const (
	handSize         = 8
	maxSelect        = 5
	playsPerRound    = 4
	discardsPerRound = 3
	maxMods          = 5
	numRounds        = 24
	rerollCost       = 2
	shopOffers       = 3
	countSteps       = 14
)

type phase int

const (
	phasePlay phase = iota
	phaseScoring
	phaseRoundWon
	phaseShop
	phaseLost
	phaseWon
	phasePreview
	phaseStart
)

// Event is one step of a scoring breakdown, revealed one tick at a time.
type Event struct {
	Label   string
	Chips   int
	Mult    float64
	Card    int // hand index, or -1
	C       Card
	HasCard bool
	Mod     int // modifier index, or -1
}

type reward struct {
	label  string
	amount int
}

type model struct {
	targetingSlot                        int // One-based consumable slot, or zero when browsing.
	hidden, playedAnte                   map[int]bool
	roundTypes                           [len(handBases)]bool
	forcedID, disabledJoker, trackedAnte int
	oxHand                               HandType
	firstDraw, afterHand                 bool
	discoveries                          int

	lastSupply supply
	ectoUses   int
	lastHand   HandType

	prefs                                        preferences
	settings                                     bool
	collection                                   bool
	collectionGroup, collectionPage, unlockRound int
	profile                                      *profile
	deckID, difficulty, challengeID              int
	bestDifficulty                               [15]int
	discardBonus, playBonus, bagBonus            int
	tagQueue                                     []int
	doubleTags, freePacks, freeRerolls           int
	upgrades                                     [len(upgradeLines)]int
	upgradeOffer, packKind, packSize, packPicks  int
	packSold                                     bool
	packFrame                                    int
	packChoices                                  []packChoice
	packRNG                                      *rand.Rand
	viewingDeck, remainingDeck                   bool
	handBonus, slotBonus                         int
	stock                                        []Card
	nextID                                       int
	bag                                          []supply
	stockItem                                    *supply
	inventory                                    bool
	itemRNG                                      *rand.Rand
	bosses                                       []boss
	bossDisabled                                 bool
	firstType                                    HandType
	bestGain, skipped                            int
	seed                                         string
	deckRNG, shopRNG                             *rand.Rand
	phase                                        phase

	round, score, money int
	plays, discards     int
	deck, hand          []Card
	selected            []bool
	cursor              int
	bySuit              bool
	mods                []*Modifier

	// scoring animation
	events     []Event
	shown      int
	scoring    map[int]bool
	playedIdx  []int
	gain       int
	baseScore  int
	shownScore int
	step       int

	rewards []reward

	offers      []*Modifier
	shopCursor  int
	rerollPrice int

	levels, timesPlayed [len(handBases)]int
	overview            bool
	msg                 string
}

func newModel(seed string) *model {
	m := &model{
		seed: seed, deckRNG: runStream(seed, "deck"), shopRNG: runStream(seed, "shop"),
		money: 4, stock: newDeck(), nextID: 52, itemRNG: runStream(seed, "items"), packRNG: runStream(seed, "packs"),
	}
	m.bosses = append(m.bosses, m.nextBoss(0))
	m.phase = phaseStart
	return m
}

func (m *model) target() int {
	stage := m.round / 3

	curves := [3][8]int{{300, 800, 2000, 5000, 11000, 20000, 35000, 50000}, {300, 900, 2600, 8000, 20000, 36000, 60000, 100000}, {300, 1000, 3200, 9000, 25000, 60000, 110000, 200000}}
	curve := 0
	if m.difficulty >= 5 {
		curve = 2
	} else if m.difficulty >= 2 {
		curve = 1
	}
	base := float64(curves[curve][min(stage, 7)])
	if stage >= 8 {
		c := float64(stage - 7)
		base *= math.Pow(1.6+math.Pow(0.75*c, 1+0.2*c), c)
		if math.IsInf(base, 1) {
			return math.MaxInt
		}
		unit := math.Pow(10, math.Floor(math.Log10(base))-1)
		base = math.Floor(base/unit) * unit
	}
	if m.deckIs(13) || m.restriction("high targets") {
		base *= 2
	}
	factor := 1 + float64(m.round%3)/2
	if b := m.boss(); b.Rule == "target" {
		factor = float64(b.Amount)
	} else if b.Rule == "needle" {
		factor = 1
	}
	target := base * factor
	if target >= float64(math.MaxInt) {
		return math.MaxInt
	}
	return int(target)
}

func (m *model) startRound() {
	m.deck = append([]Card(nil), m.stock...)
	m.deckRNG.Shuffle(len(m.deck), func(i, j int) { m.deck[i], m.deck[j] = m.deck[j], m.deck[i] })
	m.hand, m.selected = nil, nil
	m.plays, m.discards = playsPerRound+m.upgrades[rehearsal]+m.playBonus, discardsPerRound+m.upgrades[wastebasket]+m.discardBonus
	if m.difficulty >= 4 {
		m.discards--
	}
	m.applyTags("round")
	m.firstType = -1
	m.firstDraw = true
	m.hidden = map[int]bool{}
	m.roundTypes = [len(handBases)]bool{}
	m.forcedID = 0
	m.disabledJoker = -1
	m.afterHand = false
	if m.playedAnte == nil || m.trackedAnte != m.round/3 {
		m.playedAnte = map[int]bool{}
		m.trackedAnte = m.round / 3
	}
	m.oxHand = HighCard
	for i, n := range m.timesPlayed {
		if n > m.timesPlayed[m.oxHand] {
			m.oxHand = HandType(i)
		}
	}
	b := m.boss()
	switch b.Rule {
	case "water":
		m.discards = 0
	case "needle":
		m.plays = 1
	case "heart":
		if len(m.mods) > 0 {
			m.disabledJoker = m.deckRNG.IntN(len(m.mods))
		}
	case "acorn":
		m.deckRNG.Shuffle(len(m.mods), func(i, j int) { m.mods[i], m.mods[j] = m.mods[j], m.mods[i] })
	}
	m.discards = max(0, m.discards)
	if m.restriction("no discards") {
		m.discards = 0
	}
	if m.restriction("single play") {
		m.plays = 1
	}
	m.score, m.shownScore = 0, 0
	m.events, m.shown, m.scoring = nil, 0, nil
	m.cursor = 0
	m.msg = ""
	m.draw()
	m.phase = phasePlay
	m.emit("start")
}

func (m *model) draw() {
	b := m.boss()
	drawn := 0
	for len(m.deck) > 0 && ((b.Rule == "serpent" && !m.firstDraw && drawn < 3) || ((b.Rule != "serpent" || m.firstDraw) && len(m.hand) < m.drawLimit())) {
		drawn++
		c := m.deck[0]
		if (b.Rule == "house" && m.firstDraw) || (b.Rule == "fish" && m.afterHand) || (b.Rule == "mark" && c.Rank >= 11) || (b.Rule == "wheel" && m.deckRNG.IntN(7) < m.odds()) {
			m.hidden[c.ID] = true
		}
		m.hand = append(m.hand, m.deck[0])
		m.selected = append(m.selected, false)
		m.deck = m.deck[1:]
	}
	m.sortHand()
	m.firstDraw = false
	m.afterHand = false
	m.forceCard()
}

func (m *model) sortHand() {
	idx := make([]int, len(m.hand))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		ca, cb := m.hand[idx[a]], m.hand[idx[b]]
		if m.bySuit && ca.Suit != cb.Suit {
			return ca.Suit < cb.Suit
		}
		if ca.Order() != cb.Order() {
			return ca.Order() > cb.Order()
		}
		return ca.Suit < cb.Suit
	})
	hand := make([]Card, len(idx))
	sel := make([]bool, len(idx))
	cursor := m.cursor
	for ni, oi := range idx {
		hand[ni], sel[ni] = m.hand[oi], m.selected[oi]
		if oi == m.cursor {
			cursor = ni
		}
	}
	m.hand, m.selected, m.cursor = hand, sel, cursor
	m.clampCursor()
}

func (m *model) clampCursor() {
	if m.cursor >= len(m.hand) {
		m.cursor = len(m.hand) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *model) selectedIdx() []int {
	var out []int
	for i, s := range m.selected {
		if s {
			out = append(out, i)
		}
	}
	return out
}

func (m *model) selectedCards() []Card {
	var out []Card
	for _, i := range m.selectedIdx() {
		out = append(out, m.hand[i])
	}
	return out
}

func (m *model) removeSelected() {
	var hand []Card
	var sel []bool
	for i, c := range m.hand {
		if !m.selected[i] {
			hand = append(hand, c)
			sel = append(sel, false)
		}
	}
	m.hand, m.selected = hand, sel
	m.clampCursor()
}

func (m *model) play() bool {
	sel := m.selectedIdx()
	if len(sel) == 0 {
		m.msg = "select at least one card"
		return false
	}
	m.plays--
	played := m.selectedCards()
	ht, local := evaluate(played)

	b := m.boss()
	valid := local[:0]
	for _, i := range local {
		if !m.cardDebuffed(played[i]) {
			valid = append(valid, i)
		}
	}
	local = valid
	if m.firstType < 0 {
		m.firstType = ht
	}
	mods := m.mods
	if b.Rule == "heart" && m.disabledJoker >= 0 && m.disabledJoker < len(mods) {
		mods = slices.Clone(mods)
		mods[m.disabledJoker] = &Modifier{}
	}
	scoringIdx := make([]int, len(local))
	scoringCards := make([]Card, len(local))
	m.scoring = map[int]bool{}
	for i, li := range local {
		scoringIdx[i] = sel[li]
		scoringCards[i] = played[li]
		m.scoring[sel[li]] = true
	}

	held := 0
	for i, c := range m.hand {
		if !m.selected[i] && !m.cardDebuffed(c) {
			m.emit("held")
			if c.Enhancement == 3 {
				held++
				if c.Seal == 1 {
					held++
				}
			}

		}
	}
	onScore := func(c Card) {
		m.emit("score")
		if c.Seal == 2 {
			m.money += 3
		}

	}
	m.lastHand = ht
	previousType := m.roundTypes[ht]
	m.roundTypes[ht] = true
	if b.Rule == "arm" {
		m.levels[ht] = max(0, m.levels[ht]-1)
	}
	if b.Rule == "tooth" {
		m.money -= len(sel)
	}
	if b.Rule == "ox" && ht == m.oxHand {
		m.money = 0
	}
	m.timesPlayed[ht]++
	ctx := ScoreCtx{Flint: b.Rule == "flint", Plasma: m.deckIs(13), Lucky: func() float64 {
		mult := 0.
		if m.itemRNG.IntN(5) < m.odds() {
			mult = 20
		}
		if m.itemRNG.IntN(15) < m.odds() {
			m.money += 20
		}
		return mult
	}, Previous: m.timesPlayed[ht] - 1, OnScore: onScore, HeldPatient: held, Training: m.levels[ht], Type: ht, Scoring: scoringCards, Played: len(played), PlaysLeft: m.plays, Money: m.money}
	m.events, m.gain = scoreHand(m.hand, scoringIdx, ht, mods, ctx)
	if (b.Rule == "eye" && previousType) || (b.Rule == "lock" && ht != m.firstType) || (b.Rule == "minimum" && len(sel) < b.Amount) {
		m.gain = 0
		m.events = append(m.events, Event{Label: b.Name, Card: -1, Mod: -1})
	}
	m.bestGain = max(m.bestGain, m.gain)
	m.shown, m.step = 0, 0
	m.baseScore = m.score
	m.msg = ""
	m.phase = phaseScoring
	return true
}

func scoreHand(hand []Card, scoring []int, ht HandType, mods []*Modifier, ctx ScoreCtx) ([]Event, int) {
	b := trainedBase(ht, ctx.Training)

	chips, mult := b.Chips, float64(b.Mult)
	if ctx.Flint {
		chips = (chips + 1) / 2
		mult = math.Ceil(mult / 2)
	}
	ev := []Event{{Label: b.Name, Chips: chips, Mult: mult, Card: -1, Mod: -1}}

	for _, hi := range scoring {
		c := hand[hi]
		repeats := 1
		if c.Seal == 1 {
			repeats++
		}
		for i := range mods {
			if scoringSource(mods, i).RepeatRank == c.Rank && c.Enhancement != 7 {
				repeats++
			}
		}
		for range repeats {
			if ctx.OnScore != nil {
				ctx.OnScore(c)
			}
			chips += c.Chips()
			chips, mult = editionScore(c.Edition, chips, mult)
			switch c.Enhancement {
			case 1:
				chips += 30
			case 2:
				mult += 4
			case 8:
				if ctx.Lucky != nil {
					mult += ctx.Lucky()
				}
			case 5:
				mult *= 2
			}
			ev = append(ev, Event{Label: fmt.Sprintf("+%d chips", c.Chips()), Chips: chips, Mult: mult, Card: hi, C: c, HasCard: true, Mod: -1})
			for mi, mod := range mods {
				source := scoringSource(mods, mi)
				if source.card == nil || c.Enhancement == 7 {
					continue
				}
				dc, dm := source.card(source, c)
				if dc == 0 && dm == 0 {
					continue
				}
				chips += dc
				mult += float64(dm)
				ev = append(ev, Event{Label: mod.Name + " " + deltaLabel(dc, dm), Chips: chips, Mult: mult, Card: hi, C: c, HasCard: true, Mod: mi})
			}
		}

	}
	for range ctx.HeldPatient {
		factor := 1.5

		mult *= factor
		ev = append(ev, Event{Label: fmt.Sprintf("Steel ×%g", factor), Chips: chips, Mult: mult, Card: -1, Mod: -1})
	}
	for mi, mod := range mods {
		if mod.Edition > 0 && mod.Edition < 4 {
			chips, mult = editionScore(mod.Edition, chips, mult)
			ev = append(ev, Event{Label: editions[mod.Edition] + " " + mod.Name, Chips: chips, Mult: mult, Card: -1, Mod: mi})
		}
		source := scoringSource(mods, mi)
		if source.hand == nil {
			continue
		}
		label, c2, m2 := source.hand(source, ctx, chips, mult)
		if label == "" {
			continue
		}
		chips, mult = c2, m2
		ev = append(ev, Event{Label: mod.Name + " " + label, Chips: chips, Mult: mult, Card: -1, Mod: mi})
	}

	for _, mod := range mods {
		if mod.after != nil {
			mod.after(mod, ctx)
		}
	}
	total := float64(chips) * mult
	if ctx.Plasma {
		balanced := math.Floor((float64(chips) + mult) / 2)
		total = balanced * balanced
	}
	if total >= float64(int(^uint(0)>>1)) {
		return ev, int(^uint(0) >> 1)
	}
	return ev, int(total)
}

func deltaLabel(dc, dm int) string {
	switch {
	case dc != 0 && dm != 0:
		return fmt.Sprintf("+%d chips +%d mult", dc, dm)
	case dc != 0:
		return fmt.Sprintf("+%d chips", dc)
	}
	return fmt.Sprintf("+%d mult", dm)
}

func (m *model) finishPlay() {
	m.score = addScore(m.baseScore, m.gain)
	m.shownScore = m.score
	for i := range m.hand {
		c := m.hand[i]
		if m.scoring[i] && c.Enhancement == 5 && len(m.stock) > 1 && m.itemRNG.IntN(4) < m.odds() {
			m.stock = slices.DeleteFunc(m.stock, func(other Card) bool { return other.ID == c.ID })
			m.emit("destroyed")
		}
	}
	m.scoring = nil
	if m.score >= m.target() {
		for i, c := range m.hand {
			if m.selected[i] || m.cardDebuffed(c) {
				continue
			}
			repeats := 1
			if c.Seal == 1 {
				repeats++
			}
			if c.Enhancement == 6 {
				m.money += 3 * repeats
			}
			if c.Seal == 3 && len(m.bag) < m.bagSlots() {
				for _, it := range supplies() {
					if it.Action == "level" && it.Value == int(m.lastHand) {
						m.bag = append(m.bag, it)
						break
					}
				}
			}
		}
	}
	m.removeSelected()
	for range min(m.boss().Amount, len(m.hand)) {
		if m.boss().Rule != "shed" {
			break
		}
		m.selected[m.deckRNG.IntN(len(m.hand))] = true
		m.removeSelected()
	}
	m.afterHand = true
	if m.boss().Rule == "heart" && len(m.mods) > 0 {
		m.disabledJoker = m.deckRNG.IntN(len(m.mods))
	}
	m.draw()

	switch {
	case m.score >= m.target():
		m.winRound()
	case m.plays == 0 || len(m.hand) == 0:
		m.phase = phaseLost
	default:
		m.phase = phasePlay
	}
}

func (m *model) discard() {
	sel := m.selectedIdx()
	switch {
	case len(sel) == 0:
		m.msg = "select cards to discard"
		return
	case m.discards == 0:
		m.msg = "no discards left this round"
		return
	}
	for _, i := range sel {
		if m.hand[i].Seal == 4 && len(m.bag) < m.bagSlots() {
			m.bag = append(m.bag, m.randomSupply("Tarot", ""))
		}
	}
	m.emit("discard")
	m.discards--
	m.removeSelected()
	m.draw()
	m.msg = fmt.Sprintf("discarded %d", len(sel))
}

func (m *model) winRound() {
	m.rewards = []reward{{"round cleared", m.roundReward()}}
	if m.deckIs(12) && m.round%3 == 2 {
		m.doubleTags++
	}
	if m.deckIs(3) && m.discards > 0 {
		m.rewards = append(m.rewards, reward{"unused discards", m.discards})
	}
	handPay := 1 + m.upgrades[patience]
	if m.deckIs(3) {
		handPay++
	}
	if m.plays > 0 {
		m.rewards = append(m.rewards, reward{fmt.Sprintf("%d unused hands", m.plays), m.plays * handPay})
	}
	divisor := 5

	if interest := min(m.money/divisor, 5+5*m.upgrades[savings]); interest > 0 && !m.deckIs(3) && !m.restriction("no interest") {
		m.rewards = append(m.rewards, reward{fmt.Sprintf("interest ($1 per $%d)", divisor), interest})
	}
	for _, r := range m.rewards {
		m.money += r.amount
	}
	m.emit("end")
	m.settleStickers()
	m.applyTags("win")
	m.phase = phaseRoundWon
}

func (m *model) openShop() {
	m.hand, m.selected = nil, nil
	m.cursor = 0
	items := supplies()
	for {
		it := items[m.shopRNG.IntN(len(items))]
		if it.cardFamily() != "Spectral" || (m.deckIs(7) && m.shopRNG.IntN(6) == 0 && it.Name != "Black Hole" && it.Name != "The Soul") {
			m.stockItem = &it
			break
		}
	}
	m.emit("shop")
	m.prepareMarket()
	m.applyTags("shop")
	m.applyTags("pack")
	m.rerollPrice = rerollCost - m.upgrades[coupons]
	m.rollOffers()
	m.shopCursor = 0
	m.msg = ""
	m.phase = phaseShop
}

func (m *model) rollOffers() {
	owned := map[string]bool{}
	for _, mod := range m.mods {
		owned[mod.Name] = true
	}
	var pool []Modifier
	for _, mod := range catalog() {
		if !owned[mod.Name] && m.modifierUnlocked(mod.Name) {
			pool = append(pool, mod)
		}
	}

	m.offers = nil
	for len(m.offers) < shopOffers+m.upgrades[shelf] && len(pool) > 0 {
		total := 0
		for _, mod := range pool {
			total += rarityWeight[mod.Rarity]
		}
		draw := m.shopRNG.IntN(total)
		i := 0
		for draw >= rarityWeight[pool[i].Rarity] {
			draw -= rarityWeight[pool[i].Rarity]
			i++
		}
		mod := pool[i]
		pool = append(pool[:i], pool[i+1:]...)
		mod.Desc = rarities[mod.Rarity] + " · " + mod.Desc
		if m.shopRNG.IntN(8) <= m.upgrades[gilding] {
			mod.Edition = 1 + m.shopRNG.IntN(4)
			mod.Price += mod.Edition
			mod.Desc = editions[mod.Edition] + " edition · " + mod.Desc
		}
		m.stickerOffer(&mod)
		mod.Price = max(1, mod.Price-m.upgrades[wholesale])
		m.offers = append(m.offers, &mod)
	}
	m.applyTags("offer")
}

func sellPrice(mod *Modifier) int { return max(mod.Price/2, 1) + mod.ResaleGrowth }

func (m *model) shopRowCount() int {
	n := len(m.offers) + len(m.mods)
	if !m.packSold {
		n++
	}
	return n
}

func (m *model) packSelected() bool {
	return !m.packSold && m.shopCursor == len(m.offers)+len(m.mods)
}

func (m *model) shopSelect() {
	if m.packSelected() {
		m.openPack()
		return
	}
	if m.shopCursor < len(m.offers) {
		mod := m.offers[m.shopCursor]
		switch {
		case len(m.mods) >= m.modSlots() && mod.Edition != 4:
			m.msg = "all slots full, sell one first"
		case m.money < mod.Price:
			m.msg = "not enough money"
		default:
			m.money -= mod.Price
			m.mods = append(m.mods, mod)
			m.offers = append(m.offers[:m.shopCursor], m.offers[m.shopCursor+1:]...)
			m.emit("buy")
			m.msg = "bought " + mod.Name
		}
	} else {
		i := m.shopCursor - len(m.offers)
		mod := m.mods[i]
		if mod.Sticker == "Bound" {
			m.msg = "Eternal Jokers cannot be sold"
			return
		}
		m.money += sellPrice(mod) + m.upgrades[resale]
		m.mods = append(m.mods[:i], m.mods[i+1:]...)
		if m.boss().Rule == "leaf" {
			m.bossDisabled = true
		}
		m.emit("sell")
		m.msg = "sold " + mod.Name
	}
	m.clampShopCursor()
}

func (m *model) clampShopCursor() {
	n := m.shopRowCount()
	if m.shopCursor >= n {
		m.shopCursor = n - 1
	}
	if m.shopCursor < 0 {
		m.shopCursor = 0
	}
}

func trainedBase(ht HandType, training int) handBase {
	b := handBases[ht]
	b.Chips += training * planetGrowth[ht].chips
	b.Mult += training * planetGrowth[ht].mult
	return b
}

func addScore(a, b int) int {
	limit := int(^uint(0) >> 1)
	if b > limit-a {
		return limit
	}
	return a + b
}
func multiplyScore(a, b int) int {
	limit := int(^uint(0) >> 1)
	if a > 0 && b > limit/a {
		return limit
	}
	return a * b
}

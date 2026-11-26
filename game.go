package main

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

var targets = [8]int{300, 500, 800, 1200, 1800, 2700, 4000, 6000}

type phase int

const (
	phasePlay phase = iota
	phaseScoring
	phaseRoundWon
	phaseShop
	phaseLost
	phaseWon
	phasePreview
)

// Event is one step of a scoring breakdown, revealed one tick at a time.
type Event struct {
	Label       string
	Chips, Mult int // running totals after this step
	Card        int // hand index, or -1
	C           Card
	HasCard     bool
	Mod         int // modifier index, or -1
}

type reward struct {
	label  string
	amount int
}

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(170*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

type model struct {
	bosses            []boss
	bossDisabled      bool
	firstType         HandType
	bestGain, skipped int
	seed              string
	deckRNG, shopRNG  *rand.Rand
	phase             phase

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
	cursorHand          int
	overview            bool
	msg                 string
}

func newModel(seed string) *model {
	m := &model{
		seed: seed, deckRNG: runStream(seed, "deck"), shopRNG: runStream(seed, "shop"),
		money: 4,
	}
	rng := runStream(seed, "boss")
	for _, i := range rng.Perm(len(regularBosses)) {
		m.bosses = append(m.bosses, regularBosses[i])
	}
	m.bosses = append(m.bosses[:7:7], append([]boss{finalBosses[rng.IntN(len(finalBosses))]}, m.bosses[7:]...)...)
	m.phase = phasePreview
	return m
}

func (m *model) target() int {
	stage := m.round / 3
	base := targets[min(stage, 7)]
	for i := 7; i < stage; i++ {
		if base > int(^uint(0)>>1)/16 {
			return int(^uint(0) >> 1)
		}
		base *= 4
	}
	return base * (2 + m.round%3) / 2
}

func (m *model) startRound() {
	m.deck = newDeck()
	m.deckRNG.Shuffle(len(m.deck), func(i, j int) { m.deck[i], m.deck[j] = m.deck[j], m.deck[i] })
	m.hand, m.selected = nil, nil
	m.plays, m.discards = playsPerRound, discardsPerRound
	m.firstType = -1
	b := m.boss()
	switch b.Rule {
	case "plays", "finalmute":
		m.plays--
	case "discards":
		m.discards -= b.Amount
	}
	m.score, m.shownScore = 0, 0
	m.events, m.shown, m.scoring = nil, 0, nil
	m.cursor = 0
	m.msg = ""
	m.draw()
	m.phase = phasePlay
}

func (m *model) draw() {
	limit := handSize
	if b := m.boss(); b.Rule == "size" {
		limit -= b.Amount
	}
	for len(m.hand) < limit && len(m.deck) > 0 {
		m.hand = append(m.hand, m.deck[0])
		m.selected = append(m.selected, false)
		m.deck = m.deck[1:]
	}
	m.sortHand()
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

func (m *model) play() tea.Cmd {
	sel := m.selectedIdx()
	if len(sel) == 0 {
		m.msg = "select at least one card"
		return nil
	}
	m.plays--
	played := m.selectedCards()
	ht, local := evaluate(played)

	b := m.boss()
	valid := local[:0]
	for _, i := range local {
		if !b.debuffs(played[i]) {
			valid = append(valid, i)
		}
	}
	local = valid
	if m.firstType < 0 {
		m.firstType = ht
	}
	mods := m.mods
	if b.Rule == "mute" || b.Rule == "finalmute" {
		mods = nil
	}
	if b.Rule == "toll" {
		m.money = max(0, m.money-b.Amount)
	}
	scoringIdx := make([]int, len(local))
	scoringCards := make([]Card, len(local))
	m.scoring = map[int]bool{}
	for i, li := range local {
		scoringIdx[i] = sel[li]
		scoringCards[i] = played[li]
		m.scoring[sel[li]] = true
	}

	m.timesPlayed[ht]++
	ctx := ScoreCtx{Training: m.levels[ht], Type: ht, Scoring: scoringCards, Played: len(played), PlaysLeft: m.plays, Money: m.money}
	m.events, m.gain = scoreHand(m.hand, scoringIdx, ht, mods, ctx)
	if (b.Rule == "lock" && ht != m.firstType) || (b.Rule == "late" && m.plays == playsPerRound-1) || (b.Rule == "limit" && len(sel) > b.Amount) || (b.Rule == "minimum" && len(sel) < b.Amount) || (b.Rule == "ban" && int(ht) == b.Amount) {
		m.gain = 0
		m.events = append(m.events, Event{Label: b.Name, Card: -1, Mod: -1})
	}
	m.bestGain = max(m.bestGain, m.gain)
	m.shown, m.step = 0, 0
	m.baseScore = m.score
	m.msg = ""
	m.phase = phaseScoring
	return tick()
}

func scoreHand(hand []Card, scoring []int, ht HandType, mods []*Modifier, ctx ScoreCtx) ([]Event, int) {
	b := trainedBase(ht, ctx.Training)
	chips, mult := b.Chips, b.Mult
	ev := []Event{{Label: b.Name, Chips: chips, Mult: mult, Card: -1, Mod: -1}}

	for _, hi := range scoring {
		c := hand[hi]
		chips += c.Chips()
		ev = append(ev, Event{Label: fmt.Sprintf("+%d chips", c.Chips()), Chips: chips, Mult: mult, Card: hi, C: c, HasCard: true, Mod: -1})
		for mi, mod := range mods {
			if mod.card == nil {
				continue
			}
			dc, dm := mod.card(mod, c)
			if dc == 0 && dm == 0 {
				continue
			}
			chips += dc
			mult += dm
			ev = append(ev, Event{Label: mod.Name + " " + deltaLabel(dc, dm), Chips: chips, Mult: mult, Card: hi, C: c, HasCard: true, Mod: mi})
		}
	}

	for mi, mod := range mods {
		if mod.hand == nil {
			continue
		}
		label, c2, m2 := mod.hand(mod, ctx, chips, mult)
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
	return ev, chips * mult
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

func (m *model) advanceScoring() tea.Cmd {
	if m.shown < len(m.events) {
		m.shown++
		return tick()
	}
	if m.step < countSteps {
		m.step++
		m.shownScore = m.baseScore + m.gain*m.step/countSteps
		return tick()
	}
	m.finishPlay()
	return nil
}

func (m *model) finishPlay() {
	m.score = m.baseScore + m.gain
	m.shownScore = m.score
	m.scoring = nil
	m.removeSelected()
	if m.boss().Rule == "shed" && len(m.hand) > 0 {
		m.selected[m.deckRNG.IntN(len(m.hand))] = true
		m.removeSelected()
	}
	m.draw()

	switch {
	case m.score >= m.target():
		m.winRound()
	case m.plays == 0:
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
	m.discards--
	m.removeSelected()
	m.draw()
	m.msg = fmt.Sprintf("discarded %d", len(sel))
}

func (m *model) winRound() {
	m.rewards = []reward{{"round cleared", 3 + m.round%3}}
	if m.plays > 0 {
		m.rewards = append(m.rewards, reward{fmt.Sprintf("%d unused plays", m.plays), m.plays})
	}
	if interest := min(m.money/5, 5); interest > 0 {
		m.rewards = append(m.rewards, reward{"interest ($1 per $5)", interest})
	}
	for _, r := range m.rewards {
		m.money += r.amount
	}
	m.phase = phaseRoundWon
}

func (m *model) openShop() {
	m.rerollPrice = rerollCost
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
		if !owned[mod.Name] {
			pool = append(pool, mod)
		}
	}
	m.shopRNG.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	m.offers = nil
	for i := 0; i < shopOffers && i < len(pool); i++ {
		mod := pool[i]
		m.offers = append(m.offers, &mod)
	}
}

func sellPrice(mod *Modifier) int { return max(mod.Price/2, 1) }

func (m *model) shopSelect() {
	if m.shopCursor < len(m.offers) {
		mod := m.offers[m.shopCursor]
		switch {
		case len(m.mods) >= maxMods:
			m.msg = "all slots full, sell one first"
		case m.money < mod.Price:
			m.msg = "not enough money"
		default:
			m.money -= mod.Price
			m.mods = append(m.mods, mod)
			m.offers = append(m.offers[:m.shopCursor], m.offers[m.shopCursor+1:]...)
			m.msg = "bought " + mod.Name
		}
	} else {
		i := m.shopCursor - len(m.offers)
		mod := m.mods[i]
		m.money += sellPrice(mod)
		m.mods = append(m.mods[:i], m.mods[i+1:]...)
		m.msg = "sold " + mod.Name
	}
	m.clampShopCursor()
}

func (m *model) clampShopCursor() {
	n := len(m.offers) + len(m.mods)
	if m.shopCursor >= n {
		m.shopCursor = n - 1
	}
	if m.shopCursor < 0 {
		m.shopCursor = 0
	}
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.phase == phaseScoring {
			return m, m.advanceScoring()
		}
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" || key == "q" {
			return m, tea.Quit
		}
		if key == "h" && m.phase != phaseScoring {
			m.overview = !m.overview
			return m, nil
		}
		if m.overview {
			if key == "enter" && m.phase == phaseShop && m.money >= 4 {
				m.money -= 4
				m.levels[m.cursorHand]++
			}
			if key == "down" {
				m.cursorHand = (m.cursorHand + 1) % len(handBases)
			}
			if key == "up" {
				m.cursorHand = (m.cursorHand + len(handBases) - 1) % len(handBases)
			}
			return m, nil
		}
		switch m.phase {
		case phasePreview:
			if key == "enter" {
				m.startRound()
			} else if key == "s" && m.round%3 != 2 {
				m.money += 5
				m.skipped++
				m.round++
			}
		case phasePlay:
			return m, m.updatePlay(key)
		case phaseRoundWon:
			if key == "enter" || key == " " {
				if m.round == numRounds-1 {
					m.phase = phaseWon
				} else {
					m.openShop()
				}
			}
		case phaseShop:
			m.updateShop(key)
		case phaseLost, phaseWon:
			if key == "e" && m.phase == phaseWon {
				m.openShop()
			}
			if key == "n" {
				*m = *newModel(m.seed)
			}
		}
	}
	return m, nil
}

func (m *model) updatePlay(key string) tea.Cmd {
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.hand)-1 {
			m.cursor++
		}
	case " ", "x":
		if len(m.hand) == 0 {
			return nil
		}
		if !m.selected[m.cursor] && len(m.selectedIdx()) >= maxSelect {
			m.msg = fmt.Sprintf("you can select up to %d", maxSelect)
			return nil
		}
		m.selected[m.cursor] = !m.selected[m.cursor]
		m.msg = ""
	case "p", "enter":
		return m.play()
	case "d":
		m.discard()
	case "[", "]":
		dest := m.cursor - 1
		if key == "]" {
			dest = m.cursor + 1
		}
		if dest >= 0 && dest < len(m.hand) {
			m.hand[m.cursor], m.hand[dest] = m.hand[dest], m.hand[m.cursor]
			m.selected[m.cursor], m.selected[dest] = m.selected[dest], m.selected[m.cursor]
			m.cursor = dest
		}
	case "s":
		m.bySuit = !m.bySuit
		m.sortHand()
	case "c":
		for i := range m.selected {
			m.selected[i] = false
		}
	}
	return nil
}

func (m *model) updateShop(key string) {
	switch key {
	case "up", "k":
		if m.shopCursor > 0 {
			m.shopCursor--
		}
	case "down", "j":
		if m.shopCursor < len(m.offers)+len(m.mods)-1 {
			m.shopCursor++
		}
	case "enter", " ":
		if len(m.offers)+len(m.mods) > 0 {
			m.shopSelect()
		}
	case "r":
		if m.money < m.rerollPrice {
			m.msg = "not enough money to reroll"
			return
		}
		m.money -= m.rerollPrice
		m.rerollPrice++
		m.rollOffers()
		m.clampShopCursor()
		m.msg = "rerolled"
	case "c", "n":
		m.round++
		m.bossDisabled = false
		m.phase = phasePreview
	}
}

func trainedBase(ht HandType, training int) handBase {
	b := handBases[ht]
	b.Chips += training * (5 + int(ht)*2)
	b.Mult += training * (1 + int(ht)/4)
	return b
}

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	suitGlyphs = [4]string{"♠", "♥", "♦", "♣"}
	suitNames  = [4]string{"Spade", "Heart", "Diamond", "Club"}
	suitColors = [4]lipgloss.Color{"#818cf8", "#fb7185", "#fbbf24", "#34d399"}

	cText   = lipgloss.AdaptiveColor{Light: "#1f2937", Dark: "#e5e7eb"}
	cDim    = lipgloss.AdaptiveColor{Light: "#9ca3af", Dark: "#6b7280"}
	cBorder = lipgloss.AdaptiveColor{Light: "#d1d5db", Dark: "#374151"}
	cAccent = lipgloss.Color("#38bdf8")
	cChips  = lipgloss.Color("#7dd3fc")
	cMult   = lipgloss.Color("#f472b6")
	cGood   = lipgloss.Color("#4ade80")
	cWarn   = lipgloss.Color("#fbbf24")
	cBad    = lipgloss.Color("#f87171")

	sText   = lipgloss.NewStyle().Foreground(cText)
	sBold   = lipgloss.NewStyle().Foreground(cText).Bold(true)
	sDim    = lipgloss.NewStyle().Foreground(cDim)
	sAccent = lipgloss.NewStyle().Foreground(cAccent).Bold(true)
	sChips  = lipgloss.NewStyle().Foreground(cChips).Bold(true)
	sMult   = lipgloss.NewStyle().Foreground(cMult).Bold(true)
	sWarn   = lipgloss.NewStyle().Foreground(cWarn)
	sGood   = lipgloss.NewStyle().Foreground(cGood).Bold(true)
	sBad    = lipgloss.NewStyle().Foreground(cBad).Bold(true)

	panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cBorder).
		Padding(0, 1)
)

const (
	leftW  = 26 // inner content widths
	rightW = 34
	bodyH  = 13
)

func commas(n int) string {
	if n >= 1e15 || n <= -1e15 {
		return strings.Replace(fmt.Sprintf("%.2g", float64(n)), "e+", "e", 1)
	}
	s := strconv.Itoa(n)
	if n < 0 {
		return "-" + commas(-n)
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func spread(left, right string, w int) string {
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func padLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	for len(lines) < n {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func suit(c Card) string {
	if c.Enhancement == 7 {
		return "·"
	}
	return lipgloss.NewStyle().Foreground(suitColors[c.Suit]).Render(suitGlyphs[c.Suit])
}

func cardTag(c Card) string {
	return sBold.Render(fmt.Sprintf("%2s", c.Label())) + " " + suit(c)
}

func pips(n, total int, color lipgloss.TerminalColor) string {
	on := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("●", n))
	off := sDim.Render(strings.Repeat("○", total-n))
	return on + off
}

func (m *model) currentEvent() (Event, bool) {
	if m.phase != phaseScoring || m.shown == 0 || m.shown > len(m.events) {
		return Event{}, false
	}
	return m.events[m.shown-1], true
}

func (m *model) View() string {
	lipgloss.SetHasDarkBackground(m.prefs.Theme == 2 || (m.prefs.Theme == 0 && nativeDark))
	if m.settings {
		return m.settingsView()
	}
	if m.collection {
		return m.collectionView()
	}
	if len(m.packChoices) > 0 {
		return m.packView()
	}
	if m.viewingDeck {
		return m.deckView()
	}
	if m.inventory {
		return m.inventoryView()
	}
	if m.overview {
		var b strings.Builder
		fmt.Fprintf(&b, "Seed %q%s\n", m.seed, m.tagSummary())
		for i, n := range m.upgrades {
			if n > 0 {
				fmt.Fprintf(&b, "%s %d · ", upgradeLines[i].Name, n)
			}
		}
		b.WriteString("\nBoss Blinds · ")
		for i := 0; i <= m.round/3; i++ {
			boss := m.stageBoss(i)
			fmt.Fprintf(&b, "%d %s  ", i+1, boss.Name)
		}
		b.WriteString("\n")
		b.WriteString("Hand overview · g to close\n")
		for i := range handBases {
			base := trainedBase(HandType(i), m.levels[i])
			fmt.Fprintf(&b, "%-18s Lv %d   %d × %d   played %d\n", base.Name, 1+m.levels[i], base.Chips, base.Mult, m.timesPlayed[i])
		}
		return panel.Padding(1, 2).Render(b.String())
	}
	switch m.phase {
	case phaseStart:
		return m.startView()
	case phasePreview:
		return m.preview()
	case phaseRoundWon:
		return m.viewRoundWon()
	case phaseShop:
		return m.viewShop()
	case phaseLost, phaseWon:
		return m.viewEnd()
	}
	return m.viewTable()
}

func (m *model) viewTable() string {
	left := panel.Width(leftW + 2).Render(padLines(m.viewHand(), bodyH))
	right := panel.Width(rightW + 2).Render(padLines(m.viewScore(), bodyH))
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	w := lipgloss.Width(body)

	header := spread(
		sAccent.Render(" tally")+sDim.Render(fmt.Sprintf("  Ante %d/8 · %s", m.round/3+1, roundNames[m.round%3]+" "+m.boss().Name)),
		sDim.Render("target ")+sBold.Render(commas(m.target()))+" ",
		w,
	)

	if b := m.boss(); b.Rule != "" {
		header += "\n " + sWarn.Render(b.Effect)
	}
	msg := ""
	if m.msg != "" {
		msg = " " + sWarn.Render(m.msg)
	}
	help := sDim.Render(" w/s move · space select · a play · d discard · 1/2 shift · t sort · c clear · v deck · z collection · f consumables · g hands · , settings · ctrl+n new game · q quit")
	return lipgloss.JoinVertical(lipgloss.Left, header, body, m.viewMods(w), msg, help)
}

func (m *model) viewHand() string {
	var b strings.Builder
	b.WriteString(spread(sDim.Render("HAND"), sDim.Render(fmt.Sprintf("deck %d", len(m.deck))), leftW))
	b.WriteString("\n\n")

	active, hasActive := m.currentEvent()
	for i, c := range m.hand {
		cur := "  "
		if i == m.cursor && m.phase == phasePlay {
			cur = sAccent.Render("› ")
		}

		indent, bar := "", " "
		rankStyle := sText
		if m.selected[i] {
			indent, bar = "  ", sAccent.Render("┃")
			rankStyle = sBold
		}
		if hasActive && active.HasCard && active.Card == i {
			rankStyle = rankStyle.Reverse(true)
		}

		tail := ""
		if m.phase == phaseScoring && m.selected[i] {
			if m.scoring[i] {
				tail = sChips.Render(fmt.Sprintf("+%d", c.Chips()))
			} else {
				tail = sDim.Render("—")
			}
		}

		row := cur + bar + indent + rankStyle.Render(fmt.Sprintf("%2s", c.Label())) + " " + suit(c)
		if m.cardDebuffed(c) {
			tail = "debuffed"
		}
		if m.faceDown(c) && m.phase == phasePlay {
			row = cur + bar + indent + "??"
		}
		b.WriteString(spread(row, tail, leftW))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(sDim.Render("hands ") + pips(m.plays, max(m.plays, playsPerRound+m.upgrades[rehearsal]), cAccent) +
		sDim.Render("   discards ") + pips(m.discards, max(m.discards, discardsPerRound+m.upgrades[wastebasket]), cWarn))
	return b.String()
}

func (m *model) viewScore() string {
	var lines []string
	lines = append(lines, sDim.Render("SCORE"), "")

	sel := m.selectedCards()
	if m.phase == phasePlay {
		for _, c := range sel {
			if m.faceDown(c) {
				sel = nil
				break
			}
		}
	}
	const window = 5

	switch {
	case m.phase == phaseScoring || (len(sel) == 0 && m.events != nil):
		evs := m.events[:m.shown]
		if len(evs) > 0 {
			lines = append(lines, sBold.Render(evs[0].Label))
			rest := evs[1:]
			if len(rest) > window {
				rest = rest[len(rest)-window:]
			}
			for _, e := range rest {
				lines = append(lines, m.eventLine(e))
			}
			for len(lines) < 3+window {
				lines = append(lines, "")
			}
			last := evs[len(evs)-1]
			lines = append(lines, sChips.Render(commas(last.Chips))+sDim.Render(" chips  × ")+sMult.Render(strconv.FormatFloat(last.Mult, 'f', -1, 64))+sDim.Render(" mult"))
			if m.shown == len(m.events) {
				lines = append(lines, sDim.Render("= ")+sBold.Render(commas(m.gain)))
			}
		}
	case len(sel) > 0:
		ht, scoring := evaluate(sel)
		base := trainedBase(ht, m.levels[ht])
		cardChips := 0
		for _, i := range scoring {
			if !m.cardDebuffed(sel[i]) {
				cardChips += sel[i].Chips()
			}
		}
		lines = append(lines,
			sBold.Render(base.Name),
			fmt.Sprintf("%d base + %d card chips", base.Chips, cardChips),
			fmt.Sprintf("%d × %d = %s before effects", base.Chips+cardChips, base.Mult, commas(multiplyScore(base.Chips+cardChips, base.Mult))),
			"",
			sDim.Render("a to play · d to discard"),
		)
	default:
		lines = append(lines, sDim.Render("select up to 5 cards"))
	}

	for len(lines) < bodyH-1 {
		lines = append(lines, "")
	}
	lines = append(lines[:bodyH-1], m.progressBar())
	return strings.Join(lines, "\n")
}

func (m *model) eventLine(e Event) string {
	if e.Mod >= 0 {
		name := m.mods[e.Mod].Name
		return sWarn.Render("⚡ ") + sAccent.Render(name) + sText.Render(strings.TrimPrefix(e.Label, name))
	}
	label := e.Label
	if strings.Contains(label, "chips") {
		label = sChips.Render(label)
	}
	return cardTag(e.C) + "  " + label
}

func (m *model) progressBar() string {
	t := m.target()
	nums := sBold.Render(commas(m.shownScore)) + sDim.Render(" / "+commas(t))
	barW := rightW - lipgloss.Width(nums) - 1
	ratio := float64(m.shownScore) / float64(t)
	filled := min(int(ratio*float64(barW)), barW)

	color := cBad
	switch {
	case ratio >= 1:
		color = cGood
	case ratio >= 0.5:
		color = cWarn
	}
	bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("━", filled)) +
		sDim.Render(strings.Repeat("─", barW-filled))
	return bar + " " + nums
}

func (m *model) viewMods(w int) string {
	active, hasActive := m.currentEvent()
	var badges []string
	for i := 0; i < m.modSlots(); i++ {
		if i >= len(m.mods) {
			badges = append(badges, sDim.Render("◇"))
			continue
		}
		mod := m.mods[i]
		label := mod.Name
		if m.boss().Rule == "acorn" {
			label = "??"
		}
		if m.boss().Rule == "heart" && i == m.disabledJoker {
			label += " (disabled)"
		}
		if mod.Counter > 0 {
			label += fmt.Sprintf(" +%d", mod.Counter)
		}
		st := lipgloss.NewStyle().Foreground(cAccent)
		if hasActive && active.Mod == i {
			st = st.Reverse(true).Bold(true)
		}
		badges = append(badges, st.Render("◆ "+label))
	}
	money := sGood.Render(fmt.Sprintf("$%d", m.money))
	return spread(" "+strings.Join(badges, "  "), money+" ", w)
}

func (m *model) viewRoundWon() string {
	var b strings.Builder
	b.WriteString(sGood.Render(fmt.Sprintf("round %d cleared", m.round+1)))
	b.WriteString("\n")
	b.WriteString(sDim.Render(fmt.Sprintf("%s / %s", commas(m.score), commas(m.target()))))
	b.WriteString("\n\n")
	total := 0
	for _, r := range m.rewards {
		total += r.amount
		b.WriteString(spread(sText.Render(r.label), sGood.Render(fmt.Sprintf("+$%d", r.amount)), 30))
		b.WriteString("\n")
	}
	b.WriteString(sDim.Render(strings.Repeat("─", 30)) + "\n")
	b.WriteString(spread(sBold.Render("money"), sGood.Render(fmt.Sprintf("$%d", m.money)), 30))
	b.WriteString("\n\n")
	next := "e → shop"
	if m.round == numRounds-1 {
		next = "e → finish"
	}
	b.WriteString(sDim.Render(next))
	return panel.Padding(1, 3).Render(b.String())
}

func (m *model) viewShop() string {
	const w = 64
	var b strings.Builder
	b.WriteString(spread(sAccent.Render("SHOP")+sDim.Render(fmt.Sprintf("  before round %d", m.round+2)), sGood.Render(fmt.Sprintf("$%d", m.money)), w))
	b.WriteString("\n\n")

	row := func(i int, name, price, desc string, priceStyle lipgloss.Style) {
		cur := "  "
		nameStyle := sText
		if i == m.shopCursor {
			cur = sAccent.Render("› ")
			nameStyle = sBold
		}
		b.WriteString(spread(cur+nameStyle.Render(name), priceStyle.Render(price), w))
		b.WriteString("\n")
		b.WriteString("    " + sDim.Render(desc) + "\n")
	}

	b.WriteString(sDim.Render("JOKERS FOR SALE") + "\n")
	if len(m.offers) == 0 {
		b.WriteString(sDim.Render("  sold out, r to reroll") + "\n")
	}
	for i, mod := range m.offers {
		ps := sGood
		if mod.Price > m.money {
			ps = sBad
		}
		row(i, mod.Name, fmt.Sprintf("$%d", mod.Price), mod.Desc, ps)
	}

	b.WriteString("\n" + sDim.Render(fmt.Sprintf("JOKERS %d/%d", len(m.mods), m.modSlots())) + "\n")
	if len(m.mods) == 0 {
		b.WriteString(sDim.Render("  nothing yet") + "\n")
	}
	for i, mod := range m.mods {
		name := mod.Name
		if mod.Counter > 0 {
			name += fmt.Sprintf(" (+%d)", mod.Counter)
		}
		row(len(m.offers)+i, name, fmt.Sprintf("sell $%d", sellPrice(mod)+m.upgrades[resale]), mod.Desc, sDim)
	}

	if m.msg != "" {
		b.WriteString("\n" + sWarn.Render(m.msg) + "\n")
	}
	if m.stockItem != nil {
		fmt.Fprintf(&b, "\nf consumables · %s ($3)", m.stockItem.Name)
	}
	if !m.packSold {
		fmt.Fprintf(&b, "\nb open %s %s pack ($%d)", packSizes[m.packSize], packKinds[m.packKind], m.packPrice())
	}
	if i := m.upgradeOffer; i >= 0 {
		fmt.Fprintf(&b, "\nt redeem %s %d Voucher ($%d) · %s", upgradeLines[i].Name, m.upgrades[i]+1, m.upgradePrice(), upgradeLines[i].Desc)
	}
	b.WriteString("\n" + sDim.Render(fmt.Sprintf("w/s move · e buy/sell · 1/2 reorder owned · r reroll ($%d) · c next Blind · v deck · z collection · f consumables · g hands · , settings · ctrl+n new game · q quit", m.rerollCharge())))
	return panel.Padding(1, 2).Render(b.String())
}

func (m *model) viewEnd() string {
	var b strings.Builder
	if m.phase == phaseWon {
		b.WriteString(sGood.Render(fmt.Sprintf("all %d rounds cleared", numRounds)) + "\n\n")
		b.WriteString(sText.Render(fmt.Sprintf("finished with $%d and %d Jokers", m.money, len(m.mods))) + "\n")
	} else {
		b.WriteString(sBad.Render("out of hands") + "\n\n")
		b.WriteString(sText.Render(fmt.Sprintf("round %d: %s / %s", m.round+1, commas(m.score), commas(m.target()))) + "\n")
	}
	fmt.Fprintf(&b, "\nBest play %s · Blinds skipped %d\n", commas(m.bestGain), m.skipped)
	if m.phase == phaseWon {
		b.WriteString("e continue endless\n")
	}
	b.WriteString("\n" + sDim.Render("n replay seed · v deck · z collection · f consumables · g hands · , settings · ctrl+n new game · q quit"))
	return panel.Padding(1, 3).Render(b.String())
}

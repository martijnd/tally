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
	switch m.phase {
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
		sAccent.Render(" tally")+sDim.Render(fmt.Sprintf("  round %d/%d", m.round+1, numRounds)),
		sDim.Render("target ")+sBold.Render(commas(m.target()))+" ",
		w,
	)

	msg := ""
	if m.msg != "" {
		msg = " " + sWarn.Render(m.msg)
	}
	help := sDim.Render(" ↑↓ move · space select · p play · d discard · s sort · c clear · q quit")
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
		b.WriteString(spread(row, tail, leftW))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(sDim.Render("plays ") + pips(m.plays, playsPerRound, cAccent) +
		sDim.Render("   discards ") + pips(m.discards, discardsPerRound, cWarn))
	return b.String()
}

func (m *model) viewScore() string {
	var lines []string
	lines = append(lines, sDim.Render("SCORE"), "")

	sel := m.selectedCards()
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
			lines = append(lines, sChips.Render(commas(last.Chips))+sDim.Render(" chips  × ")+sMult.Render(commas(last.Mult))+sDim.Render(" mult"))
			if m.shown == len(m.events) {
				lines = append(lines, sDim.Render("= ")+sBold.Render(commas(m.gain)))
			}
		}
	case len(sel) > 0:
		ht, _ := evaluate(sel)
		base := handBases[ht]
		lines = append(lines,
			sBold.Render(base.Name),
			sChips.Render(strconv.Itoa(base.Chips))+sDim.Render(" chips  × ")+sMult.Render(strconv.Itoa(base.Mult))+sDim.Render(" mult"),
			"",
			sDim.Render("p to play · d to discard"),
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
	for i := 0; i < maxMods; i++ {
		if i >= len(m.mods) {
			badges = append(badges, sDim.Render("◇"))
			continue
		}
		mod := m.mods[i]
		label := mod.Name
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
	next := "enter → shop"
	if m.round == numRounds-1 {
		next = "enter → finish"
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

	b.WriteString(sDim.Render("FOR SALE") + "\n")
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

	b.WriteString("\n" + sDim.Render(fmt.Sprintf("OWNED %d/%d", len(m.mods), maxMods)) + "\n")
	if len(m.mods) == 0 {
		b.WriteString(sDim.Render("  nothing yet") + "\n")
	}
	for i, mod := range m.mods {
		name := mod.Name
		if mod.Counter > 0 {
			name += fmt.Sprintf(" (+%d)", mod.Counter)
		}
		row(len(m.offers)+i, name, fmt.Sprintf("sell $%d", sellPrice(mod)), mod.Desc, sDim)
	}

	if m.msg != "" {
		b.WriteString("\n" + sWarn.Render(m.msg) + "\n")
	}
	b.WriteString("\n" + sDim.Render(fmt.Sprintf("↑↓ move · enter buy/sell · r reroll ($%d) · c next round · q quit", rerollCost)))
	return panel.Padding(1, 2).Render(b.String())
}

func (m *model) viewEnd() string {
	var b strings.Builder
	if m.phase == phaseWon {
		b.WriteString(sGood.Render(fmt.Sprintf("all %d rounds cleared", numRounds)) + "\n\n")
		b.WriteString(sText.Render(fmt.Sprintf("finished with $%d and %d modifiers", m.money, len(m.mods))) + "\n")
	} else {
		b.WriteString(sBad.Render("out of plays") + "\n\n")
		b.WriteString(sText.Render(fmt.Sprintf("round %d: %s / %s", m.round+1, commas(m.score), commas(m.target()))) + "\n")
	}
	b.WriteString("\n" + sDim.Render("n new run · q quit"))
	return panel.Padding(1, 3).Render(b.String())
}

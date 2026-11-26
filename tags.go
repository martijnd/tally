package main

import (
	"fmt"
	"strings"
)

type tag struct {
	Name, Desc, When, Effect string
	Value                    int
}

var tags = []tag{
	{"Payroll", "Collect $5 immediately", "skip", "cash", 5}, {"Deferred wage", "Collect $9 in the next shop", "shop", "cash", 9}, {"Settlement", "Collect $12 after clearing a round", "win", "cash", 12}, {"Toll slip", "Next shop reroll costs $0", "reroll", "free", 0}, {"First refusal", "Next shop modifier is free", "offer", "free", 0}, {"Parcel", "Next booster pack is free", "pack", "free", 0}, {"Carbon sheet", "Apply the next earned tag twice", "skip", "double", 0},
	{"Etcher", "Etch the next shop modifier", "offer", "edition", 1}, {"Kiln", "Flare the next shop modifier", "offer", "edition", 2}, {"Polisher", "Mirror the next shop modifier", "offer", "edition", 3}, {"Annex", "Next shop modifier is spacious", "offer", "edition", 4},
	{"Overtime", "One extra play next round", "round", "play", 1}, {"Sorting tray", "Two extra discards next round", "round", "discard", 2}, {"Shift pay", "Collect $6 at next round start", "round", "cash", 6},
	{"Pocket manual", "Receive a hand lesson when space permits", "shop", "lesson", 1}, {"Dividend slip", "Collect $1 per owned modifier next shop", "shop", "owned", 1}, {"Headcount", "Collect $2 per unused play next win", "win", "plays", 2},
	{"Pair seminar", "Train pair immediately", "skip", "level", int(Pair)}, {"Survey", "Train high card immediately", "skip", "level", int(HighCard)}, {"Sequence chart", "Train straight immediately", "skip", "level", int(Straight)}, {"Colour wheel", "Train flush immediately", "skip", "level", int(Flush)}, {"Triplicate", "Train three of a kind immediately", "skip", "level", int(ThreeKind)},
	{"Night class", "Train full house on next win", "win", "level", int(FullHouse)}, {"Tuition refund", "Gain $4 on next reroll", "reroll", "cash", 4},
}

func (m *model) skipTag() int {
	return runStream(m.seed, fmt.Sprintf("skip-%d", m.round)).IntN(len(tags))
}
func (m *model) earnTag() {
	id := m.skipTag()
	n := 1
	if tags[id].Effect != "double" {
		n += m.doubleTags
		m.doubleTags = 0
	}
	for range n {
		m.tagQueue = append(m.tagQueue, id)
	}
	m.applyTags("skip")
}
func (m *model) applyTags(when string) {
	offerIndex := 0
	keep := m.tagQueue[:0]
	for _, id := range m.tagQueue {
		t := tags[id]
		if t.When != when {
			keep = append(keep, id)
			continue
		}
		switch t.Effect {
		case "double":
			m.doubleTags++
		case "cash":
			m.money += t.Value
		case "owned":
			m.money += len(m.mods) * t.Value
		case "plays":
			m.money += m.plays * t.Value
		case "level":
			m.levels[t.Value]++
		case "play":
			m.plays += t.Value
		case "discard":
			m.discards += t.Value
		case "lesson":
			if len(m.bag) >= m.bagSlots() {
				keep = append(keep, id)
				continue
			}
			m.bag = append(m.bag, m.lesson())
		case "free":
			switch when {
			case "reroll":
				m.freeRerolls++
			case "pack":
				m.freePacks++
			case "offer":
				if offerIndex >= len(m.offers) {
					keep = append(keep, id)
					continue
				}
				m.offers[offerIndex].Price = 0
				offerIndex++
			}
		case "edition":
			if offerIndex >= len(m.offers) {
				keep = append(keep, id)
				continue
			}
			mod := m.offers[offerIndex]
			offerIndex++
			mod.Edition = t.Value
			mod.Desc = editions[t.Value] + " edition · " + mod.Desc
		}
	}
	m.tagQueue = keep
}
func (m *model) tagSummary() string {
	var names []string
	for _, i := range m.tagQueue {
		names = append(names, tags[i].Name)
	}
	if m.doubleTags > 0 {
		names = append(names, fmt.Sprintf("Carbon sheet ×%d", m.doubleTags))
	}
	if m.freePacks > 0 {
		names = append(names, fmt.Sprintf("Free packs ×%d", m.freePacks))
	}
	if m.freeRerolls > 0 {
		names = append(names, fmt.Sprintf("Free rerolls ×%d", m.freeRerolls))
	}
	if len(names) == 0 {
		return ""
	}
	return "\nTags · " + strings.Join(names, ", ")
}
func (m *model) rerollCharge() int {
	if m.freeRerolls > 0 {
		return 0
	}
	for _, id := range m.tagQueue {
		if tags[id].When == "reroll" && tags[id].Effect == "free" {
			return 0
		}
	}
	return m.rerollPrice
}

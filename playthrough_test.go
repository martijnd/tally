package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestKeyboardPlaythrough(t *testing.T) {
	m := newModel("table")
	key := func(k string) { m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}); _ = m.View() }
	key("e")
	key("e")
	key("f")
	key("x")
	if !m.selected[m.cursor] {
		t.Fatal("x did not select the consumable target")
	}
	key("f")
	for actions := 0; actions < 300; actions++ {
		switch m.phase {
		case phasePlay:
			key("c")
			m.simulateSelection()
			wanted := append([]bool(nil), m.selected...)
			clear(m.selected)
			for m.cursor > 0 {
				key("w")
			}
			for i, on := range wanted {
				if on {
					key("x")
				}
				if i+1 < len(wanted) {
					key("s")
				}
			}
			key("a")
		case phaseScoring, phasePreview, phaseRoundWon:
			key("e")
		case phaseShop:
			key("f")
			key("b")
			key("b")
			if m.msg == "" {
				t.Fatal("missing purchase feedback")
			}
			key("f")
			key("e")
			key("c")
		case phaseLost, phaseWon:
			t.Log("finished round", m.round+1, "score", m.score)
			return
		}
	}
	t.Fatal("run stalled")
}

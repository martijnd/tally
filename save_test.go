package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"path/filepath"
	"reflect"
	"testing"
)

func TestResumeScoringAndProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tally.json")
	s, err := loadSession(path, "resume", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []tea.KeyType{tea.KeyDown, tea.KeyEnter, tea.KeyEnter, tea.KeySpace, tea.KeyEnter} {
		s.Update(tea.KeyMsg{Type: k})
	}
	if s.game.phase != phaseScoring || s.game.discards != 4 {
		t.Fatal("round did not start with selected deck")
	}
	s.Update(tickMsg{})
	s.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	other, err := loadSession(path, "different", false)
	if err != nil {
		t.Fatal(err)
	}
	if other.game.phase != phaseScoring || other.game.shown != 1 || !reflect.DeepEqual(s.game.hand, other.game.hand) || other.game.gain != s.game.gain || other.game.seed != "resume" {
		t.Fatal("resume lost game state")
	}
	for other.game.phase == phaseScoring {
		other.Update(tickMsg{})
	}
	fresh, err := loadSession(path, "fresh", true)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.game.phase != phaseStart || fresh.data.Profile.Hands[HighCard] != 1 {
		t.Fatal("career totals lost or duplicated")
	}
}
func TestUnlocksAndStickers(t *testing.T) {
	m := newModel("stakes")
	m.deckID = 14
	m.startKey("enter")
	if m.phase != phaseStart {
		t.Fatal("locked deck selectable")
	}
	m.unlockRound = 9
	m.bestDifficulty[14] = 7
	m.difficulty = 7
	m.startKey("enter")
	m.startRound()
	if m.discards != 2 || m.roundReward() != 0 || m.target() <= 300 {
		t.Fatal("difficulty rules missing")
	}
	m.mods = []*Modifier{{Name: "Lease", Sticker: "Rent"}, {Name: "Ink", Sticker: "Fading", Life: 1}, {Name: "Lock", Sticker: "Bound"}}
	m.money = 4
	m.settleStickers()
	if len(m.mods) != 2 || m.money != 3 {
		t.Fatal("sticker settlement")
	}
	m.shopCursor = 1
	m.shopSelect()
	if len(m.mods) != 2 {
		t.Fatal("bound modifier sold")
	}
}

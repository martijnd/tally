package main

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"path/filepath"
)

type profile struct {
	Prefs       preferences
	Highest     [15]int
	Wins        [15][8]int
	Best, Round int
	Hands       [len(handBases)]int
	Seen        map[string]bool
}
type input struct {
	Mouse *tea.MouseMsg `json:",omitempty"`
	Key   *tea.KeyMsg   `json:",omitempty"`
	Ticks int           `json:",omitempty"`
}
type recording struct {
	Discoveries   int
	LegacyEffects bool
	LegacyPacks   bool
	Prefs         preferences
	UnlockRound   int
	Seed          string
	Opening       [15]int
	Inputs        []input
}
type saveFile struct {
	Version int
	Profile profile
	Run     *recording
}
type session struct {
	game    *model
	path    string
	data    saveFile
	saveErr error
}

func loadSession(path, seed string, fresh bool) (*session, error) {
	s := &session{path: path, data: saveFile{Version: 4}}
	raw, err := os.ReadFile(path)
	if err == nil {
		if err = json.Unmarshal(raw, &s.data); err != nil {
			return nil, err
		}
		if s.data.Version != 1 && s.data.Version != 2 && s.data.Version != 3 && s.data.Version != 4 {
			return nil, fmt.Errorf("unsupported save version %d", s.data.Version)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if s.data.Profile.Seen == nil {
		s.data.Profile.Seen = map[string]bool{}
	}
	r := s.data.Run
	if r != nil && s.data.Version < 4 {
		r.LegacyEffects = true
	}
	if s.data.Version < 3 {
		if r != nil {
			r.LegacyPacks = true
		}
		for old, name := range consumableNames {
			if s.data.Profile.Seen[old] {
				s.data.Profile.Seen[name] = true
				delete(s.data.Profile.Seen, old)
			}
		}
	}
	if fresh || r == nil {
		r = &recording{Prefs: s.data.Profile.Prefs, Discoveries: len(s.data.Profile.Seen), Seed: seed, UnlockRound: s.data.Profile.Round, Opening: s.data.Profile.Highest}
		s.data.Run = r
	}
	s.game = newModel(r.Seed)
	s.game.legacyEffects = r.LegacyEffects
	if r.LegacyEffects {
		s.game.restoreLegacyBosses()
	}
	s.game.legacyPacks = r.LegacyPacks
	s.game.prefs = r.Prefs
	s.game.bestDifficulty = r.Opening
	s.game.discoveries = r.Discoveries
	s.game.unlockRound = r.UnlockRound
	for _, in := range r.Inputs {
		if in.Mouse != nil {
			s.game.replayMouse(*in.Mouse)
		} else if in.Key != nil {
			if s.data.Version == 1 {
				switch in.Key.String() {
				case "a":
					in.Key.Runes = []rune{'l'}
				case "A":
					in.Key.Runes = []rune{'a'}
				case "l":
					in.Key.Runes = nil // Previously unbound.
				}
			}
			if in.Key.String() == "p" {
				in.Key.Runes = []rune{'a'}
			}
			s.game.updateResolved(*in.Key)
		} else {
			for range in.Ticks {
				s.game.Update(tickMsg{})
			}
		}
	}
	s.data.Version = 4
	s.game.profile = &s.data.Profile
	return s, nil
}
func (s *session) persist() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".tally-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	err = json.NewEncoder(f).Encode(s.data)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), s.path)
}
func (s *session) Init() tea.Cmd {
	if s.game.phase == phaseScoring {
		return s.game.tick()
	}
	return nil
}
func (s *session) View() string {
	v := s.game.View()
	if s.saveErr != nil {
		return v + "\nSave failed · " + s.saveErr.Error()
	}
	return v
}
func (s *session) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m := s.game
	oldPrefs := m.prefs
	before, counts := m.phase, m.timesPlayed
	msg = m.resolveKey(msg)
	_, cmd := m.updateResolved(msg)
	r := s.data.Run
	if (r == nil || before == phaseWon || before == phaseLost) && m.phase == phaseStart {
		m.prefs = oldPrefs
		m.discoveries = len(s.data.Profile.Seen)
		m.unlockRound = s.data.Profile.Round
		r = &recording{Prefs: m.prefs, Discoveries: m.discoveries, Seed: m.seed, UnlockRound: m.unlockRound, Opening: m.bestDifficulty}
		s.data.Run = r
	}
	m.profile = &s.data.Profile
	write := false
	switch v := msg.(type) {
	case tea.KeyMsg:
		write = true
		if r != nil && v.String() != "q" && v.String() != "ctrl+c" && !(m.phase == phaseStart && (before == phaseWon || before == phaseLost)) {
			r.Inputs = append(r.Inputs, input{Key: &v})
		}
	case tickMsg:
		if before != phaseScoring {
			return s, cmd
		}
		if r != nil {
			n := len(r.Inputs)
			if n == 0 || r.Inputs[n-1].Key != nil || r.Inputs[n-1].Mouse != nil {
				r.Inputs = append(r.Inputs, input{})
			}
			r.Inputs[len(r.Inputs)-1].Ticks++
		}
		write = m.phase != before
	default:
		return s, cmd
	}
	p := &s.data.Profile
	for i, n := range m.timesPlayed {
		if n > counts[i] {
			p.Hands[i] += n - counts[i]
			p.Seen[handBases[i].Name] = true
		}
	}
	p.Prefs = m.prefs
	p.Best = max(p.Best, m.bestGain)
	if m.phase != phaseStart {
		p.Round = max(p.Round, m.round+1)
		p.Seen[m.decks()[m.deckID].Name] = true
		if m.challengeID > 0 {
			p.Seen[challenges[m.challengeID-1].Name] = true
		}
		if m.phase == phasePreview {
			if m.round%3 == 2 {
				p.Seen[m.boss().Name] = true
			} else if !m.restriction("no skips") {
				p.Seen[tags[m.skipTag()].Name] = true
			}
		}
		for _, c := range m.stock {
			p.Seen[enhancements[c.Enhancement]] = true
			p.Seen[editions[c.Edition]] = true
			p.Seen[seals[c.Seal]] = true
		}
	}
	for _, mod := range m.mods {
		p.Seen[mod.Name] = true
		p.Seen[editions[mod.Edition]] = true
	}
	for _, it := range m.bag {
		p.Seen[it.Name] = true
	}
	for i, n := range m.upgrades {
		if n > 0 {
			p.Seen[upgradeLines[i].Name] = true
		}
	}
	if before != phaseWon && m.phase == phaseWon {
		p.Highest = m.bestDifficulty
		p.Wins[m.deckID][m.difficulty]++
	}
	if m.phase == phaseLost {
		s.data.Run = nil
	}
	if write {
		s.saveErr = s.persist()
	}
	return s, cmd
}

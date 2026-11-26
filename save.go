package main

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"path/filepath"
)

type profile struct {
	Highest     [15]int
	Wins        [15][8]int
	Best, Round int
	Hands       [len(handBases)]int
	Seen        map[string]bool
}
type input struct {
	Key   *tea.KeyMsg `json:",omitempty"`
	Ticks int         `json:",omitempty"`
}
type recording struct {
	UnlockRound int
	Seed        string
	Opening     [15]int
	Inputs      []input
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
	s := &session{path: path, data: saveFile{Version: 1}}
	raw, err := os.ReadFile(path)
	if err == nil {
		if err = json.Unmarshal(raw, &s.data); err != nil {
			return nil, err
		}
		if s.data.Version != 1 {
			return nil, fmt.Errorf("unsupported save version %d", s.data.Version)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if s.data.Profile.Seen == nil {
		s.data.Profile.Seen = map[string]bool{}
	}
	r := s.data.Run
	if fresh || r == nil {
		r = &recording{Seed: seed, UnlockRound: s.data.Profile.Round, Opening: s.data.Profile.Highest}
		s.data.Run = r
	}
	s.game = newModel(r.Seed)
	s.game.bestDifficulty = r.Opening
	s.game.unlockRound = r.UnlockRound
	for _, in := range r.Inputs {
		if in.Key != nil {
			s.game.Update(*in.Key)
		} else {
			for range in.Ticks {
				s.game.Update(tickMsg{})
			}
		}
	}
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
		return tick()
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
	before, counts := m.phase, m.timesPlayed
	_, cmd := m.Update(msg)
	r := s.data.Run
	if (r == nil || before == phaseWon || before == phaseLost) && m.phase == phaseStart {
		m.unlockRound = s.data.Profile.Round
		r = &recording{Seed: m.seed, UnlockRound: m.unlockRound, Opening: m.bestDifficulty}
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
			if n == 0 || r.Inputs[n-1].Key != nil {
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
		}
	}
	p.Best = max(p.Best, m.bestGain)
	if m.phase != phaseStart {
		p.Round = max(p.Round, m.round+1)
		p.Seen[startingDecks[m.deckID].Name] = true
	}
	for _, mod := range m.mods {
		p.Seen[mod.Name] = true
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

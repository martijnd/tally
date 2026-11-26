package main

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"math/rand/v2"
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
	Key   *tea.KeyMsg `json:",omitempty"`
	Ticks int         `json:",omitempty"`
}
type recording struct {
	Discoveries int
	Prefs       preferences
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
	game       *model
	path       string
	data       saveFile
	saveErr    error
	confirmNew bool
}

func loadSession(path, seed string, fresh bool) (*session, error) {
	const saveVersion = 5
	s := &session{path: path}
	raw, err := os.ReadFile(path)
	if err == nil {
		if err = json.Unmarshal(raw, &s.data); err != nil {
			return nil, err
		}
		if s.data.Version != saveVersion {
			return nil, fmt.Errorf("unsupported save version %d in %s; move the file aside to start a new profile", s.data.Version, path)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	s.data.Version = saveVersion
	if s.data.Profile.Seen == nil {
		s.data.Profile.Seen = map[string]bool{}
	}
	r := s.data.Run
	if fresh || r == nil {
		r = s.newRecording(seed)
	}
	s.game = newModel(r.Seed)
	s.game.prefs = r.Prefs
	s.game.bestDifficulty = r.Opening
	s.game.discoveries = r.Discoveries
	s.game.unlockRound = r.UnlockRound
	for _, in := range r.Inputs {
		if in.Key != nil {
			s.game.updateResolved(*in.Key)
		} else {
			for range in.Ticks {
				s.game.Update(tickMsg{})
			}
		}
	}
	s.game.profile = &s.data.Profile
	return s, nil
}
func (s *session) newRecording(seed string) *recording {
	p := &s.data.Profile
	r := &recording{Prefs: p.Prefs, Discoveries: len(p.Seen), Seed: seed, UnlockRound: p.Round, Opening: p.Highest}
	s.data.Run = r
	return r
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
	if s.confirmNew {
		return panel.Padding(1, 2).Render("Replace this run with a new game?\nYour profile and unlocks are kept.\n\nCtrl+N start new game · Esc cancel")
	}
	v := s.game.View()
	if s.saveErr != nil {
		return v + "\nSave failed · " + s.saveErr.Error()
	}
	return v
}
func (s *session) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		if k.Type == tea.KeyCtrlN {
			if !s.confirmNew {
				s.confirmNew = true
				return s, nil
			}
			s.confirmNew = false
			r := s.newRecording(fmt.Sprintf("%016x", rand.Uint64()))
			m := newModel(r.Seed)
			m.prefs, m.bestDifficulty, m.discoveries, m.unlockRound = r.Prefs, r.Opening, r.Discoveries, r.UnlockRound
			m.profile = &s.data.Profile
			s.game = m
			s.saveErr = s.persist()
			return s, nil
		}
		if s.confirmNew && k.Type != tea.KeyCtrlC {
			if k.Type == tea.KeyEsc {
				s.confirmNew = false
			}
			return s, nil
		}
	}
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
			p.Seen[handBases[i].Name] = true
		}
	}
	p.Prefs = m.prefs
	p.Best = max(p.Best, m.bestGain)
	if m.phase != phaseStart {
		p.Round = max(p.Round, m.round+1)
		p.Seen[startingDecks[m.deckID].Name] = true
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

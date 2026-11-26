package main

import (
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func runStream(seed, system string) *rand.Rand {
	digest := sha256.Sum256([]byte(system + "\x00" + seed))
	return rand.New(rand.NewPCG(binary.BigEndian.Uint64(digest[:8]), binary.BigEndian.Uint64(digest[8:16])))
}
func main() {
	seed := flag.String("seed", "", "replay a named seed")
	daily := flag.Bool("daily", false, "play today's shared UTC seed")
	fresh := flag.Bool("new", false, "start a fresh run and keep your profile")
	runs := flag.Int("simulate", 0, "simulate seeded runs and report progression by Joker")
	flag.Parse()
	if *runs > 0 {
		simulate(os.Stdout, *runs, *seed)
		return
	}
	explicit := *seed != "" || *daily || *fresh
	if *daily && *seed != "" {
		fmt.Fprintln(os.Stderr, "choose either -seed or -daily")
		os.Exit(2)
	}
	if *daily {
		*seed = time.Now().UTC().Format("2006-01-02")
	}
	if *seed == "" {
		*seed = fmt.Sprintf("%016x", rand.Uint64())
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	s, err := loadSession(filepath.Join(dir, "tally", "save.json"), *seed, explicit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	p := tea.NewProgram(s, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

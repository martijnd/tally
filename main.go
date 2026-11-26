package main

import (
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
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
	flag.Parse()
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
	p := tea.NewProgram(newModel(*seed), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

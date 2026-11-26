# tally

A roguelike poker-scoring game for the terminal, built with Bubble Tea and Lip Gloss.

```sh
go run .
# or
go build -o tally . && ./tally
```

Use `go run . -seed NAME` to replay a seed or `go run . -daily` for the shared UTC daily run. Press `h` to view the seed and hand levels.

Starting the game resumes your saved run. Use `-new` for a fresh run. At deck selection, use the arrows to choose a deck and difficulty, `c` to cycle challenges, and `a` to inspect your collection and career.

## How to play

Each round you have a target score, 4 plays and 3 discards. Select up to 5 cards and play them as a poker hand.

Score = chips × mult. Each hand type has base chips and mult, and every scoring card adds chips: its number, 10 for J/Q/K, 11 for A. Modifiers you buy in the shop add chips, add mult or multiply mult.

Clear all 24 rounds across eight stages to win. You earn money for each cleared round, for unused plays and as interest on savings.

| Key | Action |
|---|---|
| ↑↓ / j k | move |
| space / x | select card |
| p / enter | play selected |
| d | discard selected |
| s | sort by rank or suit |
| c | clear selection |
| h / v / i | inspect run, deck, or supplies |
| [ / ] | move the current card or owned shop modifier |
| b / u in shop | open a pack or buy an upgrade |
| , | configure animation speed, theme, and bell |
| enter while scoring | finish the animation |
| q | save and quit |

Click a card to select it, or click the score panel to play. The scroll wheel moves the cursor.

## Development

Run `go test ./...` to check the rules and terminal controls. Use `go run . -simulate 3000 -seed audit` for a repeatable balance report with round reach and results by starting modifier. The simulated player buys affordable modifiers and plays for base hand value.

Run `make release` to build macOS and Linux binaries for Intel and ARM64, plus Windows for x86-64, in `dist/`.

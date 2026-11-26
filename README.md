# tally

A roguelike poker-scoring game for the terminal, built with Bubble Tea and Lip Gloss.

```sh
go run .
# or
go build -o tally . && ./tally
```

Use `go run . -seed NAME` to replay a seed or `go run . -daily` for the shared UTC daily run. Press `h` to view the seed and hand levels.

Starting the game resumes your saved run. Use `-new` for a fresh run. At deck selection, use the arrows to choose a Deck and Stake, `c` to cycle challenges, and `a` to inspect your collection and career.

Celestial Packs contain Planet cards that level up poker hands. Arcana Packs contain Tarot cards for deck edits and utility effects. Spectral Packs contain powerful cards with a cost. Buffoon Packs offer Jokers, and Standard Packs offer playing cards.

## How to play

Each round you have a target score, 4 hands and 3 discards. Select up to 5 cards and play them as a poker hand.

Score = chips × mult. Each hand type has base chips and mult, and every scoring card adds chips: its number, 10 for J/Q/K, 11 for A. Jokers you buy in the shop add chips, add mult or multiply mult.

Clear all eight Antes, each with a Small Blind, Big Blind, and Boss Blind, to win. You earn money for each cleared round, for unused hands and as interest on savings.

| Key | Action |
|---|---|
| ↑↓ / j k | move |
| space / x | select card |
| p / enter | play selected |
| d | discard selected |
| s | sort by rank or suit |
| c | clear selection |
| h / v / i | inspect run, deck, or consumables |
| [ / ] | move the current card or owned shop Joker |
| b / u in shop | open a pack or redeem a Voucher |
| , | configure animation speed, theme, and bell |
| enter while scoring | finish the animation |
| q | save and quit |

Click a card to select it, or click the score panel to play. The scroll wheel moves the cursor.

## Development

Run `go test ./...` to check the rules and terminal controls. Use `go run . -simulate 3000 -seed audit` for a repeatable balance report with round reach and results by starting Joker. The simulated player buys affordable Jokers and plays for base hand value.

Run `make release` to build macOS and Linux binaries for Intel and ARM64, plus Windows for x86-64, in `dist/`.

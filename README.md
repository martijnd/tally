# tally

A roguelike poker-scoring game for the terminal, built with Bubble Tea and Lip Gloss.

```sh
go run .
# or
go build -o tally . && ./tally
```

Use `go run . -seed NAME` to replay a seed or `go run . -daily` for the shared UTC daily run. Press `h` to view the seed and hand levels.

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
| q | quit |

# Roadmap

The goal is the same depth of systems as the big games in the poker-roguelike genre. Mechanics may match the genre, but names, content and look stay our own: no copied card names, effects lists, art or text.

Status: `[x]` done, `[ ]` to do.

## 0. What exists today

- [x] 52-card deck, hand of 8, select up to 5
- [x] 4 plays and 3 discards per round
- [x] Nine standard poker hands with base chips × mult
- [x] Card chips: number cards at face value, J/Q/K 10, A 11
- [x] 8 rounds with fixed, rising targets
- [x] Money: round reward, unused plays, interest
- [x] Shop with 3 offers, buy, sell, reroll
- [x] 13 modifiers in 5 slots (additive chips, additive mult, ×mult, stateful)
- [x] Animated scoring breakdown, sort by rank or suit

## 1. Run structure

- [x] Stages: each stage has three rounds (opening, main, boss) with score multipliers of 1×, 1.5× and 2× the stage base
- [x] Stage base targets that scale steeply, win after stage 8
- [x] Endless mode after a win, with targets scaling much faster
- [x] Skip a non-boss round in exchange for a **tag** (a one-off reward)
- [x] Round preview screen: target, reward, boss effect, skip tag
- [x] Rewards differ per round type (opening $3, main $4, boss $5)
- [ ] Interest cap that can be raised by upgrades
- [x] Game over when out of plays below target; keep run stats

## 2. Boss rounds

- [x] Boss pool of ~25 original effects, one drawn per stage, no repeats in a run
- [x] Effect types: debuff a suit, debuff face cards, hand size -1, only one hand type allowed, cards drawn face down, discard random cards after each play, first hand scores nothing, money drains per play, modifiers disabled
- [x] Final-stage bosses: harsher effects, used only on stage 8
- [x] Boss effect shown in the header and on affected cards
- [ ] Some modifiers or consumables can disable the boss effect

## 3. Hand levels

- [x] Each hand type has a level; levelling adds base chips and mult per hand type
- [x] Track times played per hand type (used by some modifiers and the stats screen)
- [x] Hand overview screen: level, chips, mult, times played
- [x] Secret hands, only possible with modified decks: five of a kind, flush house, flush five

## 4. Modifiers (the main build engine)

- [ ] Grow the pool to ~150 original modifiers
- [ ] Rarity tiers: common, uncommon, rare, legendary, with shop weights
- [ ] Slot order matters and can be rearranged
- [ ] Trigger points: on card scored, on hand scored, on card held in hand, on discard, on round start/end, on shop enter, on reroll, on buy/sell, on card added/destroyed
- [ ] Effect families: +chips, +mult, ×mult, money, retrigger cards, copy another modifier, create consumables, add cards to the deck, destroy cards, scale over time, probability-based effects
- [ ] Probability system with a global "odds" multiplier that some items double
- [ ] Sell value that grows on some modifiers
- [ ] Stickers on modifiers (higher difficulties): can't be sold, expires after N rounds, costs $1 per round
- [ ] Slot count can be raised by upgrades and editions

## 5. Playing card upgrades

- [ ] **Enhancements** (one per card), e.g. extra chips, extra mult, ×mult when held, wild suit, no rank or suit but big chips, chance to break, money when scored
- [ ] **Editions** for cards and modifiers: +chips, +mult, ×mult, and one that adds a slot
- [ ] **Seals**: retrigger, money when played, create a consumable when held or discarded
- [ ] Card detail view showing enhancement, edition and seal
- [ ] Deck can grow and shrink; cards can be copied, destroyed, or changed in rank and suit

## 6. Consumables

- [ ] Consumable slots (2 by default), usable during rounds and in the shop
- [ ] **Level-up items**: one per hand type, level that hand
- [ ] **Card-edit items** (~22): enhance, change suit, change rank, copy, destroy, make money, create other consumables
- [ ] **Rare high-risk items** (~18): powerful effects with a cost, e.g. add an edition but destroy something, raise hand size but lose money
- [ ] Targeting flow: select cards in hand, then use the consumable

## 7. Shop

- [ ] Shop sections: modifiers and consumables, booster packs, one upgrade
- [ ] **Booster packs**: open and pick 1 or 2 of 3 to 5 options; pack types for modifiers, playing cards, level-ups, card edits and rare items; normal, jumbo and mega sizes
- [ ] **Run upgrades** (~32, in pairs where the second needs the first): extra shop slot, cheaper rerolls, higher interest cap, +1 play, +1 discard, +1 hand size, better odds of editions, more consumable slots
- [x] Reroll cost that rises within a shop visit and resets each visit
- [ ] Price scaling with editions, rarity and run modifiers

## 8. Tags

- [ ] ~24 tags earned by skipping rounds, e.g. free shop item, free reroll, a booster pack, money, an edition on the next shop modifier, double the next tag
- [ ] Tag queue shown in the UI, applied at the right moment

## 9. Starting decks

- [ ] ~15 starting decks with one rule each, e.g. +1 discard, +$10, extra consumable slot, no face cards, suits merged into two colours, abstract deck with all cards enhanced
- [ ] Deck select screen at the start of a run

## 10. Difficulty levels

- [ ] ~8 stacking difficulty levels, unlocked by winning on the previous one
- [ ] Effects per level: no reward on the opening round, faster target scaling, stickers on shop modifiers, -1 discard, lower interest
- [ ] Track the highest level won per deck

## 11. Meta progression

- [ ] Persistent save file (run in progress plus profile)
- [ ] Resume an unfinished run
- [ ] Collection screen: everything discovered, undiscovered shown as `?`
- [ ] Unlock conditions for modifiers, decks and upgrades
- [ ] Profile stats: best hand, most played hand, highest round, wins per deck and difficulty
- [ ] Achievements

## 12. Run options

- [ ] Seeded runs: show the seed, start a run from a seed, deterministic RNG per system (shop, deck, packs, bosses)
- [ ] Challenge runs: ~20 hand-designed rule sets with fixed starting modifiers and restrictions
- [x] Daily run with a shared seed

## 13. UI and feel

- [ ] Deck viewer: remaining and full deck, grouped by suit, with modifications
- [ ] Run info screen: hand levels, boss list, upgrades owned
- [ ] Modifier and consumable detail on hover or cursor
- [ ] Mouse support
- [ ] Faster or skippable scoring animation, configurable speed
- [x] Rearrange cards in hand by hand, not only sort
- [ ] Big-number formatting for endless mode (1.2e15 and up)
- [ ] Sound bell on big scores (optional), colour themes, light terminal support
- [ ] Settings screen

## 14. Engineering

- [ ] Separate game rules from the TUI so rules can be tested without rendering
- [ ] Event-driven scoring pipeline that handles retriggers, copies and trigger order
- [ ] Data-driven content definitions for modifiers, bosses, consumables and decks
- [ ] Simulation harness to balance targets and item strength over thousands of runs
- [ ] Release builds for macOS, Linux and Windows

package main

var stakeNames = [...]string{"White", "Red", "Green", "Black", "Blue", "Purple", "Orange", "Gold"}

// The legacy keys keep collection discoveries and seeded item ordering stable.
var consumableNames = map[string]string{
	"Tracing Paper": "Cryptid", "Margin Knife": "The Hanged Man", "Ascending Ink": "Strength", "Descending Ink": "The Fool", "Bridge Stamp": "Strength (+3 ranks)", "Overprint": "The Hermit", "Proof Slip": "The High Priestess", "Quiet Writ": "Judgement",
	"Spade Plate": "The World", "Heart Plate": "The Sun", "Diamond Plate": "The Star", "Club Plate": "The Moon",
	"Weighted Ink": "The Hierophant", "Loud Ink": "The Empress", "Patient Ink": "The Chariot", "Prismatic Ink": "The Lovers", "Brittle Ink": "Justice", "Dividend Ink": "The Devil", "Blank Ink": "The Tower",
	"Echo Press": "Deja Vu", "Mint Press": "Talisman", "Archive Press": "Trance", "Scrap Press": "Medium",
	"Deep Pocket": "Ouija", "Broad Shelf": "Wraith", "Lost Receipt": "Immolate", "Missing Page": "Black Hole", "Empty Purse": "The Soul", "Bent Frame": "Familiar", "Loose Leaf": "Incantation", "Silent Auction": "Temperance", "Last Lecture": "Grim", "Open Vault": "The Emperor", "Wide Margin": "Ankh", "Copper Bath": "Aura", "Hot Press": "The Magician", "Mirror Bath": "The Wheel of Fortune", "Wide Press": "Ectoplasm", "Deep Cut": "Hex", "Lost Mould": "Sigil", "Last Casting": "Death",
}
var planetNames = [...]string{"Pluto", "Mercury", "Uranus", "Venus", "Saturn", "Jupiter", "Earth", "Mars", "Neptune", "Planet X", "Ceres", "Eris"}

func init() {
	for i, b := range handBases {
		consumableNames["Study of "+b.Name] = planetNames[i]
	}
}
func (s supply) cardFamily() string {
	if s.Action == "level" {
		return "Planet"
	}
	switch s.Name {
	case "Familiar", "Grim", "Incantation", "Talisman", "Aura", "Wraith", "Sigil", "Ouija", "Ectoplasm", "Immolate", "Ankh", "Deja Vu", "Hex", "Trance", "Medium", "Cryptid", "The Soul", "Black Hole":
		return "Spectral"
	}
	return "Tarot"
}
func (m *model) supplyFamily(s supply) string {
	if m.legacyPacks {
		switch s.Action {
		case "level":
			return "Planet"
		case "bargain":
			return "Spectral"
		default:
			return "Tarot"
		}
	}
	return s.cardFamily()
}

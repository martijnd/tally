package main

var stakeNames = [...]string{"White", "Red", "Green", "Black", "Blue", "Purple", "Orange", "Gold"}

var planetNames = [...]string{"Pluto", "Mercury", "Uranus", "Venus", "Saturn", "Jupiter", "Earth", "Mars", "Neptune", "Planet X", "Ceres", "Eris"}

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

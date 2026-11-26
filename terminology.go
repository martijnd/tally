package main

var stakeNames = [...]string{"White", "Red", "Green", "Black", "Blue", "Purple", "Orange", "Gold"}

func (s supply) cardFamily() string {
	switch s.Action {
	case "level":
		return "Planet"
	case "bargain":
		return "Spectral"
	default:
		return "Tarot"
	}
}

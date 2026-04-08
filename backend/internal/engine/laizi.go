package engine

import "github.com/mahjong/backend/internal/models"

// DetermineLaizi flips the indicator tile from the wall and returns
// the indicator tile code and the resulting laizi (wild card) tile code.
func DetermineLaizi(wall *Wall) (indicator models.TileCode, laizi models.TileCode, ok bool) {
	t, drawn := wall.Draw()
	if !drawn {
		return "", "", false
	}

	indicator = t.Code
	laizi, exists := models.LaiziSequence[indicator]
	if !exists {
		return "", "", false
	}

	return indicator, laizi, true
}

// CountLaizi counts how many laizi tiles are in a hand.
func CountLaizi(hand []models.TileCode, laizi models.TileCode) int {
	count := 0
	for _, t := range hand {
		if t == laizi {
			count++
		}
	}
	return count
}

// SeparateLaizi splits a hand into non-laizi tiles and laizi count.
func SeparateLaizi(hand []models.TileCode, laizi models.TileCode) (regular []models.TileCode, laiziCount int) {
	regular = make([]models.TileCode, 0, len(hand))
	for _, t := range hand {
		if t == laizi {
			laiziCount++
		} else {
			regular = append(regular, t)
		}
	}
	return regular, laiziCount
}

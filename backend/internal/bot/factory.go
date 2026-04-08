package bot

import (
	"math/rand"

	"github.com/mahjong/backend/internal/models"
)

// NewStrategy creates a Strategy for the given difficulty level.
func NewStrategy(difficulty models.BotDifficulty) Strategy {
	switch difficulty {
	case models.BotDifficultyEasy:
		return &EasyStrategy{}
	case models.BotDifficultyMedium:
		return &MediumStrategy{}
	case models.BotDifficultyHard:
		return &HardStrategy{}
	default:
		return &EasyStrategy{}
	}
}

// RandomDifficulty returns a random difficulty level.
func RandomDifficulty() models.BotDifficulty {
	options := []models.BotDifficulty{
		models.BotDifficultyEasy,
		models.BotDifficultyMedium,
		models.BotDifficultyHard,
	}
	return options[rand.Intn(len(options))]
}

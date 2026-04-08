package bot

import (
	"math/rand"

	"github.com/mahjong/backend/internal/models"
)

// EasyStrategy makes random valid moves. Always declares hu when possible.
type EasyStrategy struct{}

func (s *EasyStrategy) ChooseTurnAction(ctx GameContext) TurnAction {
	// Always hu if possible
	if ctx.CanHu {
		return TurnAction{Type: "hu"}
	}

	// Pick a random tile from hand to discard
	hand := ctx.Hand
	if ctx.DrawnTile != nil {
		hand = append(append([]models.TileCode{}, hand...), *ctx.DrawnTile)
	}

	if len(hand) == 0 {
		return TurnAction{Type: "discard"}
	}

	tile := hand[rand.Intn(len(hand))]
	return TurnAction{Type: "discard", Tile: tile}
}

func (s *EasyStrategy) ChooseReaction(ctx GameContext) ReactionAction {
	// Always hu if available
	for _, action := range ctx.AvailableActions {
		if action == "hu" {
			return ReactionAction{Type: "hu"}
		}
	}

	// Otherwise always pass
	return ReactionAction{Type: "pass"}
}

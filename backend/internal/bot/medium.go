package bot

import (
	"math/rand"

	"github.com/mahjong/backend/internal/engine"
	"github.com/mahjong/backend/internal/models"
)

// MediumStrategy uses simple heuristics: keep pairs, discard isolated tiles,
// always pong/gang, always hu.
type MediumStrategy struct{}

func (s *MediumStrategy) ChooseTurnAction(ctx GameContext) TurnAction {
	if ctx.CanHu {
		return TurnAction{Type: "hu"}
	}

	// Take closed gang if available
	if len(ctx.CanClosedGang) > 0 {
		return TurnAction{Type: "closed_gang", Tile: ctx.CanClosedGang[0]}
	}
	// Take add gang if available
	if len(ctx.CanAddGang) > 0 {
		return TurnAction{Type: "add_gang", Tile: ctx.CanAddGang[0]}
	}

	// Build full hand for evaluation
	hand := ctx.Hand
	if ctx.DrawnTile != nil {
		hand = append(append([]models.TileCode{}, hand...), *ctx.DrawnTile)
	}

	if len(hand) == 0 {
		return TurnAction{Type: "discard"}
	}

	// Score each tile by "isolation" — higher score = more disposable
	bestTile := hand[0]
	bestScore := -1

	for _, tile := range hand {
		score := isolationScore(tile, hand, ctx.LaiziTile)
		if score > bestScore {
			bestScore = score
			bestTile = tile
		}
	}

	return TurnAction{Type: "discard", Tile: bestTile}
}

func (s *MediumStrategy) ChooseReaction(ctx GameContext) ReactionAction {
	for _, action := range ctx.AvailableActions {
		if action == "hu" {
			return ReactionAction{Type: "hu"}
		}
	}

	for _, action := range ctx.AvailableActions {
		if action == "gang" {
			return ReactionAction{Type: "gang"}
		}
	}

	for _, action := range ctx.AvailableActions {
		if action == "pong" {
			return ReactionAction{Type: "pong"}
		}
	}

	// Chi sometimes (50% chance if available)
	for _, action := range ctx.AvailableActions {
		if action == "chi" && len(ctx.ChiOptions) > 0 && rand.Intn(2) == 0 {
			opt := ctx.ChiOptions[0]
			return ReactionAction{Type: "chi", ChiTiles: opt}
		}
	}

	return ReactionAction{Type: "pass"}
}

// isolationScore returns how "isolated" a tile is in the hand.
// Higher score = tile contributes less to the hand (better discard candidate).
// Never discard laizi tiles.
func isolationScore(tile models.TileCode, hand []models.TileCode, laiziTile models.TileCode) int {
	if tile == laiziTile {
		return -100 // never discard laizi
	}

	freq := engine.TileCodesToMap(hand)
	score := 10

	// Pairs and triples are valuable — penalize discarding them
	count := freq[tile]
	if count >= 3 {
		score -= 6
	} else if count >= 2 {
		score -= 3
	}

	// For suited tiles, check adjacency
	if engine.IsSuited(tile) {
		v := engine.TileValue(tile)
		suit := engine.TileSuit(tile)
		makeTile := func(val int) models.TileCode {
			return models.TileCode(string(rune('0'+val)) + string(suit))
		}

		// Check for adjacent tiles (sequence potential)
		adjacent := 0
		if v > 1 && freq[makeTile(v-1)] > 0 {
			adjacent++
		}
		if v < 9 && freq[makeTile(v+1)] > 0 {
			adjacent++
		}
		if v > 2 && freq[makeTile(v-2)] > 0 {
			adjacent++
		}
		if v < 8 && freq[makeTile(v+2)] > 0 {
			adjacent++
		}
		score -= adjacent * 2
	}

	// Honor tiles with no duplicates are more isolated
	if !engine.IsSuited(tile) && count == 1 {
		score += 2
	}

	return score
}

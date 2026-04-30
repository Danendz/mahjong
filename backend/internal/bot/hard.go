package bot

import (
	"github.com/mahjong/backend/internal/engine"
	"github.com/mahjong/backend/internal/models"
)

// HardStrategy uses tile counting, hand potential evaluation, and defensive play.
type HardStrategy struct{}

func (s *HardStrategy) ChooseTurnAction(ctx GameContext) TurnAction {
	if ctx.CanHu {
		return TurnAction{Type: "hu"}
	}

	// Take closed gang if available (usually beneficial)
	if len(ctx.CanClosedGang) > 0 {
		return TurnAction{Type: "closed_gang", Tile: ctx.CanClosedGang[0]}
	}
	// Take add gang if available
	if len(ctx.CanAddGang) > 0 {
		return TurnAction{Type: "add_gang", Tile: ctx.CanAddGang[0]}
	}

	hand := ctx.Hand
	if ctx.DrawnTile != nil {
		hand = append(append([]models.TileCode{}, hand...), *ctx.DrawnTile)
	}

	if len(hand) == 0 {
		return TurnAction{Type: "discard"}
	}

	// Count all visible tiles
	visible := s.countVisibleTiles(ctx)

	// Evaluate each possible discard: pick the one that maximizes winning outs
	bestTile := hand[0]
	bestOuts := -1
	bestSafe := 0

	for _, candidate := range hand {
		if candidate == ctx.LaiziTile {
			continue // never discard laizi
		}

		remaining := removeTile(hand, candidate)
		outs := len(engine.FindWinningDiscards(remaining, ctx.Melds, ctx.LaiziTile))

		// Count how "safe" this discard is (tiles already visible = less likely to be needed by others)
		safe := visible[candidate]

		if outs > bestOuts || (outs == bestOuts && safe > bestSafe) {
			bestOuts = outs
			bestTile = candidate
			bestSafe = safe
		}
	}

	// Defensive mode: if we have 0 outs and wall is low, prefer safe tiles
	if bestOuts == 0 && ctx.WallRemaining < 30 {
		bestTile = s.safestDiscard(hand, visible, ctx.LaiziTile)
	}

	return TurnAction{Type: "discard", Tile: bestTile}
}

func (s *HardStrategy) ChooseReaction(ctx GameContext) ReactionAction {
	for _, action := range ctx.AvailableActions {
		if action == "hu" {
			return ReactionAction{Type: "hu"}
		}
	}

	// For gang: always take it (gets a replacement draw)
	for _, action := range ctx.AvailableActions {
		if action == "gang" {
			return ReactionAction{Type: "gang"}
		}
	}

	// For pong/chi: only take if it improves hand potential
	hand := ctx.Hand

	for _, action := range ctx.AvailableActions {
		if action == "pong" {
			// Simulate pong: remove 2 matching tiles, evaluate remaining hand
			afterPong := removeTileN(hand, ctx.DiscardedTile, 2)
			meldsAfter := append(append([]models.MeldInfo{}, ctx.Melds...), models.MeldInfo{
				Type:  models.MeldPong,
				Tiles: []models.TileCode{ctx.DiscardedTile, ctx.DiscardedTile, ctx.DiscardedTile},
			})
			outsBefore := len(engine.FindWinningDiscards(hand, ctx.Melds, ctx.LaiziTile))
			outsAfter := len(engine.FindWinningDiscards(afterPong, meldsAfter, ctx.LaiziTile))
			if outsAfter >= outsBefore {
				return ReactionAction{Type: "pong"}
			}
		}
	}

	for _, action := range ctx.AvailableActions {
		if action == "chi" && len(ctx.ChiOptions) > 0 {
			// Try the first chi option
			opt := ctx.ChiOptions[0]
			afterChi := removeTile(removeTile(hand, opt[0]), opt[1])
			meldsAfter := append(append([]models.MeldInfo{}, ctx.Melds...), models.MeldInfo{
				Type:  models.MeldChi,
				Tiles: []models.TileCode{opt[0], opt[1], ctx.DiscardedTile},
			})
			outsBefore := len(engine.FindWinningDiscards(hand, ctx.Melds, ctx.LaiziTile))
			outsAfter := len(engine.FindWinningDiscards(afterChi, meldsAfter, ctx.LaiziTile))
			if outsAfter >= outsBefore {
				return ReactionAction{Type: "chi", ChiTiles: opt}
			}
		}
	}

	return ReactionAction{Type: "pass"}
}

// countVisibleTiles counts how many copies of each tile code are visible
// (in discards and open melds of all players).
func (s *HardStrategy) countVisibleTiles(ctx GameContext) map[models.TileCode]int {
	counts := make(map[models.TileCode]int)

	for seat := range 4 {
		for _, tile := range ctx.Discards[seat] {
			counts[tile]++
		}
		for _, meld := range ctx.OpenMelds[seat] {
			for _, tile := range meld.Tiles {
				counts[tile]++
			}
		}
	}

	return counts
}

// safestDiscard finds the tile in hand with the most visible copies (least likely needed by opponents).
func (s *HardStrategy) safestDiscard(hand []models.TileCode, visible map[models.TileCode]int, laiziTile models.TileCode) models.TileCode {
	best := hand[0]
	bestSafe := -1

	for _, tile := range hand {
		if tile == laiziTile {
			continue
		}
		safe := visible[tile]
		if safe > bestSafe {
			bestSafe = safe
			best = tile
		}
	}

	return best
}

// removeTile returns a copy of hand with one instance of the given tile removed.
func removeTile(hand []models.TileCode, tile models.TileCode) []models.TileCode {
	result := make([]models.TileCode, 0, len(hand)-1)
	removed := false
	for _, t := range hand {
		if t == tile && !removed {
			removed = true
			continue
		}
		result = append(result, t)
	}
	return result
}

// removeTileN returns a copy of hand with n instances of the given tile removed.
func removeTileN(hand []models.TileCode, tile models.TileCode, n int) []models.TileCode {
	result := make([]models.TileCode, 0, len(hand)-n)
	removed := 0
	for _, t := range hand {
		if t == tile && removed < n {
			removed++
			continue
		}
		result = append(result, t)
	}
	return result
}

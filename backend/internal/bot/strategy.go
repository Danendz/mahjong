package bot

import "github.com/mahjong/backend/internal/models"

// TurnAction represents a bot's decision during its turn.
type TurnAction struct {
	Type string          // "discard", "closed_gang", "add_gang", "hu"
	Tile models.TileCode // tile to discard or gang
}

// ReactionAction represents a bot's decision in response to a discard.
type ReactionAction struct {
	Type     string             // "hu", "gang", "pong", "chi", "pass"
	ChiTiles [2]models.TileCode // only used for chi
}

// GameContext provides the bot with all visible game information needed to make a decision.
type GameContext struct {
	Hand          []models.TileCode
	DrawnTile     *models.TileCode
	Melds         []models.MeldInfo
	Discards      [4][]models.TileCode // all players' discards
	OpenMelds     [4][]models.MeldInfo // all players' open melds
	LaiziTile     models.TileCode
	WallRemaining int
	Seat          int

	// For reactions
	DiscardedTile models.TileCode
	DiscardedFrom int

	// Available options
	AvailableActions []string
	ChiOptions       [][2]models.TileCode

	// For turn actions
	CanClosedGang []models.TileCode
	CanAddGang    []models.TileCode
	CanHu         bool
}

// Strategy defines the interface for bot decision-making.
type Strategy interface {
	ChooseTurnAction(ctx GameContext) TurnAction
	ChooseReaction(ctx GameContext) ReactionAction
}

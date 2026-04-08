package engine

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mahjong/backend/internal/models"
)

// GamePhase represents the current phase of the game.
type GamePhase string

const (
	PhaseWaiting          GamePhase = "waiting"
	PhaseDealing          GamePhase = "dealing"
	PhasePlayerTurn       GamePhase = "player_turn"
	PhaseAwaitingReaction GamePhase = "awaiting_reaction"
	PhaseAwaitingRobKong  GamePhase = "awaiting_rob_kong"
	PhaseRoundEnd         GamePhase = "round_end"
	PhaseFinished         GamePhase = "finished"
)

// ReactionType represents a player's declared reaction to a discard.
type ReactionType string

const (
	ReactionNone ReactionType = ""
	ReactionPass ReactionType = "pass"
	ReactionChi  ReactionType = "chi"
	ReactionPong ReactionType = "pong"
	ReactionGang ReactionType = "gang"
	ReactionHu   ReactionType = "hu"
)

// reactionPriority returns the priority of a reaction (higher = takes precedence).
func reactionPriority(r ReactionType) int {
	switch r {
	case ReactionHu:
		return 4
	case ReactionGang:
		return 3
	case ReactionPong:
		return 2
	case ReactionChi:
		return 1
	case ReactionPass:
		return 0
	default:
		return -1
	}
}

// PlayerReaction holds a player's declared reaction and associated data.
type PlayerReaction struct {
	Type     ReactionType
	ChiTiles [2]models.TileCode // tiles from hand used for chi
}

// PlayerState holds the state of a single player within the game.
type PlayerState struct {
	Seat      int
	Hand      []models.TileCode // tiles in hand (hidden)
	Melds     []models.MeldInfo // open melds
	Discards  []models.TileCode // tiles this player has discarded
	DrawnTile *models.TileCode  // the tile just drawn (nil if not their turn)
	Connected bool
}

// TileCount returns the number of tiles in the player's concealed hand.
func (p *PlayerState) TileCount() int {
	return len(p.Hand)
}

// RemoveTileFromHand removes one instance of a tile from the hand.
// Returns an error if the tile is not found.
func (p *PlayerState) RemoveTileFromHand(tile models.TileCode) error {
	for i, t := range p.Hand {
		if t == tile {
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("tile %s not in hand", tile)
}

// GameEvent represents an event that occurred in the game (for event sourcing).
type GameEvent struct {
	Seq        int                    `json:"seq"`
	Type       string                 `json:"type"`
	PlayerSeat int                    `json:"player_seat,omitempty"`
	Payload    map[string]interface{} `json:"payload"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Game holds the complete state of a single game round.
type Game struct {
	mu sync.RWMutex

	// Game configuration
	Config    models.RoomConfig
	DealerSeat int

	// Game state
	Phase         GamePhase
	Wall          *Wall
	Players       [4]*PlayerState
	CurrentTurn   int // seat index of the player whose turn it is
	LaiziIndicator models.TileCode
	LaiziTile     models.TileCode

	// Reaction state
	LastDiscard     *models.TileCode
	LastDiscardSeat int
	Reactions       [4]*PlayerReaction
	ReactionsNeeded map[int]bool // seats that still need to react

	// Kong state for rob-kong detection
	PendingKongSeat int
	PendingKongTile models.TileCode

	// Round tracking
	Round       int
	Scores      [4]int
	IsKongDraw  bool // was the current draw from a kong replacement

	// Events
	Events []GameEvent
	SeqCounter int

	// Callbacks
	OnEvent func(GameEvent) // called when an event is recorded
}

// NewGame creates a new game with the given config and initial scores.
func NewGame(config models.RoomConfig, dealerSeat int, scores [4]int, round int) *Game {
	g := &Game{
		Config:     config,
		DealerSeat: dealerSeat,
		Phase:      PhaseDealing,
		Scores:     scores,
		Round:      round,
	}
	for i := range 4 {
		g.Players[i] = &PlayerState{
			Seat:      i,
			Hand:      make([]models.TileCode, 0, 14),
			Melds:     make([]models.MeldInfo, 0),
			Discards:  make([]models.TileCode, 0),
			Connected: true,
		}
	}
	return g
}

// Deal shuffles the wall, deals hands, and determines the laizi.
func (g *Game) Deal() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Phase != PhaseDealing {
		return errors.New("not in dealing phase")
	}

	// Create and shuffle wall
	g.Wall = NewWall()
	g.Wall.Shuffle()

	// Deal hands
	hands, ok := DealHands(g.Wall, g.DealerSeat)
	if !ok {
		return errors.New("failed to deal hands")
	}

	for i := range 4 {
		g.Players[i].Hand = TilesToCodes(hands[i])
	}

	// Determine laizi
	indicator, laizi, ok := DetermineLaizi(g.Wall)
	if !ok {
		return errors.New("failed to determine laizi")
	}
	g.LaiziIndicator = indicator
	g.LaiziTile = laizi

	// Record deal event
	g.recordEvent("game_start", -1, map[string]interface{}{
		"dealer_seat":     g.DealerSeat,
		"laizi_indicator": string(indicator),
		"laizi_tile":      string(laizi),
		"wall_remaining":  g.Wall.Remaining(),
	})

	// Dealer starts
	g.CurrentTurn = g.DealerSeat
	g.Phase = PhasePlayerTurn

	return nil
}

// Discard processes a player discarding a tile.
func (g *Game) Discard(seat int, tile models.TileCode) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Phase != PhasePlayerTurn {
		return errors.New("not in player turn phase")
	}
	if seat != g.CurrentTurn {
		return fmt.Errorf("not your turn (current: %d, you: %d)", g.CurrentTurn, seat)
	}

	player := g.Players[seat]

	// Remove tile from hand
	if err := player.RemoveTileFromHand(tile); err != nil {
		return err
	}
	player.DrawnTile = nil
	player.Discards = append(player.Discards, tile)

	g.recordEvent("tile_discarded", seat, map[string]interface{}{
		"tile": string(tile),
	})

	// Check if any other player can react
	g.LastDiscard = &tile
	g.LastDiscardSeat = seat
	g.IsKongDraw = false

	if g.checkForReactions(tile, seat) {
		g.Phase = PhaseAwaitingReaction
	} else {
		// No reactions possible, next player's turn
		g.advanceTurn()
	}

	return nil
}

// DeclareReaction processes a player's reaction to a discard.
func (g *Game) DeclareReaction(seat int, reaction PlayerReaction) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Phase != PhaseAwaitingReaction && g.Phase != PhaseAwaitingRobKong {
		return errors.New("not awaiting reactions")
	}

	if !g.ReactionsNeeded[seat] {
		return errors.New("reaction not expected from this seat")
	}

	g.Reactions[seat] = &reaction
	delete(g.ReactionsNeeded, seat)

	g.recordEvent(string(reaction.Type), seat, map[string]interface{}{})

	// Check if all reactions are in
	if len(g.ReactionsNeeded) == 0 {
		return g.resolveReactions()
	}

	return nil
}

// DeclareClosedKong processes a player declaring a concealed kong on their turn.
func (g *Game) DeclareClosedKong(seat int, tile models.TileCode) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Phase != PhasePlayerTurn || seat != g.CurrentTurn {
		return errors.New("not your turn")
	}

	player := g.Players[seat]
	count := 0
	for _, t := range player.Hand {
		if t == tile {
			count++
		}
	}
	if count < 4 {
		return fmt.Errorf("need 4 copies of %s for closed kong, have %d", tile, count)
	}

	// Remove 4 tiles from hand
	newHand := make([]models.TileCode, 0, len(player.Hand)-4)
	removed := 0
	for _, t := range player.Hand {
		if t == tile && removed < 4 {
			removed++
			continue
		}
		newHand = append(newHand, t)
	}
	player.Hand = newHand

	player.Melds = append(player.Melds, models.MeldInfo{
		Type:  models.MeldClosedGang,
		Tiles: []models.TileCode{tile, tile, tile, tile},
	})

	g.recordEvent("closed_gang_declared", seat, map[string]interface{}{
		"tile": string(tile),
	})

	// Draw replacement from back of wall
	return g.drawKongReplacement(seat)
}

// DeclareAddKong processes a player adding to an existing open pong.
func (g *Game) DeclareAddKong(seat int, tile models.TileCode) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Phase != PhasePlayerTurn || seat != g.CurrentTurn {
		return errors.New("not your turn")
	}

	player := g.Players[seat]

	// Find the pong meld
	meldIdx := -1
	for i, m := range player.Melds {
		if m.Type == models.MeldPong && m.Tiles[0] == tile {
			meldIdx = i
			break
		}
	}
	if meldIdx == -1 {
		return fmt.Errorf("no open pong of %s to add to", tile)
	}

	// Remove tile from hand
	if err := player.RemoveTileFromHand(tile); err != nil {
		return err
	}

	// Before upgrading, check for 抢杠胡
	g.PendingKongSeat = seat
	g.PendingKongTile = tile

	// Check if anyone can rob the kong
	hasRob := false
	g.ReactionsNeeded = make(map[int]bool)
	g.Reactions = [4]*PlayerReaction{}
	for s := range 4 {
		if s == seat {
			continue
		}
		if CanWinWithTile(g.Players[s].Hand, tile, g.LaiziTile).IsWin {
			g.ReactionsNeeded[s] = true
			hasRob = true
		}
	}

	if hasRob {
		g.Phase = PhaseAwaitingRobKong
		g.recordEvent("add_gang_declared", seat, map[string]interface{}{
			"tile":             string(tile),
			"awaiting_rob_kong": true,
		})
		return nil
	}

	// No one can rob — upgrade the meld
	player.Melds[meldIdx] = models.MeldInfo{
		Type:  models.MeldAddGang,
		Tiles: []models.TileCode{tile, tile, tile, tile},
	}

	g.recordEvent("add_gang_declared", seat, map[string]interface{}{
		"tile": string(tile),
	})

	return g.drawKongReplacement(seat)
}

// DeclareSelfDrawWin processes a player declaring a self-draw win.
func (g *Game) DeclareSelfDrawWin(seat int) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Phase != PhasePlayerTurn || seat != g.CurrentTurn {
		return errors.New("not your turn")
	}

	player := g.Players[seat]
	analysis := IsWinningHand(player.Hand, g.LaiziTile)
	if !analysis.IsWin {
		return errors.New("hand is not a winning hand")
	}

	winCtx := WinContext{
		Hand:         player.Hand,
		Melds:        player.Melds,
		WinType:      WinSelfDraw,
		LaiziTile:    g.LaiziTile,
		Analysis:     analysis,
		OpenCallMode: g.Config.OpenCallMode,
		ScoreCap:     g.Config.ScoreCap,
		IsKongDraw:   g.IsKongDraw,
		IsLastTile:   g.Wall.Remaining() == 0,
	}

	return g.endRound(seat, winCtx)
}

// GetPlayerView returns the game state visible to a specific player.
func (g *Game) GetPlayerView(seat int) PlayerView {
	g.mu.RLock()
	defer g.mu.RUnlock()

	view := PlayerView{
		Phase:          g.Phase,
		YourSeat:       seat,
		YourHand:       g.Players[seat].Hand,
		DrawnTile:      g.Players[seat].DrawnTile,
		OpenMelds:      make(map[int][]models.MeldInfo),
		Discards:       make(map[int][]models.TileCode),
		TileCounts:     make(map[int]int),
		CurrentTurn:    g.CurrentTurn,
		LaiziIndicator: g.LaiziIndicator,
		LaiziTile:      g.LaiziTile,
		WallRemaining:  g.Wall.Remaining(),
		Scores:         g.Scores,
		DealerSeat:     g.DealerSeat,
	}

	for i := range 4 {
		view.OpenMelds[i] = g.Players[i].Melds
		view.Discards[i] = g.Players[i].Discards
		view.TileCounts[i] = g.Players[i].TileCount()
	}

	return view
}

// PlayerView is the game state as seen by a specific player.
type PlayerView struct {
	Phase          GamePhase
	YourSeat       int
	YourHand       []models.TileCode
	DrawnTile      *models.TileCode
	OpenMelds      map[int][]models.MeldInfo
	Discards       map[int][]models.TileCode
	TileCounts     map[int]int
	CurrentTurn    int
	LaiziIndicator models.TileCode
	LaiziTile      models.TileCode
	WallRemaining  int
	Scores         [4]int
	DealerSeat     int
}

// GetAvailableActions returns what actions a player can take right now.
func (g *Game) GetAvailableActions(seat int) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var actions []string

	if g.Phase == PhasePlayerTurn && seat == g.CurrentTurn {
		player := g.Players[seat]

		// Can always discard
		actions = append(actions, "discard")

		// Check closed kongs
		if len(FindClosedGangs(player.Hand)) > 0 {
			actions = append(actions, "closed_gang")
		}

		// Check add kongs
		if len(FindAddGangs(player.Hand, player.Melds)) > 0 {
			actions = append(actions, "add_gang")
		}

		// Check self-draw win
		if IsWinningHand(player.Hand, g.LaiziTile).IsWin {
			actions = append(actions, "hu")
		}
	}

	if (g.Phase == PhaseAwaitingReaction || g.Phase == PhaseAwaitingRobKong) && g.ReactionsNeeded[seat] {
		actions = append(actions, "pass")

		if g.Phase == PhaseAwaitingRobKong {
			actions = append(actions, "hu")
		} else if g.LastDiscard != nil {
			discard := *g.LastDiscard
			player := g.Players[seat]

			if CanWinWithTile(player.Hand, discard, g.LaiziTile).IsWin {
				actions = append(actions, "hu")
			}
			if CanOpenGang(player.Hand, discard) {
				actions = append(actions, "gang")
			}
			if CanPong(player.Hand, discard) {
				actions = append(actions, "pong")
			}
			// Chi only from left player
			if g.LastDiscardSeat == (seat+3)%4 {
				if len(CanChi(player.Hand, discard)) > 0 {
					actions = append(actions, "chi")
				}
			}
		}
	}

	return actions
}

// --- Internal methods ---

// checkForReactions checks if any player can react to a discard and sets up reaction state.
func (g *Game) checkForReactions(tile models.TileCode, discardSeat int) bool {
	g.ReactionsNeeded = make(map[int]bool)
	g.Reactions = [4]*PlayerReaction{}

	for seat := range 4 {
		if seat == discardSeat {
			continue
		}
		player := g.Players[seat]

		canReact := false
		if CanWinWithTile(player.Hand, tile, g.LaiziTile).IsWin {
			canReact = true
		}
		if CanOpenGang(player.Hand, tile) {
			canReact = true
		}
		if CanPong(player.Hand, tile) {
			canReact = true
		}
		// Chi only from left player
		if discardSeat == (seat+3)%4 && len(CanChi(player.Hand, tile)) > 0 {
			canReact = true
		}

		if canReact {
			g.ReactionsNeeded[seat] = true
		}
	}

	return len(g.ReactionsNeeded) > 0
}

// resolveReactions resolves all collected reactions by priority.
func (g *Game) resolveReactions() error {
	// Find highest priority reaction
	bestSeat := -1
	bestPriority := -1

	for seat := range 4 {
		r := g.Reactions[seat]
		if r == nil {
			continue
		}
		p := reactionPriority(r.Type)
		if p > bestPriority {
			bestPriority = p
			bestSeat = seat
		}
	}

	// All passed or no reactions
	if bestSeat == -1 || bestPriority <= 0 {
		g.advanceTurn()
		return nil
	}

	reaction := g.Reactions[bestSeat]
	player := g.Players[bestSeat]

	switch reaction.Type {
	case ReactionHu:
		if g.Phase == PhaseAwaitingRobKong {
			return g.resolveRobKong(bestSeat)
		}
		return g.resolveDiscardWin(bestSeat)

	case ReactionGang:
		tile := *g.LastDiscard
		// Remove 3 from hand
		removed := 0
		newHand := make([]models.TileCode, 0, len(player.Hand))
		for _, t := range player.Hand {
			if t == tile && removed < 3 {
				removed++
				continue
			}
			newHand = append(newHand, t)
		}
		player.Hand = newHand
		player.Melds = append(player.Melds, models.MeldInfo{
			Type:  models.MeldOpenGang,
			Tiles: []models.TileCode{tile, tile, tile, tile},
		})
		g.recordEvent("open_gang_declared", bestSeat, map[string]interface{}{
			"tile": string(tile),
		})
		g.CurrentTurn = bestSeat
		return g.drawKongReplacement(bestSeat)

	case ReactionPong:
		tile := *g.LastDiscard
		removed := 0
		newHand := make([]models.TileCode, 0, len(player.Hand))
		for _, t := range player.Hand {
			if t == tile && removed < 2 {
				removed++
				continue
			}
			newHand = append(newHand, t)
		}
		player.Hand = newHand
		player.Melds = append(player.Melds, models.MeldInfo{
			Type:  models.MeldPong,
			Tiles: []models.TileCode{tile, tile, tile},
		})
		g.recordEvent("pong_declared", bestSeat, map[string]interface{}{
			"tile": string(tile),
		})
		g.CurrentTurn = bestSeat
		g.Phase = PhasePlayerTurn
		return nil

	case ReactionChi:
		tile := *g.LastDiscard
		t1, t2 := reaction.ChiTiles[0], reaction.ChiTiles[1]
		if err := player.RemoveTileFromHand(t1); err != nil {
			return fmt.Errorf("chi failed: %w", err)
		}
		if err := player.RemoveTileFromHand(t2); err != nil {
			return fmt.Errorf("chi failed: %w", err)
		}
		chiTiles := []models.TileCode{t1, t2, tile}
		player.Melds = append(player.Melds, models.MeldInfo{
			Type:  models.MeldChi,
			Tiles: chiTiles,
		})
		g.recordEvent("chi_declared", bestSeat, map[string]interface{}{
			"tiles": []string{string(t1), string(t2), string(tile)},
		})
		g.CurrentTurn = bestSeat
		g.Phase = PhasePlayerTurn
		return nil
	}

	return nil
}

// resolveDiscardWin processes a win by claiming a discard.
func (g *Game) resolveDiscardWin(seat int) error {
	tile := *g.LastDiscard
	player := g.Players[seat]

	fullHand := make([]models.TileCode, len(player.Hand), len(player.Hand)+1)
	copy(fullHand, player.Hand)
	fullHand = append(fullHand, tile)

	analysis := IsWinningHand(fullHand, g.LaiziTile)

	winCtx := WinContext{
		Hand:         fullHand,
		Melds:        player.Melds,
		WinningTile:  tile,
		WinType:      WinDiscard,
		LaiziTile:    g.LaiziTile,
		Analysis:     analysis,
		OpenCallMode: g.Config.OpenCallMode,
		ScoreCap:     g.Config.ScoreCap,
	}

	return g.endRound(seat, winCtx)
}

// resolveRobKong processes a win by robbing a kong.
func (g *Game) resolveRobKong(seat int) error {
	tile := g.PendingKongTile
	player := g.Players[seat]

	fullHand := make([]models.TileCode, len(player.Hand), len(player.Hand)+1)
	copy(fullHand, player.Hand)
	fullHand = append(fullHand, tile)

	analysis := IsWinningHand(fullHand, g.LaiziTile)

	winCtx := WinContext{
		Hand:         fullHand,
		Melds:        player.Melds,
		WinningTile:  tile,
		WinType:      WinRobKong,
		LaiziTile:    g.LaiziTile,
		Analysis:     analysis,
		OpenCallMode: g.Config.OpenCallMode,
		ScoreCap:     g.Config.ScoreCap,
		IsRobKong:    true,
	}

	return g.endRound(seat, winCtx)
}

// drawKongReplacement draws a replacement tile from the back of the wall after a kong.
func (g *Game) drawKongReplacement(seat int) error {
	t, ok := g.Wall.DrawBack()
	if !ok {
		return g.endRoundDraw()
	}

	player := g.Players[seat]
	player.Hand = append(player.Hand, t.Code)
	player.DrawnTile = &t.Code
	g.IsKongDraw = true
	g.CurrentTurn = seat
	g.Phase = PhasePlayerTurn

	g.recordEvent("gang_replacement_drawn", seat, map[string]interface{}{
		"tile": string(t.Code),
	})

	return nil
}

// advanceTurn moves to the next player's turn, drawing a tile.
func (g *Game) advanceTurn() {
	nextSeat := (g.CurrentTurn + 1) % 4

	t, ok := g.Wall.Draw()
	if !ok {
		g.endRoundDraw()
		return
	}

	player := g.Players[nextSeat]
	player.Hand = append(player.Hand, t.Code)
	player.DrawnTile = &t.Code
	g.CurrentTurn = nextSeat
	g.Phase = PhasePlayerTurn
	g.IsKongDraw = false

	g.recordEvent("tile_drawn", nextSeat, map[string]interface{}{
		"tile": string(t.Code),
	})
}

// endRound ends the round with a winner.
func (g *Game) endRound(winnerSeat int, winCtx WinContext) error {
	score := CalculateScore(winCtx)

	discardSeat := g.LastDiscardSeat
	if winCtx.WinType == WinRobKong {
		discardSeat = g.PendingKongSeat
	}

	payments := CalculatePayments(winnerSeat, winCtx.WinType, discardSeat, score.TotalPerLoser)

	for seat, delta := range payments {
		g.Scores[seat] += delta
	}

	g.Phase = PhaseRoundEnd

	g.recordEvent("round_end", winnerSeat, map[string]interface{}{
		"result":        "hu",
		"win_type":      string(winCtx.WinType),
		"scoring":       score.Breakdown,
		"payments":      payments,
		"total_scores":  g.Scores,
	})

	return nil
}

// endRoundDraw ends the round as a draw (wall exhausted).
func (g *Game) endRoundDraw() error {
	g.Phase = PhaseRoundEnd

	g.recordEvent("round_end", -1, map[string]interface{}{
		"result":       "draw",
		"total_scores": g.Scores,
	})

	return nil
}

// AutoPass registers a pass reaction for a player (used on timeout).
func (g *Game) AutoPass(seat int) error {
	return g.DeclareReaction(seat, PlayerReaction{Type: ReactionPass})
}

// AutoDiscard discards the drawn tile for a player (used on timeout).
func (g *Game) AutoDiscard(seat int) error {
	g.mu.RLock()
	player := g.Players[seat]
	var tile models.TileCode
	if player.DrawnTile != nil {
		tile = *player.DrawnTile
	} else if len(player.Hand) > 0 {
		tile = player.Hand[len(player.Hand)-1]
	}
	g.mu.RUnlock()

	return g.Discard(seat, tile)
}

// recordEvent creates and stores a game event.
func (g *Game) recordEvent(eventType string, seat int, payload map[string]interface{}) {
	g.SeqCounter++
	event := GameEvent{
		Seq:        g.SeqCounter,
		Type:       eventType,
		PlayerSeat: seat,
		Payload:    payload,
		Timestamp:  time.Now(),
	}
	g.Events = append(g.Events, event)
	if g.OnEvent != nil {
		g.OnEvent(event)
	}
}

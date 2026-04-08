package room

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/mahjong/backend/internal/engine"
	"github.com/mahjong/backend/internal/models"
	"github.com/mahjong/backend/internal/ws"
)

// Room codes use this alphabet (no confusable chars: 0/O, 1/I/L)
const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
const codeLength = 6

// RoomStatus represents the current status of a room.
type RoomStatus string

const (
	StatusWaiting  RoomStatus = "waiting"
	StatusPlaying  RoomStatus = "playing"
	StatusFinished RoomStatus = "finished"
)

// Player represents a player in a room.
type Player struct {
	UserID       string
	SessionToken string
	Nickname     string
	Seat         int
	Ready        bool
	Connected    bool
}

// Room represents a game room.
type Room struct {
	mu          sync.RWMutex
	ID          string
	Code        string
	HostUserID  string
	Config      models.RoomConfig
	Status      RoomStatus
	Players     [4]*Player
	PlayerCount int
	Game        *engine.Game
	Round       int
	Scores      [4]int
	DealerSeat  int
	TurnTimer   *time.Timer
	CreatedAt   time.Time
}

// Manager manages all active rooms.
type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room // code -> room
	hub   *ws.Hub
}

// NewManager creates a new room manager.
func NewManager(hub *ws.Hub) *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
		hub:   hub,
	}
}

// CreateRoom creates a new room with a unique code.
func (m *Manager) CreateRoom(hostUserID, hostToken, nickname string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	code := m.generateUniqueCode()

	room := &Room{
		Code:       code,
		HostUserID: hostUserID,
		Status:     StatusWaiting,
		Config: models.RoomConfig{
			ScoreCap:      500,
			OpenCallMode:  models.OpenCallModeKouKou,
			TurnTimer:     15,
			ReactionTimer: 8,
			NumRounds:     8,
		},
		CreatedAt: time.Now(),
	}

	// Host takes seat 0
	room.Players[0] = &Player{
		UserID:       hostUserID,
		SessionToken: hostToken,
		Nickname:     nickname,
		Seat:         0,
		Ready:        false,
		Connected:    true,
	}
	room.PlayerCount = 1

	m.rooms[code] = room
	return room, nil
}

// JoinRoom adds a player to a room by code.
func (m *Manager) JoinRoom(code, userID, token, nickname string) (*Room, int, error) {
	m.mu.RLock()
	room, ok := m.rooms[code]
	m.mu.RUnlock()

	if !ok {
		return nil, -1, fmt.Errorf("room %s not found", code)
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Status != StatusWaiting {
		// Check if this is a reconnecting player
		for _, p := range room.Players {
			if p != nil && p.SessionToken == token {
				p.Connected = true
				return room, p.Seat, nil
			}
		}
		return nil, -1, fmt.Errorf("game already in progress")
	}

	// Check if player is already in the room
	for _, p := range room.Players {
		if p != nil && p.SessionToken == token {
			p.Connected = true
			p.Nickname = nickname
			return room, p.Seat, nil
		}
	}

	// Find empty seat
	seat := -1
	for i, p := range room.Players {
		if p == nil {
			seat = i
			break
		}
	}
	if seat == -1 {
		return nil, -1, fmt.Errorf("room is full")
	}

	room.Players[seat] = &Player{
		UserID:       userID,
		SessionToken: token,
		Nickname:     nickname,
		Seat:         seat,
		Ready:        false,
		Connected:    true,
	}
	room.PlayerCount++

	return room, seat, nil
}

// LeaveRoom removes a player from a room.
func (m *Manager) LeaveRoom(code string, seat int) {
	m.mu.RLock()
	room, ok := m.rooms[code]
	m.mu.RUnlock()
	if !ok {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Status == StatusWaiting && room.Players[seat] != nil {
		room.Players[seat] = nil
		room.PlayerCount--

		// If host left, assign new host or delete room
		if seat == 0 {
			newHost := -1
			for i, p := range room.Players {
				if p != nil {
					newHost = i
					break
				}
			}
			if newHost == -1 {
				m.mu.Lock()
				delete(m.rooms, code)
				m.mu.Unlock()
				return
			}
			room.HostUserID = room.Players[newHost].UserID
		}
	}
}

// SetReady marks a player as ready/not ready.
func (m *Manager) SetReady(code string, seat int) error {
	room := m.GetRoom(code)
	if room == nil {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Players[seat] == nil {
		return fmt.Errorf("no player at seat %d", seat)
	}
	room.Players[seat].Ready = !room.Players[seat].Ready
	return nil
}

// ConfigureRoom updates the room configuration (host only).
func (m *Manager) ConfigureRoom(code string, userID string, config models.RoomConfig) error {
	room := m.GetRoom(code)
	if room == nil {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.HostUserID != userID {
		return fmt.Errorf("only the host can configure the room")
	}
	if room.Status != StatusWaiting {
		return fmt.Errorf("cannot configure while game is in progress")
	}

	room.Config = config
	return nil
}

// StartGame starts the game if all 4 players are ready.
func (m *Manager) StartGame(code string, userID string) error {
	room := m.GetRoom(code)
	if room == nil {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.HostUserID != userID {
		return fmt.Errorf("only the host can start the game")
	}
	if room.Status != StatusWaiting {
		return fmt.Errorf("game already started")
	}
	if room.PlayerCount != 4 {
		return fmt.Errorf("need 4 players, have %d", room.PlayerCount)
	}
	for _, p := range room.Players {
		if p == nil || !p.Ready {
			return fmt.Errorf("all players must be ready")
		}
	}

	room.Status = StatusPlaying
	room.Round = 1
	room.DealerSeat = 0

	return m.startRound(room)
}

// startRound begins a new round within a room. Room must be locked by caller.
func (m *Manager) startRound(room *Room) error {
	game := engine.NewGame(room.Config, room.DealerSeat, room.Scores, room.Round)

	game.OnEvent = func(event engine.GameEvent) {
		log.Printf("[room %s] event: %s seat=%d", room.Code, event.Type, event.PlayerSeat)
	}

	if err := game.Deal(); err != nil {
		return fmt.Errorf("deal failed: %w", err)
	}

	room.Game = game

	// Send game_started to each player with their private hand
	for seat := range 4 {
		view := game.GetPlayerView(seat)
		seatVal := seat
		dealerSeat := room.DealerSeat
		wallRemaining := view.WallRemaining

		m.hub.SendToSeat(room.Code, seat, models.ServerMessage{
			Type:           models.MsgGameStarted,
			YourHand:       view.YourHand,
			DealerSeat:     &dealerSeat,
			LaiziIndicator: game.LaiziIndicator,
			LaiziTile:      game.LaiziTile,
			WallRemaining:  &wallRemaining,
			YourSeat:       &seatVal,
		})
	}

	// Notify dealer it's their turn
	m.sendTurnNotification(room)

	return nil
}

// HandleDiscard processes a discard action.
func (m *Manager) HandleDiscard(code string, seat int, tile models.TileCode) error {
	room := m.GetRoom(code)
	if room == nil {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Game == nil {
		return fmt.Errorf("no active game")
	}

	if err := room.Game.Discard(seat, tile); err != nil {
		return err
	}

	m.stopTurnTimer(room)

	// Broadcast the discard to all players
	wallRemaining := room.Game.Wall.Remaining()
	seatVal := seat
	m.hub.BroadcastToRoom(code, models.ServerMessage{
		Type:          models.MsgTileDiscarded,
		Seat:          &seatVal,
		Tile:          tile,
		WallRemaining: &wallRemaining,
	})

	// Handle resulting phase
	return m.handlePhaseTransition(room)
}

// HandleReaction processes a player's reaction to a discard.
func (m *Manager) HandleReaction(code string, seat int, reaction engine.PlayerReaction) error {
	room := m.GetRoom(code)
	if room == nil {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Game == nil {
		return fmt.Errorf("no active game")
	}

	if err := room.Game.DeclareReaction(seat, reaction); err != nil {
		return err
	}

	return m.handlePhaseTransition(room)
}

// HandleClosedKong processes a concealed kong declaration.
func (m *Manager) HandleClosedKong(code string, seat int, tile models.TileCode) error {
	room := m.GetRoom(code)
	if room == nil {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Game == nil {
		return fmt.Errorf("no active game")
	}

	m.stopTurnTimer(room)

	if err := room.Game.DeclareClosedKong(seat, tile); err != nil {
		return err
	}

	seatVal := seat
	m.hub.BroadcastToRoom(code, models.ServerMessage{
		Type:     models.MsgGangResult,
		Seat:     &seatVal,
		GangType: "closed",
	})

	return m.handlePhaseTransition(room)
}

// HandleAddKong processes an add kong declaration.
func (m *Manager) HandleAddKong(code string, seat int, tile models.TileCode) error {
	room := m.GetRoom(code)
	if room == nil {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Game == nil {
		return fmt.Errorf("no active game")
	}

	m.stopTurnTimer(room)

	if err := room.Game.DeclareAddKong(seat, tile); err != nil {
		return err
	}

	seatVal := seat
	m.hub.BroadcastToRoom(code, models.ServerMessage{
		Type:     models.MsgGangResult,
		Seat:     &seatVal,
		GangType: "add",
		Tile:     tile,
	})

	return m.handlePhaseTransition(room)
}

// HandleSelfDrawWin processes a self-draw win declaration.
func (m *Manager) HandleSelfDrawWin(code string, seat int) error {
	room := m.GetRoom(code)
	if room == nil {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Game == nil {
		return fmt.Errorf("no active game")
	}

	m.stopTurnTimer(room)

	if err := room.Game.DeclareSelfDrawWin(seat); err != nil {
		return err
	}

	return m.handlePhaseTransition(room)
}

// HandleDisconnect marks a player as disconnected.
func (m *Manager) HandleDisconnect(code string, seat int) {
	room := m.GetRoom(code)
	if room == nil {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Players[seat] != nil {
		room.Players[seat].Connected = false
	}

	seatVal := seat
	timeout := 120
	m.hub.BroadcastToRoom(code, models.ServerMessage{
		Type:           models.MsgPlayerDisconnected,
		Seat:           &seatVal,
		TimeoutSeconds: &timeout,
	})

	// Start disconnect timeout
	go func() {
		time.Sleep(120 * time.Second)
		room.mu.RLock()
		stillDisconnected := room.Players[seat] != nil && !room.Players[seat].Connected
		room.mu.RUnlock()

		if stillDisconnected && room.Game != nil {
			// Auto-play for disconnected player
			game := room.Game
			if game.Phase == engine.PhasePlayerTurn && game.CurrentTurn == seat {
				m.HandleAutoDiscard(code, seat)
			}
		}
	}()
}

// HandleReconnect restores a player's connection.
func (m *Manager) HandleReconnect(code string, client *ws.Client) error {
	room := m.GetRoom(code)
	if room == nil {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// Find the player by session token
	seat := -1
	for i, p := range room.Players {
		if p != nil && p.SessionToken == client.SessionToken {
			seat = i
			p.Connected = true
			break
		}
	}
	if seat == -1 {
		return fmt.Errorf("player not found in room")
	}

	client.Seat = seat

	// Send full game state
	if room.Game != nil {
		view := room.Game.GetPlayerView(seat)
		m.sendGameState(client, room, view)
	}

	seatVal := seat
	m.hub.BroadcastToRoom(code, models.ServerMessage{
		Type: models.MsgPlayerReconnected,
		Seat: &seatVal,
	})

	return nil
}

// HandleAutoDiscard auto-discards for a timed-out or disconnected player.
func (m *Manager) HandleAutoDiscard(code string, seat int) {
	room := m.GetRoom(code)
	if room == nil {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Game == nil {
		return
	}

	if err := room.Game.AutoDiscard(seat); err != nil {
		log.Printf("auto-discard error: %v", err)
		return
	}

	// Broadcast
	player := room.Game.Players[seat]
	if len(player.Discards) > 0 {
		tile := player.Discards[len(player.Discards)-1]
		wallRemaining := room.Game.Wall.Remaining()
		seatVal := seat
		m.hub.BroadcastToRoom(code, models.ServerMessage{
			Type:          models.MsgTileDiscarded,
			Seat:          &seatVal,
			Tile:          tile,
			WallRemaining: &wallRemaining,
		})
	}

	m.handlePhaseTransition(room)
}

// HandleAutoPass auto-passes for a timed-out player.
func (m *Manager) HandleAutoPass(code string, seat int) {
	room := m.GetRoom(code)
	if room == nil {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Game == nil {
		return
	}

	if err := room.Game.AutoPass(seat); err != nil {
		log.Printf("auto-pass error: %v", err)
		return
	}

	m.handlePhaseTransition(room)
}

// GetRoom returns a room by code.
func (m *Manager) GetRoom(code string) *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rooms[code]
}

// GetPlayerInfos returns player info for all players in a room.
func (r *Room) GetPlayerInfos() []models.PlayerInfo {
	var infos []models.PlayerInfo
	for _, p := range r.Players {
		if p != nil {
			infos = append(infos, models.PlayerInfo{
				Seat:      p.Seat,
				Nickname:  p.Nickname,
				Ready:     p.Ready,
				Connected: p.Connected,
			})
		}
	}
	return infos
}

// RoomInfo is the JSON-serializable snapshot of a room for the REST API.
type RoomInfo struct {
	Code        string              `json:"code"`
	Status      RoomStatus          `json:"status"`
	PlayerCount int                 `json:"player_count"`
	Players     []models.PlayerInfo `json:"players"`
	Config      models.RoomConfig   `json:"config"`
}

// GetInfo returns a thread-safe snapshot of the room for the REST API.
func (r *Room) GetInfo() RoomInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return RoomInfo{
		Code:        r.Code,
		Status:      r.Status,
		PlayerCount: r.PlayerCount,
		Players:     r.GetPlayerInfos(),
		Config:      r.Config,
	}
}

// --- Internal helpers ---

// handlePhaseTransition checks the current game phase and sends appropriate messages.
// Room must be locked by caller.
func (m *Manager) handlePhaseTransition(room *Room) error {
	game := room.Game
	if game == nil {
		return nil
	}

	switch game.Phase {
	case engine.PhasePlayerTurn:
		m.sendTurnNotification(room)
		m.startTurnTimer(room)

	case engine.PhaseAwaitingReaction:
		m.sendReactionPrompts(room)
		m.startReactionTimer(room)

	case engine.PhaseAwaitingRobKong:
		m.sendRobKongPrompts(room)
		m.startReactionTimer(room)

	case engine.PhaseRoundEnd:
		m.handleRoundEnd(room)
	}

	return nil
}

func (m *Manager) sendTurnNotification(room *Room) {
	game := room.Game
	seat := game.CurrentTurn
	player := game.Players[seat]

	actions := game.GetAvailableActions(seat)
	canGang := engine.FindClosedGangs(player.Hand)
	canGang = append(canGang, engine.FindAddGangs(player.Hand, player.Melds)...)
	canHu := engine.IsWinningHand(player.Hand, game.LaiziTile).IsWin

	timeLimit := room.Config.TurnTimer
	wallRemaining := game.Wall.Remaining()

	var drawnTile models.TileCode
	if player.DrawnTile != nil {
		drawnTile = *player.DrawnTile
	}

	_ = actions // available actions computed by client from can_gang/can_hu

	m.hub.SendToSeat(room.Code, seat, models.ServerMessage{
		Type:          models.MsgYourTurn,
		DrawnTile:     drawnTile,
		TimeLimit:     &timeLimit,
		WallRemaining: &wallRemaining,
		CanGang:       canGang,
		CanHu:         &canHu,
	})
}

func (m *Manager) sendReactionPrompts(room *Room) {
	game := room.Game
	tile := *game.LastDiscard
	fromSeat := game.LastDiscardSeat
	timeLimit := room.Config.ReactionTimer

	for seat, needed := range game.ReactionsNeeded {
		if !needed {
			continue
		}
		player := game.Players[seat]
		var available []string
		available = append(available, "pass")

		if engine.CanWinWithTile(player.Hand, tile, game.LaiziTile).IsWin {
			available = append(available, "hu")
		}
		if engine.CanOpenGang(player.Hand, tile) {
			available = append(available, "gang")
		}
		if engine.CanPong(player.Hand, tile) {
			available = append(available, "pong")
		}

		var chiOptions [][]models.TileCode
		if fromSeat == (seat+3)%4 {
			for _, opt := range engine.CanChi(player.Hand, tile) {
				chiOptions = append(chiOptions, []models.TileCode{opt[0], opt[1]})
			}
			if len(chiOptions) > 0 {
				available = append(available, "chi")
			}
		}

		from := fromSeat
		m.hub.SendToSeat(room.Code, seat, models.ServerMessage{
			Type:             models.MsgReactionPrompt,
			Tile:             tile,
			FromSeat:         &from,
			AvailableActions: available,
			ChiOptions:       chiOptions,
			TimeLimit:        &timeLimit,
		})
	}
}

func (m *Manager) sendRobKongPrompts(room *Room) {
	game := room.Game
	tile := game.PendingKongTile
	timeLimit := room.Config.ReactionTimer
	fromSeat := game.PendingKongSeat

	for seat := range game.ReactionsNeeded {
		from := fromSeat
		m.hub.SendToSeat(room.Code, seat, models.ServerMessage{
			Type:             models.MsgReactionPrompt,
			Tile:             tile,
			FromSeat:         &from,
			AvailableActions: []string{"hu", "pass"},
			TimeLimit:        &timeLimit,
		})
	}
}

func (m *Manager) handleRoundEnd(room *Room) {
	game := room.Game
	events := game.Events
	lastEvent := events[len(events)-1]

	// Extract round end info from last event
	result := "draw"
	if r, ok := lastEvent.Payload["result"].(string); ok {
		result = r
	}

	scoreDeltas := make(map[string]int)
	totalScores := make(map[string]int)
	for i := range 4 {
		key := fmt.Sprintf("%d", i)
		totalScores[key] = game.Scores[i]
	}

	// Build round_end message
	msg := models.ServerMessage{
		Type:        models.MsgRoundEnd,
		Result:      result,
		ScoreDeltas: scoreDeltas,
		TotalScores: totalScores,
	}

	if result == "hu" {
		winnerSeat := lastEvent.PlayerSeat
		msg.WinnerSeat = &winnerSeat
		msg.WinningHand = game.Players[winnerSeat].Hand

		if wt, ok := lastEvent.Payload["win_type"].(string); ok {
			msg.WinType = wt
		}
		if scoring, ok := lastEvent.Payload["scoring"].(models.ScoringBreakdown); ok {
			msg.Scoring = &scoring
		}
		if payments, ok := lastEvent.Payload["payments"].(map[int]int); ok {
			for seat, delta := range payments {
				scoreDeltas[fmt.Sprintf("%d", seat)] = delta
			}
			msg.ScoreDeltas = scoreDeltas
		}
	}

	m.hub.BroadcastToRoom(room.Code, msg)

	// Update room scores
	room.Scores = game.Scores

	// Check if more rounds
	room.Round++
	if room.Round > room.Config.NumRounds {
		room.Status = StatusFinished
		return
	}

	// Advance dealer and start next round after a delay
	room.DealerSeat = (room.DealerSeat + 1) % 4

	go func() {
		time.Sleep(5 * time.Second)
		room.mu.Lock()
		defer room.mu.Unlock()
		if room.Status == StatusPlaying {
			if err := m.startRound(room); err != nil {
				log.Printf("failed to start next round: %v", err)
			}
		}
	}()
}

func (m *Manager) sendGameState(client *ws.Client, room *Room, view engine.PlayerView) {
	openMelds := make(map[string][]models.MeldInfo)
	discards := make(map[string][]models.TileCode)
	tileCounts := make(map[string]int)
	totalScores := make(map[string]int)

	for i := range 4 {
		key := fmt.Sprintf("%d", i)
		openMelds[key] = view.OpenMelds[i]
		discards[key] = view.Discards[i]
		tileCounts[key] = view.TileCounts[i]
		totalScores[key] = view.Scores[i]
	}

	seat := view.YourSeat
	turn := view.CurrentTurn
	dealer := view.DealerSeat
	wallRemaining := view.WallRemaining

	client.SendMessage(models.ServerMessage{
		Type:            models.MsgGameState,
		YourSeat:        &seat,
		YourHand:        view.YourHand,
		OpenMelds:       openMelds,
		Discards:        discards,
		TileCounts:      tileCounts,
		CurrentTurnSeat: &turn,
		LaiziIndicator:  view.LaiziIndicator,
		LaiziTile:       view.LaiziTile,
		WallRemaining:   &wallRemaining,
		TotalScores:     totalScores,
		DealerSeat:      &dealer,
	})
}

func (m *Manager) startTurnTimer(room *Room) {
	m.stopTurnTimer(room)
	seat := room.Game.CurrentTurn
	duration := time.Duration(room.Config.TurnTimer) * time.Second
	room.TurnTimer = time.AfterFunc(duration, func() {
		m.HandleAutoDiscard(room.Code, seat)
	})
}

func (m *Manager) startReactionTimer(room *Room) {
	m.stopTurnTimer(room)
	duration := time.Duration(room.Config.ReactionTimer) * time.Second

	// Snapshot which seats need to react
	pendingSeats := make([]int, 0)
	for seat := range room.Game.ReactionsNeeded {
		pendingSeats = append(pendingSeats, seat)
	}

	room.TurnTimer = time.AfterFunc(duration, func() {
		for _, seat := range pendingSeats {
			m.HandleAutoPass(room.Code, seat)
		}
	})
}

func (m *Manager) stopTurnTimer(room *Room) {
	if room.TurnTimer != nil {
		room.TurnTimer.Stop()
		room.TurnTimer = nil
	}
}

func (m *Manager) generateUniqueCode() string {
	for {
		code := generateCode()
		if _, exists := m.rooms[code]; !exists {
			return code
		}
	}
}

func generateCode() string {
	b := make([]byte, codeLength)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
		b[i] = codeAlphabet[n.Int64()]
	}
	return string(b)
}

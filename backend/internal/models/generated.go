// AUTO-GENERATED from schema/*.schema.json — do not edit manually
// Run: scripts/generate-types.sh

package models

// ============ Tile Types ============

type TileCode string

const (
	Tile1m TileCode = "1m"
	Tile2m TileCode = "2m"
	Tile3m TileCode = "3m"
	Tile4m TileCode = "4m"
	Tile5m TileCode = "5m"
	Tile6m TileCode = "6m"
	Tile7m TileCode = "7m"
	Tile8m TileCode = "8m"
	Tile9m TileCode = "9m"
	Tile1s TileCode = "1s"
	Tile2s TileCode = "2s"
	Tile3s TileCode = "3s"
	Tile4s TileCode = "4s"
	Tile5s TileCode = "5s"
	Tile6s TileCode = "6s"
	Tile7s TileCode = "7s"
	Tile8s TileCode = "8s"
	Tile9s TileCode = "9s"
	Tile1p TileCode = "1p"
	Tile2p TileCode = "2p"
	Tile3p TileCode = "3p"
	Tile4p TileCode = "4p"
	Tile5p TileCode = "5p"
	Tile6p TileCode = "6p"
	Tile7p TileCode = "7p"
	Tile8p TileCode = "8p"
	Tile9p TileCode = "9p"
	TileWE TileCode = "we"
	TileWS TileCode = "ws"
	TileWW TileCode = "ww"
	TileWN TileCode = "wn"
	TileDZ TileCode = "dz"
	TileDF TileCode = "df"
	TileDB TileCode = "db"
)

// AllTileCodes contains all 34 unique tile codes
var AllTileCodes = []TileCode{
	Tile1m, Tile2m, Tile3m, Tile4m, Tile5m, Tile6m, Tile7m, Tile8m, Tile9m,
	Tile1s, Tile2s, Tile3s, Tile4s, Tile5s, Tile6s, Tile7s, Tile8s, Tile9s,
	Tile1p, Tile2p, Tile3p, Tile4p, Tile5p, Tile6p, Tile7p, Tile8p, Tile9p,
	TileWE, TileWS, TileWW, TileWN,
	TileDZ, TileDF, TileDB,
}

// Valid258Pairs contains valid pair tiles for winning hands
var Valid258Pairs = []TileCode{
	Tile2m, Tile5m, Tile8m,
	Tile2s, Tile5s, Tile8s,
	Tile2p, Tile5p, Tile8p,
}

// LaiziSequence maps indicator tile to laizi (wild card) tile
var LaiziSequence = map[TileCode]TileCode{
	Tile1m: Tile2m, Tile2m: Tile3m, Tile3m: Tile4m,
	Tile4m: Tile5m, Tile5m: Tile6m, Tile6m: Tile7m,
	Tile7m: Tile8m, Tile8m: Tile9m, Tile9m: Tile1m,
	Tile1s: Tile2s, Tile2s: Tile3s, Tile3s: Tile4s,
	Tile4s: Tile5s, Tile5s: Tile6s, Tile6s: Tile7s,
	Tile7s: Tile8s, Tile8s: Tile9s, Tile9s: Tile1s,
	Tile1p: Tile2p, Tile2p: Tile3p, Tile3p: Tile4p,
	Tile4p: Tile5p, Tile5p: Tile6p, Tile6p: Tile7p,
	Tile7p: Tile8p, Tile8p: Tile9p, Tile9p: Tile1p,
	TileWE: TileWS, TileWS: TileWW, TileWW: TileWN,
	TileWN: TileDF, // skip 红中
	TileDZ: TileDF, TileDF: TileDB, TileDB: TileWE,
}

// ============ Bot Types ============

type BotDifficulty string

const (
	BotDifficultyEasy   BotDifficulty = "easy"
	BotDifficultyMedium BotDifficulty = "medium"
	BotDifficultyHard   BotDifficulty = "hard"
)

// ============ Room Types ============

type OpenCallMode string

const (
	OpenCallModeKouKou OpenCallMode = "koukou" // 口口翻
	OpenCallModeKaiKou OpenCallMode = "kaikou" // 开口翻
)

type RoomConfig struct {
	ScoreCap      int          `json:"score_cap"`
	OpenCallMode  OpenCallMode `json:"open_call_mode"`
	TurnTimer     int          `json:"turn_timer"`
	ReactionTimer int          `json:"reaction_timer"`
	NumRounds     int          `json:"num_rounds"`
}

type PlayerInfo struct {
	Seat       int           `json:"seat"`
	Nickname   string        `json:"nickname"`
	Ready      bool          `json:"ready"`
	Connected  bool          `json:"connected"`
	IsBot      bool          `json:"is_bot,omitempty"`
	Difficulty BotDifficulty `json:"difficulty,omitempty"`
}

type MeldType string

const (
	MeldChi       MeldType = "chi"
	MeldPong      MeldType = "pong"
	MeldOpenGang  MeldType = "open_gang"
	MeldClosedGang MeldType = "closed_gang"
	MeldAddGang   MeldType = "add_gang"
)

type MeldInfo struct {
	Type  MeldType   `json:"type"`
	Tiles []TileCode `json:"tiles"`
}

type ScoringMultiplier struct {
	Reason string `json:"reason"`
	Value  int    `json:"value"`
}

type ScoringBreakdown struct {
	BasePoints   int                 `json:"base_points"`
	Multipliers  []ScoringMultiplier `json:"multipliers"`
	TotalPerLoser int               `json:"total_per_loser"`
	Capped       bool               `json:"capped"`
}

// ============ Client → Server Messages ============

type ClientMessageType string

const (
	MsgJoinRoom      ClientMessageType = "join_room"
	MsgLeaveRoom     ClientMessageType = "leave_room"
	MsgPlayerReady   ClientMessageType = "player_ready"
	MsgStartGame     ClientMessageType = "start_game"
	MsgConfigureRoom ClientMessageType = "configure_room"
	MsgDiscard       ClientMessageType = "discard"
	MsgChi           ClientMessageType = "chi"
	MsgPong          ClientMessageType = "pong"
	MsgGang          ClientMessageType = "gang"
	MsgHu               ClientMessageType = "hu"
	MsgPass             ClientMessageType = "pass"
	MsgAddBot           ClientMessageType = "add_bot"
	MsgRemoveBot        ClientMessageType = "remove_bot"
	MsgSetBotDifficulty ClientMessageType = "set_bot_difficulty"
)

type ClientMessage struct {
	Type         ClientMessageType `json:"type"`
	Code         string            `json:"code,omitempty"`
	Nickname     string            `json:"nickname,omitempty"`
	SessionToken string            `json:"session_token,omitempty"`
	Config       *RoomConfig       `json:"config,omitempty"`
	Tile         TileCode          `json:"tile,omitempty"`
	Tiles        []TileCode        `json:"tiles,omitempty"`
	GangType     string            `json:"gang_type,omitempty"`
	TargetSeat   *int              `json:"target_seat,omitempty"`
	Difficulty   BotDifficulty     `json:"difficulty,omitempty"`
}

// ============ Server → Client Messages ============

type ServerMessageType string

const (
	MsgRoomJoined         ServerMessageType = "room_joined"
	MsgPlayerJoined       ServerMessageType = "player_joined"
	MsgPlayerLeft         ServerMessageType = "player_left"
	MsgPlayerReadyServer  ServerMessageType = "player_ready"
	MsgConfigUpdated      ServerMessageType = "config_updated"
	MsgGameStarted        ServerMessageType = "game_started"
	MsgYourTurn           ServerMessageType = "your_turn"
	MsgTileDiscarded      ServerMessageType = "tile_discarded"
	MsgReactionPrompt     ServerMessageType = "reaction_prompt"
	MsgActionResolved     ServerMessageType = "action_resolved"
	MsgGangResult         ServerMessageType = "gang_result"
	MsgRoundEnd           ServerMessageType = "round_end"
	MsgGameState          ServerMessageType = "game_state"
	MsgPlayerDisconnected ServerMessageType = "player_disconnected"
	MsgPlayerReconnected  ServerMessageType = "player_reconnected"
	MsgBotAdded           ServerMessageType = "bot_added"
	MsgBotRemoved         ServerMessageType = "bot_removed"
	MsgBotDiffChanged     ServerMessageType = "bot_difficulty_changed"
	MsgError              ServerMessageType = "error"
)

type ServerMessage struct {
	Type ServerMessageType `json:"type"`

	// room_joined
	RoomID   string       `json:"room_id,omitempty"`
	RoomCode string       `json:"code,omitempty"`
	YourSeat *int         `json:"your_seat,omitempty"`
	Players  []PlayerInfo `json:"players,omitempty"`
	Config   *RoomConfig  `json:"config,omitempty"`

	// player_joined / player_left / player_ready / bot_added / bot_removed
	Seat     *int   `json:"seat,omitempty"`
	Nickname string `json:"nickname,omitempty"`

	// bot_added / bot_difficulty_changed
	IsBot      *bool         `json:"is_bot,omitempty"`
	Difficulty BotDifficulty `json:"difficulty,omitempty"`

	// game_started
	YourHand       []TileCode `json:"your_hand,omitempty"`
	DealerSeat     *int       `json:"dealer_seat,omitempty"`
	LaiziIndicator TileCode   `json:"laizi_indicator,omitempty"`
	LaiziTile      TileCode   `json:"laizi_tile,omitempty"`
	WallRemaining  *int       `json:"wall_remaining,omitempty"`

	// your_turn
	DrawnTile      TileCode   `json:"drawn_tile,omitempty"`
	TimeLimit      *int       `json:"time_limit,omitempty"`
	CanGang        []TileCode `json:"can_gang,omitempty"`
	CanHu          *bool      `json:"can_hu,omitempty"`
	HuScorePreview *int       `json:"hu_score_preview,omitempty"`
	WaitingTiles   []TileCode `json:"waiting_tiles,omitempty"`

	// tile_discarded / reaction_prompt
	Tile     TileCode `json:"tile,omitempty"`
	FromSeat *int     `json:"from_seat,omitempty"`

	// reaction_prompt
	AvailableActions []string       `json:"available_actions,omitempty"`
	ChiOptions       [][]TileCode   `json:"chi_options,omitempty"`

	// action_resolved
	Action        string     `json:"action,omitempty"`
	TilesRevealed []TileCode `json:"tiles_revealed,omitempty"`
	NextTurnSeat  *int       `json:"next_turn_seat,omitempty"`

	// gang_result
	GangType string `json:"gang_type,omitempty"`

	// round_end
	Result           string            `json:"result,omitempty"`
	WinnerSeat       *int              `json:"winner_seat,omitempty"`
	WinningHand      []TileCode        `json:"winning_hand,omitempty"`
	WinningTile      TileCode          `json:"winning_tile,omitempty"`
	WinType          string            `json:"win_type,omitempty"`
	Scoring          *ScoringBreakdown `json:"scoring,omitempty"`
	ScoreDeltas      map[string]int    `json:"score_deltas,omitempty"`
	TotalScores      map[string]int    `json:"total_scores,omitempty"`

	// game_state
	OpenMelds        map[string][]MeldInfo    `json:"open_melds,omitempty"`
	Discards         map[string][]TileCode    `json:"discards,omitempty"`
	TileCounts       map[string]int           `json:"tile_counts,omitempty"`
	CurrentTurnSeat  *int                     `json:"current_turn_seat,omitempty"`
	TurnTimeRemaining *int                    `json:"turn_time_remaining,omitempty"`

	// player_disconnected
	TimeoutSeconds *int `json:"timeout_seconds,omitempty"`

	// error
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"message,omitempty"`
}

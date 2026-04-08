// AUTO-GENERATED from schema/*.schema.json — do not edit manually
// Run: scripts/generate-types.sh

// ============ Tile Types ============

export type Suit = 'm' | 's' | 'p'

export type WindCode = 'we' | 'ws' | 'ww' | 'wn'
export type DragonCode = 'dz' | 'df' | 'db'
export type HonorCode = WindCode | DragonCode

export type SuitedTileCode =
  | '1m' | '2m' | '3m' | '4m' | '5m' | '6m' | '7m' | '8m' | '9m'
  | '1s' | '2s' | '3s' | '4s' | '5s' | '6s' | '7s' | '8s' | '9s'
  | '1p' | '2p' | '3p' | '4p' | '5p' | '6p' | '7p' | '8p' | '9p'

export type TileCode = SuitedTileCode | HonorCode

export type Valid258Pair = '2m' | '5m' | '8m' | '2s' | '5s' | '8s' | '2p' | '5p' | '8p'

// ============ Room Types ============

export type OpenCallMode = 'koukou' | 'kaikou'

export interface RoomConfig {
  score_cap: 200 | 500 | 1000 | 0
  open_call_mode: OpenCallMode
  turn_timer: 10 | 15 | 20 | 30
  reaction_timer: 5 | 8 | 10 | 15
  num_rounds: 4 | 8 | 16
}

export interface PlayerInfo {
  seat: number
  nickname: string
  ready: boolean
  connected: boolean
}

export interface MeldInfo {
  type: 'chi' | 'pong' | 'open_gang' | 'closed_gang' | 'add_gang'
  tiles: TileCode[]
}

export interface ScoringMultiplier {
  reason: string
  value: number
}

export interface ScoringBreakdown {
  base_points: number
  multipliers: ScoringMultiplier[]
  total_per_loser: number
  capped: boolean
}

// ============ Client → Server Messages ============

export type ClientMessage =
  | { type: 'join_room'; code: string; nickname: string; session_token: string }
  | { type: 'leave_room' }
  | { type: 'player_ready' }
  | { type: 'start_game' }
  | { type: 'configure_room'; config: RoomConfig }
  | { type: 'discard'; tile: TileCode }
  | { type: 'chi'; tiles: [TileCode, TileCode] }
  | { type: 'pong' }
  | { type: 'gang'; gang_type: 'open' | 'closed' | 'add'; tile: TileCode }
  | { type: 'hu' }
  | { type: 'pass' }

// ============ Server → Client Messages ============

export type ServerMessage =
  | RoomJoinedMsg
  | PlayerJoinedMsg
  | PlayerLeftMsg
  | PlayerReadyServerMsg
  | ConfigUpdatedMsg
  | GameStartedMsg
  | YourTurnMsg
  | TileDiscardedMsg
  | ReactionPromptMsg
  | ActionResolvedMsg
  | GangResultMsg
  | RoundEndMsg
  | GameStateMsg
  | PlayerDisconnectedMsg
  | PlayerReconnectedMsg
  | ErrorMsg

export interface RoomJoinedMsg {
  type: 'room_joined'
  room_id: string
  code: string
  your_seat: number
  players: PlayerInfo[]
  config: RoomConfig
}

export interface PlayerJoinedMsg {
  type: 'player_joined'
  seat: number
  nickname: string
}

export interface PlayerLeftMsg {
  type: 'player_left'
  seat: number
}

export interface PlayerReadyServerMsg {
  type: 'player_ready'
  seat: number
}

export interface ConfigUpdatedMsg {
  type: 'config_updated'
  config: RoomConfig
}

export interface GameStartedMsg {
  type: 'game_started'
  your_hand: TileCode[]
  dealer_seat: number
  laizi_indicator: TileCode
  laizi_tile: TileCode
  wall_remaining: number
}

export interface YourTurnMsg {
  type: 'your_turn'
  drawn_tile: TileCode
  time_limit: number
  wall_remaining: number
  can_gang: TileCode[]
  can_hu: boolean
}

export interface TileDiscardedMsg {
  type: 'tile_discarded'
  seat: number
  tile: TileCode
  wall_remaining: number
}

export type ReactionAction = 'chi' | 'pong' | 'gang' | 'hu' | 'pass'

export interface ReactionPromptMsg {
  type: 'reaction_prompt'
  tile: TileCode
  from_seat: number
  available_actions: ReactionAction[]
  chi_options?: [TileCode, TileCode][]
  time_limit: number
}

export interface ActionResolvedMsg {
  type: 'action_resolved'
  seat: number
  action: string
  tiles_revealed: TileCode[]
  next_turn_seat: number
}

export interface GangResultMsg {
  type: 'gang_result'
  seat: number
  gang_type: 'open' | 'closed' | 'add'
  tile?: TileCode
}

export interface RoundEndMsg {
  type: 'round_end'
  result: 'hu' | 'draw'
  winner_seat: number | null
  winning_hand?: TileCode[]
  winning_tile?: TileCode
  win_type?: 'self_draw' | 'discard' | 'rob_kong'
  scoring?: ScoringBreakdown
  score_deltas: Record<string, number>
  total_scores: Record<string, number>
}

export interface GameStateMsg {
  type: 'game_state'
  your_seat: number
  your_hand: TileCode[]
  open_melds: Record<string, MeldInfo[]>
  discards: Record<string, TileCode[]>
  tile_counts: Record<string, number>
  current_turn_seat: number
  laizi_indicator: TileCode
  laizi_tile: TileCode
  wall_remaining: number
  total_scores: Record<string, number>
  dealer_seat: number
  turn_time_remaining?: number
}

export interface PlayerDisconnectedMsg {
  type: 'player_disconnected'
  seat: number
  timeout_seconds: number
}

export interface PlayerReconnectedMsg {
  type: 'player_reconnected'
  seat: number
}

export interface ErrorMsg {
  type: 'error'
  code: string
  message: string
}

# Wuhan Mahjong (武汉麻将) — Game Design Reference

> Authoritative reference for implementing the game. All decisions are final unless explicitly revised.

---

## Table of Contents

1. [Game Rules](#1-game-rules)
2. [Tile Catalog](#2-tile-catalog)
3. [Game State Machine](#3-game-state-machine)
4. [Scoring System](#4-scoring-system)
5. [WebSocket Protocol](#5-websocket-protocol)
6. [Room Lifecycle](#6-room-lifecycle)
7. [Reconnection Flow](#7-reconnection-flow)
8. [Event Sourcing](#8-event-sourcing)
9. [Architecture Decisions](#9-architecture-decisions)

---

## 1. Game Rules

### 1.1 Overview

Wuhan Mahjong (武汉麻将), also known as 红中赖子杠 or 开口翻, is a 4-player regional mahjong variant from Wuhan, Hubei province. It uses the standard 136-tile set with a unique wild card (赖子) mechanic and multiplicative scoring.

### 1.2 Tile Set

Standard 136 tiles:
- **万 (Characters/Man)**: 1-9, four copies each = 36 tiles
- **条 (Bamboo/Sou)**: 1-9, four copies each = 36 tiles
- **筒 (Dots/Pin)**: 1-9, four copies each = 36 tiles
- **风 (Winds)**: 东南西北, four copies each = 16 tiles
- **箭 (Dragons)**: 红中/发财/白板, four copies each = 12 tiles

**Total: 136 tiles**

### 1.3 Dealing

1. Shuffle all 136 tiles into the wall
2. Dealer (庄家) is determined (first game: random, subsequent: winner or rotation)
3. Each player draws 13 tiles; dealer draws 14th
4. **Laizi determination**: Flip the next tile from the wall after the deal. The tile that follows it in sequence becomes the wild card (赖子). This indicator tile is set aside face-up for all to see.

### 1.4 赖子 (Laizi / Wild Card)

**Determination sequence:**

For numbered tiles (万/条/筒): Next value in same suit, wrapping 9→1
- Indicator shows 5万 → 6万 is laizi
- Indicator shows 9条 → 1条 is laizi

For honor tiles, the sequence is: 东→南→西→北→(skip 红中)→发→白→东
- Indicator shows 北 → 发财 is laizi (红中 is skipped)
- Indicator shows 白板 → 东 is laizi

**Usage:**
- A laizi tile can substitute for ANY tile to complete a winning hand
- **软胡 (Soft win)**: Using laizi as a substitute = 1x multiplier
- **硬胡 (Hard win)**: Using laizi as its natural tile value = 2x multiplier
- Forming a kong with all 4 laizi tiles = 4x multiplier

### 1.5 Winning Conditions

A standard winning hand consists of **14 tiles**: 4 sets + 1 pair.

**Sets can be:**
- **顺子 (Sequence/Chi)**: Three consecutive tiles of the same suit (e.g., 3m 4m 5m)
- **刻子 (Triplet/Pong)**: Three identical tiles (e.g., 7p 7p 7p)
- **杠子 (Kong/Gang)**: Four identical tiles

**258将 Rule (mandatory):**
The pair (将) in a winning hand MUST be a 2, 5, or 8 of any suit (万/条/筒). Pairs of honor tiles or other numbers are NOT valid for winning.

### 1.6 Player Actions

**On your turn (after drawing a tile):**
- **打牌 (Discard)**: Discard one tile from your hand

**Reacting to another player's discard:**
- **吃 (Chi)**: Claim the discard to complete a sequence (only from the player to your left)
- **碰 (Pong)**: Claim the discard to complete a triplet (from any player)
- **杠 (Gang/Kong)**: Claim the discard to complete a kong (from any player)
- **胡 (Hu/Win)**: Claim the discard to complete a winning hand
- **过 (Pass)**: Do nothing

**On your turn (special actions):**
- **暗杠 (Concealed Kong)**: Declare a kong from 4 identical tiles in your hand
- **补杠 (Add Kong)**: Add a drawn tile to an existing open triplet to make a kong
- **自摸 (Self-draw win)**: Win by drawing the completing tile yourself

**Reaction priority (when multiple players want to react to the same discard):**
1. 胡 (Hu) — highest priority
2. 杠 (Gang)
3. 碰 (Pong)
4. 吃 (Chi) — lowest priority (only from left player)

### 1.7 Special Rules

**抢杠胡 (Robbing the Kong):**
When a player declares 补杠 (add kong), other players may declare 胡 if that tile completes their hand. This takes priority over the kong.

**杠上开花 (Win off Kong Draw):**
After declaring any kong, the player draws a replacement tile from the back of the wall. If this tile completes a winning hand, it counts as a special self-draw win with bonus scoring.

**海底捞月 (Last Tile Win):**
Winning on the very last drawable tile from the wall.

**Draw (流局):**
If the wall is exhausted and no one has won, the round is a draw. No payments.

---

## 2. Tile Catalog

### 2.1 String Encoding

| Suit | Tiles | Codes |
|------|-------|-------|
| 万 (Characters) | 一万 through 九万 | `1m`, `2m`, `3m`, `4m`, `5m`, `6m`, `7m`, `8m`, `9m` |
| 条 (Bamboo) | 一条 through 九条 | `1s`, `2s`, `3s`, `4s`, `5s`, `6s`, `7s`, `8s`, `9s` |
| 筒 (Dots) | 一筒 through 九筒 | `1p`, `2p`, `3p`, `4p`, `5p`, `6p`, `7p`, `8p`, `9p` |
| 风 (Winds) | 东南西北 | `we`, `ws`, `ww`, `wn` |
| 箭 (Dragons) | 红中, 发财, 白板 | `dz`, `df`, `db` |

### 2.2 Instance IDs

Each physical tile has a unique instance ID for server-side tracking:
- Format: `{code}_{copy}` where copy is 0-3
- Example: `5m_0`, `5m_1`, `5m_2`, `5m_3`
- Clients only see the tile code (e.g., `5m`), never the instance ID

### 2.3 Laizi Sequence Lookup Table

| Indicator | Laizi (Wild Card) |
|-----------|-------------------|
| `1m` | `2m` |
| `2m` | `3m` |
| ... | ... |
| `9m` | `1m` |
| `1s` | `2s` |
| ... | ... |
| `9s` | `1s` |
| `1p` | `2p` |
| ... | ... |
| `9p` | `1p` |
| `we` (东) | `ws` (南) |
| `ws` (南) | `ww` (西) |
| `ww` (西) | `wn` (北) |
| `wn` (北) | `df` (发) — **skip 红中** |
| `dz` (红中) | `df` (发) |
| `df` (发) | `db` (白) |
| `db` (白) | `we` (东) |

---

## 3. Game State Machine

```
[WAITING]
    │ (4 players ready, host starts)
    ▼
[DEALING]
    │ (tiles dealt, laizi determined)
    ▼
[PLAYER_TURN] ◄──────────────────────┐
    │ (player draws tile)             │
    │                                 │
    ├──► Player discards ──► [AWAITING_REACTIONS]
    │                              │
    │                              ├── No reactions ──► next [PLAYER_TURN]
    │                              ├── Chi/Pong ──► claiming player's [PLAYER_TURN]
    │                              ├── Gang ──► claiming player draws, [PLAYER_TURN]
    │                              └── Hu ──► [ROUND_END]
    │
    ├──► Concealed Kong ──► draw replacement ──► [PLAYER_TURN]
    ├──► Add Kong ──► [AWAITING_ROB_KONG]
    │                      ├── No rob ──► draw replacement ──► [PLAYER_TURN]
    │                      └── Rob hu ──► [ROUND_END]
    ├──► Self-draw Hu ──► [ROUND_END]
    └──► Wall exhausted ──► [ROUND_END] (draw)

[ROUND_END]
    │
    ├──► Score calculation & display
    ├──► More rounds? ──► [DEALING]
    └──► Game over ──► [FINISHED]
```

### 3.1 State Details

**WAITING**: Room created, players joining. Host can configure rules.

**DEALING**: Server shuffles wall, deals 13 tiles to each player (14 to dealer), flips laizi indicator. Server sends each player only their own hand.

**PLAYER_TURN**: Active player has drawn a tile (or it's the dealer's first turn with 14 tiles). They must discard, declare a kong, or declare a win. Timer starts (default 15s).

**AWAITING_REACTIONS**: A tile was discarded. Server sends `reaction_prompt` to eligible players with their valid actions. Timer starts (default 8s). Server waits for all responses or timeout, then resolves by priority.

**AWAITING_ROB_KONG**: A player declared 补杠. Server briefly checks if any player can 抢杠胡. If yes, hu takes priority. If no, kong proceeds.

**ROUND_END**: Scoring calculated, results broadcast. Brief display period, then next round or game end.

**FINISHED**: All rounds complete. Final scores displayed.

### 3.2 Turn Order

Counter-clockwise (standard Chinese mahjong): Seat 0 → Seat 1 → Seat 2 → Seat 3 → Seat 0...

Chi is only allowed from the player to your left (the player whose turn was immediately before yours).

---

## 4. Scoring System

### 4.1 Base Points

Standard win (basic hu): **1 point** base

大胡 (Big win) patterns add to the base:
- Each 大胡 pattern = +10 base points
- Multiple 大胡 patterns stack: 2 patterns = 20 base, 3 = 30 base, etc.

Common 大胡 patterns include:
- **碰碰胡**: All triplets, no sequences
- **清一色**: All tiles of a single suit
- **全求人**: Win entirely from others' discards (fully open hand)
- **七对**: Seven pairs
- **杠上开花**: Win off a kong replacement draw
- **抢杠胡**: Win by robbing a kong
- **海底捞月**: Win on the last tile

### 4.2 Multipliers

Multipliers are applied multiplicatively to the base points:

| Condition | Multiplier |
|-----------|-----------|
| Each open call (口口翻 mode) | ×2 per call |
| Any open call (开口翻 mode) | ×2 once |
| 软胡 (laizi used as substitute) | ×1 |
| 硬胡 (laizi used as natural value) | ×2 |
| 自摸 (self-draw win) | ×2 |
| 杠上开花 | ×2 |
| 抢杠胡 | ×2 |
| 4-laizi kong | ×4 |

### 4.3 Formula

```
Total = Base Points × (Multiplier₁ × Multiplier₂ × ... × Multiplierₙ)
```

Capped at configurable maximum (default: **500 points**).

### 4.4 Payment

- **Self-draw win (自摸)**: Each of the 3 losers pays the winner the full calculated amount
- **Discard win (点炮)**: The player who discarded the winning tile pays the winner. Other players pay nothing.
- **抢杠胡**: The player whose kong was robbed pays as if they discarded

### 4.5 Scoring Example

Player wins with:
- 2 open calls (碰 twice) in 口口翻 mode
- Used laizi as substitute (软胡)
- Won by self-draw (自摸)
- Base hand = standard hu (1 point)

```
Total = 1 × 2 × 2 × 1 × 2 = 8 points
         └─┘   └─┘   └─┘ └─┘
       call 1  call 2  软胡  自摸

Each loser pays: 8 points
Winner receives: 24 points total
```

---

## 5. WebSocket Protocol

### 5.1 Message Format

All messages are JSON with a `type` field for routing:

```json
{"type": "message_type", ...payload}
```

### 5.2 Client → Server Messages

#### `join_room`
```json
{
  "type": "join_room",
  "code": "ABC123",
  "nickname": "Player1",
  "session_token": "tok_xxxxx"
}
```

#### `leave_room`
```json
{"type": "leave_room"}
```

#### `player_ready`
```json
{"type": "player_ready"}
```

#### `start_game` (host only)
```json
{"type": "start_game"}
```

#### `configure_room` (host only)
```json
{
  "type": "configure_room",
  "config": {
    "score_cap": 500,
    "open_call_mode": "koukou",
    "turn_timer": 15,
    "reaction_timer": 8
  }
}
```

#### `discard`
```json
{
  "type": "discard",
  "tile": "5m"
}
```

#### `chi`
```json
{
  "type": "chi",
  "tiles": ["3m", "4m"]
}
```
Note: The claimed discard tile is implicit. The two tiles listed are from the player's hand.

#### `pong`
```json
{"type": "pong"}
```

#### `gang`
```json
{
  "type": "gang",
  "gang_type": "open|closed|add",
  "tile": "7p"
}
```
- `open`: Claiming a discard to form a kong
- `closed`: Declaring a concealed kong from 4 tiles in hand
- `add`: Adding to an existing open triplet

#### `hu`
```json
{"type": "hu"}
```

#### `pass`
```json
{"type": "pass"}
```

### 5.3 Server → Client Messages

#### `room_joined`
```json
{
  "type": "room_joined",
  "room_id": "uuid",
  "code": "ABC123",
  "your_seat": 2,
  "players": [
    {"seat": 0, "nickname": "Host", "ready": true},
    {"seat": 1, "nickname": "Alice", "ready": false},
    {"seat": 2, "nickname": "You", "ready": false}
  ],
  "config": {
    "score_cap": 500,
    "open_call_mode": "koukou",
    "turn_timer": 15,
    "reaction_timer": 8
  }
}
```

#### `player_joined`
```json
{
  "type": "player_joined",
  "seat": 3,
  "nickname": "Bob"
}
```

#### `player_left`
```json
{
  "type": "player_left",
  "seat": 3
}
```

#### `player_ready`
```json
{
  "type": "player_ready",
  "seat": 1
}
```

#### `config_updated`
```json
{
  "type": "config_updated",
  "config": { ... }
}
```

#### `game_started`
```json
{
  "type": "game_started",
  "your_hand": ["1m", "1m", "3m", "5p", "6p", "7p", "2s", "3s", "4s", "we", "we", "dz", "5m"],
  "dealer_seat": 0,
  "laizi_indicator": "4s",
  "laizi_tile": "5s",
  "wall_remaining": 83
}
```
Note: Dealer receives 14 tiles (their turn starts immediately).

#### `your_turn`
```json
{
  "type": "your_turn",
  "drawn_tile": "8p",
  "time_limit": 15,
  "wall_remaining": 72,
  "can_gang": ["7p"],
  "can_hu": true
}
```

#### `tile_discarded`
```json
{
  "type": "tile_discarded",
  "seat": 1,
  "tile": "3m",
  "wall_remaining": 72
}
```

#### `reaction_prompt`
```json
{
  "type": "reaction_prompt",
  "tile": "3m",
  "from_seat": 1,
  "available_actions": ["pong", "hu", "pass"],
  "chi_options": [],
  "time_limit": 8
}
```
Note: `chi_options` lists possible chi combinations if chi is available, e.g., `[["1m","2m"], ["4m","5m"]]`

#### `action_resolved`
```json
{
  "type": "action_resolved",
  "seat": 2,
  "action": "pong",
  "tiles_revealed": ["3m", "3m", "3m"],
  "next_turn_seat": 2
}
```

#### `gang_result`
```json
{
  "type": "gang_result",
  "seat": 0,
  "gang_type": "closed",
  "tile": "7p"
}
```
Note: For concealed kong, other players only see that a kong was declared, not the tile (unless rules dictate otherwise).

#### `round_end`
```json
{
  "type": "round_end",
  "result": "hu",
  "winner_seat": 2,
  "winning_hand": ["1m", "2m", "3m", "5p", "5p", "5p", "7s", "8s", "9s", "we", "we", "we", "2m", "2m"],
  "winning_tile": "2m",
  "win_type": "self_draw",
  "scoring": {
    "base_points": 1,
    "multipliers": [
      {"reason": "open_call", "value": 2},
      {"reason": "open_call", "value": 2},
      {"reason": "self_draw", "value": 2}
    ],
    "total_per_loser": 8,
    "capped": false
  },
  "score_deltas": {
    "0": -8,
    "1": -8,
    "2": 24,
    "3": -8
  },
  "total_scores": {
    "0": 492,
    "1": 492,
    "2": 524,
    "3": 492
  }
}
```

#### `round_end` (draw)
```json
{
  "type": "round_end",
  "result": "draw",
  "winner_seat": null,
  "score_deltas": {"0": 0, "1": 0, "2": 0, "3": 0},
  "total_scores": {"0": 500, "1": 500, "2": 500, "3": 500}
}
```

#### `game_state` (for reconnection)
```json
{
  "type": "game_state",
  "your_seat": 2,
  "your_hand": ["1m", "3m", "5p", ...],
  "open_melds": {
    "0": [{"type": "pong", "tiles": ["3m", "3m", "3m"]}],
    "1": [],
    "2": [{"type": "chi", "tiles": ["4s", "5s", "6s"]}],
    "3": []
  },
  "discards": {
    "0": ["9p", "1s", ...],
    "1": ["we", "7m", ...],
    "2": ["db", ...],
    "3": ["4p", ...]
  },
  "tile_counts": {"0": 10, "1": 13, "2": 11, "3": 13},
  "current_turn_seat": 1,
  "laizi_indicator": "4s",
  "laizi_tile": "5s",
  "wall_remaining": 45,
  "total_scores": {"0": 500, "1": 500, "2": 500, "3": 500},
  "dealer_seat": 0,
  "turn_time_remaining": 8
}
```

#### `player_disconnected`
```json
{
  "type": "player_disconnected",
  "seat": 3,
  "timeout_seconds": 120
}
```

#### `player_reconnected`
```json
{
  "type": "player_reconnected",
  "seat": 3
}
```

#### `error`
```json
{
  "type": "error",
  "code": "invalid_action",
  "message": "You cannot chi this tile"
}
```

---

## 6. Room Lifecycle

```
[CREATE] Host creates room → gets room code
    │
[LOBBY] Players join via code → see player list
    │    Host configures rules
    │    Players mark ready
    │
[START] Host starts game (requires 4 ready players)
    │
[PLAYING] Game rounds proceed
    │    Multiple rounds per game
    │
[FINISHED] Final scores displayed
    │    Option to play again (returns to LOBBY)
    │
[EXPIRED] Room cleaned up after inactivity (e.g., 1 hour)
```

### 6.1 Room Code

- 6 characters, uppercase alphanumeric (excluding confusable characters: 0/O, 1/I/L)
- Character set: `ABCDEFGHJKMNPQRSTUVWXYZ23456789`
- Example: `H7KM3P`

### 6.2 Room Configuration (set by host)

| Setting | Default | Options |
|---------|---------|---------|
| Score cap | 500 | 200, 500, 1000, none |
| Open call mode | 口口翻 | 口口翻, 开口翻 |
| Turn timer | 15s | 10s, 15s, 20s, 30s |
| Reaction timer | 8s | 5s, 8s, 10s, 15s |
| Number of rounds | 8 | 4, 8, 16 |

---

## 7. Reconnection Flow

```
1. WebSocket drops
    │
2. Server detects disconnect (ping/pong timeout ~10s)
    │
3. Server marks player as disconnected
    │
4. Server broadcasts `player_disconnected` to room (120s countdown)
    │
5. Game continues — disconnected player auto-passes reactions,
   auto-discards drawn tile on their turn
    │
6a. Player reconnects within 120s:
    │   - Client connects with same session_token
    │   - Server validates token, matches to user/room
    │   - Server sends `game_state` (from latest snapshot + replaying recent events)
    │   - Server broadcasts `player_reconnected`
    │   - Player resumes normal play
    │
6b. Player does NOT reconnect within 120s:
    │   - Player continues as auto-pilot for rest of the round
    │   - At round end, seat opens for potential reconnection
    │   - If still disconnected after game ends, removed from room
```

### 7.1 Client-Side Reconnection

- Client stores `session_token` in localStorage
- On page load, client checks for existing token and room state
- WebSocket composable implements auto-reconnect with exponential backoff:
  - Attempt 1: immediate
  - Attempt 2: 1s delay
  - Attempt 3: 2s delay
  - Attempt 4: 4s delay
  - Max: 30s between attempts
- Show "Reconnecting..." overlay during attempts

---

## 8. Event Sourcing

### 8.1 Event Types

| Event Type | Payload | Description |
|-----------|---------|-------------|
| `game_start` | `{wall, hands, dealer_seat, laizi_indicator, laizi_tile}` | New round begins |
| `tile_drawn` | `{seat, tile}` | Player draws from wall |
| `tile_discarded` | `{seat, tile}` | Player discards |
| `chi_declared` | `{seat, tiles, claimed_tile}` | Player claims chi |
| `pong_declared` | `{seat, claimed_tile}` | Player claims pong |
| `open_gang_declared` | `{seat, claimed_tile}` | Player claims open kong |
| `closed_gang_declared` | `{seat, tile}` | Player declares concealed kong |
| `add_gang_declared` | `{seat, tile}` | Player adds to existing pong |
| `gang_replacement_drawn` | `{seat, tile}` | Player draws after kong |
| `hu_declared` | `{seat, tile, win_type, scoring}` | Player wins |
| `pass` | `{seat}` | Player passes on reaction |
| `auto_pass` | `{seat}` | Server auto-passes for timeout |
| `auto_discard` | `{seat, tile}` | Server auto-discards for timeout |
| `round_end` | `{result, scores}` | Round concludes |
| `player_disconnected` | `{seat}` | Player connection lost |
| `player_reconnected` | `{seat}` | Player reconnected |

### 8.2 Snapshots

- Taken every 20 events or at round boundaries
- Contains full serialized game state
- Used for fast reconnection (load snapshot, replay events after it)
- Stored in `game_snapshots` table

### 8.3 Replay Reconstruction

To replay a game:
1. Load `game_start` event to get initial state
2. Step through each event in sequence order
3. Client can play forward/backward, pause, adjust speed

---

## 9. Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Mahjong variant | 武汉麻将 (红中赖子杠) | User's preference |
| Platform | Browser, mobile + desktop | Maximum reach |
| Multiplayer | Invite-only via room code (MVP) | Simplest viable approach |
| Auth | Guest + nickname (MVP) | Zero friction; auth microservice integration later |
| Game logic | Server-authoritative (Go) | Hidden information makes client-auth unsafe |
| Client role | Pure renderer + action sender | No game logic in frontend |
| Wire format | JSON over WebSocket | Tiny payloads, easy debugging, no build step |
| Tile encoding | String codes (`1m`, `5p`, `we`) | Human-readable, compact |
| Persistence | Event sourcing + snapshots (PostgreSQL) | Enables replays, debugging, fast reconnect |
| Type sharing | JSON Schema → codegen Go + TS | Single source of truth across languages |
| Frontend framework | Vue 3 + Pinia + SCSS | User's preference |
| Tile rendering | SVG | Scalable, animatable, styleable |
| UI components | All custom (no component library) | Game UI is 95% custom anyway |
| Animations | Vue Motion (motion-v), ≤300ms | Functional, not flashy |
| Local dev | Vite (frontend), Docker (backend + PG) | Fast frontend iteration |
| Deployment | Railway, separate projects | Simple PaaS deployment |
| Repo structure | Monorepo | Atomic changes, simpler management |

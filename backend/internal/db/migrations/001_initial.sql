-- 001_initial.sql
-- Initial schema for Wuhan Mahjong

-- Guest users (every player gets a row, even guests)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname VARCHAR(32) NOT NULL,
    session_token VARCHAR(128) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Game rooms
CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(8) UNIQUE NOT NULL,
    host_user_id UUID REFERENCES users(id),
    config JSONB NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'waiting',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Players in a room (seat assignment)
CREATE TABLE room_players (
    room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    seat INTEGER NOT NULL CHECK (seat BETWEEN 0 AND 3),
    connected BOOLEAN DEFAULT true,
    PRIMARY KEY (room_id, user_id),
    UNIQUE (room_id, seat)
);

-- Event sourcing: every game action
CREATE TABLE game_events (
    id BIGSERIAL PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    seq INTEGER NOT NULL,
    event_type VARCHAR(32) NOT NULL,
    player_seat INTEGER,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (room_id, seq)
);

CREATE INDEX idx_game_events_room_seq ON game_events (room_id, seq);

-- Periodic snapshots for fast reconnection
CREATE TABLE game_snapshots (
    id BIGSERIAL PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    seq INTEGER NOT NULL,
    state JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_game_snapshots_room ON game_snapshots (room_id, seq DESC);

-- Game results (per round)
CREATE TABLE game_results (
    id BIGSERIAL PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    round INTEGER NOT NULL,
    winner_seat INTEGER,
    scores JSONB NOT NULL,
    scoring_breakdown JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_game_results_room ON game_results (room_id, round);

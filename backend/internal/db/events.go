package db

import (
	"context"
	"encoding/json"
	"fmt"
)

// SaveGameEvent persists a game event to the database.
func (db *DB) SaveGameEvent(ctx context.Context, roomID string, seq int, eventType string, playerSeat int, payload map[string]interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	_, err = db.Pool.Exec(ctx,
		"INSERT INTO game_events (room_id, seq, event_type, player_seat, payload) VALUES ($1, $2, $3, $4, $5)",
		roomID, seq, eventType, playerSeat, payloadJSON,
	)
	return err
}

// SaveGameSnapshot persists a game state snapshot.
func (db *DB) SaveGameSnapshot(ctx context.Context, roomID string, seq int, state interface{}) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	_, err = db.Pool.Exec(ctx,
		"INSERT INTO game_snapshots (room_id, seq, state) VALUES ($1, $2, $3)",
		roomID, seq, stateJSON,
	)
	return err
}

// GetGameEvents retrieves all events for a room, optionally starting from a sequence number.
func (db *DB) GetGameEvents(ctx context.Context, roomID string, fromSeq int) ([]map[string]interface{}, error) {
	rows, err := db.Pool.Query(ctx,
		"SELECT seq, event_type, player_seat, payload FROM game_events WHERE room_id = $1 AND seq > $2 ORDER BY seq",
		roomID, fromSeq,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var seq, playerSeat int
		var eventType string
		var payloadJSON []byte

		if err := rows.Scan(&seq, &eventType, &playerSeat, &payloadJSON); err != nil {
			return nil, err
		}

		var payload map[string]interface{}
		json.Unmarshal(payloadJSON, &payload)

		events = append(events, map[string]interface{}{
			"seq":         seq,
			"event_type":  eventType,
			"player_seat": playerSeat,
			"payload":     payload,
		})
	}

	return events, rows.Err()
}

// GetLatestSnapshot retrieves the latest snapshot for a room.
func (db *DB) GetLatestSnapshot(ctx context.Context, roomID string) (int, []byte, error) {
	var seq int
	var state []byte
	err := db.Pool.QueryRow(ctx,
		"SELECT seq, state FROM game_snapshots WHERE room_id = $1 ORDER BY seq DESC LIMIT 1",
		roomID,
	).Scan(&seq, &state)
	if err != nil {
		return 0, nil, err
	}
	return seq, state, nil
}

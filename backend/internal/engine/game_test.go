package engine

import (
	"testing"

	"github.com/mahjong/backend/internal/models"
)

func TestGameDeal(t *testing.T) {
	config := models.RoomConfig{
		ScoreCap:      500,
		OpenCallMode:  models.OpenCallModeKouKou,
		TurnTimer:     15,
		ReactionTimer: 8,
		NumRounds:     8,
	}

	game := NewGame(config, 0, [4]int{500, 500, 500, 500}, 1)
	if err := game.Deal(); err != nil {
		t.Fatalf("Deal failed: %v", err)
	}

	// Dealer (seat 0) should have 14 tiles, others 13
	if len(game.Players[0].Hand) != 14 {
		t.Errorf("dealer should have 14 tiles, got %d", len(game.Players[0].Hand))
	}
	for seat := 1; seat < 4; seat++ {
		if len(game.Players[seat].Hand) != 13 {
			t.Errorf("player %d should have 13 tiles, got %d", seat, len(game.Players[seat].Hand))
		}
	}

	// Laizi should be set
	if game.LaiziIndicator == "" {
		t.Error("laizi indicator should be set")
	}
	if game.LaiziTile == "" {
		t.Error("laizi tile should be set")
	}

	// Phase should be player_turn (dealer's turn)
	if game.Phase != PhasePlayerTurn {
		t.Errorf("expected PhasePlayerTurn, got %s", game.Phase)
	}
	if game.CurrentTurn != 0 {
		t.Errorf("expected dealer (seat 0) to have first turn, got seat %d", game.CurrentTurn)
	}

	// Events should have been recorded
	if len(game.Events) == 0 {
		t.Error("expected at least one event to be recorded")
	}
}

func TestGameDiscard(t *testing.T) {
	config := models.RoomConfig{
		ScoreCap:      500,
		OpenCallMode:  models.OpenCallModeKouKou,
		TurnTimer:     15,
		ReactionTimer: 8,
		NumRounds:     8,
	}

	game := NewGame(config, 0, [4]int{500, 500, 500, 500}, 1)
	if err := game.Deal(); err != nil {
		t.Fatalf("Deal failed: %v", err)
	}

	// Dealer discards their first tile
	tile := game.Players[0].Hand[0]
	handSizeBefore := len(game.Players[0].Hand)

	if err := game.Discard(0, tile); err != nil {
		t.Fatalf("Discard failed: %v", err)
	}

	// Hand should be one smaller
	if len(game.Players[0].Hand) != handSizeBefore-1 {
		t.Errorf("hand size should be %d after discard, got %d", handSizeBefore-1, len(game.Players[0].Hand))
	}

	// Tile should be in discards
	if len(game.Players[0].Discards) != 1 || game.Players[0].Discards[0] != tile {
		t.Error("discarded tile should be in discard pile")
	}
}

func TestGameDiscardWrongTurn(t *testing.T) {
	config := models.RoomConfig{
		ScoreCap:      500,
		OpenCallMode:  models.OpenCallModeKouKou,
		TurnTimer:     15,
		ReactionTimer: 8,
		NumRounds:     8,
	}

	game := NewGame(config, 0, [4]int{500, 500, 500, 500}, 1)
	game.Deal()

	// Player 1 tries to discard when it's dealer's (seat 0) turn
	if err := game.Discard(1, game.Players[1].Hand[0]); err == nil {
		t.Error("should not be able to discard out of turn")
	}
}

func TestGamePlayerView(t *testing.T) {
	config := models.RoomConfig{
		ScoreCap:      500,
		OpenCallMode:  models.OpenCallModeKouKou,
		TurnTimer:     15,
		ReactionTimer: 8,
		NumRounds:     8,
	}

	game := NewGame(config, 0, [4]int{500, 500, 500, 500}, 1)
	game.Deal()

	view := game.GetPlayerView(0)
	if view.YourSeat != 0 {
		t.Errorf("expected seat 0, got %d", view.YourSeat)
	}
	if len(view.YourHand) != 14 {
		t.Errorf("expected 14 tiles in view, got %d", len(view.YourHand))
	}
	if view.LaiziTile == "" {
		t.Error("laizi tile should be visible in player view")
	}
	if view.WallRemaining <= 0 {
		t.Error("wall should have remaining tiles")
	}
}

func TestGameGetAvailableActions(t *testing.T) {
	config := models.RoomConfig{
		ScoreCap:      500,
		OpenCallMode:  models.OpenCallModeKouKou,
		TurnTimer:     15,
		ReactionTimer: 8,
		NumRounds:     8,
	}

	game := NewGame(config, 0, [4]int{500, 500, 500, 500}, 1)
	game.Deal()

	// Dealer should be able to discard
	actions := game.GetAvailableActions(0)
	hasDiscard := false
	for _, a := range actions {
		if a == "discard" {
			hasDiscard = true
		}
	}
	if !hasDiscard {
		t.Error("dealer should have 'discard' as available action")
	}

	// Non-current player should have no actions
	actions = game.GetAvailableActions(1)
	if len(actions) != 0 {
		t.Errorf("player 1 should have no actions, got %v", actions)
	}
}

func TestGameFullRoundNoReactions(t *testing.T) {
	config := models.RoomConfig{
		ScoreCap:      500,
		OpenCallMode:  models.OpenCallModeKouKou,
		TurnTimer:     15,
		ReactionTimer: 8,
		NumRounds:     8,
	}

	game := NewGame(config, 0, [4]int{500, 500, 500, 500}, 1)
	game.Deal()

	// Play through several turns: each player discards their first tile
	for i := 0; i < 20; i++ {
		seat := game.CurrentTurn
		if game.Phase == PhaseRoundEnd {
			break
		}
		if game.Phase != PhasePlayerTurn {
			// If awaiting reactions, auto-pass all
			if game.Phase == PhaseAwaitingReaction {
				for s := range game.ReactionsNeeded {
					game.AutoPass(s)
				}
				continue
			}
			break
		}

		hand := game.Players[seat].Hand
		if len(hand) == 0 {
			break
		}

		if err := game.Discard(seat, hand[0]); err != nil {
			t.Fatalf("turn %d seat %d discard failed: %v", i, seat, err)
		}
	}

	// Game should still be running or ended normally
	if game.Phase != PhasePlayerTurn && game.Phase != PhaseAwaitingReaction && game.Phase != PhaseRoundEnd {
		t.Errorf("unexpected phase after playing turns: %s", game.Phase)
	}
}

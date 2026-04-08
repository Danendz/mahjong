package engine

import (
	"testing"

	"github.com/mahjong/backend/internal/models"
)

func TestNewWall(t *testing.T) {
	wall := NewWall()
	if wall.Remaining() != 136 {
		t.Errorf("expected 136 tiles, got %d", wall.Remaining())
	}
}

func TestWallShuffle(t *testing.T) {
	w1 := NewWall()
	w2 := NewWall()
	w2.Shuffle()

	// After shuffle, walls should differ (extremely unlikely to be same)
	same := true
	for i := range w1.tiles {
		if w1.tiles[i] != w2.tiles[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("shuffled wall should differ from unshuffled")
	}
}

func TestWallDraw(t *testing.T) {
	wall := NewWall()
	tile, ok := wall.Draw()
	if !ok {
		t.Fatal("draw failed")
	}
	if tile.Code == "" {
		t.Error("drawn tile has empty code")
	}
	if wall.Remaining() != 135 {
		t.Errorf("expected 135 remaining, got %d", wall.Remaining())
	}
}

func TestWallDrawBack(t *testing.T) {
	wall := NewWall()
	tile, ok := wall.DrawBack()
	if !ok {
		t.Fatal("draw back failed")
	}
	if tile.Code == "" {
		t.Error("drawn tile has empty code")
	}
	if wall.Remaining() != 135 {
		t.Errorf("expected 135 remaining, got %d", wall.Remaining())
	}
}

func TestWallExhaustion(t *testing.T) {
	wall := NewWall()
	for range 136 {
		_, ok := wall.Draw()
		if !ok {
			t.Fatal("draw failed before exhaustion")
		}
	}
	_, ok := wall.Draw()
	if ok {
		t.Error("draw should fail on exhausted wall")
	}
	if wall.Remaining() != 0 {
		t.Errorf("expected 0 remaining, got %d", wall.Remaining())
	}
}

func TestDealHands(t *testing.T) {
	wall := NewWall()
	wall.Shuffle()
	hands, ok := DealHands(wall, 0)
	if !ok {
		t.Fatal("deal failed")
	}

	// Dealer should have 14, others 13
	if len(hands[0]) != 14 {
		t.Errorf("dealer should have 14 tiles, got %d", len(hands[0]))
	}
	for seat := 1; seat < 4; seat++ {
		if len(hands[seat]) != 13 {
			t.Errorf("player %d should have 13 tiles, got %d", seat, len(hands[seat]))
		}
	}

	// Total dealt = 14 + 13*3 = 53
	expectedRemaining := 136 - 53
	if wall.Remaining() != expectedRemaining {
		t.Errorf("expected %d remaining in wall, got %d", expectedRemaining, wall.Remaining())
	}
}

func TestTileSuit(t *testing.T) {
	tests := []struct {
		code models.TileCode
		want byte
	}{
		{"5m", 'm'}, {"1s", 's'}, {"9p", 'p'},
		{"we", 'e'}, {"dz", 'z'},
	}
	for _, tt := range tests {
		got := TileSuit(tt.code)
		if got != tt.want {
			t.Errorf("TileSuit(%s) = %c, want %c", tt.code, got, tt.want)
		}
	}
}

func TestTileValue(t *testing.T) {
	tests := []struct {
		code models.TileCode
		want int
	}{
		{"1m", 1}, {"5s", 5}, {"9p", 9},
		{"we", 0}, {"dz", 0},
	}
	for _, tt := range tests {
		got := TileValue(tt.code)
		if got != tt.want {
			t.Errorf("TileValue(%s) = %d, want %d", tt.code, got, tt.want)
		}
	}
}

func TestIsSuited(t *testing.T) {
	if !IsSuited("5m") {
		t.Error("5m should be suited")
	}
	if IsSuited("we") {
		t.Error("we should not be suited")
	}
	if IsSuited("dz") {
		t.Error("dz should not be suited")
	}
}

func TestIsValid258Pair(t *testing.T) {
	valid := []models.TileCode{"2m", "5m", "8m", "2s", "5s", "8s", "2p", "5p", "8p"}
	invalid := []models.TileCode{"1m", "3m", "4m", "6m", "7m", "9m", "we", "dz"}

	for _, tc := range valid {
		if !IsValid258Pair(tc) {
			t.Errorf("%s should be valid 258 pair", tc)
		}
	}
	for _, tc := range invalid {
		if IsValid258Pair(tc) {
			t.Errorf("%s should not be valid 258 pair", tc)
		}
	}
}

func TestNextInSequence(t *testing.T) {
	tests := []struct {
		code models.TileCode
		want models.TileCode
	}{
		{"1m", "2m"}, {"8s", "9s"}, {"9p", "1p"},
		{"we", ""}, // honor tiles have no sequence
	}
	for _, tt := range tests {
		got := NextInSequence(tt.code)
		if got != tt.want {
			t.Errorf("NextInSequence(%s) = %s, want %s", tt.code, got, tt.want)
		}
	}
}

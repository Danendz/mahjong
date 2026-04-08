package engine

import (
	"testing"

	"github.com/mahjong/backend/internal/models"
)

func TestIsWinningHand_BasicWin(t *testing.T) {
	// 4 sets + pair with 258 pair
	// 1m2m3m 4s5s6s 7p8p9p wewewewe 2m2m (pair)
	// Simplified: 3 sequences + 1 triplet + 258 pair
	hand := []models.TileCode{
		"1m", "2m", "3m", // sequence
		"4s", "5s", "6s", // sequence
		"7p", "8p", "9p", // sequence
		"we", "we", "we", // triplet
		"2m", "2m", // pair (258 valid)
	}
	result := IsWinningHand(hand, "dz") // dz as laizi (not in hand)
	if !result.IsWin {
		t.Error("expected winning hand")
	}
	if result.UsesLaizi {
		t.Error("should not use laizi")
	}
}

func TestIsWinningHand_Invalid258Pair(t *testing.T) {
	// Same structure but pair is 1m1m (not 258)
	hand := []models.TileCode{
		"1m", "2m", "3m",
		"4s", "5s", "6s",
		"7p", "8p", "9p",
		"we", "we", "we",
		"1m", "1m", // invalid pair (1 is not 2/5/8)
	}
	result := IsWinningHand(hand, "dz")
	if result.IsWin {
		t.Error("pair of 1m should be invalid (not 258)")
	}
}

func TestIsWinningHand_WithLaizi(t *testing.T) {
	// Use laizi as substitute for one tile
	hand := []models.TileCode{
		"1m", "2m", "3m",
		"4s", "5s", "6s",
		"7p", "8p", "dz", // dz is laizi, substitutes for 9p
		"we", "we", "we",
		"2s", "2s", // valid 258 pair
	}
	result := IsWinningHand(hand, "dz")
	if !result.IsWin {
		t.Error("expected winning hand with laizi substitute")
	}
	if !result.UsesLaizi {
		t.Error("should flag laizi usage")
	}
}

func TestIsWinningHand_HardHu(t *testing.T) {
	// Laizi used as its natural tile value
	// If laizi is dz, and hand has dz in a valid triplet
	hand := []models.TileCode{
		"1m", "2m", "3m",
		"4s", "5s", "6s",
		"7p", "8p", "9p",
		"dz", "dz", "dz", // laizi used as natural 红中
		"5m", "5m", // valid 258 pair
	}
	result := IsWinningHand(hand, "dz")
	if !result.IsWin {
		t.Error("expected winning hand (hard hu)")
	}
	if !result.IsHardHu {
		t.Error("should be hard hu when laizi used as natural value")
	}
}

func TestIsWinningHand_NotWin(t *testing.T) {
	hand := []models.TileCode{
		"1m", "2m", "4m", // not a valid sequence
		"4s", "5s", "6s",
		"7p", "8p", "9p",
		"we", "we", "we",
		"2m", "2m",
	}
	result := IsWinningHand(hand, "dz")
	if result.IsWin {
		t.Error("should not be a winning hand")
	}
}

func TestCanChi(t *testing.T) {
	hand := []models.TileCode{"3m", "4m", "7p", "8p", "1s"}

	// Can chi 5m with 3m,4m
	options := CanChi(hand, "5m")
	if len(options) != 1 {
		t.Fatalf("expected 1 chi option for 5m, got %d", len(options))
	}

	// Can chi 2m with 3m (2m+3m+4m would need 4m too but 2m,3m,4m — we have 3m,4m)
	options = CanChi(hand, "2m")
	if len(options) != 1 {
		t.Fatalf("expected 1 chi option for 2m, got %d", len(options))
	}

	// Can't chi honor tiles
	options = CanChi(hand, "we")
	if len(options) != 0 {
		t.Error("should not be able to chi honor tiles")
	}
}

func TestCanPong(t *testing.T) {
	hand := []models.TileCode{"5m", "5m", "3s", "7p"}
	if !CanPong(hand, "5m") {
		t.Error("should be able to pong 5m")
	}
	if CanPong(hand, "3s") {
		t.Error("should not be able to pong 3s with only 1 copy")
	}
}

func TestCanOpenGang(t *testing.T) {
	hand := []models.TileCode{"5m", "5m", "5m", "3s"}
	if !CanOpenGang(hand, "5m") {
		t.Error("should be able to open gang 5m")
	}
	if CanOpenGang(hand, "3s") {
		t.Error("should not be able to open gang 3s")
	}
}

func TestFindClosedGangs(t *testing.T) {
	hand := []models.TileCode{"5m", "5m", "5m", "5m", "3s", "3s"}
	gangs := FindClosedGangs(hand)
	if len(gangs) != 1 {
		t.Fatalf("expected 1 closed gang, got %d", len(gangs))
	}
	if gangs[0] != "5m" {
		t.Errorf("expected closed gang of 5m, got %s", gangs[0])
	}
}

func TestFindAddGangs(t *testing.T) {
	hand := []models.TileCode{"5m", "3s", "7p"}
	melds := []models.MeldInfo{
		{Type: models.MeldPong, Tiles: []models.TileCode{"5m", "5m", "5m"}},
		{Type: models.MeldChi, Tiles: []models.TileCode{"1s", "2s", "3s"}},
	}
	gangs := FindAddGangs(hand, melds)
	if len(gangs) != 1 {
		t.Fatalf("expected 1 add gang, got %d", len(gangs))
	}
	if gangs[0] != "5m" {
		t.Errorf("expected add gang of 5m, got %s", gangs[0])
	}
}

func TestCanWinWithTile(t *testing.T) {
	// 13-tile hand that's one tile away from winning
	hand := []models.TileCode{
		"1m", "2m", "3m",
		"4s", "5s", "6s",
		"7p", "8p", "9p",
		"we", "we", "we",
		"2m", // need another 2m for pair
	}
	result := CanWinWithTile(hand, "2m", "dz")
	if !result.IsWin {
		t.Error("should win with 2m completing the pair")
	}

	result = CanWinWithTile(hand, "3m", "dz")
	if result.IsWin {
		t.Error("should not win with 3m")
	}
}

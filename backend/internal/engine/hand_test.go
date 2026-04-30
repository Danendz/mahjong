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
	result := IsWinningHand(hand, nil, "dz") // dz as laizi (not in hand)
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
	result := IsWinningHand(hand, nil, "dz")
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
	result := IsWinningHand(hand, nil, "dz")
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
	result := IsWinningHand(hand, nil, "dz")
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
	result := IsWinningHand(hand, nil, "dz")
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
	result := CanWinWithTile(hand, "2m", nil, "dz")
	if !result.IsWin {
		t.Error("should win with 2m completing the pair")
	}

	result = CanWinWithTile(hand, "3m", nil, "dz")
	if result.IsWin {
		t.Error("should not win with 3m")
	}
}

// --- Meld-aware win detection ---

func TestIsWinningHand_OneMeldPlusClosed(t *testing.T) {
	// Player has 1 pong meld (we), closed hand forms 3 sets + 258 pair = 11 tiles
	// 1m2m3m + 4s5s6s + 7p8p9p + 2m2m (pair)
	closed := []models.TileCode{
		"1m", "2m", "3m",
		"4s", "5s", "6s",
		"7p", "8p", "9p",
		"2m", "2m",
	}
	melds := []models.MeldInfo{
		{Type: models.MeldPong, Tiles: []models.TileCode{"we", "we", "we"}},
	}
	result := IsWinningHand(closed, melds, "dz")
	if !result.IsWin {
		t.Error("expected winning hand with 1 pong meld + 3 closed sets + 258 pair")
	}
}

func TestIsWinningHand_TwoMeldsPlusClosed(t *testing.T) {
	// Player has 2 melds (1 pong, 1 chi), closed hand forms 2 sets + 258 pair = 8 tiles
	closed := []models.TileCode{
		"4s", "5s", "6s",
		"7p", "8p", "9p",
		"5m", "5m",
	}
	melds := []models.MeldInfo{
		{Type: models.MeldPong, Tiles: []models.TileCode{"we", "we", "we"}},
		{Type: models.MeldChi, Tiles: []models.TileCode{"1m", "2m", "3m"}},
	}
	result := IsWinningHand(closed, melds, "dz")
	if !result.IsWin {
		t.Error("expected winning hand with 2 melds + 2 closed sets + 258 pair")
	}
}

func TestIsWinningHand_OneMeldNoWin(t *testing.T) {
	// Player has 1 chi meld, closed hand has 11 tiles but doesn't form 3 sets + 258 pair
	closed := []models.TileCode{
		"1m", "2m", "4m", // gap — not a valid sequence
		"4s", "5s", "6s",
		"7p", "8p", "9p",
		"5m", "5m",
	}
	melds := []models.MeldInfo{
		{Type: models.MeldChi, Tiles: []models.TileCode{"1s", "2s", "3s"}},
	}
	result := IsWinningHand(closed, melds, "dz")
	if result.IsWin {
		t.Error("should not win — closed hand has gap in sequence")
	}
}

func TestIsWinningHand_OneMeldWithLaiziSubstitute(t *testing.T) {
	// 1 chi meld + closed hand needing laizi substitute for one tile
	closed := []models.TileCode{
		"1m", "2m", "3m",
		"4s", "5s", "6s",
		"7p", "8p", "dz", // dz substitutes for 9p
		"5m", "5m",
	}
	melds := []models.MeldInfo{
		{Type: models.MeldPong, Tiles: []models.TileCode{"we", "we", "we"}},
	}
	result := IsWinningHand(closed, melds, "dz")
	if !result.IsWin {
		t.Error("expected winning hand with laizi substitute and 1 meld")
	}
	if !result.UsesLaizi {
		t.Error("should flag laizi usage")
	}
}

func TestIsWinningHand_FourMeldsClosedPair(t *testing.T) {
	// Edge: 4 melds + just a 258 pair in closed hand
	closed := []models.TileCode{"5m", "5m"}
	melds := []models.MeldInfo{
		{Type: models.MeldPong, Tiles: []models.TileCode{"we", "we", "we"}},
		{Type: models.MeldChi, Tiles: []models.TileCode{"1m", "2m", "3m"}},
		{Type: models.MeldChi, Tiles: []models.TileCode{"4s", "5s", "6s"}},
		{Type: models.MeldChi, Tiles: []models.TileCode{"7p", "8p", "9p"}},
	}
	result := IsWinningHand(closed, melds, "dz")
	if !result.IsWin {
		t.Error("expected winning hand with 4 melds + 258 pair")
	}
}

func TestIsWinningHand_ClosedGangCountsAsSet(t *testing.T) {
	// Closed gang (4 tiles in meld) counts as one set; replacement tile is in closed hand
	// Closed: 1m 2m 3m, 4s 5s 6s, 7p 8p 9p, 5m 5m → 3 sets + pair (after replacement draw)
	closed := []models.TileCode{
		"1m", "2m", "3m",
		"4s", "5s", "6s",
		"7p", "8p", "9p",
		"5m", "5m",
	}
	melds := []models.MeldInfo{
		{Type: models.MeldClosedGang, Tiles: []models.TileCode{"we", "we", "we", "we"}},
	}
	result := IsWinningHand(closed, melds, "dz")
	if !result.IsWin {
		t.Error("expected winning hand with closed gang counting as one set")
	}
}

func TestCanWinWithTile_WithMelds(t *testing.T) {
	// Player has 1 pong meld + 10-tile closed; drawing 5m completes a 258 pair
	closed := []models.TileCode{
		"1m", "2m", "3m",
		"4s", "5s", "6s",
		"7p", "8p", "9p",
		"5m",
	}
	melds := []models.MeldInfo{
		{Type: models.MeldPong, Tiles: []models.TileCode{"we", "we", "we"}},
	}
	if !CanWinWithTile(closed, "5m", melds, "dz").IsWin {
		t.Error("should win with 5m completing pair (1 meld + 3 closed sets + pair)")
	}
	if CanWinWithTile(closed, "1s", melds, "dz").IsWin {
		t.Error("should not win with 1s — no valid arrangement")
	}
}

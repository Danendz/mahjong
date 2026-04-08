package engine

import (
	"testing"

	"github.com/mahjong/backend/internal/models"
)

func TestCalculateScore_BasicSelfDraw(t *testing.T) {
	ctx := WinContext{
		Hand:         []models.TileCode{"1m", "2m", "3m", "4s", "5s", "6s", "7p", "8p", "9p", "we", "we", "we", "2m", "2m"},
		Melds:        nil,
		WinType:      WinSelfDraw,
		LaiziTile:    "dz",
		Analysis:     HandAnalysis{IsWin: true},
		OpenCallMode: models.OpenCallModeKouKou,
		ScoreCap:     500,
	}

	result := CalculateScore(ctx)
	// base=1, self_draw=x2 → total=2
	if result.TotalPerLoser != 2 {
		t.Errorf("expected 2 per loser, got %d", result.TotalPerLoser)
	}
	// Self draw: each of 3 losers pays
	if result.WinnerGains != 6 {
		t.Errorf("expected winner gains 6, got %d", result.WinnerGains)
	}
}

func TestCalculateScore_DiscardWin(t *testing.T) {
	ctx := WinContext{
		Hand:         []models.TileCode{"1m", "2m", "3m", "4s", "5s", "6s", "7p", "8p", "9p", "we", "we", "we", "2m", "2m"},
		Melds:        nil,
		WinType:      WinDiscard,
		LaiziTile:    "dz",
		Analysis:     HandAnalysis{IsWin: true},
		OpenCallMode: models.OpenCallModeKouKou,
		ScoreCap:     500,
	}

	result := CalculateScore(ctx)
	// base=1, no multipliers for discard win → total=1
	if result.TotalPerLoser != 1 {
		t.Errorf("expected 1 per loser, got %d", result.TotalPerLoser)
	}
	// Discard: only one person pays
	if result.WinnerGains != 1 {
		t.Errorf("expected winner gains 1, got %d", result.WinnerGains)
	}
}

func TestCalculateScore_KouKouFan(t *testing.T) {
	ctx := WinContext{
		Hand: []models.TileCode{"7p", "8p", "9p", "2s", "2s"},
		Melds: []models.MeldInfo{
			{Type: models.MeldPong, Tiles: []models.TileCode{"5m", "5m", "5m"}},
			{Type: models.MeldChi, Tiles: []models.TileCode{"1s", "2s", "3s"}},
			{Type: models.MeldPong, Tiles: []models.TileCode{"we", "we", "we"}},
		},
		WinType:      WinSelfDraw,
		LaiziTile:    "dz",
		Analysis:     HandAnalysis{IsWin: true},
		OpenCallMode: models.OpenCallModeKouKou,
		ScoreCap:     500,
	}

	result := CalculateScore(ctx)
	// base=1, 3 open calls (口口翻: x2 x2 x2 = x8), self_draw=x2 → 1*8*2=16
	if result.TotalPerLoser != 16 {
		t.Errorf("expected 16 per loser (口口翻), got %d", result.TotalPerLoser)
	}
}

func TestCalculateScore_KaiKouFan(t *testing.T) {
	ctx := WinContext{
		Hand: []models.TileCode{"7p", "8p", "9p", "2s", "2s"},
		Melds: []models.MeldInfo{
			{Type: models.MeldPong, Tiles: []models.TileCode{"5m", "5m", "5m"}},
			{Type: models.MeldChi, Tiles: []models.TileCode{"1s", "2s", "3s"}},
			{Type: models.MeldPong, Tiles: []models.TileCode{"we", "we", "we"}},
		},
		WinType:      WinSelfDraw,
		LaiziTile:    "dz",
		Analysis:     HandAnalysis{IsWin: true},
		OpenCallMode: models.OpenCallModeKaiKou,
		ScoreCap:     500,
	}

	result := CalculateScore(ctx)
	// base=1, 开口翻: x2 once, self_draw=x2 → 1*2*2=4
	if result.TotalPerLoser != 4 {
		t.Errorf("expected 4 per loser (开口翻), got %d", result.TotalPerLoser)
	}
}

func TestCalculateScore_HardHu(t *testing.T) {
	ctx := WinContext{
		Hand:         []models.TileCode{"1m", "2m", "3m", "4s", "5s", "6s", "7p", "8p", "9p", "dz", "dz", "dz", "5m", "5m"},
		WinType:      WinSelfDraw,
		LaiziTile:    "dz",
		Analysis:     HandAnalysis{IsWin: true, IsHardHu: true},
		OpenCallMode: models.OpenCallModeKouKou,
		ScoreCap:     500,
	}

	result := CalculateScore(ctx)
	// base=1, hard_hu=x2, self_draw=x2 → 1*2*2=4
	if result.TotalPerLoser != 4 {
		t.Errorf("expected 4 per loser (hard hu + self draw), got %d", result.TotalPerLoser)
	}
}

func TestCalculateScore_Cap(t *testing.T) {
	ctx := WinContext{
		Hand: []models.TileCode{"5m", "5m"},
		Melds: []models.MeldInfo{
			{Type: models.MeldPong, Tiles: []models.TileCode{"1m", "1m", "1m"}},
			{Type: models.MeldPong, Tiles: []models.TileCode{"2m", "2m", "2m"}},
			{Type: models.MeldPong, Tiles: []models.TileCode{"3m", "3m", "3m"}},
			{Type: models.MeldOpenGang, Tiles: []models.TileCode{"4m", "4m", "4m", "4m"}},
		},
		WinType:      WinSelfDraw,
		LaiziTile:    "dz",
		Analysis:     HandAnalysis{IsWin: true, IsHardHu: true},
		OpenCallMode: models.OpenCallModeKouKou,
		ScoreCap:     500,
		IsKongDraw:   true,
	}

	result := CalculateScore(ctx)
	if result.TotalPerLoser > 500 {
		t.Errorf("score should be capped at 500, got %d", result.TotalPerLoser)
	}
	if !result.Breakdown.Capped {
		t.Error("should be flagged as capped")
	}
}

func TestCalculatePayments_SelfDraw(t *testing.T) {
	payments := CalculatePayments(2, WinSelfDraw, -1, 10)
	if payments[2] != 30 {
		t.Errorf("winner should gain 30, got %d", payments[2])
	}
	for seat := range 4 {
		if seat != 2 && payments[seat] != -10 {
			t.Errorf("seat %d should pay -10, got %d", seat, payments[seat])
		}
	}
}

func TestCalculatePayments_DiscardWin(t *testing.T) {
	payments := CalculatePayments(2, WinDiscard, 0, 10)
	if payments[2] != 10 {
		t.Errorf("winner should gain 10, got %d", payments[2])
	}
	if payments[0] != -10 {
		t.Errorf("discarder should pay -10, got %d", payments[0])
	}
	if payments[1] != 0 || payments[3] != 0 {
		t.Error("non-involved players should pay 0")
	}
}

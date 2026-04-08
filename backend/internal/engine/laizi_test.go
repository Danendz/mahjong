package engine

import (
	"testing"

	"github.com/mahjong/backend/internal/models"
)

func TestDetermineLaizi(t *testing.T) {
	wall := NewWall()
	// Don't shuffle so the result is deterministic
	indicator, laizi, ok := DetermineLaizi(wall)
	if !ok {
		t.Fatal("DetermineLaizi failed")
	}

	expected, exists := models.LaiziSequence[indicator]
	if !exists {
		t.Fatalf("indicator %s not in LaiziSequence", indicator)
	}
	if laizi != expected {
		t.Errorf("for indicator %s: got laizi %s, want %s", indicator, laizi, expected)
	}
}

func TestDetermineLaiziHonorSkip(t *testing.T) {
	// Verify 北 → 发 (skipping 红中)
	laizi, exists := models.LaiziSequence[models.TileWN]
	if !exists {
		t.Fatal("WN not in LaiziSequence")
	}
	if laizi != models.TileDF {
		t.Errorf("北 indicator should give 发 as laizi, got %s", laizi)
	}

	// Verify 红中 → 发
	laizi, exists = models.LaiziSequence[models.TileDZ]
	if !exists {
		t.Fatal("DZ not in LaiziSequence")
	}
	if laizi != models.TileDF {
		t.Errorf("红中 indicator should give 发 as laizi, got %s", laizi)
	}

	// Verify 白板 → 东
	laizi, exists = models.LaiziSequence[models.TileDB]
	if !exists {
		t.Fatal("DB not in LaiziSequence")
	}
	if laizi != models.TileWE {
		t.Errorf("白板 indicator should give 东 as laizi, got %s", laizi)
	}
}

func TestCountLaizi(t *testing.T) {
	hand := []models.TileCode{"1m", "5m", "5m", "3s", "5m"}
	count := CountLaizi(hand, "5m")
	if count != 3 {
		t.Errorf("expected 3 laizi, got %d", count)
	}
}

func TestSeparateLaizi(t *testing.T) {
	hand := []models.TileCode{"1m", "5m", "3s", "5m", "7p"}
	regular, count := SeparateLaizi(hand, "5m")
	if count != 2 {
		t.Errorf("expected 2 laizi, got %d", count)
	}
	if len(regular) != 3 {
		t.Errorf("expected 3 regular tiles, got %d", len(regular))
	}
	for _, r := range regular {
		if r == "5m" {
			t.Error("regular tiles should not contain laizi")
		}
	}
}

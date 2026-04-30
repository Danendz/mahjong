package engine

import (
	"fmt"
	"sort"

	"github.com/mahjong/backend/internal/models"
)

// HandAnalysis holds the result of analyzing a hand for winning.
type HandAnalysis struct {
	IsWin          bool
	IsHardHu       bool // laizi used as its natural tile (硬胡)
	UsesLaizi      bool // laizi used as substitute (软胡)
	LaiziAsNatural int  // how many laizi used as their own tile code
}

// IsWinningHand checks if the player's closed hand together with their existing
// melds forms a valid winning hand: 4 sets total + 1 pair, where the pair must be 258.
// Each meld in `melds` (chi/pong/open_gang/closed_gang/add_gang) counts as one
// already-formed set; the closed hand needs to provide the remaining sets and the pair.
// laiziTile is the current laizi code; laizi tiles in the closed hand can substitute for anything.
func IsWinningHand(closedHand []models.TileCode, melds []models.MeldInfo, laiziTile models.TileCode) HandAnalysis {
	requiredSets := 4 - len(melds)
	if requiredSets < 0 {
		// Defensive: more than 4 melds shouldn't happen in legal play
		return HandAnalysis{IsWin: false}
	}

	regular, laiziCount := SeparateLaizi(closedHand, laiziTile)

	// Try 硬胡 first: use laizi as their natural tile value (gives 2x multiplier)
	if laiziCount > 0 {
		hardHand := make([]models.TileCode, len(regular), len(regular)+laiziCount)
		copy(hardHand, regular)
		for range laiziCount {
			hardHand = append(hardHand, laiziTile)
		}
		if canFormWin(hardHand, 0, requiredSets) {
			return HandAnalysis{
				IsWin:          true,
				IsHardHu:       true,
				UsesLaizi:      false,
				LaiziAsNatural: laiziCount,
			}
		}
	}

	// Try 软胡: use laizi as wildcards (substitutes)
	if canFormWin(regular, laiziCount, requiredSets) {
		return HandAnalysis{
			IsWin:     true,
			UsesLaizi: laiziCount > 0,
			IsHardHu:  false,
		}
	}

	return HandAnalysis{IsWin: false}
}

// canFormWin checks if regular tiles + wildcardCount wildcards can form `requiredSets`
// sets + 1 pair with a valid 258 pair.
func canFormWin(regular []models.TileCode, wildcards int, requiredSets int) bool {
	freq := TileCodesToMap(regular)
	return tryFormSets(freq, wildcards, 0, false, requiredSets)
}

// tryFormSets recursively tries to form `requiredSets` sets + 1 pair.
// It iterates through tiles in sorted order, trying to form a pair first,
// then triplets and sequences.
func tryFormSets(freq map[models.TileCode]int, wildcards int, setsFormed int, pairUsed bool, requiredSets int) bool {
	// Check if we've formed enough sets
	totalTiles := 0
	for _, c := range freq {
		totalTiles += c
	}
	totalTiles += wildcards

	if totalTiles == 0 && setsFormed == requiredSets && pairUsed {
		return true
	}
	if setsFormed > requiredSets {
		return false
	}

	// Find the smallest tile that has count > 0
	smallest := findSmallestTile(freq)
	if smallest == "" && wildcards == 0 {
		return pairUsed && setsFormed == requiredSets
	}

	// If only wildcards remain, check if they can fill remaining needs
	if smallest == "" {
		needed := 0
		if !pairUsed {
			needed += 2
		}
		needed += (requiredSets - setsFormed) * 3
		return wildcards >= needed
	}

	// Try using this tile as part of a pair (if pair not yet used)
	if !pairUsed {
		if IsValid258Pair(smallest) {
			// Pair with 2 of this tile
			if freq[smallest] >= 2 {
				freq[smallest] -= 2
				if tryFormSets(freq, wildcards, setsFormed, true, requiredSets) {
					freq[smallest] += 2
					return true
				}
				freq[smallest] += 2
			}

			// Pair with 1 of this tile + 1 wildcard
			if freq[smallest] >= 1 && wildcards >= 1 {
				freq[smallest]--
				if tryFormSets(freq, wildcards-1, setsFormed, true, requiredSets) {
					freq[smallest]++
					return true
				}
				freq[smallest]++
			}
		}

		// Pair with 2 wildcards on a 258 tile (wildcard acts as a 258 tile)
		if wildcards >= 2 {
			// Try each valid 258 pair that isn't in freq
			for _, pair := range models.Valid258Pairs {
				if tryFormSets(freq, wildcards-2, setsFormed, true, requiredSets) {
					_ = pair
					return true
				}
				break // Only need to try once since wildcards are interchangeable
			}
		}
	}

	// Try using this tile in a triplet (刻子)
	if freq[smallest] >= 3 {
		freq[smallest] -= 3
		if tryFormSets(freq, wildcards, setsFormed+1, pairUsed, requiredSets) {
			freq[smallest] += 3
			return true
		}
		freq[smallest] += 3
	}

	// Triplet with wildcards
	if freq[smallest] >= 2 && wildcards >= 1 {
		freq[smallest] -= 2
		if tryFormSets(freq, wildcards-1, setsFormed+1, pairUsed, requiredSets) {
			freq[smallest] += 2
			return true
		}
		freq[smallest] += 2
	}
	if freq[smallest] >= 1 && wildcards >= 2 {
		freq[smallest]--
		if tryFormSets(freq, wildcards-2, setsFormed+1, pairUsed, requiredSets) {
			freq[smallest]++
			return true
		}
		freq[smallest]++
	}

	// Try using this tile in a sequence (顺子) — only for suited tiles
	if IsSuited(smallest) {
		next1 := NextInSequence(smallest)
		v := TileValue(smallest)

		// Only form ascending sequences (value ≤ 7 for first tile of sequence)
		if v <= 7 {
			next2 := NextInSequence(next1)

			// All 3 tiles present
			if freq[next1] > 0 && freq[next2] > 0 {
				freq[smallest]--
				freq[next1]--
				freq[next2]--
				if tryFormSets(freq, wildcards, setsFormed+1, pairUsed, requiredSets) {
					freq[smallest]++
					freq[next1]++
					freq[next2]++
					return true
				}
				freq[smallest]++
				freq[next1]++
				freq[next2]++
			}

			// 2 tiles + 1 wildcard
			if wildcards >= 1 {
				// Have smallest + next1, wildcard as next2
				if freq[next1] > 0 {
					freq[smallest]--
					freq[next1]--
					if tryFormSets(freq, wildcards-1, setsFormed+1, pairUsed, requiredSets) {
						freq[smallest]++
						freq[next1]++
						return true
					}
					freq[smallest]++
					freq[next1]++
				}
				// Have smallest + next2, wildcard as next1
				if freq[next2] > 0 {
					freq[smallest]--
					freq[next2]--
					if tryFormSets(freq, wildcards-1, setsFormed+1, pairUsed, requiredSets) {
						freq[smallest]++
						freq[next2]++
						return true
					}
					freq[smallest]++
					freq[next2]++
				}
			}

			// 1 tile + 2 wildcards
			if wildcards >= 2 {
				freq[smallest]--
				if tryFormSets(freq, wildcards-2, setsFormed+1, pairUsed, requiredSets) {
					freq[smallest]++
					return true
				}
				freq[smallest]++
			}
		}

		// Handle sequences starting at value 8: 8-9-wildcard (only if v == 8)
		if v == 8 && wildcards >= 1 && freq[next1] > 0 {
			// 8, 9, wildcard-as-7(invalid) — no, sequences must be consecutive
			// Actually 8-9 needs a 10 which doesn't exist, so this is already
			// handled: v <= 7 check above covers valid sequence starts.
			// For v == 8: could be part of 6-7-8 or 7-8-9, but those would be
			// handled when processing 6 or 7 as the smallest tile.
		}
	}

	// If we can't use this tile in any valid combination, the hand is invalid
	// (a tile that can't form part of any set means the hand can't win)
	// But we should try consuming it with wildcards forming a triplet above,
	// which we already did.
	return false
}

// findSmallestTile returns the smallest tile code with count > 0, sorted by a canonical order.
func findSmallestTile(freq map[models.TileCode]int) models.TileCode {
	var tiles []models.TileCode
	for code, count := range freq {
		if count > 0 {
			tiles = append(tiles, code)
		}
	}
	if len(tiles) == 0 {
		return ""
	}
	sort.Slice(tiles, func(i, j int) bool {
		return tileOrder(tiles[i]) < tileOrder(tiles[j])
	})
	return tiles[0]
}

// tileOrder returns a sortable integer for a tile code.
func tileOrder(code models.TileCode) int {
	for i, c := range models.AllTileCodes {
		if c == code {
			return i
		}
	}
	return 999
}

// CanChi checks if a player can chi (claim a sequence) with the given discard tile
// using tiles from their hand. Returns all valid chi combinations.
// Each combination is a pair of tiles from the hand that form a sequence with the discard.
func CanChi(hand []models.TileCode, discard models.TileCode) [][2]models.TileCode {
	if !IsSuited(discard) {
		return nil // Can't chi honor tiles
	}

	freq := TileCodesToMap(hand)
	var options [][2]models.TileCode
	v := TileValue(discard)
	suit := TileSuit(discard)

	makeTile := func(val int) models.TileCode {
		return models.TileCode(fmt.Sprintf("%d%c", val, suit))
	}

	// discard is the low tile: discard, discard+1, discard+2
	if v <= 7 {
		t1 := makeTile(v + 1)
		t2 := makeTile(v + 2)
		if freq[t1] > 0 && freq[t2] > 0 {
			options = append(options, [2]models.TileCode{t1, t2})
		}
	}

	// discard is the middle tile: discard-1, discard, discard+1
	if v >= 2 && v <= 8 {
		t1 := makeTile(v - 1)
		t2 := makeTile(v + 1)
		if freq[t1] > 0 && freq[t2] > 0 {
			options = append(options, [2]models.TileCode{t1, t2})
		}
	}

	// discard is the high tile: discard-2, discard-1, discard
	if v >= 3 {
		t1 := makeTile(v - 2)
		t2 := makeTile(v - 1)
		if freq[t1] > 0 && freq[t2] > 0 {
			options = append(options, [2]models.TileCode{t1, t2})
		}
	}

	return options
}

// CanPong checks if a player can pong (claim a triplet) with the given discard tile.
func CanPong(hand []models.TileCode, discard models.TileCode) bool {
	count := 0
	for _, t := range hand {
		if t == discard {
			count++
		}
	}
	return count >= 2
}

// CanOpenGang checks if a player can declare an open kong with the discard tile.
func CanOpenGang(hand []models.TileCode, discard models.TileCode) bool {
	count := 0
	for _, t := range hand {
		if t == discard {
			count++
		}
	}
	return count >= 3
}

// FindClosedGangs returns tile codes for which the player has all 4 copies in hand.
func FindClosedGangs(hand []models.TileCode) []models.TileCode {
	freq := TileCodesToMap(hand)
	var gangs []models.TileCode
	for code, count := range freq {
		if count == 4 {
			gangs = append(gangs, code)
		}
	}
	return gangs
}

// FindAddGangs returns tile codes where the player can add to an existing open pong.
func FindAddGangs(hand []models.TileCode, melds []models.MeldInfo) []models.TileCode {
	var gangs []models.TileCode
	handSet := TileCodesToMap(hand)
	for _, meld := range melds {
		if meld.Type == models.MeldPong {
			tile := meld.Tiles[0]
			if handSet[tile] > 0 {
				gangs = append(gangs, tile)
			}
		}
	}
	return gangs
}

// CanWinWithTile checks if adding `tile` to the closed hand (combined with existing melds)
// results in a winning hand.
func CanWinWithTile(closedHand []models.TileCode, tile models.TileCode, melds []models.MeldInfo, laiziTile models.TileCode) HandAnalysis {
	fullHand := make([]models.TileCode, len(closedHand), len(closedHand)+1)
	copy(fullHand, closedHand)
	fullHand = append(fullHand, tile)
	return IsWinningHand(fullHand, melds, laiziTile)
}

// FindWinningDiscards returns all tiles that, if drawn, would complete the hand,
// accounting for any existing melds.
func FindWinningDiscards(closedHand []models.TileCode, melds []models.MeldInfo, laiziTile models.TileCode) []models.TileCode {
	var winning []models.TileCode
	for _, code := range models.AllTileCodes {
		if CanWinWithTile(closedHand, code, melds, laiziTile).IsWin {
			winning = append(winning, code)
		}
	}
	return winning
}

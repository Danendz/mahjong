package engine

import "github.com/mahjong/backend/internal/models"

// WinType describes how the player won.
type WinType string

const (
	WinSelfDraw WinType = "self_draw" // 自摸
	WinDiscard  WinType = "discard"   // 点炮
	WinRobKong  WinType = "rob_kong"  // 抢杠胡
)

// WinContext contains all information needed to calculate the score for a win.
type WinContext struct {
	Hand          []models.TileCode
	Melds         []models.MeldInfo
	WinningTile   models.TileCode
	WinType       WinType
	LaiziTile     models.TileCode
	Analysis      HandAnalysis
	OpenCallMode  models.OpenCallMode
	ScoreCap      int
	IsKongDraw    bool // 杠上开花
	IsRobKong     bool // 抢杠胡
	IsLastTile    bool // 海底捞月
}

// ScoreResult contains the calculated score.
type ScoreResult struct {
	Breakdown     models.ScoringBreakdown
	TotalPerLoser int
	WinnerGains   int // Total the winner receives
}

// CalculateScore computes the score for a winning hand.
func CalculateScore(ctx WinContext) ScoreResult {
	base := 1 // Standard hu base
	var multipliers []models.ScoringMultiplier

	// Check for 大胡 patterns
	bigWins := checkBigWins(ctx)
	if len(bigWins) > 0 {
		base = len(bigWins) * 10
		for _, name := range bigWins {
			multipliers = append(multipliers, models.ScoringMultiplier{
				Reason: name,
				Value:  1, // Big wins add to base, not multiply
			})
		}
	}

	// Open call multipliers
	openCallCount := countOpenCalls(ctx.Melds)
	if openCallCount > 0 {
		switch ctx.OpenCallMode {
		case models.OpenCallModeKouKou:
			// 口口翻: each open call doubles
			for i := range openCallCount {
				_ = i
				multipliers = append(multipliers, models.ScoringMultiplier{
					Reason: "open_call",
					Value:  2,
				})
			}
		case models.OpenCallModeKaiKou:
			// 开口翻: any open call = one double
			multipliers = append(multipliers, models.ScoringMultiplier{
				Reason: "open_call",
				Value:  2,
			})
		}
	}

	// 软胡/硬胡 multiplier
	if ctx.Analysis.IsHardHu {
		multipliers = append(multipliers, models.ScoringMultiplier{
			Reason: "hard_hu",
			Value:  2,
		})
	}
	// 软胡 = 1x, no multiplier added

	// Self-draw multiplier
	if ctx.WinType == WinSelfDraw {
		multipliers = append(multipliers, models.ScoringMultiplier{
			Reason: "self_draw",
			Value:  2,
		})
	}

	// 杠上开花
	if ctx.IsKongDraw {
		multipliers = append(multipliers, models.ScoringMultiplier{
			Reason: "kong_draw",
			Value:  2,
		})
	}

	// 抢杠胡
	if ctx.IsRobKong {
		multipliers = append(multipliers, models.ScoringMultiplier{
			Reason: "rob_kong",
			Value:  2,
		})
	}

	// 海底捞月
	if ctx.IsLastTile {
		multipliers = append(multipliers, models.ScoringMultiplier{
			Reason: "last_tile",
			Value:  2,
		})
	}

	// Calculate total
	total := base
	for _, m := range multipliers {
		if m.Value > 1 {
			total *= m.Value
		}
	}

	// Apply cap
	capped := false
	if ctx.ScoreCap > 0 && total > ctx.ScoreCap {
		total = ctx.ScoreCap
		capped = true
	}

	winnerGains := total * 3
	if ctx.WinType == WinDiscard || ctx.WinType == WinRobKong {
		winnerGains = total // Only one person pays for discard win
	}

	return ScoreResult{
		Breakdown: models.ScoringBreakdown{
			BasePoints:    base,
			Multipliers:   multipliers,
			TotalPerLoser: total,
			Capped:        capped,
		},
		TotalPerLoser: total,
		WinnerGains:   winnerGains,
	}
}

// countOpenCalls counts chi and pong melds (open calls that trigger 翻).
func countOpenCalls(melds []models.MeldInfo) int {
	count := 0
	for _, m := range melds {
		switch m.Type {
		case models.MeldChi, models.MeldPong, models.MeldOpenGang:
			count++
		}
	}
	return count
}

// checkBigWins checks for 大胡 patterns and returns their names.
func checkBigWins(ctx WinContext) []string {
	var wins []string

	allTiles := gatherAllTiles(ctx.Hand, ctx.Melds)

	if isPengPengHu(ctx.Hand, ctx.Melds) {
		wins = append(wins, "peng_peng_hu") // 碰碰胡
	}
	if isQingYiSe(allTiles) {
		wins = append(wins, "qing_yi_se") // 清一色
	}
	if ctx.IsKongDraw {
		wins = append(wins, "gang_shang_kai_hua") // 杠上开花
	}
	if ctx.IsRobKong {
		wins = append(wins, "qiang_gang_hu") // 抢杠胡
	}
	if ctx.IsLastTile {
		wins = append(wins, "hai_di_lao_yue") // 海底捞月
	}
	if isSevenPairs(ctx.Hand, ctx.Melds) {
		wins = append(wins, "qi_dui") // 七对
	}

	return wins
}

// gatherAllTiles combines hand tiles and meld tiles into one slice.
func gatherAllTiles(hand []models.TileCode, melds []models.MeldInfo) []models.TileCode {
	all := make([]models.TileCode, len(hand))
	copy(all, hand)
	for _, m := range melds {
		all = append(all, m.Tiles...)
	}
	return all
}

// isPengPengHu checks if the hand is all triplets/kongs + pair (no sequences).
func isPengPengHu(hand []models.TileCode, melds []models.MeldInfo) bool {
	// All melds must be pong or gang
	for _, m := range melds {
		if m.Type == models.MeldChi {
			return false
		}
	}
	// Remaining hand tiles should form only triplets + 1 pair
	freq := TileCodesToMap(hand)
	pairs := 0
	for _, count := range freq {
		if count == 1 {
			return false // Can't form triplet or pair with 1 tile
		}
		if count == 2 {
			pairs++
		}
		// count == 3 is a triplet, count == 4 is a closed kong
	}
	return pairs == 1
}

// isQingYiSe checks if all tiles are the same suit.
func isQingYiSe(tiles []models.TileCode) bool {
	if len(tiles) == 0 {
		return false
	}
	var suit byte
	for _, t := range tiles {
		if IsHonor(t) {
			return false
		}
		s := TileSuit(t)
		if suit == 0 {
			suit = s
		} else if s != suit {
			return false
		}
	}
	return true
}

// isSevenPairs checks if the hand is 7 pairs (no melds).
func isSevenPairs(hand []models.TileCode, melds []models.MeldInfo) bool {
	if len(melds) > 0 || len(hand) != 14 {
		return false
	}
	freq := TileCodesToMap(hand)
	for _, count := range freq {
		if count != 2 && count != 4 {
			return false
		}
	}
	return len(freq) == 7 || (len(freq) < 7 && countTotalTilesFromFreq(freq) == 14)
}

func countTotalTilesFromFreq(freq map[models.TileCode]int) int {
	total := 0
	for _, c := range freq {
		total += c
	}
	return total
}

// CalculatePayments computes score deltas for each seat.
// Returns map[seat]delta where positive = gains, negative = losses.
func CalculatePayments(winnerSeat int, winType WinType, discardSeat int, totalPerLoser int) map[int]int {
	deltas := map[int]int{0: 0, 1: 0, 2: 0, 3: 0}

	switch winType {
	case WinSelfDraw:
		// Each loser pays totalPerLoser
		for seat := range 4 {
			if seat == winnerSeat {
				deltas[seat] = totalPerLoser * 3
			} else {
				deltas[seat] = -totalPerLoser
			}
		}
	case WinDiscard, WinRobKong:
		// Only the discarder / kong declarer pays
		deltas[discardSeat] = -totalPerLoser
		deltas[winnerSeat] = totalPerLoser
	}

	return deltas
}

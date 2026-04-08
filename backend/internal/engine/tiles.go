package engine

import (
	"fmt"
	"math/rand/v2"

	"github.com/mahjong/backend/internal/models"
)

// Tile represents a single physical tile with a unique instance ID.
type Tile struct {
	Code     models.TileCode
	Instance int // 0-3, which copy of this tile
}

// ID returns the unique instance identifier (e.g., "5m_2").
func (t Tile) ID() string {
	return fmt.Sprintf("%s_%d", t.Code, t.Instance)
}

// Wall represents the shuffled wall of tiles to draw from.
type Wall struct {
	tiles []Tile
	drawn int // index of next tile to draw from front
	back  int // index of next tile to draw from back (for kong replacements)
}

// NewWall creates a full 136-tile wall.
func NewWall() *Wall {
	tiles := make([]Tile, 0, 136)
	for _, code := range models.AllTileCodes {
		for i := 0; i < 4; i++ {
			tiles = append(tiles, Tile{Code: code, Instance: i})
		}
	}
	return &Wall{tiles: tiles, drawn: 0, back: 135}
}

// Shuffle randomizes the wall order.
func (w *Wall) Shuffle() {
	rand.Shuffle(len(w.tiles), func(i, j int) {
		w.tiles[i], w.tiles[j] = w.tiles[j], w.tiles[i]
	})
}

// Draw takes the next tile from the front of the wall.
// Returns the tile and true, or zero Tile and false if wall is exhausted.
func (w *Wall) Draw() (Tile, bool) {
	if w.drawn > w.back {
		return Tile{}, false
	}
	t := w.tiles[w.drawn]
	w.drawn++
	return t, true
}

// DrawBack takes a tile from the back of the wall (used after kong declarations).
func (w *Wall) DrawBack() (Tile, bool) {
	if w.drawn > w.back {
		return Tile{}, false
	}
	t := w.tiles[w.back]
	w.back--
	return t, true
}

// Remaining returns how many tiles are left to draw.
func (w *Wall) Remaining() int {
	if w.drawn > w.back {
		return 0
	}
	return w.back - w.drawn + 1
}

// Peek returns the next tile without drawing it (used for laizi indicator).
func (w *Wall) Peek() (Tile, bool) {
	if w.drawn > w.back {
		return Tile{}, false
	}
	return w.tiles[w.drawn], true
}

// TileSuit returns the suit character for a tile code, or empty for honors.
func TileSuit(code models.TileCode) byte {
	s := string(code)
	if len(s) == 2 {
		return s[1]
	}
	return 0
}

// TileValue returns the numeric value for suited tiles (1-9), or 0 for honors.
func TileValue(code models.TileCode) int {
	s := string(code)
	if len(s) == 2 {
		ch := s[0]
		if ch >= '1' && ch <= '9' {
			return int(ch - '0')
		}
	}
	return 0
}

// IsSuited returns true if the tile is a numbered suit tile (万/条/筒).
func IsSuited(code models.TileCode) bool {
	suit := TileSuit(code)
	return suit == 'm' || suit == 's' || suit == 'p'
}

// IsHonor returns true if the tile is a wind or dragon tile.
func IsHonor(code models.TileCode) bool {
	return !IsSuited(code)
}

// IsValid258Pair returns true if the tile is a valid pair for winning (2, 5, or 8 of any suit).
func IsValid258Pair(code models.TileCode) bool {
	if !IsSuited(code) {
		return false
	}
	v := TileValue(code)
	return v == 2 || v == 5 || v == 8
}

// NextInSequence returns the next tile in the same suit (wrapping 9→1).
// Returns empty string for honor tiles.
func NextInSequence(code models.TileCode) models.TileCode {
	if !IsSuited(code) {
		return ""
	}
	v := TileValue(code)
	suit := TileSuit(code)
	next := v%9 + 1
	return models.TileCode(fmt.Sprintf("%d%c", next, suit))
}

// PrevInSequence returns the previous tile in the same suit (wrapping 1→9).
func PrevInSequence(code models.TileCode) models.TileCode {
	if !IsSuited(code) {
		return ""
	}
	v := TileValue(code)
	suit := TileSuit(code)
	prev := v - 1
	if prev == 0 {
		prev = 9
	}
	return models.TileCode(fmt.Sprintf("%d%c", prev, suit))
}

// DealHands deals 13 tiles to each of 4 players plus 1 extra to the dealer.
// Returns [4][]Tile where the dealer (index dealerSeat) has 14 tiles.
func DealHands(wall *Wall, dealerSeat int) ([4][]Tile, bool) {
	var hands [4][]Tile
	for i := range 4 {
		hands[i] = make([]Tile, 0, 14)
	}

	// Deal 4 tiles at a time, 3 rounds = 12 tiles each
	for round := range 3 {
		_ = round
		for seat := range 4 {
			for range 4 {
				t, ok := wall.Draw()
				if !ok {
					return hands, false
				}
				hands[seat] = append(hands[seat], t)
			}
		}
	}

	// Deal 1 more tile to each player = 13 each
	for seat := range 4 {
		t, ok := wall.Draw()
		if !ok {
			return hands, false
		}
		hands[seat] = append(hands[seat], t)
	}

	// Dealer draws 14th tile
	t, ok := wall.Draw()
	if !ok {
		return hands, false
	}
	hands[dealerSeat] = append(hands[dealerSeat], t)

	return hands, true
}

// TileCodesToMap converts a slice of tile codes to a frequency map.
func TileCodesToMap(codes []models.TileCode) map[models.TileCode]int {
	m := make(map[models.TileCode]int)
	for _, c := range codes {
		m[c]++
	}
	return m
}

// TilesToCodes extracts tile codes from a slice of Tile.
func TilesToCodes(tiles []Tile) []models.TileCode {
	codes := make([]models.TileCode, len(tiles))
	for i, t := range tiles {
		codes[i] = t.Code
	}
	return codes
}

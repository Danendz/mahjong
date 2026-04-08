package room

import (
	"math/rand"
)

var botNamePool = []string{
	"小龙", "凤凰", "麒麟", "白虎",
	"玄武", "朱雀", "青龙", "仙鹤",
	"金鱼", "蝴蝶", "熊猫", "翡翠",
}

// pickBotName returns a random Chinese name not already in use in the room.
func pickBotName(room *Room) string {
	used := make(map[string]bool)
	for _, p := range room.Players {
		if p != nil {
			used[p.Nickname] = true
		}
	}

	var available []string
	for _, name := range botNamePool {
		if !used[name] {
			available = append(available, name)
		}
	}

	if len(available) == 0 {
		return "Bot"
	}
	return available[rand.Intn(len(available))]
}

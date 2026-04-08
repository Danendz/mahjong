package bot

import (
	"math/rand"
	"sync"
	"time"

	"github.com/mahjong/backend/internal/models"
)

// Action delays for more natural-feeling bot play.
const (
	turnDelayBase     = 500 * time.Millisecond
	turnDelayJitter   = 200 * time.Millisecond
	reactDelayBase    = 1000 * time.Millisecond
	reactDelayJitter  = 500 * time.Millisecond
)

// Controller manages bot strategies and schedules their actions with delays.
type Controller struct {
	mu         sync.Mutex
	strategies map[int]Strategy
	timers     map[int]*time.Timer
}

// NewController creates a new bot controller.
func NewController() *Controller {
	return &Controller{
		strategies: make(map[int]Strategy),
		timers:     make(map[int]*time.Timer),
	}
}

// RegisterBot registers a bot with the given strategy for a seat.
func (c *Controller) RegisterBot(seat int, difficulty models.BotDifficulty) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Cancel any existing timer for this seat
	if timer, ok := c.timers[seat]; ok {
		timer.Stop()
		delete(c.timers, seat)
	}

	c.strategies[seat] = NewStrategy(difficulty)
}

// UnregisterBot removes a bot from the controller.
func (c *Controller) UnregisterBot(seat int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if timer, ok := c.timers[seat]; ok {
		timer.Stop()
		delete(c.timers, seat)
	}
	delete(c.strategies, seat)
}

// HasBot returns true if a bot is registered for the given seat.
func (c *Controller) HasBot(seat int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.strategies[seat]
	return ok
}

// ScheduleTurnAction schedules a bot's turn action with a short delay.
// The callback is invoked on a separate goroutine after the delay.
func (c *Controller) ScheduleTurnAction(seat int, ctx GameContext, callback func(TurnAction)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	strategy, ok := c.strategies[seat]
	if !ok {
		return
	}

	// Cancel any existing timer for this seat
	if timer, ok := c.timers[seat]; ok {
		timer.Stop()
	}

	delay := turnDelayBase + time.Duration(rand.Int63n(int64(turnDelayJitter)))

	c.timers[seat] = time.AfterFunc(delay, func() {
		action := strategy.ChooseTurnAction(ctx)
		callback(action)
	})
}

// ScheduleReaction schedules a bot's reaction with a slightly longer delay.
// The callback is invoked on a separate goroutine after the delay.
func (c *Controller) ScheduleReaction(seat int, ctx GameContext, callback func(ReactionAction)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	strategy, ok := c.strategies[seat]
	if !ok {
		return
	}

	if timer, ok := c.timers[seat]; ok {
		timer.Stop()
	}

	delay := reactDelayBase + time.Duration(rand.Int63n(int64(reactDelayJitter)))

	c.timers[seat] = time.AfterFunc(delay, func() {
		action := strategy.ChooseReaction(ctx)
		callback(action)
	})
}

// CancelAll stops all pending bot timers.
func (c *Controller) CancelAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for seat, timer := range c.timers {
		timer.Stop()
		delete(c.timers, seat)
	}
}

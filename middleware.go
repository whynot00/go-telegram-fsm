package fsm

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Middleware injects the FSM instance into the context for each update.
func Middleware(fsm *FSM) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			// Store FSM in context for downstream handlers
			ctx = context.WithValue(ctx, FsmKey, fsm)

			// Access current state to update last usage timestamp (optional)
			fsm.CurrentState(update.Message.From.ID)

			next(ctx, b, update)
		}
	}
}

// WithStates returns a middleware that allows the handler
// to run only if the user's current FSM state matches one of the specified states.
func WithStates(states ...StateFSM) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			fsm := FromContext(ctx)
			userID := update.Message.From.ID
			currentState := fsm.CurrentState(userID)

			for _, state := range states {
				if state == StateAny || state == currentState {
					next(ctx, b, update)
					return
				}
			}

			// If no matching state found, handler is skipped
		}
	}
}

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

			if update != nil {

				// Access current state to update last usage timestamp (optional)
				fsm.CurrentState(extractUserID(update))
			}

			// Store FSM in context for downstream handlers
			ctx = context.WithValue(ctx, FsmKey, fsm)

			next(ctx, b, update)
		}
	}
}

// WithStates returns a middleware that allows the handler to run only if the user's current FSM state matches one of the specified states.
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

			// If no matching state is found, the handler is skipped.
		}
	}
}

// extractUserID returns the ID of the user associated with the update.
func extractUserID(u *models.Update) int64 {

	switch {
	case u.Message != nil && u.Message.From != nil:
		return u.Message.From.ID
	case u.EditedMessage != nil && u.EditedMessage.From != nil:
		return u.EditedMessage.From.ID
	case u.BusinessMessage != nil && u.BusinessMessage.From != nil:
		return u.BusinessMessage.From.ID
	case u.EditedBusinessMessage != nil && u.EditedBusinessMessage.From != nil:
		return u.EditedBusinessMessage.From.ID
	case u.CallbackQuery != nil:
		return u.CallbackQuery.From.ID
	case u.InlineQuery != nil && u.InlineQuery.From != nil:
		return u.InlineQuery.From.ID
	case u.ChosenInlineResult != nil:
		return u.ChosenInlineResult.From.ID
	case u.ShippingQuery != nil && u.ShippingQuery.From != nil:
		return u.ShippingQuery.From.ID
	case u.PreCheckoutQuery != nil && u.PreCheckoutQuery.From != nil:
		return u.PreCheckoutQuery.From.ID
	case u.PurchasedPaidMedia != nil:
		return u.PurchasedPaidMedia.From.ID
	case u.ChatMember != nil:
		return u.ChatMember.From.ID
	case u.MyChatMember != nil:
		return u.MyChatMember.From.ID
	case u.ChatJoinRequest != nil:
		return u.ChatJoinRequest.From.ID
	case u.PollAnswer != nil && u.PollAnswer.User != nil:
		return u.PollAnswer.User.ID
	case u.MessageReaction != nil && u.MessageReaction.User != nil:
		return u.MessageReaction.User.ID
	}
	return 0
}

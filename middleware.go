package fsm

import (
	"context"
	"slices"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Middleware attaches the FSM instance and user ID (if present)
// to the context for every incoming update.
// User state is created lazily only if a valid user ID (>0) is extracted.
func Middleware(fsm *FSM) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update != nil {
				uid := extractUserID(update)
				if uid > 0 {
					ctx = userWithContext(ctx, uid)
					fsm.Create(ctx)
				}
			}

			// Inject FSM instance into context for downstream handlers.
			ctx = fsmWithContext(ctx, fsm)

			next(ctx, b, update)
		}
	}
}

// WithStates restricts handler execution to specific FSM states.
// - If no states are provided → handler is always executed.
// - If StateAny is provided → handler is always executed.
// - Otherwise → handler runs only when the current state matches one of the provided states.
// If no FSM or state is found, the handler is skipped.
func WithStates(states ...StateFSM) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if len(states) == 0 {
				next(ctx, b, update)
				return
			}

			if slices.Contains(states, StateAny) {
				next(ctx, b, update)
				return
			}

			fsm := FromContext(ctx)
			if fsm == nil {
				return // no FSM → skip handler
			}

			currentState, ok := fsm.CurrentState(ctx)
			if !ok {
				return // no state → skip handler
			}

			if slices.Contains(states, currentState) {
				next(ctx, b, update)
			}
		}
	}
}

// extractUserID extracts the user ID from an incoming update.
// Returns 0 if no valid user is present in the update.
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

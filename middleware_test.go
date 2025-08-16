package fsm_test

import (
	"context"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/require"
	fsm "github.com/whynot00/go-telegram-fsm"
)

func TestMiddleware_AttachesFSMToContext(t *testing.T) {

	f := fsm.New(context.Background())

	mw := fsm.Middleware(f)

	called := false

	handler := func(ctx context.Context, b *bot.Bot, update *models.Update) {
		val := fsm.FromContext(ctx)
		require.Equal(t, f, val)
		called = true
	}

	mw(handler)(context.Background(), &bot.Bot{}, &models.Update{
		Message: &models.Message{
			From: &models.User{ID: 123},
		},
	})

	require.True(t, called)
}

func TestWithStates_MatchingState(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(1)

	f.Transition(userID, "state1")

	ctx := fsm.WithContext(context.Background(), f)

	fsmInCtx := fsm.FromContext(ctx)
	if fsmInCtx == nil {
		t.Fatal("FSM not found in context")
	}

	currentState := fsmInCtx.CurrentState(userID)
	if currentState != "state1" {
		t.Fatalf("Unexpected state for user %d: got %s, want %s", userID, currentState, "state1")
	}

	called := false

	handler := func(ctx context.Context, b *bot.Bot, update *models.Update) {
		called = true
	}

	mw := fsm.WithStates("state1")
	mw(handler)(ctx, &bot.Bot{}, &models.Update{
		Message: &models.Message{
			From: &models.User{ID: userID},
		},
	})

	if !called {
		t.Fatal("Handler was not called")
	}
}

func TestWithStates_NonMatchingState(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(2)
	f.Set(userID, "state1", "val")

	ctx := context.WithValue(context.Background(), fsm.FsmKey, f)

	called := false

	handler := func(ctx context.Context, b *bot.Bot, update *models.Update) {
		called = true
	}

	mw := fsm.WithStates("another_state")
	mw(handler)(ctx, &bot.Bot{}, &models.Update{
		Message: &models.Message{
			From: &models.User{ID: userID},
		},
	})

	require.False(t, called)
}

func TestWithStates_StateAny(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(3)

	ctx := context.WithValue(context.Background(), fsm.FsmKey, f)

	called := false

	handler := func(ctx context.Context, b *bot.Bot, update *models.Update) {
		called = true
	}

	mw := fsm.WithStates(fsm.StateAny)
	mw(handler)(ctx, &bot.Bot{}, &models.Update{
		Message: &models.Message{
			From: &models.User{ID: userID},
		},
	})

	require.True(t, called)
}

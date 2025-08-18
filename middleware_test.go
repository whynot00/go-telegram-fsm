package fsm

import (
	"context"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// spyFSM: нам важно отследить Create-вызов и что FSM попал в контекст.
type spyFSM struct {
	FSM
	createdFor int64
}

func (s *spyFSM) Create(ctx context.Context) {
	s.createdFor = userFromContext(ctx)
}

func TestMiddleware_SetsUserAndFSM_WhenUserPresent(t *testing.T) {
	fsm := &FSM{}
	mw := Middleware(fsm)

	called := false
	next := func(ctx context.Context, _ *bot.Bot, _ *models.Update) {
		called = true

		// FSM должен быть в контексте
		if FromContext(ctx) == nil {
			t.Fatalf("FSM not injected in context")
		}
		// userID должен быть установлен
		if got := userFromContext(ctx); got != 123 {
			t.Fatalf("user id not injected, got %d", got)
		}
	}

	upd := &models.Update{
		Message: &models.Message{From: &models.User{ID: 123}},
	}
	handler := mw(next)
	handler(context.Background(), nil, upd)

	if !called {
		t.Fatalf("next handler was not called")
	}

	// Проверяем, что Create реально создал запись в FSM
	if _, ok := fsm.current.Load(int64(123)); !ok {
		t.Fatalf("expected FSM entry for user 123 after Middleware")
	}
}

func TestMiddleware_DoesNotCreate_WhenNoUser(t *testing.T) {
	spy := &spyFSM{}
	mw := Middleware(&spy.FSM)

	next := func(ctx context.Context, _ *bot.Bot, upd *models.Update) {
		// FSM в контексте должен быть даже без userID
		if FromContext(ctx) == nil {
			t.Fatalf("FSM not injected in context")
		}
		// userID не должен ставиться
		if userFromContext(ctx) != 0 {
			t.Fatalf("unexpected user id %d", userFromContext(ctx))
		}
	}
	handler := mw(next)
	// update без From → uid=0
	upd := &models.Update{ChannelPost: &models.Message{}} // вариант без From
	handler(context.Background(), nil, upd)

	if spy.createdFor != 0 {
		t.Fatalf("Create must not be called for uid=0, got %d", spy.createdFor)
	}
}

func TestWithStates_BasicBranches(t *testing.T) {
	f := &FSM{}
	ctx := fsmWithContext(context.Background(), f)
	ctx = userWithContext(ctx, 77)
	f.current.Store(int64(77), stateData{state: StateFSM("A")})

	run := 0
	next := func(ctx context.Context, _ *bot.Bot, _ *models.Update) { run++ }

	// пустой список — пропуск ограничений (выполняется)
	WithStates()(next)(ctx, nil, &models.Update{})
	// StateAny — выполняется
	WithStates(StateAny)(next)(ctx, nil, &models.Update{})
	// точное совпадение — выполняется
	WithStates(StateFSM("A"))(next)(ctx, nil, &models.Update{})
	// не совпало — не выполняется
	WithStates(StateFSM("B"))(next)(ctx, nil, &models.Update{})

	if run != 3 {
		t.Fatalf("expected next to run 3 times, got %d", run)
	}
}

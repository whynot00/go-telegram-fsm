package fsm_test

import (
	"context"
	"testing"
	"time"

	fsm "github.com/whynot00/go-telegram-fsm"
	"github.com/whynot00/go-telegram-fsm/media"
	"github.com/whynot00/go-telegram-fsm/storage/memory"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const testUserID int64 = 123456

// buildUpdate — минимальный апдейт с пользователем.
func buildUpdate(uid int64) *models.Update {
	return &models.Update{
		ID: 1,
		Message: &models.Message{
			ID: 10,
			From: &models.User{
				ID: uid,
			},
			Text: "ping",
		},
	}
}

// chain — склеивает middleware в один хендлер.
func chain(h bot.HandlerFunc, mws ...bot.Middleware) bot.HandlerFunc {
	wrapped := h
	for i := len(mws) - 1; i >= 0; i-- {
		wrapped = mws[i](wrapped)
	}
	return wrapped
}

func TestIntegration_FullFlow_StateAndHandlers(t *testing.T) {
	ctx := context.Background()

	// Явно задаём память с мелкими TTL/cleanup, чтобы было дёшево.
	store := memory.NewMemoryStorage(5*time.Minute, 100*time.Millisecond)
	f := fsm.New(ctx, fsm.WithStorage(store))

	// 1) Первый обработчик разрешён в StateDefault, делает Transition в кастомный стейт.
	const NextState fsm.StateFSM = "42"

	var calledDefault, calledNext bool

	hDefault := func(ctx context.Context, _ *bot.Bot, _ *models.Update) {
		calledDefault = true

		// До перехода проверим состояние — должно быть StateDefault после Middleware.Create.
		if st, ok := f.CurrentState(ctx); !ok || st != fsm.StateDefault {
			t.Fatalf("expected initial state=StateDefault, got (%v, ok=%v)", st, ok)
		}
		f.Transition(ctx, NextState)
	}

	// 2) Второй обработчик разрешён только в NextState.
	hNext := func(ctx context.Context, _ *bot.Bot, _ *models.Update) {
		calledNext = true
		if st, ok := f.CurrentState(ctx); !ok || st != NextState {
			t.Fatalf("expected state=%s inside second handler, got (%v, ok=%v)", NextState, st, ok)
		}
	}

	final := chain(
		// Итоговый хендлер прогонит оба шага последовательно:
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			// руками запускаем вложенный пайп: сначала default-хендлер…
			chain(hDefault, fsm.WithStates(fsm.StateDefault))(ctx, b, u)
			// …после перехода — хендлер для NextState.
			chain(hNext, fsm.WithStates(NextState))(ctx, b, u)
		},
		// Глобальное Middleware, которое кладёт FSM и user в контекст и вызывает Create().
		fsm.Middleware(f),
	)

	u := buildUpdate(testUserID)
	final(ctx, nil, u)

	if !calledDefault {
		t.Fatal("default-state handler was not called")
	}
	if !calledNext {
		t.Fatal("next-state handler was not called")
	}
}

func TestIntegration_CacheAndMedia_CleanOnFinish(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStorage(5*time.Minute, 100*time.Millisecond)
	f := fsm.New(ctx, fsm.WithStorage(store))

	var (
		key          = "k1"
		mediaGroupID = "mg-777"
	)

	// Хендлер: кладём данные в кеш и медиагруппу, убеждаемся, что они читаются,
	// затем Finish() и проверяем, что кеш очищен.
	handler := func(ctx context.Context, _ *bot.Bot, _ *models.Update) {
		// Перед началом — мы уже в Default (через Middleware.Create).
		if st, ok := f.CurrentState(ctx); !ok || st != fsm.StateDefault {
			t.Fatalf("expected initial state=StateDefault, got (%v, ok=%v)", st, ok)
		}

		// Переходим в рабочий стейт, чтобы очистка не сработала преждевременно.
		const Work fsm.StateFSM = "5"
		f.Transition(ctx, Work)

		// Кладём обычный key/value.
		f.Set(ctx, testUserID, key, "value-123")
		if v, ok := f.Get(ctx, testUserID, key); !ok || v.(string) != "value-123" {
			t.Fatalf("expected cached value, got (%v, ok=%v)", v, ok)
		}

		// Кладём медиа.
		f.SetMedia(ctx, testUserID, mediaGroupID, media.File{Type: "photo", FileID: "A"})
		f.SetMedia(ctx, testUserID, mediaGroupID, media.File{Type: "video", FileID: "B"})

		md, ok := f.GetMedia(ctx, testUserID, mediaGroupID)
		if !ok {
			t.Fatal("expected media data to exist")
		}
		files := md.Files()
		if len(files) != 2 {
			t.Fatalf("expected 2 media files, got %d", len(files))
		}

		// Finish -> Transition(StateDefault) => должен вызвать CleanCache.
		f.Finish(ctx)

		// Кеш должен быть пуст.
		if _, ok := f.Get(ctx, testUserID, key); ok {
			t.Fatal("expected value cache to be cleaned after Finish")
		}

		// А медиакеш? По текущей логике CleanCache чистит всё пользовательское —
		// проверим, что этой группы больше нет (или она пуста).
		if md2, ok := f.GetMedia(ctx, testUserID, mediaGroupID); ok && len(md2.Files()) > 0 {
			t.Fatal("expected media cache to be cleaned after Finish")
		}
	}

	final := chain(handler, fsm.Middleware(f))
	final(ctx, nil, buildUpdate(testUserID))
}

func TestIntegration_WithStates_SkipsWithoutMiddleware(t *testing.T) {
	ctx := context.Background()
	f := fsm.New(ctx, fsm.WithStorage(memory.NewMemoryStorage(5*time.Minute, 100*time.Millisecond)))

	called := false
	h := func(ctx context.Context, _ *bot.Bot, _ *models.Update) {
		called = true
	}

	// ВАЖНО: не добавляем fsm.Middleware(f), значит в контексте нет ни FSM, ни user.
	onlyDefault := chain(h, fsm.WithStates(fsm.StateDefault))

	onlyDefault(ctx, nil, buildUpdate(testUserID))

	if called {
		t.Fatal("handler should have been skipped without FSM in context")
	}

	// А теперь добавим Middleware — должно пройти.
	called = false
	withMw := chain(h, fsm.Middleware(f), fsm.WithStates(fsm.StateDefault))
	withMw(ctx, nil, buildUpdate(testUserID))
	if !called {
		t.Fatal("handler should have been called with FSM middleware")
	}
}

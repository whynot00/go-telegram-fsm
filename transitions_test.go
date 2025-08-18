package fsm

import (
	"context"
	"testing"
	"time"

	"github.com/whynot00/go-telegram-fsm/media"
)

// --- stub storage to observe CleanCache calls ---

type stubStorage struct {
	cleanCalled int
	lastUserID  int64
}

func (s *stubStorage) Set(context.Context, int64, string, any)             {}
func (s *stubStorage) Get(context.Context, int64, string) (any, bool)      { return nil, false }
func (s *stubStorage) SetMedia(context.Context, int64, string, media.File) {}
func (s *stubStorage) GetMedia(context.Context, int64, string) (*media.MediaData, bool) {
	return nil, false
}
func (s *stubStorage) CleanMediaCache(context.Context, int64, string) bool { return false }
func (s *stubStorage) CleanCache(_ context.Context, userID int64) {
	s.cleanCalled++
	s.lastUserID = userID
}
func (s *stubStorage) Close() {}

// helper to build FSM with stub storage
func newTestFSM() (*FSM, *stubStorage) {
	ss := &stubStorage{}
	return &FSM{
		// current: zero value is fine
		storage:         ss,
		ownsStorage:     false,
		ttl:             time.Minute,
		cleanupInterval: time.Second,
	}, ss
}

func TestCreate_IdempotentAndDefault(t *testing.T) {
	t.Parallel()
	f, _ := newTestFSM()
	ctx := userWithContext(context.Background(), 1001)

	before := time.Now()
	f.Create(ctx)

	// first create → entry must exist with default state
	v, ok := f.current.Load(int64(1001))
	if !ok {
		t.Fatalf("expected entry after Create")
	}
	sd := v.(stateData)
	if sd.state != StateDefault {
		t.Fatalf("expected StateDefault, got %q", sd.state)
	}
	if sd.lastUse.Before(before) {
		t.Fatalf("lastUse not updated, got %v, before %v", sd.lastUse, before)
	}

	// second create → must not overwrite existing state
	// set a custom state, then call Create again
	f.current.Store(int64(1001), stateData{state: StateFSM("custom"), lastUse: time.Unix(1, 0)})
	f.Create(ctx)

	v2, _ := f.current.Load(int64(1001))
	sd2 := v2.(stateData)
	if sd2.state != StateFSM("custom") {
		t.Fatalf("Create must not overwrite existing state, got %q", sd2.state)
	}
}

func TestTransition_SetsStateAndUpdatesLastUse(t *testing.T) {
	t.Parallel()
	f, _ := newTestFSM()
	ctx := userWithContext(context.Background(), 2002)

	before := time.Now()
	f.Transition(ctx, StateFSM("step1"))

	v, ok := f.current.Load(int64(2002))
	if !ok {
		t.Fatalf("expected entry after Transition")
	}
	sd := v.(stateData)
	if sd.state != StateFSM("step1") {
		t.Fatalf("expected state step1, got %q", sd.state)
	}
	if sd.lastUse.Before(before) {
		t.Fatalf("lastUse should be updated to now, got %v (before %v)", sd.lastUse, before)
	}
}

func TestTransition_Default_CallsCleanCache(t *testing.T) {
	t.Parallel()
	f, stub := newTestFSM()
	ctx := userWithContext(context.Background(), 3003)

	f.Transition(ctx, StateDefault)

	if stub.cleanCalled == 0 {
		t.Fatalf("expected CleanCache to be called on StateDefault transition")
	}
	if stub.lastUserID != 3003 {
		t.Fatalf("expected CleanCache userID=3003, got %d", stub.lastUserID)
	}
}

func TestFinish_WrapsTransitionToDefault(t *testing.T) {
	t.Parallel()
	f, stub := newTestFSM()
	ctx := userWithContext(context.Background(), 4004)

	// set non-default first
	f.Transition(ctx, StateFSM("work"))
	// finish
	f.Finish(ctx)

	// state should be default now
	v, ok := f.current.Load(int64(4004))
	if !ok {
		t.Fatalf("expected entry after Finish")
	}
	sd := v.(stateData)
	if sd.state != StateDefault {
		t.Fatalf("expected StateDefault after Finish, got %q", sd.state)
	}
	if stub.cleanCalled == 0 {
		t.Fatalf("expected CleanCache to be called during Finish")
	}
}

func TestCurrentState_MissReturnsNilFalse(t *testing.T) {
	t.Parallel()
	f, _ := newTestFSM()
	ctx := userWithContext(context.Background(), 5005)

	state, ok := f.CurrentState(ctx)
	if ok {
		t.Fatalf("expected miss (ok=false) when no entry exists, got ok=true")
	}
	if state != StateNil {
		t.Fatalf("expected StateNil on miss, got %q", state)
	}
}

func TestCurrentState_HitUpdatesLastUseAndReturnsState(t *testing.T) {
	t.Parallel()
	f, _ := newTestFSM()
	uid := int64(6006)
	// seed entry with old lastUse and custom state
	old := time.Now().Add(-time.Hour)
	f.current.Store(uid, stateData{state: StateFSM("in_progress"), lastUse: old})

	ctx := userWithContext(context.Background(), uid)
	ret, ok := f.CurrentState(ctx)
	if !ok {
		t.Fatalf("expected hit")
	}
	if ret != StateFSM("in_progress") {
		t.Fatalf("expected state in_progress, got %q", ret)
	}
	// lastUse should be refreshed
	v, _ := f.current.Load(uid)
	sd := v.(stateData)
	if !sd.lastUse.After(old) {
		t.Fatalf("expected lastUse to be updated, old=%v new=%v", old, sd.lastUse)
	}
}

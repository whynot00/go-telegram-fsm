package fsm_test

import (
	"context"
	"testing"

	fsm "github.com/whynot00/go-telegram-fsm"
)

func TestFSM_TransitionAndCurrentState(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(1234)

	f.Transition(userID, "start")
	if f.CurrentState(userID) != "start" {
		t.Errorf("expected state 'start', got '%s'", f.CurrentState(userID))
	}

	f.Transition(userID, "next")
	if f.CurrentState(userID) != "next" {
		t.Errorf("expected state 'next', got '%s'", f.CurrentState(userID))
	}
}

func TestFSM_SetAndGet(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(5678)

	f.Set(userID, "key1", "value1")
	value, _ := f.Get(userID, "key1")
	if value != "value1" {
		t.Errorf("expected value 'value1', got '%s'", value)
	}
}

func TestFSM_CleanCache(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(9876)

	f.Set(userID, "key", "value")
	f.CleanCache(userID)

	if val, _ := f.Get(userID, "key"); val != nil {
		t.Errorf("expected empty value after CleanCache, got '%s'", val)
	}
}

func TestFSM_Finish(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(1357)

	f.Transition(userID, "processing")
	f.Set(userID, "key", "value")

	f.Finish(userID)

	if f.CurrentState(userID) != fsm.StateDefault {
		t.Errorf("expected empty state after Finish, got '%s'", f.CurrentState(userID))
	}
	if val, _ := f.Get(userID, "key"); val != nil {
		t.Errorf("expected empty cache after Finish, got '%s'", val)
	}
}

# go-telegram-fsm

`go-telegram-fsm` provides a lightweight finite state machine (FSM) for [go-telegram/bot](https://github.com/go-telegram/bot) based bots.  
The library tracks a state for each Telegram user, supplies handy middleware to inject the FSM into request context and offers a per-user cache for arbitrary data and media groups.

> **Why?**  
> Chat‑bot flows often look like conversational state machines: you ask a question, wait for a reply, move to the next step and so on.  
> This package handles the boilerplate so you can focus on your bot logic.

## Features

- Thread‑safe FSM with per‑user state and last access timestamp.
- Pluggable storage layer with an in‑memory implementation shipped by default.
- Automatic cleanup of stale states and cache entries based on TTL.
- Optional key/value cache and media‑group cache bound to a user.
- `bot.Middleware` that automatically:
  - extracts the user ID from incoming updates,
  - creates a default state entry for new users,
  - attaches both the FSM instance and user ID to `context.Context`.
- `fsm.WithStates` middleware to guard handlers by allowed states.
- Simple API: `Transition`, `Finish`, `CurrentState`, `Set`, `Get`, `SetMedia`, …
- Zero dependencies besides the Telegram SDK and the standard library.

## Installation

```bash
go get github.com/whynot00/go-telegram-fsm
```

The module requires Go 1.20+ (see `go.mod` for the exact version).

## Quick Start

```go
package main

import (
    "context"
    "time"

    fsm "github.com/whynot00/go-telegram-fsm"
    "github.com/go-telegram/bot"
    "github.com/go-telegram/bot/models"
)

func main() {
    ctx := context.Background()

    // Create FSM with custom TTL and cleanup interval.
    machine := fsm.New(ctx,
        fsm.WithTTL(30*time.Minute),
        fsm.WithCleanupInterval(30*time.Second),
    )

    b, _ := bot.New("<TOKEN>", bot.WithMiddlewares(fsm.Middleware(machine)))

    // /start is allowed only in the default state and moves user to "ask-name".
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeCommand,
		func(ctx context.Context, b *bot.Bot, upd *models.Update) {
			machine := fsm.FromContext(ctx)

			machine.Transition(ctx, "ask-name")
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: upd.Message.Chat.ID, Text: "Hi! What is your name?"})
		},
		fsm.WithStates(fsm.StateDefault),
	)

    // Handler for the next state
    b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypeExact,
        func(ctx context.Context, b *bot.Bot, upd *models.Update) {
            name := upd.Message.Text
            b.SendMessage(ctx, &bot.SendMessageParams{ChatID: upd.Message.Chat.ID, Text: "Nice to meet you, " + name})
            machine.Finish(ctx) // back to default, cache cleaned
        },
        fsm.WithStates("ask-name"),
    )

    b.Start(ctx)
}
```

## FSM Concepts

### States
A state is represented by `fsm.StateFSM` (alias of `string`).
Three special states are provided:

- `StateDefault` – automatically assigned to every new user. Transitioning back to this state also clears the user's cache.
- `StateAny` – wildcard used in `WithStates` middleware to run a handler regardless of current state.
- `StateNil` – returned by `CurrentState` when no state exists for a user.

### Creating and Accessing States
Normally you do not create state manually – `Middleware` does it lazily when a user first interacts with the bot.  
Still, you can explicitly call `f.Create(ctx)` if required.  `CurrentState` returns the current state and refreshes the "last used" timestamp:

```go
st, ok := f.CurrentState(ctx)
if !ok {
    // no state stored yet
}
```

### Transitions and Finish
Use `Transition` to move a user to another state.  
Calling `Finish` is a shortcut for `Transition(ctx, StateDefault)` and also purges all cached data for that user:

```go
f.Transition(ctx, "awaiting_email")
...
f.Finish(ctx) // back to StateDefault + cache cleanup
```

## Middleware Integration

### Middleware(fsm)
`fsm.Middleware` wraps handlers to inject FSM and user ID into the context:

1. Extracts the user ID from `models.Update` (handles most Telegram update types).
2. Creates an entry with `StateDefault` if the user was not seen before.
3. Stores both the FSM instance and user ID in `context.Context` so downstream handlers can access them with `fsm.FromContext` and `userFromContext` (internally).

Attach it globally when creating the bot:

```go
b, _ := bot.New(token, bot.WithMiddlewares(fsm.Middleware(f)))
```

### WithStates
`fsm.WithStates` is an additional middleware that allows a handler to run only when a user's state matches one of the provided states:

```go
b.RegisterHandler(bot.HandlerTypeMessageText, "", handler, fsm.WithStates("step1", "step2"))
```

Special rules:

- No states passed → handler always runs.
- `StateAny` present → handler always runs.
- No FSM or no state in context → handler is skipped.

## User Cache

Each FSM instance also serves as a small per-user cache.  The storage implements the `storage.Storage` interface.  Functions operate on the user ID you pass explicitly:

```go
fsm.Set(ctx, userID, "key", 42)
val, ok := fsm.Get(ctx, userID, "key")
```

The default memory storage keeps cache items in `sync.Map` partitions and tracks the last access time per user.  When a state expires (by TTL) or you call `Finish`, the cache for that user is dropped.

### Media Group Cache
Telegram can send media as groups.  FSM keeps an in-memory accumulator per user & media group:

```go
file := media.File{Type: "photo", FileID: someID}
fsm.SetMedia(ctx, userID, mediaGroupID, file)

md, _ := fsm.GetMedia(ctx, userID, mediaGroupID)
files := md.Files() // copy of stored files
```

You may remove media groups manually with `CleanMediaCache` or wipe everything with `CleanCache`/`Finish`.

## Custom Storage

The storage backend is abstracted by the `storage.Storage` interface:

```go
type Storage interface {
    Set(ctx context.Context, userID int64, key string, value any)
    Get(ctx context.Context, userID int64, key string) (any, bool)
    SetMedia(ctx context.Context, userID int64, mediaGroupID string, file media.File)
    GetMedia(ctx context.Context, userID int64, mediaGroupID string) (*media.MediaData, bool)
    CleanMediaCache(ctx context.Context, userID int64, mediaGroupID string) bool
    CleanCache(ctx context.Context, userID int64)
    Close()
}
```

By default the FSM uses an in-memory implementation (`storage/memory`) that:

- partitions data by user ID,
- tracks last access time and runs a background goroutine to evict idle users,
- is safe for concurrent access.

Provide your own implementation and pass it via `WithStorage` option:

```go
store := redisStorage{...} // any struct implementing storage.Storage
f := fsm.New(ctx, fsm.WithStorage(store))
```

If you supply custom storage the FSM will not manage its lifecycle (no automatic `Close`).

## Configuration Options

Options are applied when creating an FSM instance:

```go
fsm.New(ctx,
    fsm.WithStorage(store),      // custom storage instead of in-memory
    fsm.WithTTL(time.Hour),      // how long to keep user state without activity
    fsm.WithCleanupInterval(time.Minute), // how often expired states are purged
)
```

## Testing

Run the test suite with:

```bash
go test ./...
```

It includes unit tests for state transitions, middleware behaviour and integration tests covering a typical conversation flow.

## License

This project is provided without an explicit license file.  Use at your own risk or contact the author to clarify licensing terms.
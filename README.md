## How It Works

This package implements a lightweight **finite state machine (FSM)** for managing user states and per-user local storage in Telegram bot applications built with [`go-telegram/bot`](https://github.com/go-telegram/bot).

### Core Concepts

- **User State Tracking:**  
  The FSM stores the current state of each user by their unique user ID. States are represented as strings (`StateFSM`), allowing flexible definition of conversation or workflow steps.

- **Local Storage:**  
  Along with the state, the FSM maintains a concurrent key-value cache for each user to store arbitrary data associated with that user's session.

- **Automatic Cleanup:**  
  A background worker periodically scans for user states that have not been used for a configured TTL (default 30 minutes) and automatically removes them to free resources.

### Usage Flow

1. **Initialization:**  
   Create a new FSM instance with `fsm.New(ctx)`. The cleanup worker will start automatically.

2. **Middleware Integration:**  
   Inject the FSM into request contexts using the provided `Middleware(fsm *FSM) bot.Middleware`. This makes the FSM accessible in all update handlers.

3. **State Management:**  
   Use `Transition(userID, newState)` to update a user's state, which also resets the last activity timestamp. Calling `Finish(userID)` resets the state to the default and clears the user cache.

4. **Accessing State and Data:**  
   Retrieve the current state with `CurrentState(userID)`. Use `Set(userID, key, value)` and `Get(userID, key)` to manage arbitrary per-user session data.

5. **Conditional Handlers:**  
   The `WithStates(states ...StateFSM)` middleware allows you to guard handlers to execute only if the user is in one of the specified states.

### Benefits

- **Concurrency Safe:**  
  Internally uses `sync.Map` for concurrent access without explicit locking.

- **Extensible:**  
  Easily integrate with Telegram bot handlers and extend states and user session data.

- **Resource Efficient:**  
  Periodic cleanup prevents unbounded memory growth by removing inactive user data.

## Getting Started

### Installation

```bash
go get github.com/whynot00/go-teleram-fsm
```

```go
const (
    someState fsm.StateFSM = "some_state"
)

func main() {
    ctx := context.Background()

    // Create FSM Instance
    fsm := fsm.New(ctx)

    // Inject FSM into bot middleware
    opts := []bot.Option{

        // It is used to attach the FSM to each request
        bot.WithMiddlewares(fsm.Middleware(fsm)),
    }

    b, err := bot.New("YOUR_TELEGRAM_BOT_API_TOKEN", botOptions...)
    if err != nil {
        log.Fatalf("failed to create bot: %v", err)
    }

    // Register handlers with FSM state filtering
    b.RegisterHandler(
        bot.HandlerTypeMessageText,
        "start",
        bot.MatchTypeCommand,
        yourHandlerStart,
        fsm.WithStates(fsm.StateDefault),
    )

    b.RegisterHandler(
        bot.HandlerTypeMessageText,
        "next",
        bot.MatchTypeCommand,
        yourHandlerNext,
        fsm.WithStates(someState),
    )
}

// Manage user states and session data inside handlers

func yourHandlerStart(ctx context.Context, b *bot.Bot, update *models.Update) {
    userID := update.Message.From.ID
    fsm := fsm.FromContext(ctx)

    // Transition user to new state
    fsm.Transition(userID, someState)

    // Store arbitrary data
    fsm.Set(userID, "key", "value")

    b.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "Welcome! State changed.",
    })
}

func yourHandlerNext(ctx context.Context, b *bot.Bot, update *models.Update) {
    userID := update.Message.From.ID
    fsm := fsm.FromContext(ctx)

    if val, ok := fsm.Get(userID, "key"); ok {
        // Use cached data
        fmt.Println("Cached value:", val)
    }

    fsm.Finish(userID) // Reset to default state

    b.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "State reset to default.",
    })
}
```
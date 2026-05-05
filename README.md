# kick-chat

Anonymous Kick.com chat client for Go. Connects via Pusher WebSocket — no OAuth or authentication required.

## Install

```bash
go get github.com/DiegPS/kick-chat
```

## Usage

```go
package main

import (
    "context"
    "fmt"

    kickchat "github.com/DiegPS/kick-chat"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    client, err := kickchat.NewClient(ctx)
    if err != nil {
        panic(err)
    }
    defer client.Close()

    if err := client.JoinChannelBySlug("xqc"); err != nil {
        panic(err)
    }

    for msg := range client.Messages() {
        fmt.Printf("[%s] %s:\n", msg.CreatedAt.Format("15:04:05"), msg.Sender.Username)

        // Render message parts in order (text and emotes interleaved)
        for _, part := range msg.Parts {
            if part.Emote != nil {
                fmt.Printf("  [emote: %s] ", part.Emote.Name)
            } else {
                fmt.Printf("  %s", part.Text)
            }
        }
        fmt.Println()
    }
}
```

## API

### Client

```go
func NewClient(ctx context.Context) (*Client, error)
func (c *Client) JoinChannelBySlug(slug string) error   // resolves slug → chatroom ID automatically
func (c *Client) JoinChannelByID(id int) error
func (c *Client) Messages() <-chan ChatMessage
func (c *Client) Errors() <-chan error
func (c *Client) Close()
func (c *Client) SetDebug(v bool)
```

### Types

```go
type ChatMessage struct {
    ID         string
    ChatroomID int
    Content    string        // raw content, may contain [emote:id:name] tokens
    Type       string
    CreatedAt  time.Time
    Sender     Sender
    Emotes     []ParsedEmote // unique emotes present in Content (deduplicated)
    Parts      []MessagePart // Content split into text+emote segments, ready to render
}

type Sender struct {
    ID       int
    Username string
    Slug     string
    Identity Identity
}

type Identity struct {
    Color  string
    Badges []Badge
}

type Badge struct {
    Type  string // "broadcaster", "moderator", "subscriber", "sub_gifter", "og", "verified"
    Text  string
    Count int
}

type ParsedEmote struct {
    ID   string
    Name string
    URL  string // https://files.kick.com/emotes/{id}/fullsize
}

type MessagePart struct {
    Text  string       // plain text segment; empty when Emote is set
    Emote *ParsedEmote // nil when Text is set
}
```

### Utilities

```go
// Resolve a channel slug to its numeric chatroom ID (public API, no auth needed)
func GetChatroomID(slug string) (int, error)

// Extract unique emotes from raw message content
func ParseEmotes(content string) []ParsedEmote

// Split raw message content into ordered text+emote segments for direct rendering
func ParseMessage(content string) []MessagePart
```

## Notes

- Emotes in `Content` are encoded as `[emote:37225:KEKLEO]` — `ParsedEmote.URL` points to the Kick CDN.
- `msg.Parts` is the recommended way to render messages — it preserves the original order of text and emotes.
- `msg.Emotes` is useful when you only need to know which emotes are present (e.g. for pre-fetching or analytics), not their position in the text.
- `Errors()` surfaces non-fatal errors (e.g. failed reconnect attempts). The client reconnects automatically on disconnect.
- Cancel the context or call `Close()` to shut down cleanly.

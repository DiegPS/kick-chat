package kickchat

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const pusherURL = "wss://ws-us2.pusher.com/app/32cbd69e4b950bf97679?protocol=7&client=js&version=8.4.0-rc2&flash=false"

// Client connects to Kick's Pusher WebSocket and delivers chat messages.
// Create one with NewClient — it is safe to call JoinChannelBySlug and
// JoinChannelByID from any goroutine after construction.
type Client struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.Mutex
	ws       *websocket.Conn
	channels map[int]bool // chatroom IDs currently joined

	msgCh chan ChatMessage
	errCh chan error

	debug bool
}

// NewClient dials the Pusher WebSocket and starts the read loop.
// The provided context controls the lifetime of the connection —
// cancel it (or call Close) to shut down cleanly.
func NewClient(ctx context.Context) (*Client, error) {
	ws, _, err := websocket.DefaultDialer.Dial(pusherURL, nil)
	if err != nil {
		return nil, fmt.Errorf("kick: websocket dial: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	c := &Client{
		ctx:      ctx,
		cancel:   cancel,
		ws:       ws,
		channels: make(map[int]bool),
		msgCh:    make(chan ChatMessage, 64),
		errCh:    make(chan error, 8),
	}

	go c.readLoop()
	return c, nil
}

// SetDebug enables or disables verbose logging to stdout.
func (c *Client) SetDebug(v bool) {
	c.mu.Lock()
	c.debug = v
	c.mu.Unlock()
}

// JoinChannelBySlug resolves the channel slug to a chatroom ID and joins it.
func (c *Client) JoinChannelBySlug(slug string) error {
	id, err := GetChatroomID(slug)
	if err != nil {
		return err
	}
	return c.JoinChannelByID(id)
}

// JoinChannelByID subscribes to a Pusher chatroom by its numeric ID.
func (c *Client) JoinChannelByID(id int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channels[id] {
		return nil
	}
	if err := c.subscribe(id); err != nil {
		return err
	}
	c.channels[id] = true
	return nil
}

// Messages returns the channel on which parsed ChatMessages are delivered.
// The channel is closed when the client shuts down.
func (c *Client) Messages() <-chan ChatMessage {
	return c.msgCh
}

// Errors returns the channel on which non-fatal errors are reported
// (e.g. a failed reconnect attempt). Fatal errors close Messages().
func (c *Client) Errors() <-chan error {
	return c.errCh
}

// Close shuts down the client and closes the Messages channel.
func (c *Client) Close() {
	c.cancel()
	c.mu.Lock()
	c.ws.Close()
	c.mu.Unlock()
}

// ── internal ──────────────────────────────────────────────────────────────────

func (c *Client) log(msg string) {
	if c.debug {
		fmt.Println("[kick-chat]", msg)
	}
}

// subscribe sends a Pusher subscribe frame. Caller must hold c.mu.
func (c *Client) subscribe(chatroomID int) error {
	var req pusherSubscribe
	req.Event = "pusher:subscribe"
	req.Data.Channel = "chatrooms." + strconv.Itoa(chatroomID) + ".v2"

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.ws.WriteMessage(websocket.TextMessage, b)
}

func (c *Client) reconnect() error {
	c.log("reconnecting...")
	c.ws.Close()

	ws, _, err := websocket.DefaultDialer.Dial(pusherURL, nil)
	if err != nil {
		return fmt.Errorf("kick: reconnect dial: %w", err)
	}

	c.mu.Lock()
	c.ws = ws
	prev := make(map[int]bool, len(c.channels))
	for id := range c.channels {
		prev[id] = true
	}
	c.channels = make(map[int]bool)
	c.mu.Unlock()

	for id := range prev {
		c.mu.Lock()
		err := c.subscribe(id)
		if err == nil {
			c.channels[id] = true
		}
		c.mu.Unlock()
		if err != nil {
			return fmt.Errorf("kick: rejoin %d after reconnect: %w", id, err)
		}
	}

	c.log("reconnected.")
	return nil
}

func (c *Client) readLoop() {
	defer close(c.msgCh)

	for {
		// Check if context was cancelled before reading.
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.mu.Lock()
		ws := c.ws
		c.mu.Unlock()

		_, raw, err := ws.ReadMessage()
		if err != nil {
			select {
			case <-c.ctx.Done():
				return
			default:
			}

			c.log("read error: " + err.Error())
			for {
				if reconnErr := c.reconnect(); reconnErr == nil {
					break
				} else {
					select {
					case c.errCh <- reconnErr:
					default:
					}
					select {
					case <-c.ctx.Done():
						return
					case <-time.After(5 * time.Second):
					}
				}
			}
			continue
		}

		var env pusherEnvelope
		if err := json.Unmarshal(raw, &env); err != nil {
			continue
		}

		if env.Event != "App\\Events\\ChatMessageEvent" {
			continue
		}

		var msg ChatMessage
		if err := json.Unmarshal([]byte(env.Data), &msg); err != nil {
			continue
		}

		msg.Emotes = ParseEmotes(msg.Content)

		select {
		case c.msgCh <- msg:
		case <-c.ctx.Done():
			return
		}
	}
}

package kickchat

import "time"

// Badge represents a user badge (e.g. subscriber, moderator, broadcaster).
type Badge struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Count int    `json:"count"`
}

// Identity holds visual identity info for a sender.
type Identity struct {
	Color  string  `json:"color"`
	Badges []Badge `json:"badges"`
}

// Sender is the user who sent the message.
type Sender struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Slug     string   `json:"slug"`
	Identity Identity `json:"identity"`
}

// ParsedEmote is a Kick emote extracted from message content.
type ParsedEmote struct {
	ID   string // numeric emote ID as string
	Name string // emote name (e.g. "KEKLEO")
	URL  string // CDN URL: https://files.kick.com/emotes/{id}/fullsize
}

// ChatMessage is a single chat message received from Kick.
type ChatMessage struct {
	ID         string        `json:"id"`
	ChatroomID int           `json:"chatroom_id"`
	Content    string        `json:"content"`  // raw, may contain [emote:id:name] tokens
	Type       string        `json:"type"`
	CreatedAt  time.Time     `json:"created_at"`
	Sender     Sender        `json:"sender"`
	Emotes     []ParsedEmote // populated after parsing Content
}

// raw Pusher envelope — inner Data is a JSON string (double-encoded)
type pusherEnvelope struct {
	Event   string `json:"event"`
	Data    string `json:"data"`
	Channel string `json:"channel"`
}

// raw subscribe request sent to Pusher
type pusherSubscribe struct {
	Event string `json:"event"`
	Data  struct {
		Channel string `json:"channel"`
		Auth    string `json:"auth"`
	} `json:"data"`
}

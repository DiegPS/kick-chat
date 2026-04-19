package kickchat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const kickAPIBase = "https://kick.com/api/v2"

// GetChatroomID resolves a Kick channel slug to its numeric chatroom ID.
// Uses the public Kick API — no authentication required.
func GetChatroomID(slug string) (int, error) {
	req, err := http.NewRequest(http.MethodGet, kickAPIBase+"/channels/"+slug, nil)
	if err != nil {
		return 0, err
	}

	// Kick's API requires browser-like headers to avoid 403.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://kick.com/")
	req.Header.Set("Origin", "https://kick.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("kick: channel %q not found (HTTP %d)", slug, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// The chatroom ID may be at top-level or nested under "chatroom".
	var payload struct {
		ChatroomID int `json:"chatroom_id"`
		Chatroom   struct {
			ID int `json:"id"`
		} `json:"chatroom"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, fmt.Errorf("kick: failed to parse channel response: %w", err)
	}

	id := payload.ChatroomID
	if id == 0 {
		id = payload.Chatroom.ID
	}
	if id == 0 {
		return 0, fmt.Errorf("kick: chatroom ID not found in response for %q", slug)
	}
	return id, nil
}

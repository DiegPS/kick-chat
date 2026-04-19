package kickchat

import (
	"fmt"
	"regexp"
)

const emoteCDN = "https://files.kick.com/emotes/%s/fullsize"

var emotePattern = regexp.MustCompile(`\[emote:(\d+):([^\]]+)\]`)

// ParseEmotes extracts all [emote:id:name] tokens from raw message content
// and returns them as ParsedEmote slice with CDN URLs.
func ParseEmotes(content string) []ParsedEmote {
	matches := emotePattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool, len(matches))
	emotes := make([]ParsedEmote, 0, len(matches))

	for _, m := range matches {
		id, name := m[1], m[2]
		key := id + ":" + name
		if seen[key] {
			continue
		}
		seen[key] = true
		emotes = append(emotes, ParsedEmote{
			ID:   id,
			Name: name,
			URL:  fmt.Sprintf(emoteCDN, id),
		})
	}
	return emotes
}

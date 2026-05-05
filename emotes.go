package kickchat

import (
	"fmt"
	"regexp"
)

const emoteCDN = "https://files.kick.com/emotes/%s/fullsize"

var emotePattern = regexp.MustCompile(`\[emote:(\d+):([^\]]+)\]`)

// ParseEmotes extracts all [emote:id:name] tokens from raw message content
// and returns them as a deduplicated ParsedEmote slice with CDN URLs.
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

// ParseMessage splits raw message content into a []MessagePart sequence
// preserving order — plain text segments and emote items interleaved exactly
// as they appear in the message. Suitable for direct rendering.
func ParseMessage(content string) []MessagePart {
	var parts []MessagePart
	last := 0

	for _, m := range emotePattern.FindAllStringSubmatchIndex(content, -1) {
		if m[0] > last {
			parts = append(parts, MessagePart{Text: content[last:m[0]]})
		}
		id := content[m[2]:m[3]]
		name := content[m[4]:m[5]]
		parts = append(parts, MessagePart{
			Emote: &ParsedEmote{
				ID:   id,
				Name: name,
				URL:  fmt.Sprintf(emoteCDN, id),
			},
		})
		last = m[1]
	}

	if last < len(content) {
		parts = append(parts, MessagePart{Text: content[last:]})
	}
	if len(parts) == 0 {
		return []MessagePart{{Text: content}}
	}
	return parts
}

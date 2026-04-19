package kickchat_test

import (
	"context"
	"fmt"

	kickchat "github.com/DiegPS/kick-chat"
)

func ExampleClient() {
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
		fmt.Printf("[%s] %s: %s\n", msg.CreatedAt.Format("15:04:05"), msg.Sender.Username, msg.Content)
		for _, e := range msg.Emotes {
			fmt.Printf("  emote: %s → %s\n", e.Name, e.URL)
		}
	}
}

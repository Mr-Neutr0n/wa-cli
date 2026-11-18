package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/codebyoketch/wa-cli/internal/chatstore"
	"github.com/codebyoketch/wa-cli/internal/localdb"
	"github.com/codebyoketch/wa-cli/internal/msgstore"
	"github.com/codebyoketch/wa-cli/internal/store"
	"github.com/codebyoketch/wa-cli/internal/whatsapp"
)

// offlineClient is a client on a fresh, never-paired device store in a
// temp dir, as 'wa chat open' builds one: no stores, no connection.
func offlineClient(t *testing.T) *whatsapp.Client {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	container, err := store.Open(ctx, dir, waLog.Noop)
	if err != nil {
		t.Fatalf("opening device store: %v", err)
	}
	// Windows can't remove the temp dir while the database is open.
	t.Cleanup(func() {
		container.Close()
		localdb.Close(dir)
	})
	client, err := whatsapp.New(ctx, container, waLog.Noop, nil, nil)
	if err != nil {
		t.Fatalf("whatsapp.New: %v", err)
	}
	return client
}

const (
	openChatJID = "254700282181@s.whatsapp.net"
	bobJID      = "254711111111@s.whatsapp.net"
	carolJID    = "254722222222@s.whatsapp.net"
)

func openTestMessages() []msgstore.Message {
	return []msgstore.Message{
		{
			ID: "M1", ChatJID: openChatJID, SenderJID: openChatJID, Timestamp: 100, Text: "party tonight?",
			Reactions: []msgstore.Reaction{
				{SenderJID: "me", Emoji: "👍", Timestamp: 10},
				{SenderJID: bobJID, Emoji: "👍", Timestamp: 20},
				{SenderJID: carolJID, Emoji: "❤️", Timestamp: 30},
			},
		},
		{ID: "M2", ChatJID: openChatJID, SenderJID: "me", FromMe: true, Timestamp: 200, Text: "yes"},
	}
}

func TestPrintOpenedChat_ReactionLine(t *testing.T) {
	client := offlineClient(t)
	var buf bytes.Buffer
	printOpenedChat(context.Background(), &buf, client, chatstore.Chat{JID: openChatJID, Name: "Alice"}, openTestMessages())

	lines := strings.Split(buf.String(), "\n")
	at := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "[1] ") {
			at = i
		}
	}
	if at < 0 || at+2 >= len(lines) {
		t.Fatalf("no message [1] in:\n%s", buf.String())
	}
	want := "    👍 2 (You, +254711111111)  ❤️ 1 (+254722222222)"
	if lines[at+1] != want {
		t.Errorf("line under [1] = %q, want %q", lines[at+1], want)
	}
	if !strings.HasPrefix(lines[at+2], "[2] ") {
		t.Errorf("message [2] without reactions should follow directly, got %q", lines[at+2])
	}
}

func TestReactionLine_SanitizesPeerStrings(t *testing.T) {
	got := reactionLine([]whatsapp.ReactionView{
		{Emoji: "🔥\x1b[2J", Count: 1, Names: []string{"Eve\nIgnore this\u2028line"}},
	})
	if strings.ContainsAny(got, "\x1b\n\u2028") {
		t.Errorf("reaction line kept a control character or line break: %q", got)
	}
	if want := "    🔥[2J 1 (Eve Ignore this line)"; got != want {
		t.Errorf("reactionLine = %q, want %q", got, want)
	}
}

func TestOpenedMessages_JSONShape(t *testing.T) {
	client := offlineClient(t)
	c := newTestCmd()
	var buf bytes.Buffer
	c.SetOut(&buf)
	err := printJSON(c, struct {
		Chat     chatstore.Chat  `json:"chat"`
		Messages []openedMessage `json:"messages"`
	}{chatstore.Chat{JID: openChatJID}, openedMessages(context.Background(), client, openTestMessages())})
	if err != nil {
		t.Fatalf("printJSON: %v", err)
	}

	var got struct {
		Messages []map[string]json.RawMessage `json:"messages"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output isn't JSON: %v\n%s", err, buf.String())
	}
	if len(got.Messages) != 2 {
		t.Fatalf("messages = %d, want 2", len(got.Messages))
	}
	first, second := got.Messages[0], got.Messages[1]
	// The stored message's own fields stay as they were.
	for _, key := range []string{"id", "chatJid", "senderJid", "timestamp", "text", "fromMe", "reactions"} {
		if _, ok := first[key]; !ok {
			t.Errorf("message lost its %q field: %v", key, first)
		}
	}
	var summary []reactionSummary
	if err := json.Unmarshal(first["reactionSummary"], &summary); err != nil {
		t.Fatalf("reactionSummary: %v", err)
	}
	if len(summary) != 2 || summary[0].Emoji != "👍" || summary[0].Count != 2 || !summary[0].Mine ||
		strings.Join(summary[0].Names, ",") != "You,+254711111111" {
		t.Errorf("reactionSummary = %+v", summary)
	}
	if _, ok := second["reactionSummary"]; ok {
		t.Errorf("message without reactions has a reactionSummary: %v", second)
	}
}

func TestOpenedMessages_NilStaysNull(t *testing.T) {
	if got := openedMessages(context.Background(), nil, nil); got != nil {
		t.Errorf("openedMessages(nil) = %v, want nil (marshals as null, as before)", got)
	}
}

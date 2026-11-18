package cmd

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/codebyoketch/wa-cli/internal/chatstore"
	"github.com/codebyoketch/wa-cli/internal/config"
	"github.com/codebyoketch/wa-cli/internal/localdb"
	"github.com/codebyoketch/wa-cli/internal/msgstore"
)

func TestReactArgs(t *testing.T) {
	cases := []struct {
		name    string
		remove  bool
		args    []string
		wantErr string
	}{
		{"emoji", false, []string{"Alice", "3", "👍"}, ""},
		{"remove", true, []string{"Alice", "3"}, ""},
		{"emoji and remove", true, []string{"Alice", "3", "👍"}, "not both"},
		{"neither", false, []string{"Alice", "3"}, "missing <emoji>"},
		{"empty emoji", false, []string{"Alice", "3", " "}, "empty emoji"},
		{"too few", false, []string{"Alice"}, "accepts 3 arg(s)"},
		{"too few with remove", true, []string{"Alice"}, "accepts 2 arg(s)"},
		{"too many", false, []string{"Alice", "3", "👍", "❤️"}, "accepts 3 arg(s)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &cobra.Command{Use: "react"}
			c.Flags().Bool("remove", false, "")
			if tc.remove {
				if err := c.Flags().Set("remove", "true"); err != nil {
					t.Fatalf("setting --remove: %v", err)
				}
			}
			err := reactArgs(c, tc.args)
			switch {
			case tc.wantErr == "" && err != nil:
				t.Errorf("reactArgs(%q) = %v, want nil", tc.args, err)
			case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
				t.Errorf("reactArgs(%q) = %v, want an error containing %q", tc.args, err, tc.wantErr)
			}
		})
	}
}

func TestChatReactCmd_Wiring(t *testing.T) {
	if chatReactCmd.Flags().Lookup("remove") == nil {
		t.Fatal("wa chat react has no --remove flag")
	}
	if err := chatReactCmd.Args(chatReactCmd, []string{"Alice", "3"}); err == nil {
		t.Error("wa chat react accepted a message-ref without an emoji or --remove")
	}
	if chatReactCmd.ValidArgsFunction == nil {
		t.Error("wa chat react has no shell completion")
	}
	if found, _, err := chatCmd.Find([]string{"react"}); err != nil || found != chatReactCmd {
		t.Errorf("'wa chat react' isn't registered under 'wa chat': %v", err)
	}
}

func TestResolveReactTarget(t *testing.T) {
	dir := withAppDataDir(t)
	t.Cleanup(func() { localdb.Close(dir) })
	jid := "254700282181@s.whatsapp.net"
	if err := chatstore.New(dir).Upsert(chatstore.Chat{JID: jid, Name: "Alice"}); err != nil {
		t.Fatalf("seeding chatstore: %v", err)
	}
	ms := msgstore.New(dir)
	_ = ms.Append(msgstore.Message{ID: "m1", ChatJID: jid, Text: "first", Timestamp: 100})
	_ = ms.Append(msgstore.Message{ID: "3EB0ABCDEF", ChatJID: jid, Text: "second", Timestamp: 200})

	for _, ref := range []string{"2", "3EB0ABCDEF"} {
		gotJID, m, err := resolveReactTarget("Alice", ref)
		if err != nil {
			t.Fatalf("resolveReactTarget(%q): %v", ref, err)
		}
		if gotJID.String() != jid || m.ID != "3EB0ABCDEF" || m.ChatJID != jid {
			t.Errorf("resolveReactTarget(%q) = %s, %+v, want message 3EB0ABCDEF in %s", ref, gotJID, m, jid)
		}
	}
	if _, _, err := resolveReactTarget("Alice", "3"); err == nil {
		t.Error("expected an error for a message number past the end of the history")
	}
	if _, _, err := resolveReactTarget("Alice", "no-such-id"); err == nil {
		t.Error("expected an error for an unknown message ID")
	}
}

func TestPrintReaction_JSONShape(t *testing.T) {
	cases := []struct {
		r    reactionResult
		want map[string]any
	}{
		{
			reactionResult{ChatJID: "254700282181@s.whatsapp.net", MessageID: "M1", Emoji: "👍"},
			map[string]any{"chatJid": "254700282181@s.whatsapp.net", "messageId": "M1", "emoji": "👍", "removed": false},
		},
		{
			reactionResult{ChatJID: "254700282181@s.whatsapp.net", MessageID: "M1", Removed: true},
			map[string]any{"chatJid": "254700282181@s.whatsapp.net", "messageId": "M1", "emoji": "", "removed": true},
		},
	}
	for _, tc := range cases {
		c := newTestCmd()
		var buf bytes.Buffer
		c.SetOut(&buf)
		if err := c.Flags().Set("json", "true"); err != nil {
			t.Fatalf("setting --json: %v", err)
		}
		if err := printReaction(c, tc.r); err != nil {
			t.Fatalf("printReaction: %v", err)
		}
		var got map[string]any
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("output isn't JSON: %v\n%s", err, buf.String())
		}
		if len(got) != len(tc.want) {
			t.Errorf("keys = %v, want %v", sortedKeys(got), sortedKeys(tc.want))
		}
		for k, v := range tc.want {
			if got[k] != v {
				t.Errorf("%s = %#v, want %#v", k, got[k], v)
			}
		}
	}
}

func TestPrintReaction_Text(t *testing.T) {
	withApp(t, config.Config{JSONOutput: false})
	for _, tc := range []struct {
		r    reactionResult
		want string
	}{
		{reactionResult{Emoji: "👍"}, "Reacted 👍.\n"},
		{reactionResult{Removed: true}, "Reaction removed.\n"},
	} {
		c := newTestCmd()
		var buf bytes.Buffer
		c.SetOut(&buf)
		if err := printReaction(c, tc.r); err != nil {
			t.Fatalf("printReaction: %v", err)
		}
		if buf.String() != tc.want {
			t.Errorf("output = %q, want %q", buf.String(), tc.want)
		}
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

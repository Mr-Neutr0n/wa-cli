package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.mau.fi/whatsmeow/types"

	"github.com/codebyoketch/wa-cli/internal/msgstore"
)

var chatReactCmd = &cobra.Command{
	Use:   "react <chat> <message-ref> (<emoji> | --remove)",
	Short: "React to a message with an emoji, or remove your reaction",
	Long: `Set your reaction on a message in a chat. <message-ref> is the number
shown next to a message in 'wa chat open <chat>' (e.g. "3"), or a raw
WhatsApp message ID if you already have one.

Reacting again replaces your earlier reaction; --remove clears it instead
of taking an emoji.`,
	Args:              reactArgs,
	ValidArgsFunction: completeFirstArg(completeChatNames),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		jid, target, err := resolveReactTarget(args[0], args[1])
		if err != nil {
			return err
		}
		emoji := ""
		if len(args) == 3 {
			emoji = args[2]
		}

		if err := checkRateLimit(); err != nil {
			return err
		}

		client, _, err := connectForSend(ctx)
		if err != nil {
			return err
		}
		defer client.Disconnect()

		if err := client.React(ctx, target, emoji); err != nil {
			return err
		}
		recordRateLimit()

		return printReaction(cmd, reactionResult{
			ChatJID:   jid.String(),
			MessageID: target.ID,
			Emoji:     emoji,
			Removed:   emoji == "",
		})
	},
}

func init() {
	chatReactCmd.Flags().Bool("remove", false, "remove your reaction instead of setting one")
	chatCmd.AddCommand(chatReactCmd)
}

// reactArgs takes either an emoji as the third argument or --remove,
// never both.
func reactArgs(cmd *cobra.Command, args []string) error {
	remove, err := cmd.Flags().GetBool("remove")
	if err != nil {
		return err
	}
	if remove {
		if len(args) == 3 {
			return fmt.Errorf("pass either an emoji or --remove, not both")
		}
		return cobra.ExactArgs(2)(cmd, args)
	}
	if len(args) == 2 {
		return fmt.Errorf("missing <emoji>: pass one, or --remove to clear your reaction")
	}
	if err := cobra.ExactArgs(3)(cmd, args); err != nil {
		return err
	}
	if strings.TrimSpace(args[2]) == "" {
		return fmt.Errorf("empty emoji: pass --remove to clear your reaction")
	}
	return nil
}

// resolveReactTarget finds the chat and the message to react to, the
// same way 'wa chat reply' does.
func resolveReactTarget(chat, ref string) (types.JID, msgstore.Message, error) {
	jid, err := resolveJID(chat)
	if err != nil {
		return types.JID{}, msgstore.Message{}, err
	}
	m, err := resolveMessageRef(msgstore.New(a.Config.DataDir), jid.String(), ref)
	if err != nil {
		return types.JID{}, msgstore.Message{}, err
	}
	return jid, m, nil
}

// reactionResult is what 'wa chat react' reports.
type reactionResult struct {
	ChatJID   string `json:"chatJid"`
	MessageID string `json:"messageId"`
	Emoji     string `json:"emoji"` // "" when removed
	Removed   bool   `json:"removed"`
}

func printReaction(cmd *cobra.Command, r reactionResult) error {
	if useJSON(cmd) {
		return printJSON(cmd, r)
	}
	if r.Removed {
		fmt.Fprintln(cmd.OutOrStdout(), "Reaction removed.")
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Reacted %s.\n", r.Emoji)
	}
	return nil
}

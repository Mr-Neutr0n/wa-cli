package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/codebyoketch/wa-cli/internal/chatstore"
	"github.com/codebyoketch/wa-cli/internal/msgstore"
	"github.com/codebyoketch/wa-cli/internal/safety"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Stay connected and print new messages as they arrive",
	Long: `wa watch keeps a long-running connection to WhatsApp open and prints
new messages as they come in.

Unlike other commands (which connect briefly and disconnect), this runs in
the foreground until you stop it with Ctrl+C. It's also the most reliable
way to keep 'wa chat list' current if one-shot syncs aren't landing
consistently on your connection — it reconnects automatically on drops
rather than racing a short sync window each time.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		guard := safety.New(a.Config.DataDir)
		client, err := loggedInClient(ctx, chatstore.New(a.Config.DataDir), msgstore.New(a.Config.DataDir))
		if err != nil {
			return err
		}
		client.SetNotifications(a.Config.NotifyEnabled, a.Config.NotifyGroups, a.Config.NotifyShowPreview)

		return client.Watch(ctx, guard)
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}

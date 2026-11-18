package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codebyoketch/wa-cli/internal/chatstore"
	"github.com/codebyoketch/wa-cli/internal/msgstore"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to WhatsApp",
	Long:  `Login to your WhatsApp account by scanning a QR code with your phone.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := openClient(ctx, libLog(), chatstore.New(a.Config.DataDir), msgstore.New(a.Config.DataDir))
		if err != nil {
			return err
		}

		if client.IsLoggedIn() {
			fmt.Println("Already logged in. Run 'wa logout' first to switch accounts.")
			return nil
		}

		if err := client.Login(ctx); err != nil {
			return err
		}

		client.Disconnect()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

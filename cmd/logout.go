package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of WhatsApp",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := openClient(ctx, libLog(), nil, nil)
		if err != nil {
			return err
		}

		if !client.IsLoggedIn() {
			fmt.Println("Not logged in.")
			return nil
		}

		if err := client.Connect(ctx); err != nil {
			return err
		}
		defer client.Disconnect()

		if err := client.Logout(ctx); err != nil {
			return err
		}

		fmt.Println("Logged out.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

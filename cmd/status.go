package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show login status",
	RunE: func(cmd *cobra.Command, args []string) error {
		var loggedIn bool
		var user string
		err := captureLibraryStdout(func() error {
			client, err := openClient(context.Background(), libLog(), nil, nil)
			if err != nil {
				return err
			}

			loggedIn = client.IsLoggedIn()
			if loggedIn {
				user = client.WA.Store.ID.User
			}
			return nil
		})
		if err != nil {
			return err
		}

		if useJSON(cmd) {
			return printJSON(cmd, struct {
				LoggedIn bool   `json:"loggedIn"`
				User     string `json:"user,omitempty"`
			}{loggedIn, user})
		}

		if !loggedIn {
			fmt.Println("Not logged in. Run 'wa login' to get started.")
			return nil
		}
		fmt.Printf("Logged in as %s\n", user)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

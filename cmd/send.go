package cmd

import (
	"fmt"

	"github.com/spectre10/fs-cli/lib"
	"github.com/spectre10/fs-cli/session/send"
	"github.com/spf13/cobra"
)

// Represents the send command.
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "To send a file",
	Long: `This command is used to send a file. For example,
    $ fs-cli send <PathAndNameOfFile> ... ...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("Missing file path")
		}
		session := send.NewSession(len(args))

		err := session.SetupConnection(args)
		if err != nil {
			return err
		}

		err = session.PrintOffer()
		if err != nil {
			return err
		}

		answer, err := lib.SDPPrompt()
		if err != nil {
			return err
		}

		err = session.Connect(answer)
		if err != nil {
			return err
		}
		return nil
	},
}

// Add send command.
func init() {
	rootCmd.AddCommand(sendCmd)
}

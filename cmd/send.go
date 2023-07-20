package cmd

import (
	"fmt"

	"github.com/spectre10/fs-cli/session/send"
	"github.com/spf13/cobra"
)

// Represents the send command.
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "To send a file",
	Long: `This command is used to send a file. For example,
    $ fs-cli send <PathAndNameOfFile> ... ...`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Missing file path")
			return
		}
		session := send.NewSession(len(args))
		err := session.Connect(args)
		if err != nil {
			panic(err)
		}
	},
}

// Add send command.
func init() {
	rootCmd.AddCommand(sendCmd)
}

package cmd

import (
	"fmt"

	"github.com/spectre10/fileshare-cli/session/send"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "To send a file",
	Long: `This command is used to send a file. For example,
    $ fileshare-cli send <PathAndNameOfFile> ... ...`,
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

func init() {
	rootCmd.AddCommand(sendCmd)
}

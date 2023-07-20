package cmd

import (
	"github.com/spectre10/fs-cli/session/receive"
	"github.com/spf13/cobra"
)

// receiveCmd represents the receive command
var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "To receive a file",
	Long: `Receive a file via this command. 
    For example,
    $ fs-cli receive`,
	Run: func(cmd *cobra.Command, args []string) {
		session := receive.NewSession()
		err := session.Connect()
		if err != nil {
			panic(err)
		}
	},
}

// Add receive command.
func init() {
	rootCmd.AddCommand(receiveCmd)
}

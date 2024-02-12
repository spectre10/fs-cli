package cmd

import (
	"github.com/spectre10/fs-cli/lib"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		session := receive.NewSession()
		err := session.CreateConnection()
		if err != nil {
			return err
		}

		answer, err := lib.SDPPrompt()
		if err != nil {
			return err
		}

		err = session.PrintSDP(answer)
		if err != nil {
			return err
		}

		err = session.Connect(answer)
		return err
	},
}

// Add receive command.
func init() {
	rootCmd.AddCommand(receiveCmd)
}

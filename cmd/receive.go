package cmd

import (
	"github.com/spectre10/fileshare-cli/session/receive"
	"github.com/spf13/cobra"
)

// receiveCmd represents the receive command
var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "To receive a file",
	Long: `Receive a file via this command. 
    For example,
    $ fileshare-cli receive`,
	Run: func(cmd *cobra.Command, args []string) {
		session := receive.NewSession()
		err := session.Connect()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(receiveCmd)
	receiveCmd.PersistentFlags().StringP("file", "f", "", "name and path of the file")
}
